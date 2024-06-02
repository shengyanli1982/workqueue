package workqueue

// ratelimitingQueueImpl 结构体，实现了 RateLimitingQueue 接口
// The ratelimitingQueueImpl struct, which implements the RateLimitingQueue interface
type ratelimitingQueueImpl struct {
	// DelayingQueue 是一个延迟队列接口
	// DelayingQueue is a delay queue interface
	DelayingQueue

	// config 是 RateLimitingQueue 的配置
	// config is the configuration of RateLimitingQueue
	config *RateLimitingQueueConfig
}

// NewRateLimitingQueue 函数用于创建一个新的 RateLimitingQueue
// The NewRateLimitingQueue function is used to create a new RateLimitingQueue
func NewRateLimitingQueue(config *RateLimitingQueueConfig) RateLimitingQueue {
	// 检查配置是否有效，如果无效，使用默认配置
	// Check if the configuration is valid, if not, use the default configuration
	config = isRateLimitingQueueConfigEffective(config)

	// 创建一个新的 RateLimitingQueueImpl
	// Create a new RateLimitingQueueImpl
	q := &ratelimitingQueueImpl{
		// 设置配置
		// Set the configuration
		config: config,

		// 创建一个新的 DelayingQueue，并将其赋值给 q.DelayingQueue
		// Create a new DelayingQueue and assign it to q.DelayingQueue
		DelayingQueue: NewDelayingQueue(&config.DelayingQueueConfig),
	}

	// 返回新创建的 RateLimitingQueue
	// Return the newly created RateLimitingQueue
	return q
}

// Shutdown 方法用于关闭 RateLimitingQueue
// The Shutdown method is used to shut down the RateLimitingQueue
func (q *ratelimitingQueueImpl) Shutdown() { q.DelayingQueue.Shutdown() }

// PutWithLimited 方法用于将一个元素放入 RateLimitingQueue，元素的延迟时间由限流器决定。
// The PutWithLimited method is used to put an element into the RateLimitingQueue. The delay time of the element is determined by the limiter.
func (q *ratelimitingQueueImpl) PutWithLimited(value interface{}) error {
	// 如果 RateLimitingQueue 已关闭，返回错误
	// If the RateLimitingQueue is closed, return an error
	if q.IsClosed() {
		return ErrQueueIsClosed
	}

	// 如果元素值为 nil，返回错误
	// If the element value is nil, return an error
	if value == nil {
		return ErrElementIsNil
	}

	// 通过限流器获取元素的延迟时间
	// Get the delay time of the element through the limiter
	delay := q.config.limiter.When(value).Milliseconds()

	// 定义错误变量
	// Define the error variable
	var err error

	// 如果延迟时间大于 0
	// If the delay time is greater than 0
	if delay > 0 {
		// 调用 PutWithDelay 方法，将元素放入 RateLimitingQueue，并设置其延迟时间
		// Call the PutWithDelay method to put the element into the RateLimitingQueue and set its delay time
		err = q.PutWithDelay(value, int64(delay))
	} else {
		// 否则，直接将元素放入 RateLimitingQueue
		// Otherwise, put the element directly into the RateLimitingQueue
		err = q.Put(value)
	}

	// 调用回调函数，通知元素已被限流
	// Call the callback function to notify that the element has been limited
	q.config.callback.OnLimited(value)

	// 返回错误
	// Return the error
	return err
}
