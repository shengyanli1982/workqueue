package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNode_Reset(t *testing.T) {
	node := NewNode()
	node.Value = "test"
	node.Next = nil
	node.Index = 10

	node.Reset()

	// Verify that the value, index, next ptr is default
	assert.Nil(t, node.Value)
	assert.Nil(t, node.Next)
	assert.Equal(t, int64(0), node.Index)
}

func TestNodePool_Get(t *testing.T) {
	pool := NewNodePool()
	node := pool.Get()

	// Verify that the node is not nil
	assert.NotNil(t, node)

	// Verify that the value, index, next ptr is default
	assert.Nil(t, node.Value)
	assert.Nil(t, node.Next)
	assert.Equal(t, int64(0), node.Index)
}

func TestNodePool_Put(t *testing.T) {
	pool := NewNodePool()

	node := NewNode()
	node.Value = "test"
	node.Next = nil
	node.Index = 10

	// Put the node back
	pool.Put(node)

	// Verify that the value, index, next ptr is default
	assert.Nil(t, node.Value)
	assert.Nil(t, node.Next)
	assert.Equal(t, int64(0), node.Index)
}

func TestNodePool_PutAndGet(t *testing.T) {
	pool := NewNodePool()
	node := pool.Get()

	// Put the node back
	pool.Put(node)

	// Get the node again
	node = pool.Get()

	// Verify that the node is not nil
	assert.NotNil(t, node)

	// Verify that the value, index, next ptr is default
	assert.Nil(t, node.Value)
	assert.Nil(t, node.Next)
	assert.Equal(t, int64(0), node.Index)
}

func BenchmarkNodePool_Get(b *testing.B) {
	pool := NewNodePool()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pool.Get()
	}
}

func BenchmarkNodePool_Put(b *testing.B) {
	pool := NewNodePool()
	node := NewNode()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pool.Put(node)
	}
}

func BenchmarkNodePool_PutAndGet(b *testing.B) {
	pool := NewNodePool()
	node := pool.Get()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pool.Put(node)
		node = pool.Get()
	}
}
