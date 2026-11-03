package workqueue

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// timerQueueImpl 定时队列：红黑树按绝对时间排序，scheduler 逐个搬运到期元素。
// 与 delaying_queue 的批量搬运不同，此处每次仅弹出堆顶单个节点，适用于精确调度。
type timerQueueImpl struct {
	Queue
	config      *TimerQueueConfig
	sorting     *hp.RBTree
	elementpool *lst.NodePool
	lock        sync.Mutex
	once        sync.Once
	wg          sync.WaitGroup
	wake        chan struct{}
	closed      chan struct{}
	// draining drain 开始时置位，拒绝新的 Put/PutAt/PutAfter，
	// 在途项由 scheduler 继续搬运直至 drain 完成。
	draining atomic.Bool
	// inTransit 记录"已出堆未入内层队列"的搬运中节点数，
	// 与堆顶到期检查共同构成 drain 判定。
	inTransit atomic.Int64
}

// NewTimerQueue 创建定时队列。
func NewTimerQueue(config *TimerQueueConfig) TimerQueue {
	config = isTimerQueueConfigEffective(config)

	q := &timerQueueImpl{
		config:      config,
		sorting:     hp.New(),
		elementpool: lst.NewNodePool(),
		wake:        make(chan struct{}, 1),
		closed:      make(chan struct{}),
	}

	q.Queue = newQueue(&wrapInternalList{List: lst.New()}, q.elementpool, &config.QueueConfig)
	q.wg.Add(1)
	go q.scheduler()

	return q
}

// PutAt 按绝对时间点调度：到期后搬入内层队列。
// 同值重复调度会替换旧节点，保证堆内同值至多一个。
func (q *timerQueueImpl) PutAt(value any, at time.Time) error {
	if q.IsClosed() || q.draining.Load() {
		return ErrQueueIsClosed
	}
	if value == nil {
		return ErrElementIsNil
	}
	return q.putAtInternal(value, at.UnixMilli(), time.Now().UnixMilli())
}

// PutAfter 按相对延迟调度：转为绝对时间戳后委托 putAtInternal。
func (q *timerQueueImpl) PutAfter(value any, after time.Duration) error {
	if q.IsClosed() || q.draining.Load() {
		return ErrQueueIsClosed
	}
	if value == nil {
		return ErrElementIsNil
	}
	nowMillis := time.Now().UnixMilli()
	return q.putAtInternal(value, nowMillis+after.Milliseconds(), nowMillis)
}

func (q *timerQueueImpl) putAtInternal(value any, atMillis, nowMillis int64) error {
	if atMillis <= nowMillis {
		q.lock.Lock()
		old := q.findNodeLocked(value)
		if old != nil {
			q.sorting.Remove(old)
		}
		q.lock.Unlock()

		if old != nil {
			q.elementpool.Put(old)
			q.notifyWake()
		}
		return q.Queue.Put(value)
	}

	node := q.elementpool.Get()
	node.Value = value
	node.Priority = atMillis

	q.lock.Lock()
	// 锁内复查 closed/draining：防止 Shutdown 清空堆后节点静默丢失。
	if q.IsClosed() || q.draining.Load() {
		q.lock.Unlock()
		q.elementpool.Put(node)
		return ErrQueueIsClosed
	}
	// 插入前查重：同值命中旧节点时移除，确保堆内同值至多一个。
	old := q.findNodeLocked(value)
	if old != nil {
		q.sorting.Remove(old)
	}
	front := q.sorting.Front()
	q.sorting.Push(node)
	shouldWake := old != nil || front == nil || atMillis < front.Priority
	q.lock.Unlock()

	if old != nil {
		q.elementpool.Put(old)
	}

	if shouldWake {
		q.notifyWake()
	}
	q.config.callback.OnSchedule(value, atMillis)
	return nil
}

