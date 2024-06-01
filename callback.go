package workqueue

// queueCallbackImpl 结构体定义了一个队列回调的实现。
// The queueCallbackImpl struct defines an implementation of a queue callback.
type queueCallbackImpl struct{}

// NewNopQueueCallbackImpl 函数创建并返回一个新的 QueueCallbackImpl 实例。
// The NewNopQueueCallbackImpl function creates and returns a new instance of QueueCallbackImpl.
func NewNopQueueCallbackImpl() *queueCallbackImpl { return &queueCallbackImpl{} }

// OnPut 方法在将元素放入队列时被调用，此处为空实现。
// The OnPut method is called when an element is put into the queue, here is an empty implementation.
func (impl *queueCallbackImpl) OnPut(interface{}) {}

// OnGet 方法在从队列中获取元素时被调用，此处为空实现。
// The OnGet method is called when an element is gotten from the queue, here is an empty implementation.
func (impl *queueCallbackImpl) OnGet(interface{}) {}

// OnDone 方法在元素处理完成后被调用，此处为空实现。
// The OnDone method is called when the element is done processing, here is an empty implementation.
func (impl *queueCallbackImpl) OnDone(interface{}) {}

// delayingQueueCallbackImpl 结构体继承了 QueueCallbackImpl，用于实现延迟队列的回调。
// The delayingQueueCallbackImpl struct inherits from QueueCallbackImpl, used to implement callbacks for delaying queues.
type delayingQueueCallbackImpl struct {
	queueCallbackImpl
}

// NewNopDelayingQueueCallbackImpl 函数创建并返回一个新的 DelayingQueueCallbackImpl 实例。
// The NewNopDelayingQueueCallbackImpl function creates and returns a new instance of DelayingQueueCallbackImpl.
func NewNopDelayingQueueCallbackImpl() *delayingQueueCallbackImpl {
	// 其内部的 QueueCallbackImpl 是空实现。
	// Its internal QueueCallbackImpl is an empty implementation.
	return &delayingQueueCallbackImpl{
		queueCallbackImpl: queueCallbackImpl{},
	}
}

// DelayingQueueCallbackImpl 结构体的 OnDelay 方法在元素被延迟时被调用，此处为空实现。
// The OnDelay method of the DelayingQueueCallbackImpl struct is called when an element is delayed, here is an empty implementation.
func (impl *delayingQueueCallbackImpl) OnDelay(interface{}, int64) {}

// OnPullError 方法在从队列中拉取元素出错时被调用，此处为空实现。
// The OnPullError method is called when there is an error pulling an element from the queue, here is an empty implementation.
func (impl *delayingQueueCallbackImpl) OnPullError(interface{}, error) {}

// priorityQueueCallbackImpl 结构体继承了 QueueCallbackImpl，用于实现优先队列的回调。
// The priorityQueueCallbackImpl struct inherits from QueueCallbackImpl, used to implement callbacks for priority queues.
type priorityQueueCallbackImpl struct {
	queueCallbackImpl
}

// NewNopPriorityQueueCallbackImpl 函数创建并返回一个新的 PriorityQueueCallbackImpl 实例。
// The NewNopPriorityQueueCallbackImpl function creates and returns a new instance of PriorityQueueCallbackImpl.
func NewNopPriorityQueueCallbackImpl() *priorityQueueCallbackImpl {
	// 其内部的 QueueCallbackImpl 是空实现。
	// Its internal QueueCallbackImpl is an empty implementation.
	return &priorityQueueCallbackImpl{
		queueCallbackImpl: queueCallbackImpl{},
	}
}

// OnPriority 方法在元素的优先级被改变时被调用，此处为空实现。
// The OnPriority method is called when the priority of an element is changed, here is an empty implementation.
func (impl *priorityQueueCallbackImpl) OnPriority(interface{}, int64) {}

// ratelimitingQueueCallbackImpl 结构体继承了 DelayingQueueCallbackImpl，用于实现限速队列的回调。
// The ratelimitingQueueCallbackImpl struct inherits from DelayingQueueCallbackImpl, used to implement callbacks for rate limiting queues.
type ratelimitingQueueCallbackImpl struct {
	delayingQueueCallbackImpl
}

// NewNopRateLimitingQueueCallbackImpl 函数创建并返回一个新的 RateLimitingQueueCallbackImpl 实例。
// The NewNopRateLimitingQueueCallbackImpl function creates and returns a new instance of RateLimitingQueueCallbackImpl.
func NewNopRateLimitingQueueCallbackImpl() *ratelimitingQueueCallbackImpl {
	// 其内部的 DelayingQueueCallbackImpl 和 QueueCallbackImpl 都是空实现。
	// Its internal DelayingQueueCallbackImpl and QueueCallbackImpl are both empty implementations.
	return &ratelimitingQueueCallbackImpl{
		delayingQueueCallbackImpl: delayingQueueCallbackImpl{
			queueCallbackImpl: queueCallbackImpl{},
		},
	}
}

// OnLimited 方法在元素被限速时被调用，此处为空实现。
// The OnLimited method is called when an element is rate limited, here is an empty implementation.
func (impl *ratelimitingQueueCallbackImpl) OnLimited(interface{}) {}
