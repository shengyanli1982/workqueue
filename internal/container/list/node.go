package list

import (
	"sync"
	"unsafe"
)

// 定义两个常量，表示节点的颜色
// Define two constants to represent the color of the node
const (
	// RED 表示红色
	// RED represents red
	RED = uint8(0)

	// BLACK 表示黑色
	// BLACK represents black
	BLACK = uint8(1)
)

// Node 结构体代表一个节点，它可以存储任何类型的值，有一个优先级，以及指向其他节点的指针。
// Node struct represents a node that can hold a value of any type, has a priority, and pointers to other nodes.
type Node struct {
	// Value 是节点存储的值
	// Value is the value stored in the node
	Value interface{}

	// Left 是节点的左子节点
	// Left is the Left child of the node
	Left *Node

	// Right 是节点的右子节点
	// Right is the Right child of the node
	Right *Node

	// Parent 是节点的父节点
	// Parent is the parent node
	Parent *Node

	// parentRef 是一个指向父节点的指针
	// parentRef is a pointer to the parent node
	parentRef unsafe.Pointer

	// Priority 是节点的优先级
	// Priority is the Priority of the node
	Priority int64

	// Color 是节点的颜色，可以是 RED 或 BLACK
	// Color is the Color of the node, can be RED or BLACK
	Color uint8

	// 用于内存对齐的填充字段
	// Padding field for memory alignment
	_ [7]uint8
}

// Reset 方法重置节点的所有字段，将它们设置为零值。
// The Reset method resets all the fields of the node, setting them to their zero values.
func (n *Node) Reset() {
	n.Value = nil
	n.Left = nil
	n.Right = nil
	n.Parent = nil
	n.parentRef = nil
	n.Priority = 0
	n.Color = RED
}

// NewNode 函数创建并返回一个新的 Node 实例。
// The NewNode function creates and returns a new instance of Node.
func NewNode() *Node { return &Node{} }

// NodePool 结构体是一个节点池，它使用 sync.Pool 来存储和复用 Node 实例。
// The NodePool struct is a pool of nodes, it uses sync.Pool to store and reuse Node instances.
type NodePool struct {
	// pool 是一个 sync.Pool 实例，它用于存储和复用 Node 实例。
	// pool is an instance of sync.Pool, it is used to store and reuse Node instances.
	pool sync.Pool
}

// NewNodePool 函数创建并返回一个新的 NodePool 实例。
// The NewNodePool function creates and returns a new instance of NodePool.
func NewNodePool() *NodePool {
	return &NodePool{
		pool: sync.Pool{
			// New 是一个函数，当 sync.Pool 需要一个新的实例时，它会调用这个函数。
			// New is a function that sync.Pool will call when it needs a new instance.
			New: func() interface{} {
				// 这里，我们返回一个新的 Node 实例。
				// Here, we return a new instance of Node.
				return NewNode()
			},
		},
	}
}

// Get 方法从 NodePool 中获取一个 Node 实例。
// The Get method retrieves a Node instance from the NodePool.
func (p *NodePool) Get() *Node {
	// 我们从 sync.Pool 中获取一个实例，并将其转换为 *Node 类型。
	// We retrieve an instance from sync.Pool and cast it to *Node.
	return p.pool.Get().(*Node)
}

// Put 方法将一个 Node 实例放回到 NodePool 中。
// The Put method puts a Node instance back into the NodePool.
func (p *NodePool) Put(n *Node) {
	// 我们首先重置 Node 实例，清除其所有字段。
	// We first reset the Node instance, clearing all its fields.
	n.Reset()

	// 然后，我们将 Node 实例放回到 sync.Pool 中。
	// Then, we put the Node instance back into the sync.Pool.
	p.pool.Put(n)
}
