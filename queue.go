package workqueue

import (
	"sync"

	list "github.com/shengyanli1982/workqueue/internal/stl/deque"
	"github.com/shengyanli1982/workqueue/internal/stl/set"
)

// QInterface 定义了队列的方法接口
// QInterface defines the method interface of the queue
type QInterface interface {
	// Add 方法添加一个元素到队列
	// The Add method adds an element to the queue
	Add(element any) error

	// Len 方法返回队列中元素的数量
	// The Len method returns the number of elements in the queue
	Len() int

	// Range 方法遍历队列中的每个元素并调用提供的函数。如果函数返回 false，迭代停止
	// The Range method iterates over each element in the queue and calls the provided function. If the function returns false, the iteration stops
	Range(fn func(element any) bool)

	// Get 方法从队列中获取一个元素
	// The Get method retrieves an element from the queue
	Get() (element any, err error)

	// GetWithBlock 方法从队列中获取一个元素，如果队列为空，会阻塞等待
	// The GetWithBlock method retrieves an element from the queue, blocking if the queue is empty
	GetWithBlock() (element any, err error)

	// GetValues 方法返回队列中的所有元素
	// The GetValues method returns all elements in the queue
	GetValues() []any

	// Done 方法标记一个元素已经处理完成并从队列中移除
	// The Done method marks an element as processed and removes it from the queue
	Done(element any)

	// Stop 方法停止队列并释放任何资源
	// The Stop method stops the queue and releases any resources
	Stop()

	// IsClosed 方法返回队列是否已经关闭
	// The IsClosed method returns whether the queue is closed
	IsClosed() bool
}

// QCallback 定义了队列的回调接口
// QCallback defines the callback interface of the queue
type QCallback interface {
	// OnAdd 是添加元素后的回调，参数 any 是添加的元素
	// OnAdd is the callback after adding an element, the parameter any is the added element
	OnAdd(any)

	// OnGet 是获取元素后的回调，参数 any 是获取的元素
	// OnGet is the callback after getting an element, the parameter any is the gotten element
	OnGet(any)

	// OnDone 是处理完元素后的回调，参数 any 是处理完的元素
	// OnDone is the callback after an element is processed, the parameter any is the processed element
	OnDone(any)
}

// QConfig 结构体定义了队列的配置信息
// The QConfig struct defines the configuration information of the queue
type QConfig struct {
	// callback 是一个队列回调接口，用于实现队列元素的处理
	// callback is a queue callback interface, used to implement the processing of queue elements
	callback QCallback
}

// NewQConfig 方法创建一个新的队列配置
// The NewQConfig method creates a new queue configuration
func NewQConfig() *QConfig {
	return &QConfig{}
}

// WithCallback 方法设置队列的回调接口
// The WithCallback method sets the callback interface of the queue
func (c *QConfig) WithCallback(cb QCallback) *QConfig {
	c.callback = cb
	return c
}

// isQConfigValid 方法验证队列配置是否有效
// The isQConfigValid method verifies whether the queue configuration is valid
func isQConfigValid(conf *QConfig) *QConfig {
	if conf == nil {
		conf = NewQConfig().WithCallback(newEmptyCallback())
	} else {
		if conf.callback == nil {
			conf.callback = newEmptyCallback()
		}
	}
	return conf
}

// Q 结构体定义了队列的数据结构
// The Q struct defines the data structure of the queue
type Q struct {
	// queue 是一个双端队列，用于存储队列中的元素
	// queue is a deque used to store the elements in the queue
	queue *list.Deque

	// nodepool 是一个链表节点池，用于存储链表节点
	// nodepool is a list node pool used to store list nodes
	nodepool *list.ListNodePool

	// qlock 是一个互斥锁，用于保护队列的并发访问
	// qlock is a mutex used to protect concurrent access to the queue
	qlock *sync.Mutex

	// cond 是一个条件变量，用于等待或通知队列状态的改变
	// cond is a condition variable used to wait for or signal changes in the queue state
	cond *sync.Cond

	// dirty 是一个集合，用于存储被修改过的元素
	// dirty is a set used to store modified elements
	dirty set.Set

	// processing 是一个集合，用于存储正在处理的元素
	// processing is a set used to store elements that are being processed
	processing set.Set

	// plock 是一个互斥锁，用于保护 processing 集合的并发访问
	// plock is a mutex used to protect concurrent access to the processing set
	plock *sync.Mutex

	// once 是一个 sync.Once 对象，用于确保某个操作只执行一次
	// once is a sync.Once object used to ensure that an operation is performed only once
	once sync.Once

	// closed 是一个布尔值，表示队列是否已经被关闭
	// closed is a boolean indicating whether the queue has been closed
	closed bool

	// config 是一个指向 QConfig 结构体的指针，用于存储队列的配置信息
	// config is a pointer to a QConfig struct used to store the configuration information of the queue
	config *QConfig
}

