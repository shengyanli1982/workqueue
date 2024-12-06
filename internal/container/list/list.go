package list

import "unsafe"

// isPtrEqual 函数检查一个 unsafe.Pointer 和一个 *List 是否指向同一个地址。
// The isPtrEqual function checks if an unsafe.Pointer and a *List point to the same address.
func isPtrEqual(up unsafe.Pointer, lp *List) bool {
	// 我们将两个指针转换为 uintptr，然后比较它们是否相等。
	// We convert both pointers to uintptr and then compare if they are equal.
	return uintptr(up) == uintptr(unsafe.Pointer(lp))
}

// toUnsafePtr 函数将一个 *List 转换为 unsafe.Pointer。
// The toUnsafePtr function converts a *List to an unsafe.Pointer.
func toUnsafePtr(lp *List) unsafe.Pointer {
	// 我们直接使用 unsafe.Pointer(lp) 来进行转换。
	// We use unsafe.Pointer(lp) directly for the conversion.
	return unsafe.Pointer(lp)
}

// List 结构体代表一个双向链表，它有一个头节点和一个尾节点，以及一个记录链表长度的字段。
// The List struct represents a doubly linked list. It has a head node, a tail node, and a field to record the length of the list.
type List struct {
	// head 是链表的头节点。
	// head is the head node of the list.
	// tail 是链表的尾节点。
	// tail is the tail node of the list.
	head, tail *Node

	// count 是链表的长度。
	// count is the length of the list.
	count int64
}

// New 函数创建并返回一个新的 List 实例。
// The New function creates and returns a new instance of List.
func New() *List { return &List{} }

// Len 方法返回链表的长度。
// The Len method returns the length of the list.
func (l *List) Len() int64 { return l.count }

// Front 方法返回链表的头节点。
// The Front method returns the head node of the list.
func (l *List) Front() *Node { return l.head }

// Back 方法返回链表的尾节点。
// The Back method returns the tail node of the list.
func (l *List) Back() *Node { return l.tail }

// PushBack 方法将一个节点添加到链表的尾部。
// The PushBack method adds a node to the end of the list.
func (l *List) PushBack(node *Node) {
	// 如果 n 是 nil，我们就直接返回，不做任何操作。
	// If n is nil, we just return directly without doing anything.
	if node == nil {
		return
	}

	// 我们将 n 的 parentRef 设置为 l 的地址。
	// We set n's parentRef to the address of l.
	node.parentRef = toUnsafePtr(l)

	// 如果链表是空的，我们就将 n 设置为头节点。
	// If the list is empty, we set n as the head node.
	if l.head == nil {
		l.head = node
	} else {
		// 否则，我们将 n 添加到尾节点的后面，并更新尾节点。
		// Otherwise, we add n after the tail node and update the tail node.
		l.tail.Right = node
		node.Left = l.tail
	}

	// 更新尾节点为 n。
	// Update the tail node to n.
	l.tail = node

	// 链表长度加 1。
	// Increase the length of the list by 1.
	l.count++
}

// PushFront 方法将一个节点添加到链表的头部。
// The PushFront method adds a node to the beginning of the list.
func (l *List) PushFront(node *Node) {
	// 如果 n 是 nil，我们就直接返回，不做任何操作。
	// If n is nil, we just return directly without doing anything.
	if node == nil {
		return
	}

	// 我们将 n 的 parentRef 设置为 l 的地址。
	// We set n's parentRef to the address of l.
	node.parentRef = toUnsafePtr(l)

	// 如果链表是空的，我们就将 n 设置为尾节点。
	// If the list is empty, we set n as the tail node.
	if l.head == nil {
		l.tail = node
	} else {
		// 否则，我们将 n 添加到头节点的前面，并更新头节点。
		// Otherwise, we add n before the head node and update the head node.
		l.head.Left = node
		node.Right = l.head
	}

	// 更新头节点为 n。
	// Update the head node to n.
	l.head = node

	// 链表长度加 1。
	// Increase the length of the list by 1.
	l.count++
}

// PopBack 方法从链表的尾部移除一个节点并返回它。
// The PopBack method removes a node from the end of the list and returns it.
func (l *List) PopBack() *Node {
	// 如果链表是空的，我们就返回 nil。
	// If the list is empty, we return nil.
	if l.tail == nil {
		return nil
	}

	// 我们获取尾节点，并将其保存在 n 中。
	// We get the tail node and save it in n.
	n := l.tail

	// 我们将尾节点更新为尾节点的前一个节点。
	// We update the tail node to the previous node of the tail node.
	l.tail = n.Left

	// 如果新的尾节点是 nil，说明链表现在是空的，我们将头节点也设置为 nil。
	// If the new tail node is nil, it means the list is now empty, we also set the head node to nil.
	if l.tail == nil {
		l.head = nil
	} else {
		// 否则，我们将新的尾节点的 Next 设置为 nil。
		// Otherwise, we set the Next of the new tail node to nil.
		l.tail.Right = nil
	}

	// 我们将 n 的 parentRef、Prev 和 Next 都设置为 nil。
	// We set n's parentRef, Prev, and Next all to nil.
	n.parentRef = nil
	n.Left = nil
	n.Right = nil

	// 链表长度减 1。
	// Decrease the length of the list by 1.
	l.count--

	// 返回被移除的节点。
	// Return the removed node.
	return n
}

