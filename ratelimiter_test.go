package workqueue

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestItemExponentialFailureRateLimiter_BackoffSequence per-item 指数退避序列与封顶。
func TestItemExponentialFailureRateLimiter_BackoffSequence(t *testing.T) {
	limiter := NewItemExponentialFailureRateLimiter(time.Millisecond, 100*time.Millisecond)

	want := []time.Duration{
		1 * time.Millisecond,
		2 * time.Millisecond,
		4 * time.Millisecond,
		8 * time.Millisecond,
		16 * time.Millisecond,
		32 * time.Millisecond,
		64 * time.Millisecond,
		100 * time.Millisecond, // 128ms 封顶为 100ms
		100 * time.Millisecond, // 继续封顶
	}
	for i, w := range want {
		assert.Equal(t, w, limiter.When("item-a"), "backoff sequence at call %d", i+1)
	}

	// 不同 item 计数独立。
	assert.Equal(t, time.Millisecond, limiter.When("item-b"))
	assert.Equal(t, 2*time.Millisecond, limiter.When("item-b"))
}

// TestItemExponentialFailureRateLimiter_Forget Forget 清理退避状态，防内存泄漏。
func TestItemExponentialFailureRateLimiter_Forget(t *testing.T) {
	limiter := NewItemExponentialFailureRateLimiter(time.Millisecond, time.Second)

	assert.Equal(t, time.Millisecond, limiter.When("k1"))
	assert.Equal(t, 2*time.Millisecond, limiter.When("k1"))
	assert.Equal(t, time.Millisecond, limiter.When("k2"))

	failures := limiter.(*itemExponentialFailureRateLimiter).failures
	assert.Len(t, failures, 2)

	forgetter, ok := limiter.(LimiterForgetter)
	if !ok {
		t.Fatal("item exponential limiter must support LimiterForgetter")
	}

	// 泄漏防护：Forget 后条目即被删除，map 长度回落。
	forgetter.Forget("k1")
	assert.Len(t, failures, 1)

	// k1 归零；k2 不受影响。
	assert.Equal(t, time.Millisecond, limiter.When("k1"))
	assert.Equal(t, 2*time.Millisecond, limiter.When("k2"))

	forgetter.Forget("k1")
	forgetter.Forget("k2")
	assert.Empty(t, failures, "forgetting all items must leave no residual entries")

	// Forget 未知 item 为安全的 no-op。
	forgetter.Forget("never-seen")
}

// TestMaxOfRateLimiter_When MaxOf 取各限流器最大值。
func TestMaxOfRateLimiter_When(t *testing.T) {
	// a: 1,2,4,8,16,32,64,100,100…（封顶 100ms）
	a := NewItemExponentialFailureRateLimiter(time.Millisecond, 100*time.Millisecond)
	// b: 50,60,60,60…（封顶 60ms）
	b := NewItemExponentialFailureRateLimiter(50*time.Millisecond, 60*time.Millisecond)

	limiter := NewMaxOfRateLimiter(a, b)

	want := []time.Duration{
		50 * time.Millisecond,  // max(1, 50)
		60 * time.Millisecond,  // max(2, 60)
		60 * time.Millisecond,  // max(4, 60)
		60 * time.Millisecond,  // max(8, 60)
		60 * time.Millisecond,  // max(16, 60)
		60 * time.Millisecond,  // max(32, 60)
		64 * time.Millisecond,  // max(64, 60)：a 越过 b 的封顶值后接管
		100 * time.Millisecond, // max(100, 60)
		100 * time.Millisecond, // max(100, 60)
	}
	for i, w := range want {
		assert.Equal(t, w, limiter.When("x"), "max-of at call %d", i+1)
	}
}

// TestMaxOfRateLimiter_EdgeCases 空组合与 nil 子项健壮性。
func TestMaxOfRateLimiter_EdgeCases(t *testing.T) {
	// 空组合。
	assert.Equal(t, time.Duration(0), NewMaxOfRateLimiter().When("x"))

	// nil 子项跳过。
	limiter := NewMaxOfRateLimiter(nil, NewItemExponentialFailureRateLimiter(3*time.Millisecond, time.Second))
	assert.Equal(t, 3*time.Millisecond, limiter.When("x"))
}

// TestMaxOfRateLimiter_Forget MaxOf 委托 Forget 给子限流器。
func TestMaxOfRateLimiter_Forget(t *testing.T) {
	a := NewItemExponentialFailureRateLimiter(time.Millisecond, time.Second)
	b := NewItemExponentialFailureRateLimiter(50*time.Millisecond, time.Second)
	limiter := NewMaxOfRateLimiter(a, NewNopRateLimiterImpl(), b)

	assert.Equal(t, 50*time.Millisecond, limiter.When("x"))  // max(1ms, 50ms)
	assert.Equal(t, 100*time.Millisecond, limiter.When("x")) // max(2ms, 100ms)

	forgetter, ok := limiter.(LimiterForgetter)
	if !ok {
		t.Fatal("max-of limiter must support LimiterForgetter")
	}

	forgetter.Forget("x")

	// 子限流器已重置。
	assert.Equal(t, 50*time.Millisecond, limiter.When("x"))
}

