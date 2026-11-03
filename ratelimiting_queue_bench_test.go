package workqueue

// ratelimiting_queue_bench_test.go 覆盖限流队列入队/消费热路径。
//
// 口径说明：PutWithLimited 系列使用 benchRateLimiter（高速率大 burst 令牌桶），
// 测量的是"限流判定→令牌充足→直通 Put"热路径；令牌不足分支（When()>0 →
// PutWithDelay ≥1ms 入堆）必然导致堆 O(b.N) 无界增长，不作为 bench 口径，
// 其搬运链覆盖见 BenchmarkGap_DelayingQueue_ExpiredTransport。
// 已知陷阱：-bench 正则为前缀匹配（'BenchmarkRateLimitingQueue_Put' 会命中
// 本文件全部 5 个 bench 及 _Pprof/Parallel 变体），单目标运行请用 ^...$ 锚定。

import "testing"

// benchRateLimiter 构造限流 bench 统一使用的高速率大 burst 令牌桶：
// 1e9 tokens/s ≫ 本机任何 bench 消耗速率（含 RunParallel 聚合吞吐），
// When() 恒为 0 → PutWithLimited 走直通 Put 分支，完整经过
// "Reserve+Delay 限流判定→直通入队→OnLimited 回调"热路径，稳态工作集有界。
// 旧口径 NewBucketRateLimiterImpl(10,10)：burst 耗尽后 When() 恒 >0，全部
// 转 PutWithDelay（毫秒粒度，亚毫秒向上取整 1ms）入堆——堆 O(b.N) 无界增长、
// Get 恒空，实测的是堆插入与 scheduler 开销而非限流热路径（与
// BenchmarkDelayingQueue_PutWithDelayAndGet 旧 100ms 口径同类缺陷，同款修法，
// 见 delaying_queue_bench_test.go）。
func benchRateLimiter() Limiter {
	return NewBucketRateLimiterImpl(1e9, 1e9)
}

func BenchmarkRateLimitingQueue_Put(b *testing.B) {
	q := NewRateLimitingQueue(nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}
}

// BenchmarkRateLimitingQueue_PutWithLimited 测量 PutWithLimited 令牌充足热路径
// （limiter 口径见 benchRateLimiter）。Put-only 约定：无 Get，队列随迭代
// O(b.N) 增长，与 BenchmarkQueue_Put 等同款口径（节点经池分配，无堆搬运放大）。
func BenchmarkRateLimitingQueue_PutWithLimited(b *testing.B) {
	config := NewRateLimitingQueueConfig().WithLimiter(benchRateLimiter())
	q := NewRateLimitingQueue(config)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.PutWithLimited(i)
	}
}

func BenchmarkRateLimitingQueue_Get(b *testing.B) {
	q := NewRateLimitingQueue(nil)

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = q.Get()
	}
}

func BenchmarkRateLimitingQueue_PutAndGet(b *testing.B) {
	q := NewRateLimitingQueue(nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
		_, _ = q.Get()
	}
}

// BenchmarkRateLimitingQueue_PutWithLimitedAndGet 测量 PutWithLimited→Get 闭环
// 热路径：benchRateLimiter 口径下 When() 恒 0 直通入队，同轮 Get 必命中，
// 稳态工作集有界（修复旧 (10,10) 口径 burst 耗尽后入堆无界、Get 恒空——
// 旧版实测的是 PutWithDelay+空 Get，并未覆盖限流直通链）。
func BenchmarkRateLimitingQueue_PutWithLimitedAndGet(b *testing.B) {
	config := NewRateLimitingQueueConfig().WithLimiter(benchRateLimiter())
	q := NewRateLimitingQueue(config)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.PutWithLimited(i)
		_, _ = q.Get()
	}
}
