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

	err = q.Put("test6")
	assert.NoError(t, err, "Put should not return an error")

	err = q.Put("test7")
	assert.NoError(t, err, "Put should not return an error")

	// Verify the queue state
	assert.Equal(t, 7, q.Len(), "Queue length should be 7")
	assert.Equal(t, []interface{}{"test5", "test6", "test7", "test4", "test1", "test2", "test3"}, q.Values(), "Queue values should be [test5 test6 test7 test4 test1 test2 test3]")
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
	assert.Equal(t, "test4", v, "Get value should be test1")

	// Done content from queue
	q.Done(v)

	// Verify the callback
	assert.Equal(t, []interface{}{"test4"}, callback.puts, "Callback puts should be [test4]")
	assert.Equal(t, []interface{}{"test4"}, callback.gets, "Callback gets should be [test4]")
	assert.Equal(t, []interface{}{"test4"}, callback.dones, "Callback dones should be [test4]")
	assert.Equal(t, []interface{}{"test1", "test2", "test3"}, callback.priorities, "Callback priorities should be [test3, test1, test2]")
}

func TestPriorityQueueImpl_Shutdown(t *testing.T) {
	q := NewPriorityQueue(nil)
	q.Shutdown()

	// Verify that the queue is closed
	assert.True(t, q.IsClosed(), "Queue should be closed")
	assert.Equal(t, 0, q.Len(), "Queue length should be 0")
}
