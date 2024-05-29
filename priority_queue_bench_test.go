package workqueue

import "testing"

func BenchmarkPriorityQueue_Put(b *testing.B) {
	q := NewPriorityQueue(nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}
}

func BenchmarkPriorityQueue_PutWithPriority(b *testing.B) {
	q := NewPriorityQueue(nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.PutWithPriority(i, int64(b.N))
	}
}

func BenchmarkPriorityQueue_Get(b *testing.B) {
	q := NewPriorityQueue(nil)

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = q.Get()
	}
}

func BenchmarkPriorityQueue_PutAndGet(b *testing.B) {
	q := NewPriorityQueue(nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
		_, _ = q.Get()
	}
}

func BenchmarkPriorityQueue_PutWithPriorityAndGet(b *testing.B) {
	q := NewPriorityQueue(nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.PutWithPriority(i, int64(b.N))
		_, _ = q.Get()
	}
}
