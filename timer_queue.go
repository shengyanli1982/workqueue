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

// timerQueueImpl 定时队列实现：通过红黑树按绝对时间戳排序，scheduler 协程逐个搬运
// 到期元素到内层队列。与 delaying_queue 的批量搬运不同，timer_queue 每次仅弹出堆顶
// 单个到期节点，适用于精确定时调度场景。
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
	// draining 在 drain 开始时置位：拒绝新的 Put/PutAt/PutAfter，
	// 已被接受的在途项由 scheduler 继续搬运直至 drain 判定完成。
	draining atomic.Bool
	// inTransit 记录“已出堆但尚未入内层队列”的搬运中节点数，
	// 与堆顶到期检查共同构成 drain 判定，避免搬运窗口内误判 drained。
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

// PutAt 按绝对时间点调度入队：到期后由 scheduler 搬入内层队列。
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

// PutAfter 按相对延迟调度入队：将 Duration 转为绝对时间戳后委托 putAtInternal。
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
		// 立即投递同样替换堆内同值待投递节点，保证同值至多一份调度。
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
	// 锁内复查 closed：防止锁外检查通过后 Shutdown 清空堆，导致节点被 Push 进死堆
	// 静默丢失（D3 同族，与 delaying_queue.go PutWithDelay 的锁内复查对齐）。
	// IsClosed 为 atomic load，持锁调用无重入风险。
	if q.IsClosed() {
		q.lock.Unlock()
		q.elementpool.Put(node)
		return ErrQueueIsClosed
	}
	// 插入前查重：同值重复调度命中旧节点时移除之并按新时间重排，
	// 确保堆内同值至多一个节点，否则 Cancel 只能取消其一，其余到期照常投递。
	old := q.findNodeLocked(value)
	if old != nil {
		q.sorting.Remove(old)
	}
	front := q.sorting.Front()
	q.sorting.Push(node)
	shouldWake := old != nil || front == nil || atMillis < front.Priority
	q.lock.Unlock()

	if old != nil {
		// 锁外归还旧节点，缩短临界区。
		q.elementpool.Put(old)
	}

	if shouldWake {
		q.notifyWake()
	}
	q.config.callback.OnSchedule(value, atMillis)
	return nil
}

// Cancel 取消堆中尚未搬运的单个调度项：优先检查堆顶（O(1)），未命中则线性查找。
// 命中返回 true 并唤醒 scheduler 重算 timer；未命中（未入堆、已搬运或已取消）返回 false。
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

// HeapRange 持锁遍历堆中全部调度项，fn 返回 false 时提前终止。fn 为 nil 时直接返回。
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
// ∧ 堆中无到期未搬运项 ∧ 无搬运中节点后关闭；超时或取消时强制关闭并返回 ctx.Err()。
// 堆中未到期调度项在最终关闭时丢弃（契约同 DelayingQueue 的 Q4a；
// TimerQueue 未设丢弃计数器，丢弃数量可从关停前 HeapRange 快照自行留存）。
func (q *timerQueueImpl) ShutdownWithDrain(ctx context.Context) error {

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

// drainedForShutdown 判定定时队列 drain 完成：堆中无到期未搬运项 ∧ 无搬运中
// 节点 ∧ 内层队列已 drained。持 q.lock 跨越内层判定：scheduler 的
// “pop(+inTransit)→Put”过渡经 q.lock 串行化，避免判定窗口内元素既不在堆
// 也不在内层队列视野内（TOCTOU）。未到期项不阻塞 drain。
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

	// 先收集所有节点，避免遍历中归还池会 Reset 指针破坏红黑树遍历。
	nodes := make([]*lst.Node, 0, q.sorting.Len())
	q.sorting.Range(func(node *lst.Node) bool {
		nodes = append(nodes, node)
		return true
	})

	q.sorting.Cleanup()

	q.lock.Unlock()

	// 锁外统一归还池，缩短临界区。
	for _, node := range nodes {
		q.elementpool.Put(node)
	}
}

// Len 返回定时队列总元素数：堆中待到期项 + 已搬入内层队列的在队项之和。
func (q *timerQueueImpl) Len() int {
	q.lock.Lock()
	count := int(q.sorting.Len())
	q.lock.Unlock()
	return count + q.Queue.Len()
}

// GetWithContext 阻塞直到消费到一个值、ctx 完成或队列关闭，永不返回
// ErrQueueIsEmpty。委托内层队列：调度项到期后由 scheduler 搬入内层
// （内层 Put 触发广播唤醒），等待内层即正确语义。
func (q *timerQueueImpl) GetWithContext(ctx context.Context) (any, error) {
	return q.Queue.(BlockingGetQueue).GetWithContext(ctx)
}

// scheduler 定时调度协程：每次从堆顶弹出单个到期节点搬入内层队列，
// 与 delaying_queue 的批量搬运（nextBatch）不同，此处逐个处理以保证精确调度。
// 空堆时仅等待 wake/closed，不设 timer；非空时按堆顶到期时间设置精确 timer。
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

	cb := q.config.callback // 局部变量捕获回调引用，避免数据竞争

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

// nextWait 检查堆顶：已到期则弹出节点并增加 inTransit 计数，未到期则返回等待时长。
// ok=false 表示队列已关闭，调度协程应退出。堆空时 wait=0、due=nil。
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

// wait 阻塞等待下一搬运时机：d<=0（堆空）时仅等待 wake/closed，不设 timer；
// d>0 时按堆顶到期时间 Reset timer，三路 select 等待 closed/wake/timer.C。
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

// notifyWake 非阻塞发送唤醒信号：wake 为容量 1 的缓冲 channel，多个并发唤醒自然折叠。
func (q *timerQueueImpl) notifyWake() {
	select {
	case q.wake <- struct{}{}:
	default:
	}
}

// findNodeLocked 持锁线性查找堆内首个与 value 匹配的节点，调用方须已持有 q.lock。
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

// matchTimerValue 类型安全的值匹配：对常见基本类型做显式类型断言比较，
// 避免 reflect.DeepEqual 的性能开销；仅 fallback 到 reflect 处理自定义类型。
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
