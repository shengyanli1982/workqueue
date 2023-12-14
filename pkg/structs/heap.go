package structs

// 堆中元素结构
// Element is the structure of the element in the heap
type Element struct {
	data  any   // 数据
	value int64 // 值
	index int   // 索引
}

// Data 返回元素的数据
// Data returns the data of the element
func (e *Element) Data() any {
	return e.data
}

// Value 返回元素的值
// Value returns the value of the element
func (e *Element) Value() int64 {
	return e.value
}

// Index 返回元素的索引
// Index returns the index of the element
func (e *Element) Index() int {
	return e.index
}

// ResetValue 重置元素的值
// ResetValue resets the value of the element
func (e *Element) ResetValue(i int64) {
	e.value = i
}

// Reset 重置元素
// Reset resets the element
func (e *Element) Reset() {
	e.data = nil
	e.value = 0
	e.index = 0
}

// NewElement 返回一个新的元素
// NewElement returns a new element
func NewElement(data any, value int64) *Element {
	return &Element{data: data, value: value}
}

// Heap 是一个最小4叉堆
// Heap is a minimum 4-ary heap
type Heap struct {
	data []*Element
}

func NewHeap() *Heap {
	return &Heap{}
}

// Reset 重置堆
// Reset resets the heap
func (h *Heap) Reset() {
	h.data = h.data[:0]
}

// Less 比较两个元素的值
// Less compares the values of two elements
func (h *Heap) Less(i, j int) bool { return h.data[i].value < h.data[j].value }

// Update 更新元素的值
// Update updates the value of the element
func (h *Heap) Update(ele *Element, value int64) {
	if value > ele.value {
		h.down(ele.index, h.Len())
	} else {
		h.up(ele.index)
	}
	ele.value = value
}

// min 返回两个元素中值较小的元素的索引
// min returns the index of the element with the smaller value of the two elements
func (h *Heap) min(i, j int) int {
	if h.data[i].value < h.data[j].value {
		return i
	}
	return j
}

// Len 返回堆中元素的个数
// Len returns the number of elements in the heap
func (h *Heap) Len() int {
	return len(h.data)
}

// Swap 交换两个元素
// Swap swaps two elements
func (h *Heap) Swap(i, j int) {
	h.data[i].index, h.data[j].index = h.data[j].index, h.data[i].index
	h.data[i], h.data[j] = h.data[j], h.data[i]
}

// Push 添加一个元素
// Push adds an element
func (h *Heap) Push(ele *Element) {
	ele.index = h.Len()
	h.data = append(h.data, ele)
	h.up(h.Len() - 1)
}

// up 从 i 开始向上调整堆
// Adjust the heap from i upwards
func (h *Heap) up(i int) {
	for i > 0 {
		parent := (i - 1) >> 2
		if !h.Less(i, parent) {
			break
		}
		h.Swap(i, parent)
		i = parent
	}
}

// down 从 i 开始向下调整堆
// Adjust the heap from i downwards
func (h *Heap) down(i, n int) {
	for {
		child1 := i<<2 + 1
		if child1 >= n {
			break
		}

		child2 := child1 + 1
		child3 := child1 + 2
		child4 := child1 + 3
		j := child1

		if child4 < n {
			j = h.min(h.min(child1, child2), h.min(child3, child4))
		} else if child3 < n {
			j = h.min(h.min(child1, child2), child3)
		} else if child2 < n {
			j = h.min(child1, child2)
		}

		if !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		i = j
	}
}

// Pop 弹出堆顶元素
// Pop pops the top element of the heap
func (h *Heap) Pop() *Element {
	n := h.Len()
	if n == 0 {
		return nil
	}
	ele := h.data[0]
	h.Swap(0, n-1)
	h.data = h.data[:n-1]
	h.down(0, n-1)
	return ele
}

// Delete 删除堆中的第 i 个元素
// Delete deletes the i-th element in the heap
func (h *Heap) Delete(i int) {
	n := h.Len()
	if n == 0 {
		return
	}
	if i >= n {
		return
	}
	h.Swap(i, n-1)
	h.data = h.data[:n-1]
	if i < n-1 {
		h.down(i, n-1)
		h.up(i)
	}
}

// Head 返回堆顶元素
// Head returns the top element of the heap
func (h *Heap) Head() *Element {
	if h.Len() == 0 {
		return nil
	}
	return h.data[0]
}

// Slice 返回堆中的元素
// Slice returns the elements in the heap
func (h *Heap) Slice() []*Element {
	return h.data
}
