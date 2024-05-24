package heap

import (
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

const (
	INIT_NODE_SIZE = 8
	// UPWORD_LEVEL   = INIT_NODE_SIZE << 5
	// DOWNWORD_LEVEL = INIT_NODE_SIZE >> 5
)

type Heap struct {
	list    *lst.List
	mapping []*lst.Node
}

func New() *Heap {
	return &Heap{
		list:    lst.New(),
		mapping: make([]*lst.Node, 0, INIT_NODE_SIZE),
	}
}

func (h *Heap) less(i, j *lst.Node) bool { return i.Priority < j.Priority }

func (h *Heap) swap(i, j *lst.Node) {
	h.mapping[i.Index], h.mapping[j.Index] = j, i
	h.mapping[i.Index].Index, h.mapping[j.Index].Index = i.Index, j.Index
	h.list.Swap(i, j)
}

func (h *Heap) moveUp(node *lst.Node) {
	for node != nil && node.Index > 0 {
		parentIndex := (node.Index - 1) / 2

		parent := h.mapping[parentIndex]

		if parent == nil || !h.less(node, parent) {
			break
		}

		h.swap(node, parent)
	}
}

func (h *Heap) moveDown(node *lst.Node) {
	count := h.list.Len() - 1

	for node != nil {
		child1 := node.Index*2 + 1
		if child1 >= count {
			break
		}

		child2 := child1 + 1

		j := child1
		if child2 < count && h.less(h.mapping[child2], h.mapping[child1]) {
			j = child2
		}

		if !h.less(h.mapping[j], node) {
			break
		}

		h.swap(node, h.mapping[j])

		node = h.mapping[j]
	}
}

func (h *Heap) Len() int64 { return h.list.Len() }

func (h *Heap) Front() *lst.Node { return h.list.Front() }

func (h *Heap) Back() *lst.Node { return h.list.Back() }

func (h *Heap) Range(f func(*lst.Node) bool) { h.list.Range(f) }

func (h *Heap) Slice() []interface{} { return h.list.Slice() }

func (h *Heap) Cleanup() {
	h.list.Cleanup()
	h.mapping = h.mapping[:0]
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
	h.mapping = append(h.mapping, n)
	h.list.PushBack(n)
	h.moveUp(n)
}

func (h *Heap) Pop() *lst.Node {
	if h.Len() <= 0 {
		return nil
	}

	n := h.list.PopFront()
	h.mapping[n.Index] = nil
	h.moveDown(h.Front())
	return n
}
