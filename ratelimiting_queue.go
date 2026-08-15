package workqueue

import (
	"context"
)

// ratelimitingQueueImpl 组合 DelayingQueue 实现限流入队。
type ratelimitingQueueImpl struct {
	DelayingQueue
	config *RateLimitingQueueConfig
}

// NewRateLimitingQueue 创建限流队列。
func NewRateLimitingQueue(config *RateLimitingQueueConfig) RateLimitingQueue {

	config = isRateLimitingQueueConfigEffective(config)

	q := &ratelimitingQueueImpl{
		config:        config,
		DelayingQueue: NewDelayingQueue(&config.DelayingQueueConfig),
	}
	return q
}

func (q *ratelimitingQueueImpl) Shutdown() {
	q.DelayingQueue.Shutdown()
}

// ShutdownWithDrain 委托内层 DelayingQueue 完成优雅关停，无自身清理。
func (q *ratelimitingQueueImpl) ShutdownWithDrain(ctx context.Context) error {
	return q.DelayingQueue.(DrainableQueue).ShutdownWithDrain(ctx)
}

// GetWithContext 委托内层 DelayingQueue：限流延迟到期项由 scheduler 搬运，
// 等待内层即正确语义（阻塞直到有值 / ctx 完成 / 队列关闭，永不返回
// ErrQueueIsEmpty）。
func (q *ratelimitingQueueImpl) GetWithContext(ctx context.Context) (any, error) {
	return q.DelayingQueue.(BlockingGetQueue).GetWithContext(ctx)
}

// Forget 与限流器联动清理：limiter 实现 LimiterForgetter 时委托其清理该 item
// 的退避状态（防 per-item 表无限增长），否则为 no-op。
// RateLimitingQueue 接口保持不动，调用方经 `q.(LimiterForgetter)` 类型断言使用。
func (q *ratelimitingQueueImpl) Forget(value any) {
	if value == nil {
		return
	}

	if f, ok := q.config.limiter.(LimiterForgetter); ok {
		f.Forget(value)
	}
}

func (q *ratelimitingQueueImpl) PutWithLimited(value any) error {

	if q.IsClosed() || value == nil {
		if q.IsClosed() {
			return ErrQueueIsClosed
		}
		return ErrElementIsNil
	}

	delay := q.config.limiter.When(value).Milliseconds()

	// 有等待时间时转为延迟入队，否则立即入队。
	var err error
	if delay > 0 {
		err = q.PutWithDelay(value, delay)
	} else {
		err = q.Put(value)
	}

	if err == nil {
		q.config.callback.OnLimited(value)
	}

	return err
}
