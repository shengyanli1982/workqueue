package workqueue

import (
	"math"
	"sync"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

const (
	PRIORITY_SLOWEST = math.MaxInt64

	PRIORITY_LOW = math.MaxInt32

	PRIORITY_NORMAL = 0

	PRIORITY_HIGH = math.MinInt32

	PRIORITY_FASTEST = math.MinInt64
)

type priorityQueueImpl struct {
	Queue
	config      *PriorityQueueConfig
	sorting     *hp.RBTree
	elementpool *lst.NodePool
	lock        *sync.Mutex
}

func NewPriorityQueue(config *PriorityQueueConfig) PriorityQueue {
	config = isPriorityQueueConfigEffective(config)

	q := &priorityQueueImpl{
		config:      config,
		sorting:     hp.New(),
		elementpool: lst.NewNodePool(),
		lock:        &sync.Mutex{},
	}

	q.Queue = newQueue(&wrapInternalHeap{RBTree: q.sorting}, q.elementpool, &config.QueueConfig)

	return q
}

func (q *priorityQueueImpl) Shutdown() { q.Queue.Shutdown() }

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

	q.lock.Lock()

	q.sorting.Push(last)

	q.lock.Unlock()

	q.config.callback.OnPriority(value, priority)

	return nil
}

func (q *priorityQueueImpl) HeapRange(fn func(value interface{}, delay int64) bool) {
	q.lock.Lock()

	q.sorting.Range(func(node *lst.Node) bool {
		return fn(node.Value, node.Priority)
	})

	q.lock.Unlock()
}
