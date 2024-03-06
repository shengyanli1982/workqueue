package heap

import (
	"sync"
)

// HeapElementPool 结构体，用于管理 Element 对象的同步池
// HeapElementPool struct, used to manage the sync pool of Element objects
type HeapElementPool struct {
	bp sync.Pool // 同步池 (sync pool)
}

// NewHeapElementPool 函数，用于创建一个新的 HeapElementPool
// NewHeapElementPool function, used to create a new HeapElementPool
func NewHeapElementPool() *HeapElementPool {
	return &HeapElementPool{
		bp: sync.Pool{
			// 当池中没有可用对象时，会调用此函数创建一个新的 Element 对象
			// When there are no available objects in the pool, this function will be called to create a new Element object
			New: func() any {
				return NewElement(nil, -1)
			},
		},
	}
}

// Get 方法，用于从池中获取一个 Element 对象
// Get method, used to get an Element object from the pool
func (p *HeapElementPool) Get() *Element {
	// 从池中获取一个 Element 对象，并将其类型断言为 *Element
	// Get an Element object from the pool and type assert it to *Element
	return p.bp.Get().(*Element)
}

// Put 方法，用于将一个 Element 对象放回池中
// Put method, used to put an Element object back into the pool
func (p *HeapElementPool) Put(e *Element) {
	// 如果 e 不为 nil，重置 e 并将其放回池中
	// If e is not nil, reset e and put it back into the pool
	if e != nil {
		e.Reset()
		p.bp.Put(e)
	}
}
