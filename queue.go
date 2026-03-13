package workqueue

import (
	"sync"
	"sync/atomic"

	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// queueImpl 是基础队列实现，支持可选幂等语义。
type queueImpl struct {
	closed      atomic.Bool
	once        sync.Once
	lock        sync.Mutex
	config      *QueueConfig
	list        container
	elementpool *lst.NodePool
	processing  Set
	dirty       Set
}

// NewQueue 创建基础队列。
func NewQueue(config *QueueConfig) Queue {
	return newQueue(&wrapInternalList{List: lst.New()}, lst.NewNodePool(), config)
}

func newQueue(list container, elementpool *lst.NodePool, config *QueueConfig) *queueImpl {

	q := &queueImpl{
		config:      isQueueConfigEffective(config),
		list:        list,
		elementpool: elementpool,
	}

	if q.config.idempotent {
		q.processing = q.config.setCreator()
		q.dirty = q.config.setCreator()
	}

	return q
}

func (q *queueImpl) Shutdown() {

	q.once.Do(func() {

		q.closed.Store(true)

		q.lock.Lock()

		q.list.Range(func(value interface{}) bool {
			q.elementpool.Put(value.(*lst.Node))
			return true
		})

		q.list.Cleanup()

		if q.config.idempotent {
			q.processing.Cleanup()
			q.dirty.Cleanup()
		}

		q.lock.Unlock()
	})
}

func (q *queueImpl) IsClosed() bool {
	return q.closed.Load()
}

func (q *queueImpl) Len() (count int) {

	q.lock.Lock()
	count = int(q.list.Len())
	q.lock.Unlock()
	return
}

func (q *queueImpl) Values() []interface{} {

	q.lock.Lock()
	items := q.list.Slice()
	q.lock.Unlock()
	return items
}

func (q *queueImpl) Range(fn func(interface{}) bool) {

	if fn == nil {
		return
	}

	q.lock.Lock()
	q.list.Range(func(value interface{}) bool {
		node := value.(*lst.Node)
		return fn(node.Value)
	})
	q.lock.Unlock()
}

func (q *queueImpl) Put(value interface{}) error {

	if q.IsClosed() {
		return ErrQueueIsClosed
	}

	if value == nil {
		return ErrElementIsNil
	}

	if q.config.idempotent {
		// 幂等模式先判重再分配节点，减少重复入队时的对象池开销。
		q.lock.Lock()
		if q.dirty.Contains(value) || q.processing.Contains(value) {
			q.lock.Unlock()
			return ErrElementAlreadyExist
		}
		last := q.elementpool.Get()
		last.Value = value
		q.list.Push(last)
		q.dirty.Add(value)
		q.lock.Unlock()
	} else {
		// 非幂等模式在锁外申请节点，缩短临界区。
		last := q.elementpool.Get()
		last.Value = value
		q.lock.Lock()
		q.list.Push(last)
		q.lock.Unlock()
	}

	q.config.callback.OnPut(value)
	return nil
}

func (q *queueImpl) Get() (interface{}, error) {

	if q.IsClosed() {
		return nil, ErrQueueIsClosed
	}

	q.lock.Lock()

	if q.list.Len() == 0 {
		q.lock.Unlock()
		return nil, ErrQueueIsEmpty
	}

	front := q.list.Pop().(*lst.Node)
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

func (q *queueImpl) Done(value interface{}) {

	if q.IsClosed() {
		return
	}

	if q.config.idempotent {
		q.lock.Lock()

		if !q.processing.Contains(value) {
			q.lock.Unlock()
			return
		}

		q.processing.Remove(value)
		q.lock.Unlock()

		q.config.callback.OnDone(value)
	}
}
