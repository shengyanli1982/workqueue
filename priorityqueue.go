package workqueue

import (
	"context"
	"math"
	"sync"
	"time"

	st "github.com/shengyanli1982/workqueue/pkg/structs"
)

// Queue 方法接口
// Queue interface
type PriorityInterface interface {
	Interface
	// AddWeight 添加一个元素，指定权重，并在一段时间内排序
	// Add an element, add it use weight and sort it in a period of time
	AddWeight(element any, weight int) error
}

// Queue 的回调接口
// Callback interface
type PriorityCallback interface {
	Callback
	// OnWeight 添加元素后的回调
	// Callback after adding element
	OnWeight(element any, weight int)
}

// Queue 的配置
// Queue config
type PriorityQConfig struct {
	QConfig
	cb  PriorityCallback
	win int64
}

// 创建一个 Queue 的配置
// Create a new Queue config
func NewPriorityQConfig() *PriorityQConfig {
	return &PriorityQConfig{}
}

// 设置 Queue 的回调接口
// Set Queue callback
func (c *PriorityQConfig) WithCallback(cb PriorityCallback) *PriorityQConfig {
	c.cb = cb
	return c
}

// 设置 Queue 的排序窗口大小
// Set Queue sort window size
func (c *PriorityQConfig) WithWindow(win int64) *PriorityQConfig {
	c.win = win
	return c
}

type PriorityQ struct {
	*Q
	waiting *st.Heap
	stopCtx context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	once    sync.Once
	lock    *sync.Mutex
	config  *PriorityQConfig
}

// 创建一个 PriorityQueue 实例
// Create a new PriorityQueue config
func NewPriorityQueue(conf *PriorityQConfig) *PriorityQ {
	q := &PriorityQ{
		waiting: st.NewHeap(),
		wg:      sync.WaitGroup{},
		lock:    &sync.Mutex{},
		once:    sync.Once{},
		config:  conf,
	}

	q.isConfigValid()

	q.config.QConfig.cb = q.config.cb
	q.Q = NewQueue(&q.config.QConfig)
	q.lock = q.Q.lock
	q.stopCtx, q.cancel = context.WithCancel(context.Background())

	q.wg.Add(1)
	go q.loop()

	return q
}

// 判断 config 是否为空，如果为空，设置默认值
// Check if config is nil, if it is, set default value
func (q *PriorityQ) isConfigValid() {
	if q.config == nil {
		q.config = NewPriorityQConfig()
		q.config.WithCallback(emptyCallback{}).WithWindow(defaultQueueSortWin).WithCap(defaultQueueCap)
	}
	if q.config.cb == nil {
		q.config.cb = emptyCallback{}
	}
	if q.config.cap < defaultQueueCap && q.config.cap >= 0 {
		q.config.cap = defaultQueueCap
	}
	if q.config.cap < 0 {
		q.config.cap = math.MaxInt64 // 无限容量, unlimited capacity
	}
	if q.config.win <= defaultQueueSortWin {
		q.config.win = defaultQueueSortWin
	}
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

	q.lock.Lock()
	q.waiting.Push(st.NewElement(element, int64(weight)))
	q.lock.Unlock()

	q.config.cb.OnWeight(element, weight)

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
		case <-heartbeat.C: // 每隔一段时间，处理一次 Heap 中的元素。 Process the elements in the Heap every once in a while.
			var s0 []*st.Element // 用于存储 Heap 中的元素。 Used to store elements in Heap.

			q.lock.Lock()
			if q.waiting.Len() > 0 { // 如果 Heap 中有元素，则将元素复制到 s0 中。 If there are elements in the Heap, copy the elements to s0.
				s0 = make([]*st.Element, q.waiting.Len())
				copy(s0, q.waiting.Slice())
			}
			q.lock.Unlock()

			if s0 != nil {
				for i := 0; i < len(s0); i++ { // 将 s0 中的元素添加到 Queue 中。 Add the elements in s0 to the Queue.
					ele := s0[i]
					if err := q.Add(ele.Data()); err != nil { // 如果添加失败，则将元素重新添加到 Heap 中。 If the addition fails, the element is re-added to the Heap.
						q.lock.Lock()
						q.waiting.Push(ele) // 将元素重新添加到 Heap 中。 Re-add the element to the Heap.
						q.lock.Unlock()
					}
				}
				s0 = nil // 清空 s0。 Empty s0.
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
