package deque

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListNodePool(t *testing.T) {
	p := &ListNodePool{} // Create a new ListNodePool instance
	node := &Node{}      // Create a new Node instance

	// Add the node to the pool
	p.bp.Put(node)

	// Get a node from the pool
	got := p.Get()

	// Verify that the returned node is the same as the one added to the pool
	assert.Equal(t, node, got)
}
