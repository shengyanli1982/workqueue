package set

// Uint64Set 是 uint64 专用集合，基于 map[uint64]struct{} 实现，
// 避免 map[any] 的接口装箱开销，适用于高频 uint64 元素的去重场景。
type Uint64Set struct {
	m map[uint64]struct{}
}

// NewUint64 创建一个空的 uint64 专用集合。
func NewUint64() *Uint64Set {
	return &Uint64Set{m: make(map[uint64]struct{})}
}

// Add 向集合中添加 uint64 元素。
func (s *Uint64Set) Add(item uint64) {
	s.m[item] = struct{}{}
}

// Remove 从集合中删除 uint64 元素。
func (s *Uint64Set) Remove(item uint64) {
	delete(s.m, item)
}

// Contains 判断 uint64 元素是否存在于集合中。
func (s *Uint64Set) Contains(item uint64) bool {
	_, c := s.m[item]
	return c
}

// TryRemove 尝试删除 uint64 元素，不存在时返回 false，成功时返回 true。
func (s *Uint64Set) TryRemove(item uint64) bool {
	_, exists := s.m[item]
	if exists {
		delete(s.m, item)
	}
	return exists
}

// Len 返回集合中的元素数量。
func (s *Uint64Set) Len() int {
	return len(s.m)
}

// List 将集合中所有 uint64 元素收集到切片并返回。
func (s *Uint64Set) List() []uint64 {
	list := make([]uint64, 0, len(s.m))
	for item := range s.m {
		list = append(list, item)
	}
	return list
}

// Cleanup 清空集合并重新分配底层 map。
func (s *Uint64Set) Cleanup() {
	s.m = make(map[uint64]struct{})
}

// GetOrCreate 查找 uint64 元素是否存在：存在时返回 true，不存在时添加并返回 false。
func (s *Uint64Set) GetOrCreate(item uint64) (exists bool) {
	_, exists = s.m[item]
	if !exists {
		s.m[item] = struct{}{}
	}
	return
}

// RemoveIfExists 查找并删除 uint64 元素：存在时删除并返回 true，不存在时返回 false。
func (s *Uint64Set) RemoveIfExists(item uint64) (existed bool) {
	_, existed = s.m[item]
	if existed {
		delete(s.m, item)
	}
	return
}
