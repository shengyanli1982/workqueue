package workqueue

type QueueConfig struct {
	callback QueueCallback
}

func NewQueueConfig() *QueueConfig {
	return &QueueConfig{
		callback: NewNopQueueCallbackImpl(),
	}
}

func (c *QueueConfig) WithCallback(cb QueueCallback) *QueueConfig {
	c.callback = cb
	return c
}

func isQueueConfigEffective(c *QueueConfig) *QueueConfig {
	if c != nil {
		if c.callback == nil {
			c.callback = NewNopQueueCallbackImpl()
		}
	} else {
		c = NewQueueConfig()
	}
	return c
}

type DelayingQueueConfig struct {
	callback DelayingQueueCallback
}

func NewDelayingQueueConfig() *DelayingQueueConfig {
	return &DelayingQueueConfig{
		callback: NewNopDelayingQueueCallbackImpl(),
	}
}

func (impl *DelayingQueueConfig) WithCallback(cb DelayingQueueCallback) *DelayingQueueConfig {
	impl.callback = cb
	return impl
}

func isDelayingQueueConfigEffective(c *DelayingQueueConfig) *DelayingQueueConfig {
	if c != nil {
		if c.callback == nil {
			c.callback = NewNopDelayingQueueCallbackImpl()
		}
	} else {
		c = NewDelayingQueueConfig()
	}
	return c
}

type PriorityQueueConfig struct {
	callback PriorityQueueCallback
}

func NewPriorityQueueConfig() *PriorityQueueConfig {
	return &PriorityQueueConfig{
		callback: NewNopPriorityQueueCallbackImpl(),
	}
}

func (c *PriorityQueueConfig) WithCallback(cb PriorityQueueCallback) *PriorityQueueConfig {
	c.callback = cb
	return c
}

func isPriorityQueueConfigEffective(c *PriorityQueueConfig) *PriorityQueueConfig {
	if c != nil {
		if c.callback == nil {
			c.callback = NewNopPriorityQueueCallbackImpl()
		}
	} else {
		c = NewPriorityQueueConfig()
	}
	return c
}

type RateLimitingQueueConfig struct {
	callback RateLimitingQueueCallback

	limiter Limiter
}

func NewRateLimitingQueueConfig() *RateLimitingQueueConfig {
	return &RateLimitingQueueConfig{
		callback: NewNopRateLimitingQueueCallbackImpl(),
		limiter:  NewNopRateLimiterImpl(),
	}
}

func (c *RateLimitingQueueConfig) WithCallback(cb RateLimitingQueueCallback) *RateLimitingQueueConfig {
	c.callback = cb
	return c
}

func (c *RateLimitingQueueConfig) WithLimiter(limiter Limiter) *RateLimitingQueueConfig {
	c.limiter = limiter
	return c
}

func isRateLimitingQueueConfigEffective(c *RateLimitingQueueConfig) *RateLimitingQueueConfig {
	if c != nil {
		if c.callback == nil {
			c.callback = NewNopRateLimitingQueueCallbackImpl()
		}
		if c.limiter == nil {
			c.limiter = NewNopRateLimiterImpl()
		}
	} else {
		c = NewRateLimitingQueueConfig()
	}
	return c
}
