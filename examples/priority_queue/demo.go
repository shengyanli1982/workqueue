package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

func main() {
	// 创建一个新的优先级队列
	// Create a new priority queue
	q := workqueue.NewPriorityQueue(nil)

	// 启动一个新的 goroutine 来处理队列中的元素
	// Start a new goroutine to handle elements in the queue
	go func() {
		// 无限循环，直到队列为空
		// Infinite loop until the queue is empty
		for {
			// 从队列中获取元素
			// Get an element from the queue
			element, err := q.Get()

			// 如果获取元素时出现错误
			// If there is an error when getting the element
			if err != nil {
				// 如果错误不是因为队列为空
				// If the error is not because the queue is empty
				if !errors.Is(err, workqueue.ErrorQueueEmpty) {
					// 打印错误并返回
					// Print the error and return
					fmt.Println(err)
					return
				} else {
					// 如果错误是因为队列为空，继续循环
					// If the error is because the queue is empty, continue the loop
					continue
				}
			}

			// 打印获取到的元素
			// Print the obtained element
			fmt.Println("get element:", element)

			// 标记元素已完成，'Done' 是 'Get' 之后必须的
			// Mark the element as done, 'Done' is required after 'Get'
			q.Done(element)
		}
	}()

	// 向队列中添加元素
	// Add elements to the queue
	_ = q.Add("hello")
	_ = q.Add("world")
	// 添加带有优先级的元素
	// Add elements with priority
	_ = q.AddWeight("priority: 1", 1)
	_ = q.AddWeight("priority: 2", 2)

	// 等待元素被执行
	// Wait for the element to be executed
	time.Sleep(time.Second * 1)

	// 停止队列
	// Stop the queue
	q.Stop()
}
