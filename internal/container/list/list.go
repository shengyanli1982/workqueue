package list

import "unsafe"

// parentRef 存储所属链表地址，用于 O(1) 判断节点归属关系。
func isPtrEqual(up unsafe.Pointer, lp *List) bool {
	return up == unsafe.Pointer(lp)
}

func toUnsafePtr(lp *List) unsafe.Pointer {
	return unsafe.Pointer(lp)
}

// List 是一个双向链表实现，支持 O(1) 的头尾访问、插入、删除，
// 以及节点的移动、交换和任意位置插入。节点通过 parentRef 字段实现 O(1) 归属判断。
type List struct {
	head, tail *Node
	count      int64
}

// New 创建一个空的双向链表。
func New() *List { return &List{} }

// Len 返回链表中的节点数量。
func (l *List) Len() int64 { return l.count }

// Front 返回链表的头节点（首节点），链表为空时返回 nil。
func (l *List) Front() *Node { return l.head }

// Back 返回链表的尾节点（末节点），链表为空时返回 nil。
func (l *List) Back() *Node { return l.tail }

// PushBack 将节点追加到链表尾部。若节点已在当前链表中，则执行移动操作避免重复挂接；
// 若节点属于其他链表，则从原链表脱离后重新挂接。
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

// PushFront 将节点插入链表头部。若节点已在当前链表中，则执行移动操作避免重复挂接；
// 若节点属于其他链表，则从原链表脱离后重新挂接。
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

// PopBack 弹出并返回链表尾部节点，同时清除其链表归属标记。链表为空时返回 nil。
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

// PopFront 弹出并返回链表头部节点，同时清除其链表归属标记。链表为空时返回 nil。
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

// Remove 从链表中移除指定节点并清除其归属标记。节点为空、链表为空或节点不属于当前链表时不执行操作。
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

// initNodeInEmptyList 尝试在空链表中初始化节点为唯一节点。链表非空时返回 false。
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

// MoveToFront 将节点移动到链表头部。若节点不属于当前链表，则从原位置脱离后插入头部；
// 若节点已在头部则不执行操作。
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

// MoveToBack 将节点移动到链表尾部。若节点不属于当前链表，则从原位置脱离后插入尾部；
// 若节点已在尾部则不执行操作。
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

// validateSwapNodes 校验两个交换节点的有效性：均非空、互不相同且均属于当前链表。
func (l *List) validateSwapNodes(node, mark *Node) bool {
	if node == nil || mark == nil || node == mark {
		return false
	}
	return isPtrEqual(node.parentRef, l) && isPtrEqual(mark.parentRef, l)
}

// InsertBefore 将 node 插入到 mark 节点之前。mark 不属于当前链表、参数为空或两节点相同时不执行操作；
// node 已在当前链表中时先从原位置移除再插入。
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

// InsertAfter 将 node 插入到 mark 节点之后。mark 不属于当前链表、参数为空或两节点相同时不执行操作；
// node 已在当前链表中时先从原位置移除再插入。
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

// Swap 交换链表中两个节点的位置。两节点相邻和非相邻分别走不同的指针交换路径，
// 均保持 head/tail 的正确性。
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

// Range 从头部到尾部遍历链表所有节点，fn 返回 false 时提前终止遍历。
func (l *List) Range(fn func(node *Node) bool) {

	for iterNode := l.head; iterNode != nil; iterNode = iterNode.Right {

		if !fn(iterNode) {
			break
		}
	}
}

// Slice 将链表中所有节点的值从头到尾收集到切片并返回。
func (l *List) Slice() []any {
	nodes := make([]any, 0, l.count)

	l.Range(func(node *Node) bool {
		nodes = append(nodes, node.Value)
		return true
	})

	return nodes
}

// Cleanup 清空链表的所有节点引用，将链表重置为空状态。注意：节点本身不会被释放，
// 调用方需自行管理节点的生命周期。
func (l *List) Cleanup() {
	l.head = nil
	l.tail = nil
	l.count = 0
}
