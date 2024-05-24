package workqueue

import "time"

type QueueCallbackImpl struct{}

func NewNopQueueCallbackImpl() *QueueCallbackImpl { return &QueueCallbackImpl{} }

func (impl *QueueCallbackImpl) OnPut(interface{})  {}
func (impl *QueueCallbackImpl) OnGet(interface{})  {}
func (impl *QueueCallbackImpl) OnDone(interface{}) {}

type DelayingQueueCallbackImpl struct {
	QueueCallbackImpl
}

func NewNopDelayingQueueCallbackImpl() *DelayingQueueCallbackImpl {
	return &DelayingQueueCallbackImpl{
		QueueCallbackImpl: QueueCallbackImpl{},
	}
}

func (impl *DelayingQueueCallbackImpl) OnDelay(interface{}, time.Duration) {}
func (impl *DelayingQueueCallbackImpl) OnPullError(interface{}, error)     {}

type PriorityQueueCallbackImpl struct {
	QueueCallbackImpl
}

func NewNopPriorityQueueCallbackImpl() *PriorityQueueCallbackImpl {
	return &PriorityQueueCallbackImpl{
		QueueCallbackImpl: QueueCallbackImpl{},
	}
}

func (impl *PriorityQueueCallbackImpl) OnPriority(interface{}, int) {}

type RateLimitingQueueCallbackImpl struct {
	DelayingQueueCallbackImpl
}

func NewNopRateLimitingQueueCallbackImpl() *RateLimitingQueueCallbackImpl {
	return &RateLimitingQueueCallbackImpl{
		DelayingQueueCallbackImpl: DelayingQueueCallbackImpl{
			QueueCallbackImpl: QueueCallbackImpl{},
		},
	}
}

func (impl *RateLimitingQueueCallbackImpl) OnLimited(interface{}) {}
