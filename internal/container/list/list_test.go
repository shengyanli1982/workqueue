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

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")
}

func TestList_PushFront(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushFront(node1)
	l.PushFront(node2)
	l.PushFront(node3)

	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node3, l.Front(), "front node should be node3")
	assert.Equal(t, node1, l.Back(), "back node should be node1")

	assert.Equal(t, node2, node3.Right, "node3 next should be node2")
	assert.Equal(t, node1, node2.Right, "node2 next should be node1")
	assert.Nil(t, node1.Right, "node1 next should be nil")
}

func TestList_PopBack(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	poppedNode := l.PopBack()

	assert.Equal(t, node3, poppedNode, "popped node should be node3")

	assert.Equal(t, int64(2), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Nil(t, node2.Right, "node2 next should be nil")
}

func TestList_PopFront(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	poppedNode := l.PopFront()

	assert.Equal(t, node1, poppedNode, "popped node should be node1")

	assert.Equal(t, int64(2), l.Len(), "list length should be 2")
	assert.Equal(t, node2, l.Front(), "front node should be node2")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")
}

func TestList_PushAndPop(t *testing.T) {
	l := New()

	count := math.MaxUint16

	for i := 0; i < count; i++ {
		l.PushBack(&Node{Value: int64(i)})
	}

	assert.Equal(t, int64(count), l.Len(), fmt.Sprintf("list length should be %v", count))
	assert.Equal(t, int64(0), l.Front().Value, fmt.Sprintf("front node index should be %v", 0))
	assert.Equal(t, int64(count-1), l.Back().Value, fmt.Sprintf("back node index should be %v", count-1))

	for i := 0; i < count; i++ {
		node := l.PopFront()
		assert.Equal(t, int64(i), node.Value, fmt.Sprintf("popped node index should be %v", i))
	}

	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")

	nilNode := l.PopFront()
	assert.Nil(t, nilNode, "popped node should be nil")
}

func TestList_PushAndPop2(t *testing.T) {
	l := New()

	count := math.MaxUint16

	for i := 0; i < count; i++ {
		l.PushFront(&Node{Value: int64(i)})
	}

	assert.Equal(t, int64(count), l.Len(), fmt.Sprintf("list length should be %v", count))
	assert.Equal(t, int64(count-1), l.Front().Value, fmt.Sprintf("front node index should be %v", count-1))
	assert.Equal(t, int64(0), l.Back().Value, fmt.Sprintf("back node index should be %v", 0))

	for i := 0; i < count; i++ {
		node := l.PopBack()
		assert.Equal(t, int64(i), node.Value, fmt.Sprintf("popped node index should be %v", i))
	}

	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")

	nilNode := l.PopBack()
	assert.Nil(t, nilNode, "popped node should be nil")
}

func TestList_PushAndPopWithPool(t *testing.T) {
	l := New()
	pool := NewNodePool()

	count := math.MaxUint16

	for i := 0; i < count; i++ {
		n := pool.Get()
		n.Value = int64(i)
		l.PushBack(n)
	}

	assert.Equal(t, int64(count), l.Len(), fmt.Sprintf("list length should be %v", count))
	assert.Equal(t, int64(0), l.Front().Value, fmt.Sprintf("front node index should be %v", 0))
	assert.Equal(t, int64(count-1), l.Back().Value, fmt.Sprintf("back node index should be %v", count-1))

	for i := 0; i < count; i++ {
		node := l.PopFront()
		assert.Equal(t, int64(i), node.Value, fmt.Sprintf("popped node index should be %v", i))
		pool.Put(node)
	}

	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")

	nilNode := l.PopFront()
	assert.Nil(t, nilNode, "popped node should be nil")
}

func TestList_PushAndPopWithPool2(t *testing.T) {
	l := New()
	pool := NewNodePool()

	count := math.MaxUint16

	for i := 0; i < count; i++ {
		n := pool.Get()
		n.Value = int64(i)
		l.PushFront(n)
	}

	assert.Equal(t, int64(count), l.Len(), fmt.Sprintf("list length should be %v", count))
	assert.Equal(t, int64(count-1), l.Front().Value, fmt.Sprintf("front node index should be %v", count-1))
	assert.Equal(t, int64(0), l.Back().Value, fmt.Sprintf("back node index should be %v", 0))

	for i := 0; i < count; i++ {
		node := l.PopBack()
		assert.Equal(t, int64(i), node.Value, fmt.Sprintf("popped node index should be %v", i))
		pool.Put(node)
	}

	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")

	nilNode := l.PopBack()
	assert.Nil(t, nilNode, "popped node should be nil")
}

func TestList_Cleanup(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	l.Cleanup()

	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
}

func TestList_Remove(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	l.Remove(node2)

	assert.Equal(t, int64(2), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	assert.Equal(t, node3, node1.Right, "node1 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")

	l.Remove(node1)

	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node3, l.Front(), "front node should be node3")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	assert.Nil(t, node3.Right, "node3 next should be nil")

	l.Remove(node3)

	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
}

func TestList_Remove_InvalidNode(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	node4 := &Node{Value: 4}

	l.Remove(node4)

	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")
}

func TestList_Remove_EmptyList(t *testing.T) {
	l := New()

	node := &Node{Value: 1}
	node.parentRef = toUnsafePtr(l)

	l.Remove(node)

	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
}

func TestList_MoveToFront(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	PrintListValues(l)

	l.MoveToFront(node2)

	PrintListValues(l)

	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node2, l.Front(), "front node should be node2")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	assert.Equal(t, node1, node2.Right, "node2 next should be node1")
	assert.Equal(t, node2, node1.Left, "node1 prev should be node2")
	assert.Equal(t, node1, node3.Left, "node3 prev should be node1")
	assert.Nil(t, node2.Left, "node2 prev should be nil")
}

