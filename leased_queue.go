package workqueue

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// leasedItem 租约表中的单条记录：元素值与租约到期时间。
type leasedItem struct {
	value    any
	deadline time.Time
}

// LeaseInfo 是单个在租租约的只读快照。
type LeaseInfo struct {
	// LeaseID 为租约唯一标识。
	LeaseID string

	// Value 为租约持有的元素。
	Value any

	// Deadline 为租约到期时间，到期后由 reaper 回收重入队。
	Deadline time.Time
}

// leasedQueueImpl 租约队列：GetWithLease 弹出元素并登记租约，
// 消费方通过 Ack 确认或 Nack 重入队；过期租约由 reaper 自动回收，
// 实现 at-least-once 交付。
type leasedQueueImpl struct {
	Queue
	config *LeasedQueueConfig

	lock    sync.Mutex
	leases  map[string]leasedItem
	leaseID atomic.Uint64

	closed chan struct{}
	once   sync.Once
	wg     sync.WaitGroup

	// draining drain 开始时置位，拒绝新的用户 Put；
	// Nack/reaper 的归还路径不受影响。
	draining atomic.Bool
}

// NewLeasedQueue 创建租约队列。
func NewLeasedQueue(config *LeasedQueueConfig) LeasedQueue {
	config = isLeasedQueueConfigEffective(config)

	q := &leasedQueueImpl{
		Queue:  NewQueue(&config.QueueConfig),
		config: config,
		leases: make(map[string]leasedItem),
		closed: make(chan struct{}),
	}

	q.wg.Add(1)
	go q.requeueExpiredLeases()

	return q
}

// GetWithLease 弹出元素并登记租约，返回值与租约 ID。
// timeout<=0 时回退到 config.leaseDuration；pop 在锁外完成，
// 租约登记持锁，避免回调重入死锁。
func (q *leasedQueueImpl) GetWithLease(timeout time.Duration) (value any, leaseID string, err error) {
	if timeout <= 0 {
		timeout = q.config.leaseDuration
	}
	if timeout <= 0 {
		return nil, "", ErrInvalidLeaseDuration
	}

	value, err = q.Queue.Get()
	if err != nil {
		return nil, "", err
	}

	seq := q.leaseID.Add(1)
	var raw [16]byte
	leaseID = string(strconv.AppendUint(raw[:0], seq, 36))
	deadline := time.Now().Add(timeout)

	q.lock.Lock()

	if q.leases == nil {
		// 关停已清理租约表，元素随关停丢弃。
		q.lock.Unlock()
		return nil, "", ErrQueueIsClosed
	}

	q.leases[leaseID] = leasedItem{
		value:    value,
		deadline: deadline,
	}
	q.lock.Unlock()

	return value, leaseID, nil
}

// GetWithLeaseWithContext 是 GetWithLease 的阻塞变体：支持 ctx 取消，
// 语义与 GetWithLease 一致。
func (q *leasedQueueImpl) GetWithLeaseWithContext(ctx context.Context, timeout time.Duration) (value any, leaseID string, err error) {
	if timeout <= 0 {
		timeout = q.config.leaseDuration
	}
	if timeout <= 0 {
		return nil, "", ErrInvalidLeaseDuration
	}

	value, err = q.Queue.(BlockingGetQueue).GetWithContext(ctx)
	if err != nil {
		return nil, "", err
	}

	seq := q.leaseID.Add(1)
	var raw [16]byte
	leaseID = string(strconv.AppendUint(raw[:0], seq, 36))
	deadline := time.Now().Add(timeout)

	q.lock.Lock()

	if q.leases == nil {
		q.lock.Unlock()
		return nil, "", ErrQueueIsClosed
	}

	q.leases[leaseID] = leasedItem{
		value:    value,
		deadline: deadline,
	}
	q.lock.Unlock()

	return value, leaseID, nil
}

// Ack 确认租约：移除租约记录并调用 Done 释放处理中追踪。租约不存在时返回 ErrLeaseNotFound。
func (q *leasedQueueImpl) Ack(leaseID string) error {
	item, ok := q.removeLease(leaseID)
	if !ok {
		return ErrLeaseNotFound
	}

	q.Queue.Done(item.value)
	return nil
}

// Nack 否定确认：移除租约、元素重入队并触发 OnNack。
// claim-first 语义：先移除租约再入队，避免与 reaper 竞态导致重复入队。
// 入队失败时租约按原 deadline 恢复，元素不丢失。
func (q *leasedQueueImpl) Nack(leaseID string, reason error) error {
	item, ok := q.removeLease(leaseID)
	if !ok {
		return ErrLeaseNotFound
	}

	if err := q.Queue.Put(item.value); err != nil {
		q.restoreLease(leaseID, item)
		return err
	}

	q.Queue.Done(item.value)
	q.config.callback.OnNack(item.value, reason)
	return nil
}

// ExtendLease 续租：将指定租约的到期时间重置为当前时间 + timeout。
// timeout<=0 返回 ErrInvalidLeaseDuration；租约不存在返回 ErrLeaseNotFound。
func (q *leasedQueueImpl) ExtendLease(leaseID string, timeout time.Duration) error {
	if timeout <= 0 {
		return ErrInvalidLeaseDuration
	}

	q.lock.Lock()
	item, ok := q.leases[leaseID]
	if ok {
		item.deadline = time.Now().Add(timeout)
		q.leases[leaseID] = item
	}
	q.lock.Unlock()

	if !ok {
		return ErrLeaseNotFound
	}

	return nil
}

