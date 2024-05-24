package list

import "unsafe"

func isPtrEqual(up unsafe.Pointer, lp *List) bool {
	return uintptr(up) == uintptr(unsafe.Pointer(lp))
}

func toUnsafePtr(lp *List) unsafe.Pointer {
	return unsafe.Pointer(lp)
}

type List struct {
	head, tail *Node
	count      int64
}

func New() *List { return &List{} }

func (l *List) Len() int64 { return l.count }

func (l *List) Front() *Node { return l.head }

func (l *List) Back() *Node { return l.tail }

func (l *List) PushBack(n *Node) {
	if n == nil {
		return
	}

	n.parentRef = toUnsafePtr(l)

	if l.head == nil {
		l.head = n
	} else {
		l.tail.Next = n
		n.Prev = l.tail
	}
	l.tail = n
	l.count++
}

func (l *List) PushFront(n *Node) {
	if n == nil {
		return
	}

	n.parentRef = toUnsafePtr(l)

	if l.head == nil {
		l.tail = n
	} else {
		l.head.Prev = n
		n.Next = l.head
	}
	l.head = n
	l.count++
}

func (l *List) PopBack() *Node {
	if l.tail == nil {
		return nil
	}
	n := l.tail
	l.tail = n.Prev
	if l.tail == nil {
		l.head = nil
	} else {
		l.tail.Next = nil
	}

	n.parentRef = nil
	n.Prev = nil
	n.Next = nil

	l.count--
	return n
}

func (l *List) PopFront() *Node {
	if l.head == nil {
		return nil
	}
	n := l.head
	l.head = n.Next
	if l.head == nil {
		l.tail = nil
	} else {
		l.head.Prev = nil
	}

	n.parentRef = nil
	n.Prev = nil
	n.Next = nil

	l.count--
	return n
}

func (l *List) Remove(n *Node) {
	if n == nil || l.count == 0 || !isPtrEqual(n.parentRef, l) {
		return
	}

	if n.Prev == nil {
		l.head = n.Next
	} else {
		n.Prev.Next = n.Next
	}

	if n.Next == nil {
		l.tail = n.Prev
	} else {
		n.Next.Prev = n.Prev
	}

	n.parentRef = nil
	n.Prev = nil
	n.Next = nil

	l.count--
}

func (l *List) MoveToFront(n *Node) {
	if n == nil {
		return
	}

	// If the list is empty, add the node as the first element
	if l.head == nil && l.tail == nil {
		n.parentRef = toUnsafePtr(l)
		n.Prev = nil
		n.Next = nil
		l.head = n
		l.tail = n
		l.count++
		return
	}

	// If the node is already at the front, no need to do anything
	if n == l.head {
		return
	}

	// Disconnect the node from its current position
	if n.Prev != nil {
		n.Prev.Next = n.Next
	} else if n != l.head {
		// The node is not in the list, add it to the front
		n.parentRef = toUnsafePtr(l)
		n.Next = l.head
		l.head.Prev = n
		l.head = n
		l.count++
		return
	}
	if n.Next != nil {
		n.Next.Prev = n.Prev
	} else {
		l.tail = n.Prev
	}

	// Move the node to the front
	n.Prev = nil
	n.Next = l.head
	l.head.Prev = n
	l.head = n
}

func (l *List) MoveToBack(n *Node) {
	if n == nil {
		return
	}

	// If the list is empty, add the node as the last element
	if l.head == nil && l.tail == nil {
		n.parentRef = toUnsafePtr(l)
		n.Prev = nil
		n.Next = nil
		l.head = n
		l.tail = n
		l.count++
		return
	}

	// If the node is already at the back, no need to do anything
	if n == l.tail {
		return
	}

	// Disconnect the node from its current position
	if n.Prev != nil {
		n.Prev.Next = n.Next
	} else if n != l.head {
		// The node is not in the list, add it to the back
		n.parentRef = toUnsafePtr(l)
		n.Prev = l.tail
		l.tail.Next = n
		l.tail = n
		l.count++
		return
	}
	if n.Next != nil {
		n.Next.Prev = n.Prev
	} else {
		l.head = n.Next
	}

	// Move the node to the back
	n.Prev = l.tail
	n.Next = nil
	if l.tail != nil {
		l.tail.Next = n
	}
	l.tail = n
}

func (l *List) InsertBefore(n, mark *Node) {
	if n == nil || mark == nil {
		return
	}

	n.parentRef = toUnsafePtr(l)

	if mark.Prev == nil {
		l.head = n
	} else {
		mark.Prev.Next = n
	}
	n.Prev = mark.Prev
	n.Next = mark
	mark.Prev = n
	l.count++
}

func (l *List) InsertAfter(n, mark *Node) {
	if n == nil || mark == nil {
		return
	}

	n.parentRef = toUnsafePtr(l)

	if mark.Next == nil {
		l.tail = n
	} else {
		mark.Next.Prev = n
	}
	n.Next = mark.Next
	n.Prev = mark
	mark.Next = n
	l.count++
}

func (l *List) Swap(n, mark *Node) {
	if n == nil || mark == nil || n == mark {
		return
	}

	// Check if n and mark are in the same list
	if !isPtrEqual(n.parentRef, l) || !isPtrEqual(mark.parentRef, l) {
		return
	}

	if n.Next == mark {
		l.Remove(n)
		l.InsertAfter(n, mark)
		return
	}

	if n.Prev == mark {
		l.Remove(n)
		l.InsertBefore(n, mark)
		return
	}

	n.Prev, mark.Prev = mark.Prev, n.Prev
	n.Next, mark.Next = mark.Next, n.Next

	if n.Prev != nil {
		n.Prev.Next = n
	} else {
		l.head = n
	}

	if n.Next != nil {
		n.Next.Prev = n
	} else {
		l.tail = n
	}

	if mark.Prev != nil {
		mark.Prev.Next = mark
	} else {
		l.head = mark
	}

	if mark.Next != nil {
		mark.Next.Prev = mark
	} else {
		l.tail = mark
	}
}

func (l *List) Cleanup() {
	for n := l.head; n != nil; {
		next := n.Next
		n.Next = nil
		n.Prev = nil
		n.Value = nil
		n = next
	}

	l.head = nil
	l.tail = nil
	l.count = 0
}

func (l *List) Range(fn func(n *Node) bool) {
	for n := l.head; n != nil; n = n.Next {
		if !fn(n) {
			break
		}
	}
}

func (l *List) Slice() []interface{} {
	s := make([]interface{}, 0, l.count)
	l.Range(func(n *Node) bool {
		s = append(s, n.Value)
		return true
	})
	return s
}
