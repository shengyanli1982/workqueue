package list

import (
	"sync"
	"unsafe"
)

type Node struct {
	Value      interface{}
	Next, Prev *Node
	Index      int64
	parentRef  unsafe.Pointer
}

func (n *Node) Reset() {
	n.parentRef = nil
	n.Value = nil
	n.Next = nil
	n.Prev = nil
	n.Index = 0
}

func NewNode() *Node { return &Node{} }

type NodePool struct {
	pool sync.Pool
}

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