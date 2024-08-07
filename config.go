package workqueue

import "github.com/shengyanli1982/workqueue/v2/internal/container/set"

// NewSetFunc 是一个函数类型，该函数返回一个 Set 实例
// NewSetFunc is a function type that returns an instance of Set
type NewSetFunc = func() Set

// defaultNewSetFunc 是一个默认的集合创建函数，它返回一个新的 Set 实例
// defaultNewSetFunc is a default set creation function that returns a new Set instance
var defaultNewSetFunc = func() Set { return set.New() }

// QueueConfig 是一个结构体，用于配置队列
// QueueConfig is a struct used for configuring the queue
type QueueConfig struct {
	// callback 是队列的回调函数
	// callback is the callback function of the queue
	callback QueueCallback

	// idempotent 是一个布尔值，表示队列中的元素是否是幂等的
	// idempotent is a boolean value that indicates whether the elements in the queue are idempotent
	idempotent bool

	// setCreator 是一个集合创建函数，用于创建新的 Set 实例
	// setCreator is a set creation function used to create new Set instances
	setCreator NewSetFunc
}

// NewQueueConfig 函数用于创建一个新的 QueueConfig 实例
// The NewQueueConfig function is used to create a new instance of QueueConfig
func NewQueueConfig() *QueueConfig {
	// 返回一个新的 QueueConfig 实例
	// Return a new instance of QueueConfig
	return &QueueConfig{
		// 使用 NewNopQueueCallbackImpl 函数创建一个新的无操作的队列回调函数实例
		// Use the NewNopQueueCallbackImpl function to create a new instance of a no-operation queue callback function
		callback: NewNopQueueCallbackImpl(),

		// 使用默认的集合创建函数
		// Use the default set creation function
		setCreator: defaultNewSetFunc,
	}
}

// WithCallback 方法用于设置队列的回调函数
// The WithCallback method is used to set the callback function of the queue
func (c *QueueConfig) WithCallback(cb QueueCallback) *QueueConfig {
	// 设置回调函数
	// Set the callback function
	c.callback = cb

	// 返回配置实例，以便进行链式调用
	// Return the configuration instance for chaining calls
	return c
}

// WithValueIdempotent 方法用于设置队列中的元素为幂等的
// The WithValueIdempotent method is used to set the elements in the queue to be idempotent
func (c *QueueConfig) WithValueIdempotent() *QueueConfig {
	// 设置元素为幂等的
	// Set the elements to be idempotent
	c.idempotent = true

	// 返回配置实例，以便进行链式调用
	// Return the configuration instance for chaining calls
	return c
}

// WithSetCreator 方法用于设置集合创建函数
// The WithSetCreator method is used to set the set creation function
func (c *QueueConfig) WithSetCreator(fn NewSetFunc) *QueueConfig {
	// 设置集合创建函数
	// Set the set creation function
	c.setCreator = fn

	// 返回配置实例，以便进行链式调用
	// Return the configuration instance for chaining calls
	return c
}

// isQueueConfigEffective 函数用于检查队列配置是否有效，如果无效，则使用默认配置
// The isQueueConfigEffective function is used to check whether the queue configuration is effective. If it is not, the default configuration is used
func isQueueConfigEffective(c *QueueConfig) *QueueConfig {
	// 如果配置不为 nil
	// If the configuration is not nil
	if c != nil {
		// 如果回调函数为 nil，则设置为无操作的队列回调函数
		// If the callback function is nil, set it to a no-operation queue callback function
		if c.callback == nil {
			c.callback = NewNopQueueCallbackImpl()
		}

		// 如果集合创建函数为 nil，则设置为默认的集合创建函数
		// If the set creation function is nil, set it to the default set creation function
		if c.setCreator == nil {
			c.setCreator = defaultNewSetFunc
		}
	} else {
		// 如果配置为 nil，则创建一个新的队列配置
		// If the configuration is nil, create a new queue configuration
		c = NewQueueConfig()
	}

	// 返回配置
	// Return the configuration
	return c
}

// DelayingQueueConfig 结构体，用于配置延迟队列
// The DelayingQueueConfig struct, used for configuring the delaying queue
type DelayingQueueConfig struct {
	// QueueConfig 是队列的配置
	// QueueConfig is the configuration of the queue
	QueueConfig

	// callback 是延迟队列的回调函数
	// callback is the callback function of the delaying queue
	callback DelayingQueueCallback
}

