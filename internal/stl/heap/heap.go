package heap

// Element 是堆中元素的结构
// Element is the structure of the element in the heap
type Element struct {
	data  any   // 元素的数据
	value int64 // 元素的值
	index int   // 元素在堆中的索引
}

// Data 方法返回元素的数据
// The Data method returns the data of the element
func (e *Element) Data() any {
	return e.data
}

// Value 方法返回元素的值
// The Value method returns the value of the element
func (e *Element) Value() int64 {
	return e.value
}

// Index 方法返回元素的索引
// The Index method returns the index of the element
func (e *Element) Index() int {
	return e.index
}

// SetValue 方法设置元素的值
// The SetValue method sets the value of the element
func (e *Element) SetValue(i int64) {
	e.value = i
}

// SetData 方法设置元素的数据
// The SetData method sets the data of the element
func (e *Element) SetData(data any) {
	e.data = data
}

// Reset 方法重置元素
// The Reset method resets the element
func (e *Element) Reset() {
	e.data = nil
	e.value = 0
	e.index = 0
}

// NewElement 函数返回一个新的元素
// The NewElement function returns a new element
func NewElement(data any, value int64) *Element {
	return &Element{data: data, value: value}
}

// Heap 是一个最小4叉堆
// Heap is a minimum 4-ary heap
type Heap struct {
	data []*Element // 堆中的元素
}

// NewHeap 函数返回一个新的堆
// The NewHeap function returns a new heap
func NewHeap() *Heap {
	return &Heap{}
}

// Reset 方法重置堆
// The Reset method resets the heap
func (h *Heap) Reset() {
	h.data = h.data[:0] // 清空堆中的元素
}

// Less 方法用于比较堆中两个元素的值，如果第 i 个元素的值小于第 j 个元素的值，返回 true，否则返回 false
// The Less method is used to compare the values of two elements in the heap. If the value of the i-th element is less than the value of the j-th element, return true, otherwise return false
func (h *Heap) Less(i, j int) bool { return h.data[i].value < h.data[j].value }

// Update 方法用于更新元素的值，如果新的值大于元素的当前值，将元素向下移动，否则将元素向上移动
// The Update method is used to update the value of the element. If the new value is greater than the current value of the element, move the element down, otherwise move the element up
func (h *Heap) Update(ele *Element, value int64) {
	if value > ele.value {
		h.down(ele.index, h.Len())
	} else {
		h.up(ele.index)
	}
	ele.value = value
}

// Len 方法返回堆中元素的个数
// The Len method returns the number of elements in the heap
func (h *Heap) Len() int {
	return len(h.data)
}

// Swap 方法用于交换堆中两个元素的位置
// The Swap method is used to swap the positions of two elements in the heap
func (h *Heap) Swap(i, j int) {
	h.data[i].index, h.data[j].index = h.data[j].index, h.data[i].index
	h.data[i], h.data[j] = h.data[j], h.data[i]
}

// Push 方法用于向堆中添加一个元素，添加后，将元素向上移动以保持堆的性质
// The Push method is used to add an element to the heap. After adding, move the element up to maintain the property of the heap
func (h *Heap) Push(ele *Element) {
	ele.index = h.Len()
	h.data = append(h.data, ele)
	h.up(h.Len() - 1)
}

// up 方法从 i 开始向上调整堆，以保持堆的性质
// The up method adjusts the heap upwards from i to maintain the property of the heap
func (h *Heap) up(i int) {
	// 当 i 大于 0 时，继续向上调整
	// Continue to adjust upwards when i is greater than 0
	for i > 0 {
		// 计算父节点的索引
		// Calculate the index of the parent node
		parent := (i - 1) >> 2

		// 如果 i 节点的值不小于其父节点的值，停止调整
		// If the value of the i node is not less than the value of its parent node, stop adjusting
		if !h.Less(i, parent) {
			break
		}

		// 交换 i 节点和其父节点的位置
		// Swap the positions of the i node and its parent node
		h.Swap(i, parent)

		// 将 i 设置为其父节点的索引，继续向上调整
		// Set i to the index of its parent node and continue to adjust upwards
		i = parent
	}
}

