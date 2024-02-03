package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

func main() {
	q := workqueue.NewDelayingQueue(nil)

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
	_ = q.AddAfter("delay 1 sec", time.Second)
	_ = q.AddAfter("delay 2 sec", time.Second*2)

	time.Sleep(time.Second * 4) // wait for element to be executed

	q.Stop()
}
