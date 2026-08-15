package set

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSet_StringSemantics 固化 string 元素的全生命周期语义。
// 该组测试描述的是 map[any]struct{} 基线实现的行为，任何内部优化
// （包括类型特化）之后必须保持逐条成立。
func TestSet_StringSemantics(t *testing.T) {
	s := New()
	defer s.Cleanup()

	// 空集合：查询/移除均为空结果。
	assert.Zero(t, s.Len())
	assert.Empty(t, s.List())
	assert.False(t, s.Contains("a"))
	assert.False(t, s.TryRemove("a"))

	// Add 与 Contains。
	s.Add("a")
	assert.True(t, s.Contains("a"))
	assert.Equal(t, 1, s.Len())

	// 重复 Add 幂等，不改变规模。
	s.Add("a")
	assert.Equal(t, 1, s.Len())

	// TryAdd：新元素返回 true，已存在返回 false。
	assert.True(t, s.TryAdd("b"))
	assert.False(t, s.TryAdd("a"))
	assert.False(t, s.TryAdd("b"))
	assert.Equal(t, 2, s.Len())

	// Remove / TryRemove。
	s.Remove("b")
	assert.False(t, s.Contains("b"))
	assert.False(t, s.TryRemove("b"))
	assert.True(t, s.TryRemove("a"))
	assert.False(t, s.Contains("a"))
	assert.Zero(t, s.Len())

	// 清空后可继续复用。
	assert.True(t, s.TryAdd("c"))
	assert.True(t, s.Contains("c"))

	// List 返回全部元素（内容一致）。
	s.Add("d")
	assert.ElementsMatch(t, []any{"c", "d"}, s.List())

	// Cleanup 回到空集合。
	s.Cleanup()
	assert.Zero(t, s.Len())
	assert.Empty(t, s.List())
	assert.False(t, s.Contains("c"))

	// Cleanup 后仍可写入。
	assert.True(t, s.TryAdd("e"))
	assert.Equal(t, 1, s.Len())
	assert.True(t, s.Contains("e"))
}

// TestSet_Int64Semantics 固化 int64 元素的全生命周期语义。
func TestSet_Int64Semantics(t *testing.T) {
	s := NewWithCapacity(8)
	defer s.Cleanup()

	// 空集合：查询/移除均为空结果。
	assert.Zero(t, s.Len())
	assert.Empty(t, s.List())
	assert.False(t, s.Contains(int64(1)))
	assert.False(t, s.TryRemove(int64(1)))

	// Add 与 Contains（含负数与零边界）。
	s.Add(int64(0))
	s.Add(int64(-1))
	s.Add(int64(1))
	assert.True(t, s.Contains(int64(0)))
	assert.True(t, s.Contains(int64(-1)))
	assert.True(t, s.Contains(int64(1)))
	assert.Equal(t, 3, s.Len())

	// 重复 Add 幂等。
	s.Add(int64(1))
	assert.Equal(t, 3, s.Len())

	// TryAdd 语义。
	assert.True(t, s.TryAdd(int64(2)))
	assert.False(t, s.TryAdd(int64(0)))
	assert.False(t, s.TryAdd(int64(2)))
	assert.Equal(t, 4, s.Len())

	// Remove / TryRemove。
	s.Remove(int64(-1))
	assert.False(t, s.Contains(int64(-1)))
	assert.False(t, s.TryRemove(int64(-1)))
	assert.True(t, s.TryRemove(int64(0)))
	assert.Equal(t, 2, s.Len())

	// List 返回全部元素。
	assert.ElementsMatch(t, []any{int64(1), int64(2)}, s.List())

	// Cleanup 回到空集合且可复用。
	s.Cleanup()
	assert.Zero(t, s.Len())
	assert.Empty(t, s.List())
	assert.True(t, s.TryAdd(int64(9)))
	assert.True(t, s.Contains(int64(9)))
}

// TestSet_MixedTypeSemantics 固化混合类型写入的语义：不同静态类型的
// 元素绝不合并（与 map[any]struct{} 的元素同一性规则完全一致）。
func TestSet_MixedTypeSemantics(t *testing.T) {
	s := New()
	defer s.Cleanup()

	type custom struct{ A int }

	s.Add("a")
	s.Add(1)
	s.Add(int64(1))
	s.Add(uint64(1))
	s.Add(float64(1))
	s.Add(custom{A: 1})
	s.Add(nil)

	assert.Equal(t, 7, s.Len())

	// 每种类型各自可见。
	assert.True(t, s.Contains("a"))
	assert.True(t, s.Contains(1))
	assert.True(t, s.Contains(int64(1)))
	assert.True(t, s.Contains(uint64(1)))
	assert.True(t, s.Contains(float64(1)))
	assert.True(t, s.Contains(custom{A: 1}))
	assert.True(t, s.Contains(nil))

	// 相似值但不同类型不可互相命中。
	assert.False(t, s.Contains("1"))
	assert.False(t, s.Contains(int8(1)))
	assert.False(t, s.Contains(int32(1)))
	assert.False(t, s.Contains(uint(1)))
	assert.False(t, s.Contains(true))

	// TryAdd 对已存在的任一类型元素返回 false。
	assert.False(t, s.TryAdd("a"))
	assert.False(t, s.TryAdd(int64(1)))
	assert.False(t, s.TryAdd(nil))

	// 移除其中一种类型不影响其他类型。
	assert.True(t, s.TryRemove(int64(1)))
	assert.False(t, s.Contains(int64(1)))
	assert.True(t, s.Contains(1))
	assert.True(t, s.Contains(uint64(1)))
	assert.True(t, s.Contains(float64(1)))
	assert.Equal(t, 6, s.Len())

	// 删除不存在的类型返回 false。
	assert.False(t, s.TryRemove(int32(1)))
	assert.False(t, s.TryRemove("b"))

	// List 返回全部剩余元素。
	assert.ElementsMatch(t,
		[]any{"a", 1, uint64(1), float64(1), custom{A: 1}, nil},
		s.List())

	// 被删除的类型可以重新加入。
	assert.True(t, s.TryAdd(int64(1)))
	assert.True(t, s.Contains(int64(1)))
	assert.Equal(t, 7, s.Len())

	// Cleanup 全部清空。
	s.Cleanup()
	assert.Zero(t, s.Len())
	assert.Empty(t, s.List())
	assert.False(t, s.Contains("a"))
	assert.False(t, s.Contains(nil))
}