// TestItemExponentialFailureRateLimiter_Concurrent 并发下退避值合法。
func TestItemExponentialFailureRateLimiter_Concurrent(t *testing.T) {
	const goroutines = 32
	base, max := time.Millisecond, 10*time.Millisecond
	limiter := NewItemExponentialFailureRateLimiter(base, max)

	done := make(chan struct{})
	for g := 0; g < goroutines; g++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for i := 0; i < 100; i++ {
				d := limiter.When("shared")
				if d < base || d > max {
					t.Errorf("When returned %v outside [%v, %v]", d, base, max)
					return
				}
			}
		}()
	}
	for g := 0; g < goroutines; g++ {
		<-done
	}
}

// TestItemFastSlowRateLimiter_Sequence per-item 快/慢双速率限流。
func TestItemFastSlowRateLimiter_Sequence(t *testing.T) {
	limiter := NewItemFastSlowRateLimiterImpl(time.Millisecond, 10*time.Millisecond, 3)

	want := []time.Duration{
		time.Millisecond,      // 第 1 次：计数 0 < 3，fast
		time.Millisecond,      // 第 2 次：计数 1 < 3，fast
		time.Millisecond,      // 第 3 次：计数 2 < 3，fast
		10 * time.Millisecond, // 第 4 次：计数 3 达到上限，slow
		10 * time.Millisecond, // 第 5 次：此后恒 slow
	}
	for i, w := range want {
		assert.Equal(t, w, limiter.When("item-a"), "fast-slow sequence at call %d", i+1)
	}

	// 不同 item 计数独立。
	assert.Equal(t, time.Millisecond, limiter.When("item-b"))
	assert.Equal(t, time.Millisecond, limiter.When("item-b"))
}

// TestItemFastSlowRateLimiter_Forget Forget 清理计数，重回 fast 阶段。
func TestItemFastSlowRateLimiter_Forget(t *testing.T) {
	limiter := NewItemFastSlowRateLimiterImpl(time.Millisecond, 10*time.Millisecond, 2)

	forgetter, ok := limiter.(LimiterForgetter)
	if !ok {
		t.Fatal("fast-slow limiter must support LimiterForgetter")
	}

	assert.Equal(t, time.Millisecond, limiter.When("k1"))    // k1: fast，计数 0
	assert.Equal(t, time.Millisecond, limiter.When("k1"))    // k1: fast，计数 1
	assert.Equal(t, 10*time.Millisecond, limiter.When("k1")) // k1: slow，计数 2 达到上限
	assert.Equal(t, time.Millisecond, limiter.When("k2"))    // k2: fast，计数 0
	assert.Equal(t, time.Millisecond, limiter.When("k2"))    // k2: fast，计数 1

	failures := limiter.(*itemFastSlowRateLimiter).failures
	assert.Len(t, failures, 2)

	forgetter.Forget("k1")
	assert.Len(t, failures, 1)

	// k1 归零回 fast；k2 不受影响。
	assert.Equal(t, time.Millisecond, limiter.When("k1"))
	assert.Equal(t, 10*time.Millisecond, limiter.When("k2"))

	forgetter.Forget("k1")
	forgetter.Forget("k2")
	assert.Empty(t, failures, "forgetting all items must leave no residual entries")

	// Forget 未知 item 为安全的 no-op。
	forgetter.Forget("never-seen")
}

// TestItemFastSlowRateLimiter_NonPositiveDelays 非正 delay 返回 0。
func TestItemFastSlowRateLimiter_NonPositiveDelays(t *testing.T) {
	// 两阶段均非正。
	limiter := NewItemFastSlowRateLimiterImpl(-time.Second, 0, 2)
	for i := 0; i < 4; i++ {
		assert.Equal(t, time.Duration(0), limiter.When("x"), "call %d", i+1)
	}

	// fast 非正、slow 为正。
	mixed := NewItemFastSlowRateLimiterImpl(0, 5*time.Millisecond, 1)
	assert.Equal(t, time.Duration(0), mixed.When("y"))
	assert.Equal(t, 5*time.Millisecond, mixed.When("y"))
}

