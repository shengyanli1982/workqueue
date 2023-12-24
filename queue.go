package workqueue

import (
	"sync"

	list "github.com/shengyanli1982/workqueue/internal/stl/deque"
	"github.com/shengyanli1982/workqueue/internal/stl/set"
)

// 队列方法接口
// Queue interface
type Interface interface {
	Add(element any) error
	Len() int
	Get() (element any, err error)
	GetWithBlock() (element any, err error)
	Done(element any)
	Stop()
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
	cb Callback
}

// 创建一个队列的配置
// Create a new Queue config
func NewQConfig() *QConfig {
	return &QConfig{}
}

// 设置队列的回调接口
// Set Queue callback
func (c *QConfig) WithCallback(cb Callback) *QConfig {
	c.cb = cb
	return c
}

type Q struct {
	queue      *list.Deque
	elepool    *list.ListNodePool
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
	q := &Q{
		dirty:      set.NewSet(),
		processing: set.NewSet(),
		queue:      list.NewDeque(),
		qlock:      &sync.Mutex{},
		elepool:    list.NewListNodePool(),
		lock:       &sync.Mutex{},
		once:       sync.Once{},
		closed:     false,
		config:     conf,
	}
	q.cond = sync.NewCond(q.qlock)

	q.isConfigValid()

	return q
}

// 创建一个默认的 Queue 对象
// Create a new default Queue object.
func DefaultQueue() Interface {
	return NewQueue(nil)
}

// 判断 config 是否为空，如果为空，设置默认值
// Check if config is nil, if it is, set default value
func (q *Q) isConfigValid() {
	if q.config == nil {
		q.config = NewQConfig().WithCallback(emptyCallback{})
	} else {
		if q.config.cb == nil {
			q.config.cb = emptyCallback{}
		}
	}
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
	if q.IsClosed() {
		return ErrorQueueClosed
	}

	if q.isElementMarked(element) {
		return ErrorQueueElementExist
	}

	ele := q.elepool.Get()
	ele.SetData(element)

	q.cond.L.Lock()
	q.queue.Push(ele)
	q.cond.Signal()
	q.cond.L.Unlock()

	q.prepare(element)
	q.config.cb.OnAdd(element)

	return nil
}

// 从队列中获取一个元素, 如果队列为空，不阻塞等待
// Get an element from the queue.
func (q *Q) Get() (element any, err error) {
	if q.IsClosed() {
		return nil, ErrorQueueClosed
	}

	q.qlock.Lock()
	ln := q.queue.Pop()
	q.qlock.Unlock()

	if ln == nil {
		return nil, ErrorQueueEmpty
	}

	element = ln.Data()
	q.todo(element)
	q.config.cb.OnGet(element)

	return element, nil
}

// 从队列中获取一个元素，如果队列为空，阻塞等待
// Get an element from the queue, if the queue is empty, block and wait.
func (q *Q) GetWithBlock() (element any, err error) {
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
	q.todo(element)
	q.config.cb.OnGet(element)

	return element, nil
}

// 标记元素已经处理完成
// Mark an element as done processing.
func (q *Q) Done(element any) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.processing.Delete(element)
	q.config.cb.OnDone(element)
}

// 关闭队列
// Shut down the queue.
func (q *Q) Stop() {
	q.once.Do(func() {
		q.lock.Lock()
		defer q.lock.Unlock()
		q.closed = true
		wg := sync.WaitGroup{}
		wg.Add(3)
		go func() {
			q.cond.L.Lock()
			q.cond.Broadcast() // 唤醒所有等待的 goroutine (Wake up all waiting goroutines)
			q.queue.Reset()
			q.cond.L.Unlock()
			wg.Done()
		}()
		go func() {
			q.dirty.Cleanup()
			wg.Done()
		}()
		go func() {
			q.processing.Cleanup()
			wg.Done()
		}()
		wg.Wait()
	})
}
