package workqueue

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter 定义了一个限速器接口
// RateLimiter defines a rate limiter interface
type RateLimiter interface {
	// When 方法获取一个元素，并决定该元素应该等待多长时间
	// The When method gets an element and decides how long the element should wait
	When(element any) time.Duration

	// Forget 方法表示一个元素已经完成重试，无论是失败还是成功，我们都会停止追踪它
	// The Forget method indicates that an element has finished retrying. Whether it's a failure or a success, we'll stop tracking it
	Forget(element any)

	// NumLimitTimes 方法返回一个元素被限速的次数
	// The NumLimitTimes method returns the number of times an element has been rate-limited
	NumLimitTimes(element any) int

	// Stop 方法停止限速器
	// The Stop method stops the limiter
	Stop()
}

// BucketRateLimiter 结构体实现了一个基于令牌桶的限速器
// The BucketRateLimiter struct implements a rate limiter based on a token bucket
type BucketRateLimiter struct {
	*rate.Limiter
}

// NewBucketRateLimiter 函数返回一个新的基于令牌桶的限速器实例
// The NewBucketRateLimiter function returns a new instance of a rate limiter based on a token bucket
func NewBucketRateLimiter(rateLimit float64, burst int) *BucketRateLimiter {
	return &BucketRateLimiter{rate.NewLimiter(rate.Limit(rateLimit), burst)}
}

// DefaultBucketRateLimiter 函数返回一个新的具有默认值的基于令牌桶的限速器实例
// The DefaultBucketRateLimiter function returns a new instance of a rate limiter based on a token bucket with default values
func DefaultBucketRateLimiter() RateLimiter {
	return NewBucketRateLimiter(defaultQueueRateLimit, defaultQueueRateBurst)
}

// When 方法获取一个元素，并决定该元素应该等待多长时间。这是 BucketRateLimiter 的核心逻辑。等待时间由 limiter.Reserve().Delay() 计算得出
// The When method gets an element and decides how long the element should wait. This is the core logic of BucketRateLimiter. The wait time is calculated by limiter.Reserve().Delay()
func (r *BucketRateLimiter) When(element any) time.Duration {
	return r.Limiter.Reserve().Delay()
}

// NumLimitTimes 方法返回一个元素被限速的次数
// The NumLimitTimes method returns the number of times an element has been rate-limited
func (r *BucketRateLimiter) NumLimitTimes(element any) int {
	// do nothing
	return 0
}

// Stop 方法停止限速器
// The Stop method stops the limiter
func (r *BucketRateLimiter) Stop() {
	// do nothing
}

// Forget 方法表示一个元素已经完成重试，无论是失败还是成功，我们都会停止追踪它
// The Forget method indicates that an element has finished retrying. Whether it's a failure or a success, we'll stop tracking it
func (r *BucketRateLimiter) Forget(element any) {
	// do nothing
}

// ExponentialFailureRateLimiter 结构体实现了一个基于指数退避的限速器
// The ExponentialFailureRateLimiter struct implements a rate limiter that uses exponential backoff
type ExponentialFailureRateLimiter struct {
	lock      sync.Mutex    // 用于保护 failures map 的互斥锁
	failures  map[any]int   // 存储每个元素的失败次数
	basedelay time.Duration // 基础延迟时间
	maxdelay  time.Duration // 最大延迟时间
}

// NewExponentialFailureRateLimiter 函数返回一个新的基于指数退避的限速器实例
// NewExponentialFailureRateLimiter returns a new instance of an exponential failure rate limiter
func NewExponentialFailureRateLimiter(base time.Duration, max time.Duration) RateLimiter {
	return &ExponentialFailureRateLimiter{
		lock:      sync.Mutex{},
		failures:  make(map[any]int),
		basedelay: base,
		maxdelay:  max,
	}
}

// DefaultExponentialFailureRateLimiter 函数返回一个新的具有默认值的基于指数退避的限速器实例
// DefaultExponentialFailureRateLimiter returns a new instance of an exponential failure rate limiter with default values
func DefaultExponentialFailureRateLimiter() RateLimiter {
	return NewExponentialFailureRateLimiter(defaultQueueExpFailureBase, defaultQueueExpFailureMax)
}

// When 方法获取一个元素，并决定该元素应该等待多长时间。这是 ExponentialFailureRateLimiter 的核心逻辑。等待时间：2^exp * base
// When gets an element and gets to decide how long that element should wait
func (r *ExponentialFailureRateLimiter) When(item any) time.Duration {
	r.lock.Lock()
	defer r.lock.Unlock()

	// 获取元素的失败次数
	// Get the number of times an element has failed
	exp := r.failures[item]
	r.failures[item]++

	// 计算退避时间，使用向左位移
	// Calculate the backoff time using left shift
	backoff := r.basedelay << uint(exp) // base * 2^exp

	// 如果退避时间大于最大退避时间，就使用最大退避时间
	// If the backoff time is greater than the maximum backoff time, use the maximum backoff time
	if backoff > r.maxdelay {
		return r.maxdelay
	}

	return backoff
}

// NumLimitTimes 方法返回一个元素被限速的次数
// The NumLimitTimes method returns the number of times an element has been rate-limited
func (r *ExponentialFailureRateLimiter) NumLimitTimes(item any) int {
	r.lock.Lock()
	defer r.lock.Unlock()

	// 获取并返回元素的失败次数
	// Get and return the number of failures of the element
	return r.failures[item]
}

// Forget 方法表示一个元素已经完成重试，无论是失败还是成功，我们都会停止追踪它
// The Forget method indicates that an element has finished retrying. Whether it's a failure or a success, we'll stop tracking it
func (r *ExponentialFailureRateLimiter) Forget(item any) {
	r.lock.Lock()
	defer r.lock.Unlock()

	// 从 failures map 中删除元素的失败次数
	// Delete the number of failures of the element from the failures map
	delete(r.failures, item)
}

// Stop 方法停止限速器
// Stop stops the limiter
func (r *ExponentialFailureRateLimiter) Stop() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.failures = make(map[any]int)
}
