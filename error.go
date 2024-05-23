package workqueue

import (
	"errors"
)

var ErrQueueIsClosed = errors.New("queue is shutting down")
var ErrQueueIsEmpty = errors.New("queue is empty")
