package workqueue

import (
	"sync"

	list "github.com/shengyanli1982/workqueue/internal/stl/deque"
)

// SimpleQ 结构体定义了一个简单队列数据结构
// The SimpleQ struct defines a simple queue data structure
type SimpleQ struct {
	queue    *list.Deque        // 队列
	qlock    *sync.Mutex        // 用于保护队列的互斥锁
	plock    *sync.Mutex        // 用于保护关闭标志的互斥锁
	cond     *sync.Cond         // 条件变量，用于实现队列的阻塞读
	nodepool *list.ListNodePool // 节点池，用于减少内存分配
	once     sync.Once          // 用于确保队列只被关闭一次
	closed   bool               // 队列是否已经关闭的标志
	config   *QConfig           // 队列的配置
}

// NewSimpleQueue 函数创建并返回一个新的 SimpleQueue 实例
// The NewSimpleQueue function creates and returns a new SimpleQ instance
func NewSimpleQueue(conf *QConfig) *SimpleQ {
	conf = isQConfigValid(conf) // 验证配置的有效性
	q := &SimpleQ{
		queue:    list.NewDeque(),        // 创建一个新的双端队列
		nodepool: list.NewListNodePool(), // 创建一个新的节点池
		qlock:    &sync.Mutex{},          // 初始化互斥锁
		plock:    &sync.Mutex{},          // 初始化互斥锁
		once:     sync.Once{},            // 初始化 sync.Once
		closed:   false,                  // 设置 closed 为 false
		config:   conf,                   // 设置配置
	}
	q.cond = sync.NewCond(q.qlock) // 创建一个新的条件变量

	return q
}

// DefaultSimpleQueue 函数创建并返回一个新的具有默认配置的 SimpleQueue 实例
// The DefaultSimpleQueue function creates and returns a new SimpleQ instance with default configuration
func DefaultSimpleQueue() QInterface {
	return NewSimpleQueue(nil)
}

// Len 方法返回队列的长度
// The Len method returns the length of the queue
func (q *SimpleQ) Len() int {
	q.qlock.Lock()
	defer q.qlock.Unlock()

	// 返回队列的长度
	// Return the length of the queue
	return q.queue.Len()
}

// IsClosed 方法返回队列是否已经关闭
// The IsClosed method returns whether the queue is closed
func (q *SimpleQ) IsClosed() bool {
	q.plock.Lock()
	defer q.plock.Unlock()

	// 返回 closed 的值
	// Return the value of closed
	return q.closed
}

// Add 方法将一个元素添加到队列
// The Add method adds an element to the queue
func (q *SimpleQ) Add(element any) error {
	// 如果队列已经关闭，返回 ErrorQueueClosed 错误
	// If the queue is already closed, return ErrorQueueClosed
	if q.IsClosed() {
		return ErrorQueueClosed
	}

	// 从资源池中获取一个节点
	// Get a node from the resource pool
	node := q.nodepool.Get()
	node.SetData(element)

	// 添加到队列中，并发送信号
	// Add to the queue and send a signal
	q.cond.L.Lock()
	q.queue.Push(node)
	q.cond.Signal()
	q.cond.L.Unlock()

	// 回调
	// Callback
	q.config.callback.OnAdd(element)

	return nil
}

// Get 方法从队列中获取一个元素, 如果队列为空，不阻塞等待
// The Get method gets an element from the queue. If the queue is empty, it does not block and wait
func (q *SimpleQ) Get() (element any, err error) {
	// 如果队列已经关闭，返回 ErrorQueueClosed 错误
	// If the queue is already closed, return ErrorQueueClosed
	if q.IsClosed() {
		return nil, ErrorQueueClosed
	}

	// 如果队列为空，返回 ErrorQueueEmpty 错误
	// If the queue is empty, return ErrorQueueEmpty
	q.qlock.Lock()
	n := q.queue.Pop()
	q.qlock.Unlock()
	if n == nil {
		return nil, ErrorQueueEmpty
	}

	// 从节点中获取数据
	// Get data from the node
	element = n.Data()

	// 回调
	// Callback
	q.config.callback.OnGet(element)

	// 回收节点
	// Recycle node
	q.nodepool.Put(n)

	// 返回元素
	// Return the element
	return element, nil
}

// GetWithBlock 方法从队列中获取一个元素，如果队列为空，阻塞等待
// The GetWithBlock method gets an element from the queue. If the queue is empty, it blocks and waits
func (q *SimpleQ) GetWithBlock() (element any, err error) {
	// 如果队列已经关闭，返回 ErrorQueueClosed 错误
	// If the queue is already closed, return ErrorQueueClosed
	if q.IsClosed() {
		return nil, ErrorQueueClosed
	}

	// 如果队列为空，阻塞等待。否者，读取节点中的数据
	// If the queue is empty, block and wait. Otherwise, read the data from the node
	q.cond.L.Lock()
	for q.queue.Len() == 0 {
		q.cond.Wait()
	}
	node := q.queue.Pop()
	q.cond.L.Unlock()
	if node == nil {
		return nil, ErrorQueueEmpty
	}

	// 从节点中获取数据
	// Get data from the node
	element = node.Data()

	// 回调
	// Callback
	q.config.callback.OnGet(element)

	// 回收节点
	// Recycle node
	q.nodepool.Put(node)

	// 返回元素
	// Return the element
	return element, nil
}

// Done 方法标记一个元素已经处理完成
// The Done method marks an element as done processing
func (q *SimpleQ) Done(element any) {
	// 回调
	// Callback
	q.config.callback.OnDone(element)
}

// Stop 方法关闭队列
// The Stop method shuts down the queue
func (q *SimpleQ) Stop() {
	q.once.Do(func() {
		q.plock.Lock()     // 加锁以保护共享资源
		q.closed = true    // 设置 closed 为 true
		q.plock.Unlock()   // 在函数返回时解锁
		q.cond.L.Lock()    // 加锁以保护共享资源
		q.cond.Broadcast() // 唤醒所有等待的 goroutine
		q.queue.Reset()    // 重置队列
		q.cond.L.Unlock()  // 在函数返回时解锁
	})
}

// GetValues 方法获取队列中所有元素, 不会阻塞等待
// The GetValues method gets all elements in the queue. It will not block and wait
func (q *SimpleQ) GetValues() []any {
	q.qlock.Lock()         // 加锁以保护共享资源
	defer q.qlock.Unlock() // 在函数返回时解锁

	// 返回队列中所有元素
	// Return all elements in the queue
	return q.queue.SnapshotValues()
}

// Range 方法遍历队列中的元素，如果 fn 返回 false，则停止遍历
// The Range method traverses the elements in the queue. If fn returns false, it stops traversing
func (q *SimpleQ) Range(fn func(element any) bool) {
	q.qlock.Lock()         // 加锁以保护共享资源
	defer q.qlock.Unlock() // 在函数返回时解锁

	// 遍历队列中的元素
	// Traverse the elements in the queue
	q.queue.Range(func(n *list.Node) bool {
		return fn(n.Data()) // 如果 fn 返回 false，则停止遍历
	})
}
