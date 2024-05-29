package set

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet_Add(t *testing.T) {
	t.Run("Integer", func(t *testing.T) {
		// Create a new set
		s := New()
		defer s.Clear()

		// Add integers to the set
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
		// Create a new set
		s := New()
		defer s.Clear()

		// Add floats to the set
		s.Add(float32(1.1))
		assert.True(t, s.Contains(float32(1.1)))
		s.Add(float64(2.2))
		assert.True(t, s.Contains(float64(2.2)))
	})

	t.Run("String", func(t *testing.T) {
		// Create a new set
		s := New()
		defer s.Clear()

		// Add strings to the set
		s.Add("a")
		assert.True(t, s.Contains("a"))
		s.Add("b")
		assert.True(t, s.Contains("b"))
		s.Add("!@#$%^&*()_+:',./<>?;[]{}")
		assert.True(t, s.Contains("!@#$%^&*()_+:',./<>?;[]{}"))
	})

	t.Run("Boolean", func(t *testing.T) {
		// Create a new set
		s := New()
		defer s.Clear()

		// Add booleans to the set
		s.Add(true)
		assert.True(t, s.Contains(true))
		s.Add(false)
		assert.True(t, s.Contains(false))
	})

	t.Run("Complex", func(t *testing.T) {
		// Create a new set
		s := New()
		defer s.Clear()

		// Add complex numbers to the set
		s.Add(complex(1, 1))
		assert.True(t, s.Contains(complex(1, 1)))
	})

	t.Run("Nil", func(t *testing.T) {
		// Create a new set
		s := New()
		defer s.Clear()

		// Add nil to the set
		s.Add(nil)
		assert.True(t, s.Contains(nil))
	})

	t.Run("Struct", func(t *testing.T) {
		// Create a new set
		s := New()
		defer s.Clear()

		// Add struct to the set
		type testStruct struct {
			A int
			B string
		}

		s.Add(testStruct{A: 1, B: "a"})
		assert.True(t, s.Contains(testStruct{A: 1, B: "a"}))

		p := &testStruct{A: 2, B: "b"}
		s.Add(p)
		assert.True(t, s.Contains(p))

		// Add struct with nocopy
		type testStructNoCopy struct {
			lock sync.Mutex
		}

		s.Add(testStructNoCopy{lock: sync.Mutex{}})
		assert.True(t, s.Contains(testStructNoCopy{lock: sync.Mutex{}}))

		// Add struct with mixed fields
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
		// Create a new set
		s := New()
		defer s.Clear()

		// Add integers to the set
		s.Add(1)
		s.Add(2)
		s.Add(3)

		// Remove an integer from the set
		s.Remove(2)

		// Assert that the removed integer is no longer in the set
		assert.False(t, s.Contains(2))
	})

	t.Run("String", func(t *testing.T) {
		// Create a new set
		s := New()
		defer s.Clear()

		// Add strings to the set
		s.Add("a")
		s.Add("b")
		s.Add("c")

		// Remove a string from the set
		s.Remove("b")

		// Assert that the removed string is no longer in the set
		assert.False(t, s.Contains("b"))
	})

	t.Run("Struct", func(t *testing.T) {
		// Create a new set
		s := New()
		defer s.Clear()

		// Add struct to the set
		type testStruct struct {
			A int
			B string
		}

		s.Add(testStruct{A: 1, B: "a"})
		s.Add(testStruct{A: 2, B: "b"})
		s.Add(testStruct{A: 3, B: "c"})

		// Remove a struct from the set
		s.Remove(testStruct{A: 2, B: "b"})

		// Assert that the removed struct is no longer in the set
		assert.False(t, s.Contains(testStruct{A: 2, B: "b"}))
	})

	t.Run("Nil", func(t *testing.T) {
		// Create a new set
		s := New()
		defer s.Clear()

		// Add nil to the set
		s.Add(nil)

		// Remove nil from the set
		s.Remove(nil)

		// Assert that nil is no longer in the set
		assert.False(t, s.Contains(nil))
	})

	t.Run("Complex", func(t *testing.T) {
		// Create a new set
		s := New()
		defer s.Clear()

		// Add complex numbers to the set
		s.Add(complex(1, 1))
		s.Add(complex(2, 2))
		s.Add(complex(3, 3))

		// Remove a complex number from the set
		s.Remove(complex(2, 2))

		// Assert that the removed complex number is no longer in the set
		assert.False(t, s.Contains(complex(2, 2)))
	})
}

func TestSet_List(t *testing.T) {
	t.Run("EmptySet", func(t *testing.T) {
		// Create a new set
		s := New()
		defer s.Clear()

		// Get the list of items from the set
		list := s.List()

		// Assert that the list is empty
		assert.Empty(t, list)
	})

	t.Run("NonEmptySet", func(t *testing.T) {
		// Create a new set
		s := New()
		defer s.Clear()

		// Add items to the set
		s.Add(1)
		s.Add("a")
		s.Add(true)

		// Get the list of items from the set
		list := s.List()

		// Assert that the list contains the added items
		assert.ElementsMatch(t, []interface{}{1, "a", true}, list)
	})
}
