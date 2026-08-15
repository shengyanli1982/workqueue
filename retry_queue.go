package workqueue

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// retryQueueImpl 组合 DelayingQueue 实现失败重试能力。
type retryQueueImpl struct {
	DelayingQueue
	config *RetryQueueConfig

	lock     sync.RWMutex
	attempts map[string]int

	// deadLetterID 为转发死信生成 base-36 单调 ID，
	// 与 LeasedQueue leaseID（leased_queue.go）的生成模式一致。
	deadLetterID atomic.Uint64
}

// NewRetryQueue 创建重试队列。
func NewRetryQueue(config *RetryQueueConfig) RetryQueue {
	config = isRetryQueueConfigEffective(config)

	return &retryQueueImpl{
		DelayingQueue: NewDelayingQueue(&config.DelayingQueueConfig),
		config:        config,
		attempts:      make(map[string]int),
	}
}

func (q *retryQueueImpl) Retry(value any, reason error) error {
	if q.IsClosed() {
		return ErrQueueIsClosed
	}
	if value == nil {
		return ErrElementIsNil
	}

	key, err := q.keyOf(value)
	if err != nil {
		return err
	}

	// Phase 1: 预读当前 attempt，预算策略决策。
	// 读锁开销极低，可快速得到 tentativeAttempt 用于计算 delay。
	q.lock.RLock()
	tentativeAttempt := q.attempts[key] + 1
	q.lock.RUnlock()

	delay, retry := q.config.policy.NextDelay(value, tentativeAttempt, reason)
	if !retry {
		// 耗尽分支：自增取得耗尽计数后立即重置，同值此后可重启完整重试周期。
		attempt := q.incrementAttempt(key)
		q.resetAttempt(key)
		q.config.callback.OnRetryExhausted(value, attempt, reason)
		q.forwardToDeadLetter(value, attempt, reason)
		return ErrRetryExhausted
	}
	if delay < 0 {
		delay = 0
	}

	immediate := delay < time.Millisecond

	// Phase 2: 写锁完成 attempt 累积。attempts 必须跨多次 Retry 累积——
	// 它是策略决策（Phase 1 的 tentativeAttempt）与 NumRequeues 的唯一数据源；
	// immediate 与延迟路径在此无差别，重置仅发生在耗尽分支或 Forget。
	attempt := q.incrementAttempt(key)

	// PutWithDelay 以毫秒为粒度，子毫秒延迟会被截断为 0。
	// 直接走 Put 可避免进入延迟搬运路径的额外轮询开销。
	// 先入队再标记完成，避免入队失败时元素永久丢失。
	if immediate {
		err = q.Put(value)
	} else {
		err = q.PutWithDelay(value, delay.Milliseconds())
	}

	if err != nil {
		return err
	}

	q.Done(value)

	q.config.callback.OnRetry(value, attempt, delay, reason)
	return nil
}

// forwardToDeadLetter 将重试耗尽的 value 转发至配置的死信队列；未配置时为
// no-op（耗尽行为保持不变）。PutDead 失败（如死信队列已关停）时静默跳过：
// 耗尽事件已由先触发的 OnRetryExhausted 报告，且 Retry 仍返回
// ErrRetryExhausted，详见 WithDeadLetterQueue 的文档注释。
func (q *retryQueueImpl) forwardToDeadLetter(value any, attempt int, reason error) {
	dlq := q.config.deadLetterQueue
	if dlq == nil {
		return
	}

	lastError := ""
	if reason != nil {
		lastError = reason.Error()
	}

	seq := q.deadLetterID.Add(1)
	var raw [16]byte
	letter := &DeadLetter{
		ID:          string(strconv.AppendUint(raw[:0], seq, 36)),
		Payload:     value,
		SourceQueue: q.config.deadLetterSource,
		Attempts:    attempt,
		LastError:   lastError,
		FailedAt:    time.Now(),
		Meta:        nil,
	}

	_ = dlq.PutDead(letter)
}

