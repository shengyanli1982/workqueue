package heap

import (
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

type Heap struct {
	list    *lst.List
	mapping map[int64]*lst.Node
}

func New() *Heap {
	return &Heap{
		list:    lst.New(),
		mapping: make(map[int64]*lst.Node),
	}
}

func (h *Heap) less(i, j *lst.Node) bool { return i.Priority < j.Priority }

func (h *Heap) swap(i, j *lst.Node) {
	h.list.Swap(i, j)
	h.mapping[i.Index], h.mapping[j.Index] = j, i
	i.Index, j.Index = j.Index, i.Index
}

func (h *Heap) moveUp(n *lst.Node) {
	var parent *lst.Node
	for n != nil && n.Index > 0 {
		parentIndex := (n.Index - 1) / 2

		parent = h.mapping[parentIndex]

		if parent == nil || h.less(parent, n) {
			break
		}
		h.swap(parent, n)
		n = parent
	}
}

func (h *Heap) moveDown(n *lst.Node) {
	var left, right, smallest *lst.Node
	for n != nil {
		smallest = n

		leftIndex := n.Index*2 + 1
		rightIndex := n.Index*2 + 2

		left = h.mapping[leftIndex]
		right = h.mapping[rightIndex]

		if left != nil && h.less(left, smallest) {
			smallest = left
		}
		if right != nil && h.less(right, smallest) {
			smallest = right
		}
		if smallest == n {
			break
		}

		h.swap(n, smallest)
		n = smallest
	}
}

func (h *Heap) Len() int64 { return h.list.Len() }

func (h *Heap) Front() *lst.Node { return h.list.Front() }

func (h *Heap) Back() *lst.Node { return h.list.Back() }

func (h *Heap) Range(f func(*lst.Node) bool) { h.list.Range(f) }

func (h *Heap) Slice() []interface{} { return h.list.Slice() }

func (h *Heap) Cleanup() {
	h.list.Cleanup()
	h.mapping = make(map[int64]*lst.Node)
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
	n.Index = h.list.Len()
	h.mapping[n.Index] = n
	h.list.PushBack(n)
	h.moveUp(n)
}

func (h *Heap) Pop() *lst.Node {
	if h.list.Len() <= 0 {
		return nil
	}

	n := h.list.PopFront()
	delete(h.mapping, n.Index)
	n.Index = 0
	h.moveDown(h.Front())
	return n
}
