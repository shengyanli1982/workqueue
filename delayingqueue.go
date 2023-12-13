package workqueue

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"

	st "github.com/shengyanli1982/workqueue/pkg/structs"
)

// Queue 方法接口
// Queue interface
type DelayingInterface interface {
	Interface
	// AddAfter 添加一个元素，延迟一段时间后再执行
	// Add an element, add it after a delay
	AddAfter(element any, delay time.Duration) error
}

// Queue 的回调接口
// Callback interface
type DelayingCallback interface {
	Callback
	// OnAfter 添加元素后的回调
	// Callback after adding element
	OnAfter(any, time.Duration)
}

// Queue 的配置
// Queue config
type DelayingQConfig struct {
	QConfig
	cb DelayingCallback
}

// 创建一个 Queue 的配置
// Create a new Queue config
func NewDelayingQConfig() *DelayingQConfig {
	return &DelayingQConfig{}
}

// 设置 Queue 的回调接口
// Set Queue callback
func (c *DelayingQConfig) WithCallback(cb DelayingCallback) *DelayingQConfig {
	c.cb = cb
	return c
}

type DelayingQ struct {
	*Q
	waiting *st.Heap
	stopCtx context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	once    sync.Once
	lock    *sync.Mutex
	now     atomic.Int64
	config  *DelayingQConfig
}

// 创建一个 DelayingQueue 实例
// Create a new DelayingQueue config
func NewDelayingQueue(conf *DelayingQConfig) *DelayingQ {
	q := &DelayingQ{
		waiting: st.NewHeap(),
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

// 判断 config 是否为空，如果为空，设置默认值
// Check if config is nil, if it is, set default value
func (q *DelayingQ) isConfigValid() {
	if q.config == nil {
		q.config = &DelayingQConfig{}
		q.config.WithCallback(emptyCallback{}).WithCap(defaultQueueCap)
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
}

// 添加元素到队列, 延迟一段时间后再处理
// Add an element to the queue, process it after a specified delay
func (q *DelayingQ) AddAfter(element any, delay time.Duration) error {
	if q.IsClosed() {
		return ErrorQueueClosed
	}

	if delay <= 0 {
		return q.Add(element)
	}

	q.lock.Lock()
	q.waiting.Push(st.NewElement(element, time.Now().Add(delay).UnixMilli()))
	q.lock.Unlock()

	q.config.cb.OnAfter(element, delay)

	return nil
}

// 同步当前的时间
// Sync current time
func (q *DelayingQ) syncNow() {
	heartbeat := time.NewTicker(time.Millisecond * 500)

	defer func() {
		q.wg.Done()
		heartbeat.Stop()
		// fmt.Println("DelayingQ syncNow stop")
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
		// fmt.Println("DelayingQ loop stop")
	}()

	var ele *st.Element

	for {
		select {
		case <-q.stopCtx.Done():
			return
		default:
			q.lock.Lock()
			if q.waiting.Len() > 0 {
				ele = q.waiting.Head()           // 获取堆顶元素
				if ele.Value() <= q.now.Load() { // 如果堆顶元素的时间小于当前时间, 意味对象已经超时
					_ = q.waiting.Pop() // 弹出堆顶元素
				} else {
					ele = nil
				}
				q.lock.Unlock()
			} else {
				q.lock.Unlock()
				<-heartbeat.C // 500ms 后再次尝试
				break         // 跳出 select
			}

			if ele != nil {
				if err := q.Add(ele.Data()); err != nil {
					q.lock.Lock()
					ele.ResetValue(q.now.Load() + 1500) // 重置元素的值
					q.waiting.Push(ele)
					q.lock.Unlock()
				}
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
