package workqueue

import "time"

type queueCallbackImpl struct{}

// NewNopQueueCallbackImpl 返回空实现回调。
func NewNopQueueCallbackImpl() *queueCallbackImpl { return &queueCallbackImpl{} }

func (impl *queueCallbackImpl) OnPut(interface{}) {}

func (impl *queueCallbackImpl) OnGet(interface{}) {}

func (impl *queueCallbackImpl) OnDone(interface{}) {}

type delayingQueueCallbackImpl struct {
	queueCallbackImpl
}

// NewNopDelayingQueueCallbackImpl 返回空实现延迟回调。
func NewNopDelayingQueueCallbackImpl() *delayingQueueCallbackImpl {

	return &delayingQueueCallbackImpl{
		queueCallbackImpl: queueCallbackImpl{},
	}
}

func (impl *delayingQueueCallbackImpl) OnDelay(interface{}, int64) {}

func (impl *delayingQueueCallbackImpl) OnPullError(interface{}, error) {}

type timerQueueCallbackImpl struct {
	queueCallbackImpl
}

// NewNopTimerQueueCallbackImpl 返回空实现定时回调。
func NewNopTimerQueueCallbackImpl() *timerQueueCallbackImpl {

	return &timerQueueCallbackImpl{
		queueCallbackImpl: queueCallbackImpl{},
	}
}

func (impl *timerQueueCallbackImpl) OnSchedule(interface{}, int64) {}

func (impl *timerQueueCallbackImpl) OnScheduleError(interface{}, error) {}

type priorityQueueCallbackImpl struct {
	queueCallbackImpl
}

// NewNopPriorityQueueCallbackImpl 返回空实现优先级回调。
func NewNopPriorityQueueCallbackImpl() *priorityQueueCallbackImpl {

	return &priorityQueueCallbackImpl{
		queueCallbackImpl: queueCallbackImpl{},
	}
}

func (impl *priorityQueueCallbackImpl) OnPriority(interface{}, int64) {}

type ratelimitingQueueCallbackImpl struct {
	delayingQueueCallbackImpl
}

// NewNopRateLimitingQueueCallbackImpl 返回空实现限流回调。
func NewNopRateLimitingQueueCallbackImpl() *ratelimitingQueueCallbackImpl {

	return &ratelimitingQueueCallbackImpl{
		delayingQueueCallbackImpl: delayingQueueCallbackImpl{
			queueCallbackImpl: queueCallbackImpl{},
		},
	}
}

func (impl *ratelimitingQueueCallbackImpl) OnLimited(interface{}) {}

type retryQueueCallbackImpl struct {
	delayingQueueCallbackImpl
}

// NewNopRetryQueueCallbackImpl 返回空实现重试回调。
func NewNopRetryQueueCallbackImpl() *retryQueueCallbackImpl {

	return &retryQueueCallbackImpl{
		delayingQueueCallbackImpl: delayingQueueCallbackImpl{
			queueCallbackImpl: queueCallbackImpl{},
		},
	}
}

func (impl *retryQueueCallbackImpl) OnRetry(interface{}, int, time.Duration, error) {}

func (impl *retryQueueCallbackImpl) OnRetryExhausted(interface{}, int, error) {}

func (impl *retryQueueCallbackImpl) OnForget(interface{}) {}

type deadLetterQueueCallbackImpl struct {
	queueCallbackImpl
}

// NewNopDeadLetterQueueCallbackImpl 返回空实现死信回调。
func NewNopDeadLetterQueueCallbackImpl() *deadLetterQueueCallbackImpl {

	return &deadLetterQueueCallbackImpl{
		queueCallbackImpl: queueCallbackImpl{},
	}
}

func (impl *deadLetterQueueCallbackImpl) OnDead(*DeadLetter) {}

func (impl *deadLetterQueueCallbackImpl) OnAckDead(*DeadLetter) {}

func (impl *deadLetterQueueCallbackImpl) OnRequeueDead(*DeadLetter, Queue) {}
