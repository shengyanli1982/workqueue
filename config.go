package workqueue

import "github.com/shengyanli1982/workqueue/v2/internal/container/set"

// NewSetFunc 定义了创建新 Set 的函数类型
// NewSetFunc defines the function type for creating a new Set
type NewSetFunc = func() Set

// defaultNewSetFunc 是一个默认的 Set 创建器函数
// defaultNewSetFunc is a default Set creator function
var defaultNewSetFunc = func() Set { return set.New() }

// QueueConfig 定义队列的基本配置
// QueueConfig defines the basic configuration for a queue
type QueueConfig struct {
	callback   QueueCallback // 队列回调函数 / Queue callback function
	idempotent bool          // 是否开启幂等性 / Whether to enable idempotency
	setCreator NewSetFunc    // Set 创建器函数 / Set creator function
}

// NewQueueConfig 创建一个新的队列配置实例
// NewQueueConfig creates a new queue configuration instance
func NewQueueConfig() *QueueConfig {
	return &QueueConfig{
		callback:   NewNopQueueCallbackImpl(),
		setCreator: defaultNewSetFunc,
	}
}

// WithCallback 设置队列回调函数
// WithCallback sets the queue callback function
func (c *QueueConfig) WithCallback(cb QueueCallback) *QueueConfig {
	c.callback = cb
	return c
}

// WithValueIdempotent 设置队列元素的幂等性
// WithValueIdempotent sets the idempotency of queue elements
func (c *QueueConfig) WithValueIdempotent() *QueueConfig {
	c.idempotent = true

	return c
}

// WithSetCreator 设置 Set 创建器函数
// WithSetCreator sets the Set creator function
func (c *QueueConfig) WithSetCreator(fn NewSetFunc) *QueueConfig {
	c.setCreator = fn

	return c
}

// isQueueConfigEffective 确保队列配置的有效性
// isQueueConfigEffective ensures the effectiveness of queue configuration
func isQueueConfigEffective(c *QueueConfig) *QueueConfig {
	if c != nil {
		if c.callback == nil {
			c.callback = NewNopQueueCallbackImpl()
		}

		if c.setCreator == nil {
			c.setCreator = defaultNewSetFunc
		}
	} else {
		c = NewQueueConfig()
	}

	return c
}

// DelayingQueueConfig 定义延迟队列的配置，继承自 QueueConfig
// DelayingQueueConfig defines the configuration for a delaying queue, inherits from QueueConfig
type DelayingQueueConfig struct {
	QueueConfig
	callback DelayingQueueCallback // 延迟队列回调函数 / Delaying queue callback function
}

// NewDelayingQueueConfig 创建一个新的延迟队列配置实例
// NewDelayingQueueConfig creates a new delaying queue configuration instance
func NewDelayingQueueConfig() *DelayingQueueConfig {
	return &DelayingQueueConfig{
		QueueConfig: *NewQueueConfig(),

		callback: NewNopDelayingQueueCallbackImpl(),
	}
}

// WithCallback 设置延迟队列回调函数
// WithCallback sets the delaying queue callback function
func (c *DelayingQueueConfig) WithCallback(cb DelayingQueueCallback) *DelayingQueueConfig {
	c.callback = cb
	c.QueueConfig.callback = cb

	return c
}

// isDelayingQueueConfigEffective 确保延迟队列配置的有效性
// isDelayingQueueConfigEffective ensures the effectiveness of delaying queue configuration
func isDelayingQueueConfigEffective(c *DelayingQueueConfig) *DelayingQueueConfig {
	if c != nil {
		if c.callback == nil {
			c.callback = NewNopDelayingQueueCallbackImpl()
		}

		if c.QueueConfig.callback == nil {
			c.QueueConfig.callback = NewNopQueueCallbackImpl()
		}
	} else {
		c = NewDelayingQueueConfig()
	}

	return c
}

// PriorityQueueConfig 定义优先级队列的配置，继承自 QueueConfig
// PriorityQueueConfig defines the configuration for a priority queue, inherits from QueueConfig
type PriorityQueueConfig struct {
	QueueConfig
	callback PriorityQueueCallback // 优先级队列回调函数 / Priority queue callback function
}

