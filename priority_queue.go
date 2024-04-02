package workqueue

import (
	"context"
	"sync"
	"time"

	"github.com/shengyanli1982/workqueue/internal/stl/heap"
)

// PriorityQInterface 接口定义了优先级队列的基本操作
// The PriorityQInterface interface defines the basic operations of a priority queue
type PriorityQInterface interface {
	// QInterface 是队列的基本接口，包含了队列的基本操作，如添加元素、获取元素等
	// QInterface is the basic interface of the queue, which includes basic operations of the queue, such as adding elements, getting elements, etc.
	QInterface

	// AddWeight 方法用于向队列中添加元素，并设置元素的权重，权重越小，优先级越高
	// The AddWeight method is used to add elements to the queue and set the weight of the elements. The smaller the weight, the higher the priority.
	AddWeight(element any, weight int) error
}

// PriorityQCallback 接口定义了优先级队列的回调函数
// The PriorityQCallback interface defines the callback functions of a priority queue
type PriorityQCallback interface {
	// QCallback 是队列的基本回调接口，包含了队列的基本回调函数，如添加元素后的回调、获取元素后的回调等
	// QCallback is the basic callback interface of the queue, which includes basic callback functions of the queue, such as the callback after adding elements, the callback after getting elements, etc.
	QCallback

	// OnAddWeight 方法是添加元素后的回调函数，当向队列中添加元素后，会调用此函数
	// The OnAddWeight method is the callback function after adding elements. After adding elements to the queue, this function will be called.
	OnAddWeight(element any, weight int)
}

// PriorityQConfig 结构体定义了优先级队列的配置信息
// The PriorityQConfig struct defines the configuration information of the priority queue
type PriorityQConfig struct {
	// QConfig 是队列的基本配置，包含了队列的一些通用配置信息
	// QConfig is the basic configuration of the queue, containing some common configuration information of the queue
	QConfig

	// callback 是一个优先级队列回调接口，用于实现队列元素的比较和处理
	// callback is a priority queue callback interface, used to implement the comparison and processing of queue elements
	callback PriorityQCallback

	// sortwin 是一个整数，表示排序窗口的大小，用于控制队列中元素的排序范围
	// sortwin is an integer that represents the size of the sort window, used to control the sort range of elements in the queue
	sortwin int64
}

// NewPriorityQConfig 函数用于创建一个新的 PriorityQConfig 实例
// The NewPriorityQConfig function is used to create a new instance of PriorityQConfig
func NewPriorityQConfig() *PriorityQConfig {
	// 返回一个新的 PriorityQConfig 实例
	// Return a new instance of PriorityQConfig
	return &PriorityQConfig{}
}

// WithCallback 方法用于设置 PriorityQConfig 的回调函数
// The WithCallback method is used to set the callback function of PriorityQConfig
func (c *PriorityQConfig) WithCallback(cb PriorityQCallback) *PriorityQConfig {
	// 设置回调函数
	// Set the callback function
	c.callback = cb

	// 返回 PriorityQConfig 实例，以便链式调用
	// Return the instance of PriorityQConfig for chain call
	return c
}

// WithWindow 方法用于设置 PriorityQConfig 的窗口大小
// The WithWindow method is used to set the window size of PriorityQConfig
func (c *PriorityQConfig) WithWindow(win int64) *PriorityQConfig {
	// 设置窗口大小
	// Set the window size
	c.sortwin = win

	// 返回 PriorityQConfig 实例，以便链式调用
	// Return the instance of PriorityQConfig for chain call
	return c
}

// isPriorityQConfigValid 函数用于检查 PriorityQConfig 的配置是否有效
// The isPriorityQConfigValid function is used to check whether the configuration of PriorityQConfig is valid
func isPriorityQConfigValid(conf *PriorityQConfig) *PriorityQConfig {
	// 如果 conf 为 nil，则创建一个新的 PriorityQConfig 实例，并设置默认的回调函数和窗口大小
	// If conf is nil, create a new instance of PriorityQConfig and set the default callback function and window size
	if conf == nil {
		conf = NewPriorityQConfig().WithCallback(newEmptyCallback()).WithWindow(defaultQueueSortWin)
	} else {
		// 如果 conf 的回调函数为 nil，则设置默认的回调函数
		// If the callback function of conf is nil, set the default callback function
		if conf.callback == nil {
			conf.callback = newEmptyCallback()
		}

		// 如果 conf 的窗口大小小于等于默认的窗口大小，则设置默认的窗口大小
		// If the window size of conf is less than or equal to the default window size, set the default window size
		if conf.sortwin <= defaultQueueSortWin {
			conf.sortwin = defaultQueueSortWin
		}
	}

	// 返回有效的 PriorityQConfig 实例
	// Return the valid instance of PriorityQConfig
	return conf
}

