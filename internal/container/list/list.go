package list

import "unsafe"

// Utility functions / 实用函数

// isPtrEqual checks if two pointers point to the same List
// isPtrEqual 检查两个指针是否指向同一个 List
func isPtrEqual(up unsafe.Pointer, lp *List) bool {
	return uintptr(up) == uintptr(unsafe.Pointer(lp))
}

// toUnsafePtr converts a List pointer to unsafe.Pointer
// toUnsafePtr 将 List 指针转换为 unsafe.Pointer
func toUnsafePtr(lp *List) unsafe.Pointer {
	return unsafe.Pointer(lp)
}

// List represents a doubly linked list
// List 表示一个双向链表
type List struct {
	head, tail *Node // First and last nodes of the list / 链表的首尾节点
	count      int64 // Number of nodes in the list / 链表中的节点数量
}

// Basic operations / 基本操作

// New creates a new List
// New 创建一个新的 List
func New() *List { return &List{} }

// Len returns the number of nodes in the list
// Len 返回链表中的节点数量
func (l *List) Len() int64 { return l.count }

// Front returns the first node in the list
// Front 返回链表中的第一个节点
func (l *List) Front() *Node { return l.head }

// Back returns the last node in the list
// Back 返回链表中的最后一个节点
func (l *List) Back() *Node { return l.tail }

// PushBack adds a node to the end of the list
// PushBack 将节点添加到链表末尾
func (l *List) PushBack(node *Node) {
	if node == nil {
		return
	}

	// If node already exists in this list, remove it first
	// 如果节点已经在此链表中，先将其移除
	if isPtrEqual(node.parentRef, l) {
		l.Remove(node)
	}

	// Set the parent reference and clear right pointer
	// 设置父引用并清除右指针
	node.parentRef = toUnsafePtr(l)
	node.Right = nil

	// If list is empty, set as head node
	// 如果链表为空，设置为头节点
	if l.head == nil {
		l.head = node
		node.Left = nil
	} else {
		// Append to existing tail
		// 追加到现有的尾部
		l.tail.Right = node
		node.Left = l.tail
	}

	l.tail = node
	l.count++
}

// PushFront adds a node to the front of the list
// PushFront 将节点添加到链表头部
func (l *List) PushFront(node *Node) {
	if node == nil {
		return
	}

	if isPtrEqual(node.parentRef, l) {
		l.Remove(node)
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

// PopBack removes the last node from the list and returns it
// PopBack 从链表末尾移除最后一个节点并返回它
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
	n.Left = nil
	n.Right = nil

	l.count--

	return n
}

// PopFront removes the first node from the list and returns it
// PopFront 从链表头部移除第一个节点并返回它
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
	n.Right = nil

	l.count--

	return n
}

// Remove removes the given node from the list
// Remove 从链表中移除给定的节点
func (l *List) Remove(node *Node) {
	// Validate node can be removed
	// 验证节点是否可以被移除
	if node == nil || l.count == 0 || !isPtrEqual(node.parentRef, l) {
		return
	}

	// Update head if removing first node
	// 如果移除第一个节点，更新头节点
	if node.Left == nil {
		l.head = node.Right
	} else {
		node.Left.Right = node.Right
	}

	// Update tail if removing last node
	// 如果移除最后一个节点，更新尾节点
	if node.Right == nil {
		l.tail = node.Left
	} else {
		node.Right.Left = node.Left
	}

	// Clear node references
	// 清除节点引用
	node.parentRef = nil
	node.Left = nil
	node.Right = nil

	l.count--
}

// initNodeInEmptyList checks if the given node is the first node in an empty list
// initNodeInEmptyList 检查给定的节点是否是空链表中的第一个节点
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

// MoveToFront moves the given node to the front of the list
// MoveToFront 将给定节点移动到链表头部
func (l *List) MoveToFront(node *Node) {
	if node == nil {
		return
	}

	// Try to initialize if list is empty
	// 如果链表为空，尝试初始化
	if l.initNodeInEmptyList(node) {
		return
	}

	// Already at front
	// 已经在头部，无需移动
	if node == l.head {
		return
	}

	// If node is from another list, add it to front
	// 如果节点来自其他链表，将其添加到头部
	if !isPtrEqual(node.parentRef, l) {
		node.parentRef = toUnsafePtr(l)
		node.Right = l.head
		l.head.Left = node
		node.Left = nil
		l.head = node
		l.count++
		return
	}

	// Relink surrounding nodes
	// 重新链接周围的节点
	if node.Left != nil {
		node.Left.Right = node.Right
	}
	if node.Right != nil {
		node.Right.Left = node.Left
	} else {
		l.tail = node.Left
	}

	// Move to front
	// 移动到头部
	node.Left = nil
	node.Right = l.head
	l.head.Left = node
	l.head = node
}

