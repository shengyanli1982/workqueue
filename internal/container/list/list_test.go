package list

import (
	"fmt"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func PrintListValues(l *List) {
	fmt.Println("List values: ============================")
	for i := l.Front(); i != nil; i = i.Right {
		fmt.Printf("Value: %v\n", i.Value)
	}
}

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
	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")
}

func TestList_PushFront(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushFront(node1)
	l.PushFront(node2)
	l.PushFront(node3)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node3, l.Front(), "front node should be node3")
	assert.Equal(t, node1, l.Back(), "back node should be node1")

	// Verify the node order
	assert.Equal(t, node2, node3.Right, "node3 next should be node2")
	assert.Equal(t, node1, node2.Right, "node2 next should be node1")
	assert.Nil(t, node1.Right, "node1 next should be nil")
}

func TestList_PopBack(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Pop the back node
	poppedNode := l.PopBack()

	// Verify the popped node
	assert.Equal(t, node3, poppedNode, "popped node should be node3")

	// Verify the list state
	assert.Equal(t, int64(2), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	// Verify the node order
	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Nil(t, node2.Right, "node2 next should be nil")
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
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")
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

func TestList_PushAndPop2(t *testing.T) {
	l := New()

	count := math.MaxUint16

	// Create some nodes
	for i := 0; i < count; i++ {
		l.PushFront(&Node{Value: int64(i)})
	}

	// Verify
	assert.Equal(t, int64(count), l.Len(), fmt.Sprintf("list length should be %v", count))
	assert.Equal(t, int64(count-1), l.Front().Value, fmt.Sprintf("front node index should be %v", count-1))
	assert.Equal(t, int64(0), l.Back().Value, fmt.Sprintf("back node index should be %v", 0))

	// Pop all nodes
	for i := 0; i < count; i++ {
		node := l.PopBack()
		assert.Equal(t, int64(i), node.Value, fmt.Sprintf("popped node index should be %v", i))
	}

	// Verify
	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")

	// Pop from empty list
	nilNode := l.PopBack()
	assert.Nil(t, nilNode, "popped node should be nil")
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

func TestList_PushAndPopWithPool2(t *testing.T) {
	l := New()
	pool := NewNodePool()

	count := math.MaxUint16

	// Create some nodes
	for i := 0; i < count; i++ {
		n := pool.Get()
		n.Value = int64(i)
		l.PushFront(n)
	}

	// Verify
	assert.Equal(t, int64(count), l.Len(), fmt.Sprintf("list length should be %v", count))
	assert.Equal(t, int64(count-1), l.Front().Value, fmt.Sprintf("front node index should be %v", count-1))
	assert.Equal(t, int64(0), l.Back().Value, fmt.Sprintf("back node index should be %v", 0))

	// Pop all nodes
	for i := 0; i < count; i++ {
		node := l.PopBack()
		assert.Equal(t, int64(i), node.Value, fmt.Sprintf("popped node index should be %v", i))
		pool.Put(node)
	}

	// Verify
	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")

	// Pop from empty list
	nilNode := l.PopBack()
	assert.Nil(t, nilNode, "popped node should be nil")
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
	assert.Equal(t, node3, node1.Right, "node1 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")

	// Remove node1 from the list
	l.Remove(node1)

	// Verify the list state
	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node3, l.Front(), "front node should be node3")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Nil(t, node3.Right, "node3 next should be nil")

	// Remove node3 from the list
	l.Remove(node3)

	// Verify the list state
	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
}

func TestList_Remove_InvalidNode(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Create an invalid node
	node4 := &Node{Value: 4}

	// Remove the invalid node from the list
	l.Remove(node4)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")
}

func TestList_Remove_EmptyList(t *testing.T) {
	l := New()

	// Create a node
	node := &Node{Value: 1}
	node.parentRef = toUnsafePtr(l)

	// Remove the node from the list (should have no effect)
	l.Remove(node)

	// Verify the list state
	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
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

	// Print the list values
	PrintListValues(l)

	// Move node2 to the front
	l.MoveToFront(node2)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node2, l.Front(), "front node should be node2")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node1, node2.Right, "node2 next should be node1")
	assert.Equal(t, node2, node1.Left, "node1 prev should be node2")
	assert.Equal(t, node1, node3.Left, "node3 prev should be node1")
	assert.Nil(t, node2.Left, "node2 prev should be nil")
}

