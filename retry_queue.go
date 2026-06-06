package workqueue

import (
	"sync"
	"time"
)

// retryQueueImpl 组合 DelayingQueue 实现失败重试能力。
type retryQueueImpl struct {
	DelayingQueue
	config *RetryQueueConfig

	lock     sync.RWMutex
	attempts map[string]int
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

func (q *retryQueueImpl) Retry(value interface{}, reason error) error {
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
		attempt := q.incrementAndMaybeResetAttempt(key, true)
		q.config.callback.OnRetryExhausted(value, attempt, reason)
		return ErrRetryExhausted
	}
	if delay < 0 {
		delay = 0
	}

	immediate := delay < time.Millisecond

	// Phase 2: 单次写锁完成 map 更新（increment + 可能的 delete）。
	// immediate 路径下 increment 和 delete 合并到同一个锁区间，
	// 使后续 Forget 变为 no-op（key 已不在 map 中）。
	attempt := q.incrementAndMaybeResetAttempt(key, immediate)

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

func (q *retryQueueImpl) Forget(value interface{}) {
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

func (q *retryQueueImpl) NumRequeues(value interface{}) int {
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

func (q *retryQueueImpl) Shutdown() {
	q.DelayingQueue.Shutdown()

	q.lock.Lock()
	q.attempts = make(map[string]int)
	q.lock.Unlock()
}

func (q *retryQueueImpl) keyOf(value interface{}) (string, error) {
	key := q.config.keyFunc(value)
	if key == "" {
		return "", ErrRetryKeyEmpty
	}
	return key, nil
}

// incrementAndMaybeResetAttempt 在单次写锁内完成 attempt 递增和可选的重置。
// 当 reset=true 时（重试耗尽或 immediate 路径），increment 和 delete 合并到
// 同一个锁区间，减少锁竞争。
func (q *retryQueueImpl) incrementAndMaybeResetAttempt(key string, reset bool) int {
	q.lock.Lock()
	q.attempts[key]++
	attempt := q.attempts[key]
	if reset {
		delete(q.attempts, key)
	}
	q.lock.Unlock()
	return attempt
}

// resetAttempt 重置指定 key 的 attempt 计数。
// 使用 RLock 预判 key 是否存在，不存在时直接返回，避免获取写锁。
// 这使得 immediate retry 后的 Forget 调用变为廉价操作。
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
