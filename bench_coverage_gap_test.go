package workqueue

// bench_coverage_gap_test.go 补齐主路径 benchmark 覆盖缺口。
// 覆盖：Queue（drainTracking/阻塞消费/幂等泛型/生命周期/并发）、
// DelayingQueue（CancelDelay/到期搬运/生命周期）、RetryQueue（默认 keyFunc/延迟分支）、
// LeasedQueue（Nack/ExtendLease）、BoundedBlockingQueue（ctx 变体）。
// 命名约定：BenchmarkGap_QueueType_Op；全部使用 b.ReportAllocs。

import (
	"context"
	"errors"
	"testing"
	"time"
)

// gapAnyKey 复合 key 类型，不落入特化路径，驱动 kindAny 泛型回退分支。
type gapAnyKey struct {
	id   int
	name string
}

// BenchmarkGap_Queue_PutGetDone_DrainTracking 测量 drainTracking 下 Put→Get→Done 全周期。
func BenchmarkGap_Queue_PutGetDone_DrainTracking(b *testing.B) {
	conf := NewQueueConfig().WithDrainTracking()
	q := NewQueue(conf)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
		v, _ := q.Get()
		q.Done(v)
	}
}

// BenchmarkGap_Queue_GetWithContext 测量阻塞消费主路径全链路。
func BenchmarkGap_Queue_GetWithContext(b *testing.B) {
	q := NewQueue(nil)
	// GetWithContext 经 BlockingGetQueue 可选接口暴露。
	bq := q.(BlockingGetQueue)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	go func() {
		for i := 0; i < b.N; i++ {
			_ = q.Put(i)
		}
	}()

	for i := 0; i < b.N; i++ {
		if _, err := bq.GetWithContext(ctx); err != nil {
			b.Fatalf("get with context failed: %v", err)
		}
	}
}

// BenchmarkGap_Queue_Idempotent_AnyKey_PutGetDone 测量幂等模式 kindAny 泛型回退路径。
func BenchmarkGap_Queue_Idempotent_AnyKey_PutGetDone(b *testing.B) {
	conf := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(conf)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := gapAnyKey{id: i, name: "k"}
		_ = q.Put(key)
		v, _ := q.Get()
		q.Done(v)
	}
}

// BenchmarkGap_Queue_PutAndGet_Parallel 测量多 goroutine 并发 Put+Get 锁竞争吞吐。
func BenchmarkGap_Queue_PutAndGet_Parallel(b *testing.B) {
	q := NewQueue(nil)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = q.Put(i)
			_, _ = q.Get() // 并发下可能为空，忽略错误：仍计入锁竞争成本
			i++
		}
	})
}

// BenchmarkGap_Queue_Idempotent_PutGetDone_Parallel 测量幂等模式并发全周期吞吐。
func BenchmarkGap_Queue_Idempotent_PutGetDone_Parallel(b *testing.B) {
	conf := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(conf)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Put 可能因他 goroutine 同值在队/处理中而拒绝；Get 可能为空。
			// 仅对本人成功取出的值 Done，保持集合不变式，稳态工作集有界。
			if err := q.Put(i); err == nil {
				if v, gerr := q.Get(); gerr == nil {
					q.Done(v)
				}
			}
			i++
		}
	})
}

// BenchmarkGap_Queue_ShutdownWithDrain 测量基础队列完整生命周期含构造/析构成本。
func BenchmarkGap_Queue_ShutdownWithDrain(b *testing.B) {
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q := NewQueue(nil)
		for j := 0; j < 8; j++ {
			_ = q.Put(j)
		}
		for j := 0; j < 8; j++ {
			v, _ := q.Get()
			q.Done(v)
		}
		if err := q.(DrainableQueue).ShutdownWithDrain(ctx); err != nil {
			b.Fatalf("shutdown with drain failed: %v", err)
		}
	}
}

// BenchmarkGap_DelayingQueue_CancelDelay 测量 CancelDelay 堆顶命中快路径。
func BenchmarkGap_DelayingQueue_CancelDelay(b *testing.B) {
	q := NewDelayingQueue(nil)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := q.PutWithDelay(i, 60_000); err != nil {
			b.Fatalf("put with delay failed: %v", err)
		}
		if !q.CancelDelay(i) {
			b.Fatalf("cancel delay failed at %d", i)
		}
	}
}

// BenchmarkGap_DelayingQueue_ExpiredTransport 测量到期搬运链，ns/op 含 ~1ms 墙钟等待。
func BenchmarkGap_DelayingQueue_ExpiredTransport(b *testing.B) {
	q := NewDelayingQueue(nil)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := q.PutWithDelay(i, 1); err != nil {
			b.Fatalf("put with delay failed: %v", err)
		}
		v := getWithSpin(b, q)
		q.Done(v)
	}
}

