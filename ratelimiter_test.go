package workqueue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestItemExponentialFailureRateLimiter_BackoffSequence 验证 per-item 指数退避：
// 同一 item 的退避序列为 base×2^(n-1) 并按 max 封顶；不同 item 计数互相独立。
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

	// 另一个 item 从 base 重新开始，证明计数是 per-item 的。
	assert.Equal(t, time.Millisecond, limiter.When("item-b"))
	assert.Equal(t, 2*time.Millisecond, limiter.When("item-b"))
}

// TestItemExponentialFailureRateLimiter_Forget 验证 Forget 清理退避状态：
// Forget 后同一 item 的退避从 base 重新开始，且内部失败表删除该条目（防内存泄漏）。
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

	// k1 计数归零：退避从 base 重新开始；k2 不受影响。
	assert.Equal(t, time.Millisecond, limiter.When("k1"))
	assert.Equal(t, 2*time.Millisecond, limiter.When("k2"))

	forgetter.Forget("k1")
	forgetter.Forget("k2")
	assert.Empty(t, failures, "forgetting all items must leave no residual entries")

	// Forget 未知 item 为安全的 no-op。
	forgetter.Forget("never-seen")
}

// TestMaxOfRateLimiter_When 验证 MaxOf 组合取各限流器返回值的最大值，
// 包括封顶项接管另一个仍在增长的项的交叉场景。
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

// TestMaxOfRateLimiter_EdgeCases 验证空组合与 nil 子项的健壮性。
func TestMaxOfRateLimiter_EdgeCases(t *testing.T) {
	// 空组合：无等待。
	assert.Equal(t, time.Duration(0), NewMaxOfRateLimiter().When("x"))

	// nil 子项被跳过，其余子项正常取大。
	limiter := NewMaxOfRateLimiter(nil, NewItemExponentialFailureRateLimiter(3*time.Millisecond, time.Second))
	assert.Equal(t, 3*time.Millisecond, limiter.When("x"))
}

// TestMaxOfRateLimiter_Forget 验证 MaxOf 将 Forget 委托给每个实现 LimiterForgetter
// 的子限流器；未实现的子项（如 Nop）被安全跳过。
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

	// 两个子限流器均已重置：退避回到各自 base 的最大值。
	assert.Equal(t, 50*time.Millisecond, limiter.When("x"))
}

// TestItemExponentialFailureRateLimiter_Concurrent 验证并发下计数与封顶的正确性：
// N 个 goroutine 并发 When 同一 item，退避值均为合法序列成员且最终封顶于 max。
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
