package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

func main() {
	// 创建一个新的队列
	// Create a new queue
	q := workqueue.NewQueue(nil)

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

			// 打印获取到的元素
			// Print the obtained element
			fmt.Println("get element:", element)

			// 标记元素为已处理，'Done' 是在 'Get' 之后必需的
			// Mark the element as done, 'Done' is required after 'Get'
			q.Done(element)
		}
	}()

	// 向队列中添加元素
	// Add elements to the queue
	_ = q.Add("hello")
	_ = q.Add("world")

	// 添加重复元素，队列不允许添加重复元素
	// Add duplicate elements, the queue does not allow adding duplicate elements
	_ = q.Add("hello")
	_ = q.Add("world")

	// 等待元素被执行
	// Wait for the elements to be executed
	time.Sleep(time.Second)

	// 停止队列
	// Stop the queue
	q.Stop()
}
