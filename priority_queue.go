package workqueue

import (
	"math"
	"sync"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

const MINI_PRIORITY = math.MinInt64

type PriorityQueueImpl struct {
	Queue
	config      *PriorityQueueConfig
	sorting     *hp.Heap
	elementpool *lst.NodePool
	lock        *sync.Mutex
}

func NewPriorityQueue(config *PriorityQueueConfig) PriorityQueue {
	config = isPriorityQueueConfigEffective(config)

	q := &PriorityQueueImpl{
		config:      config,
		sorting:     hp.New(),
		elementpool: lst.NewNodePool(),
	}

	q.Queue = newQueue(q.sorting.GetList(), q.elementpool, &config.QueueConfig)
	q.lock = q.Queue.(*QueueImpl).lock

	return q
}

func (q *PriorityQueueImpl) Shutdown() {
	q.Queue.Shutdown()
}

func (q *PriorityQueueImpl) PutWithPriority(value interface{}, priority int64) error {
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

	if priority > MINI_PRIORITY {
		q.config.callback.OnPriority(value, priority)
	} else {
		q.config.QueueConfig.callback.OnPut(value)
	}

	return nil
}

func (q *PriorityQueueImpl) Put(value interface{}) error {
	return q.PutWithPriority(value, MINI_PRIORITY)
}
