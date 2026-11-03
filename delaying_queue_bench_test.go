package workqueue

import "testing"

func BenchmarkDelayingQueue_Put(b *testing.B) {
	q := NewDelayingQueue(nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}
}

func BenchmarkDelayingQueue_PutWithDelay(b *testing.B) {
	q := NewDelayingQueue(nil)
	b.ReportAllocs()
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

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = q.Get()
	}
}

func BenchmarkDelayingQueue_PutAndGet(b *testing.B) {
	q := NewDelayingQueue(nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
		_, _ = q.Get()
	}
}

// BenchmarkDelayingQueue_PutWithDelayAndGet 测量 PutWithDelay(delay=0)→Get→Done 全链：
// delay=0 自 C-1 起走直通分支直接入内层队列（不进延迟堆、无 scheduler 搬运，入队后
// 同步可见），负载同构退化为内层 Queue 的 Put+getWithSpin 消费→Done 完成周期。
// Get 必命中、稳态工作集有界（保留旧版对 100ms 延迟下 Get 恒失败且堆无界增长的修复意图）。
// C-1 前本 bench 测量的是"scheduler 出堆搬入内层队列"的到期搬运链，ns/op 含异步搬运
// 墙钟延迟；直通形态下该搬运税已不存在。
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
