package workqueue

import "testing"

func BenchmarkDelayingQueue_Put(b *testing.B) {
	q := NewDelayingQueue(nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}
}

func BenchmarkDelayingQueue_PutWithDelay(b *testing.B) {
	q := NewDelayingQueue(nil)
	b.ResetTimer()

	defaultDelay := int64(100)

	for i := 0; i < b.N; i++ {
		_ = q.PutWithDelay(i, defaultDelay)
	}
}

func BenchmarkDelayingQueue_Get(b *testing.B) {
	q := NewDelayingQueue(nil)

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = q.Get()
	}
}

func BenchmarkDelayingQueue_PutAndGet(b *testing.B) {
	q := NewDelayingQueue(nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
		_, _ = q.Get()
	}
}
