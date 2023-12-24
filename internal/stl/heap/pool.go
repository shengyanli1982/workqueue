package heap

import "sync"

type HeapElementPool struct {
	bp sync.Pool // 同步池 (sync pool)
}

func NewHeapElementPool() *HeapElementPool {
	return &HeapElementPool{
		bp: sync.Pool{
			New: func() any {
				return NewElement(nil, -1)
			},
		},
	}
}

func (p *HeapElementPool) Get() *Element {
	return p.bp.Get().(*Element)
}

func (p *HeapElementPool) Put(b *Element) {
	if b != nil {
		b.Reset()
		p.bp.Put(b)
	}
}
