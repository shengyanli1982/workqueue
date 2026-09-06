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

// leasedQueueImpl 租约队列实现：在内层队列上叠加租约消费语义。GetWithLease 弹出
// 元素后登记租约（value + deadline），消费方通过 Ack 确认完成或 Nack 否定重入队；
// 过期租约由 reaper 协程定期扫描并自动重入队，实现 at-least-once 交付保障。
type leasedQueueImpl struct {
	Queue
	config *LeasedQueueConfig

	lock    sync.Mutex
	leases  map[string]leasedItem
	leaseID atomic.Uint64

	closed chan struct{}
	once   sync.Once
	wg     sync.WaitGroup

	// draining 在 drain 开始时置位：拒绝新的用户 Put；
	// Nack/reaper 经内层队列的归还路径不受影响，在租项照常回收。
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

// GetWithLease 从队列弹出一个元素并登记租约：返回元素值与唯一租约 ID。
// timeout<=0 时回退到 config.leaseDuration；pop 与租约登记在同一临界区完成，
// 保证 drain 判定不会落入"元素已出队但租约未登记"的窗口。
func (q *leasedQueueImpl) GetWithLease(timeout time.Duration) (value any, leaseID string, err error) {
	if timeout <= 0 {
		timeout = q.config.leaseDuration
	}
	if timeout <= 0 {
		return nil, "", ErrInvalidLeaseDuration
	}

	seq := q.leaseID.Add(1)
	var raw [16]byte
	leaseID = string(strconv.AppendUint(raw[:0], seq, 36))
	deadline := time.Now().Add(timeout)

	// pop 与租约登记同在一个 q.lock 临界区：与 drainedForShutdown 串行化，
	// 保证 drain 判定不会落入“元素已出队但租约未登记”的窗口。
	q.lock.Lock()

	value, err = q.Queue.Get()
	if err != nil {
		q.lock.Unlock()
		return nil, "", err
	}

	if q.leases == nil {
		// 关停已清理租约表：刚弹出的元素随关停丢弃，语义与 Shutdown 一致。
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

// GetWithLeaseWithContext 是 GetWithLease 的阻塞变体：阻塞直到内层队列取到值、
// ctx 完成或队列关闭，成功后登记租约并返回值与租约 ID；timeout<=0 回退
// config.leaseDuration，语义与 GetWithLease 一致。
//
// 租约登记复用 GetWithLease 的同临界区路径（q.lock 内检查 leases 清理状态，
// 与 drainedForShutdown 串行化）；阻塞等待阶段不持 q.lock。等待期间若关停
// 恰在“取到值后、登记前”窗口清理 leases 表，该元素随关停丢弃并返回
// ErrQueueIsClosed（与 GetWithLease 遇清理的丢弃语义一致）。
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
	value, ok := q.removeLease(leaseID)
	if !ok {
		return ErrLeaseNotFound
	}

	q.Queue.Done(value)
	return nil
}

// Nack 否定确认：将元素重新入队并释放租约，触发 OnNack 回调。
// 先入队成功后再移除租约，避免入队失败时元素丢失。
func (q *leasedQueueImpl) Nack(leaseID string, reason error) error {
	value, ok := q.peekLease(leaseID)
	if !ok {
		return ErrLeaseNotFound
	}

	// 先入队，成功后再释放租约，避免入队失败时元素丢失。
	if err := q.Queue.Put(value); err != nil {
		return err
	}

	q.removeLease(leaseID)
	q.Queue.Done(value)
	q.config.callback.OnNack(value, reason)
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

// LeaseInfos 返回当前全部在租租约的只读快照。
// q.lock 持锁拷贝所有租约条目，锁外返回副本：调用方可安全遍历、修改切片
// 而不影响内部 leases 表；快照期间不阻塞并发的 GetWithLease/Ack/Nack。
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

// ShutdownWithDrain 优雅关停：置位 draining 拒绝新用户入队，等待内层队列
// drained ∧ leases 清空后关闭；超时或取消时强制关闭并返回 ctx.Err()。
// 在租项由 reaper 照常回收过期租约→重入队→被消费后 Ack 释放；
// 超时后残余租约随强制关闭丢弃（与 Shutdown 的 leases=nil 语义一致）。
func (q *leasedQueueImpl) ShutdownWithDrain(ctx context.Context) error {

	if q.IsClosed() {
		return nil
	}

	q.draining.Store(true)

	inner := q.Queue.(*queueImpl)

	err := waitForDrain(ctx, inner.IsClosed, func() bool {
		return q.drainedForShutdown(inner)
	})

	// 与 Shutdown 一致的关闭顺序：先停 reaper 并清空 leases，再关内层队列。
	q.once.Do(q.closeNow)
	q.Queue.Shutdown()

	return err
}

// drainedForShutdown 判定租约队列 drain 完成：无在租项 ∧ 内层队列已 drained。
// 持 q.lock 跨越内层判定：reaper 的“Put→removeLease”与 GetWithLease 的
// “pop→登记租约”两段过渡均经 q.lock 串行化，避免判定窗口内元素既不在
// leases 也不在内层队列视野内（TOCTOU）。
func (q *leasedQueueImpl) drainedForShutdown(inner *queueImpl) bool {

	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.leases) != 0 {
		return false
	}

	return inner.isDrained()
}

// Put 覆写嵌入 Queue：drain 期间拒绝新用户入队；
// Nack/reaper 走内层队列路径完成在租项归还，不受 draining 影响。
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

// peekLease 只读查看租约内容，不删除。
func (q *leasedQueueImpl) peekLease(leaseID string) (value any, ok bool) {
	if leaseID == "" {
		return nil, false
	}

	q.lock.Lock()
	item, ok := q.leases[leaseID]
	q.lock.Unlock()

	if !ok {
		return nil, false
	}
	return item.value, true
}

// removeLease 移除指定租约并返回其元素值：空 ID 或未命中返回 ok=false。
func (q *leasedQueueImpl) removeLease(leaseID string) (value any, ok bool) {
	if leaseID == "" {
		return nil, false
	}

	q.lock.Lock()
	item, ok := q.leases[leaseID]
	if ok {
		delete(q.leases, leaseID)
	}
	q.lock.Unlock()

	if !ok {
		return nil, false
	}
	return item.value, true
}

// requeueExpiredLeases reaper 协程：按 scanInterval 周期扫描过期租约，
// 将过期元素重新入队并释放租约，实现 at-least-once 交付保障。
// 入队失败时元素保留在 leases 中，下次扫描再试。
func (q *leasedQueueImpl) requeueExpiredLeases() {
	ticker := time.NewTicker(q.config.scanInterval)
	defer func() {
		ticker.Stop()
		q.wg.Done()
	}()

	// 复用缓冲区，稳态下（无过期租约）零分配，有过期租约时复用底层数组。
	var expiredBuf []expiredLease

	for {
		select {
		case <-q.closed:
			return
		case <-ticker.C:
			now := time.Now()
			expired := q.collectExpired(now, expiredBuf[:0])
			for _, e := range expired {
				// 先入队，成功后再释放租约，避免入队失败时元素丢失。
				// 入队失败的元素保留在 leases map 中，下次扫描时再试。
				if err := q.Queue.Put(e.value); err != nil {
					continue
				}
				q.removeLease(e.id)
				q.Queue.Done(e.value)
			}
			expiredBuf = expired
		}
	}
}

// expiredLease 过期租约的中间表示，由 collectExpired 收集供 requeueExpiredLeases 处理。
type expiredLease struct {
	id    string
	value any
}

// collectExpired 只收集过期的租约条目，不从 leases map 中移除。
// buf 为复用的缓冲区，调用方传入 buf[:0] 以复用底层数组，减少 GC 压力。
func (q *leasedQueueImpl) collectExpired(now time.Time, buf []expiredLease) []expiredLease {
	q.lock.Lock()
	for id, item := range q.leases {
		if !item.deadline.After(now) {
			buf = append(buf, expiredLease{id: id, value: item.value})
		}
	}
	q.lock.Unlock()
	return buf
}
