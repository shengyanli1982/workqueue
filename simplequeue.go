package workqueue

import (
	"sync"
	"sync/atomic"

	list "github.com/shengyanli1982/workqueue/internal/stl/deque"
)

type SimpleQ struct {
	queue   *list.Deque
	qlock   *sync.Mutex
	cond    *sync.Cond
	elepool *list.ListNodePool
	once    sync.Once
	closed  atomic.Bool
	config  *QConfig
}

// 创建一个 SimpleQueue 实例
// Create a new SimpleQueue config
func NewSimpleQueue(conf *QConfig) *SimpleQ {
	q := &SimpleQ{
		queue:   list.NewDeque(),
		elepool: list.NewListNodePool(),
		qlock:   &sync.Mutex{},
		once:    sync.Once{},
		closed:  atomic.Bool{},
		config:  conf,
	}
	q.cond = sync.NewCond(q.qlock)

	q.isConfigValid()

	return q
}

// 创建一个默认的 SimpleQueue 对象
// Create a new default SimpleQueue object
func DefaultSimpleQueue() Interface {
	return NewSimpleQueue(nil)
}

// 判断 config 是否为空，如果为空，设置默认值
// Check if config is nil, if it is, set default value
func (q *SimpleQ) isConfigValid() {
	if q.config == nil {
		q.config = NewQConfig().WithCallback(emptyCallback{})
	} else {
		if q.config.cb == nil {
			q.config.cb = emptyCallback{}
		}
	}
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
	return q.closed.Load()
}

// 添加元素到队列
// Add element to queue
func (q *SimpleQ) Add(element any) error {
	if q.IsClosed() {
		return ErrorQueueClosed
	}

	ele := q.elepool.Get()
	ele.SetData(element)

	q.cond.L.Lock()
	q.queue.Push(ele)
	q.cond.Signal()
	q.cond.L.Unlock()

	q.config.cb.OnAdd(element)

	return nil
}

// 从队列中获取一个元素, 如果队列为空，不阻塞等待
// Get an element from the queue.
func (q *SimpleQ) Get() (element any, err error) {
	if q.IsClosed() {
		return nil, ErrorQueueClosed
	}

	q.qlock.Lock()
	ln := q.queue.Pop()
	q.qlock.Unlock()

	if ln == nil { // 队列为空 (queue is empty)
		return nil, ErrorQueueEmpty
	}

	element = ln.Data()
	q.config.cb.OnGet(element)

	return element, nil
}

// 从队列中获取一个元素，如果队列为空，阻塞等待
// Get an element from the queue, if the queue is empty, block and wait.
func (q *SimpleQ) GetWithBlock() (element any, err error) {
	if q.IsClosed() {
		return nil, ErrorQueueClosed
	}

	q.cond.L.Lock()
	for q.queue.Len() == 0 {
		q.cond.Wait()
	}
	ln := q.queue.Pop()
	q.cond.L.Unlock()

	if ln == nil {
		return nil, ErrorQueueEmpty
	}

	element = ln.Data()
	q.config.cb.OnGet(element)

	return element, nil
}

// 标记元素已经处理完成
// Mark an element as done processing.
func (q *SimpleQ) Done(element any) {
	q.config.cb.OnDone(element)
}

// 关闭队列
// Shut down the queue.
func (q *SimpleQ) Stop() {
	q.once.Do(func() {
		q.closed.Store(true)
		q.cond.L.Lock()
		q.cond.Broadcast() // 唤醒所有等待的 goroutine (Wake up all waiting goroutines)
		q.queue.Reset()
		q.cond.L.Unlock()
	})
}
