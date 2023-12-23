package stl

type Stack struct {
	l *Deque
}

func NewStack() *Stack {
	return &Stack{l: NewDeque()}
}

// 重置堆
// Reset resets the heap
func (s *Stack) Reset() {
	s.l.Reset()
}

// Push 添加一个节点到堆中
// Push adds a node to the heap
func (s *Stack) Push(n *Node) {
	s.l.PushFront(n)
}

// Pop 移除并返回堆中的头部节点
// Pop removes a node from the head of the heap
func (s *Stack) Pop() *Node {
	return s.l.Pop()
}

// Len 返回堆的长度
// Len returns the length of the heap
func (s *Stack) Len() int {
	return s.l.Len()
}

// Head 返回堆的头部节点
// Head returns the head node of the heap
func (s *Stack) Head() *Node {
	return s.l.Head()
}

// Tail 返回堆的尾部节点
// Tail returns the tail node of the heap
func (s *Stack) Tail() *Node {
	return s.l.Tail()
}
