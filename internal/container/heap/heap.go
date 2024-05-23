package heap

import lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"

type Heap struct {
	elements *lst.List
}

func New() *Heap {
	return &Heap{elements: lst.New()}
}

func (h *Heap) less(i, j *lst.Node) bool { return i.Index < j.Index }

func (h *Heap) up(i *lst.Node) {
	for i != nil {
		parent := i.Prev
		if parent == nil || h.less(parent, i) {
			break
		}
		h.elements.Swap(parent, i)
		i = parent
	}
}

func (h *Heap) down(i *lst.Node) {
	for i != nil {
		left := i.Next
		if left == nil {
			break
		}
		right := left.Next
		if right != nil && h.less(right, left) {
			left = right
		}
		if h.less(i, left) {
			break
		}
		h.elements.Swap(i, left)
		i = left
	}
}

func (h *Heap) Len() int64 { return h.elements.Len() }

func (h *Heap) Front() *lst.Node { return h.elements.Front() }

func (h *Heap) Back() *lst.Node { return h.elements.Back() }

func (h *Heap) Cleanup() { h.elements.Cleanup() }

func (h *Heap) Range(f func(*lst.Node) bool) { h.elements.Range(f) }

func (h *Heap) Slice() []interface{} { return h.elements.Slice() }

func (h *Heap) Remove(n *lst.Node) {
	if n == nil {
		return
	}
	h.elements.Remove(n)
	h.down(n)
}

func (h *Heap) Push(n *lst.Node) {
	if n == nil {
		return
	}
	h.elements.PushFront(n)
	h.up(n)
}

func (h *Heap) Pop() *lst.Node {
	if h.Len() <= 0 {
		return nil
	}
	n := h.elements.PopFront()
	h.down(h.Front())
	return n
}
