package list

import (
	"sync"
	"unsafe"
)

// 红黑树颜色标记。
const (
	RED = uint8(0)

	BLACK = uint8(1)
)

// Node 是链表与红黑树共用的节点结构。
// Left/Right 在链表中表示前后节点，在红黑树中表示左右子节点；
// Parent 用于红黑树父节点回溯；parentRef 存储所属链表地址（O(1) 归属判断）；
// Priority 为堆排序优先级键；Color 为红黑树着色（RED/BLACK）。
type Node struct {
	Value any

	Left *Node

	Right *Node

	Parent *Node

	parentRef unsafe.Pointer

	Priority int64

	Color uint8

	_ [7]uint8
}

// Reset 将节点重置为初始状态，清除所有指针和优先级，着色恢复为 RED。
func (n *Node) Reset() {
	n.Value = nil
	n.Left = nil
	n.Right = nil
	n.Parent = nil
	n.parentRef = nil
	n.Priority = 0
	n.Color = RED
}

// NewNode 创建零值节点。
func NewNode() *Node { return &Node{} }

// NodePool 基于 sync.Pool 复用 Node，降低高频入队/出队场景的 GC 压力。
type NodePool struct {
	pool sync.Pool
}

// NewNodePool 创建节点池。
func NewNodePool() *NodePool {
	return &NodePool{
		pool: sync.Pool{

			New: func() any {

				return NewNode()
			},
		},
	}
}

// Get 从节点池获取一个节点。
func (p *NodePool) Get() *Node {

	return p.pool.Get().(*Node)
}

// Put 重置节点后归还节点池。
func (p *NodePool) Put(n *Node) {

	n.Reset()

	p.pool.Put(n)
}