func TestList_MoveToFront_FirstNode(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Print the list values
	PrintListValues(l)

	// Move node1 to the front
	l.MoveToFront(node1)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node1, node2.Left, "node2 prev should be node1")
	assert.Equal(t, node2, node3.Left, "node3 prev should be node2")
	assert.Nil(t, node1.Left, "node1 prev should be nil")
}

func TestList_MoveToFront_LastNode(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Print the list values
	PrintListValues(l)

	// Move node3 to the front
	l.MoveToFront(node3)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node3, l.Front(), "front node should be node3")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	// Verify the node order
	assert.Equal(t, node1, node3.Right, "node3 next should be node1")
	assert.Equal(t, node3, node1.Left, "node1 prev should be node3")
	assert.Equal(t, node1, node2.Left, "node2 prev should be node1")
	assert.Nil(t, node3.Left, "node3 prev should be nil")
}

func TestList_MoveToFront_SingleNode(t *testing.T) {
	l := New()

	// Create a single node
	node := &Node{Value: 1}

	// Push the node to the list
	l.PushBack(node)

	// Print the list values
	PrintListValues(l)

	// Move the node to the front
	l.MoveToFront(node)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node, l.Front(), "front node should be the single node")
	assert.Equal(t, node, l.Back(), "back node should be the single node")

	// Verify the node order
	assert.Nil(t, node.Left, "node prev should be nil")
	assert.Nil(t, node.Right, "node next should be nil")
}

func TestList_MoveToFront_InvalidNode(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Create an invalid node
	node4 := &Node{Value: 4}

	// Print the list values
	PrintListValues(l)

	// Move the invalid node to the front
	l.MoveToFront(node4)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 3")
	assert.Equal(t, node4, l.Front(), "front node should be node4")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	// Verify the node order
	assert.Equal(t, node1, node4.Right, "node4 next should be node1")
	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node1, node2.Left, "node2 prev should be node1")
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")

}

func TestList_MoveToFront_EmptyList(t *testing.T) {
	l := New()

	// Create a node
	node := &Node{Value: 1}

	// Print the list values
	PrintListValues(l)

	// Move the node to the front (should have no effect)
	l.MoveToFront(node)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(1), l.Len(), "list length should be 0")
	assert.Equal(t, node, l.Front(), "front node should be the node")
	assert.Equal(t, node, l.Back(), "back node should be the node")
	assert.Equal(t, node.parentRef, toUnsafePtr(l), "node parentRef should be the list")

	// Verify the node order
	assert.Nil(t, node.Left, "node prev should be nil")
	assert.Nil(t, node.Right, "node next should be nil")
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

	// Print the list values
	PrintListValues(l)

	// Move node2 to the back
	l.MoveToBack(node2)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	// Verify the node order
	assert.Equal(t, node3, node1.Right, "node1 next should be node3")
	assert.Equal(t, node3, node2.Left, "node2 prev should be node3")
	assert.Equal(t, node1, node3.Left, "node3 prev should be node1")
	assert.Nil(t, node2.Right, "node2 next should be nil")
}

func TestList_MoveToBack_LastNode(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Print the list values
	PrintListValues(l)

	// Move node3 to the back
	l.MoveToBack(node3)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node2")

	// Verify the node order
	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node1, node2.Left, "node2 prev should be node1")
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")
}

