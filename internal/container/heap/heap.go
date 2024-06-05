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
// The Remove method removes the node from the heap.
func (h *Heap) Remove(node *lst.Node) {
	// 如果 node 是 nil，我们就直接返回，不做任何操作。
	// If node is nil, we just return directly without doing anything.
	if node == nil {
		return
	}

	// 调用 list 的 Remove 方法，从链表中移除节点 node。
	// Call the Remove method of list to remove the node from the list.
	h.list.Remove(node)
}

// Push 方法将一个新的节点添加到堆中。
// The Push method adds a new node to the heap.
func (h *Heap) Push(n *lst.Node) {
	// 如果 n 是 nil 或者堆的长度为 0，我们就直接在链表的后面添加 n。
	// If n is nil or the length of the heap is 0, we just add n to the back of the list.
	if n == nil || h.list.Len() == 0 {
		h.list.PushBack(n)
		return
	}

	// 获取链表的最后一个节点。
	// Get the last node of the list.
	current := h.list.Back()

	// 如果当前节点不是 nil 并且 n 小于当前节点，我们就继续向前查找。
	// If the current node is not nil and n is less than the current node, we continue to look forward.
	for current != nil && h.less(n, current) {
		current = current.Prev
	}

	// 如果当前节点是 nil，我们就在链表的前面添加 n。
	// If the current node is nil, we add n to the front of the list.
	if current == nil {
		h.list.PushFront(n)
	} else {
		// 否则，我们就在当前节点的后面添加 n。
		// Otherwise, we add n after the current node.
		h.list.InsertAfter(n, current)
	}
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
