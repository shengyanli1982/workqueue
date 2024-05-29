package list

import (
	"sync"
	"unsafe"
)

// Node 结构体代表一个节点，它可以存储任何类型的值，有一个优先级，以及指向前一个和后一个节点的指针。
// Node struct represents a node that can hold a value of any type, has a priority, and pointers to the next and previous nodes.
type Node struct {
	// Value 是存储在节点中的值，它的类型是 interface{}，所以可以是任何类型。
	// Value is the value stored in the node. Its type is interface{}, so it can be of any type.
	Value interface{}

	// parentRef 是一个指向父节点的指针，它的类型是 unsafe.Pointer，所以可以指向任何类型的值。
	// parentRef is a pointer to the parent node. Its type is unsafe.Pointer, so it can point to a value of any type.
	parentRef unsafe.Pointer

	// Priority 是节点的优先级，类型为 int64。
	// Priority is the priority of the node. Its type is int64.
	Priority int64

	// Next 是指向下一个节点的指针。
	// Next is a pointer to the next node.
	// Prev 是指向前一个节点的指针。
	// Prev is a pointer to the previous node.
	Next, Prev *Node
}

// Reset 方法重置节点的所有字段，将它们设置为零值。
// The Reset method resets all the fields of the node, setting them to their zero values.
func (n *Node) Reset() {
	// 将 parentRef 设置为 nil，表示没有父节点。
	// Set parentRef to nil, indicating no parent node.
	n.parentRef = nil

	// 将 Value 设置为 nil，表示节点不存储任何值。
	// Set Value to nil, indicating the node does not hold any value.
	n.Value = nil

	// 将 Next 和 Prev 设置为 nil，表示节点没有前一个和后一个节点。
	// Set Next and Prev to nil, indicating the node does not have a previous and a next node.
	n.Next = nil
	n.Prev = nil

	// 将 Priority 设置为 0，表示节点没有优先级。
	// Set Priority to 0, indicating the node has no priority.
	n.Priority = 0
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
