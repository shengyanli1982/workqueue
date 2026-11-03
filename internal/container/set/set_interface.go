package set

// InterfaceSet 基于 map[any]struct{} 的通用集合，不做类型特化。
type InterfaceSet struct {
	m map[any]struct{}
}

// NewInterface 创建空的通用集合。
func NewInterface() *InterfaceSet {
	return &InterfaceSet{m: make(map[any]struct{})}
}

// NewInterfaceWithCapacity 创建指定初始容量的通用集合。
func NewInterfaceWithCapacity(capacity int) *InterfaceSet {
	return &InterfaceSet{m: make(map[any]struct{}, capacity)}
}

// Add 添加元素。
func (s *InterfaceSet) Add(item any) {
	s.m[item] = setValue
}

// Remove 删除元素。
func (s *InterfaceSet) Remove(item any) {
	delete(s.m, item)
}

// Contains 判断元素是否存在。
func (s *InterfaceSet) Contains(item any) bool {
	_, c := s.m[item]
	return c
}

// TryAdd 尝试添加元素，已存在返回 false。
func (s *InterfaceSet) TryAdd(item any) bool {
	_, exists := s.m[item]
	if !exists {
		s.m[item] = setValue
	}
	return !exists
}

// TryRemove 尝试删除元素，不存在返回 false。
func (s *InterfaceSet) TryRemove(item any) bool {
	_, exists := s.m[item]
	if exists {
		delete(s.m, item)
	}
	return exists
}

// Len 返回集合中的元素数量。
func (s *InterfaceSet) Len() int {
	return len(s.m)
}

// List 将所有元素收集到切片并返回。
func (s *InterfaceSet) List() []any {
	list := make([]any, 0, len(s.m))
	for item := range s.m {
		list = append(list, item)
	}
	return list
}

// Cleanup 清空并重新分配底层 map。
func (s *InterfaceSet) Cleanup() {
	s.m = make(map[any]struct{})
}

// GetOrCreate 查找元素：存在返回 true，不存在则添加并返回 false。
func (s *InterfaceSet) GetOrCreate(item any) (exists bool) {
	_, exists = s.m[item]
	if !exists {
		s.m[item] = setValue
	}
	return
}

// RemoveIfExists 查找并删除元素：存在返回 true，不存在返回 false。
func (s *InterfaceSet) RemoveIfExists(item any) (existed bool) {
	_, existed = s.m[item]
	if existed {
		delete(s.m, item)
	}
	return
}
