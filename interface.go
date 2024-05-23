package workqueue

type Queue interface {
}

type DelayingQueue interface {
	Queue
}

type PriorityQueue interface {
	Queue
}

type RateLimitingQueue interface {
	DelayingQueue
}

type QueueConfig interface {
}

type DelayingQueueConfig interface {
	QueueConfig
}

type PriorityQueueConfig interface {
	QueueConfig
}

type RateLimitingQueueConfig interface {
	DelayingQueueConfig
}

type QueueCallback interface {
}

type DelayingQueueCallback interface {
	QueueCallback
}

type PriorityQueueCallback interface {
	QueueCallback
}

type RateLimitingQueueCallback interface {
	DelayingQueueCallback
}
