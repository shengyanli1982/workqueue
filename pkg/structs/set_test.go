package structs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	s := make(Set)
	assert.Equal(t, 0, s.Len())
	assert.False(t, s.Has("foo"))
	s.Add("foo")
	assert.Equal(t, 1, s.Len())
	assert.True(t, s.Has("foo"))
	s.Delete("foo")
	assert.Equal(t, 0, s.Len())
	assert.False(t, s.Has("foo"))
}

func BenchmarkSetDelete(b *testing.B) {
	s := make(Set)
	for i := 0; i < b.N; i++ {
		s.Add(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Delete(i)
	}
}

func BenchmarkSetInsert(b *testing.B) {
	s := make(Set)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Add(i)
	}
}

func BenchmarkSetHas(b *testing.B) {
	s := make(Set)
	for i := 0; i < b.N; i++ {
		s.Add(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Has(i)
	}
}
