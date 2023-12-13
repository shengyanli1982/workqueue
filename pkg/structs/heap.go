package workqueue

type Element struct {
	data  any
	value int64
	index int
}

func (e *Element) Data() any {
	return e.data
}

func (e *Element) Value() int64 {
	return e.value
}

func (e *Element) Index() int {
	return e.index
}

func (e *Element) ResetValue(i int64) {
	e.value = i
}

func (e *Element) Reset() {
	e.data = nil
	e.value = 0
	e.index = 0
}

func NewElement(data any, value int64) *Element {
	return &Element{data: data, value: value}
}

type Heap struct {
	data []*Element
}

func NewHeap() *Heap {
	return &Heap{}
}

func (c *Heap) Reset() {
	c.data = c.data[:0]
}

func (c *Heap) Less(i, j int) bool { return c.data[i].value < c.data[j].value }

func (c *Heap) Update(ele *Element, value int64) {
	var down = value > ele.value
	ele.value = value
	if down {
		c.Down(ele.index, c.Len())
	} else {
		c.Up(ele.index)
	}
}

func (c *Heap) min(i, j int) int {
	if c.data[i].value < c.data[j].value {
		return i
	}
	return j
}

func (c *Heap) Len() int {
	return len(c.data)
}

func (c *Heap) Swap(i, j int) {
	c.data[i].index, c.data[j].index = c.data[j].index, c.data[i].index
	c.data[i], c.data[j] = c.data[j], c.data[i]
}

func (c *Heap) Push(ele *Element) {
	ele.index = c.Len()
	c.data = append(c.data, ele)
	c.Up(c.Len() - 1)
}

// Up 从 i 开始向上调整堆
// Adjust the heap from i upwards
func (c *Heap) Up(i int) {
	var j = (i - 1) >> 2
	if i >= 1 && c.Less(i, j) {
		c.Swap(i, j)
		c.Up(j)
	}
}

// Pop 弹出堆顶元素
// Pop the top Element of the heap
func (c *Heap) Pop() (ele *Element) {
	var n = c.Len()
	switch n {
	case 0:
	case 1:
		ele = c.data[0]
		c.data = c.data[:0]
	default:
		ele = c.data[0]
		c.Swap(0, n-1)
		c.data = c.data[:n-1]
		c.Down(0, n-1)
	}
	return
}

// Delete 删除堆中的第 i 个元素
// Delete the i-th element in the heap
func (c *Heap) Delete(i int) {
	var n = c.Len()
	switch n {
	case 1:
		c.data = c.data[:0]
	default:
		var down = c.Less(i, n-1)
		c.Swap(i, n-1)
		c.data = c.data[:n-1]
		if i < n-1 {
			if down {
				c.Down(i, n-1)
			} else {
				c.Up(i)
			}
		}
	}
}

// Down 从 i 开始向下调整堆
// Adjust the heap from i downwards
func (c *Heap) Down(i, n int) {
	var index1 = i<<2 + 1
	if index1 >= n {
		return
	}

	var index2 = i<<2 + 2
	var index3 = i<<2 + 3
	var index4 = i<<2 + 4
	var j int

	if index4 < n {
		j = c.min(c.min(index1, index2), c.min(index3, index4))
	} else if index3 < n {
		j = c.min(c.min(index1, index2), index3)
	} else if index2 < n {
		j = c.min(index1, index2)
	} else {
		j = index1
	}

	if j >= 0 && c.Less(j, i) {
		c.Swap(i, j)
		c.Down(j, n)
	}
}

// Head 访问堆顶元素
// Accessing the top Element of the heap
func (c *Heap) Head() *Element {
	return c.data[0]
}

// Slice 返回堆中的所有元素
// Return all elements in the heap
func (c *Heap) Slice() []*Element {
	return c.data
}
