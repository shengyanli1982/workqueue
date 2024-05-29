package set

import "testing"

func BenchmarkSet_Add(b *testing.B) {
	s := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Add(i)
	}
}

func BenchmarkSet_Remove(b *testing.B) {
	s := New()

	for i := 0; i < b.N; i++ {
		s.Add(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Remove(i)
	}
}

func BenchmarkSet_Contains(b *testing.B) {
	s := New()

	for i := 0; i < b.N; i++ {
		s.Add(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Contains(i)
	}
}
