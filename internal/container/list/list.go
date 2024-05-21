package list

type List struct {
	head, tail *Node
	count      int64
}

func New() *List { return &List{} }

func (l *List) Len() int64 { return l.count }

func (l *List) Front() *Node { return l.head }

func (l *List) Back() *Node { return l.tail }

func (l *List) PushBack(n *Node) {
	if l.head == nil {
		l.head = n
		l.tail = n
	} else {
		l.tail.Next = n
		l.tail = n
	}
	l.count++
}

func (l *List) PopFront() *Node {
	if l.head == nil {
		return nil
	}

	n := l.head
	l.head = n.Next
	l.count--

	if l.head == nil {
		l.tail = nil
	}

	return n
}

func (l *List) Remove(n *Node) {
	if l.head == nil {
		return
	}

	if l.head == n {
		l.head = n.Next
		l.count--

		if l.head == nil {
			l.tail = nil
		}

		return
	}

	prev := l.head
	for prev.Next != nil {
		if prev.Next == n {
			prev.Next = n.Next
			l.count--
			return
		}
		prev = prev.Next
	}
}

func (l *List) Swap(prev, n *Node) {
	if prev == nil || n == nil {
		return
	}

	if prev == n {
		return
	}

	if l.head == prev {
		l.head = n
	} else if l.head == n {
		l.head = prev
	}

	if l.tail == prev {
		l.tail = n
	} else if l.tail == n {
		l.tail = prev
	}

	if prev.Next == n {
		prev.Next = n.Next
		n.Next = prev
	} else if n.Next == prev {
		n.Next = prev.Next
		prev.Next = n
	} else {
		prev.Next, n.Next = n.Next, prev.Next
	}
}

func (l *List) InsertBefore(next, n *Node) {
	if next == nil {
		return
	}

	if l.head == next {
		n.Next = l.head
		l.head = n
		l.count++
		return
	}

	prev := l.head
	for prev.Next != nil {
		if prev.Next == next {
			n.Next = next
			prev.Next = n
			l.count++
			return
		}
		prev = prev.Next
	}
}

func (l *List) InsertAfter(prev, n *Node) {
	if prev == nil {
		return
	}

	n.Next = prev.Next
	prev.Next = n
	l.count++
}

func (l *List) MoveToFront(n *Node) {
	if l.head == nil {
		return
	}

	if l.head == n {
		return
	}

	l.Remove(n)
	l.InsertBefore(l.head, n)

	l.head = n
	l.tail.Next = nil
}

func (l *List) MoveToBack(n *Node) {
	if l.head == nil {
		return
	}

	if l.tail == n {
		return
	}

	l.Remove(n)
	l.InsertAfter(l.tail, n)

	l.tail = n
	l.tail.Next = nil
}

func (l *List) Range(fn func(n *Node) bool) {
	for n := l.head; n != nil; n = n.Next {
		if !fn(n) {
			break
		}
	}
}

func (l *List) Cleanup() {
	for l.head != nil {
		n := l.head
		l.head = n.Next
		n.Reset()
	}

	l.head = nil
	l.tail = nil
	l.count = 0
}
