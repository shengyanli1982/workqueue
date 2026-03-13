package workqueue

// ratelimitingQueueImpl 组合 DelayingQueue 实现限流入队。
type ratelimitingQueueImpl struct {
	DelayingQueue
	config *RateLimitingQueueConfig
}

// NewRateLimitingQueue 创建限流队列。
func NewRateLimitingQueue(config *RateLimitingQueueConfig) RateLimitingQueue {

	config = isRateLimitingQueueConfigEffective(config)

	q := &ratelimitingQueueImpl{
		config:        config,
		DelayingQueue: NewDelayingQueue(&config.DelayingQueueConfig),
	}
	return q
}

func (q *ratelimitingQueueImpl) Shutdown() {
	q.DelayingQueue.Shutdown()
}

func (q *ratelimitingQueueImpl) PutWithLimited(value interface{}) error {

	if q.IsClosed() || value == nil {
		if q.IsClosed() {
			return ErrQueueIsClosed
		}
		return ErrElementIsNil
	}

	delay := q.config.limiter.When(value).Milliseconds()

	// 有等待时间时转为延迟入队，否则立即入队。
	var err error
	if delay > 0 {
		err = q.PutWithDelay(value, delay)
	} else {
		err = q.Put(value)
	}

	q.config.callback.OnLimited(value)

	return err
}
