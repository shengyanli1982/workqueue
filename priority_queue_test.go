package workqueue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPriorityQueueImpl_PutWithPriority(t *testing.T) {
	q := NewPriorityQueue(nil)
	defer q.Shutdown()

	// Put content into queue
	err := q.PutWithPriority("test1", 1)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithPriority("test2", 2)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithPriority("test3", 3)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithPriority("test4", 0)
	assert.NoError(t, err, "Put should not return an error")

	time.Sleep(time.Second)

	err = q.Put("test5")
	assert.NoError(t, err, "Put should not return an error")

	// Verify the queue state
	assert.Equal(t, 5, q.Len(), "Queue length should be 5")
	assert.Equal(t, []interface{}{"test4", "test5", "test1", "test2", "test3"}, q.Values(), "Queue values should be [test4 test5 test1 test2 test3]")
}

func TestPriorityQueueImpl_PutWithPriority_Closed(t *testing.T) {
	q := NewPriorityQueue(nil)
	q.Shutdown()

	// Put content into queue
	err := q.PutWithPriority("test1", 1)
	assert.ErrorIs(t, err, ErrQueueIsClosed, "Put should return ErrQueueIsClosed")

	time.Sleep(time.Second)
}

func TestPriorityQueueImpl_PutWithPriority_Nil(t *testing.T) {
	q := NewPriorityQueue(nil)
	defer q.Shutdown()

	// Put nil content into queue
	err := q.PutWithPriority(nil, 0)
	assert.ErrorIs(t, err, ErrElementIsNil, "Put should return ErrElementIsNil")

	time.Sleep(time.Second)
}

func TestPriorityQueueImpl_PutWithPriority_Parallel(t *testing.T) {
	q := NewPriorityQueue(nil)
	defer q.Shutdown()

	count := 1000
	for i := 0; i < count; i++ {
		go func(i int) {
			err := q.PutWithPriority(i, int64(i))
			assert.NoError(t, err, "Put should not return an error")
		}(i)
	}

	time.Sleep(time.Second)

	// Verify the queue state
	assert.Equal(t, count, q.Len(), "Queue length should be 1000")
}

type testPriorityQueueCallback struct {
	puts, gets, dones, priorities []interface{}
}

func (c *testPriorityQueueCallback) OnPut(value interface{}) {
	c.puts = append(c.puts, value)
}

func (c *testPriorityQueueCallback) OnGet(value interface{}) {
	c.gets = append(c.gets, value)
}

func (c *testPriorityQueueCallback) OnDone(value interface{}) {
	c.dones = append(c.dones, value)
}

func (c *testPriorityQueueCallback) OnPriority(value interface{}, priority int64) {
	c.priorities = append(c.priorities, value)
}

func TestPriorityQueueImpl_Callback(t *testing.T) {
	callback := &testPriorityQueueCallback{}
	config := NewPriorityQueueConfig().WithCallback(callback)
	q := NewPriorityQueue(config)
	defer q.Shutdown()

	// Put content into queue
	err := q.PutWithPriority("test1", 1)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithPriority("test2", 2)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithPriority("test3", 0)
	assert.NoError(t, err, "Put should not return an error")

	time.Sleep(time.Second)

	err = q.Put("test4")
	assert.NoError(t, err, "Put should not return an error")

	// Get content from queue
	v, err := q.Get()
	assert.NoError(t, err, "Get should not return an error")
	assert.Equal(t, "test3", v, "Get value should be test3")

	// Done content from queue
	q.Done(v)

	// Verify the callback
	assert.Nil(t, callback.puts, "Callback puts should be nil")
	assert.Equal(t, []interface{}{"test3"}, callback.gets, "Callback gets should be [test3]")
	assert.Equal(t, []interface{}(nil), callback.dones, "Callback dones should be [test3]")
	assert.Equal(t, []interface{}{"test1", "test2", "test3", "test4"}, callback.priorities, "Callback priorities should be [test1 test2 test3 test4]")
}

func TestPriorityQueueImpl_Shutdown(t *testing.T) {
	q := NewPriorityQueue(nil)
	q.Shutdown()

	// Verify that the queue is closed
	assert.True(t, q.IsClosed(), "Queue should be closed")
	assert.Equal(t, 0, q.Len(), "Queue length should be 0")
}