// BenchmarkGap_DelayingQueue_ShutdownWithDrain 测量延迟队列完整生命周期含构造/析构成本。
func BenchmarkGap_DelayingQueue_ShutdownWithDrain(b *testing.B) {
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q := NewDelayingQueue(nil)
		for j := 0; j < 4; j++ {
			_ = q.Put(j)
			_ = q.PutWithDelay(j+1000, 60_000)
		}
		for j := 0; j < 4; j++ {
			v, _ := q.Get()
			q.Done(v)
		}
		if err := q.(DrainableQueue).ShutdownWithDrain(ctx); err != nil {
			b.Fatalf("shutdown with drain failed: %v", err)
		}
	}
}

// BenchmarkGap_RetryQueue_RetryPath_DefaultKeyFunc 测量默认 keyFunc（fmt.Sprintf）下的完整 Retry 路径。
func BenchmarkGap_RetryQueue_RetryPath_DefaultKeyFunc(b *testing.B) {
	cfg := NewRetryQueueConfig().
		WithPolicy(NewExponentialRetryPolicy(time.Nanosecond, time.Nanosecond, -1))
	// 使用默认 defaultRetryKeyFunc。
	q := NewRetryQueue(cfg)
	b.Cleanup(q.Shutdown)

	benchErr := errors.New("bench")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := q.Put(i); err != nil {
			b.Fatalf("put failed: %v", err)
		}
		value, err := q.Get()
		if err != nil {
			b.Fatalf("get failed: %v", err)
		}
		if err = q.Retry(value, benchErr); err != nil {
			b.Fatalf("retry failed: %v", err)
		}
		retried, err := q.Get()
		if err != nil {
			b.Fatalf("get retried value failed: %v", err)
		}
		q.Done(retried)
		q.Forget(retried)
	}
}

// BenchmarkGap_RetryQueue_RetryPath_Delayed 测量 Retry 延迟分支，ns/op 含 ~1ms 墙钟等待。
func BenchmarkGap_RetryQueue_RetryPath_Delayed(b *testing.B) {
	cfg := NewRetryQueueConfig().
		WithPolicy(NewExponentialRetryPolicy(time.Millisecond, time.Millisecond, -1)).
		WithKeyFunc(func(any) string { return "k" })
	q := NewRetryQueue(cfg)
	b.Cleanup(q.Shutdown)

	benchErr := errors.New("bench")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := q.Put(i); err != nil {
			b.Fatalf("put failed: %v", err)
		}
		value, err := q.Get()
		if err != nil {
			b.Fatalf("get failed: %v", err)
		}
		if err = q.Retry(value, benchErr); err != nil {
			b.Fatalf("retry failed: %v", err)
		}
		retried := getWithSpin(b, q)
		q.Done(retried)
		q.Forget(retried)
	}
}

// BenchmarkGap_LeasedQueue_GetNack 测量 Nack 重入队主路径。
func BenchmarkGap_LeasedQueue_GetNack(b *testing.B) {
	cfg := NewLeasedQueueConfig().
		WithLeaseDuration(time.Second).
		WithScanInterval(time.Hour)
	q := NewLeasedQueue(cfg)
	b.Cleanup(q.Shutdown)

	nackReason := errors.New("bench-nack")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := q.Put(i); err != nil {
			b.Fatalf("put failed: %v", err)
		}
		_, leaseID, err := q.GetWithLease(time.Second)
		if err != nil {
			b.Fatalf("get with lease failed: %v", err)
		}
		if err = q.Nack(leaseID, nackReason); err != nil {
			b.Fatalf("nack failed: %v", err)
		}
		_, leaseID2, err := q.GetWithLease(time.Second)
		if err != nil {
			b.Fatalf("get requeued value failed: %v", err)
		}
		if err = q.Ack(leaseID2); err != nil {
			b.Fatalf("ack failed: %v", err)
		}
	}
}

// BenchmarkGap_LeasedQueue_ExtendLease 测量 ExtendLease 续租路径。
func BenchmarkGap_LeasedQueue_ExtendLease(b *testing.B) {
	cfg := NewLeasedQueueConfig().
		WithLeaseDuration(time.Second).
		WithScanInterval(time.Hour)
	q := NewLeasedQueue(cfg)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := q.Put(i); err != nil {
			b.Fatalf("put failed: %v", err)
		}
		_, leaseID, err := q.GetWithLease(time.Second)
		if err != nil {
			b.Fatalf("get with lease failed: %v", err)
		}
		if err = q.ExtendLease(leaseID, time.Second); err != nil {
			b.Fatalf("extend lease failed: %v", err)
		}
		if err = q.Ack(leaseID); err != nil {
			b.Fatalf("ack failed: %v", err)
		}
	}
}

// BenchmarkGap_BoundedBlockingQueue_PutGetWithContext 测量 ctx 阻塞变体主路径。
func BenchmarkGap_BoundedBlockingQueue_PutGetWithContext(b *testing.B) {
	cfg := NewBoundedBlockingQueueConfig().WithCapacity(2048)
	q := NewBoundedBlockingQueue(cfg)
	b.Cleanup(q.Shutdown)

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := q.PutWithContext(ctx, i); err != nil {
			b.Fatalf("put with context failed: %v", err)
		}
		value, err := q.GetWithContext(ctx)
		if err != nil {
			b.Fatalf("get with context failed: %v", err)
		}
		q.Done(value)
	}
}
