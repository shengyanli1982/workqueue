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
func (l *List) PushBack(n *Node) {
	// 如果 n 是 nil，我们就直接返回，不做任何操作。
	// If n is nil, we just return directly without doing anything.
	if n == nil {
		return
	}

	// 我们将 n 的 parentRef 设置为 l 的地址。
	// We set n's parentRef to the address of l.
	n.parentRef = toUnsafePtr(l)

	// 如果链表是空的，我们就将 n 设置为头节点。
	// If the list is empty, we set n as the head node.
	if l.head == nil {
		l.head = n
	} else {
		// 否则，我们将 n 添加到尾节点的后面，并更新尾节点。
		// Otherwise, we add n after the tail node and update the tail node.
		l.tail.Next = n
		n.Prev = l.tail
	}

	// 更新尾节点为 n。
	// Update the tail node to n.
	l.tail = n

	// 链表长度加 1。
	// Increase the length of the list by 1.
	l.count++
}

// PushFront 方法将一个节点添加到链表的头部。
// The PushFront method adds a node to the beginning of the list.
func (l *List) PushFront(n *Node) {
	// 如果 n 是 nil，我们就直接返回，不做任何操作。
	// If n is nil, we just return directly without doing anything.
	if n == nil {
		return
	}

	// 我们将 n 的 parentRef 设置为 l 的地址。
	// We set n's parentRef to the address of l.
	n.parentRef = toUnsafePtr(l)

	// 如果链表是空的，我们就将 n 设置为尾节点。
	// If the list is empty, we set n as the tail node.
	if l.head == nil {
		l.tail = n
	} else {
		// 否则，我们将 n 添加到头节点的前面，并更新头节点。
		// Otherwise, we add n before the head node and update the head node.
		l.head.Prev = n
		n.Next = l.head
	}

	// 更新头节点为 n。
	// Update the head node to n.
	l.head = n

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
	l.tail = n.Prev

	// 如果新的尾节点是 nil，说明链表现在是空的，我们将头节点也设置为 nil。
	// If the new tail node is nil, it means the list is now empty, we also set the head node to nil.
	if l.tail == nil {
		l.head = nil
	} else {
		// 否则，我们将新的尾节点的 Next 设置为 nil。
		// Otherwise, we set the Next of the new tail node to nil.
		l.tail.Next = nil
	}

	// 我们将 n 的 parentRef、Prev 和 Next 都设置为 nil。
	// We set n's parentRef, Prev, and Next all to nil.
	n.parentRef = nil
	n.Prev = nil
	n.Next = nil

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
	l.head = n.Next

	// 如果新的头节点是 nil，说明链表现在是空的，我们将尾节点也设置为 nil。
	// If the new head node is nil, it means the list is now empty, we also set the tail node to nil.
	if l.head == nil {
		l.tail = nil
	} else {
		// 否则，我们将新的头节点的 Prev 设置为 nil。
		// Otherwise, we set the Prev of the new head node to nil.
		l.head.Prev = nil
	}

	// 我们将 n 的 parentRef、Prev 和 Next 都设置为 nil。
	// We set n's parentRef, Prev, and Next all to nil.
	n.parentRef = nil
	n.Prev = nil
	n.Next = nil

	// 链表长度减 1。
	// Decrease the length of the list by 1.
	l.count--

	// 返回被移除的节点。
	// Return the removed node.
	return n
}

// Remove 方法从链表中移除节点 n。
// The Remove method removes node n from the list.
func (l *List) Remove(n *Node) {
	// 如果 n 是 nil，或者链表是空的，或者 n 不属于链表 l，我们就直接返回，不做任何操作。
	// If n is nil, or the list is empty, or n does not belong to list l, we just return directly without doing anything.
	if n == nil || l.count == 0 || !isPtrEqual(n.parentRef, l) {
		return
	}

	// 如果 n 的 Prev 是 nil，说明 n 是头节点，我们就将链表的头节点设置为 n 的 Next。
	// If n's Prev is nil, it means n is the head node, we set the head node of the list to n's Next.
	if n.Prev == nil {
		l.head = n.Next
	} else {
		// 否则，我们将 n 的 Prev 的 Next 设置为 n 的 Next。
		// Otherwise, we set n's Prev's Next to n's Next.
		n.Prev.Next = n.Next
	}

	// 如果 n 的 Next 是 nil，说明 n 是尾节点，我们就将链表的尾节点设置为 n 的 Prev。
	// If n's Next is nil, it means n is the tail node, we set the tail node of the list to n's Prev.
	if n.Next == nil {
		l.tail = n.Prev
	} else {
		// 否则，我们将 n 的 Next 的 Prev 设置为 n 的 Prev。
		// Otherwise, we set n's Next's Prev to n's Prev.
		n.Next.Prev = n.Prev
	}

	// 我们将 n 的 parentRef、Prev 和 Next 都设置为 nil。
	// We set n's parentRef, Prev and Next all to nil.
	n.parentRef = nil
	n.Prev = nil
	n.Next = nil

	// 链表长度减 1。
	// Decrease the length of the list by 1.
	l.count--
}

