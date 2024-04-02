package workqueue

import (
	"sync"

	list "github.com/shengyanli1982/workqueue/internal/stl/deque"
)

// SimpleQ 结构体定义了一个简单队列数据结构
// The SimpleQ struct defines a simple queue data structure
type SimpleQ struct {
	// queue 是一个双端队列，用于存储队列中的元素
	// queue is a deque used to store the elements in the queue
	queue *list.Deque

	// qlock 是一个互斥锁，用于保护队列的并发访问
	// qlock is a mutex used to protect concurrent access to the queue
	qlock *sync.Mutex

	// plock 是一个互斥锁，用于保护 closed 标志的并发访问
	// plock is a mutex used to protect concurrent access to the closed flag
	plock *sync.Mutex

	// cond 是一个条件变量，用于实现队列的阻塞读
	// cond is a condition variable used to implement blocking read of the queue
	cond *sync.Cond

	// nodepool 是一个链表节点池，用于减少内存分配
	// nodepool is a list node pool used to reduce memory allocation
	nodepool *list.ListNodePool

	// once 是一个 sync.Once 对象，用于确保队列只被关闭一次
	// once is a sync.Once object used to ensure that the queue is closed only once
	once sync.Once

	// closed 是一个布尔值，表示队列是否已经被关闭
	// closed is a boolean indicating whether the queue has been closed
	closed bool

	// config 是一个指向 QConfig 结构体的指针，用于存储队列的配置信息
	// config is a pointer to a QConfig struct used to store the configuration information of the queue
	config *QConfig
}

// NewSimpleQueue 函数创建并返回一个新的 SimpleQueue 实例
// The NewSimpleQueue function creates and returns a new SimpleQ instance
func NewSimpleQueue(conf *QConfig) *SimpleQ {
	// 检查传入的配置是否有效，如果无效则返回默认配置
	// Check if the passed configuration is valid, if not, return the default configuration
	conf = isQConfigValid(conf)

	// 创建一个新的 SimpleQ 实例
	// Create a new SimpleQ instance
	q := &SimpleQ{
		// 创建一个新的双端队列
		// Create a new deque
		queue: list.NewDeque(),

		// 创建一个新的列表节点池
		// Create a new list node pool
		nodepool: list.NewListNodePool(),

		// 创建一个新的互斥锁，用于保护队列
		// Create a new mutex to protect the queue
		qlock: &sync.Mutex{},

		// 创建一个新的互斥锁，用于保护生产者
		// Create a new mutex to protect the producer
		plock: &sync.Mutex{},

		// 创建一个新的 sync.Once 实例，用于确保某些操作只执行一次
		// Create a new sync.Once instance to ensure that certain operations are only performed once
		once: sync.Once{},

		// 设置 closed 为 false，表示队列未关闭
		// Set closed to false, indicating that the queue is not closed
		closed: false,

		// 保存传入的配置
		// Save the passed configuration
		config: conf,
	}

	// 创建一个新的条件变量，用于等待和通知队列的状态改变
	// Create a new condition variable for waiting and notifying changes in the queue state
	q.cond = sync.NewCond(q.qlock)

	// 返回创建的 SimpleQ 实例
	// Return the created SimpleQ instance
	return q
}

// DefaultSimpleQueue 函数创建并返回一个新的具有默认配置的 SimpleQueue 实例
// The DefaultSimpleQueue function creates and returns a new SimpleQ instance with default configuration
func DefaultSimpleQueue() QInterface {
	// 调用 NewSimpleQueue 函数，传入 nil 作为配置，这将使用默认配置创建队列
	// Call the NewSimpleQueue function, passing nil as the configuration, this will create a queue with the default configuration
	return NewSimpleQueue(nil)
}

