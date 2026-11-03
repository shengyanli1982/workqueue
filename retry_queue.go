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

	// lock 保护 attempts 表。短临界区下独占 Mutex 优于 RWMutex。
	lock     sync.Mutex
	attempts map[string]int

	// deadLetterID 为转发死信生成 base-36 单调 ID。
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

// Retry 按重试策略将 value 重新入队：锁内预占计数后由 policy.NextDelay 决策延迟；
// retry=false 时走耗尽分支（触发 OnRetryExhausted 与死信转发）返回 ErrRetryExhausted；
// 入队失败回滚预占计数。
//
// Done 契约：成功（返回 nil）时本方法已在内部完成 Done(value)，调用方不得再次调用。
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

	// 单一锁段内完成"读取 + 预占自增"。Put 失败时经 rollbackAttempt 归还。
	q.lock.Lock()
	attempt := q.attempts[key] + 1
	q.attempts[key] = attempt
	q.lock.Unlock()

	// policy.NextDelay 是用户代码，必须在锁外调用。
	delay, retry := q.config.policy.NextDelay(value, attempt, reason)
	if !retry {
		// 耗尽分支 claim-once：锁内仅首个成功删除的调用方触发回调与死信转发。
		q.lock.Lock()
		_, fire := q.attempts[key]
		if fire {
			delete(q.attempts, key)
		}
		q.lock.Unlock()

		if fire {
			q.config.callback.OnRetryExhausted(value, attempt, reason)
			q.forwardToDeadLetter(value, attempt, reason)
		}
		return ErrRetryExhausted
	}
	if delay < 0 {
		delay = 0
	}

	immediate := delay < time.Millisecond

	// 子毫秒延迟截断为 0 时直接走 Put，避免延迟搬运路径的额外开销。
	if immediate {
		err = q.Put(value)
	} else {
		err = q.PutWithDelay(value, delay.Milliseconds())
	}

	if err != nil {
		// 入队失败回滚预占计数，避免退避预算虚耗。
		q.rollbackAttempt(key)
		return err
	}

	q.Done(value)

	q.config.callback.OnRetry(value, attempt, delay, reason)
	return nil
}

// forwardToDeadLetter 将重试耗尽的 value 转发至配置的死信队列；未配置时为 no-op。
// PutDead 失败时静默跳过，耗尽事件已由 OnRetryExhausted 报告。
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
	letter := &DeadLetter{
		ID:          strconv.FormatUint(seq, 36),
		Payload:     value,
		SourceQueue: q.config.deadLetterSource,
		Attempts:    attempt,
		LastError:   lastError,
		FailedAt:    time.Now(),
		Meta:        nil,
	}

	_ = dlq.PutDead(letter)
}

// Forget 重置元素的重试计数，触发 OnForget 回调。nil 值静默跳过。
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

// NumRequeues 返回元素的当前重试次数。nil 值或 keyFunc 出错时返回 0。
func (q *retryQueueImpl) NumRequeues(value any) int {
	if value == nil {
		return 0
	}

	key, err := q.keyOf(value)
	if err != nil {
		return 0
	}

	q.lock.Lock()
	attempt := q.attempts[key]
	q.lock.Unlock()
	return attempt
}

// RequeueCounts 返回当前重试计数表的快照副本，可安全修改。
//
// 返回 map 的 key 是 RetryKeyFunc 计算的重试 key（非原始 value），
// 默认 key 函数产出 fmt.Sprintf("%T:%#v", value, value) 形式的字符串。
func (q *retryQueueImpl) RequeueCounts() map[string]int {
	q.lock.Lock()
	counts := make(map[string]int, len(q.attempts))
	for key, attempt := range q.attempts {
		counts[key] = attempt
	}
	q.lock.Unlock()
	return counts
}

// GetWithContext 委托内层 DelayingQueue，阻塞直到有值、ctx 完成或队列关闭。
func (q *retryQueueImpl) GetWithContext(ctx context.Context) (any, error) {
	return q.DelayingQueue.(BlockingGetQueue).GetWithContext(ctx)
}

// Shutdown 立即关停重试队列并重置重试计数表。
func (q *retryQueueImpl) Shutdown() {
	q.DelayingQueue.Shutdown()

	q.lock.Lock()
	q.attempts = make(map[string]int)
	q.lock.Unlock()
}

// ShutdownWithDrain 委托内层 DelayingQueue 完成 drain 后重置重试计数表。
// drain 期间 Retry 与 Put 同样被拒绝并返回 ErrQueueIsClosed。
func (q *retryQueueImpl) ShutdownWithDrain(ctx context.Context) error {
	err := q.DelayingQueue.(DrainableQueue).ShutdownWithDrain(ctx)

	q.lock.Lock()
	q.attempts = make(map[string]int)
	q.lock.Unlock()

	return err
}

// IsDraining 返回队列是否处于排空阶段，以内层 DelayingQueue 标记为准。
// 三态语义见 DrainingQueue 可选接口。
func (q *retryQueueImpl) IsDraining() bool {
	return q.DelayingQueue.(DrainingQueue).IsDraining()
}

// keyOf 通过配置的 keyFunc 计算元素的重试 key：空字符串返回 ErrRetryKeyEmpty。
func (q *retryQueueImpl) keyOf(value any) (string, error) {
	key := q.config.keyFunc(value)
	if key == "" {
		return "", ErrRetryKeyEmpty
	}
	return key, nil
}

// rollbackAttempt 回滚一次预占的 attempt 计数：锁内减 1，≤0 时删除条目。
// 与 Retry 的 Put 失败路径配对，消除计数漂移。
// 已知近似：跨周期窗口回滚可能减掉新周期计数，漂移有界 ≤1。
func (q *retryQueueImpl) rollbackAttempt(key string) {
	q.lock.Lock()
	if next := q.attempts[key] - 1; next > 0 {
		q.attempts[key] = next
	} else {
		delete(q.attempts, key)
	}
	q.lock.Unlock()
}

// resetAttempt 重置指定 key 的 attempt 计数。
func (q *retryQueueImpl) resetAttempt(key string) {
	q.lock.Lock()
	delete(q.attempts, key)
	q.lock.Unlock()
}
