package workqueue

import "time"

type queueCallbackImpl struct{}

// NewNopQueueCallbackImpl 返回空实现回调。
func NewNopQueueCallbackImpl() *queueCallbackImpl { return &queueCallbackImpl{} }

func (impl *queueCallbackImpl) OnPut(any) {}

func (impl *queueCallbackImpl) OnGet(any) {}

func (impl *queueCallbackImpl) OnDone(any) {}

type delayingQueueCallbackImpl struct {
	queueCallbackImpl
}

// NewNopDelayingQueueCallbackImpl 返回空实现延迟回调。
func NewNopDelayingQueueCallbackImpl() *delayingQueueCallbackImpl {

	return &delayingQueueCallbackImpl{
		queueCallbackImpl: queueCallbackImpl{},
	}
}

func (impl *delayingQueueCallbackImpl) OnDelay(any, int64) {}

func (impl *delayingQueueCallbackImpl) OnPullError(any, error) {}

type timerQueueCallbackImpl struct {
	queueCallbackImpl
}

// NewNopTimerQueueCallbackImpl 返回空实现定时回调。
func NewNopTimerQueueCallbackImpl() *timerQueueCallbackImpl {

	return &timerQueueCallbackImpl{
		queueCallbackImpl: queueCallbackImpl{},
	}
}

func (impl *timerQueueCallbackImpl) OnSchedule(any, int64) {}

func (impl *timerQueueCallbackImpl) OnScheduleError(any, error) {}

type priorityQueueCallbackImpl struct {
	queueCallbackImpl
}

// NewNopPriorityQueueCallbackImpl 返回空实现优先级回调。
func NewNopPriorityQueueCallbackImpl() *priorityQueueCallbackImpl {

	return &priorityQueueCallbackImpl{
		queueCallbackImpl: queueCallbackImpl{},
	}
}

func (impl *priorityQueueCallbackImpl) OnPriority(any, int64) {}

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

func (impl *ratelimitingQueueCallbackImpl) OnLimited(any) {}

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

func (impl *retryQueueCallbackImpl) OnRetry(any, int, time.Duration, error) {}

func (impl *retryQueueCallbackImpl) OnRetryExhausted(any, int, error) {}

func (impl *retryQueueCallbackImpl) OnForget(any) {}

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

type leasedQueueCallbackImpl struct {
	queueCallbackImpl
}

// NewNopLeasedQueueCallbackImpl 返回空实现租约回调。
func NewNopLeasedQueueCallbackImpl() *leasedQueueCallbackImpl {

	return &leasedQueueCallbackImpl{
		queueCallbackImpl: queueCallbackImpl{},
	}
}

func (impl *leasedQueueCallbackImpl) OnNack(any, error) {}
