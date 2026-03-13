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
		fmt.Printf(">> priority: %d, value: %v, left: %v, right: %v\n", n.Priority, n.Value, n.Left, n.Right)
		PrintOrderTraversalIndexs(n.Right)
	}
}

func PrintNodeIndexs(nodes []*lst.Node) {
	for _, n := range nodes {
		fmt.Printf("# priority: %v\n", n.Priority)
	}
}

func TestHeap_Push(t *testing.T) {
	h := New()
	count := 4

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(i), Value: i})
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front priority should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))

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

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(count - i - 1)})
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front priority should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))

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

	h.Push(&lst.Node{Priority: int64(2)})
	h.Push(&lst.Node{Priority: int64(0)})
	h.Push(&lst.Node{Priority: int64(1)})
	h.Push(&lst.Node{Priority: int64(3)})

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

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

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(count - i - 1)})
	}

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(i)})
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count*2), h.Len(), fmt.Sprintf("heap length should be %d", count*2))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

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

	for i := 0; i < 10; i++ {
		h.Push(nil)
	}

	assert.Equal(t, int64(0), h.Len(), "heap length should be 0")
	assert.Nil(t, h.Front(), "front value should be nil")
	assert.Nil(t, h.Back(), "back value should be nil")
}

func TestHeap_Pop(t *testing.T) {
	h := New()
	count := 4

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(count - i - 1)})
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	fristNode := h.Front()

	popNode := h.Pop()

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.NotNil(t, fristNode, "first node should not be nil")
	assert.NotNil(t, popNode, "pop node should not be nil")
	assert.Equal(t, fristNode, popNode, "first node should be equal to pop node")
}

func TestHeap_PopAll(t *testing.T) {
	h := New()
	count := 4

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(count - i - 1)})
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	for i := 0; i < count; i++ {
		n := h.Pop()
		assert.NotNil(t, n, "pop value should not be nil")
		assert.Equal(t, int64(i), n.Priority, fmt.Sprintf("pop value should be %d", i))
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(0), h.Len(), "heap length should be 0")
	assert.Nil(t, h.Front(), "front value should be nil")
	assert.Nil(t, h.Back(), "back value should be nil")
}

func TestHeap_PopEmpty(t *testing.T) {
	h := New()

	n := h.Pop()
	assert.Nil(t, n, "pop value should be nil")
}

func TestHeap_PutAndPop_Intersect(t *testing.T) {
	h := New()
	count := 4

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(count - i - 1)})
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	assert.Equal(t, int64(0), h.Front().Priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(2), h.Root().Priority, "root priority should be 2")
	assert.Equal(t, int64(1), h.Root().Left.Priority, "root left priority should be 1")
	assert.Equal(t, int64(0), h.Root().Left.Left.Priority, "root left left priority should be 0")
	assert.Equal(t, int64(3), h.Root().Right.Priority, "root right priority should be 0")

	n := h.Pop()
	assert.NotNil(t, n, "pop value should not be nil")
	assert.Equal(t, int64(0), n.Priority, "pop value should be 0")

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count-1), h.Len(), fmt.Sprintf("heap length should be %d", count-1))
	assert.Equal(t, int64(1), h.Front().Priority, fmt.Sprintf("front value should be %d", 1))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	h.Push(&lst.Node{Priority: int64(0)})

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	assert.Equal(t, int64(0), h.Root().Left.Left.Priority, "root left left priority should be 0")

	h.Push(&lst.Node{Priority: int64(2)})

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count+1), h.Len(), fmt.Sprintf("heap length should be %d", count+1))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	assert.Equal(t, int64(2), h.Root().Right.Left.Priority, "root right left priority should be 2")

	h.Push(&lst.Node{Priority: int64(count)})

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count+2), h.Len(), fmt.Sprintf("heap length should be %d", count+2))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count), h.Back().Priority, fmt.Sprintf("back value should be %d", count))

	assert.Equal(t, int64(count), h.Root().Right.Right.Priority, "root right right priority should be 4")
}