func TestList_MoveToFront_FirstNode(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	PrintListValues(l)

	l.MoveToFront(node1)

	PrintListValues(l)

	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node1, node2.Left, "node2 prev should be node1")
	assert.Equal(t, node2, node3.Left, "node3 prev should be node2")
	assert.Nil(t, node1.Left, "node1 prev should be nil")
}

func TestList_MoveToFront_LastNode(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	PrintListValues(l)

	l.MoveToFront(node3)

	PrintListValues(l)

	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node3, l.Front(), "front node should be node3")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	assert.Equal(t, node1, node3.Right, "node3 next should be node1")
	assert.Equal(t, node3, node1.Left, "node1 prev should be node3")
	assert.Equal(t, node1, node2.Left, "node2 prev should be node1")
	assert.Nil(t, node3.Left, "node3 prev should be nil")
}

func TestList_MoveToFront_SingleNode(t *testing.T) {
	l := New()

	node := &Node{Value: 1}

	l.PushBack(node)

	PrintListValues(l)

	l.MoveToFront(node)

	PrintListValues(l)

	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node, l.Front(), "front node should be the single node")
	assert.Equal(t, node, l.Back(), "back node should be the single node")

	assert.Nil(t, node.Left, "node prev should be nil")
	assert.Nil(t, node.Right, "node next should be nil")
}

func TestList_MoveToFront_InvalidNode(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	node4 := &Node{Value: 4}

	PrintListValues(l)

	l.MoveToFront(node4)

	PrintListValues(l)

	assert.Equal(t, int64(4), l.Len(), "list length should be 3")
	assert.Equal(t, node4, l.Front(), "front node should be node4")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	assert.Equal(t, node1, node4.Right, "node4 next should be node1")
	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node1, node2.Left, "node2 prev should be node1")
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")

}

