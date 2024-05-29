[English](./README.md) | 中文

<div align="center">
	<img src="assets/logo.png" alt="logo" width="500px">
</div>

[![Go Report Card](https://goreportcard.com/badge/github.com/shengyanli1982/workqueue)](https://goreportcard.com/report/github.com/shengyanli1982/workqueue)
[![Build Status](https://github.com/shengyanli1982/workqueue/actions/workflows/test.yaml/badge.svg)](https://github.com/shengyanli1982/workqueue/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/shengyanli1982/workqueue.svg)](https://pkg.go.dev/github.com/shengyanli1982/workqueue)

# 简介

`WorkQueue` 是一个高性能、线程安全且内存高效的 Go 库，用于管理工作队列。它提供了各种队列实现，如 `Queue`、`DelayingQueue`、`PriorityQueue` 和 `RateLimitingQueue`，每种都针对特定的使用场景和性能需求进行了优化。该库的设计简单、用户友好且平台无关，使其适用于广泛的应用和环境。

经过多次迭代和实际使用，我们收集了宝贵的用户反馈和洞察。这导致了 WorkQueue 架构和底层代码在新版本（v2）中的全面重新设计和优化，显著提高了其健壮性、可靠性和安全性。

# 为什么使用 WorkQueue(v2)

WorkQueue(v2) 进行了全面的架构改造，大大提高了其健壮性和可靠性。这种重新设计使得库能够以更高的稳定性管理苛刻的工作负载，使其成为简单任务队列和复杂工作流的理想选择。通过使用先进的算法和优化的数据结构，WorkQueue(v2) 提供了卓越的性能，有效地管理更大的任务量，降低延迟，增加吞吐量。

WorkQueue(v2) 提供了多种队列实现以满足不同的需求，包括标准任务管理、延迟执行、任务优先级和速率限制处理。这种灵活性使您能够选择最适合您特定用例的工具，确保最佳的性能和功能。凭借其跨平台设计，WorkQueue(v2) 在各种操作系统上保证了一致的行为和性能，使其成为不同环境的通用解决方案。

WorkQueue(v2) 的开发受到了用户反馈和实际使用的深刻影响，从而产生了更好地满足用户需求的库。通过解决用户报告的问题和整合功能请求，WorkQueue(v2) 提供了更精细和以用户为中心的体验。

选择 WorkQueue(v2) 作为您的应用程序或项目可能是一个明智的决定。:)

# 优点

-   **用户友好**：直观的设计确保易于使用，使各种技能水平的用户能够快速熟练。

-   **无外部依赖**：系统独立运行，无需额外的软件或库，减少兼容性问题，简化部署。

-   **高性能**：系统针对速度和效率进行了优化，快速处理任务，提高生产力和可扩展性。

-   **最小内存使用**：设计利用最小的系统资源，即使在硬件能力有限的设备上也能顺畅运行，为其他应用程序释放内存。

-   **线程安全**：系统支持多线程，允许并发操作，无需担心数据损坏或干扰，为多用户或进程提供稳定的环境。

-   **支持动作回调函数**：系统可以在特定事件发生时执行预定义的函数，增强交互性、定制性和响应性。

-   **跨平台兼容性**：系统在不同的操作系统和设备上无缝运行，为多样化的用户环境提供灵活性。

# 安装

```bash
go get github.com/shengyanli1982/workqueue/v2
```

# 性能测试

以下的性能测试结果展示了 `WorkQueue` 库的性能表现。

## 1. 标准模板库 (STL)

### 1.1. 链表

当链表进行数据修改时，主要的变化发生在元素的指针上，而不是像动态数组那样直接添加元素。在长时间运行中，链表比动态数组更具有内存效率。

**裸性能**

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

**与标准库的比较**

标准库和本项目都使用了相同的算法，因此性能相近。然而，相比于标准库，本项目中的 `list` 提供了更多的功能。

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

### 1.2. 堆

**裸性能**

本项目使用 `插入排序` 算法对堆中的元素进行排序。在已排序的数组中，`插入排序` 算法的时间复杂度为 `O(n)`。在本项目中，使用 `list` 来存储堆中的元素。每个元素都被追加到列表的末尾，然后进行排序。

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

**与标准库的比较**

本项目中的堆使用 `插入排序` 算法对元素进行排序，而标准库使用 `container/heap` 包来实现堆。标准库的排序时间复杂度为 `O(nlogn)`，而本项目的排序时间复杂度为 `O(n^2)`。因此，本项目的排序速度比标准库的慢。然而，这是由于使用的算法不同，因此，直接比较可能并不公平。

> [!TIP]
>
> `插入排序` 算法可以提供稳定且一致的排序，不同于二叉堆。如果你有更好的建议，欢迎随时分享。

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

### 结构体内存对齐

本质上，内存对齐可以提高性能，减少 CPU 周期，降低功耗，增强稳定性，并确保行为的可预测性。这就是为什么在内存中对齐数据被视为最佳实践，特别是在现代的 64 位 CPU 上。

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

## 2. 队列性能

以下是 `WorkQueue` 库中所有队列的基准测试结果。

> [!NOTE]
>
> 由于 RateLimitingQueue 使用基于桶的速率限制，其速度相当慢。因此，不建议在高性能场景中使用。

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

# 快速开始

如需了解更多关于如何使用 WorkQueue 的示例，请参考 `examples` 目录。

## 1. 队列

`Queue` 是一个简单的 FIFO（先进先出）队列，是该项目中所有其他队列的基础。它维护一个 `dirty` 集合和一个 `processing` 集合来跟踪队列的状态。

`dirty` 集合包含已添加到队列但尚未处理的项。`processing` 集合包含当前正在处理的项。

> [!IMPORTANT]
>
> 如果你使用 `WithValueIdempotent` 配置创建新队列，队列将自动删除重复项。这意味着如果你将相同的项放入队列，队列将只保留该项的一个实例。

### 配置

创建队列时，`Queue` 有几个可以设置的配置选项。

-   `WithCallback`：设置回调函数。
-   `WithValueIdempotent`：为队列启用项幂等性。

### 方法

-   `Shutdown`：终止队列，防止其接受新任务。
-   `IsClosed`：检查队列是否已关闭，返回一个布尔值。
-   `Len`：返回队列中的元素数量。
-   `Values`：将队列中的所有元素作为切片返回。
-   `Put`：向队列添加元素。
-   `Get`：从队列中检索元素。
-   `Done`：通知队列一个元素已被处理。

> [!NOTE]
>
> `Done` 函数仅在使用 `WithValueIdempotent` 选项创建队列时使用。如果你不使用此选项，你不需要调用此函数。

### 回调

-   `OnPut`：当项被添加到队列时调用。
-   `OnGet`：当从队列检索项时调用。
-   `OnDone`：当项已被处理时调用。

### 示例

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

**输出结果**

```bash
$ go run demo.go
> get element: hello
> get element: world
queue is shutting down
```

## 2. 延迟队列

`Delaying Queue` 是一种支持延迟执行的队列。它基于 `Queue` 并使用 `Heap` 来管理元素的过期时间。当你向队列中添加一个元素时，你可以指定一个延迟时间。然后，元素会按照这个延迟时间进行排序，并在指定的延迟过后执行。

### 配置

`Delaying Queue` 继承了 `Queue` 的配置。

-   `WithCallback`：设置回调函数。

### 方法

`Delaying Queue` 继承了 `Queue` 的方法。此外，它还引入了以下方法：

-   `PutWithDelay`：向队列中添加一个元素，并指定一个延迟时间。

### 回调

`Delaying Queue` 继承了 `Queue` 的回调。此外，它还引入了以下回调：

-   `OnDelay`：当一个元素被添加到队列并指定了延迟时间时调用。
-   `OnPullError`：当从堆中拉取元素到队列时发生错误时调用。

### 示例

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

**输出结果**

```bash
$ go run demo.go
> get element: hello
> get element: world
> get element: delay 2
> get element: delay 1
queue is shutting down
```

## 3. 优先级队列

`Priority Queue` 是一种支持优先级执行的队列。它基于 `Queue` 构建，并使用 `Heap` 来管理元素的优先级。当向队列中添加元素时，可以指定其优先级。然后，元素会根据其优先级进行排序和执行。

### 配置

`Priority Queue` 继承了 `Queue` 的配置。

-   `WithCallback`：设置回调函数。

### 方法

`Priority Queue` 继承了 `Queue` 的方法。此外，它还提供以下方法：

-   `PutWithPriority`：以指定的优先级将元素添加到队列中。
-   `Put`：以默认优先级（`math.MinInt64`）将元素添加到队列中。

### 回调

`Priority Queue` 继承了 `Queue` 的回调。此外，它还提供以下回调：

-   `OnPriority`：当以指定的优先级将元素添加到队列时调用。

### 示例

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

**输出结果**

```bash
$ go run demo.go
> get element: hello
> get element: world
> get element: priority 2
> get element: priority 1
queue is shutting down
```

## 4. 限速队列

`RateLimiting Queue` 是一种支持限速执行的队列，它基于 `Delaying Queue` 构建。当向队列中添加元素时，你可以指定限速，元素将按照这个限速进行处理。

> [!TIP]
> 默认的限速基于 `令牌桶` 算法。你可以通过实现 `Limiter` 接口来定义自己的限速算法。

### 配置

`RateLimiting Queue` 继承了 `Delaying Queue` 的配置。

-   `WithCallback`：设置回调函数。
-   `WithLimiter`：为队列设置限速器。

### 方法

`RateLimiting Queue`继承了`Delaying Queue`的方法。此外，它还有以下方法：

-   `PutWithLimited`：将元素添加到队列中。元素的延迟时间由限速器决定。

### 回调

`RateLimiting Queue` 继承了 `Delaying Queue` 的回调。此外，它还有以下方法：

-   `OnLimited`：当通过 `PutWithLimited` 将元素添加到队列时调用。

### 示例

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

**输出结果**

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