// Len 方法返回队列的长度
// The Len method returns the length of the queue
func (q *SimpleQ) Len() int {
	// 锁定队列锁，防止并发操作
	// Lock the queue lock to prevent concurrent operations
	q.qlock.Lock()

	// 使用 defer 语句在函数返回时解锁队列锁
	// Use the defer statement to unlock the queue lock when the function returns
	defer q.qlock.Unlock()

	// 返回队列的长度
	// Return the length of the queue
	return q.queue.Len()
}

// IsClosed 方法返回队列是否已经关闭
// The IsClosed method returns whether the queue is closed
func (q *SimpleQ) IsClosed() bool {
	// 锁定生产者锁，防止并发操作
	// Lock the producer lock to prevent concurrent operations
	q.plock.Lock()

	// 使用 defer 语句在函数返回时解锁生产者锁
	// Use the defer statement to unlock the producer lock when the function returns
	defer q.plock.Unlock()

	// 返回队列是否已经关闭
	// Return whether the queue is closed
	return q.closed
}

// Add 方法将一个元素添加到队列
// The Add method adds an element to the queue
func (q *SimpleQ) Add(element any) error {
	// 检查队列是否已关闭，如果已关闭则返回错误
	// Check if the queue is closed, if it is closed, return an error
	if q.IsClosed() {
		return ErrorQueueClosed
	}

	// 从节点池中获取一个节点
	// Get a node from the node pool
	node := q.nodepool.Get()

	// 设置节点的数据为传入的元素
	// Set the node's data to the passed element
	node.SetData(element)

	// 锁定条件变量的锁，防止并发操作
	// Lock the lock of the condition variable to prevent concurrent operations
	q.cond.L.Lock()

	// 将节点添加到队列
	// Add the node to the queue
	q.queue.Push(node)

	// 发送信号，通知等待的 goroutine
	// Send a signal to notify the waiting goroutines
	q.cond.Signal()

	// 解锁条件变量的锁
	// Unlock the lock of the condition variable
	q.cond.L.Unlock()

	// 调用回调函数，通知元素已添加
	// Call the callback function to notify that the element has been added
	q.config.callback.OnAdd(element)

	// 返回 nil，表示添加成功
	// Return nil, indicating that the addition was successful
	return nil
}

// Get 方法从队列中获取一个元素, 如果队列为空，不阻塞等待
// The Get method gets an element from the queue. If the queue is empty, it does not block and wait
func (q *SimpleQ) Get() (element any, err error) {
	// 检查队列是否已关闭，如果已关闭则返回错误
	// Check if the queue is closed, if it is closed, return an error
	if q.IsClosed() {
		return nil, ErrorQueueClosed
	}

	// 锁定队列锁，防止并发操作
	// Lock the queue lock to prevent concurrent operations
	q.qlock.Lock()

	// 从队列中弹出一个节点
	// Pop a node from the queue
	n := q.queue.Pop()

	// 解锁队列锁
	// Unlock the queue lock
	q.qlock.Unlock()

	// 如果节点为空，表示队列为空，返回错误
	// If the node is null, it means the queue is empty, return an error
	if n == nil {
		return nil, ErrorQueueEmpty
	}

	// 获取节点的数据
	// Get the data of the node
	element = n.Data()

	// 调用回调函数，通知元素已获取
	// Call the callback function to notify that the element has been obtained
	q.config.callback.OnGet(element)

	// 将节点放回节点池
	// Put the node back into the node pool
	q.nodepool.Put(n)

	// 返回获取的元素和 nil 错误，表示获取成功
	// Return the obtained element and nil error, indicating that the acquisition was successful
	return element, nil
}

