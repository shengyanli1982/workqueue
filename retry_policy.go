package workqueue

import "time"

// nopRetryPolicyImpl 是无操作重试策略，NextDelay 始终返回不重试。
type nopRetryPolicyImpl struct{}

// NextDelay 始终返回 (0, false)，表示不执行重试。
func (p *nopRetryPolicyImpl) NextDelay(any, int, error) (time.Duration, bool) {
	return 0, false
}

// NewNopRetryPolicyImpl 返回始终不重试的策略。
func NewNopRetryPolicyImpl() RetryPolicy { return &nopRetryPolicyImpl{} }

// exponentialRetryPolicyImpl 使用指数退避策略的重试策略，
// 等待时长按 baseDelay×2^(attempt-1) 计算，以 maxDelay 为上限，
// 超过 maxRetries 次后拒绝重试（maxRetries < 0 表示不限制）。
type exponentialRetryPolicyImpl struct {
	baseDelay  time.Duration
	maxDelay   time.Duration
	maxRetries int
}

// NextDelay 根据当前重试次数 attempt 计算下一次重试的等待时长。
// attempt 不超过 maxRetries 时返回 (delay, true)，否则返回 (0, false) 终止重试。
// 内部通过循环翻倍避免溢出，封顶逻辑保证不超过 maxDelay。
func (p *exponentialRetryPolicyImpl) NextDelay(_ any, attempt int, _ error) (time.Duration, bool) {
	if attempt <= 0 {
		attempt = 1
	}
	if p.maxRetries >= 0 && attempt > p.maxRetries {
		return 0, false
	}

	delay := p.baseDelay
	for i := 1; i < attempt; i++ {
		if delay >= p.maxDelay {
			return p.maxDelay, true
		}
		if delay > p.maxDelay/2 {
			delay = p.maxDelay
			break
		}
		delay *= 2
	}

	if delay > p.maxDelay {
		delay = p.maxDelay
	}
	return delay, true
}

// NewExponentialRetryPolicy 使用指数退避策略创建重试策略。
// maxRetries 小于 0 表示不限制最大重试次数。
func NewExponentialRetryPolicy(baseDelay, maxDelay time.Duration, maxRetries int) RetryPolicy {
	if baseDelay <= 0 {
		baseDelay = 100 * time.Millisecond
	}
	if maxDelay <= 0 {
		maxDelay = 30 * time.Second
	}
	if maxDelay < baseDelay {
		maxDelay = baseDelay
	}

	return &exponentialRetryPolicyImpl{
		baseDelay:  baseDelay,
		maxDelay:   maxDelay,
		maxRetries: maxRetries,
	}
}
