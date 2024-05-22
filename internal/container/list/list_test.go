package list

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func printListNodeValues(l *List) {
	fmt.Println("List nodes: ============================")
	for i := l.head; i != nil; i = i.Next {
		fmt.Printf("Node: %v\n", i.Value)
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
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Equal(t, node3, node2.Next, "node2 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")
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
	assert.Equal(t, node2, node3.Next, "node3 next should be node2")
	assert.Equal(t, node1, node2.Next, "node2 next should be node1")
	assert.Nil(t, node1.Next, "node1 next should be nil")
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
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Nil(t, node2.Next, "node2 next should be nil")
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

func TestList_PushAndPopWithParallel(t *testing.T) {

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

	printListNodeValues(l)

	// Move node2 to the front
	l.MoveToFront(node2)

	printListNodeValues(l)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node2, l.Front(), "front node should be node2")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node1, node2.Next, "node2 next should be node1")
	assert.Equal(t, node2, node1.Prev, "node1 prev should be node2")
	assert.Equal(t, node1, node3.Prev, "node3 prev should be node1")
	assert.Nil(t, node2.Prev, "node2 prev should be nil")
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

	printListNodeValues(l)

	// Move node1 to the front
	l.MoveToFront(node1)

	printListNodeValues(l)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Equal(t, node1, node2.Prev, "node2 prev should be node1")
	assert.Equal(t, node2, node3.Prev, "node3 prev should be node2")
	assert.Nil(t, node1.Prev, "node1 prev should be nil")
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

	printListNodeValues(l)

	// Move node3 to the front
	l.MoveToFront(node3)

	printListNodeValues(l)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node3, l.Front(), "front node should be node3")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	// Verify the node order
	assert.Equal(t, node1, node3.Next, "node3 next should be node1")
	assert.Equal(t, node3, node1.Prev, "node1 prev should be node3")
	assert.Equal(t, node1, node2.Prev, "node2 prev should be node1")
	assert.Nil(t, node3.Prev, "node3 prev should be nil")
}

func TestList_MoveToFront_SingleNode(t *testing.T) {
	l := New()

	// Create a single node
	node := &Node{Value: 1}

	// Push the node to the list
	l.PushBack(node)

	// Move the node to the front
	l.MoveToFront(node)

	// Verify the list state
	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node, l.Front(), "front node should be the single node")
	assert.Equal(t, node, l.Back(), "back node should be the single node")

	// Verify the node order
	assert.Nil(t, node.Prev, "node prev should be nil")
	assert.Nil(t, node.Next, "node next should be nil")
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

	printListNodeValues(l)

	// Move the invalid node to the front
	l.MoveToFront(node4)

	printListNodeValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 3")
	assert.Equal(t, node4, l.Front(), "front node should be node4")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node1, node4.Next, "node4 next should be node1")
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Equal(t, node1, node2.Prev, "node2 prev should be node1")
	assert.Equal(t, node3, node2.Next, "node2 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")

}

func TestList_MoveToFront_EmptyList(t *testing.T) {
	l := New()

	// Create a node
	node := &Node{Value: 1}

	printListNodeValues(l)

	// Move the node to the front (should have no effect)
	l.MoveToFront(node)

	printListNodeValues(l)

	// Verify the list state
	assert.Equal(t, int64(1), l.Len(), "list length should be 0")
	assert.Equal(t, node, l.Front(), "front node should be the node")
	assert.Equal(t, node, l.Back(), "back node should be the node")

	// Verify the node order
	assert.Nil(t, node.Prev, "node prev should be nil")
	assert.Nil(t, node.Next, "node next should be nil")
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

	printListNodeValues(l)

	// Move node2 to the back
	l.MoveToBack(node2)

	printListNodeValues(l)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	// Verify the node order
	assert.Equal(t, node3, node1.Next, "node1 next should be node3")
	assert.Equal(t, node3, node2.Prev, "node2 prev should be node3")
	assert.Equal(t, node1, node3.Prev, "node3 prev should be node1")
	assert.Nil(t, node2.Next, "node2 next should be nil")
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

	printListNodeValues(l)

	// Move node3 to the back
	l.MoveToBack(node3)

	printListNodeValues(l)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node2")

	// Verify the node order
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Equal(t, node1, node2.Prev, "node2 prev should be node1")
	assert.Equal(t, node3, node2.Next, "node2 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")
}

func TestList_MoveToBack_SingleNode(t *testing.T) {
	l := New()

	// Create a single node
	node := &Node{Value: 1}

	// Push the node to the list
	l.PushBack(node)

	// Move the node to the back
	l.MoveToBack(node)

	// Verify the list state
	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node, l.Front(), "front node should be the node")
	assert.Equal(t, node, l.Back(), "back node should be the node")

	// Verify the node order
	assert.Nil(t, node.Next, "node next should be nil")
	assert.Nil(t, node.Prev, "node prev should be nil")
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

	printListNodeValues(l)

	// Move the invalid node to the back
	l.MoveToBack(node4)

	printListNodeValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node4, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Equal(t, node1, node2.Prev, "node2 prev should be node1")
	assert.Equal(t, node2, node3.Prev, "node3 prev should be node2")
	assert.Equal(t, node4, node3.Next, "node3 next should be node4")
	assert.Nil(t, node4.Next, "node4 next should be nil")
}

