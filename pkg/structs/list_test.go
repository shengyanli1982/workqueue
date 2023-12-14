package workqueue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkStandard(t *testing.T) {
	l := &Deque{}
	l.Push(&Node{data: "foo"})
	l.Push(&Node{data: "bar"})
	l.Push(&Node{data: "baz"})
	assert.Equal(t, 3, l.length)
	assert.Equal(t, "foo", l.Pop().data)
	assert.Equal(t, "bar", l.Pop().data)
	assert.Equal(t, "baz", l.Pop().data)
	assert.Equal(t, 0, l.length)
	l.Push(&Node{data: "foo"})
	l.Push(&Node{data: "bar"})
	l.Push(&Node{data: "baz"})
	assert.Equal(t, "baz", l.PopBack().data)
	assert.Equal(t, "bar", l.PopBack().data)
	assert.Equal(t, "foo", l.PopBack().data)
	assert.Equal(t, 0, l.length)
	l.PushFront(&Node{data: "foo"})
	l.PushFront(&Node{data: "bar"})
	l.PushFront(&Node{data: "baz"})
	assert.Equal(t, "baz", l.Pop().data)
	assert.Equal(t, "bar", l.Pop().data)
	assert.Equal(t, "foo", l.Pop().data)
	assert.Equal(t, 0, l.length)
}

func TestLinkReset(t *testing.T) {
	l := &Deque{}
	l.Push(&Node{data: "foo"})
	l.Push(&Node{data: "bar"})
	l.Push(&Node{data: "baz"})
	assert.Equal(t, 3, l.length)
	l.Reset()
	assert.Equal(t, 0, l.length)
}

func TestLinkDelete(t *testing.T) {
	l := &Deque{}
	l.Push(&Node{data: "foo"})
	l.Push(&Node{data: "bar"})
	l.Push(&Node{data: "baz"})
	assert.Equal(t, 3, l.length)
	l.Delete(l.head.next)
	assert.Equal(t, "foo", l.Pop().data)
	assert.Equal(t, "baz", l.Pop().data)
	assert.Equal(t, 0, l.length)
}

func TestLinkHeadAndTail(t *testing.T) {
	l := &Deque{}
	l.Push(&Node{data: "foo"})
	l.Push(&Node{data: "bar"})
	l.Push(&Node{data: "baz"})
	assert.Equal(t, "foo", l.Head().data)
	assert.Equal(t, "baz", l.Tail().data)
}

func BenchmarkLinkPush(b *testing.B) {
	l := &Deque{}
	for i := 0; i < b.N; i++ {
		l.Push(&Node{data: i})
	}
}

func BenchmarkLinkPushFront(b *testing.B) {
	l := &Deque{}
	for i := 0; i < b.N; i++ {
		l.PushFront(&Node{data: i})
	}
}

func BenchmarkLinkPop(b *testing.B) {
	l := &Deque{}
	for i := 0; i < b.N; i++ {
		l.Push(&Node{data: i})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Pop()
	}
}

func BenchmarkLinkPopBack(b *testing.B) {
	l := &Deque{}
	for i := 0; i < b.N; i++ {
		l.Push(&Node{data: i})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.PopBack()
	}
}
