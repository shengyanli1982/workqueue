package main

import (
	"fmt"
	"time"

	wkq "github.com/shengyanli1982/workqueue/v2"
)

func main() {
	dlq := wkq.NewDeadLetterQueue(nil)
	defer dlq.Shutdown()

	target := wkq.NewQueue(nil)
	defer target.Shutdown()

	err := dlq.PutDead(&wkq.DeadLetter{
		Payload:     "order-1001",
		SourceQueue: "order-worker",
		Attempts:    3,
		LastError:   "upstream timeout",
		FailedAt:    time.Now(),
		Meta: map[string]string{
			"tenant": "acme",
		},
	})
	if err != nil {
		fmt.Println("put dead letter failed:", err)
		return
	}

	letter, err := dlq.GetDead()
	if err != nil {
		fmt.Println("get dead letter failed:", err)
		return
	}
	fmt.Printf("dead letter: id=%s payload=%v attempts=%d\n", letter.ID, letter.Payload, letter.Attempts)

	if err = dlq.RequeueDead(letter, target); err != nil {
		fmt.Println("requeue dead letter failed:", err)
		return
	}

	recovered, err := target.Get()
	if err != nil {
		fmt.Println("get recovered task failed:", err)
		return
	}
	fmt.Println("recovered task:", recovered)
}