// MoveToFront 方法将节点 n 移动到链表的头部。
// The MoveToFront method moves node n to the front of the list.
func (l *List) MoveToFront(n *Node) {
	// 如果 n 是 nil，我们就直接返回，不做任何操作。
	// If n is nil, we just return directly without doing anything.
	if n == nil {
		return
	}

	// 如果链表是空的，我们就将 n 设置为头节点和尾节点，然后返回。
	// If the list is empty, we set n as the head node and tail node, and then return.
	if l.head == nil && l.tail == nil {
		n.parentRef = toUnsafePtr(l)
		n.Prev = nil
		n.Next = nil
		l.head = n
		l.tail = n
		l.count++
		return
	}

	// 如果 n 已经是头节点，我们就直接返回，不做任何操作。
	// If n is already the head node, we just return directly without doing anything.
	if n == l.head {
		return
	}

	// 如果 n 的 Prev 不是 nil，我们就更新 n 的 Prev 的 Next 为 n 的 Next。
	// If n's Prev is not nil, we update n's Prev's Next to n's Next.
	if n.Prev != nil {
		n.Prev.Next = n.Next
	} else if n != l.head {
		// 如果 n 不是头节点，我们就将 n 插入到链表的头部，然后返回。
		// If n is not the head node, we insert n at the front of the list, and then return.
		n.parentRef = toUnsafePtr(l)
		n.Next = l.head
		l.head.Prev = n
		l.head = n
		l.count++
		return
	}

	// 如果 n 的 Next 不是 nil，我们就更新 n 的 Next 的 Prev 为 n 的 Prev。
	// If n's Next is not nil, we update n's Next's Prev to n's Prev.
	if n.Next != nil {
		n.Next.Prev = n.Prev
	} else {
		// 否则，我们将链表的尾节点设置为 n 的 Prev。
		// Otherwise, we set the tail node of the list to n's Prev.
		l.tail = n.Prev
	}

	// 我们将 n 的 Prev 设置为 nil，将 n 的 Next 设置为链表的头节点。
	// We set n's Prev to nil, and n's Next to the head node of the list.
	n.Prev = nil
	n.Next = l.head

	// 我们将链表的头节点的 Prev 设置为 n，将链表的头节点设置为 n。
	// We set the head node's Prev of the list to n, and the head node of the list to n.
	l.head.Prev = n
	l.head = n
}

// MoveToBack 方法将节点 n 移动到链表的尾部。
// The MoveToBack method moves node n to the end of the list.
func (l *List) MoveToBack(n *Node) {
	// 如果 n 是 nil，我们就直接返回，不做任何操作。
	// If n is nil, we just return directly without doing anything.
	if n == nil {
		return
	}

	// 如果链表是空的，我们就将 n 设置为头节点和尾节点，然后返回。
	// If the list is empty, we set n as the head node and tail node, and then return.
	if l.head == nil && l.tail == nil {
		n.parentRef = toUnsafePtr(l)
		n.Prev = nil
		n.Next = nil
		l.head = n
		l.tail = n
		l.count++
		return
	}

	// 如果 n 已经是尾节点，我们就直接返回，不做任何操作。
	// If n is already the tail node, we just return directly without doing anything.
	if n == l.tail {
		return
	}

	// 如果 n 的 Prev 不是 nil，我们就更新 n 的 Prev 的 Next 为 n 的 Next。
	// If n's Prev is not nil, we update n's Prev's Next to n's Next.
	if n.Prev != nil {
		n.Prev.Next = n.Next
	} else if n != l.head {
		// 如果 n 不是头节点，我们就将 n 插入到链表的尾部，然后返回。
		// If n is not the head node, we insert n at the end of the list, and then return.
		n.parentRef = toUnsafePtr(l)
		n.Prev = l.tail
		l.tail.Next = n
		l.tail = n
		l.count++
		return
	}

	// 如果 n 的 Next 不是 nil，我们就更新 n 的 Next 的 Prev 为 n 的 Prev。
	// If n's Next is not nil, we update n's Next's Prev to n's Prev.
	if n.Next != nil {
		n.Next.Prev = n.Prev
	} else {
		// 否则，我们将链表的头节点设置为 n 的 Next。
		// Otherwise, we set the head node of the list to n's Next.
		l.head = n.Next
	}

	// 我们将 n 的 Prev 设置为链表的尾节点，将 n 的 Next 设置为 nil。
	// We set n's Prev to the tail node of the list, and n's Next to nil.
	n.Prev = l.tail
	n.Next = nil

	// 如果链表的尾节点不是 nil，我们就更新链表的尾节点的 Next 为 n。
	// If the tail node of the list is not nil, we update the tail node's Next to n.
	if l.tail != nil {
		l.tail.Next = n
	}

	// 我们将链表的尾节点设置为 n。
	// We set the tail node of the list to n.
	l.tail = n
}

