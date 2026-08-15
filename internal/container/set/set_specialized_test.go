package set

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 本文件验证默认 Set 的类型特化存储机制（P1 优化）：
//   - 空集合惰性建表（kindNone，不预分配任何 map）；
//   - 首写探测：string / int / int64 / uint64 进入对应专用 map，其他类型落入通用 map[any]；
//   - 同型序列保持快路径；异型写入触发整体退化到 map[any]（不可逆）；
//   - 异型读/删除不触发退化；
//   - Cleanup 重置为 kindNone，之后首写重新探测特化。

func TestSet_SpecializedStringStorage(t *testing.T) {
	s := NewWithCapacity(16)

	// 空集合：类型为待定，任何底层 map 都未创建。
	assert.Equal(t, kindNone, s.kind)
	assert.Nil(t, s.mAny)
	assert.Nil(t, s.mStr)
	assert.False(t, s.Contains("a"))
	assert.False(t, s.TryRemove("a"))

	// 首写 string：进入 map[string] 快路径。
	s.Add("a")
	assert.Equal(t, kindString, s.kind)
	require.NotNil(t, s.mStr)
	assert.Len(t, s.mStr, 1)
	assert.Nil(t, s.mAny)

	// 同型写入持续走快路径。
	require.True(t, s.TryAdd("b"))
	assert.Equal(t, kindString, s.kind)
	assert.Len(t, s.mStr, 2)
	assert.False(t, s.TryAdd("a"))
	assert.Len(t, s.mStr, 2)

	// List 输出装箱后的元素集合，内容一致。
	assert.ElementsMatch(t, []any{"a", "b"}, s.List())

	// 同型删除。
	assert.True(t, s.TryRemove("a"))
	assert.Len(t, s.mStr, 1)
	assert.Equal(t, 1, s.Len())
}

func TestSet_SpecializedIntStorage(t *testing.T) {
	t.Run("Int", func(t *testing.T) {
		s := New()
		s.Add(1)
		assert.Equal(t, kindInt, s.kind)
		require.NotNil(t, s.mInt)
		assert.Len(t, s.mInt, 1)
		assert.Nil(t, s.mAny)

		require.True(t, s.TryAdd(2))
		assert.False(t, s.TryAdd(1))
		assert.Equal(t, 2, s.Len())
		assert.True(t, s.TryRemove(1))
		assert.Equal(t, 1, s.Len())
		assert.ElementsMatch(t, []any{2}, s.List())
	})

	t.Run("Int64", func(t *testing.T) {
		s := New()
		s.Add(int64(1))
		assert.Equal(t, kindInt64, s.kind)
		require.NotNil(t, s.mInt64)
		assert.Len(t, s.mInt64, 1)
		assert.Nil(t, s.mAny)

		require.True(t, s.TryAdd(int64(2)))
		assert.False(t, s.TryAdd(int64(1)))
		assert.Equal(t, 2, s.Len())
		assert.ElementsMatch(t, []any{int64(1), int64(2)}, s.List())
	})

	t.Run("Uint64", func(t *testing.T) {
		s := New()
		s.Add(uint64(1))
		assert.Equal(t, kindUint64, s.kind)
		require.NotNil(t, s.mUint64)
		assert.Len(t, s.mUint64, 1)
		assert.Nil(t, s.mAny)

		require.True(t, s.TryAdd(uint64(2)))
		assert.False(t, s.TryAdd(uint64(1)))
		assert.Equal(t, 2, s.Len())
		assert.ElementsMatch(t, []any{uint64(1), uint64(2)}, s.List())
	})

	// 不同整型之间绝不共享快路径：首写 int 后再写 int64 触发退化。
	t.Run("DistinctIntKinds", func(t *testing.T) {
		s := New()
		s.Add(1)
		assert.Equal(t, kindInt, s.kind)

		s.Add(int64(1))
		assert.Equal(t, kindAny, s.kind)
		require.NotNil(t, s.mAny)
		assert.Len(t, s.mAny, 2)
		assert.True(t, s.Contains(1))
		assert.True(t, s.Contains(int64(1)))
	})
}