// 优先级队列数据结构
// Priority queue data structure
type PriorityQ struct {
	// 继承 Queue 接口
	// Inherit Queue interface
	QInterface

	// 等待处理的元素堆
	// Heap of elements waiting to be processed
	waiting *heap.Heap

	// 元素池
	// Element pool
	elementpool *heap.HeapElementPool

	// 上下文，用于控制队列的生命周期
	// Context, used to control the lifecycle of the queue
	ctx context.Context

	// 取消函数，用于取消上下文
	// Cancel function, used to cancel the context
	cancel context.CancelFunc

	// 等待组，用于等待所有 goroutine 完成
	// Wait group, used to wait for all goroutines to complete
	wg sync.WaitGroup

	// 用于确保某些操作只执行一次
	// Used to ensure that certain operations are only performed once
	once sync.Once

	// 用于保护等待处理的元素堆的锁
	// Lock used to protect the heap of elements waiting to be processed
	wlock *sync.Mutex

	// 队列的配置
	// Configuration of the queue
	config *PriorityQConfig
}

// newPriorityQueue 函数用于创建一个 PriorityQ 实例
// The newPriorityQueue function is used to create a PriorityQ instance
func newPriorityQueue(conf *PriorityQConfig, queue QInterface) *PriorityQ {
	// 如果传入的队列为空，则返回 nil
	// If the passed in queue is nil, return nil
	if queue == nil {
		return nil
	}

	// 检查 PriorityQConfig 的配置是否有效，如果无效则使用默认配置
	// Check if the configuration of PriorityQConfig is valid, if not, use the default configuration
	conf = isPriorityQConfigValid(conf)

	// 将 PriorityQConfig 的回调函数设置为 QConfig 的回调函数
	// Set the callback function of PriorityQConfig as the callback function of QConfig
	conf.QConfig.callback = conf.callback

	// 创建 PriorityQ 实例
	// Create a PriorityQ instance
	q := &PriorityQ{
		// 设置队列接口
		// Set the queue interface
		QInterface: queue,

		// 创建一个新的堆，用于存储等待处理的元素
		// Create a new heap for storing elements waiting to be processed
		waiting: heap.NewHeap(),

		// 创建一个新的堆元素池，用于存储堆元素
		// Create a new heap element pool for storing heap elements
		elementpool: heap.NewHeapElementPool(),

		// 创建一个新的互斥锁，用于保护等待处理的元素堆
		// Create a new mutex for protecting the heap of elements waiting to be processed
		wlock: &sync.Mutex{},

		// 创建一个新的等待组，用于等待所有 goroutine 完成
		// Create a new wait group for waiting for all goroutines to complete
		wg: sync.WaitGroup{},

		// 创建一个新的 sync.Once 实例，用于确保某些操作只执行一次
		// Create a new sync.Once instance to ensure that certain operations are only performed once
		once: sync.Once{},

		// 设置 PriorityQ 的配置
		// Set the configuration of PriorityQ
		config: conf,
	}

	// 创建一个新的 context，并设置 cancel 函数，用于取消所有基于该 context 的操作
	// Create a new context and set the cancel function to cancel all operations based on this context
	q.ctx, q.cancel = context.WithCancel(context.Background())

	// 增加等待组的计数
	// Increase the count of the wait group
	q.wg.Add(1)

	// 启动一个新的 goroutine，执行 loop 方法，用于处理等待队列中的元素
	// Start a new goroutine to execute the loop method for processing elements in the waiting queue
	go q.loop()

	// 返回创建的 PriorityQ 实例
	// Return the created PriorityQ instance
	return q
}

// NewPriorityQueue 函数用于创建一个 PriorityQueue 实例
// The NewPriorityQueue function is used to create a PriorityQueue instance
func NewPriorityQueue(conf *PriorityQConfig) *PriorityQ {
	// 验证配置是否有效并设置回调
	// Verify whether the configuration is valid and set the callback
	conf = isPriorityQConfigValid(conf)

	// 将回调接口设置到队列的配置中
	// Set the callback interface to the configuration of the queue
	conf.QConfig.callback = conf.callback

	// 创建一个 PriorityQueue 实例
	// Create a PriorityQueue instance
	return newPriorityQueue(conf, NewQueue(&conf.QConfig))
}

// NewPriorityQueueWithCustomQueue 函数用于创建一个 PriorityQueue 实例，并使用自定义的 Queue
// The NewPriorityQueueWithCustomQueue function is used to create a PriorityQueue instance and use a custom Queue
func NewPriorityQueueWithCustomQueue(conf *PriorityQConfig, queue QInterface) *PriorityQ {
	// 验证配置是否有效并设置回调
	// Verify whether the configuration is valid and set the callback
	conf = isPriorityQConfigValid(conf)

	// 将回调接口设置到队列的配置中
	// Set the callback interface to the configuration of the queue
	conf.QConfig.callback = conf.callback

	// 创建一个 PriorityQueue 实例
	// Create a PriorityQueue instance
	return newPriorityQueue(conf, queue)
}