func (q *retryQueueImpl) Forget(value any) {
	if value == nil {
		return
	}

	key, err := q.keyOf(value)
	if err != nil {
		return
	}

	q.resetAttempt(key)
	q.config.callback.OnForget(value)
}

func (q *retryQueueImpl) NumRequeues(value any) int {
	if value == nil {
		return 0
	}

	key, err := q.keyOf(value)
	if err != nil {
		return 0
	}

	q.lock.RLock()
	attempt := q.attempts[key]
	q.lock.RUnlock()
	return attempt
}

// RequeueCounts 返回当前重试计数表（attempts）的快照副本，锁内拷贝、锁外返回；
// 副本可安全修改，不影响内部状态。
//
// 返回 map 的 key **不是**入队的原始 value，而是 RetryKeyFunc（可经 WithKeyFunc
// 替换）对 value 计算出的重试 key 字符串；默认 key 函数产出
// fmt.Sprintf("%T:%#v", value, value) 形式的字符串。使用自定义 key 函数时，
// 返回的 key 跟随其输出，做结果关联需按同一 key 函数换算。
func (q *retryQueueImpl) RequeueCounts() map[string]int {
	q.lock.RLock()
	counts := make(map[string]int, len(q.attempts))
	for key, attempt := range q.attempts {
		counts[key] = attempt
	}
	q.lock.RUnlock()
	return counts
}

// GetWithContext 委托内层 DelayingQueue：重试延迟到期项由 scheduler 搬运，
// 等待内层即正确语义（阻塞直到有值 / ctx 完成 / 队列关闭，永不返回
// ErrQueueIsEmpty）。
func (q *retryQueueImpl) GetWithContext(ctx context.Context) (any, error) {
	return q.DelayingQueue.(BlockingGetQueue).GetWithContext(ctx)
}

func (q *retryQueueImpl) Shutdown() {
	q.DelayingQueue.Shutdown()

	q.lock.Lock()
	q.attempts = make(map[string]int)
	q.lock.Unlock()
}

// ShutdownWithDrain 委托内层 DelayingQueue 完成 drain 后重置重试计数表。
// 契约：drain 期间 Retry（含 in-flight 项的重入队）与 Put 同样被拒绝并返回
// ErrQueueIsClosed，调用方应 Done 释放处理中追踪（与 k8s workqueue 在
// ShutDownWithDrain 期间拒绝 Add 的语义一致）。
func (q *retryQueueImpl) ShutdownWithDrain(ctx context.Context) error {
	err := q.DelayingQueue.(DrainableQueue).ShutdownWithDrain(ctx)

	q.lock.Lock()
	q.attempts = make(map[string]int)
	q.lock.Unlock()

	return err
}

func (q *retryQueueImpl) keyOf(value any) (string, error) {
	key := q.config.keyFunc(value)
	if key == "" {
		return "", ErrRetryKeyEmpty
	}
	return key, nil
}

// incrementAttempt 递增指定 key 的 attempt 计数并返回递增后的值。
// 只累积不重置：重置仅由耗尽分支或 Forget 触发，保证零延迟重试策略下
// attempts 也能正常累积并耗尽。
func (q *retryQueueImpl) incrementAttempt(key string) int {
	q.lock.Lock()
	q.attempts[key]++
	attempt := q.attempts[key]
	q.lock.Unlock()
	return attempt
}

// resetAttempt 重置指定 key 的 attempt 计数。
// 使用 RLock 预判 key 是否存在，不存在时直接返回，避免获取写锁。
// 这使得对从未重试（或已重置）的 value 调用 Forget 成为廉价操作。
func (q *retryQueueImpl) resetAttempt(key string) {
	q.lock.RLock()
	_, exists := q.attempts[key]
	q.lock.RUnlock()
	if !exists {
		return
	}
	q.lock.Lock()
	delete(q.attempts, key)
	q.lock.Unlock()
}