// NewQueue 函数创建一个新的 Queue 对象
// The NewQueue function creates a new Queue object
func NewQueue(conf *QConfig) *Q {
	// 检查传入的配置是否有效，如果无效则返回默认配置
	// Check if the passed configuration is valid, if not, return the default configuration
	conf = isQConfigValid(conf)

	// 创建一个新的 Queue 对象
	// Create a new Queue object
	q := &Q{
		// dirty 是一个集合，用于存储“脏”元素，即需要处理的元素
		// dirty is a set used to store "dirty" elements, i.e., elements that need to be processed
		dirty: set.NewSet(),

		// processing 是一个集合，用于存储正在处理的元素
		// processing is a set used to store elements that are being processed
		processing: set.NewSet(),

		// queue 是一个双端队列，用于存储队列中的元素
		// queue is a deque used to store elements in the queue
		queue: list.NewDeque(),

		// qlock 是一个互斥锁，用于保护队列的并发操作
		// qlock is a mutex used to protect concurrent operations on the queue
		qlock: &sync.Mutex{},

		// nodepool 是一个节点池，用于存储队列节点
		// nodepool is a node pool used to store queue nodes
		nodepool: list.NewListNodePool(),

		// plock 是一个互斥锁，用于保护 processing 集合的并发操作
		// plock is a mutex used to protect concurrent operations on the processing set
		plock: &sync.Mutex{},

		// once 是一个 sync.Once 对象，用于确保某个操作只执行一次
		// once is a sync.Once object used to ensure that an operation is performed only once
		once: sync.Once{},

		// closed 是一个布尔值，表示队列是否已关闭
		// closed is a boolean value indicating whether the queue is closed
		closed: false,

		// config 是队列的配置对象
		// config is the configuration object of the queue
		config: conf,
	}

	// 创建一个新的条件变量，用于等待和通知队列的状态变化
	// Create a new condition variable for waiting and notifying changes in the queue state
	q.cond = sync.NewCond(q.qlock)

	// 返回创建的 Queue 对象
	// Return the created Queue object
	return q
}

// DefaultQueue 函数创建一个默认的 Queue 对象
// The DefaultQueue function creates a default Queue object
func DefaultQueue() QInterface {
	// 创建一个新的 Queue 对象，配置为 nil
	// Create a new Queue object with nil configuration
	return NewQueue(nil)
}

// todo 方法将元素从 dirty 集合移动到 processing 集合
// The todo method moves the element from the dirty set to the processing set
func (q *Q) todo(element any) {
	// 锁定 plock 以保护并发操作
	// Lock plock to protect concurrent operations
	q.plock.Lock()

	// 从 dirty 集合中删除元素
	// Remove the element from the dirty set
	q.dirty.Delete(element)

	// 将元素添加到 processing 集合
	// Add the element to the processing set
	q.processing.Add(element)

	// 解锁 plock
	// Unlock plock
	q.plock.Unlock()
}

// prepare 方法将元素添加到 dirty 集合
// The prepare method adds the element to the dirty set
func (q *Q) prepare(element any) {
	// 锁定 plock 以保护并发操作
	// Lock plock to protect concurrent operations
	q.plock.Lock()

	// 将元素添加到 dirty 集合
	// Add the element to the dirty set
	q.dirty.Add(element)

	// 解锁 plock
	// Unlock plock
	q.plock.Unlock()
}

