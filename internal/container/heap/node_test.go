package heap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNode_Reset(t *testing.T) {
	node := NewNode()
	node.Value = "test"
	node.left = NewNode()
	node.right = NewNode()
	node.parent = NewNode()

	node.Reset()

	// Verify that the value, index, next ptr is default
	assert.Nil(t, node.Value)
	assert.Nil(t, node.right)
	assert.Nil(t, node.left)
	assert.Nil(t, node.parent)
	assert.Equal(t, int64(0), node.color)
}

func TestNodePool_Get(t *testing.T) {
	pool := NewNodePool()
	node := pool.Get()

	// Verify that the node is not nil
	assert.NotNil(t, node)

	// Verify that the value, index, next ptr is default
	assert.Nil(t, node.Value)
	assert.Nil(t, node.right)
	assert.Nil(t, node.left)
	assert.Nil(t, node.parent)
	assert.Equal(t, int64(0), node.color)
}

func TestNodePool_Put(t *testing.T) {
	pool := NewNodePool()

	node := NewNode()
	node.Value = "test"
	node.left = NewNode()
	node.right = NewNode()
	node.parent = NewNode()

	// Put the node back
	pool.Put(node)

	// Verify that the value, index, next ptr is default
	assert.Nil(t, node.Value)
	assert.Nil(t, node.right)
	assert.Nil(t, node.left)
	assert.Nil(t, node.parent)
	assert.Equal(t, int64(0), node.color)
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
	assert.Nil(t, node.right)
	assert.Nil(t, node.left)
	assert.Nil(t, node.parent)
	assert.Equal(t, int64(0), node.color)
}
