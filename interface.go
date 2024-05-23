package workqueue

import "time"

type Queue interface {
	Put(interface{}) error
	Len() int
	Get() (interface{}, error)
	Values() []interface{}
	Shutdown()
	IsClosed() bool
}

type DelayingQueue interface {
	Queue

	PutWithDelay(interface{}, time.Duration) error
}

type PriorityQueue interface {
	Queue

	PutWithPriority(interface{}, int) error
}

type RateLimitingQueue interface {
	DelayingQueue

	PutWithLimited(interface{}) error
}

type QueueCallback interface {
	OnPut(interface{})
	OnGet(interface{})
	OnDone(interface{})
}

type DelayingQueueCallback interface {
	QueueCallback

	OnDelay(interface{}, time.Duration)
}

type PriorityQueueCallback interface {
	QueueCallback

	OnPriority(interface{}, int)
}

type RateLimitingQueueCallback interface {
	DelayingQueueCallback

	OnLimited(interface{})
}

type QueueConfig interface {
	WithCallback(QueueCallback) QueueConfig
	Effective()
}

type DelayingQueueConfig interface {
	WithCallback(DelayingQueueCallback) DelayingQueueConfig
	Effective()
}

type PriorityQueueConfig interface {
	WithCallback(PriorityQueueCallback) PriorityQueueConfig
	Effective()
}

type Limiter interface {
	When() time.Duration
}

type RateLimitingQueueConfig interface {
	WithCallback(RateLimitingQueueCallback) RateLimitingQueueConfig
	WithRateLimiter(Limiter) RateLimitingQueueConfig
	Effective()
}
