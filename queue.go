package workqueue

import (
	"sync"

	list "github.com/shengyanli1982/workqueue/internal/stl/deque"
	"github.com/shengyanli1982/workqueue/internal/stl/set"
)

// 队列方法接口
// Queue interface
type Interface interface {
	// 添加一个元素到队列
	// Add adds an element to the queue.
	Add(element any) error

	// 获得 queue 的长度
	// Len returns the number of elements in the queue.
	Len() int

	// 遍历队列中的元素，如果 fn 返回 false，则停止遍历
	// Range iterates over each element in the queue and calls the provided function.
	// If the function returns false, the iteration stops.
	Range(fn func(element any) bool)

	// 获得 queue 中的一个元素，如果 queue 为空，返回 ErrorQueueEmpty
	// Get retrieves an element from the queue.
	Get() (element any, err error)

	// 获得 queue 中的一个元素，如果 queue 为空，阻塞等待
	// GetWithBlock retrieves an element from the queue, blocking if the queue is empty.
	GetWithBlock() (element any, err error)

	// 获得 queue 中的所有元素
	// GetValues returns all elements in the queue.
	GetValues() []any

	// 标记元素已经处理完成
	// Done marks an element as processed and removes it from the queue.
	Done(element any)

	// 关闭队列
	// Stop stops the queue and releases any resources.
	Stop()

	// 判断队列是否已经关闭
	// IsClosed returns true if the queue is closed, false otherwise.
	IsClosed() bool
}

// 队列的回调接口
// Callback interface
type Callback interface {
	OnAdd(any)
	OnGet(any)
	OnDone(any)
}

// 队列的配置
// Queue config
type QConfig struct {
	callback Callback
}

// 创建一个队列的配置
// Create a new Queue config
func NewQConfig() *QConfig {
	return &QConfig{}
}

// 设置队列的回调接口
// Set Queue callback
func (c *QConfig) WithCallback(cb Callback) *QConfig {
	c.callback = cb
	return c
}

// 验证队列的配置是否有效
// Verify that the queue configuration is valid
func isQConfigValid(conf *QConfig) *QConfig {
	if conf == nil {
		conf = NewQConfig().WithCallback(emptyCallback{})
	} else {
		if conf.callback == nil {
			conf.callback = emptyCallback{}
		}
	}
	return conf
}

type Q struct {
	queue      *list.Deque
	nodepool   *list.ListNodePool
	qlock      *sync.Mutex
	cond       *sync.Cond
	dirty      set.Set
	processing set.Set
	lock       *sync.Mutex
	once       sync.Once
	closed     bool
	config     *QConfig
}

// 创建一个 Queue 对象
// Create a new Queue object.
func NewQueue(conf *QConfig) *Q {
	conf = isQConfigValid(conf)
	q := &Q{
		dirty:      set.NewSet(),
		processing: set.NewSet(),
		queue:      list.NewDeque(),
		qlock:      &sync.Mutex{},
		nodepool:   list.NewListNodePool(),
		lock:       &sync.Mutex{},
		once:       sync.Once{},
		closed:     false,
		config:     conf,
	}
	q.cond = sync.NewCond(q.qlock)

	return q
}

// 创建一个默认的 Queue 对象
// Create a new default Queue object.
func DefaultQueue() Interface {
	return NewQueue(nil)
}

// 标记已经准备好处理的元素
// Mark an element as ready to be processed.
func (q *Q) todo(element any) {
	q.lock.Lock()
	q.dirty.Delete(element)
	q.processing.Add(element)
	q.lock.Unlock()
}

// 标记待被处理的元素
// Mark an element to be processed
func (q *Q) prepare(element any) {
	q.lock.Lock()
	q.dirty.Add(element)
	q.lock.Unlock()
}

// 判断元素是否已经被标记
// Determine if an element has been marked.
func (q *Q) isElementMarked(element any) bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.dirty.Has(element) || q.processing.Has(element) {
		return true
	}
	return false
}

// 获取队列长度
// Get queue length
func (q *Q) Len() int {
	q.qlock.Lock()
	defer q.qlock.Unlock()
	return q.queue.Len()
}

// 判断队列是否已经关闭
// Determine if the queue is shutting down.
func (q *Q) IsClosed() bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.closed
}

// 添加元素到队列
// Add element to queue
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

// 从队列中获取一个元素, 如果队列为空，不阻塞等待
// Get an element from the queue.
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

// 从队列中获取一个元素，如果队列为空，阻塞等待
// Get an element from the queue, if the queue is empty, block and wait.
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

// 标记元素已经处理完成
// Mark an element as done processing.
func (q *Q) Done(element any) {
	q.lock.Lock()
	defer q.lock.Unlock()

	// 从 processing 中删除元素
	// Remove the element from processing
	q.processing.Delete(element)

	// 回调
	// Callback
	q.config.callback.OnDone(element)
}

// 关闭队列
// Shut down the queue.
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

		// 标记关闭队列
		// Mark the queue as closed
		q.lock.Lock()
		q.closed = true
		q.lock.Unlock()

		// 清理所有的元素
		// Clean up all elements
		go func() {
			q.lock.Lock()
			q.dirty.Cleanup()
			q.lock.Unlock()
			wg.Done()
		}()
		go func() {
			q.lock.Lock()
			q.processing.Cleanup()
			q.lock.Unlock()
			wg.Done()
		}()

		// 等待所有的 goroutine 完成
		// Wait for all goroutines to complete
		wg.Wait()
	})
}

// 获取队列中所有元素, 不会阻塞等待
// Get all elements in the queue. Will not block and wait.
func (q *Q) GetValues() []any {
	q.qlock.Lock()
	defer q.qlock.Unlock()
	return q.queue.SnapshotValues()
}

// 遍历队列中的元素，如果 fn 返回 false，则停止遍历
// Traverse the elements in the queue. If fn returns false, stop traversing.
func (q *Q) Range(fn func(element any) bool) {
	q.qlock.Lock()
	defer q.qlock.Unlock()
	q.queue.Range(func(n *list.Node) bool {
		return fn(n.Data())
	})
}
