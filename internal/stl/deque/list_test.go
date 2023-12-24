package deque

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkStandard(t *testing.T) {
	l := NewDeque()
	l.Push(&Node{data: "foo"})
	l.Push(&Node{data: "bar"})
	l.Push(&Node{data: "baz"})
	assert.Equal(t, 3, l.Len())
	assert.Equal(t, "foo", l.Pop().data)
	assert.Equal(t, "bar", l.Pop().data)
	assert.Equal(t, "baz", l.Pop().data)
	assert.Equal(t, 0, l.Len())
	l.Push(&Node{data: "foo"})
	l.Push(&Node{data: "bar"})
	l.Push(&Node{data: "baz"})
	assert.Equal(t, "baz", l.PopBack().data)
	assert.Equal(t, "bar", l.PopBack().data)
	assert.Equal(t, "foo", l.PopBack().data)
	assert.Equal(t, 0, l.Len())
	l.PushFront(&Node{data: "foo"})
	l.PushFront(&Node{data: "bar"})
	l.PushFront(&Node{data: "baz"})
	assert.Equal(t, 3, l.Len())
	assert.Equal(t, "baz", l.Pop().data)
	assert.Equal(t, "bar", l.Pop().data)
	assert.Equal(t, "foo", l.Pop().data)
	assert.Equal(t, 0, l.Len())
}

func TestLinkPushFront(t *testing.T) {
	l := NewDeque()
	l.PushFront(&Node{data: "foo"})
	l.PushFront(&Node{data: "bar"})
	l.PushFront(&Node{data: "baz"})
	assert.Equal(t, 3, l.Len())
	assert.Equal(t, "baz", l.Pop().data)
	assert.Equal(t, "bar", l.Pop().data)
	assert.Equal(t, "foo", l.Pop().data)
	assert.Equal(t, 0, l.Len())
}

func TestLinkReset(t *testing.T) {
	l := NewDeque()
	l.Push(&Node{data: "foo"})
	l.Push(&Node{data: "bar"})
	l.Push(&Node{data: "baz"})
	assert.Equal(t, 3, l.Len())
	l.Reset()
	assert.Equal(t, 0, l.Len())
}

func TestLinkDelete(t *testing.T) {
	l := NewDeque()
	l.Push(&Node{data: "foo"})
	l.Push(&Node{data: "bar"})
	l.Push(&Node{data: "baz"})
	assert.Equal(t, 3, l.Len())
	l.Delete(l.head.next)
	assert.Equal(t, "foo", l.Pop().data)
	assert.Equal(t, "baz", l.Pop().data)
	assert.Equal(t, 0, l.Len())
}

func TestLinkHeadAndTail(t *testing.T) {
	l := NewDeque()
	l.Push(&Node{data: "foo"})
	l.Push(&Node{data: "bar"})
	l.Push(&Node{data: "baz"})
	assert.Equal(t, "foo", l.Head().data)
	assert.Equal(t, "baz", l.Tail().data)
}

func BenchmarkLink_Push(b *testing.B) {
	l := NewDeque()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Push(&Node{data: i})
	}
}

func BenchmarkLink_PushFront(b *testing.B) {
	l := NewDeque()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.PushFront(&Node{data: i})
	}
}

func BenchmarkLink_Pop(b *testing.B) {
	l := NewDeque()
	for i := 0; i < b.N; i++ {
		l.Push(&Node{data: i})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Pop()
	}
}

func BenchmarkLink_PopBack(b *testing.B) {
	l := NewDeque()
	for i := 0; i < b.N; i++ {
		l.Push(&Node{data: i})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.PopBack()
	}
}
