package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	wkq "github.com/shengyanli1982/workqueue/v2"
)

// consumer 函数是一个消费者函数，它从队列中获取元素并处理它们
// The consumer function is a consumer function that gets elements from the queue and processes them
func consumer(queue wkq.Queue, wg *sync.WaitGroup) {
	// 当函数返回时，调用 wg.Done() 来通知 WaitGroup 一个任务已经完成
	// When the function returns, call wg.Done() to notify the WaitGroup that a task has been completed
	defer wg.Done()

	// 无限循环，直到函数返回
	// Infinite loop until the function returns
	for {
		// 从队列中获取一个元素
		// Get an element from the queue
		element, err := queue.Get()

		// 如果获取元素时发生错误，则处理错误
		// If an error occurs when getting the element, handle the error
		if err != nil {
			// 如果错误不是因为队列为空，则打印错误并返回
			// If the error is not because the queue is empty, print the error and return
			if !errors.Is(err, wkq.ErrQueueIsEmpty) {
				fmt.Println(err)
				return
			} else {
				// 如果错误是因为队列为空，则继续循环
				// If the error is because the queue is empty, continue the loop
				continue
			}
		}

		// 打印获取到的元素
		// Print the obtained element
		fmt.Println("> get element:", element)

		// 标记元素为已处理，'Done' 是在 'Get' 之后必需的
		// Mark the element as done, 'Done' is required after 'Get'
		queue.Done(element)
	}
}

func main() {
	// 创建一个 WaitGroup，用于等待所有的 goroutine 完成
	// Create a WaitGroup to wait for all goroutines to complete
	wg := sync.WaitGroup{}

	// 创建一个新的桶形限流器，参数为桶的容量和填充速度
	// Create a new bucket rate limiter, the parameters are the capacity of the bucket and the fill rate
	limiter := wkq.NewBucketRateLimiterImpl(5, 1)

	// 创建一个新的限流队列配置，并设置其限流器
	// Create a new rate limiting queue configuration and set its limiter
	config := wkq.NewRateLimitingQueueConfig().WithLimiter(limiter)

	// 使用配置创建一个新的限流队列
	// Create a new rate limiting queue with the configuration
	queue := wkq.NewRateLimitingQueue(config)

	// 增加 WaitGroup 的计数器
	// Increase the counter of the WaitGroup
	wg.Add(1)

	// 启动一个新的 goroutine 来运行 consumer 函数
	// Start a new goroutine to run the consumer function
	go consumer(queue, &wg)

	// 将 "delay 1" 放入队列，并设置其延迟时间为 200 毫秒
	// Put "delay 1" into the queue and set its delay time to 200 milliseconds
	_ = queue.PutWithDelay("delay 1", 200)

	// 将 "delay 2" 放入队列，并设置其延迟时间为 100 毫秒
	// Put "delay 2" into the queue and set its delay time to 100 milliseconds
	_ = queue.PutWithDelay("delay 2", 100)

	// 将 "hello" 放入队列
	// Put "hello" into the queue
	_ = queue.Put("hello")

	// 将 "world" 放入队列
	// Put "world" into the queue
	_ = queue.Put("world")

	// 将 "limited" 放入队列, 触发限流
	// Put "limited" into the queue, trigger rate limiting
	for i := 0; i < 10; i++ {
		go func(i int) {
			_ = queue.PutWithLimited(fmt.Sprintf("limited %d", i))
		}(i)
	}

	// 等待一秒钟，让 comsumer 有机会处理队列中的元素
	// Wait for a second to give the comsumer a chance to process the elements in the queue
	time.Sleep(time.Second)

	// 关闭队列
	// Shut down the queue
	queue.Shutdown()

	// 等待所有的 goroutine 完成
	// Wait for all goroutines to complete
	wg.Wait()
}