func TestList_MoveToFront_EmptyList(t *testing.T) {
	l := New()

	node := &Node{Value: 1}

	PrintListValues(l)

	l.MoveToFront(node)

	PrintListValues(l)

	assert.Equal(t, int64(1), l.Len(), "list length should be 0")
	assert.Equal(t, node, l.Front(), "front node should be the node")
	assert.Equal(t, node, l.Back(), "back node should be the node")
	assert.Equal(t, node.parentRef, toUnsafePtr(l), "node parentRef should be the list")

	assert.Nil(t, node.Left, "node prev should be nil")
	assert.Nil(t, node.Right, "node next should be nil")
}

func TestList_MoveToBack(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	PrintListValues(l)

	l.MoveToBack(node2)

	PrintListValues(l)

	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	assert.Equal(t, node3, node1.Right, "node1 next should be node3")
	assert.Equal(t, node3, node2.Left, "node2 prev should be node3")
	assert.Equal(t, node1, node3.Left, "node3 prev should be node1")
	assert.Nil(t, node2.Right, "node2 next should be nil")
}

func TestList_MoveToBack_LastNode(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	PrintListValues(l)

	l.MoveToBack(node3)

	PrintListValues(l)

	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node2")

	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node1, node2.Left, "node2 prev should be node1")
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")
}

func TestList_MoveToBack_SingleNode(t *testing.T) {
	l := New()

	node := &Node{Value: 1}

	l.PushBack(node)

	PrintListValues(l)

	l.MoveToBack(node)

	PrintListValues(l)

	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node, l.Front(), "front node should be the node")
	assert.Equal(t, node, l.Back(), "back node should be the node")

	assert.Nil(t, node.Right, "node next should be nil")
	assert.Nil(t, node.Left, "node prev should be nil")
}

func TestList_MoveToBack_InvalidNode(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	node4 := &Node{Value: 4}

	PrintListValues(l)

	l.MoveToBack(node4)

	PrintListValues(l)

	assert.Equal(t, int64(4), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node4, l.Back(), "back node should be node3")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node1, node2.Left, "node2 prev should be node1")
	assert.Equal(t, node2, node3.Left, "node3 prev should be node2")
	assert.Equal(t, node4, node3.Right, "node3 next should be node4")
	assert.Nil(t, node4.Right, "node4 next should be nil")

}

func TestList_MoveToBack_EmptyList(t *testing.T) {
	l := New()

	node := &Node{Value: 1}

	PrintListValues(l)

	l.MoveToBack(node)

	PrintListValues(l)

	assert.Equal(t, int64(1), l.Len(), "list length should be 0")
	assert.Equal(t, node, l.Front(), "front node should be the node")
	assert.Equal(t, node, l.Back(), "back node should be the node")
	assert.Equal(t, node.parentRef, toUnsafePtr(l), "node parentRef should be the list")

	assert.Nil(t, node.Left, "node prev should be nil")
	assert.Nil(t, node.Right, "node next should be nil")
}

func TestList_InsertBefore(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	node4 := &Node{Value: 4}

	PrintListValues(l)

	l.InsertBefore(node4, node2)

	PrintListValues(l)

	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	assert.Equal(t, node4, node1.Right, "node1 next should be newNode")
	assert.Equal(t, node2, node4.Right, "newNode next should be node2")
	assert.Equal(t, node4, node2.Left, "node2 prev should be newNode")
	assert.Equal(t, node1, node4.Left, "newNode prev should be node1")
	assert.Nil(t, node3.Right, "node3 next should be nil")
}

func TestList_InsertBefore_SameValue(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 1}
	node3 := &Node{Value: 1}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	node4 := &Node{Value: 1}

	PrintListValues(l)

	l.InsertBefore(node4, node2)

	PrintListValues(l)

	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	assert.Equal(t, node4, node1.Right, "node1 next should be newNode")
	assert.Equal(t, node2, node4.Right, "newNode next should be node2")
	assert.Equal(t, node4, node2.Left, "node2 prev should be newNode")
	assert.Equal(t, node1, node4.Left, "newNode prev should be node1")
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")

}

