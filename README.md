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

## 1. Structure

All Queue types are based on `Queue`(exclude `Simple Queue`), which mean will use `channel` to store elements and use `set` to track the state of the queue.

`Delaying Queue` and `Priority Queue` will use `heap` to maintain the expiration time and priority of the element.

`Simple Queue` is a simple queue, it is based on `channel` to store elements. No `set` and `heap` are used, so no element state is tracked and no element priority is maintained.

```bash
# go test -benchmem -run=^$ -bench ^Benchmark* github.com/shengyanli1982/workqueue/pkg/structs

goos: darwin
goarch: amd64
pkg: github.com/shengyanli1982/workqueue/pkg/structs
cpu: Intel(R) Xeon(R) CPU E5-2643 v2 @ 3.50GHz
BenchmarkHeapPush-12         	10202862	       117.9 ns/op	      88 B/op	       1 allocs/op
BenchmarkHeapPop-12          	11791902	       118.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkLinkPush-12         	16013065	        81.63 ns/op	      39 B/op	       1 allocs/op
BenchmarkLinkPushFront-12    	14881771	        80.47 ns/op	      39 B/op	       1 allocs/op
BenchmarkLinkPop-12          	244934114	         4.494 ns/op	       0 B/op	       0 allocs/op
BenchmarkLinkPopBack-12      	275577670	         4.408 ns/op	       0 B/op	       0 allocs/op
BenchmarkSetDelete-12        	11674332	       150.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkSetInsert-12        	 4378489	       347.3 ns/op	      86 B/op	       1 allocs/op
BenchmarkSetHas-12           	14812819	       156.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkStackPush-12        	13127522	        84.95 ns/op	      39 B/op	       1 allocs/op
BenchmarkStackPop-12         	270568515	         4.433 ns/op	       0 B/op	       0 allocs/op

```

## 2. Queue

Compare to [kubernetes/client-go](https://github.com/kubernetes/client-go) workqueue, WorkQueue has better performance and lower memory usage.

> [!NOTE]
> All types of queues are based on `Queue` which the same as `kubernetes/client-go` workqueue. So the performance and memory usage of all types of queues are the same as `Queue`.
>
> Why not compare to others? I think workqueue too close to the process of use and it is difficult to compare. If you have a better idea, please let me know.

```bash
# go test -benchmem -run=^$ -bench ^Benchmark* github.com/shengyanli1982/workqueue/bennchmark

goos: darwin
goarch: amd64
pkg: github.com/shengyanli1982/workqueue/bennchmark
cpu: Intel(R) Xeon(R) CPU E5-2643 v2 @ 3.50GHz
BenchmarkClientgoAdd-12           	 2631259	       435.1 ns/op	     155 B/op	       1 allocs/op
BenchmarkClientgoGet-12           	 4454460	       305.4 ns/op	       7 B/op	       0 allocs/op
BenchmarkClientgoAddAndGet-12     	 2272807	       532.4 ns/op	      86 B/op	       1 allocs/op
BenchmarkWorkqueueAdd-12          	13773114	        85.97 ns/op	       8 B/op	       0 allocs/op
BenchmarkWorkqueueGet-12          	19877964	        60.22 ns/op	       7 B/op	       0 allocs/op
BenchmarkWorkqueueAddAndGet-12    	 4895158	       249.4 ns/op	       8 B/op	       1 allocs/op
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
-   `WithCap` set the capacity of the queue, default is `2048`. If the capacity is `-1`, which mean the queue is unlimited.

### Methods

-   `Add` adds an element to the workqueue. If the element is already in the queue, it will not be added again.
-   `Get` gets an element from the workqueue. If the workqueue is empty, it will **`nonblock`** and return immediately.
-   `GetWithBlock` gets an element from the workqueue. If the workqueue is empty, it will **`blocking`** and waiting new element be added into queue.
-   `Done` marks an element as done with the workqueue. If the element is not in the workqueue, it will not be marked as done.
-   `Len` returns the elements count of the workqueue.
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
-   `WithCap` set the capacity of the queue, default is `2048`. If the capacity is `-1`, which mean the queue is unlimited.

### Methods

-   `Add` adds an element to the workqueue. If the element is already in the queue, it will not be added again.
-   `Get` gets an element from the workqueue. If the workqueue is empty, it will **`nonblock`** and return immediately.
-   `GetWithBlock` gets an element from the workqueue. If the workqueue is empty, it will **`blocking`** and waiting new element be added into queue.
-   `Done` marks an element as done with the workqueue. In fact in `Simple Queue`, it does nothing. Only left for compatibility.
-   `Len` returns the elements count of the workqueue.
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
-   `WithCap` set the capacity of the queue, default is `2048`. If the capacity is `-1`, which mean the queue is unlimited.

> [!TIP]
> please let `WithCap` behind `WithCallback`.

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
> The `Priority Queue` requires a window to sort the elements currently added to the Queue. The elements in this time window are sorted in order of `priority` from smallest to largest. The order of elements in two different time Windows is not guaranteed to be sorted by `priority`, even if the two Windows are immediately adjacent.
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
-   `WithCap` set the capacity of the queue, default is `2048`. If the capacity is `-1`, which mean the queue is unlimited.
-   `WithWindow` set the sort window size of the queue, default is `500` ms.

> [!TIP]
> please let `WithCap` behind `WithWindow`, `WithCallback`.

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
> default rate limit is based on the `token bucket` algorithm. You can define your own rate limit algorithm by implementing the `RateLimiter` interface.

### Create

-   `NewRateLimitingQueue` create a rate limiting queue, use `RateLimitingQConfig` to set config options. If the config is `nil`, the default config will be used.
-   `DefaultRateLimitingQueue` create a rate limiting queue with default config. It is equivalent to `NewRateLimitingQueue(nil)`, but return value implementing the `RateLimitingInterface` interface.

### Config

-   `WithCallback` set callback functions
-   `WithCap` set the capacity of the queue, default is `2048`. If the capacity is `-1`, which mean the queue is unlimited.
-   `WithLimiter` set the rate limiter of the queue, default is `TokenBucketRateLimiter`.

> [!TIP]
> please let `WithCap` behind `WithLimiter`, `WithCallback`.

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

## 2. Capacity

Queue capacity is a very important parameter, it determines the maximum number of elements that can be stored in the queue. If the capacity is `-1`, which mean the queue is unlimited. Default capacity is `2048`.

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
	conf.WithCap(100) // set queue capacity

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

	var err error
	for i := 0; i < 1000; i++ {
		if err = q.Add(i); err != nil {
			fmt.Println(err)
		}
	}

	time.Sleep(time.Second * 2) // wait for element to be executed

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
