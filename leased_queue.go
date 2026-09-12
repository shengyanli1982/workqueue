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
// timeout<=0 时回退到 config.leaseDuration；pop 在 q.lock 外完成、租约登记持锁，
// 与 GetWithLeaseWithContext 的临界区模式一致。pop 不得持 q.lock：内层 Get 在
// 释放内层锁后触发 OnGet（WithCallback 将同一回调注入内外两层），持锁跨越会使
// 回调内重入 Ack/ExtendLease/LeaseInfos 二次加锁自死锁。
// “已弹出未登记”窗口的 drain 安全性由内层簿记覆盖，见 drainedForShutdown 注释。
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
	item, ok := q.removeLease(leaseID)
	if !ok {
		return ErrLeaseNotFound
	}

	q.Queue.Done(item.value)
	return nil
}

// Nack 否定确认：将元素重新入队并释放租约，触发 OnNack 回调。
// claim-first：先移除租约宣告所有权（与 Ack/reaper 的 removeLease 互斥，消除
// “归还入队窗口内 reaper 抢占导致重复 Put/双重 Done”的竞态），claim 成功后入队；
// 入队失败时租约按原 deadline 重新登记，元素不丢失。
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
// 持 q.lock 跨越内层判定。两段锁外过渡窗口均由内层 popLocked 在与 pop 同一
// 临界区完成的簿记覆盖，判定不会因此提前为真：
//   - GetWithLease/GetWithLeaseWithContext 的“已弹出未登记租约”：幂等模式值已
//     搬入 processing、drainTracking 模式 inFlight 已 +1，isDrained 均返回 false；
//   - reaper/Nack claim-first 的“已移除租约未重入队”：原 pop 的 processing 标记 /
//     inFlight 计数要到入队成功后的 Done 才释放，isDrained 同样返回 false。
//
// 裸非幂等且未启用 tracking 的模式下，在途项本就不在判定视野内（内层队列的
// 文档化契约）；窗口内元素随关停丢弃，与 leases=nil 清理时的丢弃语义一致。
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

// restoreLease 按 item 原样（含其当前 deadline）重新登记租约、归还所有权：
// claim 后入队失败路径与陈旧化复查路径（快照窗口内被续期）均经此恢复。
// leases 已被 closeNow 清 nil（Nack 与关停并发的窗口）时随关停丢弃，
// 语义与 GetWithLease 遇清理的丢弃一致；nil map 不可写入，此处判空为正确性所需。
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

	// 复用缓冲区，稳态下（无过期租约）零分配，有过期租约时复用底层数组。
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

// requeueCollected 处理 collectExpired 收集的过期租约。
// claim-first：先 removeLease 宣告所有权（与 Ack/Nack 的 removeLease 互斥），
// 未命中说明租约已在快照窗口内被 Ack/Nack 处置，跳过——消除“已 Ack 的值被
// 陈旧快照 Put 复活 + Done 误减他项在途计数”的竞态。claim 成功后做陈旧化复查：
// deadline 已变（快照窗口内被 ExtendLease 续期）则原样恢复并跳过，避免把已续期
// 租约照陈旧快照错误重入队；否则入队，失败时重新登记待下轮重试，成功后才调
// Done 释放在途追踪。
func (q *leasedQueueImpl) requeueCollected(expired []expiredLease) {
	for _, e := range expired {
		item, ok := q.removeLease(e.id)
		if !ok {
			continue
		}
		// 陈旧化复查：claim 到的 deadline 若已不同于快照值，说明快照窗口内被
		// ExtendLease 续期，租约仍有效，不得照陈旧快照重入队——按当前 item 原样
		// 恢复（保留续期后的新 deadline）并跳过。
		if !item.deadline.Equal(e.deadline) {
			q.restoreLease(e.id, item)
			continue
		}
		if err := q.Queue.Put(item.value); err != nil {
			// 入队失败（如内层已关停）：按当前 item 重新登记，元素不丢失。
			q.restoreLease(e.id, item)
			continue
		}
		q.Queue.Done(item.value)
	}
}

// expiredLease 过期租约的中间表示，由 collectExpired 收集供 requeueCollected 处理。
// deadline 保存快照时的到期时间，作为 claim 后陈旧化复查的基准：与 removeLease 取回
// 的 item.deadline 不一致即说明租约在快照窗口内被 ExtendLease 续期，放弃重入队。
// 重新登记不以快照值为依据，一律以 removeLease 返回的 item 为准（其 deadline 即权威到期时间）。
type expiredLease struct {
	id       string
	deadline time.Time
}

// collectExpired 只收集过期的租约条目，不从 leases map 中移除。
// buf 为复用的缓冲区，调用方传入 buf[:0] 以复用底层数组，减少 GC 压力。
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
