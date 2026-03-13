package main

import (
	"errors"
	"fmt"
	"time"

	wkq "github.com/shengyanli1982/workqueue/v2"
)

func main() {
	cfg := wkq.NewLeasedQueueConfig().
		WithLeaseDuration(2 * time.Second).
		WithScanInterval(50 * time.Millisecond)

	q := wkq.NewLeasedQueue(cfg)
	defer q.Shutdown()

	_ = q.Put("email-user-42")

	value, leaseID, err := q.GetWithLease(500 * time.Millisecond)
	if err != nil {
		fmt.Println("get with lease failed:", err)
		return
	}
	fmt.Printf("leased task: value=%v leaseID=%s\n", value, leaseID)

	if err = q.Nack(leaseID, errors.New("temporary failure")); err != nil {
		fmt.Println("nack failed:", err)
		return
	}

	requeued, err := q.Get()
	if err != nil {
		fmt.Println("get requeued task failed:", err)
		return
	}
	fmt.Println("requeued task:", requeued)

	_ = q.Put("invoice-user-7")
	value, leaseID, err = q.GetWithLease(time.Second)
	if err != nil {
		fmt.Println("get second task with lease failed:", err)
		return
	}

	if err = q.ExtendLease(leaseID, 300*time.Millisecond); err != nil {
		fmt.Println("extend lease failed:", err)
		return
	}
	if err = q.Ack(leaseID); err != nil {
		fmt.Println("ack failed:", err)
		return
	}
	fmt.Println("acked task:", value)
}
