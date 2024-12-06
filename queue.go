package workqueue

import (
	"sync"
	"sync/atomic"

	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// queueImpl 结构体定义了一个队列的实现。
// The queueImpl struct defines an implementation of a queue.
type queueImpl struct {
	// closed 是一个原子布尔值，表示队列是否已关闭。
	// closed is an atomic boolean indicating whether the queue is closed.
	closed atomic.Bool

	// once 用于确保某些操作只执行一次。
	// once is used to ensure that some operations are performed only once.
	once sync.Once

	// lock 是用于保护队列的互斥锁。
	// lock is the mutex used to protect the queue.
	lock *sync.Mutex

	// config 是队列的配置。
	// config is the configuration of the queue.
	config *QueueConfig

	// list 是队列的元素列表。
	// list is the list of elements in the queue.
	list container

	// elementpool 是元素的内存池。
	// elementpool is the memory pool of elements.
	elementpool *lst.NodePool

	// processing 是正在处理的元素集合。
	// processing is the set of elements being processed.
	processing Set

	// dirty 是脏元素集合。
	// dirty is the set of dirty elements.
	dirty Set
}

// NewQueue 函数创建并返回一个新的 QueueImpl 实例。
// The NewQueue function creates and returns a new instance of QueueImpl.
func NewQueue(config *QueueConfig) Queue {
	return newQueue(&wrapInternalList{List: lst.New()}, lst.NewNodePool(), config)
}

// newQueue 函数创建并返回一个新的 QueueImpl 实例，它接受一个元素列表、一个元素内存池和一个队列配置作为参数。
// The newQueue function creates and returns a new instance of QueueImpl, it takes a list of elements, a memory pool of elements, and a queue configuration as parameters.
func newQueue(list container, elementpool *lst.NodePool, config *QueueConfig) *queueImpl {
	// 创建一个新的 QueueImpl 实例
	// Create a new instance of QueueImpl
	q := &queueImpl{
		// 检查队列配置是否有效，如果有效则使用，否则使用默认配置
		// Check if the queue configuration is effective, use it if it is, otherwise use the default configuration
		config: isQueueConfigEffective(config),

		// 设置队列的元素列表
		// Set the list of elements for the queue
		list: list,

		// 设置队列的元素内存池
		// Set the memory pool of elements for the queue
		elementpool: elementpool,

		// 初始化互斥锁，用于保护队列的并发操作
		// Initialize the mutex, used to protect the concurrent operations of the queue
		lock: &sync.Mutex{},
	}

	// 如果队列配置为幂等的，初始化正在处理的元素集合和脏元素集合
	// If the queue is configured as idempotent, initialize the set of elements being processed and the set of dirty elements
	if q.config.idempotent {
		// 初始化正在处理的元素集合
		// Initialize the set of elements being processed
		q.processing = q.config.setCreator()

		// 初始化脏元素集合
		// Initialize the set of dirty elements
		q.dirty = q.config.setCreator()
	}

	// 返回新创建的 Queue
	// Return the newly created Queue
	return q
}

// Shutdown 方法用于关闭队列，它会清空队列中的所有元素，并将它们放回元素内存池。
// The Shutdown method is used to shut down the queue, it will clear all the elements in the queue and put them back into the element memory pool.
func (q *queueImpl) Shutdown() {
	// 使用 once.Do 保证关闭操作只执行一次
	// Use once.Do to ensure that the shutdown operation is only performed once
	q.once.Do(func() {
		// 将 closed 设置为 true，表示队列已关闭
		// Set closed to true, indicating that the queue is closed
		q.closed.Store(true)

		// 加锁，保护队列的并发操作
		// Lock, to protect the concurrent operations of the queue
		q.lock.Lock()

		// 遍历队列中的所有元素，将它们放回元素内存池
		// Traverse all the elements in the queue and put them back into the element memory pool
		q.list.Range(func(value interface{}) bool {
			q.elementpool.Put(value.(*lst.Node))
			return true
		})

		// 清空队列
		// Clear the queue
		q.list.Cleanup()

		// 如果队列配置为幂等的，则清空正在处理的元素集合和脏元素集合
		// If the queue is configured as idempotent, clear the set of elements being processed and the set of dirty elements
		if q.config.idempotent {
			q.processing.Cleanup()
			q.dirty.Cleanup()
		}

		// 解锁
		// Unlock
		q.lock.Unlock()
	})
}

// IsClosed 方法用于检查队列是否已关闭。
// The IsClosed method is used to check if the queue is closed.
func (q *queueImpl) IsClosed() bool {
	// 返回 closed 的值
	// Return the value of closed
	return q.closed.Load()
}

// Len 方法返回队列的长度
// The Len method returns the length of the queue
func (q *queueImpl) Len() (count int) {
	// 加锁以保证线程安全
	// Lock to ensure thread safety
	q.lock.Lock()

	// 获取队列长度
	// Get the length of the queue
	count = int(q.list.Len())

	// 解锁
	// Unlock
	q.lock.Unlock()

	// 返回队列长度
	// Return the length of the queue
	return
}

// Values 方法返回队列中的所有元素
// The Values method returns all elements in the queue
func (q *queueImpl) Values() []interface{} {
	q.lock.Lock()
	items := q.list.Slice()
	q.lock.Unlock()
	return items
}

// Range 方法用于遍历队列中的所有元素。
// The Range method is used to traverse all elements in the queue.
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

// Put 方法用于将一个元素放入队列。
// The Put method is used to put an element into the queue.
func (q *queueImpl) Put(value interface{}) error {
	if q.IsClosed() {
		return ErrQueueIsClosed
	}
	if value == nil {
		return ErrElementIsNil
	}

	// 提前获取对象，减少加锁时间
	last := q.elementpool.Get()
	last.Value = value

	if q.config.idempotent {
		q.lock.Lock()
		if q.dirty.Contains(value) || q.processing.Contains(value) {
			q.elementpool.Put(last) // 记得归还对象
			q.lock.Unlock()
			return ErrElementAlreadyExist
		}
		q.lock.Unlock()
	}

	q.lock.Lock()
	q.list.Push(last)
	if q.config.idempotent {
		q.dirty.Add(value)
	}
	q.lock.Unlock()

	q.config.callback.OnPut(value)
	return nil
}

// Get 方法用于从队列中获取一个元素。
// The Get method is used to get an element from the queue.
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

// Done 方法用于标记队列中的一个元素已经处理完成。
// The Done method is used to mark an element in the queue as done.
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