func TestList_InsertBefore_FirstNode(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	node4 := &Node{Value: 4}

	PrintListValues(l)

	l.InsertBefore(node4, node1)

	PrintListValues(l)

	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node4, l.Front(), "front node should be newNode")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	assert.Equal(t, node1, node4.Right, "newNode next should be node1")
	assert.Equal(t, node4, node1.Left, "node1 prev should be newNode")
	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Equal(t, node2, node3.Left, "node3 prev should be node2")
	assert.Nil(t, node3.Right, "node3 next should be nil")
}

func TestList_InsertBefore_Nil(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}

	l.PushBack(node1)

	PrintListValues(l)

	l.InsertBefore(nil, node1)

	PrintListValues(l)

	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node1, l.Back(), "back node should be node1")
	assert.Equal(t, node1.parentRef, toUnsafePtr(l), "node1 parentRef should be the list")

	l.InsertBefore(node2, nil)

	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node1, l.Back(), "back node should be node1")
	assert.Equal(t, node1.parentRef, toUnsafePtr(l), "node1 parentRef should be the list")
}

func TestList_InsertBefore_InvalidNode(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}

	l.PushBack(node1)
	l.PushBack(node2)

	node3 := &Node{Value: 3}

	PrintListValues(l)

	l.InsertBefore(node3, node2)

	PrintListValues(l)

	assert.Equal(t, int64(3), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node3")
	assert.Equal(t, node3.parentRef, toUnsafePtr(l), "node3 parentRef should be the list")

	assert.Equal(t, node3, node1.Right, "node1 next should be node3")
	assert.Equal(t, node1, node3.Left, "node3 prev should be node1")
	assert.Equal(t, node2, node3.Right, "node3 next should be node2")
	assert.Nil(t, node2.Right, "node2 next should be nil")

	node4 := &Node{Value: 4}

	PrintListValues(l)

	l.InsertBefore(node4, node1)

	PrintListValues(l)

	assert.Equal(t, int64(4), l.Len(), "list length should be 2")
	assert.Equal(t, node4, l.Front(), "front node should be node4")
	assert.Equal(t, node2, l.Back(), "back node should be node2")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	assert.Equal(t, node1, node4.Right, "node4 next should be node1")
	assert.Equal(t, node4, node1.Left, "node1 prev should be node4")
	assert.Equal(t, node3, node1.Right, "node1 next should be node3")
	assert.Equal(t, node1, node3.Left, "node3 prev should be node1")
	assert.Equal(t, node2, node3.Right, "node3 next should be node2")
	assert.Nil(t, node2.Right, "node2 next should be nil")
}

func TestList_InsertBefore_EmptyList(t *testing.T) {
	l := New()

	node := &Node{Value: 1}

	PrintListValues(l)

	l.InsertBefore(node, nil)

	PrintListValues(l)

	assert.Equal(t, int64(0), l.Len(), "list length should be 1")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
	assert.Nil(t, node.parentRef, "node parentRef should be nil")
}

func TestList_InsertAfter(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}
	node4 := &Node{Value: 4}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	PrintListValues(l)

	l.InsertAfter(node4, node1)

	PrintListValues(l)

	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	assert.Equal(t, node4, node1.Right, "node1 next should be node4")
	assert.Equal(t, node2, node4.Right, "node4 next should be node2")
	assert.Equal(t, node4, node2.Left, "node2 prev should be node4")
	assert.Equal(t, node1, node4.Left, "node4 prev should be node1")
}

func TestList_InsertAfter_SameValue(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 1}
	node3 := &Node{Value: 1}
	node4 := &Node{Value: 1}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	PrintListValues(l)

	l.InsertAfter(node4, node1)

	PrintListValues(l)

	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

	assert.Equal(t, node4, node1.Right, "node1 next should be node4")
	assert.Equal(t, node2, node4.Right, "node4 next should be node2")
	assert.Equal(t, node4, node2.Left, "node2 prev should be node4")
	assert.Equal(t, node1, node4.Left, "node4 prev should be node1")
}

