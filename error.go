package workqueue

import (
	"errors"
)

// ErrQueueIsClosed 是一个错误，表示队列正在关闭。
// ErrQueueIsClosed is an error indicating that the queue is shutting down.
var ErrQueueIsClosed = errors.New("queue is shutting down")

// ErrQueueIsEmpty 是一个错误，表示队列为空。
// ErrQueueIsEmpty is an error indicating that the queue is empty.
var ErrQueueIsEmpty = errors.New("queue is empty")

// ErrElementIsNil 是一个错误，表示元素为 nil。
// ErrElementIsNil is an error indicating that the element is nil.
var ErrElementIsNil = errors.New("element is nil")

// ErrElementAlreadyExist 是一个错误，表示元素已经存在。
// ErrElementAlreadyExist is an error indicating that the element already exists.
var ErrElementAlreadyExist = errors.New("element already exist")
