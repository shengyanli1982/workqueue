package workqueue

import (
	"math"

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
}

// NewPriorityQueue 创建优先级队列。
func NewPriorityQueue(config *PriorityQueueConfig) PriorityQueue {

	config = isPriorityQueueConfigEffective(config)

	q := &priorityQueueImpl{
		config:      config,
		sorting:     hp.New(),
		elementpool: lst.NewNodePool(),
	}

	q.Queue = newQueue(&wrapInternalHeap{RBTree: q.sorting}, q.elementpool, &config.QueueConfig)

	return q
}

func (q *priorityQueueImpl) Shutdown() {
	q.Queue.Shutdown()
}

func (q *priorityQueueImpl) Put(value interface{}) error {
	return q.PutWithPriority(value, PRIORITY_NORMAL)
}

func (q *priorityQueueImpl) PutWithPriority(value interface{}, priority int64) error {

	if q.IsClosed() {
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
	qi.lock.Unlock()

	q.config.callback.OnPriority(value, priority)

	return nil
}

func (q *priorityQueueImpl) HeapRange(fn func(value interface{}, priority int64) bool) {
	qi := q.Queue.(*queueImpl)
	qi.lock.Lock()
	q.sorting.Range(func(node *lst.Node) bool {
		return fn(node.Value, node.Priority)
	})
	qi.lock.Unlock()
}
