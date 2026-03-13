package workqueue

import (
	"time"

	"golang.org/x/time/rate"
)

type nopRateLimiterImpl struct{}

func (rl *nopRateLimiterImpl) When(interface{}) time.Duration { return 0 }

// NewNopRateLimiterImpl 返回始终无等待的限流器。
func NewNopRateLimiterImpl() Limiter { return &nopRateLimiterImpl{} }

type bucketRateLimiterImpl struct {
	r *rate.Limiter
}

func (rl *bucketRateLimiterImpl) When(interface{}) time.Duration {
	return rl.r.Reserve().Delay()
}

// NewBucketRateLimiterImpl 使用 token bucket 策略创建限流器。
func NewBucketRateLimiterImpl(r float64, burst int64) Limiter {

	return &bucketRateLimiterImpl{
		r: rate.NewLimiter(rate.Limit(r), int(burst)),
	}
}
