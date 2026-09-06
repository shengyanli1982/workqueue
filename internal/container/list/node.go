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
// Value 存储用户数据；Left/Right 在链表中表示前后节点，在红黑树中表示左右子节点；
// Parent 用于红黑树的父节点回溯；parentRef 存储所属链表地址，用于 O(1) 归属判断；
// Priority 用于堆排序的优先级键；Color 用于红黑树着色（RED/BLACK）；
// _ 为对齐填充，确保结构体在 64 位系统上对齐到 8 字节边界。
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

// Reset 将节点重置为初始状态，清除所有指针引用和优先级，着色恢复为 RED。
// 用于节点归还节点池前的清理，确保复用时不携带残留数据。
func (n *Node) Reset() {
	n.Value = nil
	n.Left = nil
	n.Right = nil
	n.Parent = nil
	n.parentRef = nil
	n.Priority = 0
	n.Color = RED
}

// NewNode 创建一个新的零值节点。
func NewNode() *Node { return &Node{} }

// NodePool 是基于 sync.Pool 的节点对象池，通过复用已分配的 Node 降低高频入队/出队场景下的 GC 压力。
type NodePool struct {
	pool sync.Pool
}

// NewNodePool 创建可复用的节点池，降低高频入队分配成本。
func NewNodePool() *NodePool {
	return &NodePool{
		pool: sync.Pool{

			New: func() any {

				return NewNode()
			},
		},
	}
}

// Get 从节点池中获取一个节点。池为空时由 sync.Pool.New 分配新节点。
func (p *NodePool) Get() *Node {

	return p.pool.Get().(*Node)
}

// Put 将节点重置后归还节点池。先调用 Reset 清除残留数据，再放回 sync.Pool 供后续 Get 复用。
func (p *NodePool) Put(n *Node) {

	n.Reset()

	p.pool.Put(n)
}
