package deque

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeReset(t *testing.T) {
	// Create a new node
	node := &Node{
		prev: &Node{},
		next: &Node{},
		data: "test",
	}

	// Call the Reset method
	node.Reset()

	// Verify the node is reset
	assert.Nil(t, node.prev)
	assert.Nil(t, node.next)
	assert.Nil(t, node.data)
}
func TestData(t *testing.T) {
	// Create a new node
	node := &Node{
		prev: &Node{},
		next: &Node{},
		data: "test",
	}

	// Call the Data method
	data := node.Data()

	// Verify the data is correct
	assert.Equal(t, "test", data)
}
func TestNodeSetData(t *testing.T) {
	// Create a new node
	node := &Node{
		prev: &Node{},
		next: &Node{},
		data: "test",
	}

	// Call the SetData method
	node.SetData("new data")

	// Verify the data is updated
	assert.Equal(t, "new data", node.data)
}
func TestDequeReset(t *testing.T) {
	// Create a new deque
	deque := NewDeque()

	// Create multiple nodes
	node1 := &Node{data: 1}
	node2 := &Node{data: 2}
	node3 := &Node{data: 3}

	// Add some elements to the deque
	deque.Push(node1)
	deque.Push(node2)
	deque.Push(node3)

	// Call the Reset method
	deque.Reset()

	// Verify the deque is reset
	assert.Equal(t, 0, deque.Len())
	assert.Nil(t, deque.Pop())
	assert.Nil(t, deque.PopBack())
}

func TestDequeResetEmpty(t *testing.T) {
	// Create a new empty deque
	deque := NewDeque()

	// Call the Reset method
	deque.Reset()

	// Verify the deque is reset
	assert.Equal(t, 0, deque.Len())
	assert.Nil(t, deque.Pop())
	assert.Nil(t, deque.PopBack())
}
func TestDequePush(t *testing.T) {
	// Create a new deque
	deque := NewDeque()

	// Create a new node
	node := &Node{data: "test"}

	// Call the Push method
	deque.Push(node)

	// Verify the node is added to the deque
	assert.Equal(t, 1, deque.Len())
	assert.Equal(t, node, deque.Pop())
}
func TestDequePop(t *testing.T) {
	// Create a new deque
	deque := NewDeque()

	// Create multiple nodes
	node1 := &Node{data: 1}
	node2 := &Node{data: 2}
	node3 := &Node{data: 3}

	// Add some elements to the deque
	deque.Push(node1)
	deque.Push(node2)
	deque.Push(node3)

	// Call the Pop method
	node := deque.Pop()

	// Verify the popped node is correct
	assert.NotNil(t, node)
	assert.Equal(t, 1, node.Data())
	assert.Equal(t, 2, deque.Len())
	assert.Equal(t, 2, deque.Pop().Data())
	assert.Equal(t, 3, deque.PopBack().Data())
}

func TestDequePopEmpty(t *testing.T) {
	// Create a new empty deque
	deque := NewDeque()

	// Call the Pop method
	node := deque.Pop()

	// Verify the popped node is nil
	assert.Nil(t, node)
	assert.Equal(t, 0, deque.Len())
	assert.Nil(t, deque.Pop())
	assert.Nil(t, deque.PopBack())
}
func TestDequePushFront(t *testing.T) {
	// Create a new deque
	deque := NewDeque()

	// Create a new node
	node := &Node{data: "test"}

	// Call the PushFront method
	deque.PushFront(node)

	// Verify the node is added to the deque
	assert.Equal(t, 1, deque.Len())
	assert.Equal(t, node, deque.Pop())
}

func TestDequePushFrontMultiple(t *testing.T) {
	// Create a new deque
	deque := NewDeque()

	// Create multiple nodes
	node1 := &Node{data: "test1"}
	node2 := &Node{data: "test2"}
	node3 := &Node{data: "test3"}

	// Call the PushFront method multiple times
	deque.PushFront(node1)
	deque.PushFront(node2)
	deque.PushFront(node3)

	// Verify the nodes are added to the deque in the correct order
	assert.Equal(t, 3, deque.Len())
	assert.Equal(t, node3, deque.Pop())
	assert.Equal(t, node1, deque.PopBack())
}
func TestDequePopBack(t *testing.T) {
	// Create a new deque
	deque := NewDeque()

	// Create multiple nodes
	node1 := &Node{data: 1}
	node2 := &Node{data: 2}
	node3 := &Node{data: 3}

	// Add some elements to the deque
	deque.Push(node1)
	deque.Push(node2)
	deque.Push(node3)

	// Call the PopBack method
	node := deque.PopBack()

	// Verify the popped node is correct
	assert.NotNil(t, node)
	assert.Equal(t, 3, node.Data())
	assert.Equal(t, 2, deque.Len())
	assert.Equal(t, 1, deque.Pop().Data())
	assert.Equal(t, 2, deque.PopBack().Data())
}

func TestDequePopBackEmpty(t *testing.T) {
	// Create a new empty deque
	deque := NewDeque()

	// Call the PopBack method
	node := deque.PopBack()

	// Verify the popped node is nil
	assert.Nil(t, node)
	assert.Equal(t, 0, deque.Len())
	assert.Nil(t, deque.Pop())
	assert.Nil(t, deque.PopBack())
}
func TestDequeDelete(t *testing.T) {
	// Create a new deque
	deque := NewDeque()

	// Create multiple nodes
	node1 := &Node{data: "test1"}
	node2 := &Node{data: "test2"}
	node3 := &Node{data: "test3"}

	// Add the nodes to the deque
	deque.Push(node1)
	deque.Push(node2)
	deque.Push(node3)

	// Delete node2 from the deque
	deque.Delete(node2)

	// Verify the node is deleted from the deque
	assert.Equal(t, 2, deque.Len())
	assert.Equal(t, node1, deque.Pop())
	assert.Equal(t, node3, deque.PopBack())
	assert.Nil(t, node2.prev)
	assert.Nil(t, node2.next)
}
func TestDequeLen(t *testing.T) {
	// Create a new deque
	deque := NewDeque()

	// Create multiple nodes
	node1 := &Node{data: 1}
	node2 := &Node{data: 2}
	node3 := &Node{data: 3}

	// Add some elements to the deque
	deque.Push(node1)
	deque.Push(node2)
	deque.Push(node3)

	// Call the Len method
	length := deque.Len()

	// Verify the length is correct
	assert.Equal(t, 3, length)
}

func TestDequeLenEmpty(t *testing.T) {
	// Create a new empty deque
	deque := NewDeque()

	// Call the Len method
	length := deque.Len()

	// Verify the length is 0
	assert.Equal(t, 0, length)
}
func TestDequeValues(t *testing.T) {
	// Create a new deque
	deque := NewDeque()

	// Create multiple nodes
	node1 := &Node{data: 1}
	node2 := &Node{data: 2}
	node3 := &Node{data: 3}

	// Add some elements to the deque
	deque.Push(node1)
	deque.Push(node2)
	deque.Push(node3)

	// Call the Values method
	values := deque.Values()

	// Verify the values are correct
	assert.Equal(t, []any{1, 2, 3}, values)
}

func TestDequeValuesEmpty(t *testing.T) {
	// Create a new empty deque
	deque := NewDeque()

	// Call the Values method
	values := deque.Values()

	// Verify the values are empty
	assert.Equal(t, []any{}, values)
}
