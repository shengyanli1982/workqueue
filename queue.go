package workqueue

import (
	"sync"
	"sync/atomic"

	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
	"github.com/shengyanli1982/workqueue/v2/internal/container/set"
)

// QueueImpl 结构体定义了一个队列的实现。
// The QueueImpl struct defines an implementation of a queue.
type QueueImpl struct {
	// config 是队列的配置。
	// config is the configuration of the queue.
	config *QueueConfig

	// list 是队列的元素列表。
	// list is the list of elements in the queue.
	list *lst.List

	// elementpool 是元素的内存池。
	// elementpool is the memory pool of elements.
	elementpool *lst.NodePool

	// processing 是正在处理的元素集合。
	// processing is the set of elements being processed.
	processing, dirty *set.Set

	// lock 是用于保护队列的互斥锁。
	// lock is the mutex used to protect the queue.
	lock *sync.Mutex

	// once 用于确保某些操作只执行一次。
	// once is used to ensure that some operations are performed only once.
	once sync.Once

	// closed 是一个原子布尔值，表示队列是否已关闭。
	// closed is an atomic boolean indicating whether the queue is closed.
	closed atomic.Bool
}

// NewQueue 函数创建并返回一个新的 QueueImpl 实例。
// The NewQueue function creates and returns a new instance of QueueImpl.
func NewQueue(config *QueueConfig) Queue {
	return newQueue(lst.New(), lst.NewNodePool(), config)
}

// newQueue 函数创建并返回一个新的 QueueImpl 实例，它接受一个元素列表、一个元素内存池和一个队列配置作为参数。
// The newQueue function creates and returns a new instance of QueueImpl, it takes a list of elements, a memory pool of elements, and a queue configuration as parameters.
func newQueue(list *lst.List, elementpool *lst.NodePool, config *QueueConfig) *QueueImpl {
	// 返回一个新的 QueueImpl 实例
	// Return a new instance of QueueImpl
	return &QueueImpl{
		// 检查队列配置是否有效，如果有效则使用，否则使用默认配置
		// Check if the queue configuration is effective, use it if it is, otherwise use the default configuration
		config: isQueueConfigEffective(config),

		// 设置队列的元素列表
		// Set the list of elements for the queue
		list: list,

		// 设置队列的元素内存池
		// Set the memory pool of elements for the queue
		elementpool: elementpool,

		// 初始化正在处理的元素集合
		// Initialize the set of elements being processed
		processing: set.New(),

		// 初始化脏元素集合
		// Initialize the set of dirty elements
		dirty: set.New(),

		// 初始化互斥锁，用于保护队列的并发操作
		// Initialize the mutex, used to protect the concurrent operations of the queue
		lock: &sync.Mutex{},

		// 初始化一次性操作，用于保证队列的关闭操作只执行一次
		// Initialize the once operation, used to ensure that the close operation of the queue is only performed once
		once: sync.Once{},

		// 初始化原子布尔值，用于标记队列是否已关闭
		// Initialize the atomic boolean, used to mark whether the queue is closed
		closed: atomic.Bool{},
	}
}

// Shutdown 方法用于关闭队列，它会清空队列中的所有元素，并将它们放回元素内存池。
// The Shutdown method is used to shut down the queue, it will clear all the elements in the queue and put them back into the element memory pool.
func (q *QueueImpl) Shutdown() {
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
		q.list.Range(func(n *lst.Node) bool {
			q.elementpool.Put(n)
			return true
		})

		// 清空队列
		// Clear the queue
		q.list.Cleanup()

		// 如果队列配置为幂等的，则清空正在处理的元素集合和脏元素集合
		// If the queue is configured as idempotent, clear the set of elements being processed and the set of dirty elements
		if q.config.idempotent {
			q.processing.Clear()
			q.dirty.Clear()
		}

		// 解锁
		// Unlock
		q.lock.Unlock()
	})
}

// IsClosed 方法用于检查队列是否已关闭。
// The IsClosed method is used to check if the queue is closed.
func (q *QueueImpl) IsClosed() bool {
	// 返回 closed 的值
	// Return the value of closed
	return q.closed.Load()
}

// Len 方法用于获取队列的长度。
// The Len method is used to get the length of the queue.
func (q *QueueImpl) Len() int {
	// 加锁，保护队列的并发操作
	// Lock, to protect the concurrent operations of the queue
	q.lock.Lock()
	defer q.lock.Unlock()

	// 返回队列的长度
	// Return the length of the queue
	return int(q.list.Len())
}

// Values 方法用于获取队列中的所有元素。
// The Values method is used to get all the elements in the queue.
func (q *QueueImpl) Values() []interface{} {
	// 加锁，保护队列的并发操作
	// Lock, to protect the concurrent operations of the queue
	q.lock.Lock()
	defer q.lock.Unlock()

	// 返回队列中的所有元素
	// Return all the elements in the queue
	return q.list.Slice()
}

