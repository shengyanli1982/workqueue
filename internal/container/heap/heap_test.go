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
		fmt.Printf("Index: %v, Priority: %v, Value: %v\n", i.Index, i.Priority, i.Value)
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
	assert.Equal(t, int32(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int32(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int32(0), h.Front().Priority, "front value should be 0")
	assert.Equal(t, int32(1), h.Front().Next.Priority, "front next value should be 1")
	assert.Equal(t, int32(2), h.Front().Next.Next.Priority, "front next next value should be 2")
	assert.Equal(t, int32(3), h.Front().Next.Next.Next.Priority, "front next next next value should be 3")

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
		h.Push(&lst.Node{Priority: int32(i)})
	}

	// Print the list indexs
	PrintListIndexs(h.list)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int32(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int32(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int32(0), h.Front().Priority, "front value should be 0")
	assert.Equal(t, int32(1), h.Front().Next.Priority, "front next value should be 1")
	assert.Equal(t, int32(2), h.Front().Next.Next.Priority, "front next next value should be 2")
	assert.Equal(t, int32(3), h.Front().Next.Next.Next.Priority, "front next next next value should be 3")
}

func TestHeap_Push_Reverse(t *testing.T) {
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
	assert.Equal(t, int32(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int32(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int32(0), h.Front().Priority, "front value should be 0")
	assert.Equal(t, int32(1), h.Front().Next.Priority, "front next value should be 1")
	assert.Equal(t, int32(2), h.Front().Next.Next.Priority, "front next next value should be 2")
	assert.Equal(t, int32(3), h.Front().Next.Next.Next.Priority, "front next next next value should be 3")
}

func TestHeap_Push_Random(t *testing.T) {
	h := New()
	count := 4

	// Push the nodes
	h.Push(&lst.Node{Priority: int32(2)})
	h.Push(&lst.Node{Priority: int32(0)})
	h.Push(&lst.Node{Priority: int32(1)})
	h.Push(&lst.Node{Priority: int32(3)})

	// Print the list indexs
	PrintListIndexs(h.list)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int32(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int32(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int32(0), h.Front().Priority, "front value should be 0")
	assert.Equal(t, int32(1), h.Front().Next.Priority, "front next value should be 1")
	assert.Equal(t, int32(2), h.Front().Next.Next.Priority, "front next next value should be 2")
	assert.Equal(t, int32(3), h.Front().Next.Next.Next.Priority, "front next next next value should be 3")
}

func TestHeap_Push_Duplicate(t *testing.T) {
	h := New()
	count := 4

	// Push the nodes
	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int32(count - i - 1)})
	}

	// Push the nodes
	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int32(i)})
	}

	// Print the list indexs
	PrintListIndexs(h.list)

	// Verify the heap state
	assert.Equal(t, int64(count*2), h.Len(), fmt.Sprintf("heap length should be %d", count*2))
	assert.Equal(t, int32(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int32(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int32(0), h.Front().Priority, "front value should be 0")
	assert.Equal(t, int32(0), h.Front().Next.Priority, "front next value should be 0")
	assert.Equal(t, int32(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))
	assert.Equal(t, int32(count-1), h.Back().Prev.Priority, fmt.Sprintf("back value should be %d", count-1))
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
	assert.Equal(t, int32(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int32(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Pop the nodes
	for i := 0; i < count; i++ {
		n := h.Pop()
		assert.NotNil(t, n, "pop value should not be nil")
		assert.Equal(t, int32(i), n.Priority, fmt.Sprintf("pop value should be %d", i))
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
	assert.Equal(t, int32(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int32(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Verify the heap order
	assert.Equal(t, int32(0), h.Front().Priority, "front value should be 0")
	assert.Equal(t, int32(1), h.Front().Next.Priority, "front next value should be 1")
	assert.Equal(t, int32(2), h.Front().Next.Next.Priority, "front next next value should be 2")
	assert.Equal(t, int32(3), h.Front().Next.Next.Next.Priority, "front next next next value should be 3")

	// Pop the node
	n := h.Pop()
	assert.NotNil(t, n, "pop value should not be nil")
	assert.Equal(t, int32(0), n.Priority, "pop value should be 0")

	// Print the list indexs
	PrintListIndexs(h.list)

	// Verify the heap state
	assert.Equal(t, int64(count-1), h.Len(), fmt.Sprintf("heap length should be %d", count-1))
	assert.Equal(t, int32(1), h.Front().Priority, fmt.Sprintf("front value should be %d", 1))
	assert.Equal(t, int32(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Push the node
	h.Push(&lst.Node{Priority: int32(0)})

	// Print the list indexs
	PrintListIndexs(h.list)

	// Verify the heap state
	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int32(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int32(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Push the node
	h.Push(&lst.Node{Priority: int32(2)})

	// Print the list indexs
	PrintListIndexs(h.list)

	// Verify the heap state
	assert.Equal(t, int64(count+1), h.Len(), fmt.Sprintf("heap length should be %d", count+1))
	assert.Equal(t, int32(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int32(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	// Push the node
	h.Push(&lst.Node{Priority: int32(count)})

	// Print the list indexs
	PrintListIndexs(h.list)

	// Verify the heap state
	assert.Equal(t, int64(count+2), h.Len(), fmt.Sprintf("heap length should be %d", count+2))
	assert.Equal(t, int32(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int32(count), h.Back().Priority, fmt.Sprintf("back value should be %d", count))
}
