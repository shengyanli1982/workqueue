package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

func main() {
	conf := workqueue.NewRateLimitingQConfig()
	conf.WithLimiter(workqueue.NewBucketRateLimiter(float64(4), 1))

	q := workqueue.NewRateLimitingQueue(conf)

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
			fmt.Printf("[%s] get element: %s\n", time.Now().Format("04:05"), element)
			q.Done(element) // mark element as done, 'Done' is required after 'Get'
		}
	}()

	_ = q.Add("hello")
	_ = q.Add("world")

	for i := 0; i < 10; i++ {
		_ = q.AddLimited(fmt.Sprintf(">>> %d", i))
	}

	time.Sleep(time.Second * 3) // wait for element to be executed

	q.Stop()
}
