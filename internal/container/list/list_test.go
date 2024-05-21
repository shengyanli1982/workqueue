package list

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList_PushBack(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Equal(t, node3, node2.Next, "node2 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")
}

func TestList_PopFront(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Pop the front node
	poppedNode := l.PopFront()

	// Verify the popped node
	assert.Equal(t, node1, poppedNode, "popped node should be node1")

	// Verify the list state
	assert.Equal(t, int64(2), l.Len(), "list length should be 2")
	assert.Equal(t, node2, l.Front(), "front node should be node2")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node3, node2.Next, "node2 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")
}

func TestList_PushAndPop(t *testing.T) {
	l := New()

	count := math.MaxUint16

	// Create some nodes
	for i := 0; i < count; i++ {
		l.PushBack(&Node{Value: int64(i)})
	}

	// Verify
	assert.Equal(t, int64(count), l.Len(), fmt.Sprintf("list length should be %v", count))
	assert.Equal(t, int64(0), l.Front().Value, fmt.Sprintf("front node index should be %v", 0))
	assert.Equal(t, int64(count-1), l.Back().Value, fmt.Sprintf("back node index should be %v", count-1))

	// Pop all nodes
	for i := 0; i < count; i++ {
		node := l.PopFront()
		assert.Equal(t, int64(i), node.Value, fmt.Sprintf("popped node index should be %v", i))
	}

	// Verify
	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")

	// Pop from empty list
	nilNode := l.PopFront()
	assert.Nil(t, nilNode, "popped node should be nil")
}

func TestList_PushAndPopWithParallel(t *testing.T) {

}

func TestList_Remove(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Remove node2 from the list
	l.Remove(node2)

	// Verify the list state
	assert.Equal(t, int64(2), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node3, node1.Next, "node1 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")

	// Remove node1 from the list
	l.Remove(node1)

	// Verify the list state
	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node3, l.Front(), "front node should be node3")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Nil(t, node3.Next, "node3 next should be nil")

	// Remove node3 from the list
	l.Remove(node3)

	// Verify the list state
	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
}

func TestList_InsertAfter(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Insert node4 after node2
	node4 := &Node{Value: 4}
	l.InsertAfter(node2, node4)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Equal(t, node4, node2.Next, "node2 next should be node4")
	assert.Equal(t, node3, node4.Next, "node4 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")
}

func TestList_InsertBefore(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Insert node4 before node2
	node4 := &Node{Value: 4}
	l.InsertBefore(node2, node4)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node4, node1.Next, "node1 next should be node4")
	assert.Equal(t, node2, node4.Next, "node4 next should be node2")
	assert.Equal(t, node3, node2.Next, "node2 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")
}

func TestList_InsertBefore_Head(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Insert node4 before head
	node4 := &Node{Value: 4}
	l.InsertBefore(l.Front(), node4)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node4, l.Front(), "front node should be node4")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node1, node4.Next, "node4 next should be node1")
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Equal(t, node3, node2.Next, "node2 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")
}

func TestList_InsertBefore_NilNext(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Insert node4 before nil next
	node4 := &Node{Value: 4}
	l.InsertBefore(nil, node4)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3, insert before nil next should not change the list")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Equal(t, node3, node2.Next, "node2 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")
}

func TestList_MoveToFront(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Move node2 to the front
	l.MoveToFront(node2)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node2, l.Front(), "front node should be node2")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node1, node2.Next, "node2 next should be node1")
	assert.Equal(t, node3, node1.Next, "node1 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")

	// Move node3 to the front
	l.MoveToFront(node3)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node3, l.Front(), "front node should be node3")
	assert.Equal(t, node1, l.Back(), "back node should be node1")

	// Verify the node order
	assert.Equal(t, node2, node3.Next, "node3 next should be node2")
	assert.Equal(t, node1, node2.Next, "node2 next should be node1")
	assert.Nil(t, node1.Next, "node1 next should be nil")

	// Move node1 to the front
	l.MoveToFront(node1)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node3, node1.Next, "node1 next should be node3")
	assert.Equal(t, node2, node3.Next, "node3 next should be node2")
	assert.Nil(t, node2.Next, "node2 next should be nil")
}

func TestList_MoveToBack(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Move node1 to the back
	l.MoveToBack(node1)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node2, l.Front(), "front node should be node2")
	assert.Equal(t, node1, l.Back(), "back node should be node1")

	// Verify the node order
	assert.Equal(t, node3, node2.Next, "node2 next should be node3")
	assert.Equal(t, node1, node3.Next, "node3 next should be node1")
	assert.Nil(t, node1.Next, "node1 next should be nil")

	// Move node2 to the back
	l.MoveToBack(node2)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node3, l.Front(), "front node should be node3")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	// Verify the node order
	assert.Equal(t, node1, node3.Next, "node3 next should be node1")
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Nil(t, node2.Next, "node2 next should be nil")

	// Move node3 to the back
	l.MoveToBack(node3)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Equal(t, node3, node2.Next, "node2 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")
}

func TestList_Swap(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Swap node1 and node2
	l.Swap(node1, node2)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node2, l.Front(), "front node should be node2")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node1, node2.Next, "node2 next should be node1")
	assert.Equal(t, node3, node1.Next, "node1 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")

	// Swap node2 and node3
	l.Swap(node2, node3)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node3, l.Front(), "front node should be node3")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	// Verify the node order
	assert.Equal(t, node1, node3.Next, "node3 next should be node1")
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Nil(t, node2.Next, "node2 next should be nil")

	// Swap node1 and node3
	l.Swap(node1, node3)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	// Verify the node order
	assert.Equal(t, node3, node1.Next, "node1 next should be node3")
	assert.Equal(t, node2, node3.Next, "node3 next should be node2")
	assert.Nil(t, node2.Next, "node2 next should be nil")
}

func TestList_Swap_InvalidNodes(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Swap with nil nodes
	l.Swap(nil, nil)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Equal(t, node3, node2.Next, "node2 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")

	// Swap with one nil node
	l.Swap(node1, nil)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Equal(t, node3, node2.Next, "node2 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")

	// Swap with the same node
	l.Swap(node1, node1)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Equal(t, node3, node2.Next, "node2 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")
}

func TestList_Range(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Define a callback function
	var result []int64
	callback := func(n *Node) bool {
		result = append(result, int64(n.Value.(int)))
		return true
	}

	// Call the Range method
	l.Range(callback)

	// Verify the result
	expected := []int64{1, 2, 3}
	assert.Equal(t, expected, result, "result should match the expected values")
}

func TestList_Cleanup(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Call the Cleanup method
	l.Cleanup()

	// Verify the list state
	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
}

func TestList_PushAndPopWithPool(t *testing.T) {
	l := New()
	pool := NewNodePool()

	count := math.MaxUint16

	// Create some nodes
	for i := 0; i < count; i++ {
		n := pool.Get()
		n.Value = int64(i)
		l.PushBack(n)
	}

	// Verify
	assert.Equal(t, int64(count), l.Len(), fmt.Sprintf("list length should be %v", count))
	assert.Equal(t, int64(0), l.Front().Value, fmt.Sprintf("front node index should be %v", 0))
	assert.Equal(t, int64(count-1), l.Back().Value, fmt.Sprintf("back node index should be %v", count-1))

	// Pop all nodes
	for i := 0; i < count; i++ {
		node := l.PopFront()
		assert.Equal(t, int64(i), node.Value, fmt.Sprintf("popped node index should be %v", i))
		pool.Put(node)
	}

	// Verify
	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")

	// Pop from empty list
	nilNode := l.PopFront()
	assert.Nil(t, nilNode, "popped node should be nil")
}
