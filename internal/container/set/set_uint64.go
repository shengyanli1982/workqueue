package set

// Uint64Set 是 uint64 专用集合，基于 map[uint64]struct{}，避免接口装箱开销。
type Uint64Set struct {
	m map[uint64]struct{}
}

// NewUint64 创建空的 uint64 专用集合。
func NewUint64() *Uint64Set {
	return &Uint64Set{m: make(map[uint64]struct{})}
}

// Add 添加 uint64 元素。
func (s *Uint64Set) Add(item uint64) {
	s.m[item] = struct{}{}
}

// Remove 删除 uint64 元素。
func (s *Uint64Set) Remove(item uint64) {
	delete(s.m, item)
}

// Contains 判断 uint64 元素是否存在。
func (s *Uint64Set) Contains(item uint64) bool {
	_, c := s.m[item]
	return c
}

// TryRemove 尝试删除 uint64 元素，不存在返回 false。
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

// List 将所有 uint64 元素收集到切片并返回。
func (s *Uint64Set) List() []uint64 {
	list := make([]uint64, 0, len(s.m))
	for item := range s.m {
		list = append(list, item)
	}
	return list
}

// Cleanup 清空并重新分配底层 map。
func (s *Uint64Set) Cleanup() {
	s.m = make(map[uint64]struct{})
}

// GetOrCreate 查找 uint64 元素：存在返回 true，不存在则添加并返回 false。
func (s *Uint64Set) GetOrCreate(item uint64) (exists bool) {
	_, exists = s.m[item]
	if !exists {
		s.m[item] = struct{}{}
	}
	return
}

// RemoveIfExists 查找并删除 uint64 元素：存在返回 true，不存在返回 false。
func (s *Uint64Set) RemoveIfExists(item uint64) (existed bool) {
	_, existed = s.m[item]
	if existed {
		delete(s.m, item)
	}
	return
}