func TestList_InsertAfter_LastNode(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)

	PrintListValues(l)

	l.InsertAfter(node3, node2)

	PrintListValues(l)

	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")
	assert.Equal(t, node3.parentRef, toUnsafePtr(l), "node3 parentRef should be the list")

	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Nil(t, node3.Right, "node3 next should be nil")
}

func TestList_InsertAfter_Nil(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}

	l.PushBack(node1)

	PrintListValues(l)

	l.InsertAfter(nil, node1)

	PrintListValues(l)

	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node1, l.Back(), "back node should be node1")
	assert.Equal(t, node1.parentRef, toUnsafePtr(l), "node1 parentRef should be the list")

	l.InsertAfter(node2, nil)

	assert.Equal(t, int64(1), l.Len(), "list length should be 1")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node1, l.Back(), "back node should be node1")
	assert.Equal(t, node1.parentRef, toUnsafePtr(l), "node1 parentRef should be the list")
}

func TestList_InsertAfter_InvalidNode(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}

	l.PushBack(node1)
	l.PushBack(node2)

	node3 := &Node{Value: 3}

	PrintListValues(l)

	l.InsertAfter(node3, node1)

	PrintListValues(l)

	assert.Equal(t, int64(3), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node3")
	assert.Equal(t, node3.parentRef, toUnsafePtr(l), "node3 parentRef should be the list")

	assert.Equal(t, node3, node1.Right, "node1 next should be node3")
	assert.Equal(t, node1, node3.Left, "node3 prev should be node1")
	assert.Equal(t, node2, node3.Right, "node3 next should be node2")
	assert.Nil(t, node2.Right, "node2 next should be nil")

	node4 := &Node{Value: 4}

	PrintListValues(l)

	l.InsertAfter(node4, node2)

	PrintListValues(l)

	assert.Equal(t, int64(4), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node4, l.Back(), "back node should be node4")
	assert.Equal(t, node4.parentRef, toUnsafePtr(l), "node4 parentRef should be the list")

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

	node := &Node{Value: 1}

	PrintListValues(l)

	l.InsertAfter(node, nil)

	PrintListValues(l)

	assert.Equal(t, int64(0), l.Len(), "list length should be 1")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
	assert.Nil(t, node.parentRef, "node parentRef should be nil")
}