// isElementMarked 方法检查元素是否在 dirty 或 processing 集合中
// The isElementMarked method checks whether the element is in the dirty or processing set
func (q *Q) isElementMarked(element any) bool {
	// 锁定 plock 以保护并发操作
	// Lock plock to protect concurrent operations
	q.plock.Lock()

	// 使用 defer 语句在函数返回时解锁 plock
	// Use the defer statement to unlock plock when the function returns
	defer q.plock.Unlock()

	// 检查元素是否在 dirty 或 processing 集合中
	// Check if the element is in the dirty or processing set
	return q.dirty.Has(element) || q.processing.Has(element)
}

// Len 方法返回队列的长度
// The Len method returns the length of the queue
func (q *Q) Len() int {
	// 锁定 qlock 以保护并发操作
	// Lock qlock to protect concurrent operations
	q.qlock.Lock()

	// 使用 defer 语句在函数返回时解锁 qlock
	// Use the defer statement to unlock qlock when the function returns
	defer q.qlock.Unlock()

	// 返回队列的长度
	// Return the length of the queue
	return q.queue.Len()
}

// IsClosed 方法检查队列是否已关闭
// The IsClosed method checks whether the queue is closed
func (q *Q) IsClosed() bool {
	// 锁定 plock 以保护并发操作
	// Lock plock to protect concurrent operations
	q.plock.Lock()

	// 使用 defer 语句在函数返回时解锁 plock
	// Use the defer statement to unlock plock when the function returns
	defer q.plock.Unlock()

	// 返回队列是否已关闭
	// Return whether the queue is closed
	return q.closed
}

// Add 方法将元素添加到队列中
// The Add method adds an element to the queue
func (q *Q) Add(element any) error {
	// 检查队列是否已关闭
	// Check if the queue is closed
	if q.IsClosed() {
		// 如果队列已关闭，返回错误
		// If the queue is closed, return an error
		return ErrorQueueClosed
	}

	// 检查元素是否已在队列中
	// Check if the element is already in the queue
	if q.isElementMarked(element) {
		// 如果元素已在队列中，返回错误
		// If the element is already in the queue, return an error
		return ErrorQueueElementExist
	}

	// 从节点池中获取一个新的节点
	// Get a new node from the node pool
	node := q.nodepool.Get()

	// 设置节点的数据为要添加的元素
	// Set the node's data to the element to be added
	node.SetData(element)

	// 锁定条件变量的锁
	// Lock the condition variable's lock
	q.cond.L.Lock()

	// 将节点添加到队列中
	// Add the node to the queue
	q.queue.Push(node)

	// 发送信号通知等待队列的其他线程
	// Send a signal to notify other threads waiting on the queue
	q.cond.Signal()

	// 解锁条件变量的锁
	// Unlock the condition variable's lock
	q.cond.L.Unlock()

	// 将元素添加到 dirty 集合中
	// Add the element to the dirty set
	q.prepare(element)

	// 调用回调函数 OnAdd
	// Call the callback function OnAdd
	q.config.callback.OnAdd(element)

	// 返回 nil 表示添加成功
	// Return nil to indicate successful addition
	return nil
}

// Get 方法从队列中获取一个元素
// The Get method gets an element from the queue
func (q *Q) Get() (element any, err error) {
	// 检查队列是否已关闭
	// Check if the queue is closed
	if q.IsClosed() {
		// 如果队列已关闭，返回错误
		// If the queue is closed, return an error
		return nil, ErrorQueueClosed
	}

	// 锁定队列的锁
	// Lock the queue's lock
	q.qlock.Lock()

	// 从队列中弹出一个节点
	// Pop a node from the queue
	node := q.queue.Pop()

	// 解锁队列的锁
	// Unlock the queue's lock
	q.qlock.Unlock()

	// 检查节点是否为 nil
	// Check if the node is nil
	if node == nil {
		// 如果节点为 nil，表示队列为空，返回错误
		// If the node is nil, it means the queue is empty, return an error
		return nil, ErrorQueueEmpty
	}

	// 获取节点的数据
	// Get the data of the node
	element = node.Data()

	// 将元素从 dirty 集合移动到 processing 集合
	// Move the element from the dirty set to the processing set
	q.todo(element)

	// 调用回调函数 OnGet
	// Call the callback function OnGet
	q.config.callback.OnGet(element)

	// 将节点放回节点池
	// Put the node back into the node pool
	q.nodepool.Put(node)

	// 返回获取的元素和 nil 错误
	// Return the obtained element and nil error
	return element, nil
}

