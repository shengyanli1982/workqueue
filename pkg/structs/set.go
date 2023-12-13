package workqueue

// 用 map 实现一个 map[any]struct{} 的 set

// set is a set of any type.
type empty struct{}

// Set is a Set of any type.
type Set map[any]empty

// NewSet returns a new set.
func NewSet() Set {
	return make(Set)
}

// object in the set.
func (s Set) Has(i any) bool {
	_, exists := s[i]
	return exists
}

// Add object in the set.
func (s Set) Add(i any) {
	s[i] = empty{}
}

// Delete object in the set.
func (s Set) Delete(i any) {
	delete(s, i)
}

// Len returns the number of objects in the set.
func (s Set) Len() int {
	return len(s)
}

func (s Set) Cleanup() {
	for k := range s {
		delete(s, k)
	}
}
