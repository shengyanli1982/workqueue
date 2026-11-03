package workqueue

import (
	"context"
	"errors"
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
	// discardedDelayed 关停时丢弃的未搬运延迟项累计计数。
	discardedDelayed atomic.Int64
	// draining 排空标记，置位后拒绝新入队。
	draining atomic.Bool
	// inTransit 已出堆但未入内层队列的搬运中节点数，参与 drain 判定。
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

// ShutdownWithDrain 优雅关停：拒绝新入队，等待内层 drained 且无搬运中节点后关闭。
// 堆中未到期项在最终关闭时丢弃（计入 discardedDelayed）。
func (q *delayingQueueImpl) ShutdownWithDrain(ctx context.Context) error {

	if q.IsClosed() {
		return nil
	}

	q.draining.Store(true)

	inner := q.Queue.(*queueImpl)

	err := waitForDrain(ctx, inner.IsClosed, func() bool {
		return q.drainedForShutdown(inner)
	})

	q.Queue.Shutdown()
	q.once.Do(q.closeNow)

	return err
}

// drainedForShutdown 判定 drain 完成：无搬运中节点、堆中无到期未搬运项、内层已 drained。
// 持 q.lock 跨越内层判定，避免 TOCTOU。未到期项不阻塞 drain。
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

// IsDraining 返回队列是否处于排空阶段。
// 三态语义见 DrainingQueue 可选接口。
func (q *delayingQueueImpl) IsDraining() bool {
	return q.draining.Load()
}

// closeNow 停止 scheduler 并丢弃堆中未搬运的延迟项。
func (q *delayingQueueImpl) closeNow() {
	close(q.closed)
	q.notifyWake()
	q.wg.Wait()

	q.lock.Lock()

	// 先收集再清理，避免遍历中归还池破坏红黑树。
	nodes := make([]*lst.Node, 0, q.sorting.Len())
	q.sorting.Range(func(node *lst.Node) bool {
		nodes = append(nodes, node)
		return true
	})

	q.sorting.Cleanup()

	q.lock.Unlock()

	for _, node := range nodes {
		q.elementpool.Put(node)
	}
	q.discardedDelayed.Add(int64(len(nodes)))
}

// Put 覆写嵌入 Queue：drain 期间拒绝新入队。
func (q *delayingQueueImpl) Put(value any) error {

	if q.draining.Load() {
		return ErrQueueIsClosed
	}

	return q.Queue.Put(value)
}

// PutWithDelay 按相对延迟（毫秒）将元素入延迟堆，到期后由 scheduler 搬入内层队列。
// 锁内复查 closed/draining 防止关停窗口丢失节点；仅当新项成为更早堆顶时唤醒 scheduler。
//
// delay ≤ 0 直通语义：元素已到期，绕过延迟堆直接入内层队列（等价于 Put）。
// 直通项不进堆，CancelDelay 对其恒返回 false；OnPut 由内层触发，OnDelay 不触发。
func (q *delayingQueueImpl) PutWithDelay(value any, delay int64) error {

	if q.IsClosed() || q.draining.Load() {
		return ErrQueueIsClosed
	}
	if value == nil {
		return ErrElementIsNil
	}

	if delay <= 0 {
		return q.Queue.Put(value)
	}

	last := q.elementpool.Get()
	last.Value = value
	last.Priority = toDelay(delay)

	q.lock.Lock()
	// 锁内复查：防止锁外检查通过后 Shutdown 清空堆或 drain 放行导致节点丢失。
	if q.IsClosed() || q.draining.Load() {
		q.lock.Unlock()
		q.elementpool.Put(last)
		return ErrQueueIsClosed
	}
	front := q.sorting.Front()
	q.sorting.Push(last)
	// 仅当堆由空变非空或新项成为更早堆顶时唤醒 scheduler，否则现有 timer 仍有效。
	shouldWake := front == nil || last.Priority < front.Priority
	q.lock.Unlock()

	if shouldWake {
		q.notifyWake()
	}
	q.config.callback.OnDelay(value, delay)
	return nil
}

// scheduler 以"堆顶精确 timer + 事件唤醒"搬运到期元素：
// 四路 select — timer 到期批量搬运、wake 重算堆顶、300ms 兜底心跳、closed 退出。
// 空堆时不设 timer，空闲零轮询。
func (q *delayingQueueImpl) scheduler() {
	defer q.wg.Done()

	heartbeat := time.NewTicker(time.Millisecond * 300)
	defer heartbeat.Stop()

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

	cb := q.config.callback
	var expired []*lst.Node

	for {
		nodes, wait, ok := q.nextBatch(expired[:0])
		if !ok {
			return
		}
		expired = nodes

		if len(nodes) > 0 {
			for i, node := range nodes {
				value := node.Value
				q.elementpool.Put(node)
				err := q.Queue.Put(value)
				q.inTransit.Add(-1)
				if err != nil {
					if errors.Is(err, ErrQueueIsClosed) {
						// 内层已关：收尾剩余节点并计入丢弃计数。
						rest := nodes[i+1:]
						for _, n := range rest {
							q.inTransit.Add(-1)
							q.elementpool.Put(n)
						}
						q.discardedDelayed.Add(int64(len(rest) + 1))
						return
					}
					cb.OnPullError(value, err)
				}
			}
			continue
		}

		if !q.wait(timer, heartbeat, wait) {
			return
		}
	}
}

// nextBatch 锁内弹出全部已到期节点（单批上限 128）并返回下一到期项的等待时长。
// ok=false 表示队列已关闭。堆空时 wait=0。
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

// wait 阻塞等待下一搬运时机：closed 退出、wake 重算堆顶、timer 到期、heartbeat 兜底。
// d<=0（堆空）时不设 timer，空闲零轮询。
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

// notifyWake 非阻塞发送唤醒信号，多个并发唤醒自然折叠为一个。
func (q *delayingQueueImpl) notifyWake() {
	select {
	case q.wake <- struct{}{}:
	default:
	}
}

// CancelDelay 取消堆中尚未搬运的延迟项：命中返回 true，未命中返回 false。
// value 匹配语义与 TimerQueue.Cancel 一致。
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

// findNodeLocked 持锁状态下按 matchTimerValue 语义查找堆内首个匹配节点。
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

// HeapRange 持锁遍历堆中全部未到期延迟项，fn 返回 false 时提前终止。
// fn 在队列锁内执行，禁止在 fn 中调用本队列任何方法，否则死锁。
func (q *delayingQueueImpl) HeapRange(fn func(value any, delay int64) bool) {
	q.lock.Lock()
	q.sorting.Range(func(n *lst.Node) bool {
		return fn(n.Value, n.Priority)
	})
	q.lock.Unlock()
}

// Len 返回延迟队列总元素数：堆中待到期项 + 内层队列在队项之和。
func (q *delayingQueueImpl) Len() int {
	q.lock.Lock()
	count := int(q.sorting.Len())
	q.lock.Unlock()
	return count + q.Queue.Len()
}

// DiscardedDelayed 返回关停时丢弃的未到期延迟项累计计数。原子读取，并发安全。
func (q *delayingQueueImpl) DiscardedDelayed() int64 {
	return q.discardedDelayed.Load()
}

// GetWithContext 阻塞直到消费到一个值、ctx 完成或队列关闭，永不返回
// ErrQueueIsEmpty。委托内层队列，延迟项到期后由 scheduler 搬入内层。
func (q *delayingQueueImpl) GetWithContext(ctx context.Context) (any, error) {
	return q.Queue.(BlockingGetQueue).GetWithContext(ctx)
}