// GetWithBlock 方法从队列中获取一个元素，如果队列为空，则阻塞等待
// The GetWithBlock method gets an element from the queue, if the queue is empty, it blocks and waits
func (q *SimpleQ) GetWithBlock() (element any, err error) {
	// 检查队列是否已关闭，如果已关闭则返回错误
	// Check if the queue is closed, if it is closed, return an error
	if q.IsClosed() {
		return nil, ErrorQueueClosed
	}

	// 锁定条件变量的锁，防止并发操作
	// Lock the lock of the condition variable to prevent concurrent operations
	q.cond.L.Lock()

	// 如果队列为空，则等待
	// If the queue is empty, wait
	for q.queue.Len() == 0 {
		q.cond.Wait()
	}

	// 从队列中弹出一个节点
	// Pop a node from the queue
	node := q.queue.Pop()

	// 解锁条件变量的锁
	// Unlock the lock of the condition variable
	q.cond.L.Unlock()

	// 如果节点为空，表示队列为空，返回错误
	// If the node is null, it means the queue is empty, return an error
	if node == nil {
		return nil, ErrorQueueEmpty
	}

	// 获取节点的数据
	// Get the data of the node
	element = node.Data()

	// 调用回调函数，通知元素已获取
	// Call the callback function to notify that the element has been obtained
	q.config.callback.OnGet(element)

	// 将节点放回节点池
	// Put the node back into the node pool
	q.nodepool.Put(node)

	// 返回获取的元素和 nil 错误，表示获取成功
	// Return the obtained element and nil error, indicating that the acquisition was successful
	return element, nil
}

// Done 方法表示一个元素已经处理完成
// The Done method indicates that an element has been processed
func (q *SimpleQ) Done(element any) {
	// 调用回调函数，通知元素已处理完成
	// Call the callback function to notify that the element has been processed
	q.config.callback.OnDone(element)
}

// Stop 方法用于关闭简单队列的操作
// The Stop method is used to stop the operations of the simple queue
func (q *SimpleQ) Stop() {
	// 使用 sync.Once 的 Do 方法确保以下操作只执行一次
	// Use the Do method of sync.Once to ensure that the following operations are only performed once
	q.once.Do(func() {
		// 锁定生产者锁，防止并发操作
		// Lock the producer lock to prevent concurrent operations
		q.plock.Lock()

		// 将 closed 设置为 true，表示队列已关闭
		// Set closed to true, indicating that the queue is closed
		q.closed = true

		// 解锁生产者锁
		// Unlock the producer lock
		q.plock.Unlock()

		// 锁定条件变量的锁，防止并发操作
		// Lock the lock of the condition variable to prevent concurrent operations
		q.cond.L.Lock()

		// 广播条件变量，唤醒所有等待的 goroutine
		// Broadcast the condition variable to wake up all waiting goroutines
		q.cond.Broadcast()

		// 重置队列
		// Reset the queue
		q.queue.Reset()

		// 解锁条件变量的锁
		// Unlock the lock of the condition variable
		q.cond.L.Unlock()
	})
}

// GetValues 方法返回队列中所有元素的快照
// The GetValues method returns a snapshot of all elements in the queue
func (q *SimpleQ) GetValues() []any {
	// 锁定队列锁，防止并发操作
	// Lock the queue lock to prevent concurrent operations
	q.qlock.Lock()

	// 使用 defer 语句在函数返回时解锁队列锁
	// Use the defer statement to unlock the queue lock when the function returns
	defer q.qlock.Unlock()

	// 返回队列中所有元素的快照
	// Return a snapshot of all elements in the queue
	return q.queue.SnapshotValues()
}

// Range 方法对队列中的每个元素执行给定的函数，如果函数返回 false，则停止迭代
// The Range method performs the given function for each element in the queue, if the function returns false, stop iterating
func (q *SimpleQ) Range(fn func(element any) bool) {
	// 锁定队列锁，防止并发操作
	// Lock the queue lock to prevent concurrent operations
	q.qlock.Lock()

	// 使用 defer 语句在函数返回时解锁队列锁
	// Use the defer statement to unlock the queue lock when the function returns
	defer q.qlock.Unlock()

	// 对队列中的每个元素执行给定的函数，如果函数返回 false，则停止迭代
	// Perform the given function for each element in the queue, if the function returns false, stop iterating
	q.queue.Range(func(n *list.Node) bool {
		return fn(n.Data())
	})
}
