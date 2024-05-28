package workqueue

import (
	"time"

	"golang.org/x/time/rate"
)

type NopRateLimiterImpl struct{}

func (rl *NopRateLimiterImpl) When(interface{}) time.Duration { return 0 }

func NewNopRateLimiterImpl() Limiter { return &NopRateLimiterImpl{} }

type BucketRateLimiterImpl struct {
	r *rate.Limiter
}

func (rl *BucketRateLimiterImpl) When(interface{}) time.Duration {
	return rl.r.Reserve().Delay()
}

func NewBucketRateLimiterImpl(r float64, burst int64) Limiter {
	return &BucketRateLimiterImpl{
		r: rate.NewLimiter(rate.Limit(r), int(burst)),
	}
}
