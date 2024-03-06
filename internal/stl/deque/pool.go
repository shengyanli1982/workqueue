package deque

import (
	"sync"
)

// ListNodePool 结构体，用于管理 Node 对象的同步池
// ListNodePool struct, used to manage the sync pool of Node objects
type ListNodePool struct {
	bp sync.Pool // 同步池 (sync pool)
}

// NewListNodePool 函数，用于创建一个新的 ListNodePool
// NewListNodePool function, used to create a new ListNodePool
func NewListNodePool() *ListNodePool {
	return &ListNodePool{
		bp: sync.Pool{
			// 当池中没有可用对象时，会调用此函数创建一个新的 Node 对象
			// When there are no available objects in the pool, this function will be called to create a new Node object
			New: func() any {
				return NewNode(nil)
			},
		},
	}
}

// Get 方法，用于从池中获取一个 Node 对象
// Get method, used to get a Node object from the pool
func (p *ListNodePool) Get() *Node {
	// 从池中获取一个 Node 对象，并将其类型断言为 *Node
	// Get a Node object from the pool and type assert it to *Node
	return p.bp.Get().(*Node)
}

// Put 方法，用于将一个 Node 对象放回池中
// Put method, used to put a Node object back into the pool
func (p *ListNodePool) Put(n *Node) {
	// 如果 n 不为 nil，重置 n 并将其放回池中
	// If n is not nil, reset n and put it back into the pool
	if n != nil {
		n.Reset()
		p.bp.Put(n)
	}
}
