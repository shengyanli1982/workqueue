package workqueue

import (
	"math"
	"sync"
	"sync/atomic"
)

type SimpleQ struct {
	queue  chan any
	once   sync.Once
	closed atomic.Bool
	config *QConfig
}

// 创建一个 SimpleQueue 实例
// Create a new SimpleQueue config
func NewSimpleQueue(conf *QConfig) *SimpleQ {
	q := &SimpleQ{
		once:   sync.Once{},
		closed: atomic.Bool{},
		config: conf,
	}

	q.isConfigValid()

	q.queue = make(chan any, q.config.cap)

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

// 获取队列长度
// Get queue length
func (q *SimpleQ) Len() int {
	return len(q.queue)
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

	select {
	case q.queue <- element:
		q.config.cb.OnAdd(element)
		return nil
	default:
		return ErrorQueueFull
	}
}

// 从队列中获取一个元素, 如果队列为空，不阻塞等待
// Get an element from the queue.
func (q *SimpleQ) Get() (element any, err error) {
	if q.IsClosed() {
		return nil, ErrorQueueClosed
	}

	select {
	case element = <-q.queue:
		q.config.cb.OnGet(element)
		return element, nil
	default:
		return nil, ErrorQueueEmpty
	}
}

// 从队列中获取一个元素，如果队列为空，阻塞等待
// Get an element from the queue, if the queue is empty, block and wait.
func (q *SimpleQ) GetWithBlock() (element any, err error) {
	if q.IsClosed() {
		return nil, ErrorQueueClosed
	}

	element = <-q.queue
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
		close(q.queue)
		for range q.queue {
			// drain the queue
		}
	})
}
