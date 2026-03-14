package list

import (
	"sync"
	"unsafe"
)

// 红黑树颜色标记。
const (
	RED   = uint8(0)
	BLACK = uint8(1)
)

// Node 既用于链表，也用于堆中的红黑树节点。
type Node struct {
	Value     interface{}
	Left      *Node
	Right     *Node
	Parent    *Node
	parentRef unsafe.Pointer
	Priority  int64
	Color     uint8
	_         [7]uint8
}

func (n *Node) Reset() {
	n.Value = nil
	n.Left = nil
	n.Right = nil
	n.Parent = nil
	n.parentRef = nil
	n.Priority = 0
	n.Color = RED
}

func NewNode() *Node { return &Node{} }

type NodePool struct {
	pool sync.Pool
}

// NewNodePool 创建可复用的节点池，降低高频入队分配成本。
func NewNodePool() *NodePool {
	return &NodePool{
		pool: sync.Pool{
			New: func() interface{} {
				return NewNode()
			},
		},
	}
}

func (p *NodePool) Get() *Node {
	return p.pool.Get().(*Node)
}

func (p *NodePool) Put(n *Node) {
	n.Reset()
	p.pool.Put(n)
}
