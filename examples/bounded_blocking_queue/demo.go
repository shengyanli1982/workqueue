package main

import (
	"context"
	"fmt"
	"time"

	wkq "github.com/shengyanli1982/workqueue/v2"
)

func main() {
	q := wkq.NewBoundedBlockingQueue(
		wkq.NewBoundedBlockingQueueConfig().WithCapacity(1),
	)
	defer q.Shutdown()

	_ = q.Put("job-1")

	putResult := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		putResult <- q.PutWithContext(ctx, "job-2")
	}()

	time.Sleep(100 * time.Millisecond)

	v1, err := q.GetWithContext(context.Background())
	if err != nil {
		fmt.Println("get job-1 failed:", err)
		return
	}
	fmt.Println("consumed:", v1)

	if err = <-putResult; err != nil {
		fmt.Println("put job-2 failed:", err)
		return
	}

	v2, err := q.GetWithContext(context.Background())
	if err != nil {
		fmt.Println("get job-2 failed:", err)
		return
	}
	fmt.Println("consumed:", v2)
}
