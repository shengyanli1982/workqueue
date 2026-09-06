package workqueue

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// toDelay 将相对延迟（毫秒）转为绝对时间戳（毫秒），用于堆内排序。
func toDelay(duration int64) int64 {
	return time.Now().UnixMilli() + duration
}

// delayingQueueImpl 通过排序树维护尚未到期的元素。
type delayingQueueImpl struct {
	Queue
	config      *DelayingQueueConfig
	sorting     *hp.RBTree
	elementpool *lst.NodePool
	lock        sync.Mutex
	once        sync.Once
	wg          sync.WaitGroup
	wake        chan struct{}
	closed      chan struct{}
	// discardedDelayed 累计 Shutdown 时按 Q4a 契约丢弃的未搬运延迟项数量。
	// 本里程碑仅内部计数，由 G4/G5 里程碑对外暴露。
	discardedDelayed atomic.Int64
	// draining 在 drain 开始时置位：拒绝新的 Put/PutWithDelay，
	// 已被接受的在途项由 scheduler 继续搬运直至 drain 判定完成。
	draining atomic.Bool
	// inTransit 记录“已出堆但尚未入内层队列”的搬运中节点数，
	// 与堆顶到期检查共同构成 drain 判定，避免搬运窗口内误判 drained。
	inTransit atomic.Int64
}

// NewDelayingQueue 创建延迟队列并启动搬运协程。
func NewDelayingQueue(config *DelayingQueueConfig) DelayingQueue {
	config = isDelayingQueueConfigEffective(config)
	q := &delayingQueueImpl{
		config:      config,
		sorting:     hp.New(),
		elementpool: lst.NewNodePool(),
		once:        sync.Once{},
		wg:          sync.WaitGroup{},
		wake:        make(chan struct{}, 1),
		closed:      make(chan struct{}),
	}

	q.Queue = newQueue(&wrapInternalList{List: lst.New()}, q.elementpool, &config.QueueConfig)
	q.wg.Add(1)
	go q.scheduler()
	return q
}

func (q *delayingQueueImpl) Shutdown() {
	q.Queue.Shutdown()
	q.once.Do(q.closeNow)
}

// ShutdownWithDrain 优雅关停：置位 draining 拒绝新入队，等待内层队列 drained
// ∧ 堆中无到期未搬运项 ∧ 无搬运中节点后关闭；超时或取消时强制关闭并返回 ctx.Err()。
// 堆中未到期项在最终关闭时按 Q4a 契约丢弃（计入 discardedDelayed）。
func (q *delayingQueueImpl) ShutdownWithDrain(ctx context.Context) error {

	if q.IsClosed() {
		return nil
	}

	q.draining.Store(true)

	inner := q.Queue.(*queueImpl)

	err := waitForDrain(ctx, inner.IsClosed, func() bool {
		return q.drainedForShutdown(inner)
	})

	// 与 Shutdown 一致的关闭顺序：先关内层队列，再停 scheduler 并清理堆。
	q.Queue.Shutdown()
	q.once.Do(q.closeNow)

	return err
}

// drainedForShutdown 判定延迟队列 drain 完成：堆中无到期未搬运项 ∧ 无搬运中
// 节点 ∧ 内层队列已 drained。持 q.lock 跨越内层判定：scheduler 的
// “pop(+inTransit)→Put”过渡经 q.lock 串行化，避免判定窗口内元素既不在堆
// 也不在内层队列视野内（TOCTOU）。未到期项不阻塞 drain。
func (q *delayingQueueImpl) drainedForShutdown(inner *queueImpl) bool {

	q.lock.Lock()
	defer q.lock.Unlock()

	if q.inTransit.Load() != 0 {
		return false
	}

	if front := q.sorting.Front(); front != nil && front.Priority <= time.Now().UnixMilli() {
		return false
	}

	return inner.isDrained()
}

