package workqueue

import (
	"math"
	"sync"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// 定义最小优先级常量，值为 math.MinInt64
// Define the minimum priority constant, the value is math.MinInt64
const MINI_PRIORITY = math.MinInt64

// sortedQueue 结构体定义了一个排序队列
// The sortedQueue struct defines a sorted queue
type sortedQueue struct {
	// Queue 是一个队列接口
	// Queue is a queue interface
	Queue

	// sorting 是一个堆结构，用于存储和排序队列元素
	// sorting is a heap structure for storing and sorting queue elements
	sorting *hp.Heap

	// elementpool 是一个节点池，用于存储队列元素，减少内存分配
	// elementpool is a node pool for storing queue elements, reducing memory allocation
	elementpool *lst.NodePool

	// lock 是一个互斥锁，用于保护队列操作的并发安全
	// lock is a mutex for protecting the concurrency safety of queue operations
	lock *sync.Mutex
}

// newSortedQueue 函数创建一个新的排序队列
// The newSortedQueue function creates a new sorted queue
func newSortedQueue(config *QueueConfig) *sortedQueue {
	// 创建一个新的排序队列实例
	// Create a new sorted queue instance
	q := &sortedQueue{
		sorting:     hp.New(),
		elementpool: lst.NewNodePool(),
	}

	// 初始化队列
	// Initialize the queue
	q.Queue = newQueue(q.sorting.GetList(), q.elementpool, config)

	// 设置互斥锁
	// Set the mutex
	q.lock = q.Queue.(*queueImpl).lock

	// 返回新创建的排序队列
	// Return the newly created sorted queue
	return q
}

// Shutdown 方法关闭队列
// The Shutdown method closes the queue
func (q *sortedQueue) Shutdown() { q.Queue.Shutdown() }

// putWithPriority 方法将元素放入队列，并设置优先级
// The putWithPriority method puts an element into the queue and sets its priority
func (q *sortedQueue) putWithPriority(value interface{}, priority int64, sortedPutCallbackFunc func(value interface{}, priority int64), putCallbackFunc func(value interface{})) error {
	// 如果 PriorityQueue 已关闭，返回错误
	// If the PriorityQueue is closed, return an error
	if q.IsClosed() {
		return ErrQueueIsClosed
	}

	// 如果元素值为 nil，返回错误
	// If the element value is nil, return an error
	if value == nil {
		return ErrElementIsNil
	}

	// 从元素内存池中获取一个元素
	// Get an element from the element memory pool
	last := q.elementpool.Get()

	// 设置元素的值
	// Set the value of the element
	last.Value = value

	// 设置元素的优先级
	// Set the priority of the element
	last.Priority = priority

	// 加锁，保护排序堆的并发操作
	// Lock, to protect the concurrent operations of the sorting heap
	q.lock.Lock()

	// 如果优先级大于最小优先级
	// If the priority is greater than the minimum priority
	if priority > MINI_PRIORITY {
		// 将元素放入排序堆
		// Put the element into the sorting heap
		q.sorting.Push(last)

		// 解锁，以保证其他线程或方法可以访问或修改队列
		// Unlock, to ensure other threads or methods can access or modify the queue
		q.lock.Unlock()

		// 调用回调函数，通知元素已被放入并设置了优先级
		// Call the callback function to notify that the element has been put in and the priority has been set
		sortedPutCallbackFunc(value, priority)
	} else {
		// 将元素放入列表中
		// Put the element into the list
		list := q.sorting.GetList()

		// 获取队列的 peek 元素，peek 元素是队列中优先级最高的元素
		// Get the peek element of the queue, the peek element is the element with the highest priority in the queue
		peek := q.Queue.(*queueImpl).peek

		// 如果列表为空，或者第一个元素的优先级大于当前元素的优先级
		// If the list is empty, or the priority of the first element is greater than the priority of the current element
		if peek == nil || peek.Priority > priority {
			// 将当前元素放在列表的最前面，因为它的优先级最高
			// Put the current element at the front of the list, because it has the highest priority
			list.PushFront(last)
		} else {
			// 否则，将当前元素插入到 peek 指向的元素之后
			// Otherwise, insert the current element after the element pointed to by peek
			list.InsertAfter(last, peek)
		}

		// 更新队列的 peek 指针
		// Update the peek pointer of the queue
		q.Queue.(*queueImpl).peek = last

		// 解锁，以保证其他线程或方法可以访问或修改队列
		// Unlock, to ensure other threads or methods can access or modify the queue
		q.lock.Unlock()

		// 调用回调函数，通知元素已被放入队列
		// Call the callback function to notify that the element has been put into the queue
		putCallbackFunc(value)
	}

	// 返回 nil 错误
	// Return a nil error
	return nil
}

// HeapRange 方法遍历队列中的所有元素，并对每个元素调用给定的函数
// The HeapRange method traverses all elements in the queue and calls the given function for each element
func (q *sortedQueue) HeapRange(fn func(value interface{}, priority int64) bool) {
	// 加锁以保证线程安全
	// Lock to ensure thread safety
	q.lock.Lock()

	// 遍历队列中的所有元素
	// Traverse all elements in the queue
	q.sorting.Range(func(n *lst.Node) bool {
		// 调用回调函数处理元素，传入元素值和优先级（这里的优先级被用作延迟）
		// Call the callback function to process the element, passing in the element value and priority (here the priority is used as delay)
		return fn(n.Value, n.Priority)
	})

	// 解锁，以保证其他线程或方法可以访问或修改队列
	// Unlock, to ensure other threads or methods can access or modify the queue
	q.lock.Unlock()
}

// priorityQueueImpl 结构体实现了 PriorityQueue 接口，它是一个支持优先级的队列。
// The priorityQueueImpl structure implements the PriorityQueue interface, it is a queue that supports priority.
type priorityQueueImpl struct {
	// sortedQueue 是一个排序队列，它是 priorityQueueImpl 的基础结构
	// sortedQueue is a sorted queue, it is the base structure of priorityQueueImpl
	sortedQueue

	// config 是 PriorityQueue 的配置，包括队列的大小、优先级等参数
	// config is the configuration of PriorityQueue, including parameters such as the size of the queue, priority, etc.
	config *PriorityQueueConfig
}

// NewPriorityQueue 函数用于创建一个新的 PriorityQueue。
// The NewPriorityQueue function is used to create a new PriorityQueue.
func NewPriorityQueue(config *PriorityQueueConfig) PriorityQueue {
	// 检查配置是否有效，如果无效，使用默认配置
	// Check if the configuration is valid, if not, use the default configuration
	config = isPriorityQueueConfigEffective(config)

	// 创建一个新的 PriorityQueueImpl
	// Create a new PriorityQueueImpl
	return &priorityQueueImpl{
		// 设置配置
		// Set the configuration
		config: config,

		// 创建一个新的排序队列
		// Create a new sorted queue
		sortedQueue: *newSortedQueue(&config.QueueConfig),
	}
}

// Put 方法用于将一个元素放入 PriorityQueue，元素的优先级为最小优先级。
// The Put method is used to put an element into the PriorityQueue, and the priority of the element is the minimum priority.
func (q *priorityQueueImpl) Put(value interface{}) error {
	// 调用 PutWithPriority 方法，将元素放入队列，并设置其优先级为最小优先级
	// Call the PutWithPriority method to put the element into the queue and set its priority to the minimum priority
	return q.PutWithPriority(value, MINI_PRIORITY)
}

// PutWithPriority 方法用于将一个元素放入 PriorityQueue，并设置其优先级。
// The PutWithPriority method is used to put an element into the PriorityQueue and set its priority.
func (q *priorityQueueImpl) PutWithPriority(value interface{}, priority int64) error {
	// 调用 sortingQueue 的 PutWithPriority 方法，将元素放入队列，并设置其优先级
	// Call the PutWithPriority method of sortingQueue to put the element into the queue and set its priority
	return q.sortedQueue.putWithPriority(value, priority, q.config.callback.OnPriority, q.config.callback.OnPut)
}
