package workqueue

type RateLimitingQueueImpl struct {
	DelayingQueue
	config *RateLimitingQueueConfig
}

func NewRateLimitingQueue(config *RateLimitingQueueConfig) RateLimitingQueue {
	config = isRateLimitingQueueConfigEffective(config)

	q := &RateLimitingQueueImpl{
		config:        config,
		DelayingQueue: NewDelayingQueue(&config.DelayingQueueConfig),
	}

	return q
}

func (q *RateLimitingQueueImpl) Shutdown() {
	q.DelayingQueue.Shutdown()
}

func (q *RateLimitingQueueImpl) PutWithLimited(value interface{}) error {
	if q.IsClosed() {
		return ErrQueueIsClosed
	}

	if value == nil {
		return ErrElementIsNil
	}

	delay := q.config.limiter.When(value)

	var err error
	if delay > 0 {
		err = q.PutWithDelay(value, int64(delay))
	} else {
		err = q.Put(value)
	}

	q.config.callback.OnLimited(value)

	return err
}