// closeNow 停止 scheduler 并按 Q4a 契约丢弃堆中未搬运的延迟项。
func (q *delayingQueueImpl) closeNow() {
	// 关闭 closed channel 并补发 wake，确保 scheduler 无论停在哪一路 select 都立即退出；
	// 先等 scheduler 退出再清空堆，避免搬运与清理交织。
	close(q.closed)
	q.notifyWake()
	q.wg.Wait()

	q.lock.Lock()

	// 先收集所有节点，避免遍历中归还池会 Reset 指针破坏红黑树遍历。
	// Q4a 契约：堆中未搬运的延迟项随关停丢弃（不提前投递、不等待到期）。
	nodes := make([]*lst.Node, 0, q.sorting.Len())
	q.sorting.Range(func(node *lst.Node) bool {
		nodes = append(nodes, node)
		return true
	})

	q.sorting.Cleanup()

	q.lock.Unlock()

	// 锁外统一归还池，缩短临界区；并累计丢弃数量。
	for _, node := range nodes {
		q.elementpool.Put(node)
	}
	q.discardedDelayed.Add(int64(len(nodes)))
}

// Put 覆写嵌入 Queue：drain 期间拒绝新入队；
// 内层队列与 scheduler 搬运不受 draining 影响，直到 drain 判定完成后统一关闭。
func (q *delayingQueueImpl) Put(value any) error {

	if q.draining.Load() {
		return ErrQueueIsClosed
	}

	return q.Queue.Put(value)
}

// PutWithDelay 按相对延迟（毫秒）将元素入延迟堆，到期后由 scheduler 搬入内层队列。
// 锁内复查 closed 防止关停窗口丢失节点；仅当新项成为更早堆顶时唤醒 scheduler 重算 timer。
func (q *delayingQueueImpl) PutWithDelay(value any, delay int64) error {

	if q.IsClosed() || q.draining.Load() {
		return ErrQueueIsClosed
	}
	if value == nil {
		return ErrElementIsNil
	}

	last := q.elementpool.Get()
	last.Value = value
	last.Priority = toDelay(delay)

	q.lock.Lock()
	// 锁内复查 closed：防止锁外检查通过后 Shutdown 清空堆，导致节点被 Push 进死堆静默丢失。
	// IsClosed 为 atomic load，持锁调用无重入风险（吸收 M1 的 D3 热修，与 queue.go Put 的锁内复查对齐）。
	if q.IsClosed() {
		q.lock.Unlock()
		q.elementpool.Put(last)
		return ErrQueueIsClosed
	}
	front := q.sorting.Front()
	q.sorting.Push(last)
	// 仅当堆由空变非空或新项成为更早的堆顶时才需要唤醒 scheduler 重算 timer，
	// 否则现有 timer 仍然有效，避免无谓唤醒。
	shouldWake := front == nil || last.Priority < front.Priority
	q.lock.Unlock()

	if shouldWake {
		q.notifyWake()
	}
	q.config.callback.OnDelay(value, delay)
	return nil
}

// scheduler 以“堆顶精确 timer + 事件唤醒”搬运到期元素，替代旧的固定 300ms 心跳轮询：
// 对堆顶 readyAt 设置精确 timer，四路 select —— timer.C 到期批量搬运、wake 重算堆顶
// timer、300ms ticker 兜底心跳（防御 timer 漂移）、closed 退出。
// 空堆时不设置 timer，仅等待 wake/ticker/closed，空闲队列零轮询空转。
func (q *delayingQueueImpl) scheduler() {
	defer q.wg.Done()

	heartbeat := time.NewTicker(time.Millisecond * 300)
	defer heartbeat.Stop()

	// 初始为已 Stop 的空 timer，由 wait 按堆顶到期时间按需 Reset。
	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()

	cb := q.config.callback // 局部变量捕获回调引用，避免数据竞争
	var expired []*lst.Node

	for {
		nodes, wait, ok := q.nextBatch(expired[:0])
		if !ok {
			return
		}
		expired = nodes

		if len(nodes) > 0 {
			for _, node := range nodes {
				value := node.Value
				q.elementpool.Put(node)
				if err := q.Queue.Put(value); err != nil {
					cb.OnPullError(value, err)
				}
				q.inTransit.Add(-1)
			}
			continue
		}

		if !q.wait(timer, heartbeat, wait) {
			return
		}
	}
}

