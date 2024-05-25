package stack

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCasStack_Push(t *testing.T) {
	s := New()

	var wg sync.WaitGroup
	numPushes := 1000

	wg.Add(numPushes)
	for i := 0; i < numPushes; i++ {
		go func(i int) {
			defer wg.Done()
			s.Push(int64(i))
		}(i)
	}

	wg.Wait()

	assert.Equal(t, int64(numPushes), s.Len(), "stack length should be 1000")
}

func TestCasStack_Pop(t *testing.T) {
	s := New()

	var wg sync.WaitGroup
	numPushes := 1000

	wg.Add(numPushes)
	for i := 0; i < numPushes; i++ {
		go func(i int) {
			defer wg.Done()
			s.Push(int64(i))
		}(i)
	}

	wg.Wait()

	assert.Equal(t, int64(numPushes), s.Len(), "stack length should be 1000")

	wg.Add(numPushes)
	for i := 0; i < numPushes; i++ {
		go func() {
			defer wg.Done()
			s.Pop()
		}()
	}

	wg.Wait()

	assert.Equal(t, int64(0), s.Len(), "stack length should be 0")
}

func TestCasStack_Cleanup(t *testing.T) {
	s := New()

	// Push some elements to the stack
	s.Push(1)
	s.Push(2)
	s.Push(3)

	// Verify that the stack is not empty
	assert.Equal(t, int64(3), s.Len(), "stack length should be 3")

	// Call the Cleanup method
	s.Cleanup()

	// Verify that the stack is empty
	assert.Equal(t, int64(0), s.Len(), "stack length should be 0")
}
