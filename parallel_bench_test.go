package workqueue

// parallel_bench_test.go 补齐并发（RunParallel）benchmark 覆盖缺口，并提供 floor 标定基准。
//
// 口径要点：ns/op 为聚合吞吐语义；floor ≈0.2ns/op（Go 1.26），无需扣除；
// 闭包内禁止写共享 sink；跨版本对比前须重跑 floor bench。
// 工作负载闭包提取为 parallel*PutGet 助手，供 profile_bench_test.go 复用。

import "testing"

// parallelQueuePutGet Queue 并发 Put+Get 的共享 RunParallel 工作负载。
func parallelQueuePutGet(b *testing.B, q Queue) {
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = q.Put(i)
			_, _ = q.Get() // 并发下可能为空，忽略错误：仍计入锁竞争成本
			i++
		}
	})
}

// parallelDelayingQueuePutGet DelayingQueue 并发 PutWithDelay(0)+Get 工作负载，
// delay=0 直通内层队列。
func parallelDelayingQueuePutGet(b *testing.B, q DelayingQueue) {
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Get 可能为空（本 worker 的直通项已先被他 worker 消费），
			// 忽略错误，仍计入消费端队列锁的竞争成本。
			_ = q.PutWithDelay(i, 0)
			_, _ = q.Get()
			i++
		}
	})
}

// parallelPriorityQueuePutGet PriorityQueue 并发 Put+Get+Done 工作负载。
func parallelPriorityQueuePutGet(b *testing.B, q PriorityQueue) {
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = q.Put(i)
			if v, err := q.Get(); err == nil {
				q.Done(v) // Done 闭环：走 OnDone 回调与 in-flight 路径
			}
			i++
		}
	})
}

// parallelRateLimitingQueuePutGet RateLimitingQueue 并发 PutWithLimited+Get 工作负载。
func parallelRateLimitingQueuePutGet(b *testing.B, q RateLimitingQueue) {
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = q.PutWithLimited(i)
			_, _ = q.Get() // 并发下可能为空，忽略错误：仍计入锁竞争成本
			i++
		}
	})
}

// BenchmarkDelayingQueue_PutGet_Parallel 测量延迟队列 PutWithDelay(0)→Get 并发吞吐。
func BenchmarkDelayingQueue_PutGet_Parallel(b *testing.B) {
	q := NewDelayingQueue(nil)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()
	b.ResetTimer()

	parallelDelayingQueuePutGet(b, q)
}

// BenchmarkPriorityQueue_PutGet_Parallel 测量优先级队列红黑树路径并发吞吐。
func BenchmarkPriorityQueue_PutGet_Parallel(b *testing.B) {
	q := NewPriorityQueue(nil)

	b.ReportAllocs()
	b.ResetTimer()

	parallelPriorityQueuePutGet(b, q)
}

// BenchmarkRateLimitingQueue_PutGet_Parallel 测量限流队列令牌充足热路径并发吞吐。
func BenchmarkRateLimitingQueue_PutGet_Parallel(b *testing.B) {
	config := NewRateLimitingQueueConfig().WithLimiter(benchRateLimiter())
	q := NewRateLimitingQueue(config)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()
	b.ResetTimer()

	parallelRateLimitingQueuePutGet(b, q)
}

// BenchmarkHarness_ParallelFloor 全库 RunParallel 空基准 floor 标定。
// 跨 Go 版本对比前须重跑。
func BenchmarkHarness_ParallelFloor(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
		}
	})
}
