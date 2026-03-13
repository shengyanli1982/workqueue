package main

import (
	"errors"
	"fmt"
	"time"

	wkq "github.com/shengyanli1982/workqueue/v2"
)

func main() {
	q := wkq.NewDelayingQueue(nil)
	defer q.Shutdown()

	_ = q.Put("immediate")
	_ = q.PutWithDelay("delay-100ms", 100)
	_ = q.PutWithDelay("delay-200ms", 200)

	deadline := time.Now().Add(2 * time.Second)
	consumed := 0
	for time.Now().Before(deadline) && consumed < 3 {
		value, err := q.Get()
		if err == nil {
			fmt.Println("consumed:", value)
			consumed++
			continue
		}
		if !errors.Is(err, wkq.ErrQueueIsEmpty) {
			fmt.Println("get failed:", err)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	if consumed < 3 {
		fmt.Println("timeout waiting delayed tasks")
	}
}