func TestHeap_Remove(t *testing.T) {
	h := New()
	count := 4
	nodes := make([]*lst.Node, count)

	for i := 0; i < count; i++ {
		n := &lst.Node{Priority: int64(count - i - 1)}
		nodes[i] = n
		h.Push(n)
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	assert.Equal(t, int64(0), h.Front().Priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(2), h.Root().Priority, "root priority should be 2")
	assert.Equal(t, int64(1), h.Root().Left.Priority, "root left priority should be 1")
	assert.Equal(t, int64(0), h.Root().Left.Left.Priority, "root left left priority should be 0")
	assert.Equal(t, int64(3), h.Root().Right.Priority, "root right priority should be 0")

	for i := 0; i < count; i++ {
		h.Remove(nodes[i])
	}

	assert.Equal(t, int64(0), h.Len(), "heap length should be 0")
	assert.Nil(t, h.Front(), "front value should be nil")
	assert.Nil(t, h.Back(), "back value should be nil")
}

func TestHeap_ExtremeValues(t *testing.T) {
	h := New()

	h.Push(&lst.Node{Priority: int64(9223372036854775807)})
	h.Push(&lst.Node{Priority: int64(-9223372036854775808)})
	h.Push(&lst.Node{Priority: 0})

	assert.Equal(t, int64(-9223372036854775808), h.Front().Priority, "front should be MinInt64")
	assert.Equal(t, int64(9223372036854775807), h.Back().Priority, "back should be MaxInt64")
	assert.Equal(t, int64(3), h.Len(), "heap should contain 3 elements")
}

func TestHeap_NegativePriorities(t *testing.T) {
	h := New()

	priorities := []int64{-1, -5, -3, -2, -4}
	for _, p := range priorities {
		h.Push(&lst.Node{Priority: p})
	}

	assert.Equal(t, int64(-5), h.Front().Priority, "front should be -5")
	assert.Equal(t, int64(-1), h.Back().Priority, "back should be -1")

	expected := []int64{-5, -4, -3, -2, -1}
	for _, exp := range expected {
		node := h.Pop()
		assert.Equal(t, exp, node.Priority, fmt.Sprintf("expected priority %d", exp))
	}
}

func TestHeap_DuplicatePriorities(t *testing.T) {
	h := New()

	count := 5
	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: 1, Value: i})
	}

	assert.Equal(t, int64(count), h.Len(), "heap should contain 5 elements")
	assert.Equal(t, int64(1), h.Front().Priority, "front priority should be 1")
	assert.Equal(t, int64(1), h.Back().Priority, "back priority should be 1")

	for i := 0; i < count; i++ {
		node := h.Pop()
		assert.Equal(t, int64(1), node.Priority, "all nodes should have priority 1")
	}

	assert.Equal(t, int64(0), h.Len(), "heap should be empty")
}

func TestHeap_RemoveNilAndInvalid(t *testing.T) {
	h := New()

	h.Remove(nil)
	assert.Equal(t, int64(0), h.Len(), "heap should be empty after removing nil")

	validNode := &lst.Node{Priority: 1}
	h.Push(validNode)
	assert.Equal(t, int64(1), h.Len(), "heap should contain one node")

	nonExistentNode := &lst.Node{Priority: 999}
	h.Remove(nonExistentNode)
	assert.Equal(t, int64(1), h.Len(), "heap should still contain the valid node")

	h.Remove(validNode)
	assert.Equal(t, int64(0), h.Len(), "heap should be empty after removing valid node")
}

func TestHeap_LargeDataSet(t *testing.T) {
	h := New()
	count := 1000

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(i)})
	}

	assert.Equal(t, int64(count), h.Len(), "heap should contain 1000 elements")
	assert.Equal(t, int64(0), h.Front().Priority, "front should have minimum priority")
	assert.Equal(t, int64(count-1), h.Back().Priority, "back should have maximum priority")

	prev := h.Pop()
	for i := 1; i < count; i++ {
		current := h.Pop()
		assert.True(t, prev.Priority <= current.Priority, "heap order property should be maintained")
		prev = current
	}
}