func TestPriorityQueueImpl_HeapRange(t *testing.T) {
	q := NewPriorityQueue(nil)
	defer q.Shutdown()

	// Put content into queue
	err := q.PutWithPriority("test1", 0)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithPriority("test2", 1)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithPriority("test3", 2)
	assert.NoError(t, err, "Put should not return an error")

	// Range content from queue
	values := []interface{}{}
	q.HeapRange(func(value interface{}, _ int64) bool {
		values = append(values, value)
		return true
	})

	time.Sleep(time.Second)

	// Verify the queue state
	assert.Equal(t, []interface{}{"test1", "test2", "test3"}, values, "Queue values should be [test1, test2, test3]")
}

func TestPriorityQueueImpl_HeapRange_Closed(t *testing.T) {
	q := NewPriorityQueue(nil)

	// Put content into queue
	err := q.PutWithPriority("test1", 0)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithPriority("test2", 1)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithPriority("test3", 2)
	assert.NoError(t, err, "Put should not return an error")

	q.Shutdown()

	// Range content from queue
	values := []interface{}{}
	q.HeapRange(func(value interface{}, _ int64) bool {
		values = append(values, value)
		return true
	})

	assert.Equal(t, []interface{}{}, values, "Values should be []")
}

func TestPriorityQueueImpl_NegativePriority(t *testing.T) {
	q := NewPriorityQueue(nil)
	defer q.Shutdown()

	// Test negative priorities
	err := q.PutWithPriority("test1", -1)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithPriority("test2", -100)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithPriority("test3", 0)
	assert.NoError(t, err, "Put should not return an error")

	// Verify order (smaller numbers have higher priority)
	v1, _ := q.Get()
	v2, _ := q.Get()
	v3, _ := q.Get()

	assert.Equal(t, "test2", v1, "First item should be test2 (priority -100)")
	assert.Equal(t, "test1", v2, "Second item should be test1 (priority -1)")
	assert.Equal(t, "test3", v3, "Third item should be test3 (priority 0)")
}

func TestPriorityQueueImpl_SamePriority(t *testing.T) {
	q := NewPriorityQueue(nil)
	defer q.Shutdown()

	// Put multiple items with same priority
	err := q.PutWithPriority("test1", 1)
	assert.NoError(t, err)
	err = q.PutWithPriority("test2", 1)
	assert.NoError(t, err)
	err = q.PutWithPriority("test3", 1)
	assert.NoError(t, err)

	// Items with same priority should maintain FIFO order
	v1, _ := q.Get()
	v2, _ := q.Get()
	v3, _ := q.Get()

	assert.Equal(t, "test1", v1, "First item should be test1")
	assert.Equal(t, "test2", v2, "Second item should be test2")
	assert.Equal(t, "test3", v3, "Third item should be test3")
}

func TestPriorityQueueImpl_PriorityBoundaries(t *testing.T) {
	q := NewPriorityQueue(nil)
	defer q.Shutdown()

	// Test with max and min int64 values
	err := q.PutWithPriority("max", int64(^uint64(0)>>1)) // MaxInt64
	assert.NoError(t, err)
	err = q.PutWithPriority("min", -int64(^uint64(0)>>1)-1) // MinInt64
	assert.NoError(t, err)

	v1, _ := q.Get()
	v2, _ := q.Get()

	assert.Equal(t, "min", v1, "First item should be min (lowest priority)")
	assert.Equal(t, "max", v2, "Second item should be max (highest priority)")
}

func TestPriorityQueueImpl_EmptyQueueGet(t *testing.T) {
	q := NewPriorityQueue(nil)
	defer q.Shutdown()

	// Try to get from empty queue
	v, err := q.Get()
	assert.Nil(t, v, "Value should be nil for empty queue")
	assert.ErrorIs(t, err, ErrQueueIsEmpty, "Get should return ErrQueueIsEmpty")
	// Try to call Done on nil value
	q.Done(nil)
}

func TestPriorityQueueImpl_LargeNumberOfItems(t *testing.T) {
	q := NewPriorityQueue(nil)
	defer q.Shutdown()

	// Add a large number of items
	itemCount := 10000
	for i := 0; i < itemCount; i++ {
		err := q.PutWithPriority(i, int64(i%100)) // Use modulo to create priority groups
		assert.NoError(t, err)
	}

	assert.Equal(t, itemCount, q.Len(), "Queue length should match number of items added")

	// Verify items can be retrieved
	previousPriority := int64(-1)
	for i := 0; i < itemCount; i++ {
		v, err := q.Get()
		assert.NoError(t, err)
		assert.NotNil(t, v)

		// Verify priority ordering
		currentPriority := int64(v.(int) % 100)
		assert.True(t, currentPriority >= previousPriority, "Items should be in priority order")
		previousPriority = currentPriority
	}
}
