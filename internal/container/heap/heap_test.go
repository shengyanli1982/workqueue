package heap

import (
	"fmt"
	"testing"

	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
	"github.com/stretchr/testify/assert"
)

func PrintListIndexs(l *lst.List) {
	fmt.Println("List indexs: ============================")
	for i := l.Front(); i != nil; i = i.Next {
		fmt.Printf("Index: %v, Priority: %v\n", i.Index, i.Priority)
	}
}

func PrintNodeIndexs(nodes []*lst.Node) {
	fmt.Println("Node indexs: ============================")
	for _, n := range nodes {
		fmt.Printf("Index: %v, Priority: %v\n", n.Index, n.Priority)
	}
}

func TestHeap_Remove(t *testing.T) {
	h := New()
	count := 4
	nodes := make([]*lst.Node, count)

	// Push the nodes
	for i := 0; i < count; i++ {
		n := &lst.Node{Priority: int32(count - i - 1)}
		nodes[i] = n
		h.Push(n)
	}

	// Print the node indexs
	PrintNodeIndexs(nodes)

	// Print the list indexs
	PrintListIndexs(h.list)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Front().Priority, "front value should be 0")
	assert.Equal(t, int64(1), h.Front().Next.Priority, "front next value should be 1")
	assert.Equal(t, int64(2), h.Front().Next.Next.Priority, "front next next value should be 2")
	assert.Equal(t, int64(3), h.Front().Next.Next.Next.Priority, "front next next next value should be 3")

	// Remove the nodes
	for i := 0; i < count; i++ {
		h.Remove(nodes[i])
	}

	// Verify the heap state
	assert.Equal(t, int64(0), h.Len(), "heap length should be 0")
	assert.Nil(t, h.Front(), "front value should be nil")
	assert.Nil(t, h.Back(), "back value should be nil")
}

func TestHeap_Push(t *testing.T) {
	h := New()
	count := 4

	// Push the nodes
	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int32(count - i - 1)})
	}

	// Print the list indexs
	PrintListIndexs(h.list)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, uint64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, uint64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, uint64(0), h.Front().Priority, "front value should be 0")
	assert.Equal(t, uint64(1), h.Front().Next.Priority, "front next value should be 1")
	assert.Equal(t, uint64(2), h.Front().Next.Next.Priority, "front next next value should be 2")
	assert.Equal(t, uint64(3), h.Front().Next.Next.Next.Priority, "front next next next value should be 3")
}

func TestHeap_Push_Duplicate(t *testing.T) {
	h := New()

	// node1 := &lst.Node{Priority: int32(2)}
	node2 := &lst.Node{Priority: int32(3)}
	node3 := &lst.Node{Priority: int32(2)}
	// node4 := &lst.Node{Priority: int32(3)}

	// h.Push(node1)
	h.Push(node2)
	h.Push(node3)
	// h.Push(node4)

	// assert.Equal(t, node1, node3)
	// h.Push(&lst.Node{Priority: int32(3)})

	// // Push the nodes
	// for i := 0; i < count; i++ {
	// 	h.Push(&lst.Node{Priority: int32(count - i - 1)})
	// }

	// // Push the nodes
	// for i := 0; i < count; i++ {
	// 	h.Push(&lst.Node{Priority: int32(count - i - 1)})
	// }

	// // Push the nodes
	// for i := 0; i < count; i++ {
	// 	h.Push(&lst.Node{Priority: int32(count - i - 1)})
	// }

	// // Print the list indexs
	// PrintListIndexs(h.list)

	fmt.Printf("Last node next0: %v\n", h.Back().Next)
	fmt.Printf("Last node next1: %v\n", h.Back().Next)
	fmt.Printf("Last node next2: %v\n", h.Back().Next)
	fmt.Printf("Last node next3: %v\n", h.Back().Next)

	// // Verify the heap state
	// assert.Equal(t, int64(count*3), h.Len(), fmt.Sprintf("heap length should be %d", count*3))
	// assert.Equal(t, uint64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	// assert.Equal(t, uint64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// // Verify the heap order
	// assert.Equal(t, uint64(0), h.Front().Priority, "front value should be 0")
	// assert.Equal(t, uint64(0), h.Front().Next.Priority, "front next value should be 0")
	// assert.Equal(t, uint64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))
	// assert.Equal(t, uint64(count-1), h.Back().Prev.Priority, fmt.Sprintf("back value should be %d", count-1))
}

func TestHeap_Pop(t *testing.T) {
	h := New()
	count := 4

	// Push the nodes
	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int32(count - i - 1)})
	}

	// Print the list indexs
	PrintListIndexs(h.list)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, uint64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, uint64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Pop the nodes
	for i := 0; i < count; i++ {
		n := h.Pop()
		assert.NotNil(t, n, "pop value should not be nil")
		assert.Equal(t, uint64(i), n.Priority, fmt.Sprintf("pop value should be %d", i))
		// assert.Equal(t, int64(0), n.Index, fmt.Sprintf("pop index should be %d", 0))
	}

	// Print the list indexs
	PrintListIndexs(h.list)

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
		h.Push(&lst.Node{Priority: int32(count - i - 1)})
	}

	// Print the list indexs
	PrintListIndexs(h.list)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int64(0), h.Front().Priority, "front value should be 0")
	assert.Equal(t, int64(1), h.Front().Next.Priority, "front next value should be 1")
	assert.Equal(t, int64(2), h.Front().Next.Next.Priority, "front next next value should be 2")
	assert.Equal(t, int64(3), h.Front().Next.Next.Next.Priority, "front next next next value should be 3")

	// Pop the node
	n := h.Pop()
	assert.NotNil(t, n, "pop value should not be nil")
	assert.Equal(t, int64(0), n.Priority, "pop value should be 0")

	// Print the list indexs
	PrintListIndexs(h.list)

	// Verify the heap state
	assert.Equal(t, int64(count-1), h.Len(), fmt.Sprintf("heap length should be %d", count-1))
	assert.Equal(t, int64(1), h.Front().Priority, fmt.Sprintf("front value should be %d", 1))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Push the node
	h.Push(&lst.Node{Priority: int32(0)})

	// Print the list indexs
	PrintListIndexs(h.list)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Push the node
	h.Push(&lst.Node{Priority: int32(2)})

	// Print the list indexs
	PrintListIndexs(h.list)

	// Verify the heap state
	assert.Equal(t, int64(count+1), h.Len(), fmt.Sprintf("heap length should be %d", count+1))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Push the node
	h.Push(&lst.Node{Priority: int32(count)})

	// Print the list indexs
	PrintListIndexs(h.list)

	// Verify the heap state
	assert.Equal(t, int64(count+2), h.Len(), fmt.Sprintf("heap length should be %d", count+2))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count), h.Back().Priority, fmt.Sprintf("back value should be %d", count))
}
