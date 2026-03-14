package list

import "unsafe"

// parentRef 存储所属链表地址，用于 O(1) 判断节点归属关系。
func isPtrEqual(up unsafe.Pointer, lp *List) bool {
	return up == unsafe.Pointer(lp)
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

func (l *List) PushBack(node *Node) {
	if node == nil {
		return
	}

	if node.parentRef != nil {
		if isPtrEqual(node.parentRef, l) {
			// 已在当前链表中时复用移动逻辑，避免重复挂接。
			l.MoveToBack(node)
			return
		}
	}

	node.parentRef = toUnsafePtr(l)
	node.Right = nil

	if l.head == nil {
		l.head = node
		node.Left = nil
	} else {

		l.tail.Right = node
		node.Left = l.tail
	}

	l.tail = node
	l.count++
}

func (l *List) PushFront(node *Node) {
	if node == nil {
		return
	}

	if node.parentRef != nil {
		if isPtrEqual(node.parentRef, l) {
			l.MoveToFront(node)
			return
		}
	}

	node.parentRef = toUnsafePtr(l)
	node.Left = nil

	if l.head == nil {
		l.tail = node
		node.Right = nil
	} else {
		l.head.Left = node
		node.Right = l.head
	}

	l.head = node
	l.count++
}

func (l *List) PopBack() *Node {
	if l.tail == nil {
		return nil
	}

	n := l.tail
	l.tail = n.Left

	if l.tail == nil {
		l.head = nil
	} else {
		l.tail.Right = nil
	}

	n.parentRef = nil
	n.Right = nil

	l.count--

	return n
}

func (l *List) PopFront() *Node {
	if l.head == nil {
		return nil
	}

	n := l.head
	l.head = n.Right

	if l.head == nil {
		l.tail = nil
	} else {
		l.head.Left = nil
	}

	n.parentRef = nil
	n.Left = nil

	l.count--

	return n
}

func (l *List) Remove(node *Node) {

	if node == nil || l.count == 0 || !isPtrEqual(node.parentRef, l) {
		return
	}

	if node.Left == nil {
		l.head = node.Right
	} else {
		node.Left.Right = node.Right
	}

	if node.Right == nil {
		l.tail = node.Left
	} else {
		node.Right.Left = node.Left
	}

	node.parentRef = nil
	node.Left = nil
	node.Right = nil

	l.count--
}

func (l *List) initNodeInEmptyList(node *Node) bool {
	if l.head == nil {
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

	if !isPtrEqual(node.parentRef, l) {
		node.parentRef = toUnsafePtr(l)
		node.Right = l.head
		l.head.Left = node
		node.Left = nil
		l.head = node
		l.count++
		return
	}

	if node.Left != nil {
		node.Left.Right = node.Right
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

	if !isPtrEqual(node.parentRef, l) {
		node.parentRef = toUnsafePtr(l)
		node.Left = l.tail
		node.Right = nil
		l.tail.Right = node
		l.tail = node
		l.count++
		return
	}

	if node.Left != nil {
		node.Left.Right = node.Right
	} else {
		l.head = node.Right
	}
	if node.Right != nil {
		node.Right.Left = node.Left
	}

	node.Right = nil
	node.Left = l.tail
	l.tail.Right = node
	l.tail = node
}

func (l *List) validateSwapNodes(node, mark *Node) bool {
	if node == nil || mark == nil || node == mark {
		return false
	}
	return isPtrEqual(node.parentRef, l) && isPtrEqual(mark.parentRef, l)
}

func (l *List) InsertBefore(node, mark *Node) {
	if node == nil || mark == nil || node == mark {
		return
	}

	if !isPtrEqual(mark.parentRef, l) {
		return
	}

	if isPtrEqual(node.parentRef, l) {
		l.Remove(node)
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

func (l *List) InsertAfter(node, mark *Node) {
	if node == nil || mark == nil || node == mark {
		return
	}

	if !isPtrEqual(mark.parentRef, l) {
		return
	}

	if isPtrEqual(node.parentRef, l) {
		l.Remove(node)
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

func (l *List) Swap(node, mark *Node) {

	if !l.validateSwapNodes(node, mark) {
		return
	}

	if node.Right == mark {
		prev := node.Left
		next := mark.Right

		if prev != nil {
			prev.Right = mark
		} else {
			l.head = mark
		}
		mark.Left = prev
		mark.Right = node
		node.Left = mark
		node.Right = next
		if next != nil {
			next.Left = node
		} else {
			l.tail = node
		}
		return
	}
	if node.Left == mark {
		prev := mark.Left
		next := node.Right

		if prev != nil {
			prev.Right = node
		} else {
			l.head = node
		}
		node.Left = prev
		node.Right = mark
		mark.Left = node
		mark.Right = next
		if next != nil {
			next.Left = mark
		} else {
			l.tail = mark
		}
		return
	}

	node.Left, mark.Left = mark.Left, node.Left
	node.Right, mark.Right = mark.Right, node.Right

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

func (l *List) Range(fn func(node *Node) bool) {

	for iterNode := l.head; iterNode != nil; iterNode = iterNode.Right {

		if !fn(iterNode) {
			break
		}
	}
}

func (l *List) Slice() []interface{} {
	nodes := make([]interface{}, 0, l.count)

	l.Range(func(node *Node) bool {
		nodes = append(nodes, node.Value)
		return true
	})

	return nodes
}

func (l *List) Cleanup() {
	l.head = nil
	l.tail = nil
	l.count = 0
}
