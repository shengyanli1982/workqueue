package main

import (
	"errors"
	"fmt"
	"time"

	wkq "github.com/shengyanli1982/workqueue/v2"
)

func main() {
	limiter := wkq.NewBucketRateLimiterImpl(5, 1)
	cfg := wkq.NewRateLimitingQueueConfig().WithLimiter(limiter)
	q := wkq.NewRateLimitingQueue(cfg)
	defer q.Shutdown()

	_ = q.Put("immediate")
	for i := 0; i < 5; i++ {
		if err := q.PutWithLimited(fmt.Sprintf("limited-%d", i)); err != nil {
			fmt.Println("put limited failed:", err)
			return
		}
	}

	start := time.Now()
	deadline := start.Add(2 * time.Second)
	consumed := 0
	for time.Now().Before(deadline) && consumed < 6 {
		value, err := q.Get()
		if err == nil {
			fmt.Printf("consumed at %v: %v\n", time.Since(start).Round(10*time.Millisecond), value)
			consumed++
			continue
		}
		if !errors.Is(err, wkq.ErrQueueIsEmpty) {
			fmt.Println("get failed:", err)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	if consumed < 6 {
		fmt.Println("timeout waiting rate-limited tasks")
	}
}
