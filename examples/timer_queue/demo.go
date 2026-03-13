package main

import (
	"errors"
	"fmt"
	"time"

	wkq "github.com/shengyanli1982/workqueue/v2"
)

func main() {
	q := wkq.NewTimerQueue(nil)
	defer q.Shutdown()

	_ = q.PutAfter("send-metrics", 100*time.Millisecond)
	_ = q.PutAfter("flush-cache", 200*time.Millisecond)
	_ = q.PutAfter("cancel-me", 400*time.Millisecond)
	_ = q.Cancel("cancel-me")

	deadline := time.Now().Add(time.Second)
	consumed := 0

	for time.Now().Before(deadline) && consumed < 2 {
		value, err := q.Get()
		if err == nil {
			fmt.Println("scheduled task:", value)
			consumed++
			continue
		}
		if !errors.Is(err, wkq.ErrQueueIsEmpty) {
			fmt.Println("get failed:", err)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	if consumed < 2 {
		fmt.Println("timeout waiting scheduled tasks")
	}
}
