package heap

import lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"

const MINIHEAPSIZE = 4

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

func (h *Heap) less(i, j *lst.Node) bool { return i.Index < j.Index }

func (h *Heap) getChildIndex(node *lst.Node, childCount int) int64 {
	var minIndex int64 = -1
	var minNode *lst.Node
	currentNode := node.Next
	for i := 0; i < childCount && currentNode != nil; i++ {
		if minNode == nil || h.less(currentNode, minNode) {
			minNode = currentNode
			minIndex = currentNode.Index
		}
		currentNode = currentNode.Next
	}
	return minIndex
}

func (h *Heap) moveUp(node *lst.Node) {
	for node.Prev != nil && h.less(node, node.Prev) {
		h.list.Swap(node, node.Prev)
	}
}

func (h *Heap) moveDown(node *lst.Node) {
	for {
		childIndex := h.getChildIndex(node, MINIHEAPSIZE)
		if childIndex == -1 {
			break
		}
		if childNode, ok := h.mapping[childIndex]; ok && !h.less(node, childNode) {
			h.list.Swap(node, childNode)
			node = childNode
		} else {
			break
		}
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
	delete(h.mapping, node.Index)
}

func (h *Heap) Push(node *lst.Node) {
	if node == nil {
		return
	}
	h.list.PushBack(node)
	h.mapping[node.Index] = node
	h.moveUp(node)
}

func (h *Heap) Pop() *lst.Node {
	if h.list.Len() <= 0 {
		return nil
	}

	minNode := h.list.PopFront()
	if minNode == nil {
		return nil
	}
	delete(h.mapping, minNode.Index)
	if h.list.Len() > 0 {
		h.moveDown(h.Front())
	}
	return minNode
}
