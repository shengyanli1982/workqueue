package heap

import (
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// Heap 结构体表示一个堆，它包含一个链表。
// The Heap struct represents a heap, which contains a list.
type Heap struct {
	list *lst.List
}

// New 函数创建并返回一个新的 Heap 实例。
// The New function creates and returns a new Heap instance.
func New() *Heap {
	return &Heap{
		// 初始化一个新的链表。
		// Initialize a new list.
		list: lst.New(),
	}
}

// less 方法比较两个节点的优先级，如果 i 的优先级小于 j 的优先级，返回 true，否则返回 false。
// The less method compares the priorities of two nodes, returns true if the priority of i is less than the priority of j, otherwise returns false.
func (h *Heap) less(i, j *lst.Node) bool {
	return i.Priority < j.Priority
}

// moveUp 方法将节点 node 向上移动到正确的位置，以保持堆的性质。
// The moveUp method moves the node node up to the correct position to maintain the properties of the heap.
func (h *Heap) moveUp(node *lst.Node) {
	// 如果 node 是 nil 或者 node 的 Prev 是 nil，我们就直接返回，不做任何操作。
	// If node is nil or node's Prev is nil, we just return directly without doing anything.
	if node == nil || node.Prev == nil {
		return
	}

	// 设置一个 current 变量，用于保存当前节点。
	// Set a current variable to save the current node.
	current := node

	// 当 current 的 Prev 不是 nil，并且 node 的优先级小于 current 的 Prev 的优先级时，我们将 current 设置为 current 的 Prev。
	// When current's Prev is not nil, and the priority of node is less than the priority of current's Prev, we set current to current's Prev.
	for current.Prev != nil && h.less(node, current.Prev) {
		current = current.Prev
	}

	// 如果 current 不等于 node，我们就从链表的尾部弹出一个节点，并在 current 之前插入 node。
	// If current is not equal to node, we pop a node from the end of the list, and insert node before current.
	if current != node {
		h.list.PopBack()
		h.list.InsertBefore(node, current)
	}
}

// Len 方法返回堆的长度。
// The Len method returns the length of the heap.
func (h *Heap) Len() int64 { return h.list.Len() }

// Front 方法返回堆的第一个节点。
// The Front method returns the first node of the heap.
func (h *Heap) Front() *lst.Node { return h.list.Front() }

// Back 方法返回堆的最后一个节点。
// The Back method returns the last node of the heap.
func (h *Heap) Back() *lst.Node { return h.list.Back() }

// Range 方法对堆中的每个节点执行函数 fn。
// The Range method executes function fn for each node in the heap.
func (h *Heap) Range(fn func(*lst.Node) bool) { h.list.Range(fn) }

// Slice 方法返回堆中所有节点的切片。
// The Slice method returns a slice of all nodes in the heap.
func (h *Heap) Slice() []interface{} { return h.list.Slice() }

// Cleanup 方法清理堆，移除所有节点。
// The Cleanup method cleans up the heap, removing all nodes.
func (h *Heap) Cleanup() {
	h.list.Cleanup()
}

// Remove 方法从堆中移除节点 node。
// The Remove method removes node node from the heap.
func (h *Heap) Remove(node *lst.Node) {
	// 如果 node 是 nil，我们就直接返回，不做任何操作。
	// If node is nil, we just return directly without doing anything.
	if node == nil {
		return
	}

	// 调用 list 的 Remove 方法，从链表中移除节点 node。
	// Call the Remove method of list to remove node node from the list.
	h.list.Remove(node)
}

// Push 方法将节点 n 添加到堆的尾部，并调整堆的结构。
// The Push method adds node n to the end of the heap and adjusts the structure of the heap.
func (h *Heap) Push(n *lst.Node) {
	// 如果 n 是 nil，我们就直接返回，不做任何操作。
	// If n is nil, we just return directly without doing anything.
	if n == nil {
		return
	}

	// 调用 list 的 PushBack 方法，将节点 n 添加到链表的尾部。
	// Call the PushBack method of list to add node n to the end of the list.
	h.list.PushBack(n)

	// 调用 moveUp 方法，调整堆的结构，使其满足堆的性质。
	// Call the moveUp method to adjust the structure of the heap to satisfy the properties of the heap.
	h.moveUp(n)
}

// Pop 方法移除并返回堆的第一个节点。
// The Pop method removes and returns the first node of the heap.
func (h *Heap) Pop() *lst.Node {
	// 如果堆的长度小于等于 0，我们就返回 nil。
	// If the length of the heap is less than or equal to 0, we return nil.
	if h.Len() <= 0 {
		return nil
	}

	// 调用 list 的 PopFront 方法，移除并返回链表的第一个节点。
	// Call the PopFront method of list to remove and return the first node of the list.
	return h.list.PopFront()
}

// GetList 方法返回堆的内部列表。
// The GetList method returns the internal list of the heap.
func (h *Heap) GetList() *lst.List {
	return h.list
}
