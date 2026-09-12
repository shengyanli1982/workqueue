package workqueue

// bench_coverage_gap_test.go 补齐主路径 benchmark 覆盖缺口（性能评估阶段产物）。
//
// 覆盖缺口清单（对照 9 种队列主路径评估）：
//   - Queue: drainTracking Done 路径 / GetWithContext 阻塞消费 / 幂等 kindAny 泛型回退 /
//     ShutdownWithDrain 生命周期 / 并发（全库此前无任何 RunParallel benchmark）
//   - DelayingQueue: CancelDelay 快路径 / 到期搬运链（scheduler→内层 Put→Get）/
//     ShutdownWithDrain 生命周期
//   - RetryQueue: 默认 keyFunc（fmt.Sprintf("%T:%#v")）路径 / 延迟 Retry 分支
//   - LeasedQueue: Nack 重入队 / ExtendLease 续租
//   - BoundedBlockingQueue: PutWithContext + GetWithContext
//
// 命名约定：BenchmarkGap_QueueType_Op，与现有 benchmark 不重名；
// 全部使用 b.ReportAllocs，风格与 queue_pprof_bench_test.go 一致。
// 内存安全：绝大多数 benchmark 稳态工作集有界（避免 O(b.N) 预置导致大 benchtime 下 OOM）；
// 例外——BenchmarkGap_Queue_GetWithContext 的生产者协程无背压，消费跟不上时积压可达 O(b.N)。

import (
	"context"
	"errors"
	"testing"
	"time"
)

// gapAnyKey 是不可比较特化的复合 key 类型（含 string 字段的可比较 struct），
// 不落入 Set 的 string/int/int64/uint64 特化路径，用于驱动 kindAny 泛型回退分支。
type gapAnyKey struct {
	id   int
	name string
}

// BenchmarkGap_Queue_PutGetDone_DrainTracking 测量非幂等 + WithDrainTracking 的
// Put→Get→Done 全周期：Get 走 inFlight.Add(1)，Done 走 decrementInFlight 的
// CAS 循环。与 BenchmarkQueue_PutAndGet（无 tracking、无 Done）对照可得增量成本。
// 节点经 sync.Pool 循环复用，稳态内存有界。
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

// BenchmarkGap_Queue_GetWithContext 测量阻塞消费主路径：生产者协程连续 Put，
// 基准协程 GetWithContext 阻塞消费，覆盖惰性广播 channel 的
// “锁内空检查→waiters 注册→close(broadcast)+重建→select 唤醒→popLocked” 全链路。
// 此前全库无任何 GetWithContext 性能覆盖。
func BenchmarkGap_Queue_GetWithContext(b *testing.B) {
	q := NewQueue(nil)
	// GetWithContext 经 BlockingGetQueue 可选接口暴露（io.Closer 风格类型断言）。
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

// BenchmarkGap_Queue_Idempotent_AnyKey_PutGetDone 测量幂等模式 kindAny 泛型回退
// 路径（map[any]struct{} 双集合搬运）：与 BenchmarkQueue_Idempotent_PutGetDone
// （int → kindInt 特化）对照，量化类型特化收益与泛型回退成本。
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

// BenchmarkGap_Queue_PutAndGet_Parallel 测量多 goroutine 并发 Put+Get 的锁竞争
// 吞吐（GOMAXPROCS 个并行 worker 竞争同一 q.lock 与 sync.Pool）。
// 此前全库无任何 RunParallel benchmark，而并发吞吐是队列库生产环境主指标。
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

// BenchmarkGap_Queue_Idempotent_PutGetDone_Parallel 测量幂等模式并发
// Put→Get→Done 全周期：锁内双集合（state/processing）搬运在竞争下的吞吐。
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

// BenchmarkGap_Queue_ShutdownWithDrain 测量基础队列完整生命周期：
// 建队 → Put/Get/Done ×8 → ShutdownWithDrain（drained 快路径，无轮询等待）。
// 数值含队列构造/析构成本（drain 操作无法在同一队列上重复执行，此为固有测量口径）。
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
		// ShutdownWithDrain 经 DrainableQueue 可选接口暴露（io.Closer 风格类型断言）。
		if err := q.(DrainableQueue).ShutdownWithDrain(ctx); err != nil {
			b.Fatalf("shutdown with drain failed: %v", err)
		}
	}
}