// MoveToBack moves the given node to the back of the list
// MoveToBack 将给定节点移动到链表尾部
func (l *List) MoveToBack(node *Node) {
	if node == nil {
		return
	}

	// Try to initialize if list is empty
	// 如果链表为空，尝试初始化
	if l.initNodeInEmptyList(node) {
		return
	}

	// Already at back, no need to move
	// 已经在尾部，无需移动
	if node == l.tail {
		return
	}

	// If node is from another list, add it to back
	// 如果节点来自其他链表，将其添加到尾部
	if !isPtrEqual(node.parentRef, l) {
		node.parentRef = toUnsafePtr(l)
		node.Left = l.tail
		node.Right = nil
		l.tail.Right = node
		l.tail = node
		l.count++
		return
	}

	// Update surrounding nodes
	// 更新周围节点
	if node.Left != nil {
		node.Left.Right = node.Right
	} else {
		l.head = node.Right
	}
	if node.Right != nil {
		node.Right.Left = node.Left
	}

	// Move to back
	// 移动到尾部
	node.Right = nil
	node.Left = l.tail
	l.tail.Right = node
	l.tail = node
}

// validateSwapNodes checks if two nodes can be swapped
// validateSwapNodes 检查两个节点是否可以交换
func (l *List) validateSwapNodes(node, mark *Node) bool {
	if node == nil || mark == nil || node == mark {
		return false
	}
	return isPtrEqual(node.parentRef, l) && isPtrEqual(mark.parentRef, l)
}

// InsertBefore inserts a node before the given mark node
// InsertBefore 在给定标记节点之前插入一个节点
func (l *List) InsertBefore(node, mark *Node) {
	if node == nil || mark == nil {
		return
	}

	// Verify mark node belongs to this list
	// 验证标记节点属于此链表
	if !isPtrEqual(mark.parentRef, l) {
		return
	}

	// If node already exists in this list, remove it first
	// 如果节点已经在此链表中，先将其移除
	if isPtrEqual(node.parentRef, l) {
		l.Remove(node)
	}

	// Set up node's references
	// 设置节点的引用
	node.parentRef = toUnsafePtr(l)
	node.Right = mark
	node.Left = mark.Left

	// Update surrounding nodes
	// 更新周围节点
	if mark.Left == nil {
		l.head = node
	} else {
		mark.Left.Right = node
	}
	mark.Left = node
	l.count++
}

// InsertAfter inserts a node after the given mark node
// InsertAfter 在给定标记节点之后插入一个节点
func (l *List) InsertAfter(node, mark *Node) {
	if node == nil || mark == nil {
		return
	}

	// Verify mark node belongs to this list
	// 验证标记节点属于此链表
	if !isPtrEqual(mark.parentRef, l) {
		return
	}

	// If node already exists in this list, remove it first
	// 如果节点已经在此链表中，先将其移除
	if isPtrEqual(node.parentRef, l) {
		l.Remove(node)
	}

	// Set up node's references
	// 设置节点的引用
	node.parentRef = toUnsafePtr(l)
	node.Left = mark
	node.Right = mark.Right

	// Update surrounding nodes
	// 更新周围节点
	if mark.Right == nil {
		l.tail = node
	} else {
		mark.Right.Left = node
	}
	mark.Right = node
	l.count++
}

// Swap swaps two nodes in the list
// Swap 交换链表中的两个节点
func (l *List) Swap(node, mark *Node) {
	// Validate nodes can be swapped
	// 验证节点是否可以交换
	if !l.validateSwapNodes(node, mark) {
		return
	}

	// Handle adjacent nodes specially
	// 特殊处理相邻节点
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

	// Swap node positions
	// 交换节点位置
	node.Left, mark.Left = mark.Left, node.Left
	node.Right, mark.Right = mark.Right, node.Right

	// Update surrounding nodes' references
	// 更新周围节点的引用
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

// Range iterates through the list and calls fn for each node
// Range 遍历链表并为每个节点调用 fn 函数
func (l *List) Range(fn func(node *Node) bool) {
	// Iterate from head to tail
	// 从头到尾遍历
	for iterNode := l.head; iterNode != nil; iterNode = iterNode.Right {
		// Call fn and check if iteration should continue
		// 调用 fn 并检查是否应继续迭代
		if !fn(iterNode) {
			break
		}
	}
}

// Slice returns a slice containing all values in the list
// Slice 返回包含链表中所有值的切片
func (l *List) Slice() []interface{} {
	nodes := make([]interface{}, 0, l.count)

	l.Range(func(node *Node) bool {
		nodes = append(nodes, node.Value)
		return true
	})

	return nodes
}

// Cleanup resets the list to empty state
// Cleanup 重置链表为空状态
func (l *List) Cleanup() {
	l.head = nil
	l.tail = nil
	l.count = 0
}
