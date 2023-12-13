package workqueue

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
func (n *Node) Reset() {
	n.prev = nil
	n.next = nil
	n.data = nil
}

// 获取节点的数据
func (n *Node) Data() any {
	return n.data
}

// 双向链表
type Deque struct {
	head   *Node
	tail   *Node
	length int
}

func NewDeque() *Deque {
	return &Deque{}
}

// 重置链表
func (l *Deque) Reset() {
	l.head = nil
	l.tail = nil
	l.length = 0
}

// 将节点 n 添加在链表尾部
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
func (l *Deque) Len() int {
	return l.length
}

// 链表的头部
func (l *Deque) Head() *Node {
	return l.head
}

// 链表的尾部
func (l *Deque) Tail() *Node {
	return l.tail
}