func TestList_MoveToBack_SingleNode(t *testing.T) {
	l := New()

	// Create a single node
	node := &Node{Value: 1}

	// Push the node to the list
	l.PushBack(node)

	// Print the list values
	PrintListValues(l)

	// Move the node to the back
	l.MoveToBack(node)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node, l.Front(), "front node should be the node")
	assert.Equal(t, node, l.Back(), "back node should be the node")

	// Verify the node order
	assert.Nil(t, node.Right, "node next should be nil")
	assert.Nil(t, node.Left, "node prev should be nil")
}

func TestList_MoveToBack_InvalidNode(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Create an invalid node
	node4 := &Node{Value: 4}

	// Print the list values
	PrintListValues(l)

	// Move the invalid node to the back
	l.MoveToBack(node4)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node4, l.Back(), "back node should be node3")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	// Verify the node order
	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node1, node2.Left, "node2 prev should be node1")
	assert.Equal(t, node2, node3.Left, "node3 prev should be node2")
	assert.Equal(t, node4, node3.Right, "node3 next should be node4")
	assert.Nil(t, node4.Right, "node4 next should be nil")

}

func TestList_MoveToBack_EmptyList(t *testing.T) {
	l := New()

	// Create a node
	node := &Node{Value: 1}

	// Print the list values
	PrintListValues(l)

	// Move the node to the front (should have no effect)
	l.MoveToBack(node)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(1), l.Len(), "list length should be 0")
	assert.Equal(t, node, l.Front(), "front node should be the node")
	assert.Equal(t, node, l.Back(), "back node should be the node")
	assert.Equal(t, node.parentRef, toUnsafePtr(l), "node parentRef should be the list")

	// Verify the node order
	assert.Nil(t, node.Left, "node prev should be nil")
	assert.Nil(t, node.Right, "node next should be nil")
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

	// Create a new node
	node4 := &Node{Value: 4}

	// Print the list values
	PrintListValues(l)

	// Insert newNode before node2
	l.InsertBefore(node4, node2)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	// Verify the node order
	assert.Equal(t, node4, node1.Right, "node1 next should be newNode")
	assert.Equal(t, node2, node4.Right, "newNode next should be node2")
	assert.Equal(t, node4, node2.Left, "node2 prev should be newNode")
	assert.Equal(t, node1, node4.Left, "newNode prev should be node1")
	assert.Nil(t, node3.Right, "node3 next should be nil")
}

func TestList_InsertBefore_SameValue(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 1}
	node3 := &Node{Value: 1}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Create a new node
	node4 := &Node{Value: 1}

	// Print the list values
	PrintListValues(l)

	// Insert newNode before node2
	l.InsertBefore(node4, node2)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	// Verify the node order
	assert.Equal(t, node4, node1.Right, "node1 next should be newNode")
	assert.Equal(t, node2, node4.Right, "newNode next should be node2")
	assert.Equal(t, node4, node2.Left, "node2 prev should be newNode")
	assert.Equal(t, node1, node4.Left, "newNode prev should be node1")
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")

}

func TestList_InsertBefore_FirstNode(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Create a new node
	node4 := &Node{Value: 4}

	// Print the list values
	PrintListValues(l)

	// Insert newNode before node1
	l.InsertBefore(node4, node1)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node4, l.Front(), "front node should be newNode")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	// Verify the node order
	assert.Equal(t, node1, node4.Right, "newNode next should be node1")
	assert.Equal(t, node4, node1.Left, "node1 prev should be newNode")
	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Equal(t, node2, node3.Left, "node3 prev should be node2")
	assert.Nil(t, node3.Right, "node3 next should be nil")
}

func TestList_InsertBefore_Nil(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}

	// Push nodes to the list
	l.PushBack(node1)

	// Print the list values
	PrintListValues(l)

	// Insert nil after node1 (should have no effect)
	l.InsertBefore(nil, node1)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node1, l.Back(), "back node should be node1")
	assert.Equal(t, node1.parentRef, toUnsafePtr(l), "node1 parentRef should be the list")

	// Insert node after nil (should have no effect)
	l.InsertBefore(node2, nil)

	// Verify the list state
	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node1, l.Back(), "back node should be node1")
	assert.Equal(t, node1.parentRef, toUnsafePtr(l), "node1 parentRef should be the list")
}

