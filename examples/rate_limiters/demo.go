package main

import (
	"errors"
	"fmt"
	"time"

	wkq "github.com/shengyanli1982/workqueue/v2"
)

func main() {
	demoFastSlow()
	fmt.Println()
	demoMaxWait()
	fmt.Println()
	demoJitter()
}

// demoFastSlow 演示 FastSlow 限流器：前 maxFastAttempts 次返回 fastDelay，之后返回 slowDelay；
// Forget 重置计数。
func demoFastSlow() {
	fmt.Println("fast-slow limiter: fastDelay=10ms, slowDelay=1s, maxFastAttempts=3")
	limiter := wkq.NewItemFastSlowRateLimiterImpl(10*time.Millisecond, time.Second, 3)

	// 前 3 次为 fast 阶段（10ms），之后切换 slow 阶段（1s）。
	for i := 1; i <= 5; i++ {
		fmt.Printf("  When(item) #%d = %v\n", i, limiter.When("item"))
	}

	// Forget 经 LimiterForgetter 可选接口调用。
	forgetter, ok := limiter.(wkq.LimiterForgetter)
	if !ok {
		fmt.Println("  limiter does not implement LimiterForgetter")
		return
	}
	forgetter.Forget("item")
	fmt.Printf("  When(item) after Forget = %v (back to fast phase)\n", limiter.When("item"))
}

// demoMaxWait 演示 MaxWait 封顶包装：When 结果超过 maxWait 时被封顶，
// inner 退避状态照常推进。
func demoMaxWait() {
	fmt.Println("max-wait limiter: inner=exponential(base=100ms, max=30s), maxWait=2s")
	inner := wkq.NewItemExponentialFailureRateLimiter(100*time.Millisecond, 30*time.Second)
	limiter := wkq.NewWithMaxWaitRateLimiter(inner, 2*time.Second)

	// 前 5 次未触顶，原样返回；第 6、7 次超过 maxWait=2s 被封顶。
	for i := 1; i <= 7; i++ {
		fmt.Printf("  When(item) #%d = %v\n", i, limiter.When("item"))
	}
}

// demoJitter 演示两个 jitter 变体：delay 在封顶之后向下扰动，
// 均匀分布于 [delay×(1-ratio), delay]。
func demoJitter() {
	fmt.Println("jitter variants: ratio=0.5, delay ∈ [d×(1-ratio), d]")

	// Limiter 变体：When 内部自增计数，需 Forget 重置后重复采样首次层级。
	limiter := wkq.NewItemExponentialFailureRateLimiterWithJitter(100*time.Millisecond, time.Second, 0.5)
	forgetter, ok := limiter.(wkq.LimiterForgetter)
	if !ok {
		fmt.Println("  limiter does not implement LimiterForgetter")
		return
	}
	fmt.Println("  limiter samples at first level (raw=100ms, expect [50ms, 100ms]):")
	for i := 1; i <= 5; i++ {
		d := limiter.When("item")
		fmt.Printf("    sample #%d = %v\n", i, d.Round(100*time.Microsecond))
		forgetter.Forget("item")
	}

	// RetryPolicy 变体：attempt 由调用方传入，同一层级可直接重复采样。
	policy := wkq.NewExponentialRetryPolicyWithJitter(100*time.Millisecond, 2*time.Second, 3, 0.5)
	fmt.Println("  policy samples at attempt=3 (raw=400ms, expect [200ms, 400ms]):")
	for i := 1; i <= 5; i++ {
		d, retry := policy.NextDelay("item", 3, errors.New("transient failure"))
		fmt.Printf("    sample #%d = %v (retry=%v)\n", i, d.Round(100*time.Microsecond), retry)
	}

	// 超过 maxRetries 后终止重试，返回 (0, false)。
	d, retry := policy.NextDelay("item", 4, errors.New("transient failure"))
	fmt.Printf("  attempt=4 exceeds maxRetries=3: delay=%v retry=%v\n", d, retry)
}