func TestList_Swap(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}
	node4 := &Node{Value: 4}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)
	l.PushBack(node4)

	PrintListValues(l)

	l.Swap(node2, node3)

	PrintListValues(l)

	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node4, l.Back(), "back node should be node4")

	assert.Equal(t, node3, node1.Right, "node1 next should be node3")
	assert.Equal(t, node2, node3.Right, "node3 next should be node2")
	assert.Equal(t, node4, node2.Right, "node2 next should be node4")
	assert.Nil(t, node4.Right, "node4 next should be nil")

	assert.Equal(t, node1, node3.Left, "node3 prev should be node1")
	assert.Equal(t, node3, node2.Left, "node2 prev should be node3")
	assert.Equal(t, node2, node4.Left, "node4 prev should be node2")
	assert.Nil(t, node1.Left, "node2 prev should be nil")

	l.Swap(node1, node4)

	PrintListValues(l)

	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node4, l.Front(), "front node should be node4")
	assert.Equal(t, node1, l.Back(), "back node should be node1")

	assert.Equal(t, node3, node4.Right, "node4 next should be node3")
	assert.Equal(t, node2, node3.Right, "node3 next should be node2")
	assert.Equal(t, node1, node2.Right, "node2 next should be node1")
	assert.Nil(t, node1.Right, "node1 next should be nil")

	assert.Equal(t, node4, node3.Left, "node3 prev should be node4")
	assert.Equal(t, node3, node2.Left, "node2 prev should be node3")
	assert.Equal(t, node2, node1.Left, "node1 prev should be node2")
	assert.Nil(t, node4.Left, "node2 prev should be nil")

	l.Swap(node2, node4)

	PrintListValues(l)

	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node2, l.Front(), "front node should be node4")
	assert.Equal(t, node1, l.Back(), "back node should be node1")

	assert.Equal(t, node3, node2.Right, "node2 next should be node3")
	assert.Equal(t, node4, node3.Right, "node3 next should be node4")
	assert.Equal(t, node1, node4.Right, "node4 next should be node1")
	assert.Nil(t, node1.Right, "node1 next should be nil")

	assert.Equal(t, node2, node3.Left, "node3 prev should be node2")
	assert.Equal(t, node3, node4.Left, "node4 prev should be node3")
	assert.Equal(t, node4, node1.Left, "node1 prev should be node4")
	assert.Nil(t, node2.Left, "node2 prev should be nil")

	l.Swap(node1, node3)

	PrintListValues(l)

	assert.Equal(t, int64(4), l.Len(), "list length should be 4")
	assert.Equal(t, node2, l.Front(), "front node should be node4")
	assert.Equal(t, node3, l.Back(), "back node should be node1")

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

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}

	l.PushBack(node1)
	l.PushBack(node2)

	node3 := &Node{Value: 3}

	PrintListValues(l)

	l.Swap(node3, node2)

	PrintListValues(l)

	assert.Equal(t, int64(2), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node2")
	assert.Nil(t, node3.parentRef, "node3 parentRef should be nil")

	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Nil(t, node2.Right, "node2 next should be nil")
	assert.Nil(t, node1.Left, "node1 prev should be nil")

	l.Swap(node3, node1)

	PrintListValues(l)

	assert.Equal(t, int64(2), l.Len(), "list length should be 2")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node2, l.Back(), "back node should be node2")
	assert.Nil(t, node3.parentRef, "node3 parentRef should be nil")

	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Nil(t, node2.Right, "node2 next should be nil")
	assert.Nil(t, node1.Left, "node1 prev should be nil")
}

func TestList_Swap_EmptyList(t *testing.T) {
	l := New()

	node := &Node{Value: 1}

	PrintListValues(l)

	l.Swap(node, nil)

	PrintListValues(l)

	assert.Equal(t, int64(0), l.Len(), "list length should be 0")
	assert.Nil(t, l.Front(), "front node should be nil")
	assert.Nil(t, l.Back(), "back node should be nil")
	assert.Nil(t, node.parentRef, "node parentRef should be nil")
}

func TestList_Slice(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	s := l.Slice()

	assert.Equal(t, 3, len(s), "slice length should be 3")

	assert.Equal(t, node1.Value, s[0], "slice[0] should be node1 value")
	assert.Equal(t, node2.Value, s[1], "slice[1] should be node2 value")
	assert.Equal(t, node3.Value, s[2], "slice[2] should be node3 value")
}

func TestList_LargeDataSet(t *testing.T) {
	l := New()
	count := 1000000

	startTime := time.Now()
	for i := 0; i < count; i++ {
		l.PushBack(&Node{Value: i})
	}
	pushDuration := time.Since(startTime)

	assert.Less(t, pushDuration.Seconds(), float64(5), "pushing 1M nodes should take less than 5 seconds")
	assert.Equal(t, int64(count), l.Len(), "list length should be 1M")

	startTime = time.Now()
	for i := 0; i < count; i++ {
		l.PopBack()
	}
	popDuration := time.Since(startTime)

	assert.Less(t, popDuration.Seconds(), float64(5), "popping 1M nodes should take less than 5 seconds")
	assert.Equal(t, int64(0), l.Len(), "list should be empty after popping all nodes")
}

func TestList_NilValues(t *testing.T) {
	l := New()

	node1 := &Node{Value: nil}
	node2 := &Node{Value: nil}
	node3 := &Node{Value: nil}

	l.PushBack(node1)
	l.PushBack(node2)
	l.PushBack(node3)

	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node1, l.Front(), "front node should be node1")
	assert.Equal(t, node3, l.Back(), "back node should be node3")

	assert.Nil(t, l.Front().Value, "front node value should be nil")
	assert.Nil(t, l.Back().Value, "back node value should be nil")
}

