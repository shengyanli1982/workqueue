package heap

import (
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

type Heap struct {
	list *lst.List
	id   uint64
}

func New() *Heap {
	return &Heap{
		list: lst.New(),
	}
}

func (h *Heap) less(i, j *lst.Node) bool {
	return i.Priority < j.Priority
}

func (h *Heap) moveUp(node *lst.Node) {
	if node == nil || node.Prev == nil {
		return
	}

	current := node
	for current.Prev != nil && h.less(node, current.Prev) {
		current = current.Prev
	}

	if current != node {
		h.list.PopBack()
		h.list.InsertBefore(node, current)
	}
}

func (h *Heap) Len() int64 { return h.list.Len() }

func (h *Heap) Front() *lst.Node { return h.list.Front() }

func (h *Heap) Back() *lst.Node { return h.list.Back() }

func (h *Heap) Range(f func(*lst.Node) bool) { h.list.Range(f) }

func (h *Heap) Slice() []interface{} { return h.list.Slice() }

func (h *Heap) Cleanup() {
	h.list.Cleanup()
}

func (h *Heap) Remove(node *lst.Node) {
	if node == nil {
		return
	}
	h.list.Remove(node)
}

func (h *Heap) Push(n *lst.Node) {
	if n == nil {
		return
	}

	h.list.PushBack(n)
	h.moveUp(n)
}

func (h *Heap) Pop() *lst.Node {
	if h.Len() <= 0 {
		return nil
	}

	return h.list.PopFront()
}
