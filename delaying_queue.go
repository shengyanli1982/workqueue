package workqueue

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shengyanli1982/workqueue/internal/stl/heap"
)

// DelayingQInterface 是 Queue 方法接口的延迟版本
// DelayingQInterface is the delayed version of the Queue method interface
type DelayingQInterface interface {
	// 继承 Queue 接口
	// Inherit Queue
	QInterface

	// AddAfter 添加一个元素，延迟一段时间后再执行
	// Add an element, execute it after a delay
	AddAfter(element any, delay time.Duration) error
}

// DelayingQCallback 是 Queue 的回调接口的延迟版本
// DelayingQCallback is the delayed version of the Queue callback interface
type DelayingQCallback interface {
	// 继承 Callback 接口
	// Inherit Callback
	QCallback

	// OnAddAfter 添加元素后的回调
	// Callback after adding element
	OnAddAfter(any, time.Duration)
}

// DelayingQConfig 是 Queue 的配置的延迟版本
// DelayingQConfig is the delayed version of the Queue config
type DelayingQConfig struct {
	QConfig
	callback DelayingQCallback
}

// NewDelayingQConfig 创建一个 DelayingQConfig 实例
// Create a new DelayingQConfig instance
func NewDelayingQConfig() *DelayingQConfig {
	return &DelayingQConfig{}
}

// WithCallback 设置 Queue 的回调接口
// Set Queue callback
func (c *DelayingQConfig) WithCallback(cb DelayingQCallback) *DelayingQConfig {
	c.callback = cb
	return c
}

// 验证队列的配置是否有效
// Verify that the queue configuration is valid
func isDelayingQConfigValid(conf *DelayingQConfig) *DelayingQConfig {
	if conf == nil {
		conf = NewDelayingQConfig()
		conf.WithCallback(emptyCallback{})
	} else {
		if conf.callback == nil {
			conf.callback = emptyCallback{}
		}
	}

	return conf
}

// 延迟队列数据结构
// Delayed queue data structure
type DelayingQ struct {
	QInterface
	waiting     *heap.Heap
	elementpool *heap.HeapElementPool
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	once        sync.Once
	wlock       *sync.Mutex
	now         atomic.Int64
	config      *DelayingQConfig
}

// 创建 DelayingQueue 实例
// Create a DelayingQueue instance
func newDelayingQueue(conf *DelayingQConfig, queue QInterface) *DelayingQ {
	if queue == nil {
		return nil
	}

	conf = isDelayingQConfigValid(conf)
	conf.QConfig.callback = conf.callback

	q := &DelayingQ{
		QInterface:  queue,
		waiting:     heap.NewHeap(),
		elementpool: heap.NewHeapElementPool(),
		wlock:       &sync.Mutex{},
		wg:          sync.WaitGroup{},
		now:         atomic.Int64{},
		once:        sync.Once{},
		config:      conf,
	}

	q.ctx, q.cancel = context.WithCancel(context.Background())

	q.wg.Add(2)
	go q.loop()
	go q.syncNow()

	return q
}

// 创建一个 DelayingQueue 实例
// Create a new DelayingQueue config
func NewDelayingQueue(conf *DelayingQConfig) *DelayingQ {
	conf = isDelayingQConfigValid(conf)
	conf.QConfig.callback = conf.callback
	return newDelayingQueue(conf, NewQueue(&conf.QConfig))
}

// 创建一个 DelayingQueue 实例, 使用自定义 Queue (实现了 Q 接口)
// Create a new DelayingQueue config, use custom Queue (implement Q interface)
func NewDelayingQueueWithCustomQueue(conf *DelayingQConfig, queue QInterface) *DelayingQ {
	conf = isDelayingQConfigValid(conf)
	conf.QConfig.callback = conf.callback
	return newDelayingQueue(conf, queue)
}

// 创建一个默认的 DelayingQueue 对象
// Create a new default DelayingQueue object
func DefaultDelayingQueue() DelayingQInterface {
	return NewDelayingQueue(nil)
}

// AddAfter 将元素添加到队列中，在延迟一段时间后处理
// Add an element to the queue and process it after a specified delay
func (q *DelayingQ) AddAfter(element any, delay time.Duration) error {
	// 如果队列已经关闭，返回 ErrorQueueClosed 错误
	// If the queue is already closed, return ErrorQueueClosed
	if q.IsClosed() {
		return ErrorQueueClosed
	}

	// 如果延迟时间小于等于 0，直接添加到队列中
	// If the delay time is less than or equal to 0, add it directly to the queue
	if delay <= 0 {
		return q.Add(element)
	}

	// 从对象池中获取一个元素
	// Get an element from the object pool
	elem := q.elementpool.Get()
	elem.SetData(element)
	elem.SetValue(time.Now().Add(delay).UnixMilli())

	// 添加到堆中
	// Add to the heap
	q.wlock.Lock()
	q.waiting.Push(elem)
	q.wlock.Unlock()

	// 回调
	// Callback
	q.config.callback.OnAddAfter(element, delay)

	return nil
}

// 同步当前的时间
// Sync current time
func (q *DelayingQ) syncNow() {
	// 心跳
	// Heartbeat
	heartbeat := time.NewTicker(time.Millisecond * defaultQueueHeartbeat)

	defer func() {
		q.wg.Done()
		heartbeat.Stop()
	}()

	// 循环同步当前时间
	// Loop to sync current time
	for {
		select {
		case <-q.ctx.Done():
			return
		case <-heartbeat.C:
			q.now.Store(time.Now().UnixMilli())
		}
	}
}

// 循环处理 Heap 中的元素
// Loop to process elements in Heap
func (q *DelayingQ) loop() {
	// 心跳
	// Heartbeat
	heartbeat := time.NewTicker(time.Millisecond * defaultQueueHeartbeat)

	defer func() {
		q.wg.Done()
		heartbeat.Stop()
	}()

	// 循环处理堆中的元素
	// Loop to process elements in the heap
	for {
		select {
		case <-q.ctx.Done():
			return
		default:
			q.wlock.Lock()
			// 如果堆中有元素
			// If there are elements in the heap
			if q.waiting.Len() > 0 {
				// 获取堆顶元素
				// Get the top element of the heap
				elem := q.waiting.Head()

				// 如果堆顶元素的时间小于当前时间, 意味对象已经超时
				// If the time of the top element of the heap is less than the current time, it means the object has timed out
				if elem.Value() <= q.now.Load() {
					// 弹出堆顶元素
					// Pop the top element of the heap
					_ = q.waiting.Pop()
					q.wlock.Unlock()

					// 添加到队列中
					// Add to the queue
					if err := q.Add(elem.Data()); err != nil {
						q.wlock.Lock()
						// 重置元素的值 Reset the value of the element
						// 1500ms 后再次处理元素
						elem.SetValue(q.now.Load() + defaultQueueHeartbeat*3)

						// 将元素重新添加到堆中 Re-add the element to the heap
						// Re-add the element to the heap
						q.waiting.Push(elem)
						q.wlock.Unlock()
					} else {
						// 释放元素
						// Free element
						q.elementpool.Put(elem)
					}
				} else {
					q.wlock.Unlock()
				}
			} else {
				q.wlock.Unlock()
				// 500ms 后再次检查堆中的元素
				// Check the elements in the heap again after 500ms
				<-heartbeat.C
			}
		}
	}
}

// 关闭队列
// Close queue
func (q *DelayingQ) Stop() {
	q.QInterface.Stop()
	q.once.Do(func() {
		q.cancel()
		q.wg.Wait()
		q.waiting.Reset()
	})
}
