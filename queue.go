package workqueue

import (
	"math"
	"sync"

	st "github.com/shengyanli1982/workqueue/pkg/structs"
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
	cb  Callback
	cap int
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

// 设置队列的容量
// Set Queue capacity
func (c *QConfig) WithCap(cap int) *QConfig {
	c.cap = cap
	return c
}

type Q struct {
	queue      chan any
	dirty      st.Set
	processing st.Set
	lock       *sync.Mutex
	once       sync.Once
	closed     bool
	config     *QConfig
}

// 创建一个 Queue 对象
// Create a new Queue object.
func NewQueue(conf *QConfig) *Q {
	q := &Q{
		dirty:      st.NewSet(),
		processing: st.NewSet(),
		lock:       &sync.Mutex{},
		once:       sync.Once{},
		closed:     false,
		config:     conf,
	}

	q.isConfigValid()

	q.queue = make(chan any, q.config.cap)

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
		q.config = &QConfig{}
		q.config.WithCallback(emptyCallback{})
		q.config.WithCap(defaultQueueCap)
	} else {
		if q.config.cb == nil {
			q.config.cb = emptyCallback{}
		}
		if q.config.cap < defaultQueueCap && q.config.cap >= 0 {
			q.config.cap = defaultQueueCap
		}
		if q.config.cap < 0 {
			q.config.cap = math.MaxInt64 // 无限容量, unlimited capacity
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
	q.lock.Lock()
	defer q.lock.Unlock()
	return len(q.queue)
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

	select {
	case q.queue <- element:
		q.prepare(element)
		q.config.cb.OnAdd(element)
		return nil
	default:
		return ErrorQueueFull
	}
}

// 从队列中获取一个元素, 如果队列为空，不阻塞等待
// Get an element from the queue.
func (q *Q) Get() (element any, err error) {
	if q.IsClosed() {
		return nil, ErrorQueueClosed
	}

	select {
	case element = <-q.queue:
		q.todo(element)
		q.config.cb.OnGet(element)
		return element, nil
	default:
		return nil, ErrorQueueEmpty
	}
}

// 从队列中获取一个元素，如果队列为空，阻塞等待
// Get an element from the queue, if the queue is empty, block and wait.
func (q *Q) GetWithBlock() (element any, err error) {
	if q.IsClosed() {
		return nil, ErrorQueueClosed
	}

	element = <-q.queue
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
			defer wg.Done()
			close(q.queue)
			for range q.queue {
				// drain the queue
			}
		}()
		go func() {
			defer wg.Done()
			q.dirty.Cleanup()
		}()
		go func() {
			defer wg.Done()
			q.processing.Cleanup()
		}()
		wg.Wait()
	})
}