// Put 方法用于将一个元素放入队列。
// The Put method is used to put an element into the queue.
func (q *QueueImpl) Put(value interface{}) error {
	// 如果队列已关闭，返回错误
	// If the queue is closed, return an error
	if q.IsClosed() {
		return ErrQueueIsClosed
	}

	// 如果元素值为 nil，返回错误
	// If the element value is nil, return an error
	if value == nil {
		return ErrElementIsNil
	}

	// 如果队列配置为幂等的，并且元素已被标记为脏或正在处理，返回错误
	// If the queue is configured as idempotent, and the element has been marked as dirty or being processed, return an error
	if q.config.idempotent && q.isElementMarked(value) {
		return ErrElementAlreadyExist
	}

	// 从元素内存池中获取一个元素
	// Get an element from the element memory pool
	last := q.elementpool.Get()

	// 设置元素的值
	// Set the value of the element
	last.Value = value

	// 加锁，保护队列的并发操作
	// Lock, to protect the concurrent operations of the queue
	q.lock.Lock()

	// 将元素放入队列的后端
	// Put the element into the back of the queue
	q.list.PushBack(last)

	// 如果队列配置为幂等的，将元素添加到脏元素集合
	// If the queue is configured as idempotent, add the element to the set of dirty elements
	if q.config.idempotent {
		q.dirty.Add(value)
	}

	// 解锁
	// Unlock
	q.lock.Unlock()

	// 调用回调函数，通知元素已被放入
	// Call the callback function to notify that the element has been put
	q.config.callback.OnPut(value)

	// 返回 nil 错误
	// Return a nil error
	return nil
}

// Get 方法用于从队列中获取一个元素。
// The Get method is used to get an element from the queue.
func (q *QueueImpl) Get() (interface{}, error) {
	// 如果队列已关闭，返回错误
	// If the queue is closed, return an error
	if q.IsClosed() {
		return nil, ErrQueueIsClosed
	}

	// 加锁，保护队列的并发操作
	// Lock, to protect the concurrent operations of the queue
	q.lock.Lock()

	// 如果队列为空，解锁并返回错误
	// If the queue is empty, unlock and return an error
	if q.list.Len() == 0 {
		q.lock.Unlock()
		return nil, ErrQueueIsEmpty
	}

	// 从队列的前端弹出一个元素
	// Pop an element from the front of the queue
	front := q.list.PopFront()

	// 获取元素的值
	// Get the value of the element
	value := front.Value

	// 如果队列配置为幂等的，将元素添加到正在处理的元素集合，并从脏元素集合中移除
	// If the queue is configured as idempotent, add the element to the set of elements being processed and remove it from the set of dirty elements
	if q.config.idempotent {
		q.processing.Add(value)
		q.dirty.Remove(value)
	}

	// 解锁
	// Unlock
	q.lock.Unlock()

	// 将元素放回元素内存池
	// Put the element back into the element memory pool
	q.elementpool.Put(front)

	// 调用回调函数，通知元素已被获取
	// Call the callback function to notify that the element has been got
	q.config.callback.OnGet(value)

	// 返回元素的值和 nil 错误
	// Return the value of the element and a nil error
	return value, nil
}

// Done 方法用于标记队列中的一个元素已经处理完成。
// The Done method is used to mark an element in the queue as done.
func (q *QueueImpl) Done(value interface{}) {
	// 如果队列已关闭，直接返回
	// If the queue is closed, return directly
	if q.IsClosed() {
		return
	}

	// 如果队列配置为幂等的，从正在处理的元素集合中移除该元素 (锁保护)
	// If the queue is configured as idempotent, remove the element from the set of elements being processed (lock protection)
	if q.config.idempotent {
		q.lock.Lock()
		q.processing.Remove(value)
		q.lock.Unlock()
	}

	// 调用回调函数，通知元素已处理完成
	// Call the callback function to notify that the element is done
	q.config.callback.OnDone(value)
}

// isElementMarked 方法用于检查一个元素是否被标记为脏或正在处理。
// The isElementMarked method is used to check if an element is marked as dirty or being processed.
func (q *QueueImpl) isElementMarked(value interface{}) (result bool) {
	// 检查元素是否在脏元素集合或正在处理的元素集合中 (锁保护)
	// Check if the element is in the set of dirty elements or the set of elements being processed (lock protection)
	q.lock.Lock()
	result = q.dirty.Contains(value) || q.processing.Contains(value)
	q.lock.Unlock()

	// 返回检查结果
	// Return the check result
	return
}
