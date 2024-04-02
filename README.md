English | [中文](./README_CN.md)

<div align="center">
	<img src="assets/logo.png" alt="logo" width="500px">
</div>

[![Go Report Card](https://goreportcard.com/badge/github.com/shengyanli1982/workqueue)](https://goreportcard.com/report/github.com/shengyanli1982/workqueue)
[![Build Status](https://github.com/shengyanli1982/workqueue/actions/workflows/test.yaml/badge.svg)](https://github.com/shengyanli1982/workqueue/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/shengyanli1982/workqueue.svg)](https://pkg.go.dev/github.com/shengyanli1982/workqueue)

# Introduction

WorkQueue is a versatile, user-friendly, high-performance Go work queue. It supports multiple queue types and is designed for simplicity and easy extensibility. You can easily write a new queue type and use it with WorkQueue.

# Queue Types

-   [x] Queue
-   [x] Simple Queue
-   [x] Delaying Queue
-   [x] Priority Queue
-   [x] RateLimiting Queue

# Advantage

-   Simple and user-friendly
-   No external dependencies
-   High-performance
-   Low memory footprint
-   Utilizes a quadruple heap
-   Supports action callback functions

# Benchmark

## 1. STL

All Queue types are based on `Queue` and `Simple Queue`.

`Queue` uses `deque` to store elements and `set` to track the state of the queue. It is the **default** queue type.

`Simple Queue` also uses `deque` to store elements. It does not track element state or maintain element priority.

`Delaying Queue` and `Priority Queue` use `heap` to manage element expiration time and priority.

### 1.1. Deque

```bash
$ go test -benchmem -run=^$ -bench ^Benchmark* github.com/shengyanli1982/workqueue/internal/stl/deque
goos: darwin
goarch: amd64
pkg: github.com/shengyanli1982/workqueue/internal/stl/deque
cpu: Intel(R) Xeon(R) CPU E5-4627 v2 @ 3.30GHz
BenchmarkLink_Push-8        	11259825	        96.54 ns/op	      39 B/op	       1 allocs/op
BenchmarkLink_PushFront-8   	14764346	        92.25 ns/op	      39 B/op	       1 allocs/op
BenchmarkLink_Pop-8         	100000000	        18.21 ns/op	       0 B/op	       0 allocs/op
BenchmarkLink_PopBack-8     	250675488	         5.142 ns/op	       0 B/op	       0 allocs/op
```

### 1.2. Heap

```bash
$ go test -benchmem -run=^$ -bench ^Benchmark* github.com/shengyanli1982/workqueue/internal/stl/heap
goos: darwin
goarch: amd64
pkg: github.com/shengyanli1982/workqueue/internal/stl/heap
cpu: Intel(R) Xeon(R) CPU E5-4627 v2 @ 3.30GHz
BenchmarkHeap_Push-8   	 8891779	       138.2 ns/op	      84 B/op	       1 allocs/op
BenchmarkHeap_Pop-8    	13314109	       119.1 ns/op	       0 B/op	       0 allocs/op
```

### 1.3. Set

```bash
$ go test -benchmem -run=^$ -bench ^Benchmark* github.com/shengyanli1982/workqueue/internal/stl/set
goos: darwin
goarch: amd64
pkg: github.com/shengyanli1982/workqueue/internal/stl/set
cpu: Intel(R) Xeon(R) CPU E5-4627 v2 @ 3.30GHz
BenchmarkSet_Delete-8   	10340976	       145.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkSet_Insert-8   	 3968264	       400.3 ns/op	      94 B/op	       1 allocs/op
BenchmarkSet_Has-8      	11896728	       136.4 ns/op	       0 B/op	       0 allocs/op
```

## 2. Queue

When comparing WorkQueue to [kubernetes/client-go](https://github.com/kubernetes/client-go) workqueue, WorkQueue demonstrates better performance and lower memory usage.

> [!NOTE]
> All types of queues in WorkQueue are based on the `Queue` implementation, which is the same as the `kubernetes/client-go` workqueue. Therefore, the performance and memory usage of all queue types are equivalent to that of the `Queue`.
>
> Why not compare to other implementations? I believe workqueue is closely tied to its usage context, making it difficult to compare with other solutions. If you have any better ideas, please let me know.

In WorkQueue, elements are stored in a `deque` and the state of the queue is tracked using a `set`. Similarly, in the `kubernetes/client-go` workqueue, elements are stored in a `slice` and the state is tracked using a `set`.

While `slice` is faster than `deque`, preallocating memory for the slice can improve performance. However, expanding the slice will result in element copying, leading to increased memory usage.

```bash
$ go test -benchmem -run=^$ -bench ^Benchmark* github.com/shengyanli1982/workqueue/benchmark
goos: darwin
goarch: amd64
pkg: github.com/shengyanli1982/workqueue/benchmark
cpu: Intel(R) Xeon(R) CPU E5-4627 v2 @ 3.30GHz
BenchmarkClientgoAdd-8          	 2363143	       476.9 ns/op	     171 B/op	       1 allocs/op
BenchmarkClientgoGet-8          	 4164686	       316.3 ns/op	       7 B/op	       0 allocs/op
BenchmarkClientgoAddAndGet-8    	 2485596	       484.9 ns/op	      57 B/op	       1 allocs/op
BenchmarkWorkqueueAdd-8         	 1639712	       669.6 ns/op	      95 B/op	       2 allocs/op
BenchmarkWorkqueueGet-8         	 3914792	       327.0 ns/op	      25 B/op	       0 allocs/op
BenchmarkWorkqueueAddAndGet-8   	 1722157	       765.4 ns/op	      81 B/op	       1 allocs/op
```

# Installation

```bash
go get github.com/shengyanli1982/workqueue
```

# Quick Start

For more examples on how to use WorkQueue, you can refer to the [examples](examples) directory.

## 1. Queue

The `Queue` is a simple FIFO queue used as the foundation for all other queues in the project. It maintains a `dirty` set and a `processing` set to track the state of the queue. When adding an existing element to the queue using the `Add` method, it will not be added again.

> [!IMPORTANT]
> It is important to note that if you want to add an existing element to the queue again, you must first call the `Done` method to mark the element as done.
>
> After calling the `Get` method, it is required to call the `Done` method. Don't forget this step.

### Create

-   `NewQueue`: Creates a queue with the provided `QConfig` options. If the config is `nil`, the default config will be used.
-   `DefaultQueue`: Creates a queue with the default config. It is equivalent to `NewQueue(nil)` and returns a value that implements the `Interface` interface.

### Config

The `Queue` has some config options that can be set when creating a queue.

-   `WithCallback`: Sets callback functions.

### Methods

-   `Add`: Adds an element to the workqueue. If the element is already in the queue, it will not be added again.
-   `Get`: Gets an element from the workqueue. If the workqueue is empty, it will **`nonblock`** and return immediately.
-   `GetWithBlock`: Gets an element from the workqueue. If the workqueue is empty, it will **`block`** and wait for a new element to be added.
-   `GetValues`: Returns a snapshot of the elements in the workqueue. It is safe to iterate over them.
-   `Done`: Marks an element as done in the workqueue. If the element is not in the workqueue, it will not be marked as done.
-   `Len`: Returns the number of elements in the workqueue.
-   `Range`: Calls a function `fn` for each element in the workqueue. It blocks the workqueue.
-   `Stop`: Shuts down the workqueue and waits for all goroutines to finish.
-   `IsClosed`: Returns `true` if the workqueue is shutting down.

### Example

```go
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
```

**Result**

```bash
$ go run demo.go
get element: hello
get element: world
```

## 2. Simple Queue

## Simple Queue

`Simple Queue` is a simplified version of `Queue` that operates as a FIFO queue without tracking the state of elements. It does not have a `dirty` or `processing` set to track the state of the queue. If you add an existing element to the queue using the `Add` method, it will be added again.

> [!TIP]
> The `Simple Queue` does not track the state of the queue, so the `Done` method is not required after calling the `Get` method.
>
> The `Done` method is provided for compatibility purposes.

### Create

-   `NewSimpleQueue`: Creates a simple queue with the provided `QConfig` options. If the config is `nil`, the default config will be used.
-   `DefaultSimpleQueue`: Creates a simple queue with the default config. It is equivalent to `NewSimpleQueue(nil)` and returns a value that implements the `Interface` interface.

### Config

The `Simple Queue` has some config options that can be set when creating a queue.

-   `WithCallback`: Sets callback functions.

### Methods

-   `Add`: Adds an element to the workqueue. If the element is already in the queue, it will not be added again.
-   `Get`: Gets an element from the workqueue. If the workqueue is empty, it will **`nonblock`** and return immediately.
-   `GetWithBlock`: Gets an element from the workqueue. If the workqueue is empty, it will **`block`** and wait for a new element to be added.
-   `GetValues`: Returns a snapshot of the elements in the workqueue. It is safe to iterate over them.
-   `Done`: Marks an element as done in the workqueue. In the `Simple Queue`, this method does nothing and is only provided for compatibility.
-   `Len`: Returns the number of elements in the workqueue.
-   `Range`: Calls a function `fn` for each element in the workqueue. It blocks the workqueue.
-   `Stop`: Shuts down the workqueue and waits for all goroutines to finish.
-   `IsClosed`: Returns `true` if the workqueue is shutting down.

### Example

```go
package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

// "main" 函数是程序的入口点
// The "main" function is the entry point of the program
func main() {
	// 创建一个新的简单队列
	// Create a new simple queue
	q := workqueue.NewSimpleQueue(nil)

	// 启动一个新的 goroutine 来处理队列中的元素
	// Start a new goroutine to handle elements in the queue
	go func() {
		// 循环处理队列中的元素
		// Loop to handle elements in the queue
		for {
			// 从队列中获取一个元素
			// Get an element from the queue
			element, err := q.Get()

			// 如果获取元素时出错，则处理错误
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
	// Wait for the element to be executed
	time.Sleep(time.Second)

	// 停止队列
	// Stop the queue
	q.Stop()
}
```

**Result**

```bash
$ go run demo.go
get element: hello
get element: world
get element: hello
get element: world
```

## 3. Delaying Queue

The `Delaying Queue` is a queue that supports delaying execution. It is based on the `Queue` and uses a `heap` to maintain the expiration time of the elements. When you add an element to the queue, you can specify the delay time, and the element will be executed after the specified delay.

> [!IMPORTANT]
> The `Delaying Queue` has a `goroutine` that synchronizes the current time to update the timeout scale. This goroutine cannot be shut down or modified.
>
> The minimum resync time for the timer is `500ms`. If you set the delay time of an element to less than `500ms`, it will be processed after `500ms`.

### Create

-   `NewDelayingQueue`: Creates a delaying queue and uses `DelayingQConfig` to set configuration options. If the config is `nil`, the default config will be used.

-   `DefaultDelayingQueue`: Creates a delaying queue with the default config. It is equivalent to `NewDelayingQueue(nil)`, but the return value implements the `DelayingInterface` interface.

### Config

The `Delaying Queue` has some configuration options that can be set when creating a queue.

-   `WithCallback`: Sets callback functions.

> [!NOTE]
> Avoid setting the capacity too small, as it may cause the elements from the `heap` to fail to be added to the queue.
>
> In such cases, the element will be assigned a new delay time of `1500ms` and added to the `heap` again, resulting in a longer execution delay.

### Methods

-   `AddAfter`: Adds an element to the workqueue after the specified delay time. If the element is already in the queue, it will not be added again.

### Example

```go
package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

func main() {
	// 创建一个新的延迟队列
	// Create a new delaying queue
	q := workqueue.NewDelayingQueue(nil)

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

	// 向队列中添加延迟元素
	// Add delayed elements to the queue
	_ = q.AddAfter("delay 1 sec", time.Second)
	_ = q.AddAfter("delay 2 sec", time.Second*2)

	// 等待元素被执行
	// Wait for the elements to be executed
	time.Sleep(time.Second * 4)

	// 停止队列
	// Stop the queue
	q.Stop()
}
```

**Result**

```bash
$ go run demo.go
get element: hello
get element: world
get element: delay 1 sec
get element: delay 2 sec
[workqueue] Queue: Is closed
```

## 4. Priority Queue

## Priority Queue

The `Priority Queue` is a queue that supports priority execution. It is based on the `Queue` and uses a `heap` to maintain the priority of the elements. When adding an element to the queue, you can specify its priority, and the element will be executed according to the priority.

> [!CAUTION]
> The `Priority Queue` requires a time window to sort the elements currently added to the queue. The elements within this time window are sorted in ascending order of priority. However, the order of elements in two different time windows is not guaranteed to be sorted by priority, even if the two windows are immediately adjacent.
>
> The default window size is `500ms`, but you can set it when creating a queue.
>
> -   Avoid setting the window size too small, as it will cause frequent sorting of the queue, impacting performance.
> -   Avoid setting the window size too large, as it will cause sorted elements to wait for a long time, potentially delaying their execution.

### Create

-   `NewPriorityQueue`: Creates a priority queue using the `PriorityQConfig` to set configuration options. If the config is `nil`, the default config will be used.

-   `DefaultPriorityQueue`: Creates a priority queue with the default config. It is equivalent to `NewPriorityQueue(nil)`, but the return value implements the `PriorityInterface` interface.

### Config

-   `WithCallback`: Sets callback functions.
-   `WithWindow`: Sets the sort window size of the queue. The default is `500ms`.

### Methods

-   `AddWeight`: Adds an element to the workqueue with the specified priority. If the element is already in the queue, it will not be added again.

### Example

```go
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
```

**Result**

```bash
$ go run demo.go
get element: hello
get element: world
get element: priority: 1
get element: priority: 2
[workqueue] Queue: Is closed
```

## 5. RateLimiting Queue

The `RateLimiting Queue` is a queue that supports rate limiting execution. It is based on the `Queue` and uses a `heap` to maintain the expiration time of the elements. When adding an element to the queue, you can specify the rate limit, and the element will be executed according to the rate limit.

> [!TIP]
> The default rate limit is based on the `token bucket` algorithm. You can define your own rate limit algorithm by implementing the `RateLimiter` interface.

### Create

-   `NewRateLimitingQueue`: Creates a rate limiting queue using the `RateLimitingQConfig` to set configuration options. If the config is `nil`, the default config will be used.
-   `DefaultRateLimitingQueue`: Creates a rate limiting queue with the default config. It is equivalent to `NewRateLimitingQueue(nil)`, but the return value implements the `RateLimitingInterface` interface.

### Config

-   `WithCallback`: Sets callback functions.
-   `WithLimiter`: Sets the rate limiter of the queue. The default is `TokenBucketRateLimiter`.

### Methods

-   `AddLimited`: Adds an element to the workqueue with the specified rate limit. If the element is already in the queue, it will not be added again.
-   `Forget`: Forgets about an element in the rate limiter, which means the element is no longer limited.
-   `NumLimitTimes`: Returns the number of times an element has been limited.

### Example

```go
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
```

**Result**

```bash
$ go run demo.go
get element: hello
get element: world
get element: >>> 2024-02-03 15:04:33.111294 +0800 CST m=+0.000177452 0
get element: >>> 2024-02-03 15:04:33.111461 +0800 CST m=+0.000344599 1
get element: >>> 2024-02-03 15:04:33.111464 +0800 CST m=+0.000347704 2
get element: >>> 2024-02-03 15:04:33.111466 +0800 CST m=+0.000350012 3
get element: >>> 2024-02-03 15:04:33.111468 +0800 CST m=+0.000352241 4
get element: >>> 2024-02-03 15:04:33.111471 +0800 CST m=+0.000354489 5
get element: >>> 2024-02-03 15:04:33.111473 +0800 CST m=+0.000356461 6
get element: >>> 2024-02-03 15:04:33.111475 +0800 CST m=+0.000358548 7
get element: >>> 2024-02-03 15:04:33.111477 +0800 CST m=+0.000360545 8
get element: >>> 2024-02-03 15:04:33.111479 +0800 CST m=+0.000362747 9
[workqueue] Queue: Is closed
```

# Features

`WorkQueue` is designed to be easily extensible, allowing you to write and use custom queue types.

## Callbacks

`WorkQueue` supports callback functions that can be specified when creating a queue. These callbacks are invoked when performing certain actions.

> [!TIP]
> Callback functions are optional. You can create a queue without specifying any callbacks by setting them to `nil`.

### Example

```go
import (
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

// 定义一个回调结构体
// Define a callback struct
type callback struct {}

// "OnAdd" 方法会在元素被添加到队列时调用
// The "OnAdd" method will be called when an element is added to the queue
func (c *callback) OnAdd(element interface{}) {
	fmt.Println("add element:", element)
}

// "OnGet" 方法会在从队列获取元素时调用
// The "OnGet" method will be called when an element is obtained from the queue
func (c *callback) OnGet(element interface{}) {
	fmt.Println("get element:", element)
}

// "OnDone" 方法会在队列中的元素处理完成时调用
// The "OnDone" method will be called when an element in the queue is done
func (c *callback) OnDone(element interface{}) {
	fmt.Println("done element:", element)
}

// "main" 函数是程序的入口点
// The "main" function is the entry point of the program
func main() {
	// 创建一个新的队列配置
	// Create a new queue configuration
	conf := workqueue.NewQConfig()

	// 设置回调函数
	// Set the callback functions
	conf.WithCallback(&callback{})

	// 使用配置创建一个新的队列
	// Create a new queue with the configuration
	q := workqueue.NewQueue(conf)

	// 启动一个新的 goroutine 来处理队列中的元素
	// Start a new goroutine to handle elements in the queue
	go func() {
		// 循环处理队列中的元素
		// Loop to handle elements in the queue
		for {
			// 从队列中获取一个元素
			// Get an element from the queue
			element, err := q.Get()

			// 如果获取元素时出错，则打印错误并返回
			// If an error occurs when getting the element, print the error and return
			if err != nil {
				fmt.Println(err)
				return
			}

			// 标记元素为已处理
			// Mark the element as done
			q.Done(element)
		}
	}()

	// 向队列中添加元素
	// Add elements to the queue
	_ = q.Add("hello")
	_ = q.Add("world")

	// 等待元素被执行
	// Wait for the element to be executed
	time.Sleep(time.Second * 2)

	// 停止队列
	// Stop the queue
	q.Stop()
}
```

### Reference

The queue callback functions are flexible and can be easily extended to suit your needs.

#### Queue / Simple Queue

-   `OnAdd`: Called when adding an element to the queue.
-   `OnGet`: Called when retrieving an element from the queue.
-   `OnDone`: Called when completing processing of an element.

#### Delaying Queue

-   `OnAddAfter`: Called when adding an element with a specified delay time to the delaying queue.

#### Priority Queue

-   `OnAddWeight`: Called when adding an element with a specified priority to the priority queue.

#### RateLimiting Queue

-   `OnAddLimited`: Called when adding an element with a specified rate limit to the rate limiting queue.
-   `OnForget`: Called when removing an element from the rate limiting queue.
-   `OnGetTimes`: Called when retrieving the number of times an element has been limited in the rate limiting queue.

## 2. With Custom Queue

`WorkQueue` is designed to be easily extensible, allowing you to create your own queue types by implementing the `Interface` and `Callback` interfaces and referencing the `QConfig`.

The `WorkQueue` library provides two built-in queue types: `Queue` and `Simple Queue`. You can use them as a reference to create your own custom queue types.

For example, you can use the `Simple Queue` as a base to create a `Delaying Queue`.

### Example

```go
// 导入需要的包
// Import the required packages
import (
	"errors"
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

// "main" 函数是程序的入口点
// The "main" function is the entry point of the program
func main() {
	// 创建一个新的延迟队列
	// Create a new delaying queue
	q := workqueue.NewDelayingQueueWithCustomQueue(nil, workqueue.NewSimpleQueue(nil))

	// 启动一个新的 goroutine 来处理队列中的元素
	// Start a new goroutine to handle elements in the queue
	go func() {
		// 循环处理队列中的元素
		// Loop to handle elements in the queue
		for {
			// 从队列中获取一个元素
			// Get an element from the queue
			element, err := q.Get()

			// 如果获取元素时出错，则处理错误
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

	// 向队列中添加延迟元素
	// Add delayed elements to the queue
	_ = q.AddAfter("delay 1 sec", time.Second)
	_ = q.AddAfter("delay 2 sec", time.Second*2)

	// 等待元素被执行
	// Wait for the element to be executed
	time.Sleep(time.Second * 4)

	// 停止队列
	// Stop the queue
	q.Stop()
}
```

**Result**

```bash
$ go run demo.go
get element: hello
get element: world
get element: delay 1 sec
get element: delay 2 sec
[workqueue] Queue: Is closed
```

## 3. Limiter

The limiter only works for the `RateLimiting Queue` and determines the rate limit for each element. By default, the rate limit is based on the `token bucket` algorithm. You can implement your own rate limit algorithm by implementing the `RateLimiter` interface.

### Example

```go
// 导入需要的包
// Import the required packages
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
	// 设置限速器
	// Set the limiter
	conf.WithLimiter(workqueue.NewBucketRateLimiter(float64(4), 1))

	// 使用配置创建一个新的限速队列
	// Create a new rate limiting queue with the configuration
	q := workqueue.NewRateLimitingQueue(conf)

	// 启动一个新的 goroutine 来处理队列中的元素
	// Start a new goroutine to handle elements in the queue
	go func() {
		// 循环处理队列中的元素
		// Loop to handle elements in the queue
		for {
			// 从队列中获取一个元素
			// Get an element from the queue
			element, err := q.Get()

			// 如果获取元素时出错，则处理错误
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
	// Add limited elements to the queue
	for i := 0; i < 10; i++ {
		_ = q.AddLimited(fmt.Sprintf(">>> %d", i))
	}

	// 等待元素被执行
	// Wait for the element to be executed
	time.Sleep(time.Second * 3)

	// 停止队列
	// Stop the queue
	q.Stop()
}
```

**Result**

```bash
$ go run demo.go
[14:13] get element: hello
[14:13] get element: world
[14:13] get element: >>> 0
[14:13] get element: >>> 1
[14:13] get element: >>> 2
[14:14] get element: >>> 3
[14:14] get element: >>> 4
[14:14] get element: >>> 5
[14:14] get element: >>> 6
[14:15] get element: >>> 7
[14:15] get element: >>> 8
[14:15] get element: >>> 9
[workqueue] Queue: Is closed
```

# Thanks to

-   [kubernetes/client-go](https://github.com/kubernetes/client-go)
-   [lxzan/memorycache](https://github.com/lxzan/memorycache)
