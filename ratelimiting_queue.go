package workqueue

import (
	"sync"
)

// RateLimitingQInterface 是 Queue 方法的接口
// RateLimitingQInterface is the interface for Queue methods
type RateLimitingQInterface interface {
	// 继承 DelayingQueue 接口
	// Inherit DelayingQueue
	DelayingQInterface

	// AddLimited 添加一个元素，需要对该元素进行限速处理
	// AddLimited adds an element that needs to be rate-limited
	AddLimited(element any) error

	// Forget 忘记一个元素，不需要对该元素进行限速处理
	// Forget forgets an element that doesn't need to be rate-limited
	Forget(element any)

	// NumLimitTimes 返回一个元素被限速的次数
	// NumLimitTimes returns the number of times an element has been rate-limited
	NumLimitTimes(element any) int
}

// RateLimitingQCallback 是 Queue 的回调接口
// RateLimitingQCallback is the callback interface for Queue
type RateLimitingQCallback interface {
	// 继承 DelayingCallback 接口
	// Inherit DelayingCallback
	DelayingQCallback

	// OnAddLimited 添加元素后的回调
	// OnAddLimited is the callback after adding an element
	OnAddLimited(any)

	// OnForget 忘记元素后的回调
	// OnForget is the callback after forgetting an element
	OnForget(any)

	// OnGetTimes 返回一个元素被限速的次数
	// OnGetTimes returns the number of times an element has been rate-limited
	OnGetTimes(any, int)
}

// RateLimitingQConfig 是 Queue 的配置
// RateLimitingQConfig is the configuration for Queue
type RateLimitingQConfig struct {
	DelayingQConfig
	callback RateLimitingQCallback
	limiter  RateLimiter
}

// NewRateLimitingQConfig 创建一个 Queue 的配置
// NewRateLimitingQConfig creates a new Queue configuration
func NewRateLimitingQConfig() *RateLimitingQConfig {
	return &RateLimitingQConfig{}
}

// WithCallback 设置 Queue 的回调接口
// WithCallback sets the callback interface for Queue
func (c *RateLimitingQConfig) WithCallback(cb RateLimitingQCallback) *RateLimitingQConfig {
	c.callback = cb
	return c
}

// WithLimiter 设置 Limiter 的实例
// WithLimiter sets the Limiter instance
func (c *RateLimitingQConfig) WithLimiter(limiter RateLimiter) *RateLimitingQConfig {
	c.limiter = limiter
	return c
}

// 验证队列的配置是否有效
// Verify that the queue configuration is valid
func isRateLimitingQConfigValid(conf *RateLimitingQConfig) *RateLimitingQConfig {
	if conf == nil {
		conf = NewRateLimitingQConfig()
		conf.WithLimiter(DefaultBucketRateLimiter()).WithCallback(emptyCallback{})
	} else {
		if conf.callback == nil {
			conf.callback = emptyCallback{}
		}
		if conf.limiter == nil {
			conf.limiter = DefaultBucketRateLimiter()
		}
	}

	return conf
}

// 限速队列数据结构
// RateLimitingQueue data structure
type RateLimitingQ struct {
	DelayingQInterface
	once    sync.Once
	rlock   *sync.Mutex
	limiter RateLimiter
	config  *RateLimitingQConfig
}

// 创建 RateLimitingQueue 实例
// Create a new RateLimitingQueue config
func newRateLimitingQueue(conf *RateLimitingQConfig, queue DelayingQInterface) *RateLimitingQ {
	if queue == nil {
		return nil
	}

	conf = isRateLimitingQConfigValid(conf)
	conf.DelayingQConfig.callback = conf.callback

	q := &RateLimitingQ{
		DelayingQInterface: queue,
		once:               sync.Once{},
		rlock:              &sync.Mutex{},
		config:             conf,
	}

	q.limiter = q.config.limiter

	return q
}

// 创建一个 RateLimitingQueue 实例
// Create a new RateLimitingQueue config
func NewRateLimitingQueue(conf *RateLimitingQConfig) *RateLimitingQ {
	conf = isRateLimitingQConfigValid(conf)
	conf.DelayingQConfig.callback = conf.callback
	return newRateLimitingQueue(conf, NewDelayingQueue(&conf.DelayingQConfig))
}

// 创建一个 RateLimitingQueue 实例, 使用自定义 Queue (实现了 DelayingQ 接口)
// Create a new PriorityRateLimitingQueueQueue config, use custom Queue (implement DelayingQ interface)
func NewRateLimitingQueueWithCustomQueue(conf *RateLimitingQConfig, queue DelayingQInterface) *RateLimitingQ {
	conf = isRateLimitingQConfigValid(conf)
	conf.DelayingQConfig.callback = conf.callback
	conf.DelayingQConfig.QConfig.callback = conf.callback
	return newRateLimitingQueue(conf, queue)
}

// 创建一个默认的 RateLimitingQueue 实例
// Create a new default RateLimitingQueue config
func DefaultRateLimitingQueue() RateLimitingQInterface {
	return NewRateLimitingQueue(nil)
}

// 添加元素到队列, 加入到等待队列, 如果有 token 则直接加入到队列
// Add an element to the queue, add it to the waiting queue, and add it to the queue directly if there is a token
func (q *RateLimitingQ) AddLimited(element any) error {
	// 如果队列已经关闭，返回 ErrorQueueClosed 错误
	// If the queue is already closed, return ErrorQueueClosed
	if q.IsClosed() {
		return ErrorQueueClosed
	}

	// 加入到等待队列
	// add it to the waiting queue
	err := q.AddAfter(element, q.limiter.When(element))

	// 回调
	// Callback
	q.config.callback.OnAddLimited(element)

	return err
}

// 忘记一个元素，不需要这个元素处理经行限速
// Forget an element, don't need to limit the speed of this element
func (q *RateLimitingQ) Forget(element any) {
	// 忘记一个元素
	// Forget an element
	q.limiter.Forget(element)

	// 回调
	// Callback
	q.config.callback.OnForget(element)
}

// NumLimitTimes 返回一个元素被限速的次数
// Return the number of times an element is limited
func (q *RateLimitingQ) NumLimitTimes(element any) int {
	// 元素被限速的次数
	// The number of times an element is limited
	count := q.limiter.NumLimitTimes(element)

	// 回调
	// Callback
	q.config.callback.OnGetTimes(element, count)

	return count
}

// 关闭队列
// Close queue
func (q *RateLimitingQ) Stop() {
	q.DelayingQInterface.Stop()
	q.once.Do(func() {
		q.limiter.Stop()
	})
}
