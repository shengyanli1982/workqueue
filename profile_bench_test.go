package workqueue

// profile_bench_test.go 并发争用 pprof 采集设施：同一窗口内采集 CPU+mutex+block+goroutine
// profile，写入 pprof_output/。
//
// 要点：采集仅在本文件 _Pprof bench 内开启，退出时恢复；ns/op 仅确认窗口有效，
// 禁止与 plain bench 定量对比；精确归因请单独进程运行目标 bench；
// 禁止与 -cpuprofile flag 同用。workload 复用 parallel_bench_test.go 助手函数。

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"testing"
)

// contentionProfileRate mutex/block profile 采样率，1 = 全采样。
const contentionProfileRate = 1

// runWithContentionProfiles 在 CPU+mutex+block 采集窗口内执行 fn，结束后落盘四类 profile。
func runWithContentionProfiles(b *testing.B, name string, fn func()) {
	b.StopTimer()
	runtime.SetMutexProfileFraction(contentionProfileRate)
	runtime.SetBlockProfileRate(contentionProfileRate)
	// 任何路径都恢复采样率，避免污染后续 bench。
	defer func() {
		runtime.SetMutexProfileFraction(0)
		runtime.SetBlockProfileRate(0)
	}()
	// 清零构造段计时与分配统计。
	b.ResetTimer()

	// 复用 runWithCPUProfile 执行 fn。
	runWithCPUProfile(b, name, fn)

	// 落盘 profile。
	writeContentionProfile(b, "mutex", name)
	writeContentionProfile(b, "block", name)
	writeContentionProfile(b, "goroutine", name)
}

// writeContentionProfile 将 profile 以 protobuf 格式写入 pprof_output/。
func writeContentionProfile(b *testing.B, kind, name string) {
	p := pprof.Lookup(kind)
	if p == nil {
		b.Fatalf("profile %q not available", kind)
	}

	filename := filepath.Join("pprof_output", fmt.Sprintf("%s_%s_%s.prof", kind, b.Name(), name))
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		b.Fatalf("failed to create pprof output dir: %v", err)
	}

	f, err := os.Create(filename)
	if err != nil {
		b.Fatalf("failed to create %s profile file: %v", kind, err)
	}
	defer f.Close()

	if err := p.WriteTo(f, 0); err != nil {
		b.Fatalf("failed to write %s profile: %v", kind, err)
	}
}

// BenchmarkQueue_PutGet_Parallel_Pprof 基础队列并发锁争用 pprof 采集。
func BenchmarkQueue_PutGet_Parallel_Pprof(b *testing.B) {
	q := NewQueue(nil)

	b.ReportAllocs()

	runWithContentionProfiles(b, "parallel_put_get", func() {
		parallelQueuePutGet(b, q)
	})
}

// BenchmarkDelayingQueue_PutGet_Parallel_Pprof 延迟队列并发争用 pprof 采集。
func BenchmarkDelayingQueue_PutGet_Parallel_Pprof(b *testing.B) {
	q := NewDelayingQueue(nil)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()

	runWithContentionProfiles(b, "parallel_put_get", func() {
		parallelDelayingQueuePutGet(b, q)
	})
}

// BenchmarkPriorityQueue_PutGet_Parallel_Pprof 优先级队列并发争用 pprof 采集。
func BenchmarkPriorityQueue_PutGet_Parallel_Pprof(b *testing.B) {
	q := NewPriorityQueue(nil)

	b.ReportAllocs()

	runWithContentionProfiles(b, "parallel_put_get", func() {
		parallelPriorityQueuePutGet(b, q)
	})
}

// BenchmarkRateLimitingQueue_PutGet_Parallel_Pprof 限流队列并发争用 pprof 采集。
func BenchmarkRateLimitingQueue_PutGet_Parallel_Pprof(b *testing.B) {
	config := NewRateLimitingQueueConfig().WithLimiter(benchRateLimiter())
	q := NewRateLimitingQueue(config)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()

	runWithContentionProfiles(b, "parallel_put_get", func() {
		parallelRateLimitingQueuePutGet(b, q)
	})
}
