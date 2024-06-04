English | [中文](./README_CN.md)

<div align="center">
	<img src="assets/logo.png" alt="logo" width="500px">
</div>

[![Go Report Card](https://goreportcard.com/badge/github.com/shengyanli1982/workqueue/v2)](https://goreportcard.com/report/github.com/shengyanli1982/workqueue/v2)
[![Build Status](https://github.com/shengyanli1982/workqueue/actions/workflows/test.yaml/badge.svg)](https://github.com/shengyanli1982/workqueue/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/shengyanli1982/workqueue/v2.svg)](https://pkg.go.dev/github.com/shengyanli1982/workqueue/v2)

# Introduction

`WorkQueue` is a high-performance, thread-safe, and memory-efficient Go library for managing work queues. It offers a variety of queue implementations such as `Queue`, `DelayingQueue`, `PriorityQueue`, and `RateLimitingQueue`, each tailored for specific use cases and performance needs. The library's design is simple, user-friendly, and platform-independent, making it suitable for a broad spectrum of applications and environments.

After several iterations and real-world usage, we've gathered valuable user feedback and insights. This led to a complete redesign and optimization of WorkQueue's architecture and underlying code in the new version (v2), significantly enhancing its robustness, reliability, and security.

# Why Use WorkQueue(v2)

The WorkQueue(v2) has undergone a comprehensive architectural revamp, greatly improving its robustness and reliability. This redesign enables the library to manage demanding workloads with increased stability, making it ideal for both simple task queues and complex workflows. By utilizing advanced algorithms and optimized data structures, WorkQueue(v2) provides superior performance, efficiently managing larger task volumes, reducing latency, and increasing throughput.

WorkQueue(v2) offers a diverse set of queue implementations to cater to different needs, including standard task management, delayed execution, task prioritization, and rate-limited processing. This flexibility allows you to select the most suitable tool for your specific use case, ensuring optimal performance and functionality. With its cross-platform design, WorkQueue(v2) guarantees consistent behavior and performance across various operating systems, making it a versatile solution for different environments.

The development of WorkQueue(v2) has been heavily influenced by user feedback and real-world usage, resulting in a library that better meets the needs of its users. By addressing user-reported issues and incorporating feature requests, WorkQueue(v2) offers a more refined and user-centric experience.

Choosing WorkQueue(v2) for your application or project could be a great decision. :)

# Advantages

-   **User-Friendly**: The intuitive design ensures easy usage, allowing users of varying skill levels to quickly become proficient.

-   **No External Dependencies**: The system operates independently, without the need for additional software or libraries, reducing compatibility issues and simplifying deployment.

-   **High Performance**: The system is optimized for speed and efficiency, swiftly handling tasks to enhance productivity and scalability.

-   **Minimal Memory Usage**: The design utilizes minimal system resources, ensuring smooth operation even on devices with limited hardware capabilities, and freeing up memory for other applications.

-   **Thread-Safe**: The system supports multi-threading, allowing for concurrent operations without the risk of data corruption or interference, providing a stable environment for multiple users or processes.

-   **Supports Action Callback Functions**: The system can execute predefined functions in response to specific events, enhancing interactivity, customization, and responsiveness.

-   **Cross-Platform Compatibility**: The system operates seamlessly across different operating systems and devices, providing flexibility for diverse user environments.

# Installation

```bash
go get github.com/shengyanli1982/workqueue/v2
```

# Benchmark

The following benchmark results demonstrate the performance of the `WorkQueue` library.

## 1. STL

### 1.1. List

When a linked list undergoes data modifications, the primary changes occur in the pointers of the elements, rather than directly adding elements like dynamic arrays. Over extended periods, linked lists prove to be more memory efficient than dynamic arrays.

**Direct performance**

```bash
$ go test -benchmem -run=^$ -bench ^BenchmarkList* .
goos: darwin
goarch: amd64
pkg: github.com/shengyanli1982/workqueue/v2/internal/container/list
cpu: Intel(R) Xeon(R) CPU E5-2643 v2 @ 3.50GHz
BenchmarkList_PushBack-12        	186905107	         6.447 ns/op	       0 B/op	       0 allocs/op
BenchmarkList_PushFront-12       	157372052	         7.701 ns/op	       0 B/op	       0 allocs/op
BenchmarkList_PopBack-12         	179555846	         6.645 ns/op	       0 B/op	       0 allocs/op
BenchmarkList_PopFront-12        	180030582	         6.989 ns/op	       0 B/op	       0 allocs/op
BenchmarkList_InsertBefore-12    	189274771	         6.406 ns/op	       0 B/op	       0 allocs/op
BenchmarkList_InsertAfter-12     	160078981	         6.490 ns/op	       0 B/op	       0 allocs/op
BenchmarkList_Remove-12          	183250782	         6.440 ns/op	       0 B/op	       0 allocs/op
BenchmarkList_MoveToFront-12     	146021263	         7.837 ns/op	       0 B/op	       0 allocs/op
BenchmarkList_MoveToBack-12      	141336429	         8.589 ns/op	       0 B/op	       0 allocs/op
BenchmarkList_Swap-12            	100000000	         10.47 ns/op	       0 B/op	       0 allocs/op
```

**Compare with the standard library**

Both the standard library and this project employ the same algorithm, leading to comparable performance. However, the `list` in this project provides additional features compared to the standard library. Furthermore, the `list` node uses `sync.Pool` to minimize memory allocation. Therefore, under high concurrency, the performance of the project's `list` may surpass that of the standard library.

```bash
$ go test -benchmem -run=^$ -bench ^BenchmarkCompare* .
goos: darwin
goarch: amd64
pkg: github.com/shengyanli1982/workqueue/v2/internal/container/list
cpu: Intel(R) Xeon(R) CPU E5-2643 v2 @ 3.50GHz
BenchmarkCompareGoStdList_PushBack-12        	 8256513	       129.4 ns/op	      56 B/op	       1 allocs/op
BenchmarkCompareGoStdList_PushFront-12       	 9448060	       115.5 ns/op	      55 B/op	       1 allocs/op
BenchmarkCompareGoStdList_PopBack-12         	178923963	        23.60 ns/op	       0 B/op	       0 allocs/op
BenchmarkCompareGoStdList_PopFront-12        	33846044	        46.40 ns/op	       0 B/op	       0 allocs/op
BenchmarkCompareGoStdList_InsertBefore-12    	12046944	        93.53 ns/op	      55 B/op	       1 allocs/op
BenchmarkCompareGoStdList_InsertAfter-12     	11364718	        94.52 ns/op	      55 B/op	       1 allocs/op
BenchmarkCompareWQList_PushBack-12           	11582172	       109.7 ns/op	      55 B/op	       1 allocs/op
BenchmarkCompareWQList_PushFront-12          	10893723	        92.67 ns/op	      55 B/op	       1 allocs/op
BenchmarkCompareWQList_PopBack-12            	181593789	         6.841 ns/op	       0 B/op	       0 allocs/op
BenchmarkCompareWQList_PopFront-12           	179179370	         7.057 ns/op	       0 B/op	       0 allocs/op
BenchmarkCompareWQList_InsertBefore-12       	 9302694	       116.5 ns/op	      55 B/op	       1 allocs/op
BenchmarkCompareWQList_InsertAfter-12        	10237197	       117.7 ns/op	      55 B/op	       1 allocs/o
```

### 1.2. Heap

**Direct performance**

The project uses the `Insertion Sort` algorithm to sort elements in the heap. In a sorted array, the time complexity of the `Insertion Sort` algorithm is `O(n)`. In this project, a `list` is used to store the elements in the heap. Each element is appended to the end of the list and then sorted.

```bash
$ go test -benchmem -run=^$ -bench ^BenchmarkHeap* .
goos: darwin
goarch: amd64
pkg: github.com/shengyanli1982/workqueue/v2/internal/container/heap
cpu: Intel(R) Xeon(R) CPU E5-2643 v2 @ 3.50GHz
BenchmarkHeap_Push-12      	  112101	    	125531 ns/op	       0 B/op	       0 allocs/op
BenchmarkHeap_Pop-12       	159293402	        23.71 ns/op	       0 B/op	       0 allocs/op
BenchmarkHeap_Remove-12    	974444684	         1.271 ns/op	       0 B/op	       0 allocs/op
```

**Compare with the standard library**

The heap in this project uses the `Insertion Sort` algorithm for sorting elements, while the standard library uses the `container/heap` package to implement the heap. The time complexity of the standard library's sorting is `O(nlogn)`, while the project's sorting has a time complexity of `O(n^2)`. Therefore, the project's sorting is slower than the standard library's. However, this is due to the difference in the algorithms used, and thus, a direct comparison may not be fair.

> [!TIP]
>
> The `Insertion Sort` algorithm can provide a stable and consistent sorting, unlike the binary heap. If you have any better suggestions, please feel free to share.

```bash
$ go test -benchmem -run=^$ -bench ^BenchmarkCompare* .
goos: darwin
goarch: amd64
pkg: github.com/shengyanli1982/workqueue/v2/internal/container/heap
cpu: Intel(R) Xeon(R) CPU E5-2643 v2 @ 3.50GHz
BenchmarkCompareGoStdHeap_Push-12    	 4552110	       278.9 ns/op	      92 B/op	       1 allocs/op
BenchmarkCompareGoStdHeap_Pop-12     	 3726718	       362.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkCompareWQHeap_Push-12       	  109158	    121247 ns/op	      48 B/op	       1 allocs/op
BenchmarkCompareWQHeap_Pop-12        	174782917	        15.10 ns/op	       0 B/op	       0 allocs/op
```

### Struct Memory Alignment

In essence, memory alignment enhances performance, minimizes CPU cycles, reduces power usage, boosts stability, and ensures predictable behavior. This is why it's considered a best practice to align data in memory, especially on contemporary 64-bit CPUs.

```bash
Node struct alignment:

---- Fields in struct ----
+----+----------------+-----------+-----------+
| ID |   FIELDTYPE    | FIELDNAME | FIELDSIZE |
+----+----------------+-----------+-----------+
| A  | interface {}   | Value     | 16        |
| B  | unsafe.Pointer | parentRef | 8         |
| C  | int64          | Priority  | 8         |
| D  | *list.Node     | Next      | 8         |
| E  | *list.Node     | Prev      | 8         |
+----+----------------+-----------+-----------+
---- Memory layout ----
|A|A|A|A|A|A|A|A|
|A|A|A|A|A|A|A|A|
|B|B|B|B|B|B|B|B|
|C|C|C|C|C|C|C|C|
|D|D|D|D|D|D|D|D|
|E|E|E|E|E|E|E|E|

total cost: 48 Bytes.
```

## 2. Queues

Here are the benchmark results for all queues in the `WorkQueue` library.

> [!NOTE]
>
> The RateLimitingQueue is quite slow due to its use of bucket-based rate limiting. It's not recommended for high-performance scenarios.

```bash
$ go test -benchmem -run=^$ -bench ^Benchmark* .
goos: darwin
goarch: amd64
pkg: github.com/shengyanli1982/workqueue/v2
cpu: Intel(R) Xeon(R) CPU E5-2643 v2 @ 3.50GHz
BenchmarkDelayingQueue_Put-12                         	 4172398	       304.4 ns/op	      56 B/op	       1 allocs/op
BenchmarkDelayingQueue_PutWithDelay-12                	 2773111	       423.9 ns/op	      55 B/op	       1 allocs/op
BenchmarkDelayingQueue_Get-12                         	26794798	        46.85 ns/op	      20 B/op	       0 allocs/op
BenchmarkDelayingQueue_PutAndGet-12                   	17567817	        68.64 ns/op	       7 B/op	       0 allocs/op
BenchmarkDelayingQueue_PutWithDelayAndGet-12          	 3747397	       314.9 ns/op	      19 B/op	       1 allocs/op
BenchmarkPriorityQueue_Put-12                         	 4631265	       259.3 ns/op	      55 B/op	       1 allocs/op
BenchmarkPriorityQueue_PutWithPriority-12             	 4797620	       259.3 ns/op	      55 B/op	       1 allocs/op
BenchmarkPriorityQueue_Get-12                         	29222815	        43.84 ns/op	      18 B/op	       0 allocs/op
BenchmarkPriorityQueue_PutAndGet-12                   	16933688	        69.35 ns/op	       7 B/op	       0 allocs/op
BenchmarkPriorityQueue_PutWithPriorityAndGet-12       	17161538	        70.61 ns/op	       7 B/op	       0 allocs/op
BenchmarkQueue_Put-12                                 	 5023969	       248.2 ns/op	      55 B/op	       1 allocs/op
BenchmarkQueue_Get-12                                 	31441930	        40.20 ns/op	      17 B/op	       0 allocs/op
BenchmarkQueue_PutAndGet-12                           	18027499	        64.72 ns/op	       7 B/op	       0 allocs/op
BenchmarkQueue_Idempotent_Put-12                      	 1820281	       687.5 ns/op	     158 B/op	       3 allocs/op
BenchmarkQueue_Idempotent_Get-12                      	 2640146	       474.4 ns/op	      93 B/op	       0 allocs/op
BenchmarkQueue_Idempotent_PutAndGet-12                	 2825148	       438.6 ns/op	      69 B/op	       1 allocs/op
BenchmarkRateLimitingQueue_Put-12                     	 4836130	       256.6 ns/op	      56 B/op	       1 allocs/op
BenchmarkRateLimitingQueue_PutWithLimited-12          	 1000000	     13557 ns/op	     120 B/op	       2 allocs/op
BenchmarkRateLimitingQueue_Get-12                     	28820907	        44.27 ns/op	      18 B/op	       0 allocs/op
BenchmarkRateLimitingQueue_PutAndGet-12               	16928090	        74.94 ns/op	       7 B/op	       0 allocs/op
BenchmarkRateLimitingQueue_PutWithLimitedAndGet-12    	 1000000	     16531 ns/op	      77 B/op	       2 allocs/op
```

# Quick Start

For more examples on how to use WorkQueue, please refer to the `examples` directory.

## 1. Queue

The `Queue` is a simple FIFO (First In, First Out) queue that serves as the base for all other queues in this project. It maintains a `dirty` set and a `processing` set to keep track of the queue's state.

The `dirty` set contains items that have been added to the queue but have not yet been processed. The `processing` set contains items that are currently being processed.

> [!IMPORTANT]
>
> If you create a new queue with the `WithValueIdempotent` configuration, the queue will automatically remove duplicate items. This means that if you put the same item into the queue, the queue will only keep one instance of that item.
>
> However, this value (`PutXXX functions param`) refers to an object that can be hashed by the `map` in the `Go` standard library. If the object cannot be hashed, such as pointers or slices, the program may throw an error.

### Config

The `Queue` has several configuration options that can be set when creating a queue.

-   `WithCallback`: Sets callback functions.
-   `WithValueIdempotent`: Enables item idempotency for the queue.

### Methods

-   `Shutdown`: Terminates the queue, preventing it from accepting new tasks.
-   `IsClosed`: Checks if the queue is closed, returns a boolean.
-   `Len`: Returns the number of elements in the queue.
-   `Values`: Returns all elements in the queue as a slice.
-   `Range`: Iterates over all elements in the queue.
-   `Put`: Adds an element to the queue.
-   `Get`: Retrieves an element from the queue.
-   `Done`: Notifies the queue that an element has been processed.

> [!NOTE]
>
> The `Done` function is only used when the queue is created with the `WithValueIdempotent` option. If you don't use this option, you don't need to call this function.

### Callbacks

-   `OnPut`: Invoked when an item is added to the queue.
-   `OnGet`: Invoked when an item is retrieved from the queue.
-   `OnDone`: Invoked when an item has been processed.

### Example

```go
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

	// 创建一个新的队列
	// Create a new queue
	queue := wkq.NewQueue(nil)

	// 增加 WaitGroup 的计数器
	// Increase the counter of the WaitGroup
	wg.Add(1)

	// 启动一个新的 goroutine 来运行 comsumer 函数
	// Start a new goroutine to run the comsumer function
	go consumer(queue, &wg)

	// 将 "hello" 放入队列
	// Put "hello" into the queue
	_ = queue.Put("hello")

	// 将 "world" 放入队列
	// Put "world" into the queue
	_ = queue.Put("world")

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

```

**Result**

```bash
$ go run demo.go
> get element: hello
> get element: world
queue is shutting down
```

## 2. Delaying Queue

The `Delaying Queue` is a queue that supports delayed execution. It builds upon the `Queue` and uses a `Heap` to manage the expiration times of the elements. When you add an element to the queue, you can specify a delay time. The elements are then sorted by this delay time and executed after the specified delay has passed.

> [!TIP]
>
> When the `Delaying Queue` is empty in the `Heap` or the first element is not due, it will wait every `heartbeat` time for an element in the `Heap` that can be processed. This means that there may be a slight deviation in the actual delay time of the element. The actual delay time is the **"element delay time + 300ms"**.
>
> If precise timing is important for your project, you may consider using the `kairos` project I wrote.

### Configuration

The `Delaying Queue` inherits the configuration of the `Queue`.

-   `WithCallback`: Sets callback functions.

### Methods

The `Delaying Queue` inherits the methods of the `Queue`. Additionally, it introduces the following method:

-   `PutWithDelay`: Adds an element to the queue with a specified delay.
-   `HeapRange`: Iterates over all elements in the heap.

### Callbacks

The `Delaying Queue` inherits the callbacks of the `Queue`. Additionally, it introduces the following callbacks:

-   `OnDelay`: Invoked when an element is added to the queue with a specified delay.
-   `OnPullError`: Invoked when an error occurs while pulling an element from the heap to the queue.

### Example

```go
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

	// 创建一个新的队列
	// Create a new queue
	queue := wkq.NewDelayingQueue(nil)

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

```

**Result**

```bash
$ go run demo.go
> get element: hello
> get element: world
> get element: delay 2
> get element: delay 1
queue is shutting down
```

## 3. Priority Queue

The `Priority Queue` is a queue that supports prioritized execution. It is built on top of the `Queue` and uses a `Heap` to manage the priorities of the elements. When adding an element to the queue, you can specify its priority. The elements are then sorted and executed based on their priorities.

### Configuration

The `Priority Queue` inherits the configuration of the `Queue`.

-   `WithCallback`: Sets callback functions.

### Methods

The `Priority Queue` inherits the methods of the `Queue`. Additionally, it provides the following methods:

-   `PutWithPriority`: Adds an element to the queue with a specified priority.
-   `Put`: Adds an element to the queue with a default priority (`math.MinInt64`).

### Callbacks

The `Priority Queue` inherits the callbacks of the `Queue`. Additionally, it provides the following callback:

-   `OnPriority`: Invoked when an element is added to the queue with a specified priority.

### Example

```go
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

	// 创建一个新的队列
	// Create a new queue
	queue := wkq.NewPriorityQueue(nil)

	// 增加 WaitGroup 的计数器
	// Increase the counter of the WaitGroup
	wg.Add(1)

	// 启动一个新的 goroutine 来运行 consumer 函数
	// Start a new goroutine to run the consumer function
	go consumer(queue, &wg)

	// 将 "delay 1" 放入队列，并设置其优先级为 200
	// Put "delay 1" into the queue and set its priority to 200
	_ = queue.PutWithPriority("priority 1", 200)

	// 将 "delay 2" 放入队列，并设置其优先级为 100
	// Put "delay 2" into the queue and set its priority to 100
	_ = queue.PutWithPriority("priority 2", 100)

	// 将 "hello" 放入队列
	// Put "hello" into the queue
	_ = queue.Put("hello")

	// 将 "world" 放入队列
	// Put "world" into the queue
	_ = queue.Put("world")

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

```

**Result**

```bash
$ go run demo.go
> get element: hello
> get element: world
> get element: priority 2
> get element: priority 1
queue is shutting down
```

## 4. RateLimiting Queue

The `RateLimiting Queue` is a queue that supports rate-limited execution. It is built on top of the `Delaying Queue`. When adding an element to the queue, you can specify the rate limit, and the element will be processed according to this rate limit.

> [!TIP]
>
> The default rate limit is based on the `Nop` strategy. You can define your own rate limit algorithm by implementing the `Limiter` interface. The project provides a `token bucket` algorithm as a Limiter implementation.

### Config

The `RateLimiting Queue` inherits the configuration of the `Delaying Queue`.

-   `WithCallback`: Sets callback functions.
-   `WithLimiter`: Sets the rate limiter for the queue.

### Methods

The `RateLimiting Queue` inherits the methods of the `Delaying Queue`. Additionally, it has the following method:

-   `PutWithLimited`: Adds an element to the queue. The delay time of the element is determined by the limiter.

### Callback

The `RateLimiting Queue` inherits the callback of the `Delaying Queue`. Additionally, it has the following method:

-   `OnLimited`: Invoked when an element is added to the queue by `PutWithLimited`.

### Example

```go
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

```

**Result**

```bash
$ go run demo.go
> get element: hello
> get element: world
> get element: delay 2
> get element: delay 1
> get element: limited 9
> get element: limited 6
> get element: limited 7
> get element: limited 8
> get element: limited 3
> get element: limited 2
> get element: limited 0
> get element: limited 5
> get element: limited 1
> get element: limited 4
queue is shutting down
```
