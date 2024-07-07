package workqueue

import (
	"time"
)

// DelayTo 函数用于计算延迟时间，它接受一个持续时间（毫秒），并返回当前时间加上持续时间的 Unix 毫秒时间戳。
// The DelayTo function is used to calculate the delay time. It takes a duration (in milliseconds) and returns the Unix millisecond timestamp of the current time plus the duration.
func DelayTo(duration int64) int64 {
	// 使用 time.Now().Add 方法计算延迟时间，并使用 UnixMilli 方法将结果转换为 Unix 毫秒时间戳
	// Use the time.Now().Add method to calculate the delay time, and use the UnixMilli method to convert the result to a Unix millisecond timestamp
	return time.Now().Add(time.Millisecond * time.Duration(duration)).UnixMilli()
}

// delayingQueueImpl 结构体实现了 DelayingQueue 接口，它是一个支持延迟的队列。
// The delayingQueueImpl structure implements the DelayingQueue interface, it is a queue that supports delay.
type delayingQueueImpl struct {
	// sortedQueue 是一个排序队列，它是 delayingQueueImpl 的基础结构
	// sortedQueue is a sorted queue, it is the base structure of delayingQueueImpl
	sortedQueue

	// config 是队列的配置，包括队列的大小、延迟时间等参数
	// config is the configuration of the queue, including parameters such as the size of the queue, delay time, etc.
	config *DelayingQueueConfig
}

// NewDelayingQueue 函数用于创建一个新的 DelayingQueue。
// The NewDelayingQueue function is used to create a new DelayingQueue.
func NewDelayingQueue(config *DelayingQueueConfig) DelayingQueue {
	// 检查配置是否有效，如果无效，使用默认配置
	// Check if the configuration is valid, if not, use the default configuration
	config = isDelayingQueueConfigEffective(config)

	// 创建一个新的 DelayingQueueImpl
	// Create a new DelayingQueueImpl
	return &delayingQueueImpl{
		// 设置配置
		// Set the configuration
		config: config,

		// 创建一个新的排序队列
		// Create a new sorted queue
		sortedQueue: *newSortedQueue(&config.QueueConfig),
	}
}

// PutWithDelay 方法将元素放入队列，并设置延迟时间
// The PutWithDelay method puts an element into the queue and sets a delay time
func (q *delayingQueueImpl) PutWithDelay(value interface{}, delay int64) error {
	// 将延迟时间转换为时间戳，并将元素放入队列
	// Convert the delay time to a timestamp and put the element into the queue
	return q.PutWithTimestamp(value, DelayTo(delay))
}

// PutWithTimestamp 方法将元素放入队列，并设置时间戳
// The PutWithTimestamp method puts an element into the queue and sets a timestamp
func (q *delayingQueueImpl) PutWithTimestamp(value interface{}, ts int64) error {
	// 将元素放入排序队列，并设置优先级为时间戳
	// Put the element into the sorted queue and set the priority to the timestamp
	return q.sortedQueue.putWithPriority(value, ts, q.config.callback.OnDelay, q.config.callback.OnPut)
}
