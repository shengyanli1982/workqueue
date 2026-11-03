package workqueue

import (
	"context"
	"math"
	"sync/atomic"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// 预定义优先级，数值越小优先级越高。
const (
	PRIORITY_SLOWEST = math.MaxInt64
	PRIORITY_LOW     = math.MaxInt32
	PRIORITY_NORMAL  = 0
	PRIORITY_HIGH    = math.MinInt32
	PRIORITY_FASTEST = math.MinInt64
)

// priorityQueueImpl 使用红黑树按优先级排序。
type priorityQueueImpl struct {
	Queue
	config      *PriorityQueueConfig
	sorting     *hp.RBTree
	elementpool *lst.NodePool
	// draining 排空标记，置位后拒绝新的入队。
	draining atomic.Bool
}

// NewPriorityQueue 创建优先级队列。
//
// 不支持 WithValueIdempotent：PutWithPriority 绕过幂等簿记，且 Done 的重入队
// 推向 FIFO list 而非堆导致优先级丢失。构造期 fail-fast；幂等语义请使用其他队列。
func NewPriorityQueue(config *PriorityQueueConfig) PriorityQueue {

	config = isPriorityQueueConfigEffective(config)

	if config.idempotent {
		panic("workqueue: PriorityQueue does not support WithValueIdempotent: " +
			"PutWithPriority bypasses the inner idempotent bookkeeping and " +
			"Done re-enqueues into the FIFO list instead of the heap, so " +
			"priority would be lost; use Queue, RetryQueue, LeasedQueue, " +
			"DeadLetterQueue or TimerQueue for idempotent semantics")
	}

	q := &priorityQueueImpl{
		config:      config,
		sorting:     hp.New(),
		elementpool: lst.NewNodePool(),
	}

	q.Queue = newQueue(&wrapInternalHeap{RBTree: q.sorting}, q.elementpool, &config.QueueConfig)

	return q
}

// Shutdown 立即关停优先级队列：委托内层队列执行关停清理。
func (q *priorityQueueImpl) Shutdown() {
	q.Queue.Shutdown()
}

// ShutdownWithDrain 优雅关停：拒绝新入队，等待内层队列 drained 后关闭。
// 堆即内层存储，无独立搬运过程，drain 判定与基础队列一致。
func (q *priorityQueueImpl) ShutdownWithDrain(ctx context.Context) error {

	if q.IsClosed() {
		return nil
	}

	q.draining.Store(true)

	inner := q.Queue.(*queueImpl)

	err := waitForDrain(ctx, inner.IsClosed, inner.isDrained)

	q.Queue.Shutdown()

	return err
}

// IsDraining 返回队列是否处于排空阶段。
// 三态语义见 DrainingQueue 可选接口。
func (q *priorityQueueImpl) IsDraining() bool {
	return q.draining.Load()
}

// Put 以默认优先级（PRIORITY_NORMAL）入队，委托 PutWithPriority 实现。
func (q *priorityQueueImpl) Put(value any) error {
	return q.PutWithPriority(value, PRIORITY_NORMAL)
}

func (q *priorityQueueImpl) PutWithPriority(value any, priority int64) error {

	if q.IsClosed() || q.draining.Load() {
		return ErrQueueIsClosed
	}

	if value == nil {
		return ErrElementIsNil
	}

	last := q.elementpool.Get()
	last.Value = value
	last.Priority = priority

	qi := q.Queue.(*queueImpl)
	qi.lock.Lock()
	if q.IsClosed() {
		qi.lock.Unlock()
		q.elementpool.Put(last)
		return ErrQueueIsClosed
	}
	q.sorting.Push(last)
	// 唤醒阻塞在 GetWithContext 上的消费者。
	qi.notifyWaitersLocked()
	qi.lock.Unlock()

	q.config.callback.OnPriority(value, priority)

	return nil
}

// GetWithContext 阻塞直到消费到一个值、ctx 完成或队列关闭，永不返回
// ErrQueueIsEmpty。复用内层 queueImpl 的等待与唤醒机制。
func (q *priorityQueueImpl) GetWithContext(ctx context.Context) (any, error) {
	return q.Queue.(BlockingGetQueue).GetWithContext(ctx)
}

// HeapRange 持锁遍历堆中全部元素，fn 返回 false 时提前终止。
// fn 在队列锁内执行，禁止在 fn 中调用本队列任何方法（Put/PutWithPriority/Get 等），否则死锁。
func (q *priorityQueueImpl) HeapRange(fn func(value any, priority int64) bool) {
	qi := q.Queue.(*queueImpl)
	qi.lock.Lock()
	q.sorting.Range(func(node *lst.Node) bool {
		return fn(node.Value, node.Priority)
	})
	qi.lock.Unlock()
}
