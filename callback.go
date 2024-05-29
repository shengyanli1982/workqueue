package workqueue

// QueueCallbackImpl 结构体定义了一个队列回调的实现。
// The QueueCallbackImpl struct defines an implementation of a queue callback.
type QueueCallbackImpl struct{}

// NewNopQueueCallbackImpl 函数创建并返回一个新的 QueueCallbackImpl 实例。
// The NewNopQueueCallbackImpl function creates and returns a new instance of QueueCallbackImpl.
func NewNopQueueCallbackImpl() *QueueCallbackImpl { return &QueueCallbackImpl{} }

// OnPut 方法在将元素放入队列时被调用，此处为空实现。
// The OnPut method is called when an element is put into the queue, here is an empty implementation.
func (impl *QueueCallbackImpl) OnPut(interface{}) {}

// OnGet 方法在从队列中获取元素时被调用，此处为空实现。
// The OnGet method is called when an element is gotten from the queue, here is an empty implementation.
func (impl *QueueCallbackImpl) OnGet(interface{}) {}

// OnDone 方法在元素处理完成后被调用，此处为空实现。
// The OnDone method is called when the element is done processing, here is an empty implementation.
func (impl *QueueCallbackImpl) OnDone(interface{}) {}

// DelayingQueueCallbackImpl 结构体继承了 QueueCallbackImpl，用于实现延迟队列的回调。
// The DelayingQueueCallbackImpl struct inherits from QueueCallbackImpl, used to implement callbacks for delaying queues.
type DelayingQueueCallbackImpl struct {
	QueueCallbackImpl
}

// NewNopDelayingQueueCallbackImpl 函数创建并返回一个新的 DelayingQueueCallbackImpl 实例。
// The NewNopDelayingQueueCallbackImpl function creates and returns a new instance of DelayingQueueCallbackImpl.
func NewNopDelayingQueueCallbackImpl() *DelayingQueueCallbackImpl {
	// 其内部的 QueueCallbackImpl 是空实现。
	// Its internal QueueCallbackImpl is an empty implementation.
	return &DelayingQueueCallbackImpl{
		QueueCallbackImpl: QueueCallbackImpl{},
	}
}

// DelayingQueueCallbackImpl 结构体的 OnDelay 方法在元素被延迟时被调用，此处为空实现。
// The OnDelay method of the DelayingQueueCallbackImpl struct is called when an element is delayed, here is an empty implementation.
func (impl *DelayingQueueCallbackImpl) OnDelay(interface{}, int64) {}

// OnPullError 方法在从队列中拉取元素出错时被调用，此处为空实现。
// The OnPullError method is called when there is an error pulling an element from the queue, here is an empty implementation.
func (impl *DelayingQueueCallbackImpl) OnPullError(interface{}, error) {}

// PriorityQueueCallbackImpl 结构体继承了 QueueCallbackImpl，用于实现优先队列的回调。
// The PriorityQueueCallbackImpl struct inherits from QueueCallbackImpl, used to implement callbacks for priority queues.
type PriorityQueueCallbackImpl struct {
	QueueCallbackImpl
}

// NewNopPriorityQueueCallbackImpl 函数创建并返回一个新的 PriorityQueueCallbackImpl 实例。
// The NewNopPriorityQueueCallbackImpl function creates and returns a new instance of PriorityQueueCallbackImpl.
func NewNopPriorityQueueCallbackImpl() *PriorityQueueCallbackImpl {
	// 其内部的 QueueCallbackImpl 是空实现。
	// Its internal QueueCallbackImpl is an empty implementation.
	return &PriorityQueueCallbackImpl{
		QueueCallbackImpl: QueueCallbackImpl{},
	}
}

// OnPriority 方法在元素的优先级被改变时被调用，此处为空实现。
// The OnPriority method is called when the priority of an element is changed, here is an empty implementation.
func (impl *PriorityQueueCallbackImpl) OnPriority(interface{}, int64) {}

// RateLimitingQueueCallbackImpl 结构体继承了 DelayingQueueCallbackImpl，用于实现限速队列的回调。
// The RateLimitingQueueCallbackImpl struct inherits from DelayingQueueCallbackImpl, used to implement callbacks for rate limiting queues.
type RateLimitingQueueCallbackImpl struct {
	DelayingQueueCallbackImpl
}

// NewNopRateLimitingQueueCallbackImpl 函数创建并返回一个新的 RateLimitingQueueCallbackImpl 实例。
// The NewNopRateLimitingQueueCallbackImpl function creates and returns a new instance of RateLimitingQueueCallbackImpl.
func NewNopRateLimitingQueueCallbackImpl() *RateLimitingQueueCallbackImpl {
	// 其内部的 DelayingQueueCallbackImpl 和 QueueCallbackImpl 都是空实现。
	// Its internal DelayingQueueCallbackImpl and QueueCallbackImpl are both empty implementations.
	return &RateLimitingQueueCallbackImpl{
		DelayingQueueCallbackImpl: DelayingQueueCallbackImpl{
			QueueCallbackImpl: QueueCallbackImpl{},
		},
	}
}

// OnLimited 方法在元素被限速时被调用，此处为空实现。
// The OnLimited method is called when an element is rate limited, here is an empty implementation.
func (impl *RateLimitingQueueCallbackImpl) OnLimited(interface{}) {}
