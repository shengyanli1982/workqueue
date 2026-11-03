package workqueue

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// nopRateLimiterImpl 是无操作限流器，始终返回零等待时长。
type nopRateLimiterImpl struct{}

// When 始终返回 0，无需等待。
func (rl *nopRateLimiterImpl) When(any) time.Duration { return 0 }

// NewNopRateLimiterImpl 返回无等待的限流器。
func NewNopRateLimiterImpl() Limiter { return &nopRateLimiterImpl{} }

// bucketRateLimiterImpl 使用令牌桶策略，基于 golang.org/x/time/rate 实现。
type bucketRateLimiterImpl struct {
	r *rate.Limiter
}

// When 预留一个令牌并返回等待时长。
func (rl *bucketRateLimiterImpl) When(any) time.Duration {
	return rl.r.Reserve().Delay()
}

// NewBucketRateLimiterImpl 使用令牌桶策略创建限流器。
func NewBucketRateLimiterImpl(r float64, burst int64) Limiter {
	if burst > math.MaxInt {
		burst = math.MaxInt
	}
	return &bucketRateLimiterImpl{
		r: rate.NewLimiter(rate.Limit(r), int(burst)),
	}
}

// itemExponentialFailureRateLimiter 按 item 追踪失败次数，
// 退避时长 base×2^(n-1)，封顶 max。value 作为 map key 须可比较。
type itemExponentialFailureRateLimiter struct {
	baseDelay   time.Duration
	maxDelay    time.Duration
	jitterRatio float64

	mu       sync.Mutex
	failures map[any]int
}

// NewItemExponentialFailureRateLimiter 创建 per-item 指数退避限流器（无 jitter）。
// 同一 item 首次返回 base，此后逐次翻倍，超过 max 恒为 max。
// base 或 max 非正时始终返回 0。
func NewItemExponentialFailureRateLimiter(base, max time.Duration) Limiter {
	return NewItemExponentialFailureRateLimiterWithJitter(base, max, 0)
}

// NewItemExponentialFailureRateLimiterWithJitter 创建带抖动的 per-item 指数退避限流器：
// 封顶后向下扰动，均匀分布于 [delay×(1-ratio), delay]。
// ratio <= 0 无 jitter，ratio > 1 按 1 处理；base 或 max 非正时始终返回 0。
// 并发安全（互斥锁保护计数表）。
func NewItemExponentialFailureRateLimiterWithJitter(base, max time.Duration, ratio float64) Limiter {
	return &itemExponentialFailureRateLimiter{
		baseDelay:   base,
		maxDelay:    max,
		jitterRatio: ratio,
		failures:    make(map[any]int),
	}
}

func (rl *itemExponentialFailureRateLimiter) When(value any) time.Duration {
	rl.mu.Lock()
	exp := rl.failures[value]
	rl.failures[value] = exp + 1
	rl.mu.Unlock()

	return applyJitter(rl.backoff(exp), rl.jitterRatio)
}

// Forget 移除 value 的失败计数，退避从 base 重新开始。
func (rl *itemExponentialFailureRateLimiter) Forget(value any) {
	rl.mu.Lock()
	delete(rl.failures, value)
	rl.mu.Unlock()
}

// backoff 计算 base×2^exp 并按 maxDelay 封顶，循环最多 log2(max/base)+1 次。
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

// Forget 委托给每个实现 LimiterForgetter 的子限流器，未实现的跳过。
func (rl *maxOfRateLimiter) Forget(value any) {
	for _, l := range rl.limiters {
		if f, ok := l.(LimiterForgetter); ok {
			f.Forget(value)
		}
	}
}

// itemFastSlowRateLimiter 按 item 追踪调用次数：
// 前 maxFastAttempts 次返回 fastDelay，之后恒返回 slowDelay。
type itemFastSlowRateLimiter struct {
	fastDelay       time.Duration
	slowDelay       time.Duration
	maxFastAttempts int

	mu       sync.Mutex
	failures map[any]int
}

// NewItemFastSlowRateLimiterImpl 创建 per-item 快/慢双速率限流器：
// 前 maxFastAttempts 次返回 fastDelay，之后恒返回 slowDelay。
// delay 非正时对应阶段返回 0；maxFastAttempts 非正时恒返回 slowDelay。
// 实现 LimiterForgetter，Forget 重置计数。并发安全。
func NewItemFastSlowRateLimiterImpl(fastDelay, slowDelay time.Duration, maxFastAttempts int) Limiter {
	if fastDelay < 0 {
		fastDelay = 0
	}
	if slowDelay < 0 {
		slowDelay = 0
	}
	return &itemFastSlowRateLimiter{
		fastDelay:       fastDelay,
		slowDelay:       slowDelay,
		maxFastAttempts: maxFastAttempts,
		failures:        make(map[any]int),
	}
}

// When 返回等待时长并递增计数：计数 < maxFastAttempts 返回 fastDelay，否则 slowDelay。
func (rl *itemFastSlowRateLimiter) When(value any) time.Duration {
	rl.mu.Lock()
	n := rl.failures[value]
	rl.failures[value] = n + 1
	rl.mu.Unlock()

	if n < rl.maxFastAttempts {
		return rl.fastDelay
	}
	return rl.slowDelay
}

// Forget 移除 value 的计数，其后该 item 重新进入 fast 阶段。
func (rl *itemFastSlowRateLimiter) Forget(value any) {
	rl.mu.Lock()
	delete(rl.failures, value)
	rl.mu.Unlock()
}

// maxWaitRateLimiter 包装内层限流器，将 When 返回值封顶到 maxWait。
type maxWaitRateLimiter struct {
	inner   Limiter
	maxWait time.Duration
}

// NewWithMaxWaitRateLimiter 创建封顶包装限流器：When 超过 maxWait 时封顶。
// maxWait 非正时恒无等待；inner 为 nil 时 When 恒返回 0。
// 返回实例始终实现 LimiterForgetter：inner 实现时委托，否则 no-op。
func NewWithMaxWaitRateLimiter(limiter Limiter, maxWait time.Duration) Limiter {
	if maxWait < 0 {
		maxWait = 0
	}
	return &maxWaitRateLimiter{inner: limiter, maxWait: maxWait}
}

// When 委托 inner 并按 maxWait 封顶。inner 为 nil 时返回 0。
func (rl *maxWaitRateLimiter) When(value any) time.Duration {
	if rl.inner == nil {
		return 0
	}
	d := rl.inner.When(value)
	if d > rl.maxWait {
		return rl.maxWait
	}
	return d
}

// Forget 委托 inner（若实现 LimiterForgetter），否则 no-op。
func (rl *maxWaitRateLimiter) Forget(value any) {
	if f, ok := rl.inner.(LimiterForgetter); ok {
		f.Forget(value)
	}
}

// applyJitter 封顶后对 delay 施加向下扰动：结果均匀分布于 [delay×(1-ratio), delay]。
// ratio <= 0 或 NaN 无 jitter，ratio > 1 按 1 处理；delay 非正时原样返回。
func applyJitter(delay time.Duration, ratio float64) time.Duration {
	if !(ratio > 0) || delay <= 0 {
		return delay
	}
	if ratio > 1 {
		ratio = 1
	}
	out := time.Duration(float64(delay) * (1 - ratio*rand.Float64()))
	if out > delay {
		out = delay
	}
	return out
}
