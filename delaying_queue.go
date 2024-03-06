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
	// Inherit Queue interface
	QInterface

	// AddAfter 添加一个元素，延迟一段时间后再执行
	// AddAfter adds an element, and executes it after a delay
	AddAfter(element any, delay time.Duration) error
}

// DelayingQCallback 是 Queue 的回调接口的延迟版本
// DelayingQCallback is the delayed version of the Queue callback interface
type DelayingQCallback interface {
	// 继承 Callback 接口
	// Inherit Callback interface
	QCallback

	// OnAddAfter 添加元素后的回调
	// OnAddAfter is the callback after adding an element
	OnAddAfter(any, time.Duration)
}

// DelayingQConfig 是 Queue 的配置的延迟版本
// DelayingQConfig is the delayed version of the Queue configuration
type DelayingQConfig struct {
	QConfig
	callback DelayingQCallback
}

// NewDelayingQConfig 创建一个 DelayingQConfig 实例
// NewDelayingQConfig creates a new DelayingQConfig instance
func NewDelayingQConfig() *DelayingQConfig {
	return &DelayingQConfig{}
}

// WithCallback 设置 Queue 的回调接口
// WithCallback sets the Queue callback interface
func (c *DelayingQConfig) WithCallback(cb DelayingQCallback) *DelayingQConfig {
	c.callback = cb
	return c
}

// 验证队列的配置是否有效
// Verify that the queue configuration is valid
func isDelayingQConfigValid(conf *DelayingQConfig) *DelayingQConfig {
	if conf == nil {
		// 如果 conf 为 nil，创建一个新的 DelayingQConfig 实例，并设置一个空的回调接口
		// If conf is nil, create a new DelayingQConfig instance and set an empty callback interface
		conf = NewDelayingQConfig()
		conf.WithCallback(newEmptyCallback())
	} else {
		// 如果 conf 不为 nil，但回调接口为 nil，设置一个空的回调接口
		// If conf is not nil, but the callback interface is nil, set an empty callback interface
		if conf.callback == nil {
			conf.callback = newEmptyCallback()
		}
	}

	return conf
}

// 延迟队列数据结构
// Delayed queue data structure
type DelayingQ struct {
	QInterface                        // 队列接口
	waiting     *heap.Heap            // 等待处理的元素堆
	elementpool *heap.HeapElementPool // 元素池，用于存储等待处理的元素
	ctx         context.Context       // 上下文，用于控制并发操作
	cancel      context.CancelFunc    // 取消函数，用于取消上下文
	wg          sync.WaitGroup        // 同步等待组，用于等待并发操作完成
	once        sync.Once             // 保证某个操作只执行一次的同步原语
	wlock       *sync.Mutex           // 互斥锁，用于保护等待处理的元素堆
	now         atomic.Int64          // 当前时间，用于计算元素的延迟时间
	config      *DelayingQConfig      // 配置，包含了队列的配置和回调接口
}

// 创建 DelayingQueue 实例
// Create a DelayingQueue instance
func newDelayingQueue(conf *DelayingQConfig, queue QInterface) *DelayingQ {
	if queue == nil {
		return nil
	}

	// 验证配置是否有效
	// Verify the configuration is valid
	conf = isDelayingQConfigValid(conf)
	conf.QConfig.callback = conf.callback

	// 初始化 DelayingQ 结构体
	// Initialize the DelayingQ structure
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

	// 创建一个新的上下文和取消函数
	// Create a new context and cancel function
	q.ctx, q.cancel = context.WithCancel(context.Background())

	// 启动两个 goroutine，一个用于处理队列中的元素，一个用于同步当前时间
	// Start two goroutines, one for processing elements in the queue, and one for synchronizing the current time
	q.wg.Add(2)
	go q.loop()
	go q.syncNow()

	// 返回 DelayingQ 实例
	// Return the DelayingQ instance
	return q
}

// 创建一个 DelayingQueue 实例
// Create a new DelayingQueue instance
func NewDelayingQueue(conf *DelayingQConfig) *DelayingQ {
	// 验证并获取有效的配置
	// Validate and get the valid configuration
	conf = isDelayingQConfigValid(conf)

	// 设置回调函数
	// Set the callback function
	conf.QConfig.callback = conf.callback

	// 创建一个新的延迟队列实例，并返回
	// Create a new instance of the delaying queue and return it
	return newDelayingQueue(conf, NewQueue(&conf.QConfig))
}