// GetWithBlock 方法从队列中获取一个元素，如果队列为空，则阻塞等待
// The GetWithBlock method gets an element from the queue, if the queue is empty, it blocks and waits
func (q *Q) GetWithBlock() (element any, err error) {
	// 检查队列是否已关闭
	// Check if the queue is closed
	if q.IsClosed() {
		// 如果队列已关闭，返回错误
		// If the queue is closed, return an error
		return nil, ErrorQueueClosed
	}

	// 锁定条件变量的锁
	// Lock the condition variable's lock
	q.cond.L.Lock()

	// 当队列长度为 0 时，等待条件变量的信号
	// Wait for the condition variable's signal when the queue length is 0
	for q.queue.Len() == 0 {
		q.cond.Wait()
	}

	// 从队列中弹出一个节点
	// Pop a node from the queue
	node := q.queue.Pop()

	// 解锁条件变量的锁
	// Unlock the condition variable's lock
	q.cond.L.Unlock()

	// 检查节点是否为 nil
	// Check if the node is nil
	if node == nil {
		// 如果节点为 nil，表示队列为空，返回错误
		// If the node is nil, it means the queue is empty, return an error
		return nil, ErrorQueueEmpty
	}

	// 获取节点的数据
	// Get the data of the node
	element = node.Data()

	// 将元素从 dirty 集合移动到 processing 集合
	// Move the element from the dirty set to the processing set
	q.todo(element)

	// 调用回调函数 OnGet
	// Call the callback function OnGet
	q.config.callback.OnGet(element)

	// 将节点放回节点池
	// Put the node back into the node pool
	q.nodepool.Put(node)

	// 返回获取的元素和 nil 错误
	// Return the obtained element and nil error
	return element, nil
}

// Done 方法将元素从 processing 集合中删除，表示元素已处理完成
// The Done method removes the element from the processing set, indicating that the element has been processed
func (q *Q) Done(element any) {
	// 锁定 plock 以保护并发操作
	// Lock plock to protect concurrent operations
	q.plock.Lock()

	// 使用 defer 语句在函数返回时解锁 plock
	// Use the defer statement to unlock plock when the function returns
	defer q.plock.Unlock()

	// 从 processing 集合中删除元素
	// Remove the element from the processing set
	q.processing.Delete(element)

	// 调用回调函数 OnDone
	// Call the callback function OnDone
	q.config.callback.OnDone(element)
}

// Stop 方法关闭队列
// The Stop method shuts down the queue
func (q *Q) Stop() {
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

		// 锁定生产者锁，防止并发操作
		// Lock the producer lock to prevent concurrent operations
		q.plock.Lock()

		// 清理正在处理的元素集合
		// Clean up the set of elements being processed
		q.processing.Cleanup()

		// 清理脏元素集合
		// Clean up the set of dirty elements
		q.dirty.Cleanup()

		// 解锁生产者锁
		// Unlock the producer lock
		q.plock.Unlock()
	})
}

// GetValues 方法返回队列中所有元素的快照
// The GetValues method returns a snapshot of all elements in the queue
func (q *Q) GetValues() []any {
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
func (q *Q) Range(fn func(element any) bool) {
	// 锁定队列锁，防止并发操作
	// Lock the queue lock to prevent concurrent operations
	q.qlock.Lock()

	// 使用 defer 语句在函数返回时解锁队列锁
	// Use the defer statement to unlock the queue lock when the function returns
	defer q.qlock.Unlock()

	// 对队列中的每个元素执行给定的函数，如果函数返回 false，则停止迭代
	// Perform the given function for each element in the queue, if the function returns false, stop iterating
	q.queue.Range(func(n *list.Node) bool {
		// 调用传入的函数，传入节点的数据作为参数
		// Call the passed function, passing the data of the node as a parameter
		return fn(n.Data())
	})
}
