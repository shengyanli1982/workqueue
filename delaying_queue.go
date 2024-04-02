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
	// 继承 Queue 接口，包含了队列的一些基本操作
	// Inherits Queue interface, includes some basic operations of the queue
	QInterface

	// AddAfter 方法用于添加一个元素到队列中，该元素会在延迟一段时间后再被处理
	// The AddAfter method is used to add an element to the queue, which will be processed after a delay
	AddAfter(element any, delay time.Duration) error
}

// DelayingQCallback 是 Queue 的回调接口的延迟版本
// DelayingQCallback is the delayed version of the Queue callback interface
type DelayingQCallback interface {
	// 继承 Callback 接口，包含了队列的一些基本操作的回调
	// Inherits Callback interface, includes callbacks for some basic operations of the queue
	QCallback

	// OnAddAfter 是添加元素后的回调，参数 any 是添加的元素，time.Duration 是元素的延迟时间
	// OnAddAfter is the callback after adding an element, the parameter any is the added element, and time.Duration is the delay time of the element
	OnAddAfter(any, time.Duration)
}

// DelayingQConfig 结构体定义了延迟队列的配置信息
// The DelayingQConfig struct defines the configuration information of the delayed queue
type DelayingQConfig struct {
	// QConfig 是队列的基本配置，包含了队列的一些通用配置信息
	// QConfig is the basic configuration of the queue, containing some common configuration information of the queue
	QConfig

	// callback 是一个延迟队列回调接口，用于实现队列元素的处理
	// callback is a delayed queue callback interface, used to implement the processing of queue elements
	callback DelayingQCallback
}

// NewDelayingQConfig 函数用于创建一个新的 DelayingQConfig 实例
// The NewDelayingQConfig function is used to create a new DelayingQConfig instance
func NewDelayingQConfig() *DelayingQConfig {
	// 返回一个新的 DelayingQConfig 实例
	// Return a new DelayingQConfig instance
	return &DelayingQConfig{}
}

// WithCallback 方法用于设置 DelayingQConfig 的回调接口
// The WithCallback method is used to set the callback interface of DelayingQConfig
func (c *DelayingQConfig) WithCallback(cb DelayingQCallback) *DelayingQConfig {
	// 设置回调接口
	// Set the callback interface
	c.callback = cb

	// 返回 DelayingQConfig 实例
	// Return the DelayingQConfig instance
	return c
}

// isDelayingQConfigValid 函数用于验证 DelayingQConfig 的配置是否有效
// The isDelayingQConfigValid function is used to verify whether the configuration of DelayingQConfig is valid
func isDelayingQConfigValid(conf *DelayingQConfig) *DelayingQConfig {
	// 如果配置为空，则创建一个新的配置，并设置一个空的回调接口
	// If the configuration is nil, create a new configuration and set an empty callback interface
	if conf == nil {
		conf = NewDelayingQConfig().WithCallback(newEmptyCallback())
	} else {
		// 如果回调接口为空，则设置一个空的回调接口
		// If the callback interface is nil, set an empty callback interface
		if conf.callback == nil {
			conf.callback = newEmptyCallback()
		}
	}

	// 返回经过验证和可能的修改后的配置
	// Return the configuration after verification and possible modification
	return conf
}

// DelayingQ 结构体定义了一个延迟队列的数据结构
// The DelayingQ struct defines a data structure for a delayed queue
type DelayingQ struct {
	// QInterface 是队列的接口，定义了队列的基本操作
	// QInterface is the interface of the queue, defining the basic operations of the queue
	QInterface

	// waiting 是一个堆结构，用于存储等待处理的元素
	// waiting is a heap structure used to store elements waiting for processing
	waiting *heap.Heap

	// elementpool 是一个元素池，用于存储等待处理的元素
	// elementpool is an element pool used to store elements waiting for processing
	elementpool *heap.HeapElementPool

	// ctx 是一个上下文，用于控制并发操作
	// ctx is a context used to control concurrent operations
	ctx context.Context

	// cancel 是一个取消函数，用于取消上下文
	// cancel is a cancel function used to cancel the context
	cancel context.CancelFunc

	// wg 是一个同步等待组，用于等待并发操作完成
	// wg is a sync.WaitGroup used to wait for concurrent operations to complete
	wg sync.WaitGroup

	// once 是一个 sync.Once 对象，用于确保某个操作只执行一次
	// once is a sync.Once object used to ensure that an operation is performed only once
	once sync.Once

	// wlock 是一个互斥锁，用于保护等待处理的元素堆
	// wlock is a mutex used to protect the heap of elements waiting for processing
	wlock *sync.Mutex

	// now 是一个原子整数，用于存储当前时间，计算元素的延迟时间
	// now is an atomic integer used to store the current time and calculate the delay time of elements
	now atomic.Int64

	// config 是一个指向 DelayingQConfig 结构体的指针，用于存储队列的配置信息和回调接口
	// config is a pointer to a DelayingQConfig struct used to store the configuration information and callback interface of the queue
	config *DelayingQConfig
}

