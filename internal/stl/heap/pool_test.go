package heap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeapElementPool(t *testing.T) {
	p := NewHeapElementPool()
	h := NewHeap()

	e1 := p.Get()
	e1.SetData("foo")
	e1.SetValue(10)
	assert.Equal(t, "foo", e1.Data())
	assert.Equal(t, 10, int(e1.Value()))

	h.Push(e1)
	assert.Equal(t, 1, h.Len())

	e2 := p.Get()
	e2.SetData("bar")
	e2.SetValue(2)
	h.Push(e2)
	assert.Equal(t, 2, h.Len())

	e3 := p.Get()
	e3.SetData("baz")
	e3.SetValue(6)
	h.Push(e3)
	assert.Equal(t, 3, h.Len())

	e := h.Pop()
	assert.Equal(t, "bar", e.Data())
	e.Reset()
	p.Put(e)

	e = h.Pop()
	assert.Equal(t, "baz", e.Data())
	e.Reset()
	p.Put(e)

	e = h.Pop()
	assert.Equal(t, "foo", e.Data())
	e.Reset()
	p.Put(e)
}