func TestList_MoveToBack_EmptyList(t *testing.T) {
	l := New()

	// Create a node
	node := &Node{Value: 1}

	// Move the node to the front (should have no effect)
	l.MoveToBack(node)

	// Verify the list state
	assert.Equal(t, int64(1), l.Len(), "list length should be 0")
	assert.Equal(t, node, l.Front(), "front node should be the node")
	assert.Equal(t, node, l.Back(), "back node should be the node")

	// Verify the node order
	assert.Nil(t, node.Prev, "node prev should be nil")
	assert.Nil(t, node.Next, "node next should be nil")
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

	printListNodeValues(l)

	// Insert newNode before node2
	l.InsertBefore(node4, node2)

	printListNodeValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node4, node1.Next, "node1 next should be newNode")
	assert.Equal(t, node2, node4.Next, "newNode next should be node2")
	assert.Equal(t, node4, node2.Prev, "node2 prev should be newNode")
	assert.Equal(t, node1, node4.Prev, "newNode prev should be node1")
	assert.Nil(t, node3.Next, "node3 next should be nil")
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

	printListNodeValues(l)

	// Insert newNode before node1
	l.InsertBefore(node4, node1)

	printListNodeValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node4, l.Front(), "front node should be newNode")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node1, node4.Next, "newNode next should be node1")
	assert.Equal(t, node4, node1.Prev, "node1 prev should be newNode")
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Equal(t, node3, node2.Next, "node2 next should be node3")
	assert.Equal(t, node2, node3.Prev, "node3 prev should be node2")
	assert.Nil(t, node3.Next, "node3 next should be nil")
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

	// Insert the invalid node before node2
	l.InsertBefore(node3, node2)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node3")

	// Verify the node order
	assert.Equal(t, node3, node1.Next, "node1 next should be node3")
	assert.Equal(t, node1, node3.Prev, "node3 prev should be node1")
	assert.Equal(t, node2, node3.Next, "node3 next should be node2")
	assert.Nil(t, node2.Next, "node2 next should be nil")

	// Create an invalid node
	node4 := &Node{Value: 4}

	// Insert the invalid node before node1
	l.InsertBefore(node4, node1)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 2")
	assert.Equal(t, node4, l.Front(), "front node should be node4")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	// Verify the node order
	assert.Equal(t, node1, node4.Next, "node4 next should be node1")
	assert.Equal(t, node4, node1.Prev, "node1 prev should be node4")
	assert.Equal(t, node3, node1.Next, "node1 next should be node3")
	assert.Equal(t, node1, node3.Prev, "node3 prev should be node1")
	assert.Equal(t, node2, node3.Next, "node3 next should be node2")
	assert.Nil(t, node2.Next, "node2 next should be nil")
}

func TestList_InsertBefore_EmptyList(t *testing.T) {
	l := New()

	// Create a node
	node := &Node{Value: 1}

	// Insert the node after nil (should have no effect)
	l.InsertBefore(node, nil)

	// Verify the list state
	assert.Equal(t, int64(0), l.Len(), "list length should be 1")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
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

	printListNodeValues(l)

	// Insert node4 after node1
	l.InsertAfter(node4, node1)

	printListNodeValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node4, node1.Next, "node1 next should be node4")
	assert.Equal(t, node2, node4.Next, "node4 next should be node2")
	assert.Equal(t, node4, node2.Prev, "node2 prev should be node4")
	assert.Equal(t, node1, node4.Prev, "node4 prev should be node1")
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

	// Insert node3 after node2
	l.InsertAfter(node3, node2)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Equal(t, node3, node2.Next, "node2 next should be node3")
	assert.Nil(t, node3.Next, "node3 next should be nil")
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

	// Insert the invalid node after node2
	l.InsertAfter(node3, node1)

	// Verify the list state
	assert.Equal(t, int64(3), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node2")
	assert.Equal(t, node3, node1.Next, "node1 next should be node3")
	assert.Equal(t, node1, node3.Prev, "node3 prev should be node2")
	assert.Equal(t, node2, node3.Next, "node3 next should be node2")
	assert.Nil(t, node2.Next, "node2 next should be nil")

	// Create an invalid node
	node4 := &Node{Value: 4}

	// Insert the invalid node after node2
	l.InsertAfter(node4, node2)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node4, l.Back(), "back node should be node4")
	assert.Equal(t, node3, node1.Next, "node1 next should be node3")
	assert.Equal(t, node1, node3.Prev, "node3 prev should be node1")
	assert.Equal(t, node2, node3.Next, "node3 next should be node2")
	assert.Equal(t, node4, node2.Next, "node2 next should be node4")
	assert.Equal(t, node2, node4.Prev, "node4 prev should be node2")
	assert.Equal(t, node4, node2.Next, "node2 next should be node4")
	assert.Nil(t, node4.Next, "node4 next should be nil")

}