// NewDelayingQueueConfig 函数用于创建一个新的 DelayingQueueConfig
// The NewDelayingQueueConfig function is used to create a new DelayingQueueConfig
func NewDelayingQueueConfig() *DelayingQueueConfig {
	// 返回一个新的 DelayingQueueConfig 实例
	// Return a new instance of DelayingQueueConfig
	return &DelayingQueueConfig{
		// 使用 NewQueueConfig 函数创建一个新的 QueueConfig 实例
		// Use the NewQueueConfig function to create a new instance of QueueConfig
		QueueConfig: *NewQueueConfig(),

		// 使用 NewNopDelayingQueueCallbackImpl 函数创建一个新的无操作的延迟队列回调函数实例
		// Use the NewNopDelayingQueueCallbackImpl function to create a new instance of a no-operation delaying queue callback function
		callback: NewNopDelayingQueueCallbackImpl(),
	}
}

// WithCallback 方法用于设置延迟队列的回调函数
// The WithCallback method is used to set the callback function of the delaying queue
func (c *DelayingQueueConfig) WithCallback(cb DelayingQueueCallback) *DelayingQueueConfig {
	// 设置回调函数
	// Set the callback function
	c.callback = cb
	c.QueueConfig.callback = cb

	// 返回配置
	// Return the configuration
	return c
}

// isDelayingQueueConfigEffective 函数用于检查延迟队列配置是否有效，如果无效，则使用默认配置
// The isDelayingQueueConfigEffective function is used to check whether the delaying queue configuration is effective. If not, use the default configuration
func isDelayingQueueConfigEffective(c *DelayingQueueConfig) *DelayingQueueConfig {
	// 如果配置不为 nil
	// If the configuration is not nil
	if c != nil {
		// 如果回调函数为 nil，则设置为无操作的延迟队列回调函数
		// If the callback function is nil, set it to a no-operation delaying queue callback function
		if c.callback == nil {
			c.callback = NewNopDelayingQueueCallbackImpl()
		}

		// 如果队列配置的回调函数为 nil，则设置为无操作的队列回调函数
		// If the callback function of the queue configuration is nil, set it to a no-operation queue callback function
		if c.QueueConfig.callback == nil {
			c.QueueConfig.callback = NewNopQueueCallbackImpl()
		}
	} else {
		// 如果配置为 nil，则创建一个新的延迟队列配置
		// If the configuration is nil, create a new delaying queue configuration
		c = NewDelayingQueueConfig()
	}

	// 返回配置
	// Return the configuration
	return c
}

// PriorityQueueConfig 结构体，用于配置优先队列
// The PriorityQueueConfig struct, used for configuring the priority queue
type PriorityQueueConfig struct {
	// QueueConfig 是队列的配置
	// QueueConfig is the configuration of the queue
	QueueConfig

	// callback 是优先队列的回调函数
	// callback is the callback function of the priority queue
	callback PriorityQueueCallback
}

// NewPriorityQueueConfig 函数用于创建一个新的 PriorityQueueConfig
// The NewPriorityQueueConfig function is used to create a new PriorityQueueConfig
func NewPriorityQueueConfig() *PriorityQueueConfig {
	// 返回一个新的 PriorityQueueConfig 实例
	// Return a new instance of PriorityQueueConfig
	return &PriorityQueueConfig{
		// 使用 NewQueueConfig 函数创建一个新的 QueueConfig 实例
		// Use the NewQueueConfig function to create a new instance of QueueConfig
		QueueConfig: *NewQueueConfig(),

		// 使用 NewNopPriorityQueueCallbackImpl 函数创建一个新的无操作的优先队列回调函数实例
		// Use the NewNopPriorityQueueCallbackImpl function to create a new instance of a no-operation priority queue callback function
		callback: NewNopPriorityQueueCallbackImpl(),
	}
}

// WithCallback 方法用于设置优先队列的回调函数
// The WithCallback method is used to set the callback function of the priority queue
func (c *PriorityQueueConfig) WithCallback(cb PriorityQueueCallback) *PriorityQueueConfig {
	// 设置回调函数
	// Set the callback function
	c.callback = cb
	c.QueueConfig.callback = cb

	// 返回配置
	// Return the configuration
	return c
}

// isPriorityQueueConfigEffective 函数用于检查优先队列配置是否有效，如果无效，则使用默认配置
// The isPriorityQueueConfigEffective function is used to check whether the priority queue configuration is effective. If not, use the default configuration
func isPriorityQueueConfigEffective(c *PriorityQueueConfig) *PriorityQueueConfig {
	// 如果配置不为 nil
	// If the configuration is not nil
	if c != nil {
		// 如果回调函数为 nil，则设置为无操作的优先队列回调函数
		// If the callback function is nil, set it to a no-operation priority queue callback function
		if c.callback == nil {
			c.callback = NewNopPriorityQueueCallbackImpl()
		}

		// 如果队列配置的回调函数为 nil，则设置为无操作的队列回调函数
		// If the callback function of the queue configuration is nil, set it to a no-operation queue callback function
		if c.QueueConfig.callback == nil {
			c.QueueConfig.callback = NewNopQueueCallbackImpl()
		}
	} else {
		// 如果配置为 nil，则创建一个新的优先队列配置
		// If the configuration is nil, create a new priority queue configuration
		c = NewPriorityQueueConfig()
	}

	// 返回配置
	// Return the configuration
	return c
}

