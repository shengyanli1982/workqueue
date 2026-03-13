package set

var setValue = struct{}{}

// Set 是一个基于 map 的轻量集合实现。
type Set struct {
	m map[interface{}]struct{}
}

func New() *Set {
	return &Set{

		m: make(map[interface{}]struct{}),
	}
}

func (s *Set) Add(item interface{}) {

	s.m[item] = setValue
}

func (s *Set) Remove(item interface{}) {

	delete(s.m, item)
}

func (s *Set) Contains(item interface{}) bool {

	_, c := s.m[item]
	return c
}

func (s *Set) Len() int {

	return len(s.m)
}

func (s *Set) List() []interface{} {

	list := make([]interface{}, 0, len(s.m))

	for item := range s.m {
		list = append(list, item)
	}

	return list
}

func (s *Set) Cleanup() {

	s.m = make(map[interface{}]struct{})
}