func TestList_InsertBefore_InvalidNode(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)

	// Create an invalid node
	node3 := &Node{Value: 3}

	// Print the list values
	PrintListValues(l)

	// Insert the invalid node before node2
	l.InsertBefore(node3, node2)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node3")
	assert.Equal(t, node3.parentRef, toUnsafePtr(l), "node3 parentRef should be the list")

	// Verify the node order
	assert.Equal(t, node3, node1.Right, "node1 next should be node3")
	assert.Equal(t, node1, node3.Left, "node3 prev should be node1")
	assert.Equal(t, node2, node3.Right, "node3 next should be node2")
	assert.Nil(t, node2.Right, "node2 next should be nil")

	// Create an invalid node
	node4 := &Node{Value: 4}

	// Print the list values
	PrintListValues(l)

	// Insert the invalid node before node1
	l.InsertBefore(node4, node1)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 2")
	assert.Equal(t, node4, l.Front(), "front node should be node4")
	assert.Equal(t, node2, l.Back(), "back node should be node2")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	// Verify the node order
	assert.Equal(t, node1, node4.Right, "node4 next should be node1")
	assert.Equal(t, node4, node1.Left, "node1 prev should be node4")
	assert.Equal(t, node3, node1.Right, "node1 next should be node3")
	assert.Equal(t, node1, node3.Left, "node3 prev should be node1")
	assert.Equal(t, node2, node3.Right, "node3 next should be node2")
	assert.Nil(t, node2.Right, "node2 next should be nil")
}

func TestList_InsertBefore_EmptyList(t *testing.T) {
	l := New()

	// Create a node
	node := &Node{Value: 1}

	// Print the list values
	PrintListValues(l)

	// Insert the node after nil (should have no effect)
	l.InsertBefore(node, nil)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(0), l.Len(), "list length should be 1")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
	assert.Nil(t, node.parentRef, "node parentRef should be nil")
}

func TestList_InsertAfter(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}
	node4 := &Node{Value: 4}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Print the list values
	PrintListValues(l)

	// Insert node4 after node1
	l.InsertAfter(node4, node1)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	// Verify the node order
	assert.Equal(t, node4, node1.Right, "node1 next should be node4")
	assert.Equal(t, node2, node4.Right, "node4 next should be node2")
	assert.Equal(t, node4, node2.Left, "node2 prev should be node4")
	assert.Equal(t, node1, node4.Left, "node4 prev should be node1")
}

func TestList_InsertAfter_SameValue(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 1}
	node3 := &Node{Value: 1}
	node4 := &Node{Value: 1}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Print the list values
	PrintListValues(l)

	// Insert node4 after node1
	l.InsertAfter(node4, node1)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	// Verify the node order
	assert.Equal(t, node4, node1.Right, "node1 next should be node4")
	assert.Equal(t, node2, node4.Right, "node4 next should be node2")
	assert.Equal(t, node4, node2.Left, "node2 prev should be node4")
	assert.Equal(t, node1, node4.Left, "node4 prev should be node1")
}

func TestList_InsertAfter_LastNode(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)

	// Print the list values
	PrintListValues(l)

	// Insert node3 after node2
	l.InsertAfter(node3, node2)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node3.parentRef, toUnsafePtr(l), "node3 parentRef should be the list")

	// Verify the node order
	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")
}

func TestList_InsertAfter_Nil(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}

	// Push nodes to the list
	l.PushBack(node1)

	// Print the list values
	PrintListValues(l)

	// Insert nil after node1 (should have no effect)
	l.InsertAfter(nil, node1)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node1, l.Back(), "back node should be node1")
	assert.Equal(t, node1.parentRef, toUnsafePtr(l), "node1 parentRef should be the list")

	// Insert node after nil (should have no effect)
	l.InsertAfter(node2, nil)

	// Verify the list state
	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node1, l.Back(), "back node should be node1")
	assert.Equal(t, node1.parentRef, toUnsafePtr(l), "node1 parentRef should be the list")
}

