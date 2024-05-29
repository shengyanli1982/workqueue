package workqueue

import (
	"sync"
	"sync/atomic"

	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
	"github.com/shengyanli1982/workqueue/v2/internal/container/set"
)

type QueueImpl struct {
	config            *QueueConfig
	list              *lst.List
	elementpool       *lst.NodePool
	processing, dirty *set.Set
	lock              *sync.Mutex
	once              sync.Once
	closed            atomic.Bool
}

func NewQueue(config *QueueConfig) Queue {
	return newQueue(lst.New(), lst.NewNodePool(), config)
}

func newQueue(list *lst.List, elementpool *lst.NodePool, config *QueueConfig) *QueueImpl {
	return &QueueImpl{
		config:      isQueueConfigEffective(config),
		list:        list,
		elementpool: elementpool,
		processing:  set.New(),
		dirty:       set.New(),
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
		q.processing.Clear()
		q.dirty.Clear()
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

	if q.config.idempotent && q.isElementMarked(value) {
		return ErrElementAlreadyExist
	}

	last := q.elementpool.Get()
	last.Value = value

	q.lock.Lock()

	q.list.PushBack(last)

	if q.config.idempotent {
		q.dirty.Add(value)
	}

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

	if q.config.idempotent {
		q.processing.Add(value)
		q.dirty.Remove(value)
	}

	q.lock.Unlock()

	q.elementpool.Put(front)

	q.config.callback.OnGet(value)

	return value, nil
}

func (q *QueueImpl) Done(value interface{}) {
	if q.IsClosed() {
		return
	}

	if q.config.idempotent {
		q.lock.Lock()
		q.processing.Remove(value)
		q.lock.Unlock()
	}

	q.config.callback.OnDone(value)
}

func (q *QueueImpl) isElementMarked(value interface{}) (result bool) {
	q.lock.Lock()
	result = q.dirty.Contains(value) || q.processing.Contains(value)
	q.lock.Unlock()
	return
}
