// Package workqueue 提供生产级工作队列库，涵盖从简单 FIFO 到复杂调度场景的完整队列族。
//
// 队列类型：
//
//   - Queue：基础 FIFO 队列，支持幂等入队与阻塞消费
//   - DelayingQueue：延迟队列，元素按指定延迟时间到期后入队
//   - PriorityQueue：优先级队列，按优先级高低排序消费
//   - RateLimitingQueue：限流队列，基于 Limiter 策略控制入队速率
//   - RetryQueue：重试队列，失败元素按 RetryPolicy 策略自动重试
//   - DeadLetterQueue：死信队列，收集重试耗尽或不可恢复的失败元素
//   - LeasedQueue：租约队列，消费方获取带租约的元素，超时未确认自动回队
//   - BoundedBlockingQueue：有界阻塞队列，固定容量、满时阻塞生产者、空时阻塞消费者
//   - TimerQueue：定时队列，按绝对时间点或相对时长调度入队
//
// 核心设计理念：
//
//   - 幂等语义：开启 WithValueIdempotent 后，相同值不会被重复入队
//   - 优雅关停：所有队列均支持 Shutdown 立即关停与 ShutdownWithDrain 等待在途项处理完毕
//   - 阻塞消费：通过 GetWithContext 或 BoundedBlockingQueue 实现生产者-消费者模型的自然阻塞
//   - 回调驱动：每种队列类型提供对应的 Callback 接口，覆盖全生命周期事件
//
// 快速示例：
//
//	q := workqueue.NewQueue(workqueue.NewQueueConfig().WithValueIdempotent())
//	defer q.Shutdown()
//
//	_ = q.Put("task-1")
//	value, err := q.Get()
//	if err == nil {
//	    // 处理 value
//	    q.Done(value)
//	}
package workqueue
