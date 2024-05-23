package workqueue

type QueueConfigImpl struct {
}

func NewQueueConfigImpl() *QueueConfigImpl { return &QueueConfigImpl{} }

func (impl *QueueConfigImpl) Effective() {

}

type DelayingQueueConfigImpl struct {
	QueueConfigImpl
}

func (impl *DelayingQueueConfigImpl) New() *DelayingQueueConfigImpl {
	return &DelayingQueueConfigImpl{}
}

func (impl *DelayingQueueConfigImpl) Effective() {

}

type PriorityQueueConfigImpl struct {
	QueueConfigImpl
}

func (impl *PriorityQueueConfigImpl) New() *PriorityQueueConfigImpl {
	return &PriorityQueueConfigImpl{}
}

func (impl *PriorityQueueConfigImpl) Effective() {

}

type RateLimitingQueueConfigImpl struct {
	DelayingQueueConfigImpl
}

func (impl *RateLimitingQueueConfigImpl) New() *RateLimitingQueueConfigImpl {
	return &RateLimitingQueueConfigImpl{}
}

func (impl *RateLimitingQueueConfigImpl) Effective() {

}
