package workqueue

type QueueCallbackImpl struct{}

func NewNopQueueCallbackImpl() *QueueCallbackImpl { return &QueueCallbackImpl{} }

type DelayingQueueCallbackImpl struct {
	QueueCallbackImpl
}

func NewNopDelayingQueueCallbackImpl() *DelayingQueueCallbackImpl {
	return &DelayingQueueCallbackImpl{}
}

type PriorityQueueCallbackImpl struct {
	QueueCallbackImpl
}

func NewNopPriorityQueueCallbackImpl() *PriorityQueueCallbackImpl {
	return &PriorityQueueCallbackImpl{}
}

type RateLimitingQueueCallbackImpl struct {
	DelayingQueueCallbackImpl
}

func NewNopRateLimitingQueueCallbackImpl() *RateLimitingQueueCallbackImpl {
	return &RateLimitingQueueCallbackImpl{}
}
