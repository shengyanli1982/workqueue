package workqueue

import (
	"strconv"
	"testing"
)

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

func BenchmarkQueue_Idempotent_PutGetDone(b *testing.B) {
	conf := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(conf)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
		v, _ := q.Get()
		q.Done(v)
	}
}

func BenchmarkQueue_Idempotent_DuplicatePut(b *testing.B) {
	conf := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(conf)
	_ = q.Put("same")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Put("same")
	}
}

// BenchmarkQueue_Idempotent_Put_String 测量字符串 key 下的幂等 Put 热路径，
// 与 BenchmarkQueue_Idempotent_Put（int key）结构一致：每元素两次 Put，
// 第二次命中去重返回 ErrElementAlreadyExist。P1 集合特化的对照基准。
func BenchmarkQueue_Idempotent_Put_String(b *testing.B) {
	conf := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(conf)

	keys := make([]string, b.N)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Put(keys[i])
		_ = q.Put(keys[i])
	}
}

// BenchmarkQueue_Idempotent_Done 测量幂等模式下 Done 的开销：
// 预先将 b.N 个元素全部 Get 走（处于处理中状态），再逐个 Done。
func BenchmarkQueue_Idempotent_Done(b *testing.B) {
	conf := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(conf)

	for i := 0; i < b.N; i++ {
		_ = q.Put(i)
	}

	values := make([]any, 0, b.N)
	for i := 0; i < b.N; i++ {
		v, _ := q.Get()
		values = append(values, v)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q.Done(values[i])
	}
}
