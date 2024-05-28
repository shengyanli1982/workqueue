package workqueue

import "testing"

func BenchmarkRateLimitingQueue_Put(b *testing.B) {
	q := NewRateLimitingQueue(nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}
}

func BenchmarkRateLimitingQueue_PutWithLimited(b *testing.B) {
	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(10, 10))
	q := NewRateLimitingQueue(config)
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

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = q.Get()
	}
}

func BenchmarkRateLimitingQueue_PutAndGet(b *testing.B) {
	q := NewRateLimitingQueue(nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
		_, _ = q.Get()
	}
}

func BenchmarkRateLimitingQueue_PutWithLimitedAndGet(b *testing.B) {
	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(10, 10))
	q := NewRateLimitingQueue(config)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.PutWithLimited(i)
		_, _ = q.Get()
	}
}
