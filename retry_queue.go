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

	// lock 保护 attempts 表。pprof 证据（P1-3）：RWMutex 的读写锁在短临界区
	// 上产生高昂的原子操作开销（atomic.Int32.Add 占采样 19.2%），而本文件
	// 全部临界区均为极短的 map 读写，独占 Mutex 反而更快；Mutex 同时使
	// “读取+预占自增”与耗尽分支的 claim-once 各自在单一锁段内原子完成。
	lock     sync.Mutex
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

	// 单一锁段内完成“读取 + 预占自增”（单次 map 查找）。attempts 必须跨多次
	// Retry 累积——它是策略决策与 NumRequeues 的唯一数据源；immediate 与延迟
	// 路径在此无差别，重置仅发生在耗尽分支或 Forget。Put 失败时经
	// rollbackAttempt 归还预占计数，消除计数漂移（#8a）。
	q.lock.Lock()
	attempt := q.attempts[key] + 1
	q.attempts[key] = attempt
	q.lock.Unlock()

	// policy.NextDelay 是用户代码，必须在锁外调用。
	delay, retry := q.config.policy.NextDelay(value, attempt, reason)
	if !retry {
		// 耗尽分支 claim-once（#8b）：锁内仅首个成功删除计数条目的调用方
		// 触发回调与死信转发，保证同值并发耗尽时副作用恰好一次；
		// 未 claim 到的调用方仍返回 ErrRetryExhausted 但不重复副作用。
		// 计数被删除后，同值此后可重启完整重试周期。
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

	// PutWithDelay 以毫秒为粒度，子毫秒延迟会被截断为 0。
	// 直接走 Put 可避免进入延迟搬运路径的额外轮询开销。
	// 先入队再标记完成，避免入队失败时元素永久丢失。
	if immediate {
		err = q.Put(value)
	} else {
		err = q.PutWithDelay(value, delay.Milliseconds())
	}

	if err != nil {
		// Put 失败回滚（#8a）：归还本次预占的计数，避免漂移导致退避预算虚耗、
		// 提前进入耗尽/死信。回滚不得破坏幂等队列“先 Put 后 Done”的挂起语义
		//（queue.go Put：processing 命中打挂起标记），仅归还本调用预占的 +1。
		q.rollbackAttempt(key)
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

	// 与 deadLetterQueueImpl.nextID 同款（P1-2）：FormatUint 直接产出 string，
	// 避免 [N]byte + AppendUint 后再 string(buf) 的两步转换，堆分配同为 1 次。
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

// Forget 重置元素的重试计数：成功处理后调用，清除累积的 attempt 并触发 OnForget 回调。
// nil 值或 keyFunc 出错时静默跳过。
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

// NumRequeues 返回元素的当前重试次数：nil 值或 keyFunc 出错时返回 0。
// 加锁访问 attempts 表，临界区仅一次 map 读取。
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

// RequeueCounts 返回当前重试计数表（attempts）的快照副本，锁内拷贝、锁外返回；
// 副本可安全修改，不影响内部状态。
//
// 返回 map 的 key **不是**入队的原始 value，而是 RetryKeyFunc（可经 WithKeyFunc
// 替换）对 value 计算出的重试 key 字符串；默认 key 函数产出
// fmt.Sprintf("%T:%#v", value, value) 形式的字符串。使用自定义 key 函数时，
// 返回的 key 跟随其输出，做结果关联需按同一 key 函数换算。
func (q *retryQueueImpl) RequeueCounts() map[string]int {
	q.lock.Lock()
	counts := make(map[string]int, len(q.attempts))
	for key, attempt := range q.attempts {
		counts[key] = attempt
	}
	q.lock.Unlock()
	return counts
}

// GetWithContext 委托内层 DelayingQueue：重试延迟到期项由 scheduler 搬运，
// 等待内层即正确语义（阻塞直到有值 / ctx 完成 / 队列关闭，永不返回
// ErrQueueIsEmpty）。
func (q *retryQueueImpl) GetWithContext(ctx context.Context) (any, error) {
	return q.DelayingQueue.(BlockingGetQueue).GetWithContext(ctx)
}

// Shutdown 立即关停重试队列：委托内层 DelayingQueue 关停后重置重试计数表，
// 释放内存并防止关停后残留计数影响后续使用。
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

// keyOf 通过配置的 keyFunc 计算元素的重试 key：空字符串返回 ErrRetryKeyEmpty。
func (q *retryQueueImpl) keyOf(value any) (string, error) {
	key := q.config.keyFunc(value)
	if key == "" {
		return "", ErrRetryKeyEmpty
	}
	return key, nil
}

// rollbackAttempt 回滚一次预占的 attempt 计数：锁内减 1，≤0 时删除条目。
// 与 Retry 的 Put 失败路径配对，消除计数漂移（#8a）。
// 只累积不重置的语义保持不变：重置仅由耗尽分支或 Forget 触发。
// 已知近似——存在跨周期窗口：若他方耗尽分支 claim-once 删除计数后、第三方重启
// 新周期预占，本次回滚可能减掉的是新周期的计数。漂移有界 ≤1，属可接受的近似；
// 完整修复需为计数附加周期标签以区分新旧周期，属过度设计，故不采用。
func (q *retryQueueImpl) rollbackAttempt(key string) {
	q.lock.Lock()
	if next := q.attempts[key] - 1; next > 0 {
		q.attempts[key] = next
	} else {
		delete(q.attempts, key)
	}
	q.lock.Unlock()
}

// resetAttempt 重置指定 key 的 attempt 计数（delete 对不存在 key 为 no-op，
// 对从未重试或已重置的 value 调用 Forget 因此是廉价的单锁段操作）。
func (q *retryQueueImpl) resetAttempt(key string) {
	q.lock.Lock()
	delete(q.attempts, key)
	q.lock.Unlock()
}
