package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

func main() {
	q := workqueue.NewSimpleQueue(nil)

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
	_ = q.Add("hello") // duplicate element
	_ = q.Add("world") // duplicate element

	time.Sleep(time.Second) // wait for element to be executed

	q.Stop()
}
