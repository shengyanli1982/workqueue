package heap

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func PrintRootIndexs(h *RBTree) {
	fmt.Printf("# root: %v\n", h.root)
}

func PrintOrderTraversalIndexs(n *Node) {
	if n != nil {
		PrintOrderTraversalIndexs(n.left)
		fmt.Printf(">> priority: %d, value: %v, left: %v, right: %v\n", n.priority, n.Value, n.left, n.right)
		PrintOrderTraversalIndexs(n.right)
	}
}

func PrintNodeIndexs(nodes []*Node) {
	for _, n := range nodes {
		fmt.Printf("# priority: %v\n", n.priority)
	}
}

func TestHeap_Push(t *testing.T) {
	h := New()
	count := 4

	// Push the nodes
	for i := 0; i < count; i++ {
		h.Push(&Node{priority: int64(i), Value: i})
	}

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().priority, fmt.Sprintf("front priority should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().priority, fmt.Sprintf("back priority should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Front().priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(1), h.Root().priority, "root priority should be 1")
	assert.Equal(t, int64(0), h.Root().left.priority, "root left priority should be 0")
	assert.Equal(t, int64(2), h.Root().right.priority, "root right priority should be 2")
	assert.Equal(t, int64(3), h.Root().right.right.priority, "root right right priority should be 3")
}

func TestHeap_Push_Reverse(t *testing.T) {
	h := New()
	count := 4

	// Push the nodes
	for i := 0; i < count; i++ {
		h.Push(&Node{priority: int64(count - i - 1)})
	}

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().priority, fmt.Sprintf("front priority should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().priority, fmt.Sprintf("back priority should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Front().priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(2), h.Root().priority, "root priority should be 2")
	assert.Equal(t, int64(1), h.Root().left.priority, "root left priority should be 1")
	assert.Equal(t, int64(0), h.Root().left.left.priority, "root left left priority should be 0")
	assert.Equal(t, int64(3), h.Root().right.priority, "root right priority should be 0")
}

func TestHeap_Push_Random(t *testing.T) {
	h := New()
	count := 4

	// Push the nodes
	h.Push(&Node{priority: int64(2)})
	h.Push(&Node{priority: int64(0)})
	h.Push(&Node{priority: int64(1)})
	h.Push(&Node{priority: int64(3)})

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Front().priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(1), h.Root().priority, "root priority should be 1")
	assert.Equal(t, int64(0), h.Root().left.priority, "root left priority should be 0")
	assert.Equal(t, int64(2), h.Root().right.priority, "root right priority should be 2")
	assert.Equal(t, int64(3), h.Root().right.right.priority, "root right right priority should be 3")
}

func TestHeap_Push_Duplicate(t *testing.T) {
	h := New()
	count := 4

	// Push the nodes
	for i := 0; i < count; i++ {
		h.Push(&Node{priority: int64(count - i - 1)})
	}

	// Push the nodes
	for i := 0; i < count; i++ {
		h.Push(&Node{priority: int64(i)})
	}

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count*2), h.Len(), fmt.Sprintf("heap length should be %d", count*2))
	assert.Equal(t, int64(0), h.Front().priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Front().priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(2), h.Root().priority, "root priority should be 2")
	assert.Equal(t, int64(0), h.Root().left.priority, "root left priority should be 0")
	assert.Equal(t, int64(3), h.Root().right.priority, "root right priority should be 3")
	assert.Equal(t, int64(0), h.Root().left.left.priority, "root left left priority should be 0")
	assert.Equal(t, int64(1), h.Root().left.right.priority, "root left right priority should be 1")
	assert.Equal(t, int64(2), h.Root().right.left.priority, "root right left priority should be 2")
	assert.Equal(t, int64(3), h.Root().right.right.priority, "root right right priority should be 3")
	assert.Equal(t, int64(1), h.Root().left.right.right.priority, "root left right priority should be 1")
}

