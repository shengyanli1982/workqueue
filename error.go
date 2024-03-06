package workqueue

import "errors"

var (
	// ErrorQueueClosed 是一个错误类型，表示队列已关闭
	// ErrorQueueClosed is an error type indicating that the queue is closed
	ErrorQueueClosed = errors.New("[workqueue] Queue: Is closed")

	// ErrorQueueEmpty 是一个错误类型，表示队列为空
	// ErrorQueueEmpty is an error type indicating that the queue is empty
	ErrorQueueEmpty = errors.New("[workqueue] Queue: Queue is empty")

	// ErrorQueueElementExist 是一个错误类型，表示元素已存在于队列中
	// ErrorQueueElementExist is an error type indicating that the element already exists in the queue
	ErrorQueueElementExist = errors.New("[workqueue] Queue: Element already exists in queue")
)
