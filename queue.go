package workqueue

import (
	"sync"
	"sync/atomic"

	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

type QueueImpl struct {
	config      *QueueConfig
	elementlist *lst.List
	elementpool *lst.NodePool
	lock        sync.Mutex
	once        sync.Once
	closed      atomic.Bool
}

func NewQueue(config *QueueConfig) Queue {
	return &QueueImpl{
		config:      isQueueConfigEffective(config),
		elementlist: lst.New(),
		elementpool: lst.NewNodePool(),
		lock:        sync.Mutex{},
		once:        sync.Once{},
		closed:      atomic.Bool{},
	}
}

func (q *QueueImpl) Shutdown() {
	q.once.Do(func() {
		q.closed.Store(true)
		q.lock.Lock()
		q.elementlist.Range(func(n *lst.Node) bool {
			q.elementpool.Put(n)
			return true
		})
		q.elementlist.Cleanup()
		q.lock.Unlock()
	})
}

func (q *QueueImpl) IsClosed() bool {
	return q.closed.Load()
}

func (q *QueueImpl) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return int(q.elementlist.Len())
}

func (q *QueueImpl) Values() []interface{} {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.elementlist.Slice()
}

func (q *QueueImpl) Put(value interface{}) error {
	if q.IsClosed() {
		return ErrQueueIsClosed
	}

	if value == nil {
		return ErrElementIsNil
	}

	e := q.elementpool.Get()
	e.Value = value

	q.lock.Lock()

	q.elementlist.PushBack(e)

	q.config.callback.OnPut(value)

	q.lock.Unlock()

	return nil
}

func (q *QueueImpl) Get() (interface{}, error) {
	if q.IsClosed() {
		return nil, ErrQueueIsClosed
	}

	q.lock.Lock()

	e := q.elementlist.PopFront()
	if e == nil {
		q.lock.Unlock()
		return nil, ErrQueueIsEmpty
	}

	value := e.Value

	q.config.callback.OnGet(value)

	q.lock.Unlock()

	q.elementpool.Put(e)

	return value, nil
}

func (q *QueueImpl) Done(value interface{}) {
	if q.IsClosed() {
		return
	}

	q.lock.Lock()

	q.config.callback.OnDone(value)

	q.lock.Unlock()
}
