package workqueue

import (
	"math"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// nopRateLimiterImpl 是无操作限流器，When 始终返回零等待时长。
type nopRateLimiterImpl struct{}

// When 始终返回 0，表示无需等待。
func (rl *nopRateLimiterImpl) When(any) time.Duration { return 0 }

// NewNopRateLimiterImpl 返回始终无等待的限流器。
func NewNopRateLimiterImpl() Limiter { return &nopRateLimiterImpl{} }

// bucketRateLimiterImpl 使用 token bucket（令牌桶）策略的限流器，
// 基于 golang.org/x/time/rate 实现。
type bucketRateLimiterImpl struct {
	r *rate.Limiter
}

// When 预留一个令牌并返回对应的等待时长。
func (rl *bucketRateLimiterImpl) When(any) time.Duration {
	return rl.r.Reserve().Delay()
}

// NewBucketRateLimiterImpl 使用 token bucket 策略创建限流器。
func NewBucketRateLimiterImpl(r float64, burst int64) Limiter {
	if burst > math.MaxInt {
		burst = math.MaxInt
	}
	return &bucketRateLimiterImpl{
		r: rate.NewLimiter(rate.Limit(r), int(burst)),
	}
}

// itemExponentialFailureRateLimiter 按 item 追踪失败次数，
// 退避时长按失败次数指数增长：base×2^(n-1)，封顶 max。
// value 作为 map key 使用，必须是可比较类型。
type itemExponentialFailureRateLimiter struct {
	baseDelay time.Duration
	maxDelay  time.Duration

	mu       sync.Mutex
	failures map[any]int
}

// NewItemExponentialFailureRateLimiter 创建 per-item 指数退避限流器。
// 同一 item 首次 When 返回 base，此后逐次翻倍，超过 max 后恒为 max；
// base 或 max 非正时视为无等待，始终返回 0。
func NewItemExponentialFailureRateLimiter(base, max time.Duration) Limiter {
	return &itemExponentialFailureRateLimiter{
		baseDelay: base,
		maxDelay:  max,
		failures:  make(map[any]int),
	}
}

func (rl *itemExponentialFailureRateLimiter) When(value any) time.Duration {
	rl.mu.Lock()
	exp := rl.failures[value]
	rl.failures[value] = exp + 1
	rl.mu.Unlock()

	return rl.backoff(exp)
}

// Forget 移除 value 的失败计数，其后退避从 base 重新开始。
func (rl *itemExponentialFailureRateLimiter) Forget(value any) {
	rl.mu.Lock()
	delete(rl.failures, value)
	rl.mu.Unlock()
}

// backoff 计算 base×2^exp 并按 maxDelay 封顶。
// doubling 循环复用 exponentialRetryPolicyImpl.NextDelay 的溢出防护模式，
// 最多执行约 log2(max/base)+1 次。
func (rl *itemExponentialFailureRateLimiter) backoff(exp int) time.Duration {
	if rl.baseDelay <= 0 || rl.maxDelay <= 0 {
		return 0
	}

	delay := rl.baseDelay
	for i := 0; i < exp; i++ {
		if delay >= rl.maxDelay {
			return rl.maxDelay
		}
		if delay > rl.maxDelay/2 {
			delay = rl.maxDelay
			break
		}
		delay *= 2
	}

	if delay > rl.maxDelay {
		delay = rl.maxDelay
	}
	return delay
}

// maxOfRateLimiter 组合多个限流器，When 取各返回值的最大值。
type maxOfRateLimiter struct {
	limiters []Limiter
}

// NewMaxOfRateLimiter 创建组合限流器：When 返回所有子限流器等待时长的最大值。
// 空组合或全部子项为 nil 时始终返回 0；nil 子项被跳过。
func NewMaxOfRateLimiter(limiters ...Limiter) Limiter {
	filtered := make([]Limiter, 0, len(limiters))
	for _, l := range limiters {
		if l != nil {
			filtered = append(filtered, l)
		}
	}

	return &maxOfRateLimiter{limiters: filtered}
}

func (rl *maxOfRateLimiter) When(value any) time.Duration {
	var longest time.Duration
	for _, l := range rl.limiters {
		if d := l.When(value); d > longest {
			longest = d
		}
	}
	return longest
}

// Forget 委托给每个实现 LimiterForgetter 的子限流器，未实现的子项跳过。
func (rl *maxOfRateLimiter) Forget(value any) {
	for _, l := range rl.limiters {
		if f, ok := l.(LimiterForgetter); ok {
			f.Forget(value)
		}
	}
}