func TestList_InsertAfter_EmptyList(t *testing.T) {
	l := New()

	// Create a node
	node := &Node{Value: 1}

	// Insert the node after nil (should have no effect)
	l.InsertAfter(node, nil)

	// Verify the list state
	assert.Equal(t, int64(0), l.Len(), "list length should be 1")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
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

	printListNodeValues(l)

	// Swap node2 and node3
	l.Swap(node2, node3)

	printListNodeValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node4, l.Back(), "back node should be node4")

	// Verify the node order
	assert.Equal(t, node3, node1.Next, "node1 next should be node3")
	assert.Equal(t, node2, node3.Next, "node3 next should be node2")
	assert.Equal(t, node4, node2.Next, "node2 next should be node4")
	assert.Nil(t, node4.Next, "node4 next should be nil")

	assert.Equal(t, node1, node3.Prev, "node3 prev should be node1")
	assert.Equal(t, node3, node2.Prev, "node2 prev should be node3")
	assert.Equal(t, node2, node4.Prev, "node4 prev should be node2")
	assert.Nil(t, node1.Prev, "node2 prev should be nil")

	// Swap node1 and node4
	l.Swap(node1, node4)

	printListNodeValues(l)

	// Verify the list state
	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node4, l.Front(), "front node should be node4")
	assert.Equal(t, node1, l.Back(), "back node should be node1")

	// Verify the node order
	assert.Equal(t, node3, node4.Next, "node4 next should be node3")
	assert.Equal(t, node2, node3.Next, "node3 next should be node2")
	assert.Equal(t, node1, node2.Next, "node2 next should be node1")
	assert.Nil(t, node1.Next, "node1 next should be nil")

	assert.Equal(t, node4, node3.Prev, "node3 prev should be node4")
	assert.Equal(t, node3, node2.Prev, "node2 prev should be node3")
	assert.Equal(t, node2, node1.Prev, "node1 prev should be node2")
	assert.Nil(t, node4.Prev, "node2 prev should be nil")
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

	printListNodeValues(l)

	// Swap the invalid node with node2
	l.Swap(node3, node2)

	printListNodeValues(l)

	// Verify the list state
	assert.Equal(t, int64(2), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	// Verify the node order
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Nil(t, node2.Next, "node2 next should be nil")
	assert.Nil(t, node1.Prev, "node1 prev should be nil")

	// Swap the invalid node with node1
	l.Swap(node3, node1)

	// Verify the list state
	assert.Equal(t, int64(2), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	// Verify the node order
	assert.Equal(t, node2, node1.Next, "node1 next should be node2")
	assert.Nil(t, node2.Next, "node2 next should be nil")
	assert.Nil(t, node1.Prev, "node1 prev should be nil")
}

func TestList_Swap_EmptyList(t *testing.T) {
	l := New()

	// Create a node
	node := &Node{Value: 1}

	// Swap the node with nil (should have no effect)
	l.Swap(node, nil)

	// Verify the list state
	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
}

func BenchmarkList_PushBack(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PushBack(nodes[i])
	}
}

func BenchmarkList_PushFront(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PushFront(nodes[i])
	}
}

func BenchmarkList_PopBack(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
		l.PushBack(nodes[i])
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PopBack()
	}
}

func BenchmarkList_PopFront(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
		l.PushBack(nodes[i])
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PopFront()
	}
}

func BenchmarkList_InsertBefore(b *testing.B) {
	l := New()
	n := &Node{Value: int64(-1)}
	l.PushBack(n)
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.InsertBefore(nodes[i], n)
	}
}

func BenchmarkList_InsertAfter(b *testing.B) {
	l := New()
	n := &Node{Value: int64(-1)}
	l.PushBack(n)
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.InsertAfter(nodes[i], n)
	}
}

func BenchmarkList_Remove(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
		l.PushBack(nodes[i])
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Remove(nodes[i])
	}
}

func BenchmarkList_MoveToFront(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
		l.PushBack(nodes[i])
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.MoveToFront(nodes[i])
	}
}

func BenchmarkList_MoveToBack(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
		l.PushBack(nodes[i])
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.MoveToBack(nodes[i])
	}
}

func BenchmarkList_Swap(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
		l.PushBack(nodes[i])
	}

	b.ResetTimer()

	for i := 0; i < b.N-1; i++ {
		l.Swap(nodes[i], nodes[i+1])
	}
}