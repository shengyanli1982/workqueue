package set

var setValue = struct{}{}

// setKind 描述 Set 当前的特化存储类型。
type setKind uint8

const (
	// kindNone 空集合：类型未定，全部底层 map 均为 nil（零预分配）。
	kindNone setKind = iota

	// kindString 元素全为 string，走 mStr。
	kindString

	// kindInt 元素全为 int，走 mInt。
	kindInt

	// kindInt64 元素全为 int64，走 mInt64。
	kindInt64

	// kindUint64 元素全为 uint64，走 mUint64。
	kindUint64

	// kindAny 通用路径：元素为其他类型，或混合类型退化后的终态，走 mAny。
	kindAny
)

// Set 是一个基于 map 的轻量集合实现。
//
// 为避免 map[any] 在热路径上的接口装箱与动态哈希分发开销，Set 在首次写入
// 时探测元素类型并做类型特化：string / int / int64 / uint64 直接存入对应
// 标量 map（map[string]struct{} 等），其他类型落入通用 map[any]struct{}。
//
// 退化规则（不可逆）：同型集合在写入（Add/TryAdd）到异型元素时整体迁移回
// 通用 map[any]，以严格保持与 map[any] 相同的元素同一性语义——不同静态
// 类型即使值相同（如 1、int64(1)、uint64(1)）也永不合并。异型元素的读
// （Contains）与删除（Remove/TryRemove）是空操作，不触发退化。
//
// 存储惰性创建：New/NewWithCapacity 不预分配任何 map，首写时按探测到的
// 类型建表；空集合的所有读操作零成本返回空结果。
type Set struct {
	kind    setKind
	initCap int

	mStr    map[string]struct{}
	mInt    map[int]struct{}
	mInt64  map[int64]struct{}
	mUint64 map[uint64]struct{}
	mAny    map[any]struct{}
}

// New 创建一个空的类型特化集合，底层 map 惰性创建，首次写入时才分配。
func New() *Set {
	return &Set{}
}

// NewWithCapacity 创建一个指定初始容量的类型特化集合，底层 map 在首次写入时按该容量创建。
func NewWithCapacity(capacity int) *Set {
	return &Set{initCap: capacity}
}

// Add 向集合中添加元素。首次写入时探测类型并走专用 map 路径；
// 同型集合写入异型元素时不可逆退化到通用 map[any]。
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

// Remove 从集合中删除元素。异型元素的删除是空操作，不触发退化。
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

// Contains 判断元素是否存在于集合中。异型元素的查询返回 false。
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

// TryAdd 尝试添加元素到集合中，元素已存在时返回 false，成功添加时返回 true。
// 同型集合写入异型元素时不可逆退化到通用 map[any]。
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

// TryRemove 尝试从集合中删除元素，元素不存在或类型不匹配时返回 false，成功删除时返回 true。
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

// List 将集合中所有元素收集到切片并返回。空集合返回空切片。
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

// Cleanup 清空集合中的所有底层 map 并将类型重置为 kindNone，集合回到空状态。
func (s *Set) Cleanup() {
	s.kind = kindNone
	s.mStr = nil
	s.mInt = nil
	s.mInt64 = nil
	s.mUint64 = nil
	s.mAny = nil
}

// insertFirst 处理空集合的首次写入：探测元素类型，创建对应专用 map 并插入。
// 非特化类型（含 nil）直接落入通用 map[any]。
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

// demoteToAny 在异型元素写入时把特化集合整体迁移到通用 map[any]（不可逆），
// 并释放专用 map。仅在 kind 为特化类型时执行迁移。
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
		// kindAny/kindNone 无需迁移；kindNone 不应到达此处（调用方先走 insertFirst）。
		return
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
