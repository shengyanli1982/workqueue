package workqueue

// ratelimitingQueueImpl implements the RateLimitingQueue interface
// ratelimitingQueueImpl 结构体实现了 RateLimitingQueue 接口，提供限速队列功能
type ratelimitingQueueImpl struct {
	DelayingQueue                          // 内嵌延迟队列 / Embedded DelayingQueue
	config        *RateLimitingQueueConfig // 限速队列配置 / Rate limiting queue configuration
}

// NewRateLimitingQueue creates a new RateLimitingQueue with the given configuration
// NewRateLimitingQueue 函数用于创建一个新的限速队列，接收配置参数
func NewRateLimitingQueue(config *RateLimitingQueueConfig) RateLimitingQueue {
	// Validate and set default configuration if needed
	// 验证配置并设置默认值（如果需要）
	config = isRateLimitingQueueConfigEffective(config)

	q := &ratelimitingQueueImpl{
		config:        config,
		DelayingQueue: NewDelayingQueue(&config.DelayingQueueConfig),
	}
	return q
}

// Shutdown stops the queue from accepting new items and shuts down internal workers
// Shutdown 方法用于关闭限速队列，停止接收新元素并关闭内部工作协程
func (q *ratelimitingQueueImpl) Shutdown() {
	q.DelayingQueue.Shutdown()
}

// PutWithLimited adds an item to the queue with rate limiting applied
// If the limiter returns a delay > 0, the item will be delayed for that duration
// PutWithLimited 方法用于将元素添加到限速队列中
// 如果限速器返回的延迟时间大于0，该元素将被延迟相应时间后再加入队列
func (q *ratelimitingQueueImpl) PutWithLimited(value interface{}) error {
	// Check for queue closure and nil value
	// 检查队列是否已关闭以及输入值是否为nil
	if q.IsClosed() || value == nil {
		if q.IsClosed() {
			return ErrQueueIsClosed
		}
		return ErrElementIsNil
	}

	// Get delay duration from rate limiter
	// 从限速器获取延迟时间
	delay := q.config.limiter.When(value).Milliseconds()

	// Add element either with delay or immediately based on limiter response
	// 根据限速器返回的延迟时间决定是直接添加还是延迟添加
	var err error
	if delay > 0 {
		err = q.PutWithDelay(value, delay)
	} else {
		err = q.Put(value)
	}

	// Notify callback of rate limiting event
	// 触发限速回调通知
	q.config.callback.OnLimited(value)

	return err
}
