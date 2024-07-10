package workqueue

import (
	"math"
	"sync"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

const (
	// PRIORITY_SLOWEST 定义了最慢的优先级，值为最大的 int64
	// PRIORITY_SLOWEST defines the slowest priority, which is the maximum int64
	PRIORITY_SLOWEST = math.MaxInt64

	// PRIORITY_LOW 定义了低优先级，值为最大的 int32
	// PRIORITY_LOW defines low priority, which is the maximum int32
	PRIORITY_LOW = math.MaxInt32

	// PRIORITY_NORMAL 定义了正常优先级，值为 0
	// PRIORITY_NORMAL defines normal priority, which is 0
	PRIORITY_NORMAL = 0

	// PRIORITY_HIGH 定义了高优先级，值为最小的 int32
	// PRIORITY_HIGH defines high priority, which is the minimum int32
	PRIORITY_HIGH = math.MinInt32

	// PRIORITY_FASTEST 定义了最快的优先级，值为最小的 int64
	// PRIORITY_FASTEST defines the fastest priority, which is the minimum int64
	PRIORITY_FASTEST = math.MinInt64
)

// priorityQueueImpl 结构体，实现了 PriorityQueue 接口
// The priorityQueueImpl struct, which implements the PriorityQueue interface
type priorityQueueImpl struct {
	// Queue 是一个队列接口
	// Queue is a queue interface
	Queue

	// config 是 PriorityQueue 的配置
	// config is the configuration of PriorityQueue
	config *PriorityQueueConfig

	// sorting 是一个堆结构，用于存储和排序红黑树
	// sorting is a heap structure for storing and sorting red-black trees
	sorting *hp.RBTree

	// elementpool 是一个节点池，用于存储元素，减少内存分配
	// elementpool is a node pool for storing elements, reducing memory allocation
	elementpool *lst.NodePool

	// lock 是一个互斥锁，用于保护队列操作的并发安全
	// lock is a mutex for protecting the concurrency safety of queue operations
	lock *sync.Mutex
}

// NewPriorityQueue 函数用于创建一个新的 PriorityQueue
// The NewPriorityQueue function is used to create a new PriorityQueue
func NewPriorityQueue(config *PriorityQueueConfig) PriorityQueue {
	// 检查配置是否有效，如果无效，使用默认配置
	// Check if the configuration is valid, if not, use the default configuration
	config = isPriorityQueueConfigEffective(config)

	// 创建一个新的 PriorityQueueImpl
	// Create a new PriorityQueueImpl
	q := &priorityQueueImpl{
		// 设置配置
		// Set the configuration
		config: config,

		// 创建一个新的排序堆，用于存储延迟元素
		// Create a new sorting heap for storing delayed elements
		sorting: hp.New(),

		// 创建一个新的元素内存池，用于存储队列元素，减少内存分配
		// Create a new element memory pool for storing queue elements, reducing memory allocation
		elementpool: lst.NewNodePool(),
	}

	// 使用 newQueue 创建一个新的队列，并将其赋值给 q.Queue
	// Use newQueue to create a new queue, and assign it to q.Queue
	q.Queue = newQueue(&wrapInternalHeap{RBTree: q.sorting}, q.elementpool, &config.QueueConfig)

	// 将 q.Queue 的锁赋值给 q.lock
	// Assign the lock of q.Queue to q.lock
	q.lock = q.Queue.(*queueImpl).lock

	// 返回新创建的 PriorityQueue
	// Return the newly created PriorityQueue
	return q
}

// Shutdown 方法用于关闭 PriorityQueue。
// The Shutdown method is used to shut down the PriorityQueue.
func (q *priorityQueueImpl) Shutdown() { q.Queue.Shutdown() }

// Put 方法用于将一个元素放入 PriorityQueue，元素的优先级为最小优先级。
// The Put method is used to put an element into the PriorityQueue, and the priority of the element is the minimum priority.
func (q *priorityQueueImpl) Put(value interface{}) error {
	return q.PutWithPriority(value, PRIORITY_NORMAL)
}

// PutWithPriority 方法用于将一个元素放入 PriorityQueue，并设置其优先级。
// The PutWithPriority method is used to put an element into the PriorityQueue and set its priority.
func (q *priorityQueueImpl) PutWithPriority(value interface{}, priority int64) error {
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

	// 将元素放入排序堆
	// Put the element into the sorting heap
	q.sorting.Push(last)

	// 解锁
	// Unlock
	q.lock.Unlock()

	// 调用回调函数，通知已经按照 priority 添加了 value 值的元素
	// Call the callback function to notify that an element with the value of value has been added according to priority
	q.config.callback.OnPriority(value, priority)

	// 返回 nil 错误
	// Return a nil error
	return nil
}

// HeapRange 方法用于遍历 sorted 堆中的所有元素。
// The HeapRange method is used to traverse all elements in the sorted heap.
func (q *priorityQueueImpl) HeapRange(fn func(value interface{}, delay int64) bool) {
	// 加锁以保证线程安全
	// Lock to ensure thread safety
	q.lock.Lock()

	// 遍历队列中的所有元素
	// Traverse all elements in the queue
	q.sorting.Range(func(node *lst.Node) bool {
		// 调用回调函数处理元素
		// Call the callback function to process the element
		return fn(node.Value, node.Priority)
	})

	// 解锁
	// Unlock
	q.lock.Unlock()
}
