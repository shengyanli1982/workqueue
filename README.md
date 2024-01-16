<div align="center">
	<h1>WorkQueue</h1>
	<img src="assets/logo.png" alt="logo" width="300px">
</div>

# Introduction

WorkQueue is a simple, fast, reliable work queue written in Go. It supports multiple queue types and is designed to be easily extensible which mean you can easily write a new queue type and use it with WorkQueue.

# Queue Types

-   [x] Queue
-   [x] Simple Queue
-   [x] Delaying Queue
-   [x] Priority Queue
-   [x] RateLimiting Queue

# Advantage

-   Simple and easy to use
-   No third-party dependencies
-   High performance
-   Low memory usage
-   Use `quadruple heap`
-   Support action callback functions

# Benchmark

## 1. STL

All Queue types can be based on `Queue` and `Simple Queue`.

`Queue` use `deque` to store elements and use `set` to track the state of the queue. `Queue` is **default**.

`Simple Queue` also use `deque` to store elements. No `set` be used, so no element state is tracked and no element priority is maintained.

`Delaying Queue` and `Priority Queue` will use `heap` to maintain the expiration time and priority of the element.

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

Compare to [kubernetes/client-go](https://github.com/kubernetes/client-go) workqueue, WorkQueue has better performance and lower memory usage.

> [!NOTE]
> All types of queues are based on `Queue` which the same as `kubernetes/client-go` workqueue. So the performance and memory usage of all types of queues are the same as `Queue`.
>
> Why not compare to others? I think workqueue too close to the process of use and it is difficult to compare. If you have a better idea, please let me know.

`WorkQueue` is same as `kubernetes/client-go` workqueue, no used `channel`, `WorkQueue` use `deque` to store elements and use `set` to track the state of the queue. `kubernetes/client-go` workqueue use `slice` to store elements and use `set` to track the state of the queue.

`slice` is faster than `deque`, preallocate memory for slice will improve performance.

But `slice` will cause the element to be copied when the slice is expanded, **which will cause the memory usage to increase**.

```bash
$ go test -benchmem -run=^$ -bench ^Benchmark* github.com/shengyanli1982/workqueue/bennchmark
goos: darwin
goarch: amd64
pkg: github.com/shengyanli1982/workqueue/bennchmark
cpu: Intel(R) Xeon(R) CPU E5-4627 v2 @ 3.30GHz
BenchmarkClientgoAdd-8          	 2439475	       463.0 ns/op	     166 B/op	       1 allocs/op
BenchmarkClientgoGet-8          	 4138384	       314.1 ns/op	       7 B/op	       0 allocs/op
BenchmarkClientgoAddAndGet-8    	 2355006	       488.5 ns/op	      57 B/op	       1 allocs/op
BenchmarkWorkqueueAdd-8         	 1585618	       677.8 ns/op	      96 B/op	       2 allocs/op
BenchmarkWorkqueueGet-8         	 4445774	       293.8 ns/op	       7 B/op	       0 allocs/op
BenchmarkWorkqueueAddAndGet-8   	 1734568	       764.4 ns/op	      90 B/op	       2 allocs/op
```

# Installation

```bash
go get github.com/shengyanli1982/workqueue
```

# Quick Start

Here are some examples of how to use WorkQueue. but you can also refer to the [examples](examples) directory for more examples.

## 1. Queue

`Queue` is a simple queue in project, all queues are based on it. It is a FIFO queue and has `dirty` and `processing` set to track the state of the queue. If you want to `Add` an exist element to the queue, unfortunately, it will not be added to the queue again.

> [!IMPORTANT]
> Here is an very important thing to note, if you want to add exist one to the queue again, you must call `Done` method to mark the element as done.
>
> `Done` method is required after `Get` method, don't forget it.

### Create

-   `NewQueue` create a queue, use `QConfig` to set config options. If the config is `nil`, the default config will be used.
-   `DefaultQueue` create a queue with default config. It is equivalent to `NewQueue(nil)`, but return value implementing the `Interface` interface.

### Config

The `Queue` has some config options, you can set it when create a queue.

-   `WithCallback` set callback functions

### Methods

-   `Add` adds an element to the workqueue. If the element is already in the queue, it will not be added again.
-   `Get` gets an element from the workqueue. If the workqueue is empty, it will **`nonblock`** and return immediately.
-   `GetWithBlock` gets an element from the workqueue. If the workqueue is empty, it will **`blocking`** and waiting new element be added into queue.
-   `GetValues` returns the elements of the workqueue. elements are `snapshot` of the workqueue, so it is safe to iterate over them.
-   `Done` marks an element as done with the workqueue. If the element is not in the workqueue, it will not be marked as done.
-   `Len` returns the elements count of the workqueue.
-   `Range` calls `fn` for each element in the workqueue. elements are `current` of the workqueue, so it block the workqueue.
-   `Stop` shuts down the workqueue and waits for all the goroutines to finish.
-   `IsClosed` returns true if the workqueue is shutting down.

### Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

func main() {
	q := workqueue.NewQueue(nil) // create a queue

	go func() {
		for {
			element, err := q.Get() // get element from queue
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("get element:", element)
			q.Done(element) // mark element as done, 'Done' is required after 'Get'
		}
	}()

	_ = q.Add("hello") // add element to queue
	_ = q.Add("world")

	time.Sleep(time.Second * 2) // wait for element to be executed

	q.Stop()
}
```

## 2. Simple Queue

`Simple Queue` is a simple queue that does not track the state of the element. It is simple version of `Queue`, which mean it is a FIFO queue and has no `dirty` and `processing` set to track the state of the queue. If you want to `Add` an exist element to the queue, it will be added to the queue again.

> [!TIP]
> The `Simple Queue` have no `dirty` and `processing` set to track the state of the queue, so `Done` method is not required after `Get` method.
>
> The `Done` method is left for compatibility

### Create

-   `NewSimpleQueue` create a simple queue, use `QConfig` to set config options. If the config is `nil`, the default config will be used.
-   `DefaultSimpleQueue` create a simple queue with default config. It is equivalent to `NewSimpleQueue(nil)`, but return value implementing the `Interface` interface.

### Config

The `Queue` has some config options, you can set it when create a queue.

-   `WithCallback` set callback functions

### Methods

-   `Add` adds an element to the workqueue. If the element is already in the queue, it will not be added again.
-   `Get` gets an element from the workqueue. If the workqueue is empty, it will **`nonblock`** and return immediately.
-   `GetWithBlock` gets an element from the workqueue. If the workqueue is empty, it will **`blocking`** and waiting new element be added into queue.
-   `GetValues` returns the elements of the workqueue. elements are `snapshot` of the workqueue, so it is safe to iterate over them.
-   `Done` marks an element as done with the workqueue. In fact in `Simple Queue`, it does nothing. Only left for compatibility.
-   `Len` returns the elements count of the workqueue.
-   `Range` calls `fn` for each element in the workqueue. elements are `current` of the workqueue, so it block the workqueue.
-   `Stop` shuts down the workqueue and waits for all the goroutines to finish.
-   `IsClosed` returns true if the workqueue is shutting down.

### Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

func main() {
	q := workqueue.NewSimpleQueue(nil)

	go func() {
		for {
			element, err := q.Get()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("get element:", element)
			q.Done(element) // mark element as done, 'Done' is not required after 'Get'
		}
	}()

	_ = q.Add("hello")
	_ = q.Add("world")

	time.Sleep(time.Second * 2)

	q.Stop()
}
```

## 3. Delaying Queue

`Delaying Queue` is a queue that supports delaying execution. It is based on `Queue` and uses a `heap` to maintain the expiration time of the element. When you add an element to the queue, you can specify the delay time, and the element will be executed after the delay time.

> [!IMPORTANT]
> The `Delaying Queue` has a `goroutine` that is sync the current time, used to update timeout scale. It can not be shut down and modified.
>
> Timer minimum resync time is `500ms`, which mean if you set the element's delay time less than `500ms`, it will be processed after `500ms`.

### Create

-   `NewDelayingQueue` create a delaying queue, use `DelayingQConfig` to set config options. If the config is `nil`, the default config will be used.

-   `DefaultDelayingQueue` create a delaying queue with default config. It is equivalent to `NewDelayingQueue(nil)`, but return value implementing the `DelayingInterface` interface.

### Config

The `Delaying Queue` has some config options, you can set it when create a queue.

-   `WithCallback` set callback functions

> [!NOTE]
> Don't set the capacity too small, it will cause the element from `heap` to be added to Queue failed.
>
> Then the element will be set new delay time(`1500ms`) and added to `heap` again, which will cause the element to be executed after a long time.

### Methods

-   `AddAfter` adds an element to the workqueue after the specified delay time. If the element is already in the queue, it will not be added again.

### Example

```go
package main

import (
	"fmt"
	"errors"
	"time"

	"github.com/shengyanli1982/workqueue"
)

func main() {
	q := workqueue.NewDelayingQueue(nil)

	go func() {
		for {
			element, err := q.Get()
			if err != nil && !errors.Is(err, workqueue.ErrorQueueEmpty) {
				fmt.Println(err)
				return
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
```

## 4. Priority Queue

`Priority Queue` is a queue that supports priority execution. It is based on `Queue` and uses a `heap` to maintain the priority of the element. When you add an element to the queue, you can specify the priority of the element, and the element will be executed according to the priority.

> [!CAUTION]
> The `Priority Queue` requires a window to sort the elements currently added to the Queue. The elements in this time window are sorted in order of `priority` from smallest to largestl The order of elements in two different time Windows is not guaranteed to be sorted by `priority`, even if the two Windows are immediately adjacent.
>
> The default window size is `500ms`, you can set it when create a queue.
>
> -   Dont't set the window size too small, it will cause the queue to be sorted frequently, which will affect the performance of the queue.
> -   Dont't set the window size too large, it will cause the elements sorted to wait for a long time, which will affect elements to be executed in time.

### Create

-   `NewPriorityQueue` create a priority queue, use `PriorityQConfig` to set config options. If the config is `nil`, the default config will be used.

-   `DefaultPriorityQueue` create a priority queue with default config. It is equivalent to `NewPriorityQueue(nil)`, but return value implementing the `PriorityInterface` interface.

### Config

-   `WithCallback` set callback functions
-   `WithWindow` set the sort window size of the queue, default is `500` ms.

### Methods

-   `AddWeight` adds an element to the workqueue with the specified priority. If the element is already in the queue, it will not be added again.

### Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

func main() {
	q := workqueue.NewPriorityQueue(nil)

	go func() {
		for {
			element, shutdown := q.Get()
			if shutdown {
				fmt.Println("shutdown")
				return
			}
			fmt.Println("get element:", element)
			q.Done(element) // mark element as done, 'Done' is required after 'Get'
		}
	}()

	_ = q.Add("hello")
	_ = q.Add("world")
	_ = q.AddWeight("priority: 1", 1) // add element with priority
	_ = q.AddWeight("priority: 2", 2)

	time.Sleep(time.Second * 2) // wait for element to be executed

	q.Stop()
}
```

## 5. RateLimiting Queue

`RateLimiting Queue` is a queue that supports rate limiting execution. It is based on `Queue` and uses a `heap` to maintain the expiration time of the element. When you add an element to the queue, you can specify the rate limit of the element, and the element will be executed according to the rate limit.

> [!TIP]
> Default rate limit is based on the `token bucket` algorithm. You can define your own rate limit algorithm by implementing the `RateLimiter` interface.

### Create

-   `NewRateLimitingQueue` create a rate limiting queue, use `RateLimitingQConfig` to set config options. If the config is `nil`, the default config will be used.
-   `DefaultRateLimitingQueue` create a rate limiting queue with default config. It is equivalent to `NewRateLimitingQueue(nil)`, but return value implementing the `RateLimitingInterface` interface.

### Config

-   `WithCallback` set callback functions
-   `WithLimiter` set the rate limiter of the queue, default is `TokenBucketRateLimiter`.

### Methods

-   `AddLimited` adds an element to the workqueue with the specified rate limit. If the element is already in the queue, it will not be added again.
-   `Forget` forgets about an element in the rate limiter. Which mean the element not limit anymore.
-   `NumLimitTimes` returns the number of times an element has been limited.

### Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

func main() {
	q := workqueue.NewRateLimitingQueue(nil)

	go func() {
		for {
			element, shutdown := q.Get()
			if shutdown {
				fmt.Println("shutdown")
				return
			}
			fmt.Println("get element:", element)
			q.Done(element) // mark element as done, 'Done' is required after 'Get'
		}
	}()

	_ = q.AddLimited("hello", time.Second)
	_ = q.AddLimited("world", time.Second)

	time.Sleep(time.Second * 2) // wait for element to be executed

	q.Stop()
}

```

# Features

`WorkQueue` also has interesting properties. It is designed to be easily extensible which mean you can easily write a new queue type and use it with WorkQueue.

## 1. Callback

`WorkQueue` supports action callback function. Specify a callback functions when create a queue, and the callback function will be called when do some action.

> [!TIP]
> Callback functions is not required that you can use `WorkQueue` without callback functions. Set `nil` when create a queue, and the callback function will not be called.

### Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

type callback struct {}

func (c *callback) OnAdd(element interface{}) { // OnAdd will be called when add an element to the queue
	fmt.Println("add element:", element)
}

func (c *callback) OnGet(element interface{}) { // OnGet will be called when get an element from the queue
	fmt.Println("get element:", element)
}

func (c *callback) OnDone(element interface{}) { // OnDone will be called when done an element from the queue
	fmt.Println("done element:", element)
}

func main() {
	conf := workqueue.NewQConfig()
	conf.WithCallback(&callback{}) // set callback functions

	q := workqueue.NewQueue(conf)

	go func() {
		for {
			element, err := q.Get()
			if err != nil {
				fmt.Println(err)
				return
			}
			q.Done(element)
		}
	}()

	_ = q.Add("hello")
	_ = q.Add("world")

	time.Sleep(time.Second * 2) // wait for element to be executed

	q.Stop()
}
```

### Reference

The queue callback functions are loosely used and can be easily extended, you can use it as you like.

#### Queue / Simple Queue

-   `OnAdd` will be called when add an element to the queue
-   `OnGet` will be called when get an element from the queue
-   `OnDone` will be called when done an element from the queue

#### Delaying Queue

-   `OnAddAfter` will be called when add an specified delay time element to the delaying queue

#### Priority Queue

-   `OnAddWeight` will be called when add an specified priority element to the priority queue

#### RateLimiting Queue

-   `OnAddLimited` will be called when add an specified rate limit element to the rate limiting queue
-   `OnForget` will be called when forget an element from the rate limiting queue
-   `OnGetTimes` will be called when get the number of times an element has been limited from the rate limiting queue

## 2. With Custom Queue

`WorkQueue` is designed to be easily extensible which mean you can easily write a new queue type and use it with WorkQueue. You can implement the `Interface`, `Callback` interfaces and reference `QConfig` to create a new queue type.

`WorkQueue` provides two queue types, `Queue` and `Simple Queue`. You can refer to them to create your own queue type.

Following is used the `Simple Queue` as an example to set base for `Delaying Queue`.

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
	base := workqueue.NewSimpleQueue(nil)
	q := workqueue.NewDelayingQueueWithCustomQueue(nil, base)

	go func() {
		for {
			element, err := q.Get()
			if err != nil && !errors.Is(err, workqueue.ErrorQueueEmpty) {
				fmt.Println(err)
				return
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

```

## 3. Limiter

The limiter only works for `RateLimiting Queue`, it determines the rate limit of the element. Default rate limit is based on the `token bucket` algorithm. You can define your own rate limit algorithm by implementing the `RateLimiter` interface.

### Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/shengyanli1982/workqueue"
)

func main() {
	conf := workqueue.NewQConfig()
	conf.WithLimiter(workqueue.NewTokenBucketRateLimiter(3, 1)) // set rate limit

	q := workqueue.NewRateLimitingQueue(conf)

	go func() {
		for {
			element, shutdown := q.Get()
			if shutdown {
				fmt.Println("shutdown")
				return
			}
			fmt.Println("get element:", element)
			q.Done(element) // mark element as done, 'Done' is required after 'Get'
		}
	}()

	_ = q.AddLimited("hello", time.Second)
	_ = q.AddLimited("world", time.Second)

	time.Sleep(time.Second * 2) // wait for element to be executed

	q.Stop()
}
```

# Thanks to

-   [kubernetes/client-go](https://github.com/kubernetes/client-go)
-   [lxzan/memorycache](https://github.com/lxzan/memorycache)
