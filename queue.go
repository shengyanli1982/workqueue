package workqueue

import (
	"sync"
	"sync/atomic"

	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// queueImpl 实现了一个队列的具体实现
// queueImpl implements a concrete queue implementation
type queueImpl struct {
	closed      atomic.Bool   // 队列是否关闭的标志 / Flag indicating if queue is closed
	once        sync.Once     // 确保 Shutdown 只执行一次 / Ensures Shutdown is executed only once
	lock        *sync.Mutex   // 用于保护队列操作的互斥锁 / Mutex for protecting queue operations
	config      *QueueConfig  // 队列配置 / Queue configuration
	list        container     // 存储元素的容器 / Container for storing elements
	elementpool *lst.NodePool // 节点对象池 / Node object pool
	processing  Set           // 正在处理的元素集合 / Set of elements being processed
	dirty       Set           // 待处理的元素集合 / Set of elements to be processed
}

// NewQueue 创建一个新的队列实例
// NewQueue creates a new queue instance
func NewQueue(config *QueueConfig) Queue {
	return newQueue(&wrapInternalList{List: lst.New()}, lst.NewNodePool(), config)
}

// newQueue 初始化并返回一个新的队列实例
// newQueue initializes and returns a new queue instance
func newQueue(list container, elementpool *lst.NodePool, config *QueueConfig) *queueImpl {
	// 创建新的队列实例
	// Create new queue instance
	q := &queueImpl{
		config:      isQueueConfigEffective(config),
		list:        list,
		elementpool: elementpool,
		lock:        &sync.Mutex{},
	}

	// 如果启用了幂等性，初始化处理集合和脏数据集合
	// If idempotency is enabled, initialize processing and dirty sets
	if q.config.idempotent {
		q.processing = q.config.setCreator()
		q.dirty = q.config.setCreator()
	}

	return q
}

// Shutdown 关闭队列并清理资源
// Shutdown closes the queue and cleans up resources
func (q *queueImpl) Shutdown() {
	// 使用 sync.Once 确保 Shutdown 只执行一次
	// Use sync.Once to ensure Shutdown is executed only once
	q.once.Do(func() {
		// 将队列标记为已关闭状态
		// Mark the queue as closed
		q.closed.Store(true)

		// 加锁保护资源清理过程
		// Lock to protect resource cleanup process
		q.lock.Lock()

		// 遍历并清理所有节点，将节点返回到对象池
		// Iterate and clean up all nodes, return nodes to object pool
		q.list.Range(func(value interface{}) bool {
			q.elementpool.Put(value.(*lst.Node))
			return true
		})

		// 清理底层列表
		// Clean up the underlying list
		q.list.Cleanup()

		// 如果启用了幂等性，清理处理集合和脏数据集合
		// If idempotency is enabled, clean up processing and dirty sets
		if q.config.idempotent {
			q.processing.Cleanup()
			q.dirty.Cleanup()
		}

		q.lock.Unlock()
	})
}

// IsClosed 检查队列是否已关闭
// IsClosed checks if the queue is closed
func (q *queueImpl) IsClosed() bool {
	return q.closed.Load()
}

// Len 返回队列中的元素数量
// Len returns the number of elements in the queue
func (q *queueImpl) Len() (count int) {
	// 加锁保护并发访问
	// Lock to protect concurrent access
	q.lock.Lock()
	count = int(q.list.Len())
	q.lock.Unlock()
	return
}

// Values 返回队列中所有元素的切片
// Values returns a slice of all elements in the queue
func (q *queueImpl) Values() []interface{} {
	// 加锁保护并发访问
	// Lock to protect concurrent access
	q.lock.Lock()
	items := q.list.Slice()
	q.lock.Unlock()
	return items
}

// Range 遍历队列中的所有元素
// Range iterates over all elements in the queue
func (q *queueImpl) Range(fn func(interface{}) bool) {
	// 检查回调函数是否为空
	// Check if callback function is nil
	if fn == nil {
		return
	}

	// 加锁保护并发访问
	// Lock to protect concurrent access
	q.lock.Lock()
	q.list.Range(func(value interface{}) bool {
		node := value.(*lst.Node)
		return fn(node.Value)
	})
	q.lock.Unlock()
}

// Put 将元素添加到队列中
// Put adds an element to the queue
func (q *queueImpl) Put(value interface{}) error {
	// 检查队列是否已关闭
	// Check if queue is closed
	if q.IsClosed() {
		return ErrQueueIsClosed
	}

	// 检查元素是否为空
	// Check if element is nil
	if value == nil {
		return ErrElementIsNil
	}

	// 从对象池获取一个新节点
	// Get a new node from object pool
	last := q.elementpool.Get()
	last.Value = value

	// 如果启用了幂等性，检查元素是否已存在
	// If idempotency is enabled, check if element already exists
	if q.config.idempotent {
		q.lock.Lock()
		if q.dirty.Contains(value) || q.processing.Contains(value) {
			// 如果元素已存在，将节点返回到对象池
			// If element exists, return node to object pool
			q.elementpool.Put(last)
			q.lock.Unlock()
			return ErrElementAlreadyExist
		}
		q.lock.Unlock()
	}

	// 加锁并将元素添加到队列
	// Lock and add element to queue
	q.lock.Lock()
	q.list.Push(last)
	if q.config.idempotent {
		// 如果启用了幂等性，将元素添加到脏数据集合
		// If idempotency is enabled, add element to dirty set
		q.dirty.Add(value)
	}
	q.lock.Unlock()

	// 触发添加元素的回调函数
	// Trigger callback for element addition
	q.config.callback.OnPut(value)
	return nil
}

// Get 从队列中获取一个元素
// Get retrieves an element from the queue
func (q *queueImpl) Get() (interface{}, error) {
	// 检查队列是否已关闭
	// Check if queue is closed
	if q.IsClosed() {
		return nil, ErrQueueIsClosed
	}

	// 加锁保护并发访问
	// Lock to protect concurrent access
	q.lock.Lock()

	// 检查队列是否为空
	// Check if queue is empty
	if q.list.Len() == 0 {
		q.lock.Unlock()
		return nil, ErrQueueIsEmpty
	}

	// 从队列头部取出元素
	// Get element from queue head
	front := q.list.Pop().(*lst.Node)
	value := front.Value

	// 如果启用了幂等性，更新元素状态
	// If idempotency is enabled, update element status
	if q.config.idempotent {
		q.processing.Add(value) // 添加到处理集合 / Add to processing set
		q.dirty.Remove(value)   // 从脏数据集合中移除 / Remove from dirty set
	}
	q.lock.Unlock()

	// 将节点返回到对象池
	// Return node to object pool
	q.elementpool.Put(front)

	// 触发获取元素的回调函数
	// Trigger callback for element retrieval
	q.config.callback.OnGet(value)

	return value, nil
}

// Done 标记一个元素已经处理完成
// Done marks an element as processed
func (q *queueImpl) Done(value interface{}) {
	// 检查队列是否已关闭
	// Check if queue is closed
	if q.IsClosed() {
		return
	}

	// 如果启用了幂等性，处理元素完成状态
	// If idempotency is enabled, handle element completion status
	if q.config.idempotent {
		q.lock.Lock()
		// 检查元素是否在处理集合中
		// Check if element is in processing set
		if !q.processing.Contains(value) {
			q.lock.Unlock()
			return
		}
		// 从处理集合中移除元素
		// Remove element from processing set
		q.processing.Remove(value)
		q.lock.Unlock()

		// 触发元素处理完成的回调函数
		// Trigger callback for element completion
		q.config.callback.OnDone(value)
	}
}