func TestList_InsertAfter_InvalidNode(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)

	// Create an invalid node
	node3 := &Node{Value: 3}

	// Print the list values
	PrintListValues(l)

	// Insert the invalid node after node2
	l.InsertAfter(node3, node1)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node3")
	assert.Equal(t, node3.parentRef, toUnsafePtr(l), "node3 parentRef should be the list")

	// Verify the node order
	assert.Equal(t, node3, node1.Right, "node1 next should be node3")
	assert.Equal(t, node1, node3.Left, "node3 prev should be node1")
	assert.Equal(t, node2, node3.Right, "node3 next should be node2")
	assert.Nil(t, node2.Right, "node2 next should be nil")

	// Create an invalid node
	node4 := &Node{Value: 4}

	// Print the list values
	PrintListValues(l)

	// Insert the invalid node after node2
	l.InsertAfter(node4, node2)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node4, l.Back(), "back node should be node4")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	// Verify the node order
	assert.Equal(t, node3, node1.Right, "node1 next should be node3")
	assert.Equal(t, node1, node3.Left, "node3 prev should be node1")
	assert.Equal(t, node2, node3.Right, "node3 next should be node2")
	assert.Equal(t, node4, node2.Right, "node2 next should be node4")
	assert.Equal(t, node2, node4.Left, "node4 prev should be node2")
	assert.Equal(t, node4, node2.Right, "node2 next should be node4")
	assert.Nil(t, node4.Right, "node4 next should be nil")

}

func TestList_InsertAfter_EmptyList(t *testing.T) {
	l := New()

	// Create a node
	node := &Node{Value: 1}

	// Print the list values
	PrintListValues(l)

	// Insert the node after nil (should have no effect)
	l.InsertAfter(node, nil)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(0), l.Len(), "list length should be 1")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
	assert.Nil(t, node.parentRef, "node parentRef should be nil")
}

func TestList_Swap(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}
	node4 := &Node{Value: 4}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)
	l.PushBack(node4)

	// Print the list values
	PrintListValues(l)

	// Swap node2 and node3
	l.Swap(node2, node3)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node4, l.Back(), "back node should be node4")

	// Verify the node order
	assert.Equal(t, node3, node1.Right, "node1 next should be node3")
	assert.Equal(t, node2, node3.Right, "node3 next should be node2")
	assert.Equal(t, node4, node2.Right, "node2 next should be node4")
	assert.Nil(t, node4.Right, "node4 next should be nil")

	assert.Equal(t, node1, node3.Left, "node3 prev should be node1")
	assert.Equal(t, node3, node2.Left, "node2 prev should be node3")
	assert.Equal(t, node2, node4.Left, "node4 prev should be node2")
	assert.Nil(t, node1.Left, "node2 prev should be nil")

	// Swap node1 and node4
	l.Swap(node1, node4)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node4, l.Front(), "front node should be node4")
	assert.Equal(t, node1, l.Back(), "back node should be node1")

	// Verify the node order
	assert.Equal(t, node3, node4.Right, "node4 next should be node3")
	assert.Equal(t, node2, node3.Right, "node3 next should be node2")
	assert.Equal(t, node1, node2.Right, "node2 next should be node1")
	assert.Nil(t, node1.Right, "node1 next should be nil")

	assert.Equal(t, node4, node3.Left, "node3 prev should be node4")
	assert.Equal(t, node3, node2.Left, "node2 prev should be node3")
	assert.Equal(t, node2, node1.Left, "node1 prev should be node2")
	assert.Nil(t, node4.Left, "node2 prev should be nil")

	// Swap node2 and node4
	l.Swap(node2, node4)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node2, l.Front(), "front node should be node4")
	assert.Equal(t, node1, l.Back(), "back node should be node1")

	// Verify the node order
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Equal(t, node4, node3.Right, "node3 next should be node4")
	assert.Equal(t, node1, node4.Right, "node4 next should be node1")
	assert.Nil(t, node1.Right, "node1 next should be nil")

	assert.Equal(t, node2, node3.Left, "node3 prev should be node2")
	assert.Equal(t, node3, node4.Left, "node4 prev should be node3")
	assert.Equal(t, node4, node1.Left, "node1 prev should be node4")
	assert.Nil(t, node2.Left, "node2 prev should be nil")

	// Swap node1 and node3
	l.Swap(node1, node3)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node2, l.Front(), "front node should be node4")
	assert.Equal(t, node3, l.Back(), "back node should be node1")

	// Verify the node order
	assert.Equal(t, node1, node2.Right, "node2 next should be node1")
	assert.Equal(t, node4, node1.Right, "node1 next should be node4")
	assert.Equal(t, node3, node4.Right, "node4 next should be node3")
	assert.Nil(t, node3.Right, "node1 next should be nil")

	assert.Equal(t, node2, node1.Left, "node1 prev should be node2")
	assert.Equal(t, node1, node4.Left, "node4 prev should be node1")
	assert.Equal(t, node2, node1.Left, "node1 prev should be node2")
	assert.Nil(t, node2.Left, "node2 prev should be nil")
}

