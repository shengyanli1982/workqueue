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
	RED = 0

	// BLACK 表示黑色
	// BLACK represents black
	BLACK = 1
)

// Node 结构体代表一个节点，它可以存储任何类型的值，有一个优先级，以及指向前一个和后一个节点的指针。
// Node struct represents a node that can hold a value of any type, has a priority, and pointers to the next and previous nodes.
type Node struct {
	// parentRef 是一个指向父节点的指针，它的类型是 unsafe.Pointer，所以可以指向任何类型的值。
	// parentRef is a pointer to the parent node. Its type is unsafe.Pointer, so it can point to a value of any type.
	parentRef unsafe.Pointer

	// Priority 是节点的优先级
	// Priority is the Priority of the node
	Priority int64

	// Color 是节点的颜色，可以是 RED 或 BLACK
	// Color is the Color of the node, can be RED or BLACK
	Color int64

	// Left 是节点的左子节点
	// Left is the Left child of the node
	Left *Node

	// Right 是节点的右子节点
	// Right is the Right child of the node
	Right *Node

	// Parent 是节点的父节点
	// Parent is the Parent of the node
	Parent *Node

	// Value 是节点存储的值
	// Value is the value stored in the node
	Value interface{}
}

// Reset 方法重置节点的所有字段，将它们设置为零值。
// The Reset method resets all the fields of the node, setting them to their zero values.
func (n *Node) Reset() {
	// 将 parentRef 设置为 nil，表示没有父节点。
	// Set parentRef to nil, indicating no parent node.
	n.parentRef = nil

	// 将 priority 设置为 0，表示节点的优先级为 0。
	// Set priority to 0, indicating the priority of the node is 0.
	n.Priority = 0

	// 将 color 设置为 0，表示节点的颜色为 RED。
	// Set color to 0, indicating the color of the node is RED.
	n.Color = RED

	// 将 left 设置为 nil，表示节点没有左子节点。
	// Set left to nil, indicating the node has no left child.
	n.Left = nil

	// 将 right 设置为 nil，表示节点没有右子节点。
	// Set right to nil, indicating the node has no right child.
	n.Right = nil

	// 将 parent 设置为 nil，表示节点没有父节点。
	// Set parent to nil, indicating the node has no parent node.
	n.Parent = nil

	// 将 Value 设置为 nil，表示节点不存储任何值。
	// Set Value to nil, indicating the node does not hold any value.
	n.Value = nil
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