// 创建一个 DelayingQueue 实例, 使用自定义 Queue (实现了 Q 接口)
// Create a new DelayingQueue instance, using a custom Queue (which implements the Q interface)
func NewDelayingQueueWithCustomQueue(conf *DelayingQConfig, queue QInterface) *DelayingQ {
	// 验证并获取有效的配置
	// Validate and get the valid configuration
	conf = isDelayingQConfigValid(conf)

	// 设置回调函数
	// Set the callback function
	conf.QConfig.callback = conf.callback

	// 使用自定义的队列创建一个新的延迟队列实例，并返回
	// Create a new instance of the delaying queue using the custom queue and return it
	return newDelayingQueue(conf, queue)
}

// 创建一个默认的 DelayingQueue 对象
// Create a new default DelayingQueue object
func DefaultDelayingQueue() DelayingQInterface {
	// 使用默认的配置创建一个新的延迟队列实例，并返回
	// Create a new instance of the delaying queue using the default configuration and return it
	return NewDelayingQueue(nil)
}

// AddAfter 方法将元素添加到队列中，在延迟一段时间后处理
// The AddAfter method adds an element to the queue and processes it after a specified delay
func (q *DelayingQ) AddAfter(element any, delay time.Duration) error {
	// 如果队列已经关闭，返回 ErrorQueueClosed 错误
	// If the queue is already closed, return the ErrorQueueClosed error
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
	// 设置元素的数据和值
	// Set the data and value of the element
	elem.SetData(element)
	elem.SetValue(time.Now().Add(delay).UnixMilli())

	// 添加到堆中
	// Add to the heap
	q.wlock.Lock()
	q.waiting.Push(elem)
	q.wlock.Unlock()

	// 执行回调函数
	// Execute the callback function
	q.config.callback.OnAddAfter(element, delay)

	// 返回 nil，表示没有错误
	// Return nil, indicating no error
	return nil
}

// syncNow 方法同步当前的时间
// The syncNow method synchronizes the current time
func (q *DelayingQ) syncNow() {
	// 创建一个心跳计时器
	// Create a heartbeat timer
	heartbeat := time.NewTicker(time.Millisecond * defaultQueueHeartbeat)

	// 在函数结束时，停止心跳计时器，并通知等待组一个操作已完成
	// At the end of the function, stop the heartbeat timer and notify the wait group that an operation has been completed
	defer func() {
		q.wg.Done()
		heartbeat.Stop()
	}()

	// 循环同步当前时间
	// Loop to sync current time
	for {
		select {
		// 如果上下文已完成，结束循环
		// If the context is done, end the loop
		case <-q.ctx.Done():
			return
		// 如果收到心跳信号，更新当前时间
		// If a heartbeat signal is received, update the current time
		case <-heartbeat.C:
			q.now.Store(time.Now().UnixMilli())
		}
	}
}

// loop 方法用于循环处理堆中的元素
// The loop method is used to process elements in the heap in a loop
func (q *DelayingQ) loop() {
	// 创建一个心跳计时器
	// Create a heartbeat timer
	heartbeat := time.NewTicker(time.Millisecond * defaultQueueHeartbeat)

	// 在函数结束时，停止心跳计时器，并通知等待组一个操作已完成
	// At the end of the function, stop the heartbeat timer and notify the wait group that an operation has been completed
	defer func() {
		q.wg.Done()
		heartbeat.Stop()
	}()

	// 循环处理堆中的元素
	// Loop to process elements in the heap
	for {
		select {
		// 如果上下文已完成，结束循环
		// If the context is done, end the loop
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

				// 如果堆顶元素的时间小于当前时间，意味着对象已经超时
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
						// 重置元素的值，1500ms 后再次处理元素
						// Reset the value of the element, process the element again after 1500ms
						elem.SetValue(q.now.Load() + defaultQueueHeartbeat*3)

						// 将元素重新添加到堆中
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

// Stop 方法用于关闭队列
// The Stop method is used to close the queue
func (q *DelayingQ) Stop() {
	// 调用 QInterface 的 Stop 方法，关闭队列
	// Call the Stop method of QInterface to close the queue
	q.QInterface.Stop()

	// 使用 sync.Once 确保以下操作只执行一次
	// Use sync.Once to ensure that the following operations are only performed once
	q.once.Do(func() {
		// 调用 context 的 cancel 函数，取消所有基于该 context 的操作
		// Call the cancel function of context to cancel all operations based on this context
		q.cancel()

		// 调用 sync.WaitGroup 的 Wait 方法，等待所有 goroutine 完成
		// Call the Wait method of sync.WaitGroup to wait for all goroutines to complete
		q.wg.Wait()

		// 调用 heap 的 Reset 方法，重置等待处理的元素堆
		// Call the Reset method of heap to reset the heap of elements waiting to be processed
		q.waiting.Reset()
	})
}
