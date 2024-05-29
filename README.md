English | [中文](./README_CN.md)

<div align="center">
	<img src="assets/logo.png" alt="logo" width="500px">
</div>

[![Go Report Card](https://goreportcard.com/badge/github.com/shengyanli1982/workqueue/v2)](https://goreportcard.com/report/github.com/shengyanli1982/workqueue/v2)
[![Build Status](https://github.com/shengyanli1982/workqueue/v2/actions/workflows/test.yaml/badge.svg)](https://github.com/shengyanli1982/workqueue/v2/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/shengyanli1982/workqueue/v2.svg)](https://pkg.go.dev/github.com/shengyanli1982/workqueue/v2)

# Introduction

`WorkQueue` is a high-performance, thread-safe, and memory-efficient Go library for managing work queues. It offers a variety of queue implementations such as `Queue`, `DelayingQueue`, `PriorityQueue`, and `RateLimitingQueue`, each tailored for specific use cases and performance needs. The library's design is simple, user-friendly, and platform-independent, making it suitable for a broad spectrum of applications and environments.

After several iterations and real-world usage, we've gathered valuable user feedback and insights. This led to a complete redesign and optimization of WorkQueue's architecture and underlying code in the new version (v2), significantly enhancing its robustness, reliability, and security.

# Why Use WorkQueue(v2)

The WorkQueue(v2) has undergone a comprehensive architectural revamp, greatly improving its robustness and reliability. This redesign enables the library to manage demanding workloads with increased stability, making it ideal for both simple task queues and complex workflows. By utilizing advanced algorithms and optimized data structures, WorkQueue(v2) provides superior performance, efficiently managing larger task volumes, reducing latency, and increasing throughput.

WorkQueue(v2) offers a diverse set of queue implementations to cater to different needs, including standard task management, delayed execution, task prioritization, and rate-limited processing. This flexibility allows you to select the most suitable tool for your specific use case, ensuring optimal performance and functionality. With its cross-platform design, WorkQueue(v2) guarantees consistent behavior and performance across various operating systems, making it a versatile solution for different environments.

The development of WorkQueue(v2) has been heavily influenced by user feedback and real-world usage, resulting in a library that better meets the needs of its users. By addressing user-reported issues and incorporating feature requests, WorkQueue(v2) offers a more refined and user-centric experience.

Choosing WorkQueue(v2) for your application or project could be a great decision. :):)

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

Both the standard library and the project utilize the same algorithm, resulting in similar performance. However, the `list` in this project offers more features compared to the standard library.

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
