package workqueue

// jitter_probe_bench_test.go P2-7 jitter 全局 rand 锁争用测量探针。
// 分离 limiter mu / applyJitter rand.Float64 / 纯计算三层成本。
// 口径对齐 parallel_bench_test.go；key 采用有界预生成 keyspace（4096 个）。

import (
	"strconv"
	"testing"
	"time"
)

// jitterProbeKeyspace 有界 key 集，failures map 至多 4096 条。
var jitterProbeKeyspace = func() []string {
	keys := make([]string, 4096)
	for i := range keys {
		keys[i] = "jitter-probe-key-" + strconv.Itoa(i)
	}
	return keys
}()

// jitterProbeSinkGuard 防止累加器被 DCE。
func jitterProbeSinkGuard(sum time.Duration) {
	if sum < 0 {
		panic("workqueue: unreachable jitter probe sink")
	}
}

// jitterProbeDelay/Ratio 固定探针参数。
const (
	jitterProbeDelay = 100 * time.Millisecond
	jitterProbeRatio = 0.5
)

// BenchmarkJitterProbe_ApplyJitter_Parallel 裸压 applyJitter，隔离 rand 成本。
func BenchmarkJitterProbe_ApplyJitter_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var sum time.Duration
		for pb.Next() {
			sum += applyJitter(jitterProbeDelay, jitterProbeRatio)
		}
		jitterProbeSinkGuard(sum)
	})
}

// BenchmarkJitterProbe_NoJitter_Parallel 对照组：无 rand 调用。
func BenchmarkJitterProbe_NoJitter_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var sum time.Duration
		for pb.Next() {
			sum += applyJitter(jitterProbeDelay, 0)
		}
		jitterProbeSinkGuard(sum)
	})
}

// BenchmarkJitterProbe_ApplyJitter_Serial 单线程基线。
func BenchmarkJitterProbe_ApplyJitter_Serial(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	var sum time.Duration
	for i := 0; i < b.N; i++ {
		sum += applyJitter(jitterProbeDelay, jitterProbeRatio)
	}
	jitterProbeSinkGuard(sum)
}

// BenchmarkJitterProbe_When_Jitter_Parallel 全路径实验组：When 并发含 jitter。
func BenchmarkJitterProbe_When_Jitter_Parallel(b *testing.B) {
	limiter := NewItemExponentialFailureRateLimiterWithJitter(
		time.Millisecond, time.Second, jitterProbeRatio)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var sum time.Duration
		i := 0
		for pb.Next() {
			sum += limiter.When(jitterProbeKeyspace[i&(len(jitterProbeKeyspace)-1)])
			i++
		}
		jitterProbeSinkGuard(sum)
	})
}

// BenchmarkJitterProbe_When_NoJitter_Parallel 全路径对照组：无 jitter。
func BenchmarkJitterProbe_When_NoJitter_Parallel(b *testing.B) {
	limiter := NewItemExponentialFailureRateLimiter(time.Millisecond, time.Second)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var sum time.Duration
		i := 0
		for pb.Next() {
			sum += limiter.When(jitterProbeKeyspace[i&(len(jitterProbeKeyspace)-1)])
			i++
		}
		jitterProbeSinkGuard(sum)
	})
}
