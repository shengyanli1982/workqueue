package workqueue

import (
	"context"
)

// ratelimitingQueueImpl 通过组合 DelayingQueue 实现限流入队：
// Limiter 返回的等待时长大于零时委托 PutWithDelay，否则直接入队。
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

// IsDraining 返回队列是否处于排空阶段，以内层 DelayingQueue 标记为准。
// 三态语义见 DrainingQueue 可选接口。
func (q *ratelimitingQueueImpl) IsDraining() bool {
	return q.DelayingQueue.(DrainingQueue).IsDraining()
}

// GetWithContext 委托内层 DelayingQueue，阻塞直到有值、ctx 完成或队列关闭。
func (q *ratelimitingQueueImpl) GetWithContext(ctx context.Context) (any, error) {
	return q.DelayingQueue.(BlockingGetQueue).GetWithContext(ctx)
}

// Forget 与限流器联动清理：limiter 实现 LimiterForgetter 时委托清理退避状态。
// 调用方经 q.(LimiterForgetter) 类型断言使用。
func (q *ratelimitingQueueImpl) Forget(value any) {
	if value == nil {
		return
	}

	if f, ok := q.config.limiter.(LimiterForgetter); ok {
		f.Forget(value)
	}
}

// PutWithLimited 执行限流入队：通过 Limiter 计算等待时长，
// 大于零时转为延迟入队，否则直接入队。入队成功后触发 OnLimited 回调。
//
// Limiter 状态契约：When 已推进限流器内部状态，后续入队失败不回退 limiter。
// drain 拒绝发生在 When 之后，限流预算已消耗；关停与 nil 值拒绝在 When 之前，
// 不消耗预算。
func (q *ratelimitingQueueImpl) PutWithLimited(value any) error {

	if q.IsClosed() || value == nil {
		if q.IsClosed() {
			return ErrQueueIsClosed
		}
		return ErrElementIsNil
	}

	when := q.config.limiter.When(value)

	// PutWithDelay 以毫秒为粒度：亚毫秒等待截断为 0 时限流失效，
	// 向上取整至 1ms 以保持限流语义。
	var err error
	if when > 0 {
		delay := when.Milliseconds()
		if delay == 0 {
			delay = 1
		}
		err = q.PutWithDelay(value, delay)
	} else {
		err = q.Put(value)
	}

	if err == nil {
		q.config.callback.OnLimited(value)
	}

	return err
}