// newDelayingQueue 函数用于创建一个新的 DelayingQueue 实例
// The newDelayingQueue function is used to create a new DelayingQueue instance
func newDelayingQueue(conf *DelayingQConfig, queue QInterface) *DelayingQ {
	// 如果传入的队列为空，则直接返回 nil
	// If the passed in queue is nil, return nil directly
	if queue == nil {
		return nil
	}

	// 验证并修正传入的配置，确保配置是有效的
	// Verify and correct the passed in configuration to ensure that the configuration is valid
	conf = isDelayingQConfigValid(conf)

	// 将回调接口设置到队列的配置中
	// Set the callback interface to the configuration of the queue
	conf.QConfig.callback = conf.callback

	// 创建一个新的 DelayingQ 实例
	// Create a new DelayingQ instance
	q := &DelayingQ{
		// 设置队列接口
		// Set the queue interface
		QInterface: queue,

		// 初始化等待队列
		// Initialize the waiting queue
		waiting: heap.NewHeap(),

		// 初始化元素池
		// Initialize the element pool
		elementpool: heap.NewHeapElementPool(),

		// 初始化互斥锁
		// Initialize the mutex
		wlock: &sync.Mutex{},

		// 初始化等待组
		// Initialize the wait group
		wg: sync.WaitGroup{},

		// 初始化当前时间
		// Initialize the current time
		now: atomic.Int64{},

		// 初始化 sync.Once 对象
		// Initialize the sync.Once object
		once: sync.Once{},

		// 设置配置
		// Set the configuration
		config: conf,
	}

	// 创建一个新的上下文和取消函数
	// Create a new context and cancel function
	q.ctx, q.cancel = context.WithCancel(context.Background())

	// 增加等待组的计数器，表示有两个 goroutine 需要等待
	// Increase the counter of the wait group, indicating that there are two goroutines to wait for
	q.wg.Add(2)

	// 启动一个新的 goroutine 来处理队列中的元素
	// Start a new goroutine to process elements in the queue
	go q.loop()

	// 启动一个新的 goroutine 来同步当前时间
	// Start a new goroutine to synchronize the current time
	go q.syncNow()

	// 返回创建的 DelayingQ 实例
	// Return the created DelayingQ instance
	return q
}

// NewDelayingQueue 函数用于创建一个新的 DelayingQueue 实例
// The NewDelayingQueue function is used to create a new DelayingQueue instance
func NewDelayingQueue(conf *DelayingQConfig) *DelayingQ {
	// 验证并修正传入的配置，确保配置是有效的
	// Verify and correct the passed in configuration to ensure that the configuration is valid
	conf = isDelayingQConfigValid(conf)

	// 将回调接口设置到队列的配置中
	// Set the callback interface to the configuration of the queue
	conf.QConfig.callback = conf.callback

	// 使用配置和新的队列创建一个新的 DelayingQueue 实例
	// Create a new DelayingQueue instance with the configuration and a new queue
	return newDelayingQueue(conf, NewQueue(&conf.QConfig))
}

// NewDelayingQueueWithCustomQueue 函数用于创建一个新的 DelayingQueue 实例，使用自定义的 QInterface 队列
// The NewDelayingQueueWithCustomQueue function is used to create a new DelayingQueue instance, using a custom QInterface queue
func NewDelayingQueueWithCustomQueue(conf *DelayingQConfig, queue QInterface) *DelayingQ {
	// 验证并修正传入的配置，确保配置是有效的
	// Verify and correct the passed in configuration to ensure that the configuration is valid
	conf = isDelayingQConfigValid(conf)

	// 将回调接口设置到队列的配置中
	// Set the callback interface to the configuration of the queue
	conf.QConfig.callback = conf.callback

	// 使用配置和自定义的 QInterface 队列创建一个新的 DelayingQueue 实例
	// Create a new DelayingQueue instance with the configuration and the custom QInterface queue
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
// The AddAfter method adds an element to the queue and processes it after a delay
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

	// 设置元素的值为当前时间加上延迟时间，单位为毫秒
	// Set the value of the element to the current time plus the delay time, in milliseconds
	elem.SetValue(time.Now().Add(delay).UnixMilli())

	// 首先，我们需要锁定队列以防止并发操作
	// First, we need to lock the queue to prevent concurrent operations
	q.wlock.Lock()

	// 将元素添加到等待队列中
	// Add the element to the waiting queue
	q.waiting.Push(elem)

	// 添加完成后，解锁队列
	// After the addition is complete, unlock the queue
	q.wlock.Unlock()

	// 执行添加元素后的回调函数
	// Execute the callback function after adding the element
	q.config.callback.OnAddAfter(element, delay)

	// 返回 nil，表示没有错误
	// Return nil, indicating no error
	return nil
}