// InsertBefore 方法将节点 n 插入到节点 mark 的前面。
// The InsertBefore method inserts node n before node mark.
func (l *List) InsertBefore(n, mark *Node) {
	// 如果 n 或 mark 是 nil，我们就直接返回，不做任何操作。
	// If n or mark is nil, we just return directly without doing anything.
	if n == nil || mark == nil {
		return
	}

	// 我们将 n 的 parentRef 设置为链表 l。
	// We set n's parentRef to list l.
	n.parentRef = toUnsafePtr(l)

	// 如果 mark 的 Prev 是 nil，说明 mark 是头节点，我们就将 n 插入到链表的头部。
	// If mark's Prev is nil, it means mark is the head node, we insert n at the head of the list.
	if mark.Prev == nil {
		n.Prev = nil
		l.head = n
	} else {
		// 否则，我们将 n 插入到 mark 的前面。
		// Otherwise, we insert n before mark.
		n.Prev = mark.Prev
		mark.Prev.Next = n
	}

	// 我们将 n 的 Next 设置为 mark，将 mark 的 Prev 设置为 n。
	// We set n's Next to mark, and mark's Prev to n.
	n.Next = mark
	mark.Prev = n

	// 链表长度加 1。
	// Increase the length of the list by 1.
	l.count++
}

// InsertAfter 方法将节点 n 插入到节点 mark 的后面。
// The InsertAfter method inserts node n after node mark.
func (l *List) InsertAfter(n, mark *Node) {
	// 如果 n 或 mark 是 nil，我们就直接返回，不做任何操作。
	// If n or mark is nil, we just return directly without doing anything.
	if n == nil || mark == nil {
		return
	}

	// 我们将 n 的 parentRef 设置为链表 l。
	// We set n's parentRef to list l.
	n.parentRef = toUnsafePtr(l)

	// 如果 mark 的 Next 是 nil，说明 mark 是尾节点，我们就将 n 插入到链表的尾部。
	// If mark's Next is nil, it means mark is the tail node, we insert n at the end of the list.
	if mark.Next == nil {
		n.Next = nil
		l.tail = n
	} else {
		// 否则，我们将 n 插入到 mark 的后面。
		// Otherwise, we insert n after mark.
		n.Next = mark.Next
		mark.Next.Prev = n
	}

	// 我们将 n 的 Prev 设置为 mark，将 mark 的 Next 设置为 n。
	// We set n's Prev to mark, and mark's Next to n.
	n.Prev = mark
	mark.Next = n

	// 链表长度加 1。
	// Increase the length of the list by 1.
	l.count++
}

