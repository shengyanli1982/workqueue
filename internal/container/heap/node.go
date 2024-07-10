package heap

import "sync"

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

// Node 是一个结构体，表示一个节点
// Node is a struct that represents a node
type Node struct {
	// priority 是节点的优先级
	// priority is the priority of the node
	priority int64

	// color 是节点的颜色，可以是 RED 或 BLACK
	// color is the color of the node, can be RED or BLACK
	color int64

	// left 是节点的左子节点
	// left is the left child of the node
	left *Node

	// right 是节点的右子节点
	// right is the right child of the node
	right *Node

	// parent 是节点的父节点
	// parent is the parent of the node
	parent *Node

	// Value 是节点存储的值
	// Value is the value stored in the node
	Value interface{}
}

// Reset 方法用于重置节点的所有属性
// The Reset method is used to reset all properties of the node
func (n *Node) Reset() {
	// 将优先级设置为 0
	// Set the priority to 0
	n.priority = 0

	// 将颜色设置为 0
	// Set the color to 0
	n.color = 0

	// 将左子节点设置为 nil
	// Set the left child to nil
	n.left = nil

	// 将右子节点设置为 nil
	// Set the right child to nil
	n.right = nil

	// 将父节点设置为 nil
	// Set the parent to nil
	n.parent = nil

	// 将存储的值设置为 nil
	// Set the stored value to nil
	n.Value = nil
}

// NewNode 函数用于创建一个新的节点
// The NewNode function is used to create a new node
func NewNode() *Node {
	// 返回一个新的 Node 结构体实例
	// Return a new instance of the Node struct
	return &Node{}
}

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