// down 方法从 i 开始向下调整堆，以保持堆的性质
// The down method adjusts the heap downwards from i to maintain the property of the heap
func (h *Heap) down(i, n int) {
	// 当 i 的子节点存在时，继续向下调整
	// Continue to adjust downwards when the child node of i exists
	for {
		// 计算 i 节点的第一个子节点的索引
		// Calculate the index of the first child node of the i node
		c1 := i<<2 + 1
		if c1 >= n {
			break
		}

		// 计算 i 节点的其他子节点的索引
		// Calculate the indexes of the other child nodes of the i node
		c2 := c1 + 1
		c3 := c1 + 2
		c4 := c1 + 3

		// 初始化 j 为 i 节点的第一个子节点的索引
		// Initialize j as the index of the first child node of the i node
		j := c1

		// 找出 i 节点的子节点中值最小的节点的索引
		// Find the index of the node with the smallest value among the child nodes of the i node
		if c2 < n && h.Less(c2, j) {
			j = c2
		} else if c3 < n && h.Less(c3, j) {
			j = c3
		} else if c4 < n && h.Less(c4, j) {
			j = c4
		}

		// 如果 i 节点的值不大于其子节点中值最小的节点的值，停止调整
		// If the value of the i node is not greater than the value of the node with the smallest value among its child nodes, stop adjusting
		if !h.Less(j, i) {
			break
		}

		// 交换 i 节点和其子节点中值最小的节点的位置
		// Swap the positions of the i node and the node with the smallest value among its child nodes
		h.Swap(i, j)

		// 将 i 设置为其子节点中值最小的节点的索引，继续向下调整
		// Set i to the index of the node with the smallest value among its child nodes and continue to adjust downwards
		i = j
	}
}

// Pop 弹出堆顶元素
// Pop pops the top element of the heap
func (h *Heap) Pop() *Element {
	// 获取堆的长度
	// Get the length of the heap
	n := h.Len()

	// 如果堆为空，返回 nil
	// If the heap is empty, return nil
	if n == 0 {
		return nil
	}

	// 获取堆顶元素
	// Get the top element of the heap
	ele := h.data[0]

	// 将堆顶元素与堆的最后一个元素交换位置
	// Swap the top element of the heap with the last element of the heap
	h.Swap(0, n-1)

	// 删除堆的最后一个元素
	// Delete the last element of the heap
	h.data = h.data[:n-1]

	// 从堆顶开始向下调整堆
	// Adjust the heap downwards from the top of the heap
	h.down(0, n-1)

	// 返回原堆顶元素
	// Return the original top element of the heap
	return ele
}

// Delete 删除堆中的第 i 个元素
// Delete deletes the i-th element in the heap
func (h *Heap) Delete(i int) {
	// 获取堆的长度
	// Get the length of the heap
	n := h.Len()

	// 如果堆为空，或者 i 不在堆的索引范围内，直接返回
	// If the heap is empty, or i is not within the index range of the heap, return directly
	if n == 0 || i >= n {
		return
	}

	// 将第 i 个元素与堆的最后一个元素交换位置
	// Swap the i-th element with the last element of the heap
	h.Swap(i, n-1)

	// 删除堆的最后一个元素
	// Delete the last element of the heap
	h.data = h.data[:n-1]

	// 如果 i 小于原堆的长度减一，从 i 开始向下调整堆，然后从 i 开始向上调整堆
	// If i is less than the length of the original heap minus one, adjust the heap downwards from i, and then adjust the heap upwards from i
	if i < n-1 {
		h.down(i, n-1)
		h.up(i)
	}
}

// Head 返回堆顶元素
// Head returns the top element of the heap
func (h *Heap) Head() *Element {
	// 如果堆为空，返回 nil
	// If the heap is empty, return nil
	if h.Len() == 0 {
		return nil
	}

	// 返回堆顶元素
	// Return the top element of the heap
	return h.data[0]
}

// Slice 返回堆中的元素
// Slice returns the elements in the heap
func (h *Heap) Slice() []*Element {
	// 返回堆中的所有元素
	// Return all elements in the heap
	return h.data
}