func TestList_Swap_InvalidNode(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)

	// Create an invalid node
	node3 := &Node{Value: 3}

	// Print the list values
	PrintListValues(l)

	// Swap the invalid node with node2
	l.Swap(node3, node2)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(2), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node2")
	assert.Nil(t, node3.parentRef, "node3 parentRef should be nil")

	// Verify the node order
	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Nil(t, node2.Right, "node2 next should be nil")
	assert.Nil(t, node1.Left, "node1 prev should be nil")

	// Swap the invalid node with node1
	l.Swap(node3, node1)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(2), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node2")
	assert.Nil(t, node3.parentRef, "node3 parentRef should be nil")

	// Verify the node order
	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Nil(t, node2.Right, "node2 next should be nil")
	assert.Nil(t, node1.Left, "node1 prev should be nil")
}

func TestList_Swap_EmptyList(t *testing.T) {
	l := New()

	// Create a node
	node := &Node{Value: 1}

	// Print the list values
	PrintListValues(l)

	// Swap the node with nil (should have no effect)
	l.Swap(node, nil)

	// Print the list values
	PrintListValues(l)

	// Verify the list state
	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
	assert.Nil(t, node.parentRef, "node parentRef should be nil")
}

func TestList_Slice(t *testing.T) {
	l := New()

	// Create some nodes
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Push nodes to the list
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Call the Slice method
	s := l.Slice()

	// Verify the slice length
	assert.Equal(t, 3, len(s), "slice length should be 3")

	// Verify the slice values
	assert.Equal(t, node1.Value, s[0], "slice[0] should be node1 value")
	assert.Equal(t, node2.Value, s[1], "slice[1] should be node2 value")
	assert.Equal(t, node3.Value, s[2], "slice[2] should be node3 value")
}

func TestList_LargeDataSet(t *testing.T) {
	l := New()
	count := 1000000 // 1 million nodes

	// Test push performance
	startTime := time.Now()
	for i := 0; i < count; i++ {
		l.PushBack(&Node{Value: i})
	}
	pushDuration := time.Since(startTime)

	// Verify push performance
	assert.Less(t, pushDuration.Seconds(), float64(5), "pushing 1M nodes should take less than 5 seconds")
	assert.Equal(t, int64(count), l.Len(), "list length should be 1M")

	// Test pop performance
	startTime = time.Now()
	for i := 0; i < count; i++ {
		l.PopBack()
	}
	popDuration := time.Since(startTime)

	// Verify pop performance
	assert.Less(t, popDuration.Seconds(), float64(5), "popping 1M nodes should take less than 5 seconds")
	assert.Equal(t, int64(0), l.Len(), "list should be empty after popping all nodes")
}