// TestItemFastSlowRateLimiter_NonPositiveMaxFastAttempts maxFastAttempts 非正时无 fast 阶段。
func TestItemFastSlowRateLimiter_NonPositiveMaxFastAttempts(t *testing.T) {
	for _, attempts := range []int{0, -1} {
		limiter := NewItemFastSlowRateLimiterImpl(time.Millisecond, 10*time.Millisecond, attempts)
		for i := 0; i < 3; i++ {
			assert.Equal(t, 10*time.Millisecond, limiter.When("x"), "maxFastAttempts=%d call %d", attempts, i+1)
		}
	}
}

// TestItemFastSlowRateLimiter_Concurrent 并发 When/Forget 安全。
func TestItemFastSlowRateLimiter_Concurrent(t *testing.T) {
	const goroutines = 16
	fast, slow := time.Millisecond, 20*time.Millisecond
	limiter := NewItemFastSlowRateLimiterImpl(fast, slow, 3)
	forgetter := limiter.(LimiterForgetter)

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				d := limiter.When("shared")
				if d != fast && d != slow {
					t.Errorf("When returned %v, want %v or %v", d, fast, slow)
					return
				}
				if i%10 == 0 {
					forgetter.Forget("shared")
				}
			}
		}()
	}
	wg.Wait()
}

// TestWithMaxWaitRateLimiter_CapsExceededWait 超过 maxWait 被截断，内层状态照常推进。
func TestWithMaxWaitRateLimiter_CapsExceededWait(t *testing.T) {
	inner := NewItemExponentialFailureRateLimiter(time.Millisecond, time.Second) // 1,2,4,8,16ms…
	limiter := NewWithMaxWaitRateLimiter(inner, 3*time.Millisecond)

	want := []time.Duration{
		time.Millisecond,     // 1ms < 3ms：原样返回
		2 * time.Millisecond, // 2ms < 3ms：原样返回
		3 * time.Millisecond, // 4ms > 3ms：封顶
		3 * time.Millisecond, // 8ms > 3ms：封顶
		3 * time.Millisecond, // 16ms > 3ms：封顶
	}
	for i, w := range want {
		assert.Equal(t, w, limiter.When("x"), "max-wait at call %d", i+1)
	}
}

// TestWithMaxWaitRateLimiter_ForgetDelegation Forget 委托重置 inner 退避。
func TestWithMaxWaitRateLimiter_ForgetDelegation(t *testing.T) {
	inner := NewItemExponentialFailureRateLimiter(time.Millisecond, time.Second)
	limiter := NewWithMaxWaitRateLimiter(inner, time.Second)

	assert.Equal(t, time.Millisecond, limiter.When("x"))
	assert.Equal(t, 2*time.Millisecond, limiter.When("x"))

	forgetter, ok := limiter.(LimiterForgetter)
	if !ok {
		t.Fatal("max-wait limiter must always support LimiterForgetter")
	}
	forgetter.Forget("x")

	// inner 已重置。
	assert.Equal(t, time.Millisecond, limiter.When("x"))
}

// TestWithMaxWaitRateLimiter_ForgetWithoutInnerForgetter inner 无 Forget 时包装器 Forget 为 no-op。
func TestWithMaxWaitRateLimiter_ForgetWithoutInnerForgetter(t *testing.T) {
	limiter := NewWithMaxWaitRateLimiter(NewNopRateLimiterImpl(), time.Second)

	forgetter, ok := limiter.(LimiterForgetter)
	if !ok {
		t.Fatal("max-wait limiter must always support LimiterForgetter")
	}
	forgetter.Forget("x") // 不得 panic
	assert.Equal(t, time.Duration(0), limiter.When("x"))
}

// TestWithMaxWaitRateLimiter_NonPositiveMaxWait maxWait 非正时 When 返回 0。
func TestWithMaxWaitRateLimiter_NonPositiveMaxWait(t *testing.T) {
	inner := NewItemExponentialFailureRateLimiter(time.Millisecond, time.Second)

	for _, maxWait := range []time.Duration{0, -time.Second} {
		limiter := NewWithMaxWaitRateLimiter(inner, maxWait)
		assert.Equal(t, time.Duration(0), limiter.When("x"), "maxWait=%v", maxWait)
	}
}

// TestWithMaxWaitRateLimiter_NilInner inner 为 nil 时 When 返回 0，Forget 为 no-op。
func TestWithMaxWaitRateLimiter_NilInner(t *testing.T) {
	limiter := NewWithMaxWaitRateLimiter(nil, time.Second)
	assert.Equal(t, time.Duration(0), limiter.When("x"))

	forgetter, ok := limiter.(LimiterForgetter)
	if !ok {
		t.Fatal("max-wait limiter must always support LimiterForgetter")
	}
	forgetter.Forget("x") // 不得 panic
}

