package heap

import (
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
	"github.com/shengyanli1982/workqueue/v2/internal/lockfree/stack"
)

const (
	INIT_NODE_SIZE = 8
	// UPWORD_LEVEL   = INIT_NODE_SIZE << 5
	// DOWNWORD_LEVEL = INIT_NODE_SIZE >> 5
)

type Heap struct {
	list    *lst.List
	mapping []*lst.Node
	cache   *stack.Stack
}

func New() *Heap {
	return &Heap{
		list:    lst.New(),
		cache:   stack.New(),
		mapping: make([]*lst.Node, 0, INIT_NODE_SIZE),
	}
}

func (h *Heap) less(i, j *lst.Node) bool { return i.Priority < j.Priority }

// func (h *Heap) swap(i, j *lst.Node) {
// 	h.mapping[i.Index], h.mapping[j.Index] = j, i
// 	h.mapping[i.Index].Index, h.mapping[j.Index].Index = i.Index, j.Index
// 	h.list.Swap(i, j)
// }

func getMiddleNode(low, high *lst.Node) *lst.Node {
	if low == nil || high == nil {
		return nil
	}

	slow := low
	fast := low

	for fast != high && fast.Next != high {
		slow = slow.Next
		fast = fast.Next.Next
	}

	return slow
}

func (h *Heap) moveUp(node *lst.Node) {
	if node == nil || node.Prev == nil {
		return
	}

	low := h.list.Front()
	high := h.list.Back()

	for low != high {
		mid := getMiddleNode(low, high)

		if h.less(node, mid) {
			high = mid
		} else {
			low = mid.Next
		}
	}

	for high != nil && h.less(node, high) {
		high = high.Prev
	}

	h.list.Remove(node)

	if high == nil {
		h.list.PushFront(node)
	} else {
		h.list.InsertAfter(node, high)
	}
}

func (h *Heap) Len() int64 { return h.list.Len() }

func (h *Heap) Front() *lst.Node { return h.list.Front() }

func (h *Heap) Back() *lst.Node { return h.list.Back() }

func (h *Heap) Range(f func(*lst.Node) bool) { h.list.Range(f) }

func (h *Heap) Slice() []interface{} { return h.list.Slice() }

func (h *Heap) Cleanup() {
	h.list.Cleanup()
	h.mapping = h.mapping[:0]
}

func (h *Heap) Remove(node *lst.Node) {
	if node == nil {
		return
	}
	h.list.Remove(node)
}

func (h *Heap) Push(n *lst.Node) {
	if n == nil {
		return
	}

	if h.cache.Len() > 0 {
		n.Index = h.cache.Pop()
		h.mapping[n.Index] = n
	} else {
		n.Index = h.list.Len()
		h.mapping = append(h.mapping, n)
	}

	h.list.PushBack(n)
	h.moveUp(n)
}

func (h *Heap) Pop() *lst.Node {
	if h.Len() <= 0 {
		return nil
	}

	n := h.list.PopFront()
	h.mapping[n.Index] = nil
	return n
}
