package workqueue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeapStandard(t *testing.T) {
	h := &Heap{}
	h.Push(NewElement("foo", 1))
	h.Push(NewElement("bar", 2))
	h.Push(NewElement("baz", 3))
	assert.Equal(t, 3, h.Len())
	assert.Equal(t, "foo", h.Pop().data)
	assert.Equal(t, "bar", h.Pop().data)
	assert.Equal(t, "baz", h.Pop().data)
	assert.Equal(t, 0, h.Len())
}

func TestHeapReset(t *testing.T) {
	h := &Heap{}
	h.Push(NewElement("foo", 1))
	h.Push(NewElement("bar", 2))
	h.Push(NewElement("baz", 3))
	assert.Equal(t, 3, h.Len())
	h.Reset()
	assert.Equal(t, 0, h.Len())
}

func TestHeapDelete(t *testing.T) {
	h := &Heap{}
	h.Push(NewElement("foo", 1))
	h.Push(NewElement("bar", 2))
	h.Push(NewElement("baz", 3))
	assert.Equal(t, 3, h.Len())
	h.Delete(1)
	assert.Equal(t, "foo", h.Pop().data)
	assert.Equal(t, "baz", h.Pop().data)
	assert.Equal(t, 0, h.Len())
}

func TestHeapUpdate(t *testing.T) {
	h := &Heap{}
	h.Push(NewElement("foo", 1))
	h.Push(NewElement("bar", 2))
	h.Push(NewElement("baz", 3))
	assert.Equal(t, 3, h.Len())
	h.Update(h.data[1], 4)
	assert.Equal(t, "foo", h.Pop().data)
	assert.Equal(t, "baz", h.Pop().data)
	assert.Equal(t, "bar", h.Pop().data)
	assert.Equal(t, 0, h.Len())
}

func TestHeapHead(t *testing.T) {
	h := &Heap{}
	h.Push(NewElement("foo", 1))
	h.Push(NewElement("bar", 2))
	h.Push(NewElement("baz", 3))
	assert.Equal(t, "foo", h.Head().data)
	assert.Equal(t, "foo", h.Head().data)
	assert.Equal(t, "foo", h.Head().data)
	assert.Equal(t, 3, h.Len())
}

func BenchmarkHeapPush(b *testing.B) {
	h := &Heap{}
	for i := 0; i < b.N; i++ {
		h.Push(NewElement(i, int64(i)))
	}
}

func BenchmarkHeapPop(b *testing.B) {
	h := &Heap{}
	for i := 0; i < b.N; i++ {
		h.Push(NewElement(i, int64(i)))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Pop()
	}
}
