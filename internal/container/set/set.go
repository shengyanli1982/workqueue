package set

var setValue = struct{}{}

// setKind 描述 Set 当前的特化存储类型。
type setKind uint8

const (
	kindNone setKind = iota // 空集合，所有 map 为 nil
	kindString              // 元素全为 string
	kindInt                 // 元素全为 int
	kindInt64               // 元素全为 int64
	kindUint64              // 元素全为 uint64
	kindAny                 // 通用路径：混合类型退化后的终态
)

// Set 是基于 map 的类型特化集合。首次写入时探测元素类型，string/int/int64/uint64
// 走专用标量 map，其他类型落入通用 map[any]struct{}。
//
// 退化规则（不可逆）：写入异型元素时整体迁移到 map[any]，不同静态类型即使值相同也不合并。
// 异型读/删除是空操作，不触发退化。存储惰性创建，首写时按类型建表。
type Set struct {
	kind    setKind
	initCap int

	mStr    map[string]struct{}
	mInt    map[int]struct{}
	mInt64  map[int64]struct{}
	mUint64 map[uint64]struct{}
	mAny    map[any]struct{}
}

// New 创建空的类型特化集合，底层 map 惰性创建。
func New() *Set {
	return &Set{}
}

// NewWithCapacity 创建指定初始容量的类型特化集合。
func NewWithCapacity(capacity int) *Set {
	return &Set{initCap: capacity}
}

// Add 添加元素。写入异型元素时不可逆退化到 map[any]。
func (s *Set) Add(item any) {
	switch s.kind {
	case kindNone:
		s.insertFirst(item)
	case kindString:
		if v, ok := item.(string); ok {
			s.mStr[v] = setValue
		} else {
			s.addMixed(item)
		}
	case kindInt:
		if v, ok := item.(int); ok {
			s.mInt[v] = setValue
		} else {
			s.addMixed(item)
		}
	case kindInt64:
		if v, ok := item.(int64); ok {
			s.mInt64[v] = setValue
		} else {
			s.addMixed(item)
		}
	case kindUint64:
		if v, ok := item.(uint64); ok {
			s.mUint64[v] = setValue
		} else {
			s.addMixed(item)
		}
	default: // kindAny
		s.mAny[item] = setValue
	}
}

// Remove 删除元素。异型元素的删除是空操作。
func (s *Set) Remove(item any) {
	switch s.kind {
	case kindNone:
		return
	case kindString:
		if v, ok := item.(string); ok {
			delete(s.mStr, v)
		}
	case kindInt:
		if v, ok := item.(int); ok {
			delete(s.mInt, v)
		}
	case kindInt64:
		if v, ok := item.(int64); ok {
			delete(s.mInt64, v)
		}
	case kindUint64:
		if v, ok := item.(uint64); ok {
			delete(s.mUint64, v)
		}
	default: // kindAny
		delete(s.mAny, item)
	}
}

// Contains 判断元素是否存在。异型元素返回 false。
func (s *Set) Contains(item any) bool {
	switch s.kind {
	case kindNone:
		return false
	case kindString:
		v, ok := item.(string)
		if !ok {
			return false
		}

		_, c := s.mStr[v]

		return c
	case kindInt:
		v, ok := item.(int)
		if !ok {
			return false
		}

		_, c := s.mInt[v]

		return c
	case kindInt64:
		v, ok := item.(int64)
		if !ok {
			return false
		}

		_, c := s.mInt64[v]

		return c
	case kindUint64:
		v, ok := item.(uint64)
		if !ok {
			return false
		}

		_, c := s.mUint64[v]

		return c
	default: // kindAny
		_, c := s.mAny[item]

		return c
	}
}

// TryAdd 尝试添加元素，已存在返回 false。写入异型元素时退化到 map[any]。
func (s *Set) TryAdd(item any) bool {
	switch s.kind {
	case kindNone:
		s.insertFirst(item)

		return true
	case kindString:
		v, ok := item.(string)
		if !ok {
			return s.tryAddMixed(item)
		}

		if _, exists := s.mStr[v]; exists {
			return false
		}

		s.mStr[v] = setValue

		return true
	case kindInt:
		v, ok := item.(int)
		if !ok {
			return s.tryAddMixed(item)
		}

		if _, exists := s.mInt[v]; exists {
			return false
		}

		s.mInt[v] = setValue

		return true
	case kindInt64:
		v, ok := item.(int64)
		if !ok {
			return s.tryAddMixed(item)
		}

		if _, exists := s.mInt64[v]; exists {
			return false
		}

		s.mInt64[v] = setValue

		return true
	case kindUint64:
		v, ok := item.(uint64)
		if !ok {
			return s.tryAddMixed(item)
		}

		if _, exists := s.mUint64[v]; exists {
			return false
		}

		s.mUint64[v] = setValue

		return true
	default: // kindAny
		_, exists := s.mAny[item]
		if !exists {
			s.mAny[item] = setValue
		}

		return !exists
	}
}

