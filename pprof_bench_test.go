package workqueue

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"testing"
	"time"
)

// runWithCPUProfile 在 fn 执行窗口内采集 CPU profile，写入 pprof_output/。
// 禁止与 -cpuprofile flag 同用。计时口径：setup 与收尾均在计时段外。
func runWithCPUProfile(b *testing.B, name string, fn func()) {
	b.StopTimer()

	filename := filepath.Join("pprof_output", fmt.Sprintf("cpu_%s_%s.prof", b.Name(), name))
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		b.Fatalf("failed to create pprof output dir: %v", err)
	}

	f, err := os.Create(filename)
	if err != nil {
		b.Fatalf("failed to create cpu profile file: %v", err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		b.Fatalf("failed to start CPU profile: %v", err)
	}

	b.ResetTimer()
	b.StartTimer()

	fn()

	// CPU 采样收尾。
	b.StopTimer()
	pprof.StopCPUProfile()
}

func runWithMemProfile(b *testing.B, name string, fn func()) {
	filename := filepath.Join("pprof_output", fmt.Sprintf("mem_%s_%s.prof", b.Name(), name))
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		b.Fatalf("failed to create pprof output dir: %v", err)
	}

	f, err := os.Create(filename)
	if err != nil {
		b.Fatalf("failed to create mem profile file: %v", err)
	}
	defer f.Close()

	// 先回收残留堆，排除在 ns/op 之外。
	runtime.GC()
	b.ResetTimer()

	fn()

	// 堆转储退出计时。
	b.StopTimer()
	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		b.Fatalf("failed to write heap profile: %v", err)
	}
}

func runWithTrace(b *testing.B, name string, fn func()) {
	filename := filepath.Join("pprof_output", fmt.Sprintf("trace_%s_%s.trace", b.Name(), name))
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		b.Fatalf("failed to create trace output dir: %v", err)
	}

	f, err := os.Create(filename)
	if err != nil {
		b.Fatalf("failed to create trace file: %v", err)
	}
	defer f.Close()

	if err := trace.Start(f); err != nil {
		b.Fatalf("failed to start trace: %v", err)
	}
	defer trace.Stop()

	fn()
}

func BenchmarkQueue_Put_Pprof(b *testing.B) {
	q := NewQueue(nil)
	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "put", func() {
		for i := 0; i < b.N; i++ {
			_ = q.Put(i)
		}
	})
}

func BenchmarkQueue_Get_Pprof(b *testing.B) {
	q := NewQueue(nil)
	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "get", func() {
		for i := 0; i < b.N; i++ {
			_, _ = q.Get()
		}
	})
}

func BenchmarkQueue_PutAndGet_Pprof(b *testing.B) {
	q := NewQueue(nil)
	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "put_get", func() {
		for i := 0; i < b.N; i++ {
			_ = q.Put(i)
			_, _ = q.Get()
		}
	})
}

func BenchmarkQueue_Idempotent_Put_Pprof(b *testing.B) {
	conf := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(conf)
	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "idempotent_put", func() {
		for i := 0; i < b.N; i++ {
			_ = q.Put(i)
		}
	})
}

func BenchmarkQueue_Idempotent_Get_Pprof(b *testing.B) {
	conf := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(conf)

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "idempotent_get", func() {
		for i := 0; i < b.N; i++ {
			_, _ = q.Get()
		}
	})
}

func BenchmarkQueue_Idempotent_PutGetDone_Pprof(b *testing.B) {
	conf := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(conf)
	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "idempotent_put_get_done", func() {
		for i := 0; i < b.N; i++ {
			_ = q.Put(i)
			v, _ := q.Get()
			q.Done(v)
		}
	})
}

func BenchmarkQueue_Idempotent_DuplicatePut_Pprof(b *testing.B) {
	conf := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(conf)
	_ = q.Put("same")

	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "idempotent_duplicate_put", func() {
		for i := 0; i < b.N; i++ {
			_ = q.Put("same")
		}
	})
}

func BenchmarkQueue_Mem_Pprof(b *testing.B) {
	q := NewQueue(nil)

	b.ReportAllocs()

	runWithMemProfile(b, "put", func() {
		for i := 0; i < b.N; i++ {
			_ = q.Put(i)
		}
	})
}

func BenchmarkQueue_Idempotent_Mem_Pprof(b *testing.B) {
	conf := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(conf)

	b.ReportAllocs()

	runWithMemProfile(b, "idempotent_put", func() {
		for i := 0; i < b.N; i++ {
			_ = q.Put(i)
		}
	})
}

func BenchmarkDelayingQueue_Put_Pprof(b *testing.B) {
	q := NewDelayingQueue(nil)
	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "put", func() {
		for i := 0; i < b.N; i++ {
			_ = q.Put(i)
		}
	})
}

func BenchmarkDelayingQueue_PutWithDelay_Pprof(b *testing.B) {
	q := NewDelayingQueue(nil)
	b.ReportAllocs()
	b.ResetTimer()

	defaultDelay := int64(100)

	runWithCPUProfile(b, "put_with_delay", func() {
		for i := 0; i < b.N; i++ {
			_ = q.PutWithDelay(i, defaultDelay)
		}
	})
}