func TestList_EdgeCases(t *testing.T) {
	l := New()

	assert.Nil(t, l.PopBack(), "PopBack on empty list should return nil")
	assert.Nil(t, l.PopFront(), "PopFront on empty list should return nil")
	assert.Nil(t, l.Front(), "Front on empty list should return nil")
	assert.Nil(t, l.Back(), "Back on empty list should return nil")

	node := &Node{Value: 1}
	l.PushBack(node)
	l.MoveToFront(node)
	assert.Equal(t, node, l.Front(), "node should still be at front")
	l.MoveToBack(node)
	assert.Equal(t, node, l.Back(), "node should still be at back")

	l.Remove(node)
	l.Remove(node)
	assert.Equal(t, int64(0), l.Len(), "list should be empty after removing node")

	l.PushBack(nil)
	l.PushFront(nil)
	l.Remove(nil)
	assert.Equal(t, int64(0), l.Len(), "list should remain empty after nil operations")
}

func TestList_ChainedOperations(t *testing.T) {
	l := New()

	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node3 := &Node{Value: 3}

	l.PushBack(node1)
	l.MoveToFront(node1)
	l.PushBack(node2)
	l.MoveToBack(node1)
	l.PushFront(node3)
	l.Remove(node2)
	l.PushBack(node2)

	assert.Equal(t, int64(3), l.Len(), "list length should be 3")
	assert.Equal(t, node3, l.Front(), "front node should be node3")
	assert.Equal(t, node2, l.Back(), "back node should be node2")

	assert.Equal(t, node1, node3.Right, "node3 next should be node1")
	assert.Equal(t, node2, node1.Right, "node1 next should be node2")
	assert.Nil(t, node2.Right, "node2 next should be nil")
}

func TestList_ConcurrentOperations(t *testing.T) {
	l := New()
	done := make(chan bool)
	count := 1000
	var mu sync.Mutex

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

	<-done
	<-done

	assert.Equal(t, int64(count*2), l.Len(), "list length should be double the count")

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

	<-done
	<-done

	assert.Equal(t, int64(0), l.Len(), "list should be empty after concurrent pops")
}

func TestList_ConcurrentOperations_Extended(t *testing.T) {
	l := New()
	done := make(chan bool)
	count := 1000
	valueMap := make(map[interface{}]bool)
	var mu sync.Mutex

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
		time.Sleep(100 * time.Millisecond)
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
		time.Sleep(100 * time.Millisecond)
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

	for i := 0; i < 4; i++ {
		<-done
	}

	actualValues := make(map[interface{}]bool)
	for node := l.Front(); node != nil; node = node.Right {
		actualValues[node.Value] = true
	}

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

	for i := 0; i < count; i++ {
		nodes[i] = &Node{Value: i}
		insertNodes[i] = &Node{Value: fmt.Sprintf("insert_%d", i)}
		l.PushBack(nodes[i])
	}

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

	<-done
	<-done

	expectedLen := int64(count + count)
	assert.Equal(t, expectedLen, l.Len(), "list length should match expected")

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

	for i := 0; i < count; i++ {
		nodes[i] = &Node{Value: i}
		l.PushBack(nodes[i])
	}

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

	<-done
	<-done

	assert.Equal(t, int64(count), l.Len(), "list length should remain unchanged")

	nodeMap := make(map[interface{}]bool)
	for node := l.Front(); node != nil; node = node.Right {
		nodeMap[node.Value] = true
	}
	assert.Equal(t, count, len(nodeMap), "all nodes should still be present")

	var prev *Node
	for node := l.Front(); node != nil; node = node.Right {
		if prev != nil {
			assert.Equal(t, prev, node.Left, "node links should be consistent")
		}
		prev = node
	}
}