// RateLimitingQueueConfig 结构体，用于配置限流队列
// The RateLimitingQueueConfig struct, used for configuring the rate limiting queue
type RateLimitingQueueConfig struct {
	// DelayingQueueConfig 是延迟队列的配置
	// DelayingQueueConfig is the configuration of the delaying queue
	DelayingQueueConfig

	// callback 是限流队列的回调函数
	// callback is the callback function of the rate limiting queue
	callback RateLimitingQueueCallback

	// limiter 是限流器
	// limiter is the rate limiter
	limiter Limiter
}

// NewRateLimitingQueueConfig 函数用于创建一个新的 RateLimitingQueueConfig
// The NewRateLimitingQueueConfig function is used to create a new RateLimitingQueueConfig
func NewRateLimitingQueueConfig() *RateLimitingQueueConfig {
	// 返回一个新的 RateLimitingQueueConfig 实例
	// Return a new instance of RateLimitingQueueConfig
	return &RateLimitingQueueConfig{
		// 使用 NewDelayingQueueConfig 函数创建一个新的 DelayingQueueConfig 实例
		// Use the NewDelayingQueueConfig function to create a new instance of DelayingQueueConfig
		DelayingQueueConfig: *NewDelayingQueueConfig(),

		// 使用 NewNopRateLimitingQueueCallbackImpl 函数创建一个新的无操作的限流队列回调函数实例
		// Use the NewNopRateLimitingQueueCallbackImpl function to create a new instance of a no-operation rate limiting queue callback function
		callback: NewNopRateLimitingQueueCallbackImpl(),

		// 使用 NewNopRateLimiterImpl 函数创建一个新的无操作的限流器实例
		// Use the NewNopRateLimiterImpl function to create a new instance of a no-operation rate limiter
		limiter: NewNopRateLimiterImpl(),
	}
}

// WithCallback 方法用于设置限流队列的回调函数
// The WithCallback method is used to set the callback function of the rate limiting queue
func (c *RateLimitingQueueConfig) WithCallback(cb RateLimitingQueueCallback) *RateLimitingQueueConfig {
	// 设置回调函数
	// Set the callback function
	c.callback = cb
	c.DelayingQueueConfig.callback = cb
	c.DelayingQueueConfig.QueueConfig.callback = cb

	// 返回配置
	// Return the configuration
	return c
}

// WithLimiter 方法用于设置限流器
// The WithLimiter method is used to set the rate limiter
func (c *RateLimitingQueueConfig) WithLimiter(limiter Limiter) *RateLimitingQueueConfig {
	// 设置限流器
	// Set the rate limiter
	c.limiter = limiter

	// 返回配置
	// Return the configuration
	return c
}

// isRateLimitingQueueConfigEffective 函数用于检查限流队列配置是否有效，如果无效，则使用默认配置
// The isRateLimitingQueueConfigEffective function is used to check whether the rate limiting queue configuration is effective. If not, use the default configuration
func isRateLimitingQueueConfigEffective(c *RateLimitingQueueConfig) *RateLimitingQueueConfig {
	// 如果配置不为 nil
	// If the configuration is not nil
	if c != nil {
		// 如果回调函数为 nil，则设置为无操作的限流队列回调函数
		// If the callback function is nil, set it to a no-operation rate limiting queue callback function
		if c.callback == nil {
			c.callback = NewNopRateLimitingQueueCallbackImpl()
		}

		// 如果延迟队列配置的回调函数为 nil，则设置为无操作的延迟队列回调函数
		// If the callback function of the delaying queue configuration is nil, set it to a no-operation delaying queue callback function
		if c.DelayingQueueConfig.callback == nil {
			c.DelayingQueueConfig.callback = NewNopDelayingQueueCallbackImpl()
		}

		// 如果队列配置的回调函数为 nil，则设置为无操作的队列回调函数
		// If the callback function of the queue configuration is nil, set it to a no-operation queue callback function
		if c.DelayingQueueConfig.QueueConfig.callback == nil {
			c.DelayingQueueConfig.QueueConfig.callback = NewNopQueueCallbackImpl()
		}

		// 如果限流器为 nil，则设置为无操作的限流器
		// If the rate limiter is nil, set it to a no-operation rate limiter
		if c.limiter == nil {
			c.limiter = NewNopRateLimiterImpl()
		}
	} else {
		// 如果配置为 nil，则创建一个新的限流队列配置
		// If the configuration is nil, create a new rate limiting queue configuration
		c = NewRateLimitingQueueConfig()
	}

	// 返回配置
	// Return the configuration
	return c
}