// Cancel 取消堆中未搬运的调度项：优先检查堆顶，未命中则线性查找。
// 命中返回 true 并唤醒 scheduler；未命中返回 false。
func (q *timerQueueImpl) Cancel(value any) bool {
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

// HeapRange 持锁遍历堆中全部调度项，fn 返回 false 时提前终止。
// fn 在锁内执行，禁止调用本队列任何方法，否则死锁。
func (q *timerQueueImpl) HeapRange(fn func(value any, at int64) bool) {
	if fn == nil {
		return
	}

	q.lock.Lock()
	q.sorting.Range(func(node *lst.Node) bool {
		return fn(node.Value, node.Priority)
	})
	q.lock.Unlock()
}

func (q *timerQueueImpl) Shutdown() {
	q.Queue.Shutdown()

	q.once.Do(q.closeNow)
}

// ShutdownWithDrain 优雅关停：置位 draining 拒绝新入队，等待内层队列 drained
// ∧ 堆中无到期未搬运项 ∧ 无搬运中节点；超时或取消时强制关闭并返回 ctx.Err()。
// 未到期调度项在最终关闭时丢弃。
func (q *timerQueueImpl) ShutdownWithDrain(ctx context.Context) error {

	if q.IsClosed() {
		return nil
	}

	q.draining.Store(true)

	inner := q.Queue.(*queueImpl)

	err := waitForDrain(ctx, inner.IsClosed, func() bool {
		return q.drainedForShutdown(inner)
	})

	// 与 Shutdown 一致的关闭顺序。
	q.Queue.Shutdown()
	q.once.Do(q.closeNow)

	return err
}

// drainedForShutdown 判定 drain 完成：堆中无到期未搬运项 ∧ 无搬运中节点
// ∧ 内层队列已 drained。持锁跨越内层判定，避免 TOCTOU。
func (q *timerQueueImpl) drainedForShutdown(inner *queueImpl) bool {

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

// IsDraining 返回是否处于排空阶段。三态语义见 DrainingQueue 接口。
func (q *timerQueueImpl) IsDraining() bool {
	return q.draining.Load()
}

// Put 覆写嵌入 Queue：drain 期间拒绝新入队。
func (q *timerQueueImpl) Put(value any) error {
	if q.draining.Load() {
		return ErrQueueIsClosed
	}

	return q.Queue.Put(value)
}

// closeNow 停止 scheduler 并清空堆中未搬运的调度项。
func (q *timerQueueImpl) closeNow() {
	close(q.closed)
	q.notifyWake()
	q.wg.Wait()

	q.lock.Lock()
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
}

// Len 返回总元素数：堆中待到期项 + 内层队列在队项之和。
func (q *timerQueueImpl) Len() int {
	q.lock.Lock()
	count := int(q.sorting.Len())
	q.lock.Unlock()
	return count + q.Queue.Len()
}

// GetWithContext 阻塞直到消费到值、ctx 完成或队列关闭，永不返回 ErrQueueIsEmpty。
func (q *timerQueueImpl) GetWithContext(ctx context.Context) (any, error) {
	return q.Queue.(BlockingGetQueue).GetWithContext(ctx)
}

// scheduler 定时调度协程：逐个弹出到期节点搬入内层队列，
// 空堆时仅等待 wake/closed，非空时按堆顶设置精确 timer。
func (q *timerQueueImpl) scheduler() {
	defer q.wg.Done()

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

	for {
		wait, due, ok := q.nextWait()
		if !ok {
			return
		}

		if due != nil {
			value := due.Value
			q.elementpool.Put(due)
			err := q.Queue.Put(value)
			q.inTransit.Add(-1)
			if err != nil {
				if errors.Is(err, ErrQueueIsClosed) {
					return
				}
				cb.OnScheduleError(value, err)
			}
			continue
		}

		if !q.wait(timer, wait) {
			return
		}
	}
}

// nextWait 检查堆顶：已到期则弹出节点并增加 inTransit，未到期则返回等待时长。
// ok=false 表示队列已关闭；堆空时 wait=0、due=nil。
func (q *timerQueueImpl) nextWait() (time.Duration, *lst.Node, bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	select {
	case <-q.closed:
		return 0, nil, false
	default:
	}

	front := q.sorting.Front()
	if front == nil {
		return 0, nil, true
	}

	now := time.Now().UnixMilli()
	if front.Priority <= now {
		q.inTransit.Add(1)
		return 0, q.sorting.Pop(), true
	}

	return time.Duration(front.Priority-now) * time.Millisecond, nil, true
}

// wait 阻塞等待下一搬运时机：d<=0（堆空）时仅等待 wake/closed；
// d>0 时按堆顶时间 Reset timer，三路 select。
func (q *timerQueueImpl) wait(timer *time.Timer, d time.Duration) bool {
	if d <= 0 {
		select {
		case <-q.closed:
			return false
		case <-q.wake:
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
	}
}

// notifyWake 非阻塞发送唤醒信号，容量 1 的缓冲 channel 自然折叠并发唤醒。
func (q *timerQueueImpl) notifyWake() {
	select {
	case q.wake <- struct{}{}:
	default:
	}
}

// findNodeLocked 持锁线性查找匹配节点，调用方须已持有 q.lock。
func (q *timerQueueImpl) findNodeLocked(value any) *lst.Node {
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

// matchTimerValue 类型安全的值匹配：基本类型做显式断言避免 reflect 开销，
// 仅 fallback 到 reflect.DeepEqual 处理自定义类型。
func matchTimerValue(candidate, target any) bool {
	switch tv := target.(type) {
	case int:
		cv, ok := candidate.(int)
		return ok && cv == tv
	case int8:
		cv, ok := candidate.(int8)
		return ok && cv == tv
	case int16:
		cv, ok := candidate.(int16)
		return ok && cv == tv
	case int32:
		cv, ok := candidate.(int32)
		return ok && cv == tv
	case int64:
		cv, ok := candidate.(int64)
		return ok && cv == tv
	case uint:
		cv, ok := candidate.(uint)
		return ok && cv == tv
	case uint8:
		cv, ok := candidate.(uint8)
		return ok && cv == tv
	case uint16:
		cv, ok := candidate.(uint16)
		return ok && cv == tv
	case uint32:
		cv, ok := candidate.(uint32)
		return ok && cv == tv
	case uint64:
		cv, ok := candidate.(uint64)
		return ok && cv == tv
	case uintptr:
		cv, ok := candidate.(uintptr)
		return ok && cv == tv
	case string:
		cv, ok := candidate.(string)
		return ok && cv == tv
	case float32:
		cv, ok := candidate.(float32)
		return ok && cv == tv
	case float64:
		cv, ok := candidate.(float64)
		return ok && cv == tv
	default:
		return reflect.DeepEqual(candidate, target)
	}
}
