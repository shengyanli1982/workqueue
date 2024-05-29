package workqueue

import "testing"

func BenchmarkQueue_Put(b *testing.B) {
	q := NewQueue(nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}
}

func BenchmarkQueue_Get(b *testing.B) {
	q := NewQueue(nil)

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = q.Get()
	}
}

func BenchmarkQueue_PutAndGet(b *testing.B) {
	q := NewQueue(nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
		_, _ = q.Get()
	}
}

func BenchmarkQueue_Idempotent_Put(b *testing.B) {
	conf := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(conf)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
		_ = q.Put(i)
	}
}

func BenchmarkQueue_Idempotent_Get(b *testing.B) {
	conf := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(conf)

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = q.Get()
	}
}

func BenchmarkQueue_Idempotent_PutAndGet(b *testing.B) {
	conf := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(conf)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
		_, _ = q.Get()
	}
}
