package workqueue

import (
	"errors"
)

// ErrQueueIsClosed 表示队列已关闭或正在关闭。
var ErrQueueIsClosed = errors.New("queue is shutting down")

// ErrQueueIsEmpty 表示当前没有可消费元素。
var ErrQueueIsEmpty = errors.New("queue is empty")

// ErrElementIsNil 表示入队元素为空。
var ErrElementIsNil = errors.New("element is nil")

// ErrElementAlreadyExist 表示幂等模式下重复入队。
var ErrElementAlreadyExist = errors.New("element already exist")

// ErrInvalidQueueCapacity 表示队列容量配置不合法。
var ErrInvalidQueueCapacity = errors.New("invalid queue capacity")

// ErrInvalidLeaseDuration 表示租约时长不合法。
var ErrInvalidLeaseDuration = errors.New("invalid lease duration")

// ErrLeaseNotFound 表示租约不存在或已失效。
var ErrLeaseNotFound = errors.New("lease not found")

// ErrRetryExhausted 表示元素重试次数已达上限。
var ErrRetryExhausted = errors.New("retry exhausted")

// ErrRetryKeyEmpty 表示重试 key 生成结果为空。
var ErrRetryKeyEmpty = errors.New("retry key is empty")

// ErrInvalidDeadLetter 表示死信对象不合法。
var ErrInvalidDeadLetter = errors.New("invalid dead letter")

// ErrInvalidTargetQueue 表示死信重放目标队列不合法。
var ErrInvalidTargetQueue = errors.New("invalid target queue")