func TestSet_SpecializedOtherTypeGoesGeneric(t *testing.T) {
	type custom struct{ A int }

	for _, first := range []any{custom{A: 1}, true, 1.5, 1i, int8(7), uint(3)} {
		s := New()
		s.Add(first)
		assert.Equal(t, kindAny, s.kind, "first=%#v", first)
		require.NotNil(t, s.mAny, "first=%#v", first)
		assert.Len(t, s.mAny, 1)
		assert.Nil(t, s.mStr)
		assert.Nil(t, s.mInt)
		assert.True(t, s.Contains(first))
	}

	// nil 元素同样落入通用路径。
	s := New()
	s.Add(nil)
	assert.Equal(t, kindAny, s.kind)
	require.NotNil(t, s.mAny)
	assert.True(t, s.Contains(nil))

	// 不可哈希元素与基线 map[any] 行为一致：panic。
	unhashable := []int{1}
	assert.Panics(t, func() { New().Add(unhashable) })
	assert.Panics(t, func() {
		grown := New()
		grown.Add("a")
		grown.Add(unhashable)
	})
}

func TestSet_MixedTypeDemotesToGeneric(t *testing.T) {
	s := New()
	s.Add("a")
	s.Add("b")
	assert.Equal(t, kindString, s.kind)

	// 异型写入：整体迁移到通用 map，内容不丢失。
	s.Add(1)
	assert.Equal(t, kindAny, s.kind)
	require.NotNil(t, s.mAny)
	assert.Len(t, s.mAny, 3)
	assert.Nil(t, s.mStr) // 特化 map 已释放
	assert.True(t, s.Contains("a"))
	assert.True(t, s.Contains("b"))
	assert.True(t, s.Contains(1))
	assert.False(t, s.Contains("c"))

	// 退化后所有操作走通用路径，语义与基线一致。
	assert.True(t, s.TryAdd(nil))
	assert.False(t, s.TryAdd(1))
	assert.Equal(t, kindAny, s.kind)
	assert.Equal(t, 4, s.Len())
	assert.True(t, s.TryRemove("a"))
	assert.Equal(t, 3, s.Len())
	assert.ElementsMatch(t, []any{"b", 1, nil}, s.List())
}

func TestSet_ForeignTypeReadDoesNotDemote(t *testing.T) {
	s := New()
	s.Add("a")
	assert.Equal(t, kindString, s.kind)

	// 异型查询/删除是空操作，绝不触发退化。
	assert.False(t, s.Contains(1))
	s.Remove(1)
	assert.False(t, s.TryRemove(1))
	assert.False(t, s.Contains(nil))

	assert.Equal(t, kindString, s.kind)
	require.NotNil(t, s.mStr)
	assert.Len(t, s.mStr, 1)
	assert.Equal(t, 1, s.Len())
}

func TestSet_CleanupResetsSpecialization(t *testing.T) {
	s := New()
	s.Add("a")
	s.Add(1)
	assert.Equal(t, kindAny, s.kind)

	s.Cleanup()
	assert.Equal(t, kindNone, s.kind)
	assert.Zero(t, s.Len())
	assert.Empty(t, s.List())
	assert.Nil(t, s.mAny)
	assert.Nil(t, s.mStr)
	assert.False(t, s.Contains("a"))

	// Cleanup 后首写重新探测特化类型。
	s.Add(uint64(7))
	assert.Equal(t, kindUint64, s.kind)
	require.NotNil(t, s.mUint64)
	assert.True(t, s.Contains(uint64(7)))
}

// TestSet_NewWithCapacityNegativeTolerated 固化基线行为：Go 运行时对负
// hint 的 make(map, hint) 按 0 处理而不 panic，因此负容量集合仍可正常使用。
func TestSet_NewWithCapacityNegativeTolerated(t *testing.T) {
	assert.NotPanics(t, func() {
		s := NewWithCapacity(-1)
		assert.Zero(t, s.Len())
		assert.True(t, s.TryAdd("a"))
		assert.True(t, s.Contains("a"))
		assert.Equal(t, 1, s.Len())
	})
}