func BenchmarkDelayingQueue_Get_Pprof(b *testing.B) {
	q := NewDelayingQueue(nil)

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "get", func() {
		for i := 0; i < b.N; i++ {
			_, _ = q.Get()
		}
	})
}

func BenchmarkDelayingQueue_PutAndGet_Pprof(b *testing.B) {
	q := NewDelayingQueue(nil)
	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "put_get", func() {
		for i := 0; i < b.N; i++ {
			_ = q.Put(i)
			_, _ = q.Get()
		}
	})
}

func BenchmarkPriorityQueue_Put_Pprof(b *testing.B) {
	q := NewPriorityQueue(nil)
	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "put", func() {
		for i := 0; i < b.N; i++ {
			_ = q.Put(i)
		}
	})
}

func BenchmarkPriorityQueue_PutWithPriority_Pprof(b *testing.B) {
	q := NewPriorityQueue(nil)
	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "put_with_priority", func() {
		for i := 0; i < b.N; i++ {
			_ = q.PutWithPriority(i, int64(i))
		}
	})
}

func BenchmarkPriorityQueue_Get_Pprof(b *testing.B) {
	q := NewPriorityQueue(nil)

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "get", func() {
		for i := 0; i < b.N; i++ {
			_, _ = q.Get()
		}
	})
}

func BenchmarkPriorityQueue_PutAndGet_Pprof(b *testing.B) {
	q := NewPriorityQueue(nil)
	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "put_get", func() {
		for i := 0; i < b.N; i++ {
			_ = q.Put(i)
			_, _ = q.Get()
		}
	})
}

// BenchmarkRateLimitingQueue_PutWithLimited_Pprof 限流队列令牌充足热路径 CPU profile。
func BenchmarkRateLimitingQueue_PutWithLimited_Pprof(b *testing.B) {
	config := NewRateLimitingQueueConfig().WithLimiter(benchRateLimiter())
	q := NewRateLimitingQueue(config)
	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "put_with_limited", func() {
		for i := 0; i < b.N; i++ {
			_ = q.PutWithLimited(i)
		}
	})
}

func BenchmarkRetryQueue_RetryPath_Pprof(b *testing.B) {
	cfg := NewRetryQueueConfig().
		WithPolicy(NewExponentialRetryPolicy(time.Nanosecond, time.Nanosecond, -1)).
		WithKeyFunc(func(any) string { return "k" })
	q := NewRetryQueue(cfg)
	b.Cleanup(q.Shutdown)

	benchErr := errors.New("bench")

	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "retry_path", func() {
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
	})
}

func BenchmarkDeadLetterQueue_PutGetAck_Pprof(b *testing.B) {
	q := NewDeadLetterQueue(nil)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "put_get_ack", func() {
		for i := 0; i < b.N; i++ {
			letter := &DeadLetter{
				Payload: i,
				Meta: map[string]string{
					"trace": "bench",
				},
			}
			if err := q.PutDead(letter); err != nil {
				b.Fatalf("put dead letter failed: %v", err)
			}
			got, err := q.GetDead()
			if err != nil {
				b.Fatalf("get dead letter failed: %v", err)
			}
			if err = q.AckDead(got); err != nil {
				b.Fatalf("ack dead letter failed: %v", err)
			}
		}
	})
}

func BenchmarkLeasedQueue_GetAck_Pprof(b *testing.B) {
	cfg := NewLeasedQueueConfig().
		WithLeaseDuration(time.Second).
		WithScanInterval(time.Hour)
	q := NewLeasedQueue(cfg)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "get_ack", func() {
		for i := 0; i < b.N; i++ {
			if err := q.Put(i); err != nil {
				b.Fatalf("put failed: %v", err)
			}
			_, leaseID, err := q.GetWithLease(time.Second)
			if err != nil {
				b.Fatalf("get with lease failed: %v", err)
			}
			if err = q.Ack(leaseID); err != nil {
				b.Fatalf("ack failed: %v", err)
			}
		}
	})
}

func BenchmarkBoundedBlockingQueue_PutGet_Pprof(b *testing.B) {
	cfg := NewBoundedBlockingQueueConfig().WithCapacity(2048)
	q := NewBoundedBlockingQueue(cfg)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "put_get", func() {
		for i := 0; i < b.N; i++ {
			if err := q.Put(i); err != nil {
				b.Fatalf("put failed: %v", err)
			}
			value, err := q.Get()
			if err != nil {
				b.Fatalf("get failed: %v", err)
			}
			q.Done(value)
		}
	})
}

func BenchmarkTimerQueue_PutAtGet_Pprof(b *testing.B) {
	q := NewTimerQueue(nil)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "put_at_get", func() {
		for i := 0; i < b.N; i++ {
			if err := q.PutAt(i, time.Now()); err != nil {
				b.Fatalf("put at failed: %v", err)
			}
			value := getWithSpin(b, q)
			q.Done(value)
		}
	})
}

func BenchmarkTimerQueue_Cancel_Pprof(b *testing.B) {
	q := NewTimerQueue(nil)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()
	b.ResetTimer()

	runWithCPUProfile(b, "cancel", func() {
		for i := 0; i < b.N; i++ {
			if err := q.PutAfter(i, time.Second); err != nil {
				b.Fatalf("put after failed: %v", err)
			}
			if !q.Cancel(i) {
				b.Fatalf("cancel failed at %d", i)
			}
		}
	})
}
