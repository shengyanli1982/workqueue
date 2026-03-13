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

	attempt := q.incrementAttempt(key)
	delay, retry := q.config.policy.NextDelay(value, attempt, reason)
	if !retry {
		q.resetAttempt(key)
		q.config.callback.OnRetryExhausted(value, attempt, reason)
		return ErrRetryExhausted
	}
	if delay < 0 {
		delay = 0
	}

	// 先标记处理完成，避免幂等模式下重入队失败。
	q.Done(value)

	// PutWithDelay 以毫秒为粒度，子毫秒延迟会被截断为 0。
	// 直接走 Put 可避免进入延迟搬运路径的额外轮询开销。
	if delay < time.Millisecond {
		err = q.Put(value)
	} else {
		err = q.PutWithDelay(value, delay.Milliseconds())
	}

	if err != nil {
		return err
	}

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

func (q *retryQueueImpl) incrementAttempt(key string) int {
	q.lock.Lock()
	q.attempts[key]++
	attempt := q.attempts[key]
	q.lock.Unlock()
	return attempt
}

func (q *retryQueueImpl) resetAttempt(key string) {
	q.lock.Lock()
	delete(q.attempts, key)
	q.lock.Unlock()
}