// LeaseInfos 返回当前全部在租租约的只读快照，调用方可安全遍历。
func (q *leasedQueueImpl) LeaseInfos() []LeaseInfo {
	q.lock.Lock()

	infos := make([]LeaseInfo, 0, len(q.leases))
	for id, item := range q.leases {
		infos = append(infos, LeaseInfo{
			LeaseID:  id,
			Value:    item.value,
			Deadline: item.deadline,
		})
	}

	q.lock.Unlock()

	return infos
}

func (q *leasedQueueImpl) Shutdown() {
	q.once.Do(q.closeNow)

	q.Queue.Shutdown()
}

// ShutdownWithDrain 优雅关停：置位 draining 拒绝新入队，等待内层队列
// drained ∧ leases 清空；超时或取消时强制关闭并返回 ctx.Err()。
func (q *leasedQueueImpl) ShutdownWithDrain(ctx context.Context) error {

	if q.IsClosed() {
		return nil
	}

	q.draining.Store(true)

	inner := q.Queue.(*queueImpl)

	err := waitForDrain(ctx, inner.IsClosed, func() bool {
		return q.drainedForShutdown(inner)
	})

	// 与 Shutdown 一致的关闭顺序。
	q.once.Do(q.closeNow)
	q.Queue.Shutdown()

	return err
}

// drainedForShutdown 判定 drain 完成：无在租项 ∧ 内层队列已 drained。
// 持锁跨越内层判定，避免 TOCTOU；内层簿记覆盖过渡窗口。
func (q *leasedQueueImpl) drainedForShutdown(inner *queueImpl) bool {

	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.leases) != 0 {
		return false
	}

	return inner.isDrained()
}

// IsDraining 返回是否处于排空阶段。三态语义见 DrainingQueue 接口。
func (q *leasedQueueImpl) IsDraining() bool {
	return q.draining.Load()
}

// Put 覆写嵌入 Queue：drain 期间拒绝新用户入队。
func (q *leasedQueueImpl) Put(value any) error {

	if q.draining.Load() {
		return ErrQueueIsClosed
	}

	return q.Queue.Put(value)
}

// closeNow 停止 reaper 并清空 leases。
func (q *leasedQueueImpl) closeNow() {
	close(q.closed)
	q.wg.Wait()

	q.lock.Lock()
	q.leases = nil
	q.lock.Unlock()
}

// restoreLease 按原样恢复租约：claim 后入队失败或陈旧化复查路径均经此恢复。
// leases 已被 closeNow 清 nil 时随关停丢弃。
func (q *leasedQueueImpl) restoreLease(leaseID string, item leasedItem) {
	q.lock.Lock()
	if q.leases != nil {
		q.leases[leaseID] = item
	}
	q.lock.Unlock()
}

// removeLease 移除指定租约并返回其完整记录（供失败/陈旧化路径原样恢复）：
// 空 ID 或未命中返回 ok=false。
func (q *leasedQueueImpl) removeLease(leaseID string) (item leasedItem, ok bool) {
	if leaseID == "" {
		return leasedItem{}, false
	}

	q.lock.Lock()
	item, ok = q.leases[leaseID]
	if ok {
		delete(q.leases, leaseID)
	}
	q.lock.Unlock()

	if !ok {
		return leasedItem{}, false
	}
	return item, true
}

// requeueExpiredLeases reaper 协程：按 scanInterval 周期扫描过期租约，
// 将过期元素重新入队并释放租约，实现 at-least-once 交付保障。
// 入队失败时租约按原 deadline 重新登记，下次扫描再试（见 requeueCollected）。
func (q *leasedQueueImpl) requeueExpiredLeases() {
	ticker := time.NewTicker(q.config.scanInterval)
	defer func() {
		ticker.Stop()
		q.wg.Done()
	}()

	var expiredBuf []expiredLease

	for {
		select {
		case <-q.closed:
			return
		case <-ticker.C:
			now := time.Now()
			expired := q.collectExpired(now, expiredBuf[:0])
			q.requeueCollected(expired)
			expiredBuf = expired
		}
	}
}

// requeueCollected 处理过期租约：claim-first 移除后做陈旧化复查，
// deadline 已变（被续期）则恢复跳过，否则入队；失败时恢复待下轮重试。
func (q *leasedQueueImpl) requeueCollected(expired []expiredLease) {
	for _, e := range expired {
		item, ok := q.removeLease(e.id)
		if !ok {
			continue
		}
		// 陈旧化复查：deadline 已变说明被续期，恢复并跳过。
		if !item.deadline.Equal(e.deadline) {
			q.restoreLease(e.id, item)
			continue
		}
		if err := q.Queue.Put(item.value); err != nil {
			q.restoreLease(e.id, item)
			continue
		}
		q.Queue.Done(item.value)
	}
}

// expiredLease 过期租约的中间表示，deadline 作为陈旧化复查基准。
type expiredLease struct {
	id       string
	deadline time.Time
}

// collectExpired 收集过期租约条目（不移除），buf 为复用缓冲区。
func (q *leasedQueueImpl) collectExpired(now time.Time, buf []expiredLease) []expiredLease {
	q.lock.Lock()
	for id, item := range q.leases {
		if !item.deadline.After(now) {
			buf = append(buf, expiredLease{id: id, deadline: item.deadline})
		}
	}
	q.lock.Unlock()
	return buf
}
