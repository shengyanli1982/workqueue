package workqueue

import (
	"errors"
	"runtime"
	"testing"
	"time"
)

func getWithSpin(b *testing.B, q Queue) interface{} {
	for {
		value, err := q.Get()
		if err == nil {
			return value
		}
		if errors.Is(err, ErrQueueIsEmpty) {
			runtime.Gosched()
			continue
		}
		b.Fatalf("queue get failed: %v", err)
	}
}

func BenchmarkDeadLetterQueue_PutGetAck(b *testing.B) {
	q := NewDeadLetterQueue(nil)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()
	b.ResetTimer()

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
}

func BenchmarkRetryQueue_RetryPath(b *testing.B) {
	cfg := NewRetryQueueConfig().
		WithPolicy(NewExponentialRetryPolicy(time.Nanosecond, time.Nanosecond, -1)).
		WithKeyFunc(func(interface{}) string { return "k" })
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

func BenchmarkLeasedQueue_GetAck(b *testing.B) {
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
		if err = q.Ack(leaseID); err != nil {
			b.Fatalf("ack failed: %v", err)
		}
	}
}

func BenchmarkBoundedBlockingQueue_PutGet(b *testing.B) {
	cfg := NewBoundedBlockingQueueConfig().WithCapacity(2048)
	q := NewBoundedBlockingQueue(cfg)
	b.Cleanup(q.Shutdown)

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
		q.Done(value)
	}
}

func BenchmarkTimerQueue_PutAtGet(b *testing.B) {
	q := NewTimerQueue(nil)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := q.PutAt(i, time.Now()); err != nil {
			b.Fatalf("put at failed: %v", err)
		}
		value := getWithSpin(b, q)
		q.Done(value)
	}
}

func BenchmarkTimerQueue_Cancel(b *testing.B) {
	q := NewTimerQueue(nil)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := q.PutAfter(i, time.Second); err != nil {
			b.Fatalf("put after failed: %v", err)
		}
		if !q.Cancel(i) {
			b.Fatalf("cancel failed at %d", i)
		}
	}
}
