package workqueue

import (
	"errors"
)

var (
	ErrQueueIsClosed      = errors.New("queue is shutting down")      // 队列已关闭或正在关闭
	ErrQueueIsEmpty       = errors.New("queue is empty")              // 无可消费元素
	ErrElementIsNil       = errors.New("element is nil")              // 入队元素为空
	ErrElementAlreadyExist = errors.New("element already exist")      // 幂等模式下重复入队
	ErrInvalidQueueCapacity = errors.New("invalid queue capacity")    // 队列容量不合法
	ErrInvalidLeaseDuration = errors.New("invalid lease duration")    // 租约时长不合法
	ErrLeaseNotFound       = errors.New("lease not found")            // 租约不存在或已失效
	ErrRetryExhausted      = errors.New("retry exhausted")            // 重试次数已达上限
	ErrRetryKeyEmpty       = errors.New("retry key is empty")         // 重试 key 生成结果为空
	ErrInvalidDeadLetter   = errors.New("invalid dead letter")        // 死信对象不合法
	ErrInvalidTargetQueue  = errors.New("invalid target queue")       // 死信重放目标队列不合法
)
