package workqueue

import (
	"context"
	"sync"
	"time"

	"github.com/shengyanli1982/workqueue/internal/stl/heap"
)

// 优先级队列方法接口
// Priority queue interface
type PriorityInterface interface {
	// 继承 Queue 接口
	// Inherit Queue
	Interface

	// AddWeight 添加一个元素，指定权重，并在一段时间内排序
	// Add an element with specified weight and sort it within a period of time
	AddWeight(element any, weight int) error
}

// 优先级队列的回调接口
// Priority queue callback interface
type PriorityCallback interface {
	// 继承 Callback 接口
	// Inherit Callback
	Callback

	// OnAddWeight 添加元素后的回调
	// Callback after adding an element
	OnAddWeight(element any, weight int)
}

// 优先级队列的配置
// Priority queue configuration
type PriorityQConfig struct {
	QConfig
	callback PriorityCallback
	sortwin  int64
}

// 创建一个优先级队列的配置
// Create a new priority queue configuration
func NewPriorityQConfig() *PriorityQConfig {
	return &PriorityQConfig{}
}

// 设置优先级队列的回调接口
// Set the callback interface for the priority queue
func (c *PriorityQConfig) WithCallback(cb PriorityCallback) *PriorityQConfig {
	c.callback = cb
	return c
}

// 设置优先级队列的排序窗口大小
// Set the sort window size for the priority queue
func (c *PriorityQConfig) WithWindow(win int64) *PriorityQConfig {
	c.sortwin = win
	return c
}

// 验证队列的配置是否有效
// Verify that the queue configuration is valid
func isPriorityQConfigValid(conf *PriorityQConfig) *PriorityQConfig {
	if conf == nil {
		conf = NewPriorityQConfig()
		conf.WithCallback(emptyCallback{}).WithWindow(defaultQueueSortWin)
	} else {
		if conf.callback == nil {
			conf.callback = emptyCallback{}
		}
		if conf.sortwin <= defaultQueueSortWin {
			conf.sortwin = defaultQueueSortWin
		}
	}

	return conf
}

// PriorityQ 是 PriorityQueue 的实现
// PriorityQ is the implementation of PriorityQueue
type PriorityQ struct {
	*Q
	waiting     *heap.Heap
	elementpool *heap.HeapElementPool
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	once        sync.Once
	lock        *sync.Mutex
	config      *PriorityQConfig
}

// 创建一个 PriorityQueue 实例, 使用自定义 Queue (实现了 Q 接口)
// Create a new PriorityQueue config, use custom Queue (implement Q interface)
func NewPriorityQueueWithCustomQueue(conf *PriorityQConfig, queue *Q) *PriorityQ {
	if queue == nil {
		return nil
	}

	conf = isPriorityQConfigValid(conf)
	conf.QConfig.callback = conf.callback

	q := &PriorityQ{
		Q:           queue,
		waiting:     heap.NewHeap(),
		elementpool: heap.NewHeapElementPool(),
		wg:          sync.WaitGroup{},
		lock:        &sync.Mutex{},
		once:        sync.Once{},
		config:      conf,
	}

	q.lock = q.Q.lock
	q.ctx, q.cancel = context.WithCancel(context.Background())

	q.wg.Add(1)
	go q.loop()

	return q
}

// 创建一个 PriorityQueue 实例
// Create a new PriorityQueue config
func NewPriorityQueue(conf *PriorityQConfig) *PriorityQ {
	conf = isPriorityQConfigValid(conf)
	conf.QConfig.callback = conf.callback
	return NewPriorityQueueWithCustomQueue(conf, NewQueue(&conf.QConfig))
}

// 创建一个默认的 PriorityQueue 对象
// Create a new default PriorityQueue object
func DefaultPriorityQueue() PriorityInterface {
	return NewPriorityQueue(nil)
}

// AddWeight 添加一个元素，指定权重，并在一段时间内排序
// Add an element, add it use weight and sort it in a period of time
func (q *PriorityQ) AddWeight(element any, weight int) error {
	// 如果队列已经关闭，返回 ErrorQueueClosed 错误
	// If the queue is already closed, return ErrorQueueClosed
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
	q.lock.Lock()
	q.waiting.Push(elem)
	q.lock.Unlock()

	// 回调
	// Callback
	q.config.callback.OnAddWeight(element, weight)

	return nil
}

// 循环处理 Heap 中的元素
// Loop to process elements in Heap
func (q *PriorityQ) loop() {
	// 心跳
	// Heartbeat
	heartbeat := time.NewTicker(time.Duration(q.config.sortwin) * time.Millisecond)

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

		// 每隔一段时间，处理一次 Heap 中的元素
		// Process the elements in the Heap every once in a while.
		case <-heartbeat.C:
			q.lock.Lock()
			// 获取 Heap 中的元素
			// Get the elements in the Heap.
			elems := q.waiting.Slice()

			// 重置 Heap
			// Reset the Heap.
			q.waiting.Reset()
			q.lock.Unlock()

			// 将 Heap 中的元素添加到 Queue 中
			// Add the elements in the Heap to the Queue.
			if len(elems) > 0 {
				// 创建一个临时的切片，用于存储需要重新添加到 Heap 中的元素
				// Create a temporary slice to store the elements that need to be re-added to the Heap.
				var reAddElems []*heap.Element

				// 将 s0 中的元素添加到 Queue 中
				// Add the elements in s0 to the Queue.
				for i := 0; i < len(elems); i++ {
					elem := elems[i]
					// 如果添加失败，则将元素添加到临时切片中
					// If the addition fails, add the element to the temporary slice.
					if err := q.Add(elem.Data()); err != nil {
						reAddElems = append(reAddElems, elem)
					} else {
						// 释放元素 Free element
						q.elementpool.Put(elem)
					}
				}

				// 将需要重新添加的元素重新添加到 Heap 中
				// Re-add the elements that need to be re-added to the Heap.
				q.lock.Lock()
				for i := 0; i < len(reAddElems); i++ {
					q.waiting.Push(reAddElems[i])
				}
				q.lock.Unlock()
			}
		}
	}
}

// Close 关闭 Queue
// Close Queue
func (q *PriorityQ) Stop() {
	q.Q.Stop()
	q.once.Do(func() {
		q.cancel()
		q.wg.Wait()
		q.waiting.Reset()
	})
}
