package workqueue

import (
	"sync"
)

// RateLimitingInterface 是 Queue 方法的接口
// RateLimitingInterface is the interface for Queue methods
type RateLimitingInterface interface {
	Interface
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

// RateLimitingCallback 是 Queue 的回调接口
// RateLimitingCallback is the callback interface for Queue
type RateLimitingCallback interface {
	DelayingCallback
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
	cb      RateLimitingCallback
	limiter RateLimiter
}

// NewRateLimitingQConfig 创建一个 Queue 的配置
// NewRateLimitingQConfig creates a new Queue configuration
func NewRateLimitingQConfig() *RateLimitingQConfig {
	return &RateLimitingQConfig{}
}

// WithCallback 设置 Queue 的回调接口
// WithCallback sets the callback interface for Queue
func (c *RateLimitingQConfig) WithCallback(cb RateLimitingCallback) *RateLimitingQConfig {
	c.cb = cb
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
		conf = &RateLimitingQConfig{}
		conf.WithLimiter(DefaultBucketRateLimiter()).WithCallback(emptyCallback{})
	} else {
		if conf.cb == nil {
			conf.cb = emptyCallback{}
		}
		if conf.limiter == nil {
			conf.limiter = DefaultBucketRateLimiter()
		}
	}

	return conf
}

type RateLimitingQ struct {
	*DelayingQ
	once    sync.Once
	lock    *sync.Mutex
	limiter RateLimiter
	config  *RateLimitingQConfig
}

// 创建一个 RateLimitingQueue 实例
// Create a new RateLimitingQueue config
func NewRateLimitingQueue(conf *RateLimitingQConfig) *RateLimitingQ {
	conf = isRateLimitingQConfigValid(conf)
	conf.DelayingQConfig.cb = conf.cb

	q := &RateLimitingQ{
		DelayingQ: NewDelayingQueue(&conf.DelayingQConfig),
		once:      sync.Once{},
		config:    conf,
	}

	q.lock = q.DelayingQ.lock
	q.limiter = q.config.limiter

	return q
}

// 创建一个默认的 RateLimitingQueue 实例
// Create a new default RateLimitingQueue config
func DefaultRateLimitingQueue() RateLimitingInterface {
	return NewRateLimitingQueue(nil)
}

// 添加元素到队列, 加入到等待队列, 如果有 token 则直接加入到队列
// Add an element to the queue, add it to the waiting queue, and add it to the queue directly if there is a token
func (q *RateLimitingQ) AddLimited(element any) error {
	if q.IsClosed() {
		return ErrorQueueClosed
	}

	err := q.AddAfter(element, q.limiter.When(element)) // 加入到等待队列, add it to the waiting queue
	q.config.cb.OnAddLimited(element)
	return err
}

// 忘记一个元素，不需要这个元素处理经行限速
// Forget an element, don't need to limit the speed of this element
func (q *RateLimitingQ) Forget(element any) {
	q.limiter.Forget(element)
	q.config.cb.OnForget(element)
}

// NumLimitTimes 返回一个元素被限速的次数
// Return the number of times an element is limited
func (q *RateLimitingQ) NumLimitTimes(element any) int {
	count := q.limiter.NumLimitTimes(element)
	q.config.cb.OnGetTimes(element, count)
	return count
}

// 关闭队列
// Close queue
func (q *RateLimitingQ) Stop() {
	q.DelayingQ.Stop()
	q.once.Do(func() {
		q.limiter.Stop()
	})
}
