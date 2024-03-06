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
	OnAdd(any)
	OnGet(any)
	OnDone(any)
}

// QConfig 定义了队列的配置
// QConfig defines the configuration of the queue
type QConfig struct {
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
	queue      *list.Deque        // 队列
	nodepool   *list.ListNodePool // 节点池
	qlock      *sync.Mutex        // 队列锁
	cond       *sync.Cond         // 条件变量
	dirty      set.Set            // 待处理的元素集合
	processing set.Set            // 正在处理的元素集合
	plock      *sync.Mutex        // 处理过程锁
	once       sync.Once          // 用于确保某个操作只执行一次
	closed     bool               // 队列是否已关闭的标志
	config     *QConfig           // 队列配置
}

// NewQueue 函数创建一个新的 Queue 对象
// The NewQueue function creates a new Queue object
func NewQueue(conf *QConfig) *Q {
	// 验证配置是否有效
	// Validate if the configuration is valid
	conf = isQConfigValid(conf)

	// 创建一个新的 Queue 对象
	// Create a new Queue object
	q := &Q{
		dirty:      set.NewSet(),           // 初始化待处理的元素集合，Initialize the set of elements to be processed
		processing: set.NewSet(),           // 初始化正在处理的元素集合，Initialize the set of elements being processed
		queue:      list.NewDeque(),        // 初始化队列，Initialize the queue
		qlock:      &sync.Mutex{},          // 初始化队列锁，Initialize the queue lock
		nodepool:   list.NewListNodePool(), // 初始化节点池，Initialize the node pool
		plock:      &sync.Mutex{},          // 初始化处理过程锁，Initialize the processing lock
		once:       sync.Once{},            // 初始化 sync.Once 对象，Initialize the sync.Once object
		closed:     false,                  // 设置 closed 为 false，Set closed to false
		config:     conf,                   // 设置配置，Set the configuration
	}

	// 创建一个新的条件变量
	// Create a new condition variable
	q.cond = sync.NewCond(q.qlock)

	// 返回 Queue 对象
	// Return the Queue object
	return q
}

// DefaultQueue 函数创建一个默认的 Queue 对象
// The DefaultQueue function creates a default Queue object
func DefaultQueue() QInterface {
	// 创建一个新的 Queue 对象，配置为 nil
	// Create a new Queue object with nil configuration
	return NewQueue(nil)
}

// todo 方法标记一个元素已经准备好处理
// The todo method marks an element as ready to be processed
func (q *Q) todo(element any) {
	q.plock.Lock()
	// 从 dirty 集合中删除元素
	// Remove the element from the dirty set
	q.dirty.Delete(element)
	// 将元素添加到 processing 集合中
	// Add the element to the processing set
	q.processing.Add(element)
	q.plock.Unlock()
}

// prepare 方法标记一个元素待处理
// The prepare method marks an element to be processed
func (q *Q) prepare(element any) {
	q.plock.Lock()
	// 将元素添加到 dirty 集合中
	// Add the element to the dirty set
	q.dirty.Add(element)
	q.plock.Unlock()
}

// isElementMarked 方法判断一个元素是否已经被标记
// The isElementMarked method determines whether an element has been marked
func (q *Q) isElementMarked(element any) bool {
	q.plock.Lock()
	defer q.plock.Unlock()

	// 如果元素在 dirty 集合或 processing 集合中，返回 true
	// If the element is in the dirty set or processing set, return true
	return q.dirty.Has(element) || q.processing.Has(element)
}

// Len 方法获取队列的长度
// The Len method gets the length of the queue
func (q *Q) Len() int {
	q.qlock.Lock()
	defer q.qlock.Unlock()

	// 返回队列的长度
	// Return the length of the queue
	return q.queue.Len()
}

// IsClosed 方法判断队列是否已经关闭
// The IsClosed method determines whether the queue is closed
func (q *Q) IsClosed() bool {
	q.plock.Lock()
	defer q.plock.Unlock()

	// 返回 closed 的值
	// Return the value of closed
	return q.closed
}

