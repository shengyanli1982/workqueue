package heap

import (
	"fmt"
	"testing"

	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
	"github.com/stretchr/testify/assert"
)

func PrintRootIndexs(h *RBTree) {
	fmt.Printf("# root: %v\n", h.root)
}

func PrintOrderTraversalIndexs(n *lst.Node) {
	if n != nil {
		PrintOrderTraversalIndexs(n.Left)
		fmt.Printf(">> Priority: %d, value: %v, left: %v, right: %v\n", n.Priority, n.Value, n.Left, n.Right)
		PrintOrderTraversalIndexs(n.Right)
	}
}

func PrintNodeIndexs(nodes []*lst.Node) {
	for _, n := range nodes {
		fmt.Printf("# Priority: %v\n", n.Priority)
	}
}

func TestHeap_Push(t *testing.T) {
	h := New()
	count := 4

	// Push the nodes
	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(i), Value: i})
	}

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front priority should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Front().Priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(1), h.Root().Priority, "root priority should be 1")
	assert.Equal(t, int64(0), h.Root().Left.Priority, "root left priority should be 0")
	assert.Equal(t, int64(2), h.Root().Right.Priority, "root right priority should be 2")
	assert.Equal(t, int64(3), h.Root().Right.Right.Priority, "root right right priority should be 3")
}

func TestHeap_Push_Reverse(t *testing.T) {
	h := New()
	count := 4

	// Push the nodes
	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(count - i - 1)})
	}

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front priority should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Front().Priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(2), h.Root().Priority, "root priority should be 2")
	assert.Equal(t, int64(1), h.Root().Left.Priority, "root left priority should be 1")
	assert.Equal(t, int64(0), h.Root().Left.Left.Priority, "root left left priority should be 0")
	assert.Equal(t, int64(3), h.Root().Right.Priority, "root right priority should be 0")
}

func TestHeap_Push_Random(t *testing.T) {
	h := New()
	count := 4

	// Push the nodes
	h.Push(&lst.Node{Priority: int64(2)})
	h.Push(&lst.Node{Priority: int64(0)})
	h.Push(&lst.Node{Priority: int64(1)})
	h.Push(&lst.Node{Priority: int64(3)})

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Front().Priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(1), h.Root().Priority, "root priority should be 1")
	assert.Equal(t, int64(0), h.Root().Left.Priority, "root left priority should be 0")
	assert.Equal(t, int64(2), h.Root().Right.Priority, "root right priority should be 2")
	assert.Equal(t, int64(3), h.Root().Right.Right.Priority, "root right right priority should be 3")
}

func TestHeap_Push_Duplicate(t *testing.T) {
	h := New()
	count := 4

	// Push the nodes
	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(count - i - 1)})
	}

	// Push the nodes
	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(i)})
	}

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count*2), h.Len(), fmt.Sprintf("heap length should be %d", count*2))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Front().Priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(2), h.Root().Priority, "root priority should be 2")
	assert.Equal(t, int64(0), h.Root().Left.Priority, "root left priority should be 0")
	assert.Equal(t, int64(3), h.Root().Right.Priority, "root right priority should be 3")
	assert.Equal(t, int64(0), h.Root().Left.Left.Priority, "root left left priority should be 0")
	assert.Equal(t, int64(1), h.Root().Left.Right.Priority, "root left right priority should be 1")
	assert.Equal(t, int64(2), h.Root().Right.Left.Priority, "root right left priority should be 2")
	assert.Equal(t, int64(3), h.Root().Right.Right.Priority, "root right right priority should be 3")
	assert.Equal(t, int64(1), h.Root().Left.Right.Right.Priority, "root left right priority should be 1")
}

func TestHeap_Push_Nil(t *testing.T) {
	h := New()

	// Push the nil lst.Node
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
		h.Push(&lst.Node{Priority: int64(count - i - 1)})
	}

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Get frist node
	fristNode := h.Front()

	// Pop the node
	popNode := h.Pop()

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the nodes
	assert.NotNil(t, fristNode, "first node should not be nil")
	assert.NotNil(t, popNode, "pop node should not be nil")
	assert.Equal(t, fristNode, popNode, "first node should be equal to pop node")
}

func TestHeap_PopAll(t *testing.T) {
	h := New()
	count := 4

	// Push the nodes
	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(count - i - 1)})
	}

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Pop the nodes
	for i := 0; i < count; i++ {
		n := h.Pop()
		assert.NotNil(t, n, "pop value should not be nil")
		assert.Equal(t, int64(i), n.Priority, fmt.Sprintf("pop value should be %d", i))
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
		h.Push(&lst.Node{Priority: int64(count - i - 1)})
	}

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Front().Priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(2), h.Root().Priority, "root priority should be 2")
	assert.Equal(t, int64(1), h.Root().Left.Priority, "root left priority should be 1")
	assert.Equal(t, int64(0), h.Root().Left.Left.Priority, "root left left priority should be 0")
	assert.Equal(t, int64(3), h.Root().Right.Priority, "root right priority should be 0")

	// Pop the lst.Node
	n := h.Pop()
	assert.NotNil(t, n, "pop value should not be nil")
	assert.Equal(t, int64(0), n.Priority, "pop value should be 0")

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count-1), h.Len(), fmt.Sprintf("heap length should be %d", count-1))
	assert.Equal(t, int64(1), h.Front().Priority, fmt.Sprintf("front value should be %d", 1))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order

	// Push the lst.Node
	h.Push(&lst.Node{Priority: int64(0)})

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Root().Left.Left.Priority, "root left left priority should be 0")

	// Push the lst.Node
	h.Push(&lst.Node{Priority: int64(2)})

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count+1), h.Len(), fmt.Sprintf("heap length should be %d", count+1))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(2), h.Root().Right.Left.Priority, "root right left priority should be 2")

	// Push the lst.Node
	h.Push(&lst.Node{Priority: int64(count)})

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count+2), h.Len(), fmt.Sprintf("heap length should be %d", count+2))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count), h.Back().Priority, fmt.Sprintf("back value should be %d", count))

	// Verify the heap order
	assert.Equal(t, int64(count), h.Root().Right.Right.Priority, "root right right priority should be 4")
}

func TestHeap_Remove(t *testing.T) {
	h := New()
	count := 4
	nodes := make([]*lst.Node, count)

	// Push the nodes
	for i := 0; i < count; i++ {
		n := &lst.Node{Priority: int64(count - i - 1)}
		nodes[i] = n
		h.Push(n)
	}

	// Prrint the tree order
	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Front().Priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(2), h.Root().Priority, "root priority should be 2")
	assert.Equal(t, int64(1), h.Root().Left.Priority, "root left priority should be 1")
	assert.Equal(t, int64(0), h.Root().Left.Left.Priority, "root left left priority should be 0")
	assert.Equal(t, int64(3), h.Root().Right.Priority, "root right priority should be 0")

	// Remove the nodes
	for i := 0; i < count; i++ {
		h.Remove(nodes[i])
	}

	// Verify the heap state
	assert.Equal(t, int64(0), h.Len(), "heap length should be 0")
	assert.Nil(t, h.Front(), "front value should be nil")
	assert.Nil(t, h.Back(), "back value should be nil")
}