// DefaultPriorityQueue 函数用于创建一个默认的 PriorityQueue 对象
// The DefaultPriorityQueue function is used to create a default PriorityQueue object
func DefaultPriorityQueue() PriorityQInterface {
	// 创建一个 PriorityQueue 实例，配置为 nil
	// Create a PriorityQueue instance with the configuration set to nil
	return NewPriorityQueue(nil)
}

// AddWeight 方法用于添加一个元素到优先级队列中，并指定其权重
// The AddWeight method is used to add an element to the priority queue and specify its weight
func (q *PriorityQ) AddWeight(element any, weight int) error {
	// 如果队列已经关闭，返回 ErrorQueueClosed 错误
	// If the queue is already closed, return the ErrorQueueClosed error
	if q.IsClosed() {
		return ErrorQueueClosed
	}

	// 如果权重小于等于 0，则直接添加到 Queue 中
	// If the weight is less than or equal to 0, add it directly to the Queue
	if weight <= 0 {
		return q.Add(element)
	}
	// 从对象池中获取一个元素
	// Get an element from the object pool
	elem := q.elementpool.Get()

	// 使用 SetData 方法设置元素的数据
	// Use the SetData method to set the data of the element
	elem.SetData(element)

	// 使用 SetValue 方法设置元素的权重，权重越小，优先级越高
	// Use the SetValue method to set the weight of the element. The smaller the weight, the higher the priority
	elem.SetValue(int64(weight))

	// 使用互斥锁保护等待处理的元素堆，防止并发操作导致的数据不一致
	// Use a mutex to protect the heap of elements waiting to be processed, to prevent data inconsistency caused by concurrent operations
	q.wlock.Lock()

	// 使用 Push 方法将元素添加到等待处理的元素堆中
	// Use the Push method to add the element to the heap of elements waiting to be processed
	q.waiting.Push(elem)

	// 解锁互斥锁，允许其他 goroutine 访问等待处理的元素堆
	// Unlock the mutex to allow other goroutines to access the heap of elements waiting to be processed
	q.wlock.Unlock()

	// 执行添加元素后的回调函数
	// Execute the callback function after adding the element
	q.config.callback.OnAddWeight(element, weight)

	// 返回 nil，表示添加元素成功
	// Return nil, indicating that the element was added successfully
	return nil
}

// loop 方法用于循环处理优先队列中的元素
// The loop method is used to process elements in the priority queue in a loop
func (q *PriorityQ) loop() {
	// 创建一个心跳计时器，每隔 q.config.sortwin 毫秒就会发送一个信号
	// Create a heartbeat timer that sends a signal every q.config.sortwin milliseconds
	heartbeat := time.NewTicker(time.Duration(q.config.sortwin) * time.Millisecond)

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

	// 循环处理优先队列中的元素
	// Loop to process elements in the priority queue
	for {
		select {
		// 如果上下文已完成，结束循环
		// If the context is done, end the loop
		case <-q.ctx.Done():
			return

		// 如果收到心跳信号
		// If a heartbeat signal is received
		case <-heartbeat.C:
			// 锁定等待队列，防止并发操作
			// Lock the waiting queue to prevent concurrent operations
			q.wlock.Lock()

			// 获取等待队列中的所有元素
			// Get all elements in the waiting queue
			elems := q.waiting.Slice()

			// 重置等待队列
			// Reset the waiting queue
			q.waiting.Reset()

			// 解锁等待队列
			// Unlock the waiting queue
			q.wlock.Unlock()

			// 如果等待队列中有元素
			// If there are elements in the waiting queue
			if len(elems) > 0 {
				// 创建一个新的切片，用于存储需要重新添加到等待队列中的元素
				// Create a new slice to store the elements that need to be re-added to the waiting queue
				var reAddedElements []*heap.Element

				// 遍历等待队列中的所有元素
				// Traverse all elements in the waiting queue
				for i := 0; i < len(elems); i++ {
					elem := elems[i]

					// 尝试将元素添加到优先队列中
					// Try to add the element to the priority queue
					if err := q.Add(elem.Data()); err != nil {
						// 如果添加失败，将元素添加到 reAddedElements 切片中
						// If the addition fails, add the element to the reAddedElements slice
						reAddedElements = append(reAddedElements, elem)
					} else {
						// 如果添加成功，将元素放回对象池
						// If the addition is successful, put the element back into the object pool
						q.elementpool.Put(elem)
					}
				}

				// 锁定等待队列，防止并发操作
				// Lock the waiting queue to prevent concurrent operations
				q.wlock.Lock()

				// 将 reAddedElements 切片中的所有元素重新添加到等待队列中
				// Re-add all elements in the reAddedElements slice to the waiting queue
				for i := 0; i < len(reAddedElements); i++ {
					q.waiting.Push(reAddedElements[i])
				}

				// 解锁等待队列
				// Unlock the waiting queue
				q.wlock.Unlock()
			}
		}
	}
}

// Stop 方法用于停止优先队列的操作
// The Stop method is used to stop the operations of the priority queue
func (q *PriorityQ) Stop() {
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