func TestList_NilValues(t *testing.T) {
	l := New()

	// Create nodes with nil values
	node1 := &Node{Value: nil}
	node2 := &Node{Value: nil}
	node3 := &Node{Value: nil}

	// Push nodes with nil values
	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify node values
	assert.Nil(t, l.Front().Value, "front node value should be nil")
	assert.Nil(t, l.Back().Value, "back node value should be nil")
}

func TestList_EdgeCases(t *testing.T) {
	l := New()

	// Test operations on empty list
	assert.Nil(t, l.PopBack(), "PopBack on empty list should return nil")
	assert.Nil(t, l.PopFront(), "PopFront on empty list should return nil")
	assert.Nil(t, l.Front(), "Front on empty list should return nil")
	assert.Nil(t, l.Back(), "Back on empty list should return nil")

	// Test single node operations
	node := &Node{Value: 1}
	l.PushBack(node)
	l.MoveToFront(node) // Move to front when already at front
	assert.Equal(t, node, l.Front(), "node should still be at front")
	l.MoveToBack(node) // Move to back when already at back
	assert.Equal(t, node, l.Back(), "node should still be at back")

	// Test removing the same node multiple times
	l.Remove(node)
	l.Remove(node) // Try to remove already removed node
	assert.Equal(t, int64(0), l.Len(), "list should be empty after removing node")

	// Test operations with nil nodes
	l.PushBack(nil)
	l.PushFront(nil)
	l.Remove(nil)
	assert.Equal(t, int64(0), l.Len(), "list should remain empty after nil operations")
}

func TestList_ChainedOperations(t *testing.T) {
	l := New()

	// Test chained push operations
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	// Chain multiple operations
	// node3 --> node1 --> node2
	l.PushBack(node1)
	l.MoveToFront(node1)
	l.PushBack(node2)
	l.MoveToBack(node1)
	l.PushFront(node3)
	l.Remove(node2)
	l.PushBack(node2)

	// Verify the final state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node3, l.Front(), "front node should be node3")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	// Verify node order
	assert.Equal(t, node1, node3.Right, "node3 next should be node1")
	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Nil(t, node2.Right, "node2 next should be nil")
}

