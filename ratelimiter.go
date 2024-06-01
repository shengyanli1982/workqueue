package workqueue

import (
	"time"

	"golang.org/x/time/rate"
)

// nopRateLimiterImpl 结构体定义了一个无操作的速率限制器实现。
// The nopRateLimiterImpl struct defines a no-operation implementation of a rate limiter.
type nopRateLimiterImpl struct{}

// When 方法返回元素应该被放入队列的时间，此处为 0，表示无延迟。
// The When method returns the time when the element should be put into the queue, here it is 0, indicating no delay.
func (rl *nopRateLimiterImpl) When(interface{}) time.Duration { return 0 }

// NewNopRateLimiterImpl 函数创建并返回一个新的 NopRateLimiterImpl 实例。
// The NewNopRateLimiterImpl function creates and returns a new instance of NopRateLimiterImpl.
func NewNopRateLimiterImpl() Limiter { return &nopRateLimiterImpl{} }

// bucketRateLimiterImpl 结构体定义了一个桶速率限制器实现。
// The bucketRateLimiterImpl struct defines an implementation of a bucket rate limiter.
type bucketRateLimiterImpl struct {
	r *rate.Limiter
}

// When 方法返回元素应该被放入队列的时间，此处为桶速率限制器的延迟。
// The When method returns the time when the element should be put into the queue, here it is the delay of the bucket rate limiter.
func (rl *bucketRateLimiterImpl) When(interface{}) time.Duration {
	return rl.r.Reserve().Delay()
}

// NewBucketRateLimiterImpl 函数创建并返回一个新的 BucketRateLimiterImpl 实例。
// The NewBucketRateLimiterImpl function creates and returns a new instance of BucketRateLimiterImpl.
func NewBucketRateLimiterImpl(r float64, burst int64) Limiter {
	// rate.NewLimiter 创建一个新的速率限制器，参数 rate.Limit(r) 设置每秒可以处理的元素数量，int(burst) 设置桶的大小。
	// rate.NewLimiter creates a new rate limiter, the parameter rate.Limit(r) sets the number of elements that can be handled per second, and int(burst) sets the size of the bucket.
	return &bucketRateLimiterImpl{
		r: rate.NewLimiter(rate.Limit(r), int(burst)),
	}
}
