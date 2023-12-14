package structs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	s := NewStack()
	s.Push(&Node{data: "foo"})
	s.Push(&Node{data: "bar"})
	s.Push(&Node{data: "baz"})
	assert.Equal(t, 3, s.Len())
	assert.Equal(t, "baz", s.Pop().data)
	assert.Equal(t, "bar", s.Pop().data)
	assert.Equal(t, "foo", s.Pop().data)
	s.Reset()
}

func BenchmarkStack_Push(b *testing.B) {
	l := NewStack()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Push(&Node{data: i})
	}
}

func BenchmarkStack_Pop(b *testing.B) {
	l := NewStack()
	for i := 0; i < b.N; i++ {
		l.Push(&Node{data: i})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Pop()
	}
}

