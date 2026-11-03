package workqueue

import "time"

// nopRetryPolicyImpl 是无操作重试策略，始终返回不重试。
type nopRetryPolicyImpl struct{}

// NextDelay 始终返回 (0, false)，不执行重试。
func (p *nopRetryPolicyImpl) NextDelay(any, int, error) (time.Duration, bool) {
	return 0, false
}

// NewNopRetryPolicyImpl 返回始终不重试的策略。
func NewNopRetryPolicyImpl() RetryPolicy { return &nopRetryPolicyImpl{} }

// exponentialRetryPolicyImpl 使用指数退避策略：
// 等待时长 baseDelay×2^(attempt-1)，以 maxDelay 为上限，
// 超过 maxRetries 后拒绝重试（maxRetries < 0 不限制）。
// jitterRatio 为封顶后的向下扰动比例（0 无 jitter，见 applyJitter）。
type exponentialRetryPolicyImpl struct {
	baseDelay   time.Duration
	maxDelay    time.Duration
	maxRetries  int
	jitterRatio float64
}

// NextDelay 计算下一次重试的等待时长。
// attempt 不超过 maxRetries 时返回 (delay, true)，否则 (0, false) 终止重试。
// 循环翻倍避免溢出，封顶后施加 jitter。
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
			return applyJitter(p.maxDelay, p.jitterRatio), true
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
	return applyJitter(delay, p.jitterRatio), true
}

// NewExponentialRetryPolicy 创建指数退避重试策略（无 jitter）。
// maxRetries < 0 不限制最大重试次数。
func NewExponentialRetryPolicy(baseDelay, maxDelay time.Duration, maxRetries int) RetryPolicy {
	return NewExponentialRetryPolicyWithJitter(baseDelay, maxDelay, maxRetries, 0)
}

// NewExponentialRetryPolicyWithJitter 创建带抖动的指数退避重试策略：
// 封顶后向下扰动，均匀分布于 [delay×(1-ratio), delay]，永不超过封顶值。
// ratio <= 0 无 jitter，ratio > 1 按 1 处理；maxRetries < 0 不限制。
// 并发安全（策略本身无可变状态）。
func NewExponentialRetryPolicyWithJitter(baseDelay, maxDelay time.Duration, maxRetries int, ratio float64) RetryPolicy {
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
		baseDelay:   baseDelay,
		maxDelay:    maxDelay,
		maxRetries:  maxRetries,
		jitterRatio: ratio,
	}
}
