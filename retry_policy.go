package workqueue

import "time"

type nopRetryPolicyImpl struct{}

func (p *nopRetryPolicyImpl) NextDelay(interface{}, int, error) (time.Duration, bool) {
	return 0, false
}

// NewNopRetryPolicyImpl 返回始终不重试的策略。
func NewNopRetryPolicyImpl() RetryPolicy { return &nopRetryPolicyImpl{} }

type exponentialRetryPolicyImpl struct {
	baseDelay  time.Duration
	maxDelay   time.Duration
	maxRetries int
}

func (p *exponentialRetryPolicyImpl) NextDelay(_ interface{}, attempt int, _ error) (time.Duration, bool) {
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
