package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

func main() {
	q := workqueue.NewPriorityQueue(nil)

	go func() {
		for {
			element, err := q.Get()
			if err != nil {
				if !errors.Is(err, workqueue.ErrorQueueEmpty) {
					fmt.Println(err)
					return
				} else {
					continue
				}
			}
			fmt.Println("get element:", element)
			q.Done(element) // mark element as done, 'Done' is required after 'Get'
		}
	}()

	_ = q.Add("hello")
	_ = q.Add("world")
	_ = q.AddWeight("priority: 1", 1) // add element with priority
	_ = q.AddWeight("priority: 2", 2) // add element with priority

	time.Sleep(time.Second * 1) // wait for element to be executed

	q.Stop()
}
