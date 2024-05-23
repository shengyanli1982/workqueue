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
	QueueConfig

	callback DelayingQueueCallback
}

func NewDelayingQueueConfig() *DelayingQueueConfig {
	return &DelayingQueueConfig{
		QueueConfig: *NewQueueConfig(),
		callback:    NewNopDelayingQueueCallbackImpl(),
	}
}

func (c *DelayingQueueConfig) WithCallback(cb DelayingQueueCallback) *DelayingQueueConfig {
	c.callback = cb
	c.QueueConfig.callback = cb
	return c
}

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

type PriorityQueueConfig struct {
	QueueConfig

	callback PriorityQueueCallback
}

func NewPriorityQueueConfig() *PriorityQueueConfig {
	return &PriorityQueueConfig{
		QueueConfig: *NewQueueConfig(),
		callback:    NewNopPriorityQueueCallbackImpl(),
	}
}

func (c *PriorityQueueConfig) WithCallback(cb PriorityQueueCallback) *PriorityQueueConfig {
	c.callback = cb
	c.QueueConfig.callback = cb
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
	DelayingQueueConfig

	callback RateLimitingQueueCallback

	limiter Limiter
}

func NewRateLimitingQueueConfig() *RateLimitingQueueConfig {
	return &RateLimitingQueueConfig{
		DelayingQueueConfig: *NewDelayingQueueConfig(),
		callback:            NewNopRateLimitingQueueCallbackImpl(),
		limiter:             NewNopRateLimiterImpl(),
	}
}

func (c *RateLimitingQueueConfig) WithCallback(cb RateLimitingQueueCallback) *RateLimitingQueueConfig {
	c.callback = cb
	c.DelayingQueueConfig.callback = cb
	c.DelayingQueueConfig.QueueConfig.callback = cb
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
		if c.DelayingQueueConfig.callback == nil {
			c.DelayingQueueConfig.callback = NewNopDelayingQueueCallbackImpl()
		}
		if c.DelayingQueueConfig.QueueConfig.callback == nil {
			c.DelayingQueueConfig.QueueConfig.callback = NewNopQueueCallbackImpl()
		}
		if c.limiter == nil {
			c.limiter = NewNopRateLimiterImpl()
		}
	} else {
		c = NewRateLimitingQueueConfig()
	}
	return c
}
