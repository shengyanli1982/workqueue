package deque

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDequeNodePool_Standard(t *testing.T) {
	p := NewListNodePool()
	assert.NotNil(t, p)

	list := NewDeque()

	b := p.Get()
	assert.NotNil(t, b)
	assert.Nil(t, b.data)
	assert.Nil(t, b.prev)
	assert.Nil(t, b.next)

	b.SetData("hello")
	assert.Equal(t, "hello", b.data)
	list.Push(b)

	b = p.Get()
	b.SetData("world")
	assert.Equal(t, "world", b.data)
	list.Push(b)
	assert.Equal(t, "hello", b.prev.data)

	ln := list.Pop()
	assert.Equal(t, "hello", ln.data)
	p.Put(ln)

	ln = list.Pop()
	assert.Equal(t, "world", ln.data)
	p.Put(ln)

	list.Reset()
}
