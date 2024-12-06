package workqueue

// ratelimitingQueueImpl 结构体，实现了 RateLimitingQueue 接口
type ratelimitingQueueImpl struct {
	DelayingQueue
	config *RateLimitingQueueConfig
}

// NewRateLimitingQueue 函数用于创建一个新的 RateLimitingQueue
func NewRateLimitingQueue(config *RateLimitingQueueConfig) RateLimitingQueue {
	config = isRateLimitingQueueConfigEffective(config)

	q := &ratelimitingQueueImpl{
		config:        config,
		DelayingQueue: NewDelayingQueue(&config.DelayingQueueConfig),
	}
	return q
}

// Shutdown 方法用于关闭 RateLimitingQueue
func (q *ratelimitingQueueImpl) Shutdown() {
	q.DelayingQueue.Shutdown()
}

// PutWithLimited 方法用于将一个元素放入 RateLimitingQueue
func (q *ratelimitingQueueImpl) PutWithLimited(value interface{}) error {
	// 合并错误检查逻辑
	if q.IsClosed() || value == nil {
		if q.IsClosed() {
			return ErrQueueIsClosed
		}
		return ErrElementIsNil
	}

	// 获取延迟时间
	delay := q.config.limiter.When(value).Milliseconds()

	// 根据延迟时间决定使用哪种方式添加元素
	var err error
	if delay > 0 {
		err = q.PutWithDelay(value, delay)
	} else {
		err = q.Put(value)
	}

	// 回调通知
	q.config.callback.OnLimited(value)

	return err
}