// Swap 方法交换链表中的两个节点 n 和 mark 的位置。
// The Swap method swaps the positions of two nodes n and mark in the list.
func (l *List) Swap(n, mark *Node) {
	// 如果 n 或 mark 是 nil，或者 n 和 mark 是同一个节点，我们就直接返回，不做任何操作。
	// If n or mark is nil, or n and mark are the same node, we just return directly without doing anything.
	if n == nil || mark == nil || n == mark {
		return
	}

	// 如果 n 或 mark 不是链表 l 的节点，我们就直接返回，不做任何操作。
	// If n or mark is not a node of list l, we just return directly without doing anything.
	if !isPtrEqual(n.parentRef, l) || !isPtrEqual(mark.parentRef, l) {
		return
	}

	// 如果 n 是 mark 的前一个节点，我们就移除 n，然后将 n 插入到 mark 的后面。
	// If n is the previous node of mark, we remove n and then insert n after mark.
	if n.Next == mark {
		l.Remove(n)
		l.InsertAfter(n, mark)
		return
	}

	// 如果 n 是 mark 的后一个节点，我们就移除 n，然后将 n 插入到 mark 的前面。
	// If n is the next node of mark, we remove n and then insert n before mark.
	if n.Prev == mark {
		l.Remove(n)
		l.InsertBefore(n, mark)
		return
	}

	// 我们交换 n 和 mark 的 Prev 和 Next。
	// We swap the Prev and Next of n and mark.
	n.Prev, mark.Prev = mark.Prev, n.Prev
	n.Next, mark.Next = mark.Next, n.Next

	// 如果 n 的 Prev 不是 nil，我们就更新 n 的 Prev 的 Next 为 n，否则，我们将链表的头节点设置为 n。
	// If n's Prev is not nil, we update n's Prev's Next to n, otherwise, we set the head node of the list to n.
	if n.Prev != nil {
		n.Prev.Next = n
	} else {
		l.head = n
	}

	// 如果 n 的 Next 不是 nil，我们就更新 n 的 Next 的 Prev 为 n，否则，我们将链表的尾节点设置为 n。
	// If n's Next is not nil, we update n's Next's Prev to n, otherwise, we set the tail node of the list to n.
	if n.Next != nil {
		n.Next.Prev = n
	} else {
		l.tail = n
	}

	// 如果 mark 的 Prev 不是 nil，我们就更新 mark 的 Prev 的 Next 为 mark，否则，我们将链表的头节点设置为 mark。
	// If mark's Prev is not nil, we update mark's Prev's Next to mark, otherwise, we set the head node of the list to mark.
	if mark.Prev != nil {
		mark.Prev.Next = mark
	} else {
		l.head = mark
	}

	// 如果 mark 的 Next 不是 nil，我们就更新 mark 的 Next 的 Prev 为 mark，否则，我们将链表的尾节点设置为 mark。
	// If mark's Next is not nil, we update mark's Next's Prev to mark, otherwise, we set the tail node of the list to mark.
	if mark.Next != nil {
		mark.Next.Prev = mark
	} else {
		l.tail = mark
	}
}

// Cleanup 方法清理链表，移除所有节点并重置链表的状态。
// The Cleanup method cleans up the list, removes all nodes and resets the state of the list.
func (l *List) Cleanup() {
	// 我们从头节点开始，遍历整个链表。
	// We start from the head node and traverse the entire list.
	for n := l.head; n != nil; {
		// 我们保存下一个节点的引用。
		// We save the reference to the next node.
		next := n.Next

		// 我们将当前节点的 Next、Prev 和 Value 都设置为 nil。
		// We set the current node's Next, Prev, and Value all to nil.
		n.Next = nil
		n.Prev = nil
		n.Value = nil

		// 我们将 n 更新为下一个节点。
		// We update n to the next node.
		n = next
	}

	// 我们将头节点、尾节点和 count 都设置为它们的零值。
	// We set the head node, tail node, and count all to their zero values.
	l.head = nil
	l.tail = nil
	l.count = 0
}

// Range 方法遍历链表，对每个节点执行 fn 函数，如果 fn 返回 false，就停止遍历。
// The Range method traverses the list, performs the fn function on each node, and stops traversing if fn returns false.
func (l *List) Range(fn func(n *Node) bool) {
	// 我们从头节点开始，遍历整个链表。
	// We start from the head node and traverse the entire list.
	for n := l.head; n != nil; n = n.Next {
		// 我们对当前节点执行 fn，如果 fn 返回 false，我们就停止遍历。
		// We perform fn on the current node, if fn returns false, we stop traversing.
		if !fn(n) {
			break
		}
	}
}

// Slice 方法将链表转换为一个切片，切片中的元素顺序和链表中的节点顺序一致。
// The Slice method converts the list to a slice, the order of the elements in the slice is consistent with the order of the nodes in the list.
func (l *List) Slice() []interface{} {
	// 我们创建一个空的切片，切片的容量为链表的长度。
	// We create an empty slice, the capacity of the slice is the length of the list.
	s := make([]interface{}, 0, l.count)

	// 我们遍历链表，将每个节点的 Value 添加到切片中。
	// We traverse the list and add the Value of each node to the slice.
	l.Range(func(n *Node) bool {
		s = append(s, n.Value)
		return true
	})

	// 返回切片。
	// Return the slice.
	return s
}
