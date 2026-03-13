package workqueue

import (
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue/v2/internal/container/set"
)

// NewSetFunc 用于自定义幂等集合实现。
type NewSetFunc = func() Set

var defaultNewSetFunc = func() Set { return set.New() }

var defaultRetryKeyFunc = func(value interface{}) string {
	if value == nil {
		return ""
	}

	return fmt.Sprintf("%T:%#v", value, value)
}

// QueueConfig 定义基础队列配置。
type QueueConfig struct {
	callback   QueueCallback
	idempotent bool
	setCreator NewSetFunc
}

// NewQueueConfig 返回带默认值的基础队列配置。
func NewQueueConfig() *QueueConfig {
	return &QueueConfig{
		callback:   NewNopQueueCallbackImpl(),
		setCreator: defaultNewSetFunc,
	}
}

// WithCallback 设置基础回调。
func (c *QueueConfig) WithCallback(cb QueueCallback) *QueueConfig {
	c.callback = cb
	return c
}

// WithValueIdempotent 开启值幂等模式。
func (c *QueueConfig) WithValueIdempotent() *QueueConfig {
	c.idempotent = true

	return c
}

// WithSetCreator 设置幂等集合构造器。
func (c *QueueConfig) WithSetCreator(fn NewSetFunc) *QueueConfig {
	c.setCreator = fn

	return c
}

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

// DelayingQueueConfig 定义延迟队列配置。
type DelayingQueueConfig struct {
	QueueConfig
	callback DelayingQueueCallback
}

// NewDelayingQueueConfig 返回带默认值的延迟队列配置。
func NewDelayingQueueConfig() *DelayingQueueConfig {
	return &DelayingQueueConfig{
		QueueConfig: *NewQueueConfig(),

		callback: NewNopDelayingQueueCallbackImpl(),
	}
}

// WithCallback 设置延迟队列回调。
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

// PriorityQueueConfig 定义优先级队列配置。
type PriorityQueueConfig struct {
	QueueConfig
	callback PriorityQueueCallback
}

// NewPriorityQueueConfig 返回带默认值的优先级队列配置。
func NewPriorityQueueConfig() *PriorityQueueConfig {
	return &PriorityQueueConfig{
		QueueConfig: *NewQueueConfig(),

		callback: NewNopPriorityQueueCallbackImpl(),
	}
}

// WithCallback 设置优先级队列回调。
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

		if c.QueueConfig.callback == nil {
			c.QueueConfig.callback = NewNopQueueCallbackImpl()
		}
	} else {
		c = NewPriorityQueueConfig()
	}

	return c
}

// LeasedQueueConfig 定义租约队列配置。
type LeasedQueueConfig struct {
	QueueConfig
	leaseDuration time.Duration
	scanInterval  time.Duration
}

// NewLeasedQueueConfig 返回带默认值的租约队列配置。
func NewLeasedQueueConfig() *LeasedQueueConfig {
	return &LeasedQueueConfig{
		QueueConfig:   *NewQueueConfig(),
		leaseDuration: 30 * time.Second,
		scanInterval:  100 * time.Millisecond,
	}
}

// WithLeaseDuration 设置默认租约时长。
func (c *LeasedQueueConfig) WithLeaseDuration(duration time.Duration) *LeasedQueueConfig {
	c.leaseDuration = duration
	return c
}

// WithScanInterval 设置租约扫描间隔。
func (c *LeasedQueueConfig) WithScanInterval(interval time.Duration) *LeasedQueueConfig {
	c.scanInterval = interval
	return c
}

func isLeasedQueueConfigEffective(c *LeasedQueueConfig) *LeasedQueueConfig {
	if c != nil {
		c.QueueConfig = *isQueueConfigEffective(&c.QueueConfig)

		if c.leaseDuration <= 0 {
			c.leaseDuration = 30 * time.Second
		}
		if c.scanInterval <= 0 {
			c.scanInterval = 100 * time.Millisecond
		}
	} else {
		c = NewLeasedQueueConfig()
	}
	return c
}

// BoundedBlockingQueueConfig 定义有界阻塞队列配置。
type BoundedBlockingQueueConfig struct {
	QueueConfig
	capacity int
}

// NewBoundedBlockingQueueConfig 返回带默认值的有界阻塞队列配置。
func NewBoundedBlockingQueueConfig() *BoundedBlockingQueueConfig {
	return &BoundedBlockingQueueConfig{
		QueueConfig: *NewQueueConfig(),
		capacity:    1024,
	}
}

// WithCapacity 设置队列容量上限。
func (c *BoundedBlockingQueueConfig) WithCapacity(capacity int) *BoundedBlockingQueueConfig {
	c.capacity = capacity
	return c
}

func isBoundedBlockingQueueConfigEffective(c *BoundedBlockingQueueConfig) *BoundedBlockingQueueConfig {
	if c != nil {
		c.QueueConfig = *isQueueConfigEffective(&c.QueueConfig)

		if c.capacity <= 0 {
			c.capacity = 1024
		}
	} else {
		c = NewBoundedBlockingQueueConfig()
	}

	return c
}

