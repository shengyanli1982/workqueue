package stl

// 使用 map 实现一个 map[any]struct{} 的 set

// Set 是任意类型的集合。
// Set is a set of any type.
type Set map[any]struct{}

// NewSet 返回一个新的集合。
// NewSet returns a new set.
func NewSet() Set {
	return make(Set)
}

// Has 判断集合中是否存在指定对象。
// Has returns true if the set contains the specified object.
func (s Set) Has(i any) bool {
	_, exists := s[i]
	return exists
}

// Add 向集合中添加对象。
// Add adds an object to the set.
func (s Set) Add(i any) {
	s[i] = struct{}{}
}

// Delete 从集合中删除对象。
// Delete removes an object from the set.
func (s Set) Delete(i any) {
	delete(s, i)
}

// Len 返回集合中对象的数量。
// Len returns the number of objects in the set.
func (s Set) Len() int {
	return len(s)
}

// Cleanup 清空集合。
// Cleanup empties the set.
func (s Set) Cleanup() {
	for k := range s {
		delete(s, k)
	}
}
