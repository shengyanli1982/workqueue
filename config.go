package workqueue

type QueueConfigImpl struct {
	callback QueueCallback
}

func NewQueueConfigImpl() *QueueConfigImpl { return &QueueConfigImpl{} }

func (impl *QueueConfigImpl) WithCallback(cb QueueCallback) *QueueConfigImpl {
	impl.callback = cb
	return impl
}

func (impl *QueueConfigImpl) Effective() {
	if impl.callback == nil {
		impl.callback = NewNopQueueCallbackImpl()
	}
}

type DelayingQueueConfigImpl struct {
	callback DelayingQueueCallback
}

func (impl *DelayingQueueConfigImpl) New() *DelayingQueueConfigImpl {
	return &DelayingQueueConfigImpl{}
}

func (impl *DelayingQueueConfigImpl) WithCallback(cb DelayingQueueCallback) *DelayingQueueConfigImpl {
	impl.callback = cb
	return impl
}

func (impl *DelayingQueueConfigImpl) Effective() {
	if impl.callback == nil {
		impl.callback = NewNopDelayingQueueCallbackImpl()
	}
}

type PriorityQueueConfigImpl struct {
	QueueConfigImpl
}

func (impl *PriorityQueueConfigImpl) New() *PriorityQueueConfigImpl {
	return &PriorityQueueConfigImpl{}
}

func (impl *PriorityQueueConfigImpl) WithCallback(cb PriorityQueueCallback) *PriorityQueueConfigImpl {
	impl.callback = cb
	return impl
}

func (impl *PriorityQueueConfigImpl) Effective() {
	if impl.callback == nil {
		impl.callback = NewNopPriorityQueueCallbackImpl()
	}
}

type RateLimitingQueueConfigImpl struct {
	DelayingQueueConfigImpl
}

func (impl *RateLimitingQueueConfigImpl) New() *RateLimitingQueueConfigImpl {
	return &RateLimitingQueueConfigImpl{}
}

func (impl *RateLimitingQueueConfigImpl) WithCallback(cb RateLimitingQueueCallback) *RateLimitingQueueConfigImpl {
	impl.callback = cb
	return impl
}

func (impl *RateLimitingQueueConfigImpl) Effective() {
	if impl.callback == nil {
		impl.callback = NewNopRateLimitingQueueCallbackImpl()
	}
}
