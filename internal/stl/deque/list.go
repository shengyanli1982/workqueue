package deque

// 双向链表的节点
// Node of the doubly linked list
type Node struct {
	data       any
	prev, next *Node
}

// 创建一个新的节点
// Create a new node
func NewNode(data any) *Node {
	return &Node{data: data}
}

// 重置节点
// Reset the node
func (n *Node) Reset() {
	n.prev = nil
	n.next = nil
	n.data = nil
}

// 获取节点的数据
// Get the data of the node
func (n *Node) Data() any {
	return n.data
}

// 设置节点的数据
// Set the data of the node
func (n *Node) SetData(data any) {
	n.data = data
}

// 双向链表
// Doubly linked list
type Deque struct {
	head   *Node
	tail   *Node
	length int
}

// 创建一个新的双向链表
// Create a new doubly linked list
func NewDeque() *Deque {
	return &Deque{}
}

// 重置链表
// Reset the list
func (l *Deque) Reset() {
	// 遍历链表，将每个节点弹出
	// Traverse the list and pop each node
	for n := l.head; n != nil; {
		n = n.next
		l.Pop()
	}

	// 将头节点、尾节点设为 nil，长度设为 0
	// Set the head node and tail node to nil, and the length to 0
	l.head = nil
	l.tail = nil
	l.length = 0
}

// 将节点 n 添加在链表尾部
// Add node n to the end of the list
func (l *Deque) Push(n *Node) {
	// 长度加 1
	// Increase the length by 1
	l.length++

	// 如果链表为空，将头节点和尾节点都设为 n
	// If the list is empty, set both the head node and the tail node to n
	if l.head == nil {
		l.head = n
		l.tail = n
		return
	}

	// 否则，将 n 添加到尾节点后面，然后更新尾节点
	// Otherwise, add n after the tail node, and then update the tail node
	l.tail.next = n
	n.prev = l.tail
	l.tail = n
}

// 将节点 n 从头部弹出
// Pop removes a node from the head of the list
func (l *Deque) Pop() *Node {
	// 如果链表为空，返回 nil
	// If the list is empty, return nil
	if l.head == nil {
		return nil
	}

	// 否则，将头节点弹出，然后更新头节点
	// Otherwise, pop the head node, and then update the head node
	n := l.head
	l.head = n.next
	if l.head != nil {
		l.head.prev = nil
	} else {
		l.tail = nil
	}
	n.next = nil

	// 长度减 1
	// Decrease the length by 1
	l.length--
	return n
}

// 将节点 n 添加在链表头部
// PushFront adds a node to the head of the list
func (l *Deque) PushFront(n *Node) {
	// 长度加 1
	// Increase the length by 1
	l.length++

	// 如果链表为空，将头节点和尾节点都设为 n
	// If the list is empty, set both the head node and the tail node to n
	if l.tail == nil {
		l.head = n
		l.tail = n
		return
	}

	// 否则，将 n 添加到头节点前面，然后更新头节点
	// Otherwise, add n before the head node, and then update the head node
	l.head.prev = n
	n.next = l.head
	l.head = n
}

// 将节点 n 从尾部弹出
// PopBack removes a node from the tail of the list
func (l *Deque) PopBack() *Node {
	// 如果链表为空，返回 nil
	// If the list is empty, return nil
	if l.tail == nil {
		return nil
	}

	// 否则，将尾节点弹出，然后更新尾节点
	// Otherwise, pop the tail node, and then update the tail node
	n := l.tail
	l.tail = n.prev
	if l.tail != nil {
		l.tail.next = nil
	} else {
		l.head = nil
	}
	n.prev = nil

	// 长度减 1
	// Decrease the length by 1
	l.length--
	return n
}

// 删除节点 n
// Delete removes a node from the list
func (l *Deque) Delete(n *Node) {
	// 如果 n 是头节点，更新头节点
	// If n is the head node, update the head node
	if n.prev == nil {
		l.head = n.next
	} else {
		// 否则，将 n 的前一个节点的 next 指向 n 的下一个节点
		// Otherwise, point the next of the previous node of n to the next node of n
		n.prev.next = n.next
	}

	// 如果 n 是尾节点，更新尾节点
	// If n is the tail node, update the tail node
	if n.next == nil {
		l.tail = n.prev
	} else {
		// 否则，将 n 的下一个节点的 prev 指向 n 的前一个节点
		// Otherwise, point the prev of the next node of n to the previous node of n
		n.next.prev = n.prev
	}

	// 将 n 的 prev 和 next 都设为 nil
	// Set both the prev and next of n to nil
	n.prev = nil
	n.next = nil

	// 长度减 1
	// Decrease the length by 1
	l.length--
}

// 链表的长度
// Len returns the length of the list
func (l *Deque) Len() int {
	// 返回链表的长度
	// Return the length of the list
	return l.length
}

// 链表的头部
// Head returns the head of the list
func (l *Deque) Head() *Node {
	// 返回链表的头节点
	// Return the head node of the list
	return l.head
}

// 链表的尾部
// Tail returns the tail of the list
func (l *Deque) Tail() *Node {
	// 返回链表的尾节点
	// Return the tail node of the list
	return l.tail
}

// 链表是否为空
// IsEmpty returns true if the list is empty
func (l *Deque) IsEmpty() bool {
	// 如果链表的长度为 0，返回 true，否则返回 false
	// If the length of the list is 0, return true, otherwise return false
	return l.length == 0
}

// 获得链表所有元素值数组
// SnapshotValues returns current snapshot list of all values
func (l *Deque) SnapshotValues() []any {
	// 创建一个空的数组，长度为链表的长度
	// Create an empty array with a length of the list length
	values := make([]any, 0, l.length)

	// 遍历链表，将每个节点的数据添加到数组中
	// Traverse the list and add the data of each node to the array
	for n := l.head; n != nil; n = n.next {
		values = append(values, n.data)
	}

	// 返回数组
	// Return the array
	return values
}

// 遍历链表
// Range iterates the list
func (l *Deque) Range(fn func(node *Node) bool) {
	// 遍历链表，对每个节点执行函数 fn
	// If fn returns false, stop the iteration
	for n := l.head; n != nil; n = n.next {
		// 如果 fn 返回 false，停止遍历
		// Traverse the list and execute the function fn on each node
		if !fn(n) {
			break
		}
	}
}
