package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNode_Reset(t *testing.T) {
	node := NewNode()
	node.Value = "test"
	node.Right = NewNode()
	node.Left = NewNode()

	node.Reset()

	assert.Nil(t, node.Value)
	assert.Nil(t, node.Right)
	assert.Nil(t, node.Left)
}

func TestNodePool_Get(t *testing.T) {
	pool := NewNodePool()
	node := pool.Get()

	assert.NotNil(t, node)

	assert.Nil(t, node.Value)
	assert.Nil(t, node.Right)
	assert.Nil(t, node.Left)
}

func TestNodePool_Put(t *testing.T) {
	pool := NewNodePool()

	node := NewNode()
	node.Value = "test"
	node.Right = NewNode()
	node.Left = NewNode()

	pool.Put(node)

	assert.Nil(t, node.Value)
	assert.Nil(t, node.Right)
	assert.Nil(t, node.Left)
}

func TestNodePool_PutAndGet(t *testing.T) {
	pool := NewNodePool()
	node := pool.Get()

	pool.Put(node)

	node = pool.Get()

	assert.NotNil(t, node)

	assert.Nil(t, node.Value)
	assert.Nil(t, node.Right)
	assert.Nil(t, node.Left)
}
