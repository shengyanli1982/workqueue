package set

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet_Add(t *testing.T) {
	t.Run("Integer", func(t *testing.T) {

		s := New()
		defer s.Cleanup()

		s.Add(-1)
		assert.True(t, s.Contains(-1))
		s.Add(int8(-2))
		assert.True(t, s.Contains(int8(-2)))
		s.Add(int16(-3))
		assert.True(t, s.Contains(int16(-3)))
		s.Add(int32(-4))
		assert.True(t, s.Contains(int32(-4)))
		s.Add(int64(-5))
		assert.True(t, s.Contains(int64(-5)))
		s.Add(uint(1))
		assert.True(t, s.Contains(uint(1)))
		s.Add(uint8(2))
		assert.True(t, s.Contains(uint8(2)))
		s.Add(uint16(3))
		assert.True(t, s.Contains(uint16(3)))
		s.Add(uint32(4))
		assert.True(t, s.Contains(uint32(4)))
		s.Add(uint64(5))
		assert.True(t, s.Contains(uint64(5)))
	})

	t.Run("Float", func(t *testing.T) {

		s := New()
		defer s.Cleanup()

		s.Add(float32(1.1))
		assert.True(t, s.Contains(float32(1.1)))
		s.Add(float64(2.2))
		assert.True(t, s.Contains(float64(2.2)))
	})

	t.Run("String", func(t *testing.T) {

		s := New()
		defer s.Cleanup()

		s.Add("a")
		assert.True(t, s.Contains("a"))
		s.Add("b")
		assert.True(t, s.Contains("b"))
		s.Add("!@#$%^&*()_+:',./<>?;[]{}")
		assert.True(t, s.Contains("!@#$%^&*()_+:',./<>?;[]{}"))
	})

	t.Run("Boolean", func(t *testing.T) {

		s := New()
		defer s.Cleanup()

		s.Add(true)
		assert.True(t, s.Contains(true))
		s.Add(false)
		assert.True(t, s.Contains(false))
	})

	t.Run("Complex", func(t *testing.T) {

		s := New()
		defer s.Cleanup()

		s.Add(complex(1, 1))
		assert.True(t, s.Contains(complex(1, 1)))
	})

	t.Run("Nil", func(t *testing.T) {

		s := New()
		defer s.Cleanup()

		s.Add(nil)
		assert.True(t, s.Contains(nil))
	})

	t.Run("Struct", func(t *testing.T) {

		s := New()
		defer s.Cleanup()

		type testStruct struct {
			A int
			B string
		}

		s.Add(testStruct{A: 1, B: "a"})
		assert.True(t, s.Contains(testStruct{A: 1, B: "a"}))

		p := &testStruct{A: 2, B: "b"}
		s.Add(p)
		assert.True(t, s.Contains(p))

		type testStructNoCopy struct {
			lock sync.Mutex
		}

		s.Add(testStructNoCopy{lock: sync.Mutex{}})
		assert.True(t, s.Contains(testStructNoCopy{lock: sync.Mutex{}}))

		type testStructMixed struct {
			a    int
			b    string
			A    int
			B    string
			lock sync.Mutex
		}

		s.Add(testStructMixed{a: 1, b: "a", A: 2, B: "b", lock: sync.Mutex{}})
		assert.True(t, s.Contains(testStructMixed{a: 1, b: "a", A: 2, B: "b", lock: sync.Mutex{}}))
	})
}

func TestSet_Remove(t *testing.T) {
	t.Run("Integer", func(t *testing.T) {

		s := New()
		defer s.Cleanup()

		s.Add(1)
		s.Add(2)
		s.Add(3)

		s.Remove(2)

		assert.False(t, s.Contains(2))
	})

	t.Run("String", func(t *testing.T) {

		s := New()
		defer s.Cleanup()

		s.Add("a")
		s.Add("b")
		s.Add("c")

		s.Remove("b")

		assert.False(t, s.Contains("b"))
	})

	t.Run("Struct", func(t *testing.T) {

		s := New()
		defer s.Cleanup()

		type testStruct struct {
			A int
			B string
		}

		s.Add(testStruct{A: 1, B: "a"})
		s.Add(testStruct{A: 2, B: "b"})
		s.Add(testStruct{A: 3, B: "c"})

		s.Remove(testStruct{A: 2, B: "b"})

		assert.False(t, s.Contains(testStruct{A: 2, B: "b"}))
	})

	t.Run("Nil", func(t *testing.T) {

		s := New()
		defer s.Cleanup()

		s.Add(nil)

		s.Remove(nil)

		assert.False(t, s.Contains(nil))
	})

	t.Run("Complex", func(t *testing.T) {

		s := New()
		defer s.Cleanup()

		s.Add(complex(1, 1))
		s.Add(complex(2, 2))
		s.Add(complex(3, 3))

		s.Remove(complex(2, 2))

		assert.False(t, s.Contains(complex(2, 2)))
	})
}

func TestSet_List(t *testing.T) {
	t.Run("EmptySet", func(t *testing.T) {

		s := New()
		defer s.Cleanup()

		list := s.List()

		assert.Empty(t, list)
	})

	t.Run("NonEmptySet", func(t *testing.T) {

		s := New()
		defer s.Cleanup()

		s.Add(1)
		s.Add("a")
		s.Add(true)

		list := s.List()

		assert.ElementsMatch(t, []interface{}{1, "a", true}, list)
	})
}