// BenchmarkGap_DelayingQueue_CancelDelay 测量 CancelDelay 堆顶命中快路径：
// 每迭代先 PutWithDelay（60s，不触发到期搬运）再 Cancel，堆内恒为 1 项，
// Front 命中 → Pop → 归还节点池 → notifyWake。线性扫描路径为 O(n) 冷路径不在覆盖范围。
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

// BenchmarkGap_DelayingQueue_ExpiredTransport 测量真实到期搬运链：
// PutWithDelay(1ms) → scheduler 堆顶 timer 到期 → nextBatch 出堆 →
// 内层 Queue.Put → getWithSpin 消费。ns/op 含 1ms 墙钟等待（解释数据时须扣除），
// CPU profile 中可分离搬运链各环节成本。
// 对照：现有 BenchmarkDelayingQueue_PutWithDelayAndGet（100ms 延迟）中 Get 恒失败，
// 实测的是 PutWithDelay+空 Get，并未覆盖搬运链。
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

// BenchmarkGap_DelayingQueue_ShutdownWithDrain 测量延迟队列完整生命周期：
// 建队（含 scheduler 协程）→ 4 即时项 Put/Get/Done + 4 未到期延迟项 →
// ShutdownWithDrain（未到期项不阻塞 drain，按 Q4a 契约关停丢弃）。
// 数值含队列/scheduler 协程构造与析构成本。
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
		// ShutdownWithDrain 经 DrainableQueue 可选接口暴露（io.Closer 风格类型断言）。
		if err := q.(DrainableQueue).ShutdownWithDrain(ctx); err != nil {
			b.Fatalf("shutdown with drain failed: %v", err)
		}
	}
}

// BenchmarkGap_RetryQueue_RetryPath_DefaultKeyFunc 测量默认 keyFunc
// （fmt.Sprintf("%T:%#v")）下的完整 Retry 路径：keyOf 在 Retry 与 Forget 中各执行
// 一次。对照 BenchmarkRetryQueue_RetryPath（常量 keyFunc "k"）可分离
// fmt.Sprintf 反射格式化的成本。
func BenchmarkGap_RetryQueue_RetryPath_DefaultKeyFunc(b *testing.B) {
	cfg := NewRetryQueueConfig().
		WithPolicy(NewExponentialRetryPolicy(time.Nanosecond, time.Nanosecond, -1))
	// 不设置 WithKeyFunc：使用默认 defaultRetryKeyFunc（fmt.Sprintf）。
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

// BenchmarkGap_RetryQueue_RetryPath_Delayed 测量 Retry 的延迟分支：
// policy 返回 1ms（≥ time.Millisecond）→ 走 PutWithDelay 入堆 → scheduler 搬运 →
// getWithSpin 消费。对照 BenchmarkRetryQueue_RetryPath（ns 级 policy → immediate
// Put 分支），补齐 Retry 两条入队分支的覆盖。ns/op 含 ~1ms 墙钟等待。
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

// BenchmarkGap_LeasedQueue_GetNack 测量 Nack 重入队主路径：
// Put → GetWithLease → Nack（claim-first: removeLease → Put 归还 → Done）→ 再次
// GetWithLease 消费重入队项 → Ack。稳态队列/租约表恒空，工作集有界。
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

// BenchmarkGap_LeasedQueue_ExtendLease 测量 ExtendLease 续租路径：
// Put → GetWithLease → ExtendLease（锁内 map 读写 + time.Now().Add）→ Ack。
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

// BenchmarkGap_BoundedBlockingQueue_PutGetWithContext 测量 ctx 阻塞变体主路径：
// PutWithContext（slots 信号量 select）→ GetWithContext（items 信号量 select）。
// 对照 BenchmarkBoundedBlockingQueue_PutGet（无 ctx 的 Put/Get）量化 ctx 分支开销。
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
