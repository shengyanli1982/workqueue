package workqueue

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNopRetryPolicyImpl_NextDelay(t *testing.T) {
	policy := NewNopRetryPolicyImpl()

	delay, retry := policy.NextDelay("task", 1, nil)
	assert.Equal(t, time.Duration(0), delay)
	assert.False(t, retry)
}

func TestExponentialRetryPolicyImpl_NextDelay(t *testing.T) {
	policy := NewExponentialRetryPolicy(100*time.Millisecond, time.Second, 3)

	delay1, retry1 := policy.NextDelay("task", 1, nil)
	assert.True(t, retry1)
	assert.Equal(t, 100*time.Millisecond, delay1)

	delay2, retry2 := policy.NextDelay("task", 2, nil)
	assert.True(t, retry2)
	assert.Equal(t, 200*time.Millisecond, delay2)

	delay3, retry3 := policy.NextDelay("task", 3, nil)
	assert.True(t, retry3)
	assert.Equal(t, 400*time.Millisecond, delay3)

	delay4, retry4 := policy.NextDelay("task", 4, nil)
	assert.False(t, retry4)
	assert.Equal(t, time.Duration(0), delay4)
}

func TestExponentialRetryPolicyImpl_DelayCap(t *testing.T) {
	policy := NewExponentialRetryPolicy(500*time.Millisecond, 750*time.Millisecond, 10)

	delay, retry := policy.NextDelay("task", 2, nil)
	assert.True(t, retry)
	assert.Equal(t, 750*time.Millisecond, delay)
}

// TestExponentialRetryPolicyWithJitter_ZeroRatioMatchesPlain 验证零抖动：
// ratio 非正时与现有构造函数逐点一致（delay 与 retry 判定，含 maxRetries 终止语义）。
func TestExponentialRetryPolicyWithJitter_ZeroRatioMatchesPlain(t *testing.T) {
	for _, ratio := range []float64{0, -0.5} {
		plain := NewExponentialRetryPolicy(100*time.Millisecond, time.Second, 3)
		withJitter := NewExponentialRetryPolicyWithJitter(100*time.Millisecond, time.Second, 3, ratio)
		for attempt := 1; attempt <= 5; attempt++ {
			wantDelay, wantRetry := plain.NextDelay("task", attempt, nil)
			gotDelay, gotRetry := withJitter.NextDelay("task", attempt, nil)
			assert.Equal(t, wantRetry, gotRetry, "ratio=%v attempt=%d", ratio, attempt)
			assert.Equal(t, wantDelay, gotDelay, "ratio=%v attempt=%d", ratio, attempt)
		}
	}
}

// TestExponentialRetryPolicyWithJitter_Range 验证 jitter 区间：
// delay 落在 [d×(1-ratio), d]（d 为封顶后的确定指数值），且多次采样观测到扰动。
func TestExponentialRetryPolicyWithJitter_Range(t *testing.T) {
	const ratio = 0.5
	base := 10 * time.Millisecond
	policy := NewExponentialRetryPolicyWithJitter(base, time.Second, -1, ratio)

	lower := time.Duration(float64(base) * (1 - ratio))
	minSeen := time.Second
	const samples = 100
	for i := 0; i < samples; i++ {
		d, retry := policy.NextDelay("task", 1, nil) // attempt=1：确定值为 base
		assert.True(t, retry, "sample %d", i)
		assert.GreaterOrEqual(t, d, lower, "sample %d", i)
		assert.LessOrEqual(t, d, base, "sample %d", i)
		if d < minSeen {
			minSeen = d
		}
	}
	assert.Less(t, minSeen, base, "jitter must be observable within %d samples", samples)
}

// TestExponentialRetryPolicyWithJitter_CapAndExhaustion 验证封顶先行与终止语义：
// 确定值封顶为 maxDelay 后 jitter 仍不超过封顶值；超过 maxRetries 返回 (0, false)
// 且不施加 jitter；attempt 非正归一为 1，与现有实现一致。
func TestExponentialRetryPolicyWithJitter_CapAndExhaustion(t *testing.T) {
	base, maxDelay, ratio := 100*time.Millisecond, 200*time.Millisecond, 0.8

	// maxRetries 不限：attempt>=2 时确定值封顶为 200ms，jitter 后落在 [40ms, 200ms]。
	unlimited := NewExponentialRetryPolicyWithJitter(base, maxDelay, -1, ratio)
	cappedLower := time.Duration(float64(maxDelay) * (1 - ratio))
	for attempt := 2; attempt <= 6; attempt++ {
		d, retry := unlimited.NextDelay("task", attempt, nil)
		assert.True(t, retry, "attempt=%d", attempt)
		assert.GreaterOrEqual(t, d, cappedLower, "attempt=%d", attempt)
		assert.LessOrEqual(t, d, maxDelay, "attempt=%d", attempt)
	}

	// attempt=1 与 attempt 非正归一：确定值为 base，jitter 后落在 [20ms, 100ms]。
	baseLower := time.Duration(float64(base) * (1 - ratio))
	for _, attempt := range []int{1, 0, -3} {
		d, retry := unlimited.NextDelay("task", attempt, nil)
		assert.True(t, retry, "attempt=%d", attempt)
		assert.GreaterOrEqual(t, d, baseLower, "attempt=%d", attempt)
		assert.LessOrEqual(t, d, base, "attempt=%d", attempt)
	}

	// 超过 maxRetries：终止语义与现有实现一致，delay 为 0 且不施加 jitter。
	limited := NewExponentialRetryPolicyWithJitter(base, maxDelay, 2, ratio)
	d, retry := limited.NextDelay("task", 3, nil)
	assert.False(t, retry)
	assert.Equal(t, time.Duration(0), d)
}

// TestExponentialRetryPolicyWithJitter_Concurrent 验证并发安全（配合 -race）：
// 并发 NextDelay 的返回值恒落在 [base×(1-ratio), maxDelay] 且始终允许重试。
func TestExponentialRetryPolicyWithJitter_Concurrent(t *testing.T) {
	base, maxDelay, ratio := time.Millisecond, 10*time.Millisecond, 0.5
	policy := NewExponentialRetryPolicyWithJitter(base, maxDelay, -1, ratio)
	lower := time.Duration(float64(base) * (1 - ratio))

	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				d, retry := policy.NextDelay("shared", i%8+1, nil)
				if !retry {
					t.Errorf("unlimited maxRetries must always allow retry")
					return
				}
				if d < lower || d > maxDelay {
					t.Errorf("jittered delay %v outside [%v, %v]", d, lower, maxDelay)
					return
				}
			}
		}()
	}
	wg.Wait()
}
