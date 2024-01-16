package workqueue

import (
	"math"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// 定义一个限速器接口
// Defines the rate limiter interface.
type RateLimiter interface {
	// 获取一个元素应该等待多长时间
	// When gets an element and gets to decide how long that element should wait
	When(element any) time.Duration

	// 停止追踪这个元素，不让他继续在队列中等待。
	// Forget indicates that an element is finished being retried.  Doesn't matter whether it's for failing
	// or for success, we'll stop tracking it
	Forget(element any)

	// 返回一个元素被限速的次数
	// NumLimitTimes returns back limit times the element has had
	NumLimitTimes(element any) int

	// 关闭限速器
	// Stop stops the limiter
	Stop()
}

// 实现一个基于令牌桶 RateLimiter
// Implements a rate limiter that uses a token bucket.
type BucketRateLimiter struct {
	*rate.Limiter
}

// 创建一个基于令牌桶 RateLimiter
// NewBucketRateLimiter returns a new instance of a bucket rate limiter.
func NewBucketRateLimiter(r float64, b int) *BucketRateLimiter {
	return &BucketRateLimiter{rate.NewLimiter(rate.Limit(r), b)}
}

// 创建一个基于令牌桶 RateLimiter, 拥有默认值
// DefaultBucketRateLimiter returns a new instance of a bucket rate limiter with default values.
func DefaultBucketRateLimiter() RateLimiter {
	return NewBucketRateLimiter(defaultQueueRateLimit, defaultQueueRateBurst)
}

// 获取一个元素，并决定该元素应该等待多长时间。这里是 BucketRateLimiter 的核心逻辑。 等待时间：limiter.Reserve().Delay()
// When gets an element and gets to decide how long that element should wait
func (r *BucketRateLimiter) When(element any) time.Duration {
	return r.Limiter.Reserve().Delay()
}

// 返回一个元素被限速的次数
// Return the number of times an element is limited
func (r *BucketRateLimiter) NumLimitTimes(element any) int {
	// do nothing
	return 0
}

// 关闭令牌桶
// Stop stops the limiter
func (r *BucketRateLimiter) Stop() {
	// do nothing
}

// 停止追踪这个元素，不让他继续在队列中等待。
// Forget indicates that an element is finished being retried.  Doesn't matter whether it's for failing
func (r *BucketRateLimiter) Forget(element any) {
	// do nothing
}

// 实现一个基于指数退避的 RateLimiter
// Implements a rate limiter that uses exponential backoff.
type ExponentialFailureRateLimiter struct {
	lock      sync.Mutex
	failures  map[any]int
	basedelay time.Duration
	maxdelay  time.Duration
}

// 创建一个基于指数退避的 RateLimiter
// NewExponentialFailureRateLimiter returns a new instance of an exponential failure rate limiter.
func NewExponentialFailureRateLimiter(base time.Duration, max time.Duration) RateLimiter {
	return &ExponentialFailureRateLimiter{
		lock:      sync.Mutex{},
		failures:  map[any]int{},
		basedelay: base,
		maxdelay:  max,
	}
}

// 创建一个基于指数退避的 RateLimiter, 拥有默认值
// DefaultExponentialFailureRateLimiter returns a new instance of an exponential failure rate limiter with default values.
func DefaultExponentialFailureRateLimiter() RateLimiter {
	return NewExponentialFailureRateLimiter(defaultQueueExpFailureBase, defaultQueueExpFailureMax)
}

// 获取一个元素，并决定该元素应该等待多长时间。这里是 ExponentialFailureRateLimiter 的核心逻辑。等待时间：2^exp * base
// When gets an element and gets to decide how long that element should wait
func (r *ExponentialFailureRateLimiter) When(item any) time.Duration {
	r.lock.Lock()
	defer r.lock.Unlock()

	exp := r.failures[item]
	r.failures[item]++

	// Calculate the backoff using exponential formula
	backoff := r.basedelay * time.Duration(math.Pow(2, float64(exp))) // 2^exp * base

	// Cap the backoff to avoid overflow
	if backoff > r.maxdelay {
		return r.maxdelay
	}

	return backoff
}

// 返回一个元素被限速的次数
// Return the number of times an element is limited
func (r *ExponentialFailureRateLimiter) NumLimitTimes(item any) int {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.failures[item]
}

// 停止追踪这个元素，不让他继续在队列中等待。
// Forget indicates that an element is finished being retried.  Doesn't matter whether it's for failing
func (r *ExponentialFailureRateLimiter) Forget(item any) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.failures, item)
}

// 关闭指数退避
// Stop stops the limiter
func (r *ExponentialFailureRateLimiter) Stop() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.failures = map[any]int{}
}
