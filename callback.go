package workqueue

import "time"

// queueCallbackImpl 是 QueueCallback 接口的空实现基类，
// 可作为各队列类型回调的默认值使用。
type queueCallbackImpl struct{}

// NewNopQueueCallbackImpl 返回空实现回调。
func NewNopQueueCallbackImpl() *queueCallbackImpl { return &queueCallbackImpl{} }

// OnPut 为空实现。
func (impl *queueCallbackImpl) OnPut(any) {}

// OnGet 为空实现。
func (impl *queueCallbackImpl) OnGet(any) {}

// OnDone 为空实现。
func (impl *queueCallbackImpl) OnDone(any) {}

// delayingQueueCallbackImpl 是 DelayingQueueCallback 接口的空实现，
// 内嵌 queueCallbackImpl 继承基础回调方法。
type delayingQueueCallbackImpl struct {
	queueCallbackImpl
}

// NewNopDelayingQueueCallbackImpl 返回空实现延迟回调。
func NewNopDelayingQueueCallbackImpl() *delayingQueueCallbackImpl {

	return &delayingQueueCallbackImpl{
		queueCallbackImpl: queueCallbackImpl{},
	}
}

// OnDelay 为空实现。
func (impl *delayingQueueCallbackImpl) OnDelay(any, int64) {}

// OnPullError 为空实现。
func (impl *delayingQueueCallbackImpl) OnPullError(any, error) {}

// timerQueueCallbackImpl 是 TimerQueueCallback 接口的空实现，
// 内嵌 queueCallbackImpl 继承基础回调方法。
type timerQueueCallbackImpl struct {
	queueCallbackImpl
}

// NewNopTimerQueueCallbackImpl 返回空实现定时回调。
func NewNopTimerQueueCallbackImpl() *timerQueueCallbackImpl {

	return &timerQueueCallbackImpl{
		queueCallbackImpl: queueCallbackImpl{},
	}
}

// OnSchedule 为空实现。
func (impl *timerQueueCallbackImpl) OnSchedule(any, int64) {}

// OnScheduleError 为空实现。
func (impl *timerQueueCallbackImpl) OnScheduleError(any, error) {}

// priorityQueueCallbackImpl 是 PriorityQueueCallback 接口的空实现，
// 内嵌 queueCallbackImpl 继承基础回调方法。
type priorityQueueCallbackImpl struct {
	queueCallbackImpl
}

// NewNopPriorityQueueCallbackImpl 返回空实现优先级回调。
func NewNopPriorityQueueCallbackImpl() *priorityQueueCallbackImpl {

	return &priorityQueueCallbackImpl{
		queueCallbackImpl: queueCallbackImpl{},
	}
}

// OnPriority 为空实现。
func (impl *priorityQueueCallbackImpl) OnPriority(any, int64) {}

// ratelimitingQueueCallbackImpl 是 RateLimitingQueueCallback 接口的空实现，
// 内嵌 delayingQueueCallbackImpl 继承延迟队列回调方法。
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

// OnLimited 为空实现。
func (impl *ratelimitingQueueCallbackImpl) OnLimited(any) {}

// retryQueueCallbackImpl 是 RetryQueueCallback 接口的空实现，
// 内嵌 delayingQueueCallbackImpl 继承延迟队列回调方法。
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

// OnRetry 为空实现。
func (impl *retryQueueCallbackImpl) OnRetry(any, int, time.Duration, error) {}

// OnRetryExhausted 为空实现。
func (impl *retryQueueCallbackImpl) OnRetryExhausted(any, int, error) {}

// OnForget 为空实现。
func (impl *retryQueueCallbackImpl) OnForget(any) {}

// deadLetterQueueCallbackImpl 是 DeadLetterQueueCallback 接口的空实现，
// 内嵌 queueCallbackImpl 继承基础回调方法。
type deadLetterQueueCallbackImpl struct {
	queueCallbackImpl
}

// NewNopDeadLetterQueueCallbackImpl 返回空实现死信回调。
func NewNopDeadLetterQueueCallbackImpl() *deadLetterQueueCallbackImpl {

	return &deadLetterQueueCallbackImpl{
		queueCallbackImpl: queueCallbackImpl{},
	}
}

// OnDead 为空实现。
func (impl *deadLetterQueueCallbackImpl) OnDead(*DeadLetter) {}

// OnAckDead 为空实现。
func (impl *deadLetterQueueCallbackImpl) OnAckDead(*DeadLetter) {}

// OnRequeueDead 为空实现。
func (impl *deadLetterQueueCallbackImpl) OnRequeueDead(*DeadLetter, Queue) {}

// leasedQueueCallbackImpl 是 LeasedQueueCallback 接口的空实现，
// 内嵌 queueCallbackImpl 继承基础回调方法。
type leasedQueueCallbackImpl struct {
	queueCallbackImpl
}

// NewNopLeasedQueueCallbackImpl 返回空实现租约回调。
func NewNopLeasedQueueCallbackImpl() *leasedQueueCallbackImpl {

	return &leasedQueueCallbackImpl{
		queueCallbackImpl: queueCallbackImpl{},
	}
}

// OnNack 为空实现。
func (impl *leasedQueueCallbackImpl) OnNack(any, error) {}
