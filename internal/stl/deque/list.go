package deque

// 用双向链表实现一个队列
// 双向链表的节点
type Node struct {
	data       any
	prev, next *Node
}

func NewNode(data any) *Node {
	return &Node{data: data}
}

// 重置节点
// Reset resets the node
func (n *Node) Reset() {
	n.prev = nil
	n.next = nil
	n.data = nil
}

// 获取节点的数据
// Data returns the data of the node
func (n *Node) Data() any {
	return n.data
}

// 设置节点的数据
// SetData sets the data of the node
func (n *Node) SetData(data any) {
	n.data = data
}

// 双向链表
// Deque is a doubly linked list
type Deque struct {
	head   *Node
	tail   *Node
	length int
}

func NewDeque() *Deque {
	return &Deque{}
}

// 重置链表
// Reset resets the list
func (l *Deque) Reset() {
	for n := l.head; n != nil; {
		n = n.next
		l.Pop()
	}
	l.head = nil
	l.tail = nil
	l.length = 0
}

// 将节点 n 添加在链表尾部
// Push adds a node to the tail of the list
func (l *Deque) Push(n *Node) {
	l.length++
	if l.head == nil {
		l.head = n
		l.tail = n
		return
	}
	l.tail.next = n
	n.prev = l.tail
	l.tail = n
}

// 将节点 n 从头部弹出
// Pop removes a node from the head of the list
func (l *Deque) Pop() *Node {
	if l.head == nil {
		return nil
	}
	n := l.head
	l.head = n.next
	if l.head != nil {
		l.head.prev = nil
	} else {
		l.tail = nil
	}
	n.next = nil
	l.length--
	return n
}

// 将节点 n 添加在链表头部
// PushFront adds a node to the head of the list
func (l *Deque) PushFront(n *Node) {
	l.length++
	if l.tail == nil {
		l.head = n
		l.tail = n
		return
	}
	l.head.prev = n
	n.next = l.head
	l.head = n
}

// 将节点 n 从尾部弹出
// PopBack removes a node from the tail of the list
func (l *Deque) PopBack() *Node {
	if l.tail == nil {
		return nil
	}
	n := l.tail
	l.tail = n.prev
	if l.tail != nil {
		l.tail.next = nil
	} else {
		l.head = nil
	}
	n.prev = nil
	l.length--
	return n
}

// 删除节点 n
// Delete removes a node from the list
func (l *Deque) Delete(n *Node) {
	if n.prev == nil {
		l.head = n.next
	} else {
		n.prev.next = n.next
	}
	if n.next == nil {
		l.tail = n.prev
	} else {
		n.next.prev = n.prev
	}
	n.prev = nil
	n.next = nil
	l.length--
}

// 链表的长度
// Len returns the length of the list
func (l *Deque) Len() int {
	return l.length
}

// 链表的头部
// Head returns the head of the list
func (l *Deque) Head() *Node {
	return l.head
}

// 链表的尾部
// Tail returns the tail of the list
func (l *Deque) Tail() *Node {
	return l.tail
}
