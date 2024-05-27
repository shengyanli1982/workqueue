package workqueue

import "time"

type Queue interface {
	Put(interface{}) error
	Get() (interface{}, error)
	Done(interface{})
	Len() int
	Values() []interface{}
	Shutdown()
	IsClosed() bool
}

type DelayingQueue interface {
	Queue

	PutWithDelay(interface{}, int64) error
}

type PriorityQueue interface {
	Queue

	PutWithPriority(interface{}, int64) error
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

	OnDelay(interface{}, int64)
	OnPullError(interface{}, error)
}

type PriorityQueueCallback interface {
	QueueCallback

	OnPriority(interface{}, int64)
}

type RateLimitingQueueCallback interface {
	DelayingQueueCallback

	OnLimited(interface{})
}

type Limiter interface {
	When(interface{}) time.Duration
}
