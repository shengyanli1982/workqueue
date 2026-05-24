package set

type Uint64Set struct {
	m map[uint64]struct{}
}

func NewUint64() *Uint64Set {
	return &Uint64Set{m: make(map[uint64]struct{})}
}

func (s *Uint64Set) Add(item uint64) {
	s.m[item] = struct{}{}
}

func (s *Uint64Set) Remove(item uint64) {
	delete(s.m, item)
}

func (s *Uint64Set) Contains(item uint64) bool {
	_, c := s.m[item]
	return c
}

func (s *Uint64Set) TryRemove(item uint64) bool {
	_, exists := s.m[item]
	if exists {
		delete(s.m, item)
	}
	return exists
}

func (s *Uint64Set) Len() int {
	return len(s.m)
}

func (s *Uint64Set) List() []uint64 {
	list := make([]uint64, 0, len(s.m))
	for item := range s.m {
		list = append(list, item)
	}
	return list
}

func (s *Uint64Set) Cleanup() {
	s.m = make(map[uint64]struct{})
}

func (s *Uint64Set) GetOrCreate(item uint64) (exists bool) {
	_, exists = s.m[item]
	if !exists {
		s.m[item] = struct{}{}
	}
	return
}

func (s *Uint64Set) RemoveIfExists(item uint64) (existed bool) {
	_, existed = s.m[item]
	if existed {
		delete(s.m, item)
	}
	return
}
