package workqueue

import "errors"

var (
	// Queue
	ErrorQueueClosed       = errors.New("[workqueue] Queue: Is closed")
	ErrorQueueFull         = errors.New("[workqueue] Queue: Is full")
	ErrorQueueEmpty        = errors.New("[workqueue] Queue: Queue is empty")
	ErrorQueueElementExist = errors.New("[workqueue] Queue: Element already exists in queue")
	ErrorQueueGetTimeout   = errors.New("[workqueue] Queue: Get element timeout")
)
