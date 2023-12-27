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
	Interface
	// AddWeight 添加一个元素，指定权重，并在一段时间内排序
	// Add an element with specified weight and sort it within a period of time
	AddWeight(element any, weight int) error
}

// 优先级队列的回调接口
// Priority queue callback interface
type PriorityCallback interface {
	Callback
	// OnAddWeight 添加元素后的回调
	// Callback after adding an element
	OnAddWeight(element any, weight int)
}

// 优先级队列的配置
// Priority queue configuration
type PriorityQConfig struct {
	QConfig
	cb  PriorityCallback
	win int64
}

// 创建一个优先级队列的配置
// Create a new priority queue configuration
func NewPriorityQConfig() *PriorityQConfig {
	return &PriorityQConfig{}
}

// 设置优先级队列的回调接口
// Set the callback interface for the priority queue
func (c *PriorityQConfig) WithCallback(cb PriorityCallback) *PriorityQConfig {
	c.cb = cb
	return c
}

// 设置优先级队列的排序窗口大小
// Set the sort window size for the priority queue
func (c *PriorityQConfig) WithWindow(win int64) *PriorityQConfig {
	c.win = win
	return c
}

// 验证队列的配置是否有效
// Verify that the queue configuration is valid
func isPriorityQConfigValid(conf *PriorityQConfig) *PriorityQConfig {
	if conf == nil {
		conf = NewPriorityQConfig()
		conf.WithCallback(emptyCallback{}).WithWindow(defaultQueueSortWin)
	} else {
		if conf.cb == nil {
			conf.cb = emptyCallback{}
		}
		if conf.win <= defaultQueueSortWin {
			conf.win = defaultQueueSortWin
		}
	}

	return conf
}

type PriorityQ struct {
	*Q
	waiting *heap.Heap
	elepool *heap.HeapElementPool
	stopCtx context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	once    sync.Once
	lock    *sync.Mutex
	config  *PriorityQConfig
}

// 创建一个 PriorityQueue 实例, 使用自定义 Queue (实现了 Q 接口)
// Create a new PriorityQueue config, use custom Queue (implement Q interface)
func NewPriorityQueueWithCustomQueue(conf *PriorityQConfig, queue *Q) *PriorityQ {
	if queue == nil {
		return nil
	}

	conf = isPriorityQConfigValid(conf)
	conf.QConfig.cb = conf.cb

	q := &PriorityQ{
		Q:       queue,
		waiting: heap.NewHeap(),
		elepool: heap.NewHeapElementPool(),
		wg:      sync.WaitGroup{},
		lock:    &sync.Mutex{},
		once:    sync.Once{},
		config:  conf,
	}

	q.lock = q.Q.lock
	q.stopCtx, q.cancel = context.WithCancel(context.Background())

	q.wg.Add(1)
	go q.loop()

	return q
}

// 创建一个 PriorityQueue 实例
// Create a new PriorityQueue config
func NewPriorityQueue(conf *PriorityQConfig) *PriorityQ {
	conf = isPriorityQConfigValid(conf)
	conf.QConfig.cb = conf.cb
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
	if q.IsClosed() {
		return ErrorQueueClosed
	}

	if weight <= 0 {
		return q.Add(element)
	}

	ele := q.elepool.Get()
	ele.SetData(element)
	ele.SetValue(int64(weight))

	q.lock.Lock()
	q.waiting.Push(ele)
	q.lock.Unlock()

	q.config.cb.OnAddWeight(element, weight)

	return nil
}

// 循环处理 Heap 中的元素
// Loop to process elements in Heap
func (q *PriorityQ) loop() {
	heartbeat := time.NewTicker(time.Duration(q.config.win) * time.Millisecond)

	defer func() {
		q.wg.Done()
		heartbeat.Stop()
	}()

	for {
		select {
		case <-q.stopCtx.Done():
			return

		// 每隔一段时间，处理一次 Heap 中的元素。 Process the elements in the Heap every once in a while.
		case <-heartbeat.C:
			q.lock.Lock()
			// 获取 Heap 中的元素。 Get the elements in the Heap.
			s0 := q.waiting.Slice()
			// 重置 Heap。 Reset the Heap.
			q.waiting.Reset()
			q.lock.Unlock()

			// 将 Heap 中的元素添加到 Queue 中。 Add the elements in the Heap to the Queue.
			if len(s0) > 0 {
				// 将 s0 中的元素添加到 Queue 中。 Add the elements in s0 to the Queue.
				for i := 0; i < len(s0); i++ {
					ele := s0[i]
					// 如果添加失败，则将元素重新添加到 Heap 中。 If the addition fails, the element is re-added to the Heap.
					if err := q.Add(ele.Data()); err != nil {
						q.lock.Lock()
						// 将元素重新添加到 Heap 中。 Re-add the element to the Heap.
						q.waiting.Push(ele)
						q.lock.Unlock()
					} else {
						// 释放元素 Free element
						q.elepool.Put(ele)
					}
				}
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
