[English](./README.md) | 中文

<div align="center">
	<img src="assets/logo.png" alt="logo" width="500px">
</div>

# 简介

WorkQueue 是一个多功能、用户友好、高性能的 Go 工作队列。它支持多种队列类型，并且设计简单易扩展。您可以轻松编写新的队列类型并与 WorkQueue 一起使用。

# 队列类型

-   [x] 队列 (Queue)
-   [x] 简单队列 (Simple Queue)
-   [x] 延迟队列 (Delaying Queue)
-   [x] 优先级队列 (Priority Queue)
-   [x] 限速队列 (RateLimiting Queue)

# 优势

-   简单易用
-   无外部依赖
-   高性能
-   低内存占用
-   四叉堆 (quadruple heap)
-   支持回调函数

# 基准测试

## 1. STL

所有队列类型都基于 `Queue` 和 `Simple Queue`。

`Queue` 使用 `deque` 存储元素，并使用 `set` 跟踪队列的状态。它是**默认**的队列类型。

`Simple Queue` 也使用 `deque` 存储元素。它不跟踪元素状态，也不维护元素优先级。

`Delaying Queue` 和 `Priority Queue` 使用 `heap` 管理元素的过期时间和优先级。

### 1.1. 双端队列

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

### 1.2. 堆

```bash
$ go test -benchmem -run=^$ -bench ^Benchmark* github.com/shengyanli1982/workqueue/internal/stl/heap
goos: darwin
goarch: amd64
pkg: github.com/shengyanli1982/workqueue/internal/stl/heap
cpu: Intel(R) Xeon(R) CPU E5-4627 v2 @ 3.30GHz
BenchmarkHeap_Push-8   	 8891779	       138.2 ns/op	      84 B/op	       1 allocs/op
BenchmarkHeap_Pop-8    	13314109	       119.1 ns/op	       0 B/op	       0 allocs/op
```

### 1.3. 集合

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

## 2. 队列

