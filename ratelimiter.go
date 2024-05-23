package workqueue

import "time"

type NopRateLimiter struct{}

func (rl *NopRateLimiter) When(interface{}) time.Duration { return 0 }

func NewNopRateLimiterImpl() Limiter { return &NopRateLimiter{} }