func TestList_ConcurrentOperations(t *testing.T) {
	l := New()
	done := make(chan bool)
	count := 1000
	var mu sync.Mutex

	// Concurrent push operations
	go func() {
		for i := 0; i < count; i++ {
			mu.Lock()
			l.PushBack(&Node{Value: i})
			mu.Unlock()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < count; i++ {
			mu.Lock()
			l.PushFront(&Node{Value: -i})
			mu.Unlock()
		}
		done <- true
	}()

	// Wait for push operations to complete
	<-done
	<-done

	// Verify list length
	assert.Equal(t, int64(count*2), l.Len(), "list length should be double the count")

	// Concurrent pop operations
	go func() {
		for i := 0; i < count; i++ {
			mu.Lock()
			l.PopBack()
			mu.Unlock()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < count; i++ {
			mu.Lock()
			l.PopFront()
			mu.Unlock()
		}
		done <- true
	}()

	// Wait for pop operations to complete
	<-done
	<-done

	// Verify list is empty
	assert.Equal(t, int64(0), l.Len(), "list should be empty after concurrent pops")
}

func TestList_ConcurrentOperations_Extended(t *testing.T) {
	l := New()
	done := make(chan bool)
	count := 1000
	valueMap := make(map[interface{}]bool)
	var mu sync.Mutex

	// Concurrent mixed operations
	go func() {
		for i := 0; i < count; i++ {
			node := &Node{Value: fmt.Sprintf("push_back_%d", i)}
			l.PushBack(node)
			mu.Lock()
			valueMap[node.Value] = true
			mu.Unlock()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < count; i++ {
			node := &Node{Value: fmt.Sprintf("push_front_%d", i)}
			l.PushFront(node)
			mu.Lock()
			valueMap[node.Value] = true
			mu.Unlock()
		}
		done <- true
	}()

	go func() {
		time.Sleep(100 * time.Millisecond) // Let some items accumulate
		for i := 0; i < count/2; i++ {
			node := l.PopBack()
			if node != nil {
				mu.Lock()
				delete(valueMap, node.Value)
				mu.Unlock()
			}
		}
		done <- true
	}()

	go func() {
		time.Sleep(100 * time.Millisecond) // Let some items accumulate
		for i := 0; i < count/2; i++ {
			node := l.PopFront()
			if node != nil {
				mu.Lock()
				delete(valueMap, node.Value)
				mu.Unlock()
			}
		}
		done <- true
	}()

	// Wait for all operations to complete
	for i := 0; i < 4; i++ {
		<-done
	}

	// Verify final state
	actualValues := make(map[interface{}]bool)
	for node := l.Front(); node != nil; node = node.Right {
		actualValues[node.Value] = true
	}

	// Compare maps
	assert.Equal(t, len(valueMap), len(actualValues), "number of values should match")
	for value := range valueMap {
		assert.True(t, actualValues[value], "value should exist in list")
	}
}

func TestList_ConcurrentInsertOperations(t *testing.T) {
	l := New()
	done := make(chan bool)
	count := 100
	nodes := make([]*Node, count)
	insertNodes := make([]*Node, count)
	var mu sync.Mutex

	// Initialize list
	for i := 0; i < count; i++ {
		nodes[i] = &Node{Value: i}
		insertNodes[i] = &Node{Value: fmt.Sprintf("insert_%d", i)}
		l.PushBack(nodes[i])
	}

	// Concurrent insert operations
	go func() {
		for i := 0; i < count; i++ {
			if i%2 == 0 {
				mu.Lock()
				l.InsertBefore(insertNodes[i], nodes[i])
				mu.Unlock()
			}
		}
		done <- true
	}()

	go func() {
		for i := 0; i < count; i++ {
			if i%2 == 1 {
				mu.Lock()
				l.InsertAfter(insertNodes[i], nodes[i])
				mu.Unlock()
			}
		}
		done <- true
	}()

	// Wait for operations to complete
	<-done
	<-done

	// Verify list length
	expectedLen := int64(count + count) // Original nodes + inserted nodes
	assert.Equal(t, expectedLen, l.Len(), "list length should match expected")

	// Verify list integrity
	var prev *Node
	nodeCount := 0
	for node := l.Front(); node != nil; node = node.Right {
		nodeCount++
		if prev != nil {
			assert.Equal(t, prev, node.Left, "node links should be consistent")
		}
		prev = node
	}
	assert.Equal(t, int(expectedLen), nodeCount, "node count should match list length")
}

func TestList_ConcurrentMoveOperations(t *testing.T) {
	l := New()
	done := make(chan bool)
	count := 100
	nodes := make([]*Node, count)
	var mu sync.Mutex

	// Initialize list
	for i := 0; i < count; i++ {
		nodes[i] = &Node{Value: i}
		l.PushBack(nodes[i])
	}

	// Concurrent move operations
	go func() {
		for i := 0; i < count; i++ {
			mu.Lock()
			l.MoveToFront(nodes[i])
			mu.Unlock()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < count; i++ {
			mu.Lock()
			l.MoveToBack(nodes[count-1-i])
			mu.Unlock()
		}
		done <- true
	}()

	// Wait for operations to complete
	<-done
	<-done

	// Verify list integrity
	assert.Equal(t, int64(count), l.Len(), "list length should remain unchanged")

	// Verify all nodes are still in the list
	nodeMap := make(map[interface{}]bool)
	for node := l.Front(); node != nil; node = node.Right {
		nodeMap[node.Value] = true
	}
	assert.Equal(t, count, len(nodeMap), "all nodes should still be present")

	// Verify list links
	var prev *Node
	for node := l.Front(); node != nil; node = node.Right {
		if prev != nil {
			assert.Equal(t, prev, node.Left, "node links should be consistent")
		}
		prev = node
	}
}
