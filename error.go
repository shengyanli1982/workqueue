package workqueue

import "errors"

var (
	// ErrorQueueClosed 队列已关闭
	// ErrorQueueClosed Queue is closed
	ErrorQueueClosed = errors.New("[workqueue] Queue: Is closed")

	// ErrorQueueEmpty 队列为空
	// ErrorQueueEmpty Queue is empty
	ErrorQueueEmpty = errors.New("[workqueue] Queue: Queue is empty")

	// ErrorQueueElementExist 元素已存在
	// ErrorQueueElementExist Element already exists
	ErrorQueueElementExist = errors.New("[workqueue] Queue: Element already exists in queue")
)
