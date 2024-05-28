package workqueue

import (
	"sync"
	"sync/atomic"

	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

type QueueImpl struct {
	config      *QueueConfig
	list        *lst.List
	elementpool *lst.NodePool
	lock        *sync.Mutex
	once        sync.Once
	closed      atomic.Bool
}

func NewQueue(config *QueueConfig) Queue {
	return newQueue(lst.New(), lst.NewNodePool(), config)
}

func newQueue(list *lst.List, elementpool *lst.NodePool, config *QueueConfig) *QueueImpl {
	return &QueueImpl{
		config:      isQueueConfigEffective(config),
		list:        list,
		elementpool: elementpool,
		lock:        &sync.Mutex{},
		once:        sync.Once{},
		closed:      atomic.Bool{},
	}
}

func (q *QueueImpl) Shutdown() {
	q.once.Do(func() {
		q.closed.Store(true)
		q.lock.Lock()
		q.list.Range(func(n *lst.Node) bool {
			q.elementpool.Put(n)
			return true
		})
		q.list.Cleanup()
		q.lock.Unlock()
	})
}

func (q *QueueImpl) IsClosed() bool {
	return q.closed.Load()
}

func (q *QueueImpl) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return int(q.list.Len())
}

func (q *QueueImpl) Values() []interface{} {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.list.Slice()
}

func (q *QueueImpl) Put(value interface{}) error {
	if q.IsClosed() {
		return ErrQueueIsClosed
	}

	if value == nil {
		return ErrElementIsNil
	}

	last := q.elementpool.Get()
	last.Value = value

	q.lock.Lock()

	q.list.PushBack(last)

	q.lock.Unlock()

	q.config.callback.OnPut(value)

	return nil
}

func (q *QueueImpl) Get() (interface{}, error) {
	if q.IsClosed() {
		return nil, ErrQueueIsClosed
	}

	q.lock.Lock()

	if q.list.Len() == 0 {
		q.lock.Unlock()
		return nil, ErrQueueIsEmpty
	}

	front := q.list.PopFront()

	value := front.Value

	q.lock.Unlock()

	q.elementpool.Put(front)

	q.config.callback.OnGet(value)

	return value, nil
}

func (q *QueueImpl) Done(value interface{}) {
	if q.IsClosed() {
		return
	}

	q.config.callback.OnDone(value)
}
