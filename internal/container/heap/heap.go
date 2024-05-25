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
	id      uint64
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

// func getMiddleNode(low, high *lst.Node) *lst.Node {
// 	if low == nil || high == nil {
// 		return nil
// 	}

// 	slow := low
// 	fast := low

// 	for fast != high && fast.Next != high {
// 		slow = slow.Next
// 		fast = fast.Next.Next
// 	}

// 	return slow
// }

// func getMiddleNode(low, high *lst.Node) *lst.Node {
// 	if low == nil || high == nil {
// 		return nil
// 	}

// 	// 计算 low 和 high 之间的距离
// 	lowIndex := low.Index
// 	highIndex := high.Index
// 	midIndex := (lowIndex + highIndex) / 2

// 	// 从 low 或 high 开始，找到中间节点
// 	mid := low
// 	if midIndex-lowIndex < highIndex-midIndex {
// 		// 从 low 开始向前移动
// 		for mid.Index < midIndex {
// 			mid = mid.Next
// 		}
// 	} else {
// 		// 从 high 开始向后移动
// 		mid = high
// 		for mid.Index > midIndex {
// 			mid = mid.Prev
// 		}
// 	}

// 	return mid
// }

// func (h *Heap) moveUp(node *lst.Node) {
// 	if node == nil || node.Prev == nil {
// 		return
// 	}

// 	low := h.list.Front()
// 	high := h.list.Back()

// 	for low != high {
// 		mid := getMiddleNode(low, high)

// 		if h.less(node, mid) {
// 			high = mid
// 		} else {
// 			low = mid.Next
// 		}

// 		fmt.Printf("## mid index: %v, mid priority: %v\n", mid.Index, mid.Priority)
// 		fmt.Printf("## low: %v, high: %v\n", low.Priority, high.Priority)
// 	}

// 	for high != nil && h.less(node, high) {
// 		fmt.Println(">> high:", high.Priority)
// 		high = high.Prev
// 	}

// 	h.list.Remove(node)

// 	if high == nil {
// 		h.list.PushFront(node)
// 	} else {
// 		h.list.InsertAfter(node, high)
// 	}

// 	fmt.Printf("xxxx\n")
// }

// func (h *Heap) moveUp(node *lst.Node) {
// 	if node == nil || node.Prev == nil {
// 		return
// 	}

// 	current := node
// 	for current.Prev != nil && h.less(node, current.Prev) {
// 		current = current.Prev
// 	}

// 	if current != node {
// 		h.list.Remove(node)
// 		h.list.InsertBefore(node, current)
// 	}
// }

func getMiddleNode(low, high *lst.Node) *lst.Node {
	if low == nil || high == nil || low == high {
		return low
	}

	// 计算 low 和 high 之间的距离
	lowIndex := low.Index
	highIndex := high.Index
	midIndex := (lowIndex + highIndex) / 2

	// 从 low 或 high 开始，找到中间节点
	mid := low
	if midIndex-lowIndex < highIndex-midIndex {
		// 从 low 开始向前移动
		for mid.Index < midIndex {
			mid = mid.Next
		}
	} else {
		// 从 high 开始向后移动
		mid = high
		for mid.Index > midIndex {
			mid = mid.Prev
		}
	}

	return mid
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

	if high == nil {
		h.list.PushFront(node)
	} else {
		h.list.InsertAfter(node, high)
		h.list.Remove(node)
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

	n.Index = h.id
	h.id++
	h.list.PushBack(n)
	h.moveUp(n)
}

func (h *Heap) Pop() *lst.Node {
	if h.Len() <= 0 {
		return nil
	}

	n := h.list.PopFront()
	// h.mapping[n.Index] = nil
	return n
}
