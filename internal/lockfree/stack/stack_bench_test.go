package stack

import "testing"

func BenchmarkCasStack_Put(b *testing.B) {
	s := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Push(int64(i))
	}
}

func BenchmarkCasStack_Get(b *testing.B) {
	s := New()
	for i := 0; i < b.N; i++ {
		s.Push(int64(i))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Pop()
	}
}

func BenchmarkCasStack_PutAndGet(b *testing.B) {
	s := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Push(int64(i))
		s.Pop()
	}
}