// TimerQueueConfig 定义定时队列配置。
type TimerQueueConfig struct {
	QueueConfig
}

// NewTimerQueueConfig 返回带默认值的定时队列配置。
func NewTimerQueueConfig() *TimerQueueConfig {
	return &TimerQueueConfig{
		QueueConfig: *NewQueueConfig(),
	}
}

func isTimerQueueConfigEffective(c *TimerQueueConfig) *TimerQueueConfig {
	if c != nil {
		c.QueueConfig = *isQueueConfigEffective(&c.QueueConfig)
	} else {
		c = NewTimerQueueConfig()
	}
	return c
}

// RetryQueueConfig 定义重试队列配置。
type RetryQueueConfig struct {
	DelayingQueueConfig
	callback RetryQueueCallback
	policy   RetryPolicy
	keyFunc  RetryKeyFunc
}

// NewRetryQueueConfig 返回带默认值的重试队列配置。
func NewRetryQueueConfig() *RetryQueueConfig {
	return &RetryQueueConfig{
		DelayingQueueConfig: *NewDelayingQueueConfig(),
		callback:            NewNopRetryQueueCallbackImpl(),
		policy:              NewExponentialRetryPolicy(100*time.Millisecond, 30*time.Second, 5),
		keyFunc:             defaultRetryKeyFunc,
	}
}

// WithCallback 设置重试队列回调。
func (c *RetryQueueConfig) WithCallback(cb RetryQueueCallback) *RetryQueueConfig {
	c.callback = cb
	c.DelayingQueueConfig.callback = cb
	c.DelayingQueueConfig.QueueConfig.callback = cb
	return c
}

// WithPolicy 设置重试策略。
func (c *RetryQueueConfig) WithPolicy(policy RetryPolicy) *RetryQueueConfig {
	c.policy = policy
	return c
}

// WithKeyFunc 设置重试 key 生成函数。
func (c *RetryQueueConfig) WithKeyFunc(fn RetryKeyFunc) *RetryQueueConfig {
	c.keyFunc = fn
	return c
}

func isRetryQueueConfigEffective(c *RetryQueueConfig) *RetryQueueConfig {
	if c != nil {
		c.DelayingQueueConfig = *isDelayingQueueConfigEffective(&c.DelayingQueueConfig)

		if c.callback == nil {
			c.callback = NewNopRetryQueueCallbackImpl()
		}
		c.DelayingQueueConfig.callback = c.callback
		c.DelayingQueueConfig.QueueConfig.callback = c.callback

		if c.policy == nil {
			c.policy = NewExponentialRetryPolicy(100*time.Millisecond, 30*time.Second, 5)
		}
		if c.keyFunc == nil {
			c.keyFunc = defaultRetryKeyFunc
		}
	} else {
		c = NewRetryQueueConfig()
	}

	return c
}

// DeadLetterQueueConfig 定义死信队列配置。
type DeadLetterQueueConfig struct {
	QueueConfig
	callback DeadLetterQueueCallback
}

// NewDeadLetterQueueConfig 返回带默认值的死信队列配置。
func NewDeadLetterQueueConfig() *DeadLetterQueueConfig {
	return &DeadLetterQueueConfig{
		QueueConfig: *NewQueueConfig(),
		callback:    NewNopDeadLetterQueueCallbackImpl(),
	}
}

// WithCallback 设置死信队列回调。
func (c *DeadLetterQueueConfig) WithCallback(cb DeadLetterQueueCallback) *DeadLetterQueueConfig {
	c.callback = cb
	c.QueueConfig.callback = cb
	return c
}

func isDeadLetterQueueConfigEffective(c *DeadLetterQueueConfig) *DeadLetterQueueConfig {
	if c != nil {
		c.QueueConfig = *isQueueConfigEffective(&c.QueueConfig)

		if c.callback == nil {
			c.callback = NewNopDeadLetterQueueCallbackImpl()
		}
		c.QueueConfig.callback = c.callback
	} else {
		c = NewDeadLetterQueueConfig()
	}

	return c
}

// RateLimitingQueueConfig 定义限流队列配置。
type RateLimitingQueueConfig struct {
	DelayingQueueConfig
	callback RateLimitingQueueCallback
	limiter  Limiter
}

// NewRateLimitingQueueConfig 返回带默认值的限流队列配置。
func NewRateLimitingQueueConfig() *RateLimitingQueueConfig {
	return &RateLimitingQueueConfig{
		DelayingQueueConfig: *NewDelayingQueueConfig(),

		callback: NewNopRateLimitingQueueCallbackImpl(),

		limiter: NewNopRateLimiterImpl(),
	}
}

// WithCallback 设置限流队列回调。
func (c *RateLimitingQueueConfig) WithCallback(cb RateLimitingQueueCallback) *RateLimitingQueueConfig {
	c.callback = cb
	c.DelayingQueueConfig.callback = cb
	c.DelayingQueueConfig.QueueConfig.callback = cb

	return c
}

// WithLimiter 设置限流器实现。
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