func TestHeap_Push_Nil(t *testing.T) {
	h := New()

	// Push the nil node
	for i := 0; i < 10; i++ {
		h.Push(nil)
	}

	// Verify the heap state
	assert.Equal(t, int64(0), h.Len(), "heap length should be 0")
	assert.Nil(t, h.Front(), "front value should be nil")
	assert.Nil(t, h.Back(), "back value should be nil")
}

func TestHeap_Pop(t *testing.T) {
	h := New()
	count := 4

	// Push the nodes
	for i := 0; i < count; i++ {
		h.Push(&Node{priority: int64(count - i - 1)})
	}

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().priority, fmt.Sprintf("back value should be %d", count-1))

	// Pop the nodes
	for i := 0; i < count; i++ {
		n := h.Pop()
		assert.NotNil(t, n, "pop value should not be nil")
		assert.Equal(t, int64(i), n.priority, fmt.Sprintf("pop value should be %d", i))
	}

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(0), h.Len(), "heap length should be 0")
	assert.Nil(t, h.Front(), "front value should be nil")
	assert.Nil(t, h.Back(), "back value should be nil")
}

func TestHeap_PopEmpty(t *testing.T) {
	h := New()

	// Pop the empty heap
	n := h.Pop()
	assert.Nil(t, n, "pop value should be nil")
}

func TestHeap_PutAndPop_Intersect(t *testing.T) {
	h := New()
	count := 4

	// Push the nodes
	for i := 0; i < count; i++ {
		h.Push(&Node{priority: int64(count - i - 1)})
	}

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Front().priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(2), h.Root().priority, "root priority should be 2")
	assert.Equal(t, int64(1), h.Root().left.priority, "root left priority should be 1")
	assert.Equal(t, int64(0), h.Root().left.left.priority, "root left left priority should be 0")
	assert.Equal(t, int64(3), h.Root().right.priority, "root right priority should be 0")

	// Pop the node
	n := h.Pop()
	assert.NotNil(t, n, "pop value should not be nil")
	assert.Equal(t, int64(0), n.priority, "pop value should be 0")

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count-1), h.Len(), fmt.Sprintf("heap length should be %d", count-1))
	assert.Equal(t, int64(1), h.Front().priority, fmt.Sprintf("front value should be %d", 1))
	assert.Equal(t, int64(count-1), h.Back().priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order

	// Push the node
	h.Push(&Node{priority: int64(0)})

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Root().left.left.priority, "root left left priority should be 0")

	// Push the node
	h.Push(&Node{priority: int64(2)})

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count+1), h.Len(), fmt.Sprintf("heap length should be %d", count+1))
	assert.Equal(t, int64(0), h.Front().priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(2), h.Root().right.left.priority, "root right left priority should be 2")

	// Push the node
	h.Push(&Node{priority: int64(count)})

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count+2), h.Len(), fmt.Sprintf("heap length should be %d", count+2))
	assert.Equal(t, int64(0), h.Front().priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count), h.Back().priority, fmt.Sprintf("back value should be %d", count))

	// Verify the heap order
	assert.Equal(t, int64(count), h.Root().right.right.priority, "root right right priority should be 4")
}

func TestHeap_Remove(t *testing.T) {
	h := New()
	count := 4
	nodes := make([]*Node, count)

	// Push the nodes
	for i := 0; i < count; i++ {
		n := &Node{priority: int64(count - i - 1)}
		nodes[i] = n
		h.Push(n)
	}

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Front().priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(2), h.Root().priority, "root priority should be 2")
	assert.Equal(t, int64(1), h.Root().left.priority, "root left priority should be 1")
	assert.Equal(t, int64(0), h.Root().left.left.priority, "root left left priority should be 0")
	assert.Equal(t, int64(3), h.Root().right.priority, "root right priority should be 0")

	// Remove the nodes
	for i := 0; i < count; i++ {
		h.Remove(nodes[i])
	}

	// Verify the heap state
	assert.Equal(t, int64(0), h.Len(), "heap length should be 0")
	assert.Nil(t, h.Front(), "front value should be nil")
	assert.Nil(t, h.Back(), "back value should be nil")
}
