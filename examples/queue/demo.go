package main

import (
	"errors"
	"fmt"

	wkq "github.com/shengyanli1982/workqueue/v2"
)

func main() {
	q := wkq.NewQueue(
		wkq.NewQueueConfig().WithValueIdempotent(),
	)
	defer q.Shutdown()

	_ = q.Put("job-1")
	if err := q.Put("job-1"); errors.Is(err, wkq.ErrElementAlreadyExist) {
		fmt.Println("dedup works: duplicate job ignored")
	}

	value, err := q.Get()
	if err != nil {
		fmt.Println("get failed:", err)
		return
	}
	fmt.Println("consumed:", value)
	q.Done(value)

	_, err = q.Get()
	if errors.Is(err, wkq.ErrQueueIsEmpty) {
		fmt.Println("queue is empty now")
	}
}