// Add 方法将元素添加到队列中
// The Add method adds an element to the queue
func (q *Q) Add(element any) error {
	// 如果队列已经关闭，返回 ErrorQueueClosed 错误
	// If the queue is already closed, return ErrorQueueClosed
	if q.IsClosed() {
		return ErrorQueueClosed
	}

	// 如果元素已经被标记，返回 ErrorQueueElementExist 错误
	// If the element has been marked, return ErrorQueueElementExist
	if q.isElementMarked(element) {
		return ErrorQueueElementExist
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

	// 标记元素已经准备好处理
	// Mark the element as ready to be processed
	q.prepare(element)

	// 回调
	// Callback
	q.config.callback.OnAdd(element)

	return nil
}

// Get 方法从队列中获取一个元素, 如果队列为空，不阻塞等待
// The Get method gets an element from the queue, if the queue is empty, it does not block and wait
func (q *Q) Get() (element any, err error) {
	// 如果队列已经关闭，返回 ErrorQueueClosed 错误
	// If the queue is already closed, return ErrorQueueClosed
	if q.IsClosed() {
		return nil, ErrorQueueClosed
	}

	// 如果队列为空，返回 ErrorQueueEmpty 错误
	// If the queue is empty, return ErrorQueueEmpty
	q.qlock.Lock()
	node := q.queue.Pop()
	q.qlock.Unlock()
	if node == nil {
		return nil, ErrorQueueEmpty
	}

	// 从节点中获取数据
	// Get data from the node
	element = node.Data()

	// 标记元素已经准备好处理
	// Mark the element as ready to be processed
	q.todo(element)

	// 回调
	// Callback
	q.config.callback.OnGet(element)

	// 回收节点
	// Recycle node
	q.nodepool.Put(node)

	return element, nil
}

// GetWithBlock 方法从队列中获取一个元素，如果队列为空，阻塞等待
// The GetWithBlock method gets an element from the queue, if the queue is empty, it blocks and waits
func (q *Q) GetWithBlock() (element any, err error) {
	// 如果队列已经关闭，返回 ErrorQueueClosed 错误
	// If the queue is already closed, return ErrorQueueClosed
	if q.IsClosed() {
		return nil, ErrorQueueClosed
	}

	// 如果队列为空，阻塞等待。否者，读取节点中的数据
	// If the queue is empty, block and wait. Otherwise, read the data from the node.
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

	// 标记元素已经准备好处理
	// Mark the element as ready to be processed
	q.todo(element)

	// 回调
	// Callback
	q.config.callback.OnGet(element)

	// 回收节点
	// Recycle node
	q.nodepool.Put(node)

	return element, nil
}

// Done 方法标记一个元素已经处理完成
// The Done method marks an element as done processing
func (q *Q) Done(element any) {
	q.plock.Lock()
	defer q.plock.Unlock()

	// 从 processing 集合中删除元素
	// Remove the element from the processing set
	q.processing.Delete(element)

	// 回调
	// Callback
	q.config.callback.OnDone(element)
}

// Stop 方法关闭队列
// The Stop method shuts down the queue
func (q *Q) Stop() {
	q.once.Do(func() {
		wg := sync.WaitGroup{}
		wg.Add(3)

		// 唤醒所有等待的 goroutine
		// Wake up all waiting goroutines
		go func() {
			q.cond.L.Lock()
			q.cond.Broadcast()
			q.queue.Reset()
			q.cond.L.Unlock()
			wg.Done()
		}()

		// 标记队列为已关闭
		// Mark the queue as closed
		q.plock.Lock()
		q.closed = true
		q.plock.Unlock()

		// 清理所有的元素
		// Clean up all elements
		go func() {
			q.plock.Lock()
			q.dirty.Cleanup()
			q.plock.Unlock()
			wg.Done()
		}()
		go func() {
			q.plock.Lock()
			q.processing.Cleanup()
			q.plock.Unlock()
			wg.Done()
		}()

		// 等待所有的 goroutine 完成
		// Wait for all goroutines to complete
		wg.Wait()
	})
}

// GetValues 方法获取队列中所有元素, 不会阻塞等待
// The GetValues method gets all elements in the queue, it will not block and wait
func (q *Q) GetValues() []any {
	q.qlock.Lock()
	defer q.qlock.Unlock()

	// 使用 SnapshotValues 方法获取队列中所有元素的快照
	// Use the SnapshotValues method to get a snapshot of all elements in the queue
	return q.queue.SnapshotValues()
}

// Range 方法遍历队列中的元素，如果 fn 返回 false，则停止遍历
// The Range method traverses the elements in the queue. If fn returns false, stop traversing
func (q *Q) Range(fn func(element any) bool) {
	q.qlock.Lock()
	defer q.qlock.Unlock()

	// 使用 Range 方法遍历队列中的每一个元素，将元素传递给 fn 函数
	// Use the Range method to traverse each element in the queue and pass the element to the fn function
	q.queue.Range(func(n *list.Node) bool {
		// 调用 fn 函数处理元素，如果 fn 返回 false，则停止遍历
		// Call the fn function to handle the element, if fn returns false, stop traversing
		return fn(n.Data())
	})
}
