package workqueue

import (
	"context"
	"sync"
	"time"

	"github.com/shengyanli1982/workqueue/internal/stl/heap"
)

// PriorityQInterface 是优先级队列的接口，它继承了 QInterface 接口，并添加了一个 AddWeight 方法
// PriorityQInterface is the interface of the priority queue, it inherits the QInterface interface and adds an AddWeight method
type PriorityQInterface interface {
	// 继承 Queue 接口
	// Inherit Queue
	QInterface

	// AddWeight 方法用于添加一个元素，并指定其权重，然后在一段时间内对队列进行排序
	// The AddWeight method is used to add an element and specify its weight, and then sort the queue within a certain period of time
	AddWeight(element any, weight int) error
}

// PriorityQCallback 是优先级队列的回调接口，它继承了 QCallback 接口，并添加了一个 OnAddWeight 方法
// PriorityQCallback is the callback interface of the priority queue, it inherits the QCallback interface and adds an OnAddWeight method
type PriorityQCallback interface {
	// 继承 Callback 接口
	// Inherit Callback
	QCallback

	// OnAddWeight 方法是添加元素后的回调
	// The OnAddWeight method is a callback after adding an element
	OnAddWeight(element any, weight int)
}

// PriorityQConfig 是优先级队列的配置，它包含了一个回调接口和一个排序窗口大小
// PriorityQConfig is the configuration of the priority queue, it contains a callback interface and a sort window size
type PriorityQConfig struct {
	QConfig
	callback PriorityQCallback
	sortwin  int64
}

// NewPriorityQConfig 方法用于创建一个新的优先级队列配置
// The NewPriorityQConfig method is used to create a new priority queue configuration
func NewPriorityQConfig() *PriorityQConfig {
	return &PriorityQConfig{}
}

// WithCallback 方法用于设置优先级队列的回调接口
// The WithCallback method is used to set the callback interface for the priority queue
func (c *PriorityQConfig) WithCallback(cb PriorityQCallback) *PriorityQConfig {
	c.callback = cb
	return c
}

// WithWindow 方法用于设置优先级队列的排序窗口大小
// The WithWindow method is used to set the sort window size for the priority queue
func (c *PriorityQConfig) WithWindow(win int64) *PriorityQConfig {
	c.sortwin = win
	return c
}

// isPriorityQConfigValid 函数用于验证队列的配置是否有效
// The isPriorityQConfigValid function is used to verify whether the queue configuration is valid
func isPriorityQConfigValid(conf *PriorityQConfig) *PriorityQConfig {
	if conf == nil {
		// 如果配置为空，则创建一个新的配置，并设置默认的回调和排序窗口大小
		// If the configuration is null, create a new configuration and set the default callback and sort window size
		conf = NewPriorityQConfig()
		conf.WithCallback(newEmptyCallback()).WithWindow(defaultQueueSortWin)
	} else {
		// 如果配置的回调为空，则设置默认的回调
		// If the callback of the configuration is null, set the default callback
		if conf.callback == nil {
			conf.callback = newEmptyCallback()
		}
		// 如果配置的排序窗口大小小于等于默认的排序窗口大小，则设置默认的排序窗口大小
		// If the sort window size of the configuration is less than or equal to the default sort window size, set the default sort window size
		if conf.sortwin <= defaultQueueSortWin {
			conf.sortwin = defaultQueueSortWin
		}
	}

	// 返回验证后的配置
	// Return the verified configuration
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
	if queue == nil {
		// 如果队列为空，则返回 nil
		// If the queue is null, return nil
		return nil
	}

	// 验证配置是否有效
	// Verify whether the configuration is valid
	conf = isPriorityQConfigValid(conf)

	// 设置回调
	// Set the callback
	conf.QConfig.callback = conf.callback

	// 创建一个 PriorityQ 实例
	// Create a PriorityQ instance
	q := &PriorityQ{
		QInterface:  queue,
		waiting:     heap.NewHeap(),
		elementpool: heap.NewHeapElementPool(),
		wlock:       &sync.Mutex{},
		wg:          sync.WaitGroup{},
		once:        sync.Once{},
		config:      conf,
	}

	// 创建一个新的上下文和取消函数
	// Create a new context and cancel function
	q.ctx, q.cancel = context.WithCancel(context.Background())

	// 增加等待组的计数
	// Increase the count of the wait group
	q.wg.Add(1)

	// 启动一个新的 goroutine 运行 loop 方法
	// Start a new goroutine to run the loop method
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
	elem.SetData(element)
	elem.SetValue(int64(weight))

	// 添加到堆中
	// Add to the heap
	q.wlock.Lock()
	q.waiting.Push(elem)
	q.wlock.Unlock()

	// 执行添加元素后的回调函数
	// Execute the callback function after adding the element
	q.config.callback.OnAddWeight(element, weight)

	return nil
}

// loop 方法用于循环处理 Heap 中的元素
// The loop method is used to process elements in the Heap in a loop
func (q *PriorityQ) loop() {
	// 创建一个心跳计时器，每隔一段时间发送一次信号
	// Create a heartbeat timer that sends a signal every once in a while
	heartbeat := time.NewTicker(time.Duration(q.config.sortwin) * time.Millisecond)

	// 在函数返回时，停止心跳计时器，并通知 WaitGroup 一个操作已经完成
	// When the function returns, stop the heartbeat timer and notify the WaitGroup that an operation has been completed
	defer func() {
		q.wg.Done()
		heartbeat.Stop()
	}()

	// 循环处理堆中的元素
	// Loop to process elements in the heap
	for {
		select {
		// 如果 context 已经被取消，就返回
		// If the context has been cancelled, return
		case <-q.ctx.Done():
			return

		// 每隔一段时间，处理一次 Heap 中的元素
		// Process the elements in the Heap every once in a while
		case <-heartbeat.C:
			q.wlock.Lock()

			// 获取 Heap 中的元素
			// Get the elements in the Heap
			elems := q.waiting.Slice()

			// 重置 Heap
			// Reset the Heap
			q.waiting.Reset()

			q.wlock.Unlock()

			// 将 Heap 中的元素添加到 Queue 中
			// Add the elements in the Heap to the Queue
			if len(elems) > 0 {
				// 创建一个临时的切片，用于存储需要重新添加到 Heap 中的元素
				// Create a temporary slice to store the elements that need to be re-added to the Heap
				var reAddElems []*heap.Element

				// 将 elems 中的元素添加到 Queue 中
				// Add the elements in elems to the Queue
				for i := 0; i < len(elems); i++ {
					elem := elems[i]
					// 如果添加失败，则将元素添加到临时切片中
					// If the addition fails, add the element to the temporary slice
					if err := q.Add(elem.Data()); err != nil {
						reAddElems = append(reAddElems, elem)
					} else {
						// 释放元素
						// Release the element
						q.elementpool.Put(elem)
					}
				}

				// 将需要重新添加的元素重新添加到 Heap 中
				// Re-add the elements that need to be re-added to the Heap
				q.wlock.Lock()
				for i := 0; i < len(reAddElems); i++ {
					q.waiting.Push(reAddElems[i])
				}

				q.wlock.Unlock()
			}
		}
	}
}

// Stop 方法用于关闭 PriorityQ 队列
// The Stop method is used to close the PriorityQ queue
func (q *PriorityQ) Stop() {
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
