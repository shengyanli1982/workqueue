package workqueue

import (
	"sync"

	list "github.com/shengyanli1982/workqueue/internal/stl/deque"
)

// 简单队列数据结构
// SimpleQueue data structure
type SimpleQ struct {
	queue    *list.Deque
	qlock    *sync.Mutex
	plock    *sync.Mutex
	cond     *sync.Cond
	nodepool *list.ListNodePool
	once     sync.Once
	closed   bool
	config   *QConfig
}

// 创建一个 SimpleQueue 实例
// Create a new SimpleQueue config
func NewSimpleQueue(conf *QConfig) *SimpleQ {
	conf = isQConfigValid(conf)
	q := &SimpleQ{
		queue:    list.NewDeque(),
		nodepool: list.NewListNodePool(),
		qlock:    &sync.Mutex{},
		plock:    &sync.Mutex{},
		once:     sync.Once{},
		closed:   false,
		config:   conf,
	}
	q.cond = sync.NewCond(q.qlock)

	return q
}

// 创建一个默认的 SimpleQueue 对象
// Create a new default SimpleQueue object
func DefaultSimpleQueue() QInterface {
	return NewSimpleQueue(nil)
}

// 获取队列长度
// Get queue length
func (q *SimpleQ) Len() int {
	q.qlock.Lock()
	defer q.qlock.Unlock()
	return q.queue.Len()
}

// 判断队列是否已经关闭
// Determine if the queue is shutting down.
func (q *SimpleQ) IsClosed() bool {
	q.plock.Lock()
	defer q.plock.Unlock()
	return q.closed
}

// 添加元素到队列
// Add element to queue
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

// 从队列中获取一个元素, 如果队列为空，不阻塞等待
// Get an element from the queue.
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

	return element, nil
}

// 从队列中获取一个元素，如果队列为空，阻塞等待
// Get an element from the queue, if the queue is empty, block and wait.
func (q *SimpleQ) GetWithBlock() (element any, err error) {
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

	// 回调
	// Callback
	q.config.callback.OnGet(element)

	// 回收节点
	// Recycle node
	q.nodepool.Put(node)

	return element, nil
}

// 标记元素已经处理完成
// Mark an element as done processing.
func (q *SimpleQ) Done(element any) {
	// 回调
	// Callback
	q.config.callback.OnDone(element)
}

// 关闭队列
// Shut down the queue.
func (q *SimpleQ) Stop() {
	q.once.Do(func() {
		q.plock.Lock()
		q.closed = true
		q.plock.Unlock()
		q.cond.L.Lock()
		q.cond.Broadcast() // 唤醒所有等待的 goroutine (Wake up all waiting goroutines)
		q.queue.Reset()
		q.cond.L.Unlock()
	})
}

// 获取队列中所有元素, 不会阻塞等待
// Get all elements in the queue. Will not block and wait.
func (q *SimpleQ) GetValues() []any {
	q.qlock.Lock()
	defer q.qlock.Unlock()
	return q.queue.SnapshotValues()
}

// 遍历队列中的元素，如果 fn 返回 false，则停止遍历
// Traverse the elements in the queue. If fn returns false, stop traversing.
func (q *SimpleQ) Range(fn func(element any) bool) {
	q.qlock.Lock()
	defer q.qlock.Unlock()
	q.queue.Range(func(n *list.Node) bool {
		return fn(n.Data())
	})
}