// PopFront 方法从链表的头部移除一个节点并返回它。
// The PopFront method removes a node from the beginning of the list and returns it.
func (l *List) PopFront() *Node {
	// 如果链表是空的，我们就返回 nil。
	// If the list is empty, we return nil.
	if l.head == nil {
		return nil
	}

	// 我们获取头节点，并将其保存在 n 中。
	// We get the head node and save it in n.
	n := l.head

	// 我们将头节点更新为头节点的下一个节点。
	// We update the head node to the next node of the head node.
	l.head = n.Right

	// 如果新的头节点是 nil，说明链表现在是空的，我们将尾节点也设置为 nil。
	// If the new head node is nil, it means the list is now empty, we also set the tail node to nil.
	if l.head == nil {
		l.tail = nil
	} else {
		// 否则，我们将新的头节点的 Prev 设置为 nil。
		// Otherwise, we set the Prev of the new head node to nil.
		l.head.Left = nil
	}

	// 我们将 n 的 parentRef、Prev 和 Next 都设置为 nil。
	// We set n's parentRef, Prev, and Next all to nil.
	n.parentRef = nil
	n.Left = nil
	n.Right = nil

	// 链表长度减 1。
	// Decrease the length of the list by 1.
	l.count--

	// 返回被移除的节点。
	// Return the removed node.
	return n
}

// Remove 方法从链表中移除节点 n。
// The Remove method removes node n from the list.
func (l *List) Remove(node *Node) {
	// 如果 n 是 nil，或者链表是空的，或者 n 不属于链表 l，我们就直接返回，不做任何操作。
	// If n is nil, or the list is empty, or n does not belong to list l, we just return directly without doing anything.
	if node == nil || l.count == 0 || !isPtrEqual(node.parentRef, l) {
		return
	}

	// 如果 n 的 Prev 是 nil，说明 n 是头节点，我们就将链表的头节点设置为 n 的 Next。
	// If n's Prev is nil, it means n is the head node, we set the head node of the list to n's Next.
	if node.Left == nil {
		l.head = node.Right
	} else {
		// 否则，我们将 n 的 Prev 的 Next 设置为 n 的 Next。
		// Otherwise, we set n's Prev's Next to n's Next.
		node.Left.Right = node.Right
	}

	// 如果 n 的 Next 是 nil，说明 n 是尾节点，我们就将链表的尾节点设置为 n 的 Prev。
	// If n's Next is nil, it means n is the tail node, we set the tail node of the list to n's Prev.
	if node.Right == nil {
		l.tail = node.Left
	} else {
		// 否则，我们将 n 的 Next 的 Prev 设置为 n 的 Prev。
		// Otherwise, we set n's Next's Prev to n's Prev.
		node.Right.Left = node.Left
	}

	// 我们将 n 的 parentRef、Prev 和 Next 都设置为 nil。
	// We set n's parentRef, Prev and Next all to nil.
	node.parentRef = nil
	node.Left = nil
	node.Right = nil

	// 链表长度减 1。
	// Decrease the length of the list by 1.
	l.count--
}

// initNodeInEmptyList 辅助方法用于在空链表中初始化一个节点
// initNodeInEmptyList helper method initializes a node in an empty list
func (l *List) initNodeInEmptyList(node *Node) bool {
	if l.head == nil && l.tail == nil {
		node.parentRef = toUnsafePtr(l)
		node.Left = nil
		node.Right = nil
		l.head = node
		l.tail = node
		l.count++
		return true
	}
	return false
}

// MoveToFront 方法将节点 n 移动到链表的头部。
// The MoveToFront method moves node n to the front of the list.
func (l *List) MoveToFront(node *Node) {
	if node == nil {
		return
	}

	if l.initNodeInEmptyList(node) {
		return
	}

	if node == l.head {
		return
	}

	if node.Left != nil {
		node.Left.Right = node.Right
	} else if node != l.head {
		node.parentRef = toUnsafePtr(l)
		node.Right = l.head
		l.head.Left = node
		l.head = node
		l.count++
		return
	}

	if node.Right != nil {
		node.Right.Left = node.Left
	} else {
		l.tail = node.Left
	}

	node.Left = nil
	node.Right = l.head
	l.head.Left = node
	l.head = node
}

