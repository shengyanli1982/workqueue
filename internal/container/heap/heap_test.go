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

func TestHeap_ExtremeValues(t *testing.T) {
	h := New()
	
	// 测试最大值和最小值
	h.Push(&lst.Node{Priority: int64(9223372036854775807)})  // math.MaxInt64
	h.Push(&lst.Node{Priority: int64(-9223372036854775808)}) // math.MinInt64
	h.Push(&lst.Node{Priority: 0})
	
	assert.Equal(t, int64(-9223372036854775808), h.Front().Priority, "front should be MinInt64")
	assert.Equal(t, int64(9223372036854775807), h.Back().Priority, "back should be MaxInt64")
	assert.Equal(t, int64(3), h.Len(), "heap should contain 3 elements")
}

func TestHeap_NegativePriorities(t *testing.T) {
	h := New()
	
	// 测试负数优先级
	priorities := []int64{-1, -5, -3, -2, -4}
	for _, p := range priorities {
		h.Push(&lst.Node{Priority: p})
	}
	
	// 验证最小堆特性
	assert.Equal(t, int64(-5), h.Front().Priority, "front should be -5")
	assert.Equal(t, int64(-1), h.Back().Priority, "back should be -1")
	
	// 按序弹出验证顺序
	expected := []int64{-5, -4, -3, -2, -1}
	for _, exp := range expected {
		node := h.Pop()
		assert.Equal(t, exp, node.Priority, fmt.Sprintf("expected priority %d", exp))
	}
}

func TestHeap_DuplicatePriorities(t *testing.T) {
	h := New()
	
	// 插入多个相同优先级的节点
	count := 5
	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: 1, Value: i})
	}
	
	assert.Equal(t, int64(count), h.Len(), "heap should contain 5 elements")
	assert.Equal(t, int64(1), h.Front().Priority, "front priority should be 1")
	assert.Equal(t, int64(1), h.Back().Priority, "back priority should be 1")
	
	// 删除所有节点，确保相同优先级的节点都能正确删除
	for i := 0; i < count; i++ {
		node := h.Pop()
		assert.Equal(t, int64(1), node.Priority, "all nodes should have priority 1")
	}
	
	assert.Equal(t, int64(0), h.Len(), "heap should be empty")
}

func TestHeap_RemoveNilAndInvalid(t *testing.T) {
	h := New()
	
	// 测试删除nil节点
	h.Remove(nil)
	assert.Equal(t, int64(0), h.Len(), "heap should be empty after removing nil")
	
	// 添加一个正常节点
	validNode := &lst.Node{Priority: 1}
	h.Push(validNode)
	assert.Equal(t, int64(1), h.Len(), "heap should contain one node")
	
	// 测试删除不存在的节点
	nonExistentNode := &lst.Node{Priority: 999}
	h.Remove(nonExistentNode)
	assert.Equal(t, int64(1), h.Len(), "heap should still contain the valid node")
	
	// 删除存在的节点
	h.Remove(validNode)
	assert.Equal(t, int64(0), h.Len(), "heap should be empty after removing valid node")
}

func TestHeap_LargeDataSet(t *testing.T) {
	h := New()
	count := 1000
	
	// 插入大量数据
	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(i)})
	}
	
	assert.Equal(t, int64(count), h.Len(), "heap should contain 1000 elements")
	assert.Equal(t, int64(0), h.Front().Priority, "front should have minimum priority")
	assert.Equal(t, int64(count-1), h.Back().Priority, "back should have maximum priority")
	
	// 验证堆的顺序性质
	prev := h.Pop()
	for i := 1; i < count; i++ {
		current := h.Pop()
		assert.True(t, prev.Priority <= current.Priority, "heap order property should be maintained")
		prev = current
	}
}
