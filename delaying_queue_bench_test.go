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

// BenchmarkDelayingQueue_PutWithDelayAndGet 测量真实到期搬运链：PutWithDelay(delay=0) →
// scheduler 出堆搬入内层队列 → getWithSpin 消费 → Done 完成周期。delay=0 使项立即到期，
// Get 必命中、堆与内层队列稳态有界（修复旧版 100ms 延迟下 Get 恒失败且堆无界增长）。
// ns/op 含 scheduler 异步搬运的墙钟延迟（解释数据时须扣除）。
func BenchmarkDelayingQueue_PutWithDelayAndGet(b *testing.B) {
	q := NewDelayingQueue(nil)
	b.Cleanup(q.Shutdown)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.PutWithDelay(i, 0)
		v := getWithSpin(b, q)
		q.Done(v)
	}
}