// TryRemove 尝试删除元素，不存在或类型不匹配返回 false。
func (s *Set) TryRemove(item any) bool {
	switch s.kind {
	case kindNone:
		return false
	case kindString:
		v, ok := item.(string)
		if !ok {
			return false
		}

		if _, exists := s.mStr[v]; !exists {
			return false
		}

		delete(s.mStr, v)

		return true
	case kindInt:
		v, ok := item.(int)
		if !ok {
			return false
		}

		if _, exists := s.mInt[v]; !exists {
			return false
		}

		delete(s.mInt, v)

		return true
	case kindInt64:
		v, ok := item.(int64)
		if !ok {
			return false
		}

		if _, exists := s.mInt64[v]; !exists {
			return false
		}

		delete(s.mInt64, v)

		return true
	case kindUint64:
		v, ok := item.(uint64)
		if !ok {
			return false
		}

		if _, exists := s.mUint64[v]; !exists {
			return false
		}

		delete(s.mUint64, v)

		return true
	default: // kindAny
		_, exists := s.mAny[item]
		if exists {
			delete(s.mAny, item)
		}

		return exists
	}
}

// Len 返回集合中的元素数量。
func (s *Set) Len() int {
	switch s.kind {
	case kindNone:
		return 0
	case kindString:
		return len(s.mStr)
	case kindInt:
		return len(s.mInt)
	case kindInt64:
		return len(s.mInt64)
	case kindUint64:
		return len(s.mUint64)
	default: // kindAny
		return len(s.mAny)
	}
}

// List 将所有元素收集到切片并返回。
func (s *Set) List() []any {
	switch s.kind {
	case kindNone:
		return make([]any, 0)
	case kindString:
		list := make([]any, 0, len(s.mStr))

		for item := range s.mStr {
			list = append(list, item)
		}

		return list
	case kindInt:
		list := make([]any, 0, len(s.mInt))

		for item := range s.mInt {
			list = append(list, item)
		}

		return list
	case kindInt64:
		list := make([]any, 0, len(s.mInt64))

		for item := range s.mInt64 {
			list = append(list, item)
		}

		return list
	case kindUint64:
		list := make([]any, 0, len(s.mUint64))

		for item := range s.mUint64 {
			list = append(list, item)
		}

		return list
	default: // kindAny
		list := make([]any, 0, len(s.mAny))

		for item := range s.mAny {
			list = append(list, item)
		}

		return list
	}
}

// Cleanup 清空所有底层 map，类型重置为 kindNone。
func (s *Set) Cleanup() {
	s.kind = kindNone
	s.mStr = nil
	s.mInt = nil
	s.mInt64 = nil
	s.mUint64 = nil
	s.mAny = nil
}

// insertFirst 处理空集合首次写入：探测类型并创建对应 map。
//
//go:noinline
func (s *Set) insertFirst(item any) {
	switch v := item.(type) {
	case string:
		s.kind = kindString
		s.mStr = make(map[string]struct{}, s.initCap)
		s.mStr[v] = setValue
	case int:
		s.kind = kindInt
		s.mInt = make(map[int]struct{}, s.initCap)
		s.mInt[v] = setValue
	case int64:
		s.kind = kindInt64
		s.mInt64 = make(map[int64]struct{}, s.initCap)
		s.mInt64[v] = setValue
	case uint64:
		s.kind = kindUint64
		s.mUint64 = make(map[uint64]struct{}, s.initCap)
		s.mUint64[v] = setValue
	default:
		s.kind = kindAny
		s.mAny = make(map[any]struct{}, s.initCap)
		s.mAny[item] = setValue
	}
}

// demoteToAny 将特化集合整体迁移到 map[any]（不可逆），释放专用 map。
//
//go:noinline
func (s *Set) demoteToAny() {
	m := make(map[any]struct{}, s.Len()+1)

	switch s.kind {
	case kindString:
		for k := range s.mStr {
			m[k] = setValue
		}

		s.mStr = nil
	case kindInt:
		for k := range s.mInt {
			m[k] = setValue
		}

		s.mInt = nil
	case kindInt64:
		for k := range s.mInt64 {
			m[k] = setValue
		}

		s.mInt64 = nil
	case kindUint64:
		for k := range s.mUint64 {
			m[k] = setValue
		}

		s.mUint64 = nil
	default:
		return // kindAny/kindNone 无需迁移
	}

	s.mAny = m
	s.kind = kindAny
}

//go:noinline
func (s *Set) addMixed(item any) {
	s.demoteToAny()
	s.mAny[item] = setValue
}

//go:noinline
func (s *Set) tryAddMixed(item any) bool {
	s.demoteToAny()

	_, exists := s.mAny[item]
	if !exists {
		s.mAny[item] = setValue
	}

	return !exists
}
