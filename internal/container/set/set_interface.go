package set

// InterfaceSet 是基于 map[any]struct{} 的通用集合，不做类型特化，
// 适用于元素类型多样且无需优化接口装箱开销的场景。
type InterfaceSet struct {
	m map[any]struct{}
}

// NewInterface 创建一个空的通用集合。
func NewInterface() *InterfaceSet {
	return &InterfaceSet{m: make(map[any]struct{})}
}

// NewInterfaceWithCapacity 创建一个指定初始容量的通用集合。
func NewInterfaceWithCapacity(capacity int) *InterfaceSet {
	return &InterfaceSet{m: make(map[any]struct{}, capacity)}
}

// Add 向集合中添加元素。
func (s *InterfaceSet) Add(item any) {
	s.m[item] = setValue
}

// Remove 从集合中删除元素。
func (s *InterfaceSet) Remove(item any) {
	delete(s.m, item)
}

// Contains 判断元素是否存在于集合中。
func (s *InterfaceSet) Contains(item any) bool {
	_, c := s.m[item]
	return c
}

// TryAdd 尝试添加元素，已存在时返回 false，成功时返回 true。
func (s *InterfaceSet) TryAdd(item any) bool {
	_, exists := s.m[item]
	if !exists {
		s.m[item] = setValue
	}
	return !exists
}

// TryRemove 尝试删除元素，不存在时返回 false，成功时返回 true。
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

// List 将集合中所有元素收集到切片并返回。
func (s *InterfaceSet) List() []any {
	list := make([]any, 0, len(s.m))
	for item := range s.m {
		list = append(list, item)
	}
	return list
}

// Cleanup 清空集合并重新分配底层 map。
func (s *InterfaceSet) Cleanup() {
	s.m = make(map[any]struct{})
}

// GetOrCreate 查找元素是否存在：存在时返回 true（不修改集合），不存在时添加并返回 false。
func (s *InterfaceSet) GetOrCreate(item any) (exists bool) {
	_, exists = s.m[item]
	if !exists {
		s.m[item] = setValue
	}
	return
}

// RemoveIfExists 查找并删除元素：存在时删除并返回 true，不存在时返回 false。
func (s *InterfaceSet) RemoveIfExists(item any) (existed bool) {
	_, existed = s.m[item]
	if existed {
		delete(s.m, item)
	}
	return
}