// nextBatch 锁内弹出全部已到期节点（保留单批 128 上限语义）并返回至下一到期项的等待时长。
// ok=false 表示队列已关闭，调度协程应退出。堆空时 wait=0（调用方不设置 timer）。
func (q *delayingQueueImpl) nextBatch(expired []*lst.Node) ([]*lst.Node, time.Duration, bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	select {
	case <-q.closed:
		return nil, 0, false
	default:
	}

	now := time.Now().UnixMilli()
	for q.sorting.Len() > 0 && q.sorting.Front().Priority <= now {
		expired = append(expired, q.sorting.Pop())
		q.inTransit.Add(1)
		if len(expired) >= 128 {
			break
		}
	}

	if front := q.sorting.Front(); front != nil {
		return expired, time.Duration(front.Priority-now) * time.Millisecond, true
	}
	return expired, 0, true
}

// wait 阻塞等待下一搬运时机，四路 select：closed 退出、wake 重算堆顶、
// timer.C 到期搬运、heartbeat 兜底心跳。d<=0（堆空）时不设置 timer，空闲零轮询。
func (q *delayingQueueImpl) wait(timer *time.Timer, heartbeat *time.Ticker, d time.Duration) bool {
	if d <= 0 {
		select {
		case <-q.closed:
			return false
		case <-q.wake:
			return true
		case <-heartbeat.C:
			return true
		}
	}

	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(d)

	select {
	case <-q.closed:
		return false
	case <-q.wake:
		return true
	case <-timer.C:
		return true
	case <-heartbeat.C:
		return true
	}
}

// notifyWake 非阻塞发送唤醒信号：wake 为容量 1 的缓冲 channel，多个并发唤醒自然折叠为一个。
func (q *delayingQueueImpl) notifyWake() {
	select {
	case q.wake <- struct{}{}:
	default:
	}
}

// CancelDelay 取消堆中尚未搬运的单个延迟项：命中则移除节点、归还节点池、
// 唤醒 scheduler 重算堆顶 timer 并返回 true；未命中（未入堆、已搬运、已取消
// 或队列已关停清空）返回 false。value 匹配语义与 TimerQueue.Cancel 一致。
func (q *delayingQueueImpl) CancelDelay(value any) bool {
	q.lock.Lock()
	var target *lst.Node
	if front := q.sorting.Front(); front != nil && matchTimerValue(front.Value, value) {
		target = q.sorting.Pop()
	} else {
		target = q.findNodeLocked(value)
		if target != nil {
			q.sorting.Remove(target)
		}
	}
	q.lock.Unlock()

	if target == nil {
		return false
	}

	q.elementpool.Put(target)
	q.notifyWake()
	return true
}

// findNodeLocked 在持锁状态下按 matchTimerValue 语义查找堆内首个匹配节点。
func (q *delayingQueueImpl) findNodeLocked(value any) *lst.Node {
	var target *lst.Node
	q.sorting.Range(func(node *lst.Node) bool {
		if matchTimerValue(node.Value, value) {
			target = node
			return false
		}
		return true
	})
	return target
}

func (q *delayingQueueImpl) HeapRange(fn func(value any, delay int64) bool) {
	q.lock.Lock()
	q.sorting.Range(func(n *lst.Node) bool {
		return fn(n.Value, n.Priority)
	})
	q.lock.Unlock()
}

// Len 返回延迟队列总元素数：堆中待到期项 + 已搬入内层队列的在队项之和。
func (q *delayingQueueImpl) Len() int {
	q.lock.Lock()
	count := int(q.sorting.Len())
	q.lock.Unlock()
	return count + q.Queue.Len()
}

// GetWithContext 阻塞直到消费到一个值、ctx 完成或队列关闭，永不返回
// ErrQueueIsEmpty。委托内层队列：堆中延迟项到期后由 scheduler 搬入内层
// （内层 Put 触发广播唤醒），等待内层即正确语义。
func (q *delayingQueueImpl) GetWithContext(ctx context.Context) (any, error) {
	return q.Queue.(BlockingGetQueue).GetWithContext(ctx)
}
