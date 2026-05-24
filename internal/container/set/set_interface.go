package set

type InterfaceSet struct {
	m map[any]struct{}
}

func NewInterface() *InterfaceSet {
	return &InterfaceSet{m: make(map[any]struct{})}
}

func NewInterfaceWithCapacity(capacity int) *InterfaceSet {
	return &InterfaceSet{m: make(map[any]struct{}, capacity)}
}

func (s *InterfaceSet) Add(item any) {
	s.m[item] = setValue
}

func (s *InterfaceSet) Remove(item any) {
	delete(s.m, item)
}

func (s *InterfaceSet) Contains(item any) bool {
	_, c := s.m[item]
	return c
}

func (s *InterfaceSet) TryAdd(item any) bool {
	_, exists := s.m[item]
	if !exists {
		s.m[item] = setValue
	}
	return !exists
}

func (s *InterfaceSet) TryRemove(item any) bool {
	_, exists := s.m[item]
	if exists {
		delete(s.m, item)
	}
	return exists
}

func (s *InterfaceSet) Len() int {
	return len(s.m)
}

func (s *InterfaceSet) List() []any {
	list := make([]any, 0, len(s.m))
	for item := range s.m {
		list = append(list, item)
	}
	return list
}

func (s *InterfaceSet) Cleanup() {
	s.m = make(map[any]struct{})
}

func (s *InterfaceSet) GetOrCreate(item any) (exists bool) {
	_, exists = s.m[item]
	if !exists {
		s.m[item] = setValue
	}
	return
}

func (s *InterfaceSet) RemoveIfExists(item any) (existed bool) {
	_, existed = s.m[item]
	if existed {
		delete(s.m, item)
	}
	return
}