// syncNow 方法用于同步当前的时间
// The syncNow method is used to synchronize the current time
func (q *DelayingQ) syncNow() {
	// 创建一个心跳计时器，每隔 defaultQueueHeartbeat 毫秒就会发送一个信号
	// Create a heartbeat timer that sends a signal every defaultQueueHeartbeat milliseconds
	heartbeat := time.NewTicker(time.Millisecond * defaultQueueHeartbeat)

	// 在函数结束时，停止心跳计时器，并通知等待组一个操作已完成
	// At the end of the function, stop the heartbeat timer and notify the wait group that an operation has been completed
	defer func() {
		// 停止心跳计时器
		// Stop the heartbeat timer
		heartbeat.Stop()

		// 通知等待组一个操作已完成
		// Notify the wait group that an operation has been completed
		q.wg.Done()
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
			// 使用 atomic 包的 Store 方法安全地更新当前时间
			// Use the Store method of the atomic package to safely update the current time
			q.now.Store(time.Now().UnixMilli())
		}
	}
}

// loop 方法用于循环处理堆中的元素
// The loop method is used to process elements in the heap in a loop
func (q *DelayingQ) loop() {
	// 创建一个心跳计时器，每隔 defaultQueueHeartbeat 毫秒就会发送一个信号
	// Create a heartbeat timer that sends a signal every defaultQueueHeartbeat milliseconds
	heartbeat := time.NewTicker(time.Millisecond * defaultQueueHeartbeat)

	// 在函数结束时，停止心跳计时器，并通知等待组一个操作已完成
	// At the end of the function, stop the heartbeat timer and notify the wait group that an operation has been completed
	defer func() {
		// 通知等待组一个操作已完成
		// Notify the wait group that an operation has been completed
		q.wg.Done()

		// 停止心跳计时器
		// Stop the heartbeat timer
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
			// 锁定等待队列，防止并发操作
			// Lock the waiting queue to prevent concurrent operations
			q.wlock.Lock()

			// 如果等待队列中有元素
			// If there are elements in the waiting queue
			if q.waiting.Len() > 0 {
				// 获取堆顶元素
				// Get the top element of the heap
				elem := q.waiting.Head()

				// 如果元素的值小于等于当前时间，表示可以处理该元素
				// If the value of the element is less than or equal to the current time, it means that the element can be processed
				if elem.Value() <= q.now.Load() {
					// 从堆中移除元素
					// Remove the element from the heap
					_ = q.waiting.Pop()

					// 解锁等待队列
					// Unlock the waiting queue
					q.wlock.Unlock()

					// 将元素添加到队列中
					// Add the element to the queue
					if err := q.Add(elem.Data()); err != nil {
						// 如果添加失败，将元素重新添加到等待队列中，并设置新的值
						// If the addition fails, re-add the element to the waiting queue and set a new value
						q.wlock.Lock()

						// 设置新的值为当前时间加上 defaultQueueHeartbeat*3 毫秒
						// Set the new value to the current time plus defaultQueueHeartbeat*3 milliseconds
						elem.SetValue(q.now.Load() + defaultQueueHeartbeat*3)

						// 将元素重新添加到等待队列中
						// Re-add the element to the waiting queue
						q.waiting.Push(elem)

						// 解锁等待队列
						// Unlock the waiting queue
						q.wlock.Unlock()
					} else {
						// 如果添加成功，将元素放回对象池
						// If the addition is successful, put the element back into the object pool
						q.elementpool.Put(elem)
					}
				} else {
					// 如果元素的值大于当前时间，解锁等待队列，等待下一次心跳
					// If the value of the element is greater than the current time, unlock the waiting queue and wait for the next heartbeat
					q.wlock.Unlock()
				}
			} else {
				// 如果等待队列中没有元素，解锁等待队列，等待下一次心跳
				// If there are no elements in the waiting queue, unlock the waiting queue and wait for the next heartbeat
				q.wlock.Unlock()

				// 等待下一次心跳
				// Wait for the next heartbeat
				<-heartbeat.C
			}
		}
	}
}

// Stop 方法用于停止延迟队列的操作
// The Stop method is used to stop the operations of the delaying queue
func (q *DelayingQ) Stop() {
	// 调用 QInterface 的 Stop 方法，停止队列的操作
	// Call the Stop method of QInterface to stop the operations of the queue
	q.QInterface.Stop()

	// 使用 sync.Once 的 Do 方法确保以下操作只执行一次
	// Use the Do method of sync.Once to ensure that the following operations are only performed once
	q.once.Do(func() {
		// 调用 cancel 函数，取消上下文
		// Call the cancel function to cancel the context
		q.cancel()

		// 等待所有的 goroutine 结束
		// Wait for all goroutines to end
		q.wg.Wait()

		// 重置等待队列
		// Reset the waiting queue
		q.waiting.Reset()
	})
}