// MoveToBack 方法将节点 n 移动到链表的尾部。
// The MoveToBack method moves node n to the end of the list.
func (l *List) MoveToBack(node *Node) {
	if node == nil {
		return
	}

	if l.initNodeInEmptyList(node) {
		return
	}

	if node == l.tail {
		return
	}

	if node.Left != nil {
		node.Left.Right = node.Right
	} else if node != l.head {
		node.parentRef = toUnsafePtr(l)
		node.Left = l.tail
		l.tail.Right = node
		l.tail = node
		l.count++
		return
	}

	if node.Right != nil {
		node.Right.Left = node.Left
	} else {
		l.head = node.Right
	}

	node.Left = l.tail
	node.Right = nil
	if l.tail != nil {
		l.tail.Right = node
	}
	l.tail = node
}

// validateSwapNodes 辅助方法用于验证交换节点的有效性
// validateSwapNodes helper method validates if two nodes can be swapped
func (l *List) validateSwapNodes(node, mark *Node) bool {
	if node == nil || mark == nil || node == mark {
		return false
	}
	return isPtrEqual(node.parentRef, l) && isPtrEqual(mark.parentRef, l)
}

// InsertBefore 方法将节点 n 插入到节点 mark 的前面。
// The InsertBefore method inserts node n before node mark.
func (l *List) InsertBefore(node, mark *Node) {
	if node == nil || mark == nil {
		return
	}

	node.parentRef = toUnsafePtr(l)
	node.Right = mark
	node.Left = mark.Left

	if mark.Left == nil {
		l.head = node
	} else {
		mark.Left.Right = node
	}
	mark.Left = node
	l.count++
}

// InsertAfter 方法将节点 n 插入到节点 mark 的后面。
// The InsertAfter method inserts node n after node mark.
func (l *List) InsertAfter(node, mark *Node) {
	if node == nil || mark == nil {
		return
	}

	node.parentRef = toUnsafePtr(l)
	node.Left = mark
	node.Right = mark.Right

	if mark.Right == nil {
		l.tail = node
	} else {
		mark.Right.Left = node
	}
	mark.Right = node
	l.count++
}

// Swap 方法交换链表中的两个节点 n 和 mark 的位置。
// The Swap method swaps the positions of two nodes n and mark in the list.
func (l *List) Swap(node, mark *Node) {
	if !l.validateSwapNodes(node, mark) {
		return
	}

	if node.Right == mark {
		l.Remove(node)
		l.InsertAfter(node, mark)
		return
	}

	if node.Left == mark {
		l.Remove(node)
		l.InsertBefore(node, mark)
		return
	}

	// 交换节点的链接
	node.Left, mark.Left = mark.Left, node.Left
	node.Right, mark.Right = mark.Right, node.Right

	// 更新相邻节点的链接
	if node.Left != nil {
		node.Left.Right = node
	} else {
		l.head = node
	}

	if node.Right != nil {
		node.Right.Left = node
	} else {
		l.tail = node
	}

	if mark.Left != nil {
		mark.Left.Right = mark
	} else {
		l.head = mark
	}

	if mark.Right != nil {
		mark.Right.Left = mark
	} else {
		l.tail = mark
	}
}

// Cleanup 方法清理链表，移除所有节点并重置链表的状态。
// The Cleanup method cleans up the list, removes all nodes and resets the state of the list.
func (l *List) Cleanup() {
	// 我们将头节点、尾节点和 count 都设置为它们的零值。
	// We set the head node, tail node, and count all to their zero values.
	l.head = nil
	l.tail = nil
	l.count = 0
}

// Range 方法遍历链表，对每个节点执行 fn 函数，如果 fn 返回 false，就停止遍历。
// The Range method traverses the list, performs the fn function on each node, and stops traversing if fn returns false.
func (l *List) Range(fn func(node *Node) bool) {
	// 我们从头节点开始，遍历整个链表。
	// We start from the head node and traverse the entire list.
	for iterNode := l.head; iterNode != nil; iterNode = iterNode.Right {
		// 我们对当前节点执行 fn，如果 fn 返回 false，我们就停止遍历。
		// We perform fn on the current node, if fn returns false, we stop traversing.
		if !fn(iterNode) {
			break
		}
	}
}

// Slice 方法将链表转换为一个切片，切片中的元素顺序和链表中的节点顺序一致。
// The Slice method converts the list to a slice, the order of the elements in the slice is consistent with the order of the nodes in the list.
func (l *List) Slice() []interface{} {
	// 我们创建一个空的切片，切片的容量为链的长度。
	// We create an empty slice, the capacity of the slice is the length of the list.
	nodes := make([]interface{}, 0, l.count)

	// 我们遍历链表，将每个节点的 Value 添加到切片中。
	// We traverse the list and add the Value of each node to the slice.
	l.Range(func(node *Node) bool {
		nodes = append(nodes, node.Value)
		return true
	})

	// 返回切片。
	// Return the slice.
	return nodes
}
