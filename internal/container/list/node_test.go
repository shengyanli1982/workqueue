package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNode_Reset(t *testing.T) {
	node := NewNode()
	node.Value = "test"
	node.next = NewNode()
	node.prev = NewNode()

	node.Reset()

	// Verify that the value, index, next ptr is default
	assert.Nil(t, node.Value)
	assert.Nil(t, node.next)
	assert.Nil(t, node.prev)
}

func TestNodePool_Get(t *testing.T) {
	pool := NewNodePool()
	node := pool.Get()

	// Verify that the node is not nil
	assert.NotNil(t, node)

	// Verify that the value, index, next ptr is default
	assert.Nil(t, node.Value)
	assert.Nil(t, node.next)
	assert.Nil(t, node.prev)
}

func TestNodePool_Put(t *testing.T) {
	pool := NewNodePool()

	node := NewNode()
	node.Value = "test"
	node.next = NewNode()
	node.prev = NewNode()

	// Put the node back
	pool.Put(node)

	// Verify that the value, index, next ptr is default
	assert.Nil(t, node.Value)
	assert.Nil(t, node.next)
	assert.Nil(t, node.prev)
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
	assert.Nil(t, node.next)
	assert.Nil(t, node.prev)
}