与 [kubernetes/client-go](https://github.com/kubernetes/client-go) 的 workqueue 相比，WorkQueue 展示了更好的性能和更低的内存占用。

> [!NOTE]
> WorkQueue 中的所有队列类型都基于 `Queue` 实现，与 `kubernetes/client-go` 的 workqueue 相同。因此，所有队列类型的性能和内存占用与 `Queue` 相当。
>
> 为什么不与其他实现进行比较？我认为 workqueue 与其使用环境紧密相关，很难与其他解决方案进行比较。如果您有更好的想法，请告诉我。

在 WorkQueue 中，元素存储在 `deque` 中，并使用 `set` 跟踪队列的状态。类似地，在 `kubernetes/client-go` 的 workqueue 中，元素存储在 `slice` 中，并使用 `set` 跟踪状态。

虽然 `slice` 比 `deque` 更快，但为 `slice` 预分配内存可以提高性能。然而，扩展 `slice` 将导致元素复制，从而增加内存使用量。

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

# 安装

```bash
go get github.com/shengyanli1982/workqueue
```

# 快速入门

有关如何使用 WorkQueue 的更多示例，请参考 [examples](examples) 目录。

## 1. 队列 (Queue)

`Queue` 是一个简单的先进先出队列，用作项目中所有其他队列的基础。它维护了一个 `dirty` 集合和一个 `processing` 集合来跟踪队列的状态。使用 `Add` 方法将现有元素添加到队列时，不会重复添加。

> [!IMPORTANT]
> 需要注意的是，如果要再次将现有元素添加到队列中，必须先调用 `Done` 方法将元素标记为已完成。
>
> 在调用 `Get` 方法后，需要调用 `Done` 方法。不要忘记这一步。

### 创建

-   `NewQueue`：使用提供的 `QConfig` 选项创建一个队列。如果配置为 `nil`，将使用默认配置。
-   `DefaultQueue`：使用默认配置创建一个队列。等效于 `NewQueue(nil)`，并返回实现 `Interface` 接口的值。

### 配置

在创建队列时，`Queue` 有一些可设置的配置选项。

-   `WithCallback`：设置回调函数。

### 方法

-   `Add`：将元素添加到工作队列。如果元素已经在队列中，不会重复添加。
-   `Get`：从工作队列中获取一个元素。如果工作队列为空，它将 **`非阻塞`** 并立即返回。
-   `GetWithBlock`：从工作队列中获取一个元素。如果工作队列为空，它将 **`阻塞`** 并等待新元素的添加。
-   `GetValues`：返回工作队列中元素的快照。可以安全地迭代它们。
-   `Done`：将工作队列中的元素标记为已完成。如果元素不在工作队列中，不会将其标记为已完成。
-   `Len`：返回工作队列中的元素数量。
-   `Range`：对工作队列中的每个元素调用函数 `fn`。它会阻塞工作队列。
-   `Stop`：关闭工作队列并等待所有 goroutine 完成。
-   `IsClosed`：如果工作队列正在关闭，则返回 `true`。

### 示例

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

**执行结果**

```bash
$ go run demo.go
get element: hello
get element: world
```

## 2. 简单队列 (Simple Queue)

而 `Simple Queue` 是一个简化版的 `Queue`，它作为一个先进先出（FIFO）队列运行，不跟踪元素的状态。它没有 `dirty` 或 `processing` 集合来跟踪队列的状态。如果使用 `Add` 方法将一个已存在的元素添加到队列中，它将会被再次添加。

> [!TIP] >
>
> 因为 `Simple Queue` 不跟踪队列的状态，所以在调用 `Get` 方法后不需要调用 `Done` 方法。
>
> `Done` 方法提供了兼容性支持。

### 创建

-   `NewSimpleQueue`：使用提供的 `QConfig` 选项创建一个简单队列。如果配置为 `nil`，将使用默认配置。
-   `DefaultSimpleQueue`：使用默认配置创建一个简单队列。它等同于 `NewSimpleQueue(nil)`，并返回实现 `Interface` 接口的值。

### 配置

`Simple Queue` 有一些可以在创建队列时设置的配置选项。

-   `WithCallback`：设置回调函数。

### 方法

-   `Add`：将元素添加到工作队列中。如果元素已经在队列中，不会再次添加。
-   `Get`：从工作队列中获取一个元素。如果工作队列为空，它将 **`非阻塞`** 并立即返回。
-   `GetWithBlock`：从工作队列中获取一个元素。如果工作队列为空，它将 **`阻塞`** 并等待新元素的添加。
-   `GetValues`：返回工作队列中元素的快照。可以安全地迭代它们。
-   `Done`：在工作队列中标记一个元素为已完成。在 `Simple Queue` 中，此方法不执行任何操作，仅提供兼容性支持。
-   `Len`：返回工作队列中的元素数量。
-   `Range`：对工作队列中的每个元素调用函数 `fn`。它会阻塞工作队列。
-   `Stop`：关闭工作队列并等待所有 goroutine 完成。
-   `IsClosed`：如果工作队列正在关闭，则返回 `true`。

### 示例

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

**执行结果**

```bash
$ go run demo.go
get element: hello
get element: world
get element: hello
get element: world
```

## 3. 延迟队列 (Delaying Queue)

`延迟队列` 是一个支持延迟执行的队列。它基于 `队列` 实现，并使用 `堆` 来维护元素的过期时间。当你向队列中添加一个元素时，可以指定延迟时间，该元素将在指定的延迟时间后执行。

> [!IMPORTANT] > `延迟队列` 有一个 `goroutine` 用于同步当前时间以更新超时时间。这个 `goroutine` 不能被关闭或修改。
>
> 定时器的最小重新同步时间为 `500ms`。如果你将元素的延迟时间设置为小于 `500ms`，它将在 `500ms` 后被处理。

### 创建

-   `NewDelayingQueue`：创建一个延迟队列，并使用 `DelayingQConfig` 来设置配置选项。如果配置为 `nil`，将使用默认配置。

-   `DefaultDelayingQueue`：使用默认配置创建一个延迟队列。它等同于 `NewDelayingQueue(nil)`，但返回值实现了 `DelayingInterface` 接口。

### 配置

`延迟队列` 有一些配置选项，在创建队列时可以设置。

-   `WithCallback`：设置回调函数。

> [!NOTE]
> 避免将容量设置得太小，否则可能导致从 `堆` 中的元素无法添加到队列中。
>
> 在这种情况下，元素将被分配一个新的延迟时间 `1500ms`，并再次添加到 `堆` 中，导致更长的执行延迟。

### 方法

-   `AddAfter`：在指定的延迟时间后将元素添加到工作队列中。如果元素已经在队列中，则不会再次添加。

### 示例

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

**执行结果**

```bash
$ go run demo.go
get element: hello
get element: world
get element: delay 1 sec
get element: delay 2 sec
[workqueue] Queue: Is closed
```

## 4. 优先级队列 (Priority Queue)

而 `优先级队列` 是一种支持优先级执行的队列。它基于 `队列`，并使用 `堆` 来维护元素的优先级。当向队列中添加元素时，可以指定其优先级，元素将按照优先级进行执行。

> [!CAUTION]
> 在 `优先级队列` 中，需要一个时间窗口来对当前添加到队列中的元素进行排序。在这个时间窗口内的元素按照优先级升序排序。然而，即使两个时间窗口紧邻，两个窗口中的元素的顺序也不能保证按照优先级排序。
>
> 默认的窗口大小是 `500ms`，但在创建队列时可以设置它。
>
> -   避免将窗口大小设置得太小，这会导致队列频繁排序，影响性能。
> -   避免将窗口大小设置得太大，这会导致排序的元素等待很长时间，可能延迟它们的执行。

### 创建

-   `NewPriorityQueue`：使用`PriorityQConfig`创建一个优先级队列来设置配置选项。如果配置为`nil`，将使用默认配置。

-   `DefaultPriorityQueue`：使用默认配置创建一个优先级队列。它等同于`NewPriorityQueue(nil)`，但返回值实现了`PriorityInterface`接口。

### 配置

-   `WithCallback`：设置回调函数。
-   `WithWindow`：设置队列的排序窗口大小。默认为`500ms`。

### 方法

-   `AddWeight`：使用指定的优先级将元素添加到工作队列中。如果元素已经在队列中，将不会再次添加。

### 示例

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

**执行结果**

```bash
$ go run demo.go
get element: hello
get element: world
get element: priority: 1
get element: priority: 2
[workqueue] Queue: Is closed
```

## 5. 限速队列

`限速队列` 是一个支持限速执行的队列。它基于 `队列` 并使用 `堆` 来维护元素的过期时间。当向队列中添加元素时，可以指定限速，元素将按照限速执行。

> [!TIP]
> 默认的限速算法基于 `令牌桶` 算法。您可以通过实现 `RateLimiter` 接口来定义自己的限速算法。

### 创建

-   `NewRateLimitingQueue`：使用 `RateLimitingQConfig` 创建一个限速队列来设置配置选项。如果配置为 `nil`，将使用默认配置。
-   `DefaultRateLimitingQueue`：使用默认配置创建一个限速队列。它等同于 `NewRateLimitingQueue(nil)`，但返回值实现了 `RateLimitingInterface` 接口。

### 配置

-   `WithCallback`：设置回调函数。
-   `WithLimiter`：设置队列的限速器。默认为 `TokenBucketRateLimiter`。

### 方法

-   `AddLimited`：使用指定的限速将元素添加到工作队列中。如果元素已经在队列中，则不会再次添加。
-   `Forget`：忽略限速器中的元素，表示该元素不再受限制。
-   `NumLimitTimes`：返回元素被限制的次数。

### 示例

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

**执行结果**

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

# 特性

`WorkQueue` 的设计目标是易于扩展，允许您编写和使用自定义的队列类型。

## 回调函数

`WorkQueue` 支持在创建队列时指定的回调函数。这些回调函数在执行某些操作时被调用。

> [!TIP]
> 回调函数是可选的。您可以通过将它们设置为 `nil` 来创建一个没有指定任何回调函数的队列。

### 示例

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

### 参考

队列的回调函数非常灵活，可以根据需要进行扩展。

#### 队列 (Queue) / 简单队列 (Simple Queue)

-   `OnAdd`: 在将元素添加到队列时调用。
-   `OnGet`: 在从队列中检索元素时调用。
-   `OnDone`: 在完成元素处理时调用。

#### 延迟队列 (Delaying Queue)

-   `OnAddAfter`: 在将具有指定延迟时间的元素添加到延迟队列时调用。

#### 优先级队列 (Priority Queue)

-   `OnAddWeight`: 在将具有指定优先级的元素添加到优先级队列时调用。

#### 限速队列 (RateLimiting Queue)

-   `OnAddLimited`: 在将具有指定速率限制的元素添加到限速队列时调用。
-   `OnForget`: 在从限速队列中移除元素时调用。
-   `OnGetTimes`: 在检索元素在限速队列中被限制的次数时调用。

## 2. 使用自定义队列

`WorkQueue` 的设计目标是易于扩展，允许您通过实现 `Interface` 和 `Callback` 接口并引用 `QConfig` 来创建自己的队列类型。

`WorkQueue` 库提供了两种内置的队列类型：`Queue` 和 `Simple Queue`。您可以将它们作为参考来创建自己的自定义队列类型。

例如，您可以使用 `Simple Queue` 作为基础来创建一个 `Delaying Queue`。

### 示例

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

**执行结果**

```bash
$ go run demo.go
get element: hello
get element: world
get element: delay 1 sec
get element: delay 2 sec
[workqueue] Queue: Is closed
```

## 3. 限速器

限速器仅适用于“限速队列”，用于确定每个元素的速率限制。默认情况下，速率限制基于“令牌桶”算法。您可以通过实现“RateLimiter”接口来实现自己的速率限制算法。

### 示例

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

**执行结果**

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

# 感谢

-   [kubernetes/client-go](https://github.com/kubernetes/client-go)
-   [lxzan/memorycache](https://github.com/lxzan/memorycache)
