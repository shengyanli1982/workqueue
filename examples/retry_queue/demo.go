package main

import (
	"errors"
	"fmt"
	"time"

	wkq "github.com/shengyanli1982/workqueue/v2"
)

func main() {
	cfg := wkq.NewRetryQueueConfig().
		WithPolicy(wkq.NewExponentialRetryPolicy(100*time.Millisecond, 500*time.Millisecond, 3))

	q := wkq.NewRetryQueue(cfg)
	defer q.Shutdown()

	_ = q.Put("sync-order-1001")

	value, err := q.Get()
	if err != nil {
		fmt.Println("first get failed:", err)
		return
	}
	fmt.Println("first consume:", value)

	if err = q.Retry(value, errors.New("upstream 503")); err != nil {
		fmt.Println("retry failed:", err)
		return
	}
	fmt.Println("retry count:", q.NumRequeues(value))

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		value, err = q.Get()
		if err == nil {
			fmt.Println("second consume:", value)
			q.Done(value)
			q.Forget(value)
			return
		}
		if !errors.Is(err, wkq.ErrQueueIsEmpty) {
			fmt.Println("second get failed:", err)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println("timeout waiting retried task")
}
