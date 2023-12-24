package deque

import "sync"

type ListNodePool struct {
	bp sync.Pool // 同步池 (sync pool)
}

func NewListNodePool() *ListNodePool {
	return &ListNodePool{
		bp: sync.Pool{
			New: func() any {
				return NewNode(nil)
			},
		},
	}
}

func (p *ListNodePool) Get() *Node {
	return p.bp.Get().(*Node) // 从池中获取 ListNode 对象 (get ListNode object from the pool)
}

func (p *ListNodePool) Put(b *Node) {
	if b != nil {
		b.Reset()
		p.bp.Put(b) // 将 ListNode 对象放回池中 (put ListNode object back into the pool)
	}
}
