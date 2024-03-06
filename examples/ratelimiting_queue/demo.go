package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

func main() {
	// 创建一个新的限速队列配置
	// Create a new rate limiting queue configuration
	conf := workqueue.NewRateLimitingQConfig()

	// 使用桶形限速器，参数为令牌生成速率和桶的大小
	// Use a bucket rate limiter, the parameters are the token generation rate and the size of the bucket
	conf.WithLimiter(workqueue.NewBucketRateLimiter(float64(4), 1))

	// 创建一个新的限速队列
	// Create a new rate limiting queue
	q := workqueue.NewRateLimitingQueue(conf)

	// 启动一个新的 goroutine 来处理队列中的元素
	// Start a new goroutine to handle elements in the queue
	go func() {
		// 循环获取队列中的元素
		// Loop to get elements from the queue
		for {
			// 从队列中获取一个元素
			// Get an element from the queue
			element, err := q.Get()

			// 如果获取元素时发生错误，则处理错误
			// If an error occurs when getting the element, handle the error
			if err != nil {
				// 如果错误不是因为队列为空，则打印错误并返回
				// If the error is not because the queue is empty, print the error and return
				if !errors.Is(err, workqueue.ErrorQueueEmpty) {
					fmt.Println(err)
					return
				} else {
					// 如果错误是因为队列为空，则继续循环
					// If the error is because the queue is empty, continue the loop
					continue
				}
			}

			// 打印获取到的元素和当前时间
			// Print the obtained element and the current time
			fmt.Printf("[%s] get element: %s\n", time.Now().Format("04:05"), element)

			// 标记元素为已处理，'Done' 是在 'Get' 之后必需的
			// Mark the element as done, 'Done' is required after 'Get'
			q.Done(element)
		}
	}()

	// 向队列中添加元素
	// Add elements to the queue
	_ = q.Add("hello")
	_ = q.Add("world")

	// 向队列中添加限速元素
	// Add rate-limited elements to the queue
	for i := 0; i < 10; i++ {
		_ = q.AddLimited(fmt.Sprintf(">>> %d", i))
	}

	// 等待元素被执行
	// Wait for the elements to be executed
	time.Sleep(time.Second * 3)

	// 停止队列
	// Stop the queue
	q.Stop()
}
