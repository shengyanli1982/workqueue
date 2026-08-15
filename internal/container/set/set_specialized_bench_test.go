package set

import (
	"strconv"
	"testing"
)

// 类型特化对照基准（P1）：覆盖四种特化类型、混合退化与通用类型路径。
// 负载形态与既有 BenchmarkSet_Add 一致：唯一 key 逐个写入，map 持续增长。

// benchStructKey 代表落入通用 map[any] 路径的自定义类型。
type benchStructKey struct {
	a int64
	b int64
}

func BenchmarkSet_TryAdd_Int(b *testing.B) {
	s := NewWithCapacity(64)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.TryAdd(i)
	}
}

func BenchmarkSet_TryAdd_Int64(b *testing.B) {
	s := NewWithCapacity(64)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.TryAdd(int64(i))
	}
}

func BenchmarkSet_TryAdd_Uint64(b *testing.B) {
	s := NewWithCapacity(64)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.TryAdd(uint64(i))
	}
}

func BenchmarkSet_TryAdd_String(b *testing.B) {
	keys := make([]string, b.N)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}

	s := NewWithCapacity(64)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.TryAdd(keys[i])
	}
}

func BenchmarkSet_TryAdd_Struct(b *testing.B) {
	s := NewWithCapacity(64)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.TryAdd(benchStructKey{a: int64(i)})
	}
}

func BenchmarkSet_TryAdd_Mixed(b *testing.B) {
	keys := make([]string, b.N)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}

	s := NewWithCapacity(64)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if i&1 == 0 {
			s.TryAdd(i)
		} else {
			s.TryAdd(keys[i])
		}
	}
}

func BenchmarkSet_Contains_Int(b *testing.B) {
	s := NewWithCapacity(64)
	for i := 0; i < b.N; i++ {
		s.Add(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Contains(i)
	}
}

func BenchmarkSet_Contains_String(b *testing.B) {
	keys := make([]string, b.N)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}

	s := NewWithCapacity(64)
	for i := 0; i < b.N; i++ {
		s.Add(keys[i])
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Contains(keys[i])
	}
}

func BenchmarkSet_TryRemove_Int(b *testing.B) {
	s := NewWithCapacity(64)
	for i := 0; i < b.N; i++ {
		s.Add(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.TryRemove(i)
	}
}

func BenchmarkSet_TryRemove_String(b *testing.B) {
	keys := make([]string, b.N)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}

	s := NewWithCapacity(64)
	for i := 0; i < b.N; i++ {
		s.Add(keys[i])
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.TryRemove(keys[i])
	}
}
