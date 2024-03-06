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
	// 如果配置为空，则创建一个新的配置，并设置默认的限速器和回调
	// If the configuration is nil, create a new configuration and set the default rate limiter and callback
	if conf == nil {
		conf = NewRateLimitingQConfig()
		conf.WithLimiter(DefaultBucketRateLimiter()).WithCallback(newEmptyCallback())
	} else {
		// 如果回调为空，则设置一个空的回调
		// If the callback is nil, set an empty callback
		if conf.callback == nil {
			conf.callback = newEmptyCallback()
		}

		// 如果限速器为空，则设置默认的限速器
		// If the rate limiter is nil, set the default rate limiter
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
	// 如果队列为空，则返回 nil
	// If the queue is nil, return nil
	if queue == nil {
		return nil
	}

	// 验证配置是否有效
	// Verify that the configuration is valid
	conf = isRateLimitingQConfigValid(conf)
	conf.DelayingQConfig.callback = conf.callback

	// 创建一个新的 RateLimitingQueue 实例
	// Create a new RateLimitingQueue instance
	q := &RateLimitingQ{
		DelayingQInterface: queue,
		once:               sync.Once{},
		rlock:              &sync.Mutex{},
		config:             conf,
	}

	// 设置限速器
	// Set the rate limiter
	q.limiter = q.config.limiter

	// 返回 RateLimitingQueue 实例
	// Return the RateLimitingQueue instance
	return q
}

// 创建一个 RateLimitingQueue 实例
// Create a new RateLimitingQueue config
func NewRateLimitingQueue(conf *RateLimitingQConfig) *RateLimitingQ {
	// 验证配置是否有效
	// Verify that the configuration is valid
	conf = isRateLimitingQConfigValid(conf)
	conf.DelayingQConfig.callback = conf.callback

	// 创建一个新的 RateLimitingQueue 实例
	// Create a new RateLimitingQueue instance
	return newRateLimitingQueue(conf, NewDelayingQueue(&conf.DelayingQConfig))
}

// 创建一个 RateLimitingQueue 实例, 使用自定义 Queue (实现了 DelayingQ 接口)
// Create a new PriorityRateLimitingQueueQueue config, use custom Queue (implement DelayingQ interface)
func NewRateLimitingQueueWithCustomQueue(conf *RateLimitingQConfig, queue DelayingQInterface) *RateLimitingQ {
	// 验证配置是否有效
	// Verify that the configuration is valid
	conf = isRateLimitingQConfigValid(conf)
	conf.DelayingQConfig.callback = conf.callback
	conf.DelayingQConfig.QConfig.callback = conf.callback

	// 创建一个新的 RateLimitingQueue 实例
	// Create a new RateLimitingQueue instance
	return newRateLimitingQueue(conf, queue)
}

// 创建一个默认的 RateLimitingQueue 实例
// Create a new default RateLimitingQueue config
func DefaultRateLimitingQueue() RateLimitingQInterface {
	// 创建一个新的 RateLimitingQueue 实例，配置为 nil
	// Create a new RateLimitingQueue instance with nil configuration
	return NewRateLimitingQueue(nil)
}

// AddLimited 方法用于将元素添加到队列中，如果元素有 token，则直接添加到队列中，否则添加到等待队列中
// The AddLimited method is used to add an element to the queue. If the element has a token, it is directly added to the queue, otherwise it is added to the waiting queue
func (q *RateLimitingQ) AddLimited(element any) error {
	// 如果队列已经关闭，则返回 ErrorQueueClosed 错误
	// If the queue is already closed, return the ErrorQueueClosed error
	if q.IsClosed() {
		return ErrorQueueClosed
	}

	// 将元素添加到等待队列中，等待时间由 limiter 的 When 方法决定
	// Add the element to the waiting queue, the waiting time is determined by the When method of the limiter
	err := q.AddAfter(element, q.limiter.When(element))

	// 调用回调函数 OnAddLimited
	// Call the callback function OnAddLimited
	q.config.callback.OnAddLimited(element)

	// 返回错误
	// Return error
	return err
}

// Forget 方法用于忘记一个元素，即不对该元素进行限速处理
// The Forget method is used to forget an element, that is, not to rate limit the element
func (q *RateLimitingQ) Forget(element any) {
	// 调用 limiter 的 Forget 方法忘记一个元素
	// Call the Forget method of the limiter to forget an element
	q.limiter.Forget(element)

	// 调用回调函数 OnForget
	// Call the callback function OnForget
	q.config.callback.OnForget(element)
}

// NumLimitTimes 方法返回一个元素被限速的次数
// The NumLimitTimes method returns the number of times an element has been rate-limited
func (q *RateLimitingQ) NumLimitTimes(element any) int {
	// 调用 limiter 的 NumLimitTimes 方法获取一个元素被限速的次数
	// Call the NumLimitTimes method of the limiter to get the number of times an element has been rate-limited
	count := q.limiter.NumLimitTimes(element)

	// 调用回调函数 OnGetTimes
	// Call the callback function OnGetTimes
	q.config.callback.OnGetTimes(element, count)

	// 返回次数
	// Return count
	return count
}

// Stop 方法用于关闭队列
// The Stop method is used to close the queue
func (q *RateLimitingQ) Stop() {
	// 调用 DelayingQInterface 的 Stop 方法关闭队列
	// Call the Stop method of DelayingQInterface to close the queue
	q.DelayingQInterface.Stop()

	// 使用 sync.Once 确保 limiter 的 Stop 方法只被调用一次
	// Use sync.Once to ensure that the Stop method of limiter is only called once
	q.once.Do(func() {
		// 调用 limiter 的 Stop 方法停止限速器
		// Call the Stop method of limiter to stop the rate limiter
		q.limiter.Stop()
	})
}
