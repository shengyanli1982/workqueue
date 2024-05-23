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
