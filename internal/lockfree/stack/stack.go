package stack

import (
	"sync/atomic"
	"unsafe"
)

type Node struct {
	value int64
	next  *Node
}

func LoadNode(p *unsafe.Pointer) *Node {
	return (*Node)(atomic.LoadPointer(p))
}

func CompareAndSwapNode(p *unsafe.Pointer, old, new *Node) bool {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}

type Stack struct {
	count int64
	top   unsafe.Pointer
}

func New() *Stack {
	return &Stack{}
}

func (s *Stack) Len() int64 {
	return atomic.LoadInt64(&s.count)
}

func (s *Stack) Empty() bool {
	return s.Len() == 0
}

func (s *Stack) Cleanup() {
	atomic.StorePointer(&s.top, nil)
	atomic.StoreInt64(&s.count, 0)
}

func (s *Stack) Push(i int64) {
	n := &Node{value: i}
	for {
		top := LoadNode(&s.top)
		n.next = top
		if CompareAndSwapNode(&s.top, top, n) {
			atomic.AddInt64(&s.count, 1)
			return
		}
	}
}

func (s *Stack) Pop() int64 {
	for {
		top := LoadNode(&s.top)
		if top == nil {
			return -1
		}
		next := top.next
		if CompareAndSwapNode(&s.top, top, next) {
			atomic.AddInt64(&s.count, -1)
			return top.value
		}
	}
}
