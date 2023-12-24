package workqueue

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shengyanli1982/workqueue/internal/stl/heap"
)

// DelayingInterface 是 Queue 方法接口的延迟版本
// DelayingInterface is the delayed version of the Queue method interface
type DelayingInterface interface {
	Interface
	// AddAfter 添加一个元素，延迟一段时间后再执行
	// Add an element, execute it after a delay
	AddAfter(element any, delay time.Duration) error
}

// DelayingCallback 是 Queue 的回调接口的延迟版本
// DelayingCallback is the delayed version of the Queue callback interface
type DelayingCallback interface {
	Callback
	// OnAddAfter 添加元素后的回调
	// Callback after adding element
	OnAddAfter(any, time.Duration)
}

// DelayingQConfig 是 Queue 的配置的延迟版本
// DelayingQConfig is the delayed version of the Queue config
type DelayingQConfig struct {
	QConfig
	cb DelayingCallback
}

// NewDelayingQConfig 创建一个 DelayingQConfig 实例
// Create a new DelayingQConfig instance
func NewDelayingQConfig() *DelayingQConfig {
	return &DelayingQConfig{}
}

// WithCallback 设置 Queue 的回调接口
// Set Queue callback
func (c *DelayingQConfig) WithCallback(cb DelayingCallback) *DelayingQConfig {
	c.cb = cb
	return c
}

// DelayingQ 是 DelayingQueue 的实现
// DelayingQ is the implementation of DelayingQueue
type DelayingQ struct {
	*Q
	waiting *heap.Heap
	elepool *heap.HeapElementPool
	stopCtx context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	once    sync.Once
	lock    *sync.Mutex
	now     atomic.Int64
	config  *DelayingQConfig
}

// NewDelayingQueue 创建一个 DelayingQueue 实例
// Create a new DelayingQueue instance
func NewDelayingQueue(conf *DelayingQConfig) *DelayingQ {
	q := &DelayingQ{
		waiting: heap.NewHeap(),
		elepool: heap.NewHeapElementPool(),
		wg:      sync.WaitGroup{},
		now:     atomic.Int64{},
		once:    sync.Once{},
		config:  conf,
	}

	q.isConfigValid()

	q.config.QConfig.cb = q.config.cb
	q.Q = NewQueue(&q.config.QConfig)
	q.lock = q.Q.lock
	q.stopCtx, q.cancel = context.WithCancel(context.Background())

	q.wg.Add(2)
	go q.loop()
	go q.syncNow()

	return q
}

// 创建一个默认的 DelayingQueue 对象
// Create a new default DelayingQueue object
func DefaultDelayingQueue() DelayingInterface {
	return NewDelayingQueue(nil)
}

// isConfigValid 检查配置是否有效，如果为空则设置默认值
// Check if the config is valid, if it is nil, set default values
func (q *DelayingQ) isConfigValid() {
	if q.config == nil {
		q.config = &DelayingQConfig{}
		q.config.WithCallback(emptyCallback{})
	} else {
		if q.config.cb == nil {
			q.config.cb = emptyCallback{}
		}
	}
}

// AddAfter 将元素添加到队列中，在延迟一段时间后处理
// Add an element to the queue and process it after a specified delay
func (q *DelayingQ) AddAfter(element any, delay time.Duration) error {
	if q.IsClosed() {
		return ErrorQueueClosed
	}

	if delay <= 0 {
		return q.Add(element)
	}

	ele := q.elepool.Get()
	ele.SetData(element)
	ele.SetValue(time.Now().Add(delay).UnixMilli())

	q.lock.Lock()
	q.waiting.Push(ele)
	q.lock.Unlock()

	q.config.cb.OnAddAfter(element, delay)

	return nil
}

// 同步当前的时间
// Sync current time
func (q *DelayingQ) syncNow() {
	heartbeat := time.NewTicker(time.Millisecond * 500)

	defer func() {
		q.wg.Done()
		heartbeat.Stop()
	}()

	for {
		select {
		case <-q.stopCtx.Done():
			return
		case <-heartbeat.C:
			q.now.Store(time.Now().UnixMilli())
		}
	}
}

// 循环处理 Heap 中的元素
// Loop to process elements in Heap
func (q *DelayingQ) loop() {
	heartbeat := time.NewTicker(time.Millisecond * 500)

	defer func() {
		q.wg.Done()
		heartbeat.Stop()
	}()

	for {
		select {
		case <-q.stopCtx.Done():
			return
		default:
			q.lock.Lock()
			// 如果堆中有元素 If there are elements in the heap
			if q.waiting.Len() > 0 {
				// 获取堆顶元素
				ele := q.waiting.Head()
				// 如果堆顶元素的时间小于当前时间, 意味对象已经超时
				if ele.Value() <= q.now.Load() {
					// 弹出堆顶元素
					_ = q.waiting.Pop()
					q.lock.Unlock()
					// 添加到队列中
					if err := q.Add(ele.Data()); err != nil {
						q.lock.Lock()
						// 重置元素的值 Reset the value of the element
						ele.SetValue(q.now.Load() + 1500)
						// 将元素重新添加到堆中 Re-add the element to the heap
						q.waiting.Push(ele)
						q.lock.Unlock()
					} else {
						// 释放元素 Free element
						q.elepool.Put(ele)
					}
				} else {
					q.lock.Unlock()
				}
			} else {
				q.lock.Unlock()
				// 500ms 后再次检查堆中的元素 Check the elements in the heap again after 500ms
				<-heartbeat.C
			}
		}
	}
}

// 关闭队列
// Close queue
func (q *DelayingQ) Stop() {
	q.Q.Stop()
	q.once.Do(func() {
		q.cancel()
		q.wg.Wait()
		q.waiting.Reset()
	})
}