// NewPriorityQueueConfig 创建一个新的优先级队列配置实例
// NewPriorityQueueConfig creates a new priority queue configuration instance
func NewPriorityQueueConfig() *PriorityQueueConfig {
	return &PriorityQueueConfig{
		QueueConfig: *NewQueueConfig(),

		callback: NewNopPriorityQueueCallbackImpl(),
	}
}

// WithCallback 设置优先级队列回调函数
// WithCallback sets the priority queue callback function
func (c *PriorityQueueConfig) WithCallback(cb PriorityQueueCallback) *PriorityQueueConfig {
	c.callback = cb
	c.QueueConfig.callback = cb

	return c
}

// isPriorityQueueConfigEffective 确保优先级队列配置的有效性
// isPriorityQueueConfigEffective ensures the effectiveness of priority queue configuration
func isPriorityQueueConfigEffective(c *PriorityQueueConfig) *PriorityQueueConfig {
	if c != nil {
		if c.callback == nil {
			c.callback = NewNopPriorityQueueCallbackImpl()
		}

		if c.QueueConfig.callback == nil {
			c.QueueConfig.callback = NewNopQueueCallbackImpl()
		}
	} else {
		c = NewPriorityQueueConfig()
	}

	return c
}

// RateLimitingQueueConfig 定义限速队列的配置，继承自 DelayingQueueConfig
// RateLimitingQueueConfig defines the configuration for a rate limiting queue, inherits from DelayingQueueConfig
type RateLimitingQueueConfig struct {
	DelayingQueueConfig
	callback RateLimitingQueueCallback // 限速队列回调函数 / Rate limiting queue callback function
	limiter  Limiter                   // 限速器 / Rate limiter
}

// NewRateLimitingQueueConfig 创建一个新的限速队列配置实例
// NewRateLimitingQueueConfig creates a new rate limiting queue configuration instance
func NewRateLimitingQueueConfig() *RateLimitingQueueConfig {
	return &RateLimitingQueueConfig{
		DelayingQueueConfig: *NewDelayingQueueConfig(),

		callback: NewNopRateLimitingQueueCallbackImpl(),

		limiter: NewNopRateLimiterImpl(),
	}
}

// WithCallback 设置限速队列回调函数
// WithCallback sets the rate limiting queue callback function
func (c *RateLimitingQueueConfig) WithCallback(cb RateLimitingQueueCallback) *RateLimitingQueueConfig {
	c.callback = cb
	c.DelayingQueueConfig.callback = cb
	c.DelayingQueueConfig.QueueConfig.callback = cb

	return c
}

// WithLimiter 设置限速器
// WithLimiter sets the rate limiter
func (c *RateLimitingQueueConfig) WithLimiter(limiter Limiter) *RateLimitingQueueConfig {
	c.limiter = limiter

	return c
}

// isRateLimitingQueueConfigEffective 确保限速队列配置的有效性
// isRateLimitingQueueConfigEffective ensures the effectiveness of rate limiting queue configuration
func isRateLimitingQueueConfigEffective(c *RateLimitingQueueConfig) *RateLimitingQueueConfig {
	if c != nil {
		// 设置默认回调函数 / Set default callbacks if not provided
		if c.callback == nil {
			c.callback = NewNopRateLimitingQueueCallbackImpl()
		}
		if c.DelayingQueueConfig.callback == nil {
			c.DelayingQueueConfig.callback = NewNopDelayingQueueCallbackImpl()
		}
		if c.DelayingQueueConfig.QueueConfig.callback == nil {
			c.DelayingQueueConfig.QueueConfig.callback = NewNopQueueCallbackImpl()
		}
		// 设置默认限速器 / Set default rate limiter if not provided
		if c.limiter == nil {
			c.limiter = NewNopRateLimiterImpl()
		}
	} else {
		c = NewRateLimitingQueueConfig()
	}
	return c
}