// TestItemExponentialFailureRateLimiterWithJitter_ZeroRatioMatchesPlain 零 ratio 与普通退避一致。
func TestItemExponentialFailureRateLimiterWithJitter_ZeroRatioMatchesPlain(t *testing.T) {
	base, maxDelay := time.Millisecond, 100*time.Millisecond

	for _, ratio := range []float64{0, -0.5} {
		plain := NewItemExponentialFailureRateLimiter(base, maxDelay)
		withJitter := NewItemExponentialFailureRateLimiterWithJitter(base, maxDelay, ratio)
		for i := 0; i < 12; i++ {
			assert.Equal(t, plain.When("x"), withJitter.When("x"), "ratio=%v call %d", ratio, i+1)
		}
	}
}

// TestItemExponentialFailureRateLimiterWithJitter_Range jitter 落在 [d×(1-ratio), d]。
func TestItemExponentialFailureRateLimiterWithJitter_Range(t *testing.T) {
	base, maxDelay, ratio := 10*time.Millisecond, time.Second, 0.5
	limiter := NewItemExponentialFailureRateLimiterWithJitter(base, maxDelay, ratio)
	forgetter := limiter.(LimiterForgetter)

	lower := time.Duration(float64(base) * (1 - ratio))
	minSeen := maxDelay
	const samples = 100
	for i := 0; i < samples; i++ {
		// 重置计数，采样 exp=0 层级。
		forgetter.Forget("x")
		d := limiter.When("x")
		assert.GreaterOrEqual(t, d, lower, "sample %d", i)
		assert.LessOrEqual(t, d, base, "sample %d", i)
		if d < minSeen {
			minSeen = d
		}
	}
	assert.Less(t, minSeen, base, "jitter must be observable within %d samples", samples)
}

// TestItemExponentialFailureRateLimiterWithJitter_CapBeforeJitter 封顶先行，jitter 不超 maxDelay。
func TestItemExponentialFailureRateLimiterWithJitter_CapBeforeJitter(t *testing.T) {
	base, maxDelay, ratio := time.Millisecond, 10*time.Millisecond, 0.9
	limiter := NewItemExponentialFailureRateLimiterWithJitter(base, maxDelay, ratio)

	lower := time.Duration(float64(base) * (1 - ratio))
	// 确定值封顶为 10ms，jitter 后不超封顶值。
	for i := 0; i < 20; i++ {
		d := limiter.When("big-exp")
		assert.GreaterOrEqual(t, d, lower, "call %d", i+1)
		assert.LessOrEqual(t, d, maxDelay, "call %d", i+1)
	}
}

// TestItemExponentialFailureRateLimiterWithJitter_RatioClampedToFullJitter ratio>1 按 full jitter 处理。
func TestItemExponentialFailureRateLimiterWithJitter_RatioClampedToFullJitter(t *testing.T) {
	base, maxDelay := 10*time.Millisecond, time.Second
	limiter := NewItemExponentialFailureRateLimiterWithJitter(base, maxDelay, 2.5)
	forgetter := limiter.(LimiterForgetter)

	minSeen := maxDelay
	const samples = 100
	for i := 0; i < samples; i++ {
		forgetter.Forget("x")
		d := limiter.When("x") // 确定值 base=10ms，full jitter 后落在 [0, 10ms]
		assert.GreaterOrEqual(t, d, time.Duration(0), "sample %d: clamped ratio must never go negative", i)
		assert.LessOrEqual(t, d, base, "sample %d", i)
		if d < minSeen {
			minSeen = d
		}
	}
	assert.Less(t, minSeen, base/2, "full jitter must spread well below the deterministic value")
}

// TestItemExponentialFailureRateLimiterWithJitter_Concurrent 并发 When/Forget 安全。
func TestItemExponentialFailureRateLimiterWithJitter_Concurrent(t *testing.T) {
	base, maxDelay, ratio := time.Millisecond, 10*time.Millisecond, 0.5
	limiter := NewItemExponentialFailureRateLimiterWithJitter(base, maxDelay, ratio)
	forgetter := limiter.(LimiterForgetter)
	lower := time.Duration(float64(base) * (1 - ratio))

	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				d := limiter.When("shared")
				if d < lower || d > maxDelay {
					t.Errorf("jittered delay %v outside [%v, %v]", d, lower, maxDelay)
					return
				}
				if i%10 == 0 {
					forgetter.Forget("shared")
				}
			}
		}()
	}
	wg.Wait()
}

// TestApplyJitter_NaNRatio 钉住 P2-1：NaN ratio 原样返回 delay。
func TestApplyJitter_NaNRatio(t *testing.T) {
	assert.Equal(t, 100*time.Millisecond, applyJitter(100*time.Millisecond, math.NaN()))

	limiter := NewItemExponentialFailureRateLimiterWithJitter(time.Second, 10*time.Second, math.NaN())
	assert.Equal(t, time.Second, limiter.When("nan-item"))
}
