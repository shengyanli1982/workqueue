package workqueue

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var DELAYDUCRATION = int64(150)

func TestDelayingQueueImpl_PutWithDelay(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	// Put content into queue
	err := q.PutWithDelay("test1", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test2", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test3", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	time.Sleep(time.Second)

	// Verify the queue state
	assert.Equal(t, 3, q.Len(), "Queue length should be 3")
	assert.Equal(t, []interface{}{"test1", "test2", "test3"}, q.Values(), "Queue values should be [test1, test2, test3]")
}

func TestDelayingQueueImpl_PutWithDelay_Closed(t *testing.T) {
	q := NewDelayingQueue(nil)
	q.Shutdown()

	// Put nil content into queue
	err := q.PutWithDelay("test", 0)
	assert.ErrorIs(t, err, ErrQueueIsClosed, "Put should return ErrQueueIsClosed")

	time.Sleep(time.Second)
}

func TestDelayingQueueImpl_PutWithDelay_Nil(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	// Put nil content into queue
	err := q.PutWithDelay(nil, 0)
	assert.ErrorIs(t, err, ErrElementIsNil, "Put should return ErrElementIsNil")

	time.Sleep(time.Second)
}

func TestDelayingQueueImpl_PutWithDelay_Parallel(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	count := 1000

	// Put content into queue in parallel
	wg := sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			err := q.PutWithDelay("test", DELAYDUCRATION)
			assert.NoError(t, err, "Put should not return an error")
		}()
	}
	wg.Wait()

	time.Sleep(time.Second)

	// Verify the queue state
	assert.Equal(t, count, q.Len(), "Queue length should be 1000")
}

func TestDelayingQueueImpl_HeapRange(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	// Put content into queue
	err := q.PutWithDelay("test1", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test2", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test3", DELAYDUCRATION)
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

func TestDelayingQueueImpl_HeapRange_Closed(t *testing.T) {
	q := NewDelayingQueue(nil)

	// Put content into queue
	err := q.PutWithDelay("test1", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test2", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test3", DELAYDUCRATION)
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

type testDelayingQueueCallback struct {
	puts, gets, dones, delays, errors []interface{}
}

func (c *testDelayingQueueCallback) OnPut(value interface{}) {
	c.puts = append(c.puts, value)
}

func (c *testDelayingQueueCallback) OnGet(value interface{}) {
	c.gets = append(c.gets, value)
}

func (c *testDelayingQueueCallback) OnDone(value interface{}) {
	c.dones = append(c.dones, value)
}

func (c *testDelayingQueueCallback) OnDelay(value interface{}, delay int64) {
	c.delays = append(c.delays, value)
}

func (c *testDelayingQueueCallback) OnPullError(value interface{}, err error) {
	c.errors = append(c.errors, value)
}

func TestDelayingQueueImpl_Callback(t *testing.T) {
	callback := &testDelayingQueueCallback{}
	config := NewDelayingQueueConfig().WithCallback(callback)
	q := NewDelayingQueue(config)
	defer q.Shutdown()

	// Put content into queue
	err := q.PutWithDelay("test1", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test2", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test3", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	time.Sleep(time.Second)

	err = q.Put("test4")
	assert.NoError(t, err, "Put should not return an error")

	// Get content from queue
	v, err := q.Get()
	assert.NoError(t, err, "Get should not return an error")
	assert.Equal(t, "test1", v, "Get value should be test1")

	// Done content from queue
	q.Done(v)

	// Verify the callback state
	assert.Equal(t, []interface{}{"test1", "test2", "test3"}, callback.delays, "Callback puts should be [test1, test2, test3]")
	assert.Equal(t, []interface{}{"test1", "test2", "test3", "test4"}, callback.puts, "Callback puts should be [test1, test2, test3, test4]")
	assert.Equal(t, []interface{}{"test1"}, callback.gets, "Callback gets should be [test1]")
	assert.Equal(t, []interface{}(nil), callback.dones, "Callback dones should be [test1]")
	assert.Equal(t, []interface{}(nil), callback.errors, "Callback errors should be []")
}

type testAccNode struct {
	value interface{}
	ts    int64
}

func TestDelayingQueueImpl_Accuracy(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	// Put content into queue
	err := q.PutWithDelay(&testAccNode{value: "test1", ts: time.Now().UnixMilli()}, DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay(&testAccNode{value: "test2", ts: time.Now().UnixMilli()}, DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay(&testAccNode{value: "test3", ts: time.Now().UnixMilli()}, DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	time.Sleep(time.Second)

	// Verify the queue state
	assert.Equal(t, 3, q.Len(), "Queue length should be 3")

	values := q.Values()
	for i, v := range values {
		node := v.(*testAccNode)
		assert.Equal(t, fmt.Sprintf("test%d", i+1), node.value, fmt.Sprintf("Value should be test%d", i+1))
		assert.True(t, time.Now().UnixMilli()-node.ts > DELAYDUCRATION, "Delay duration should be greater than 150ms")
	}
}

func TestDelayingQueueImpl_ExtremeDelays(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	// Test zero delay
	err := q.PutWithDelay("zero-delay", 0)
	assert.NoError(t, err, "Put with zero delay should not return an error")

	// Test very long delay
	err = q.PutWithDelay("long-delay", 24*60*60*1000) // 24 hours
	assert.NoError(t, err, "Put with long delay should not return an error")

	// Test very short delay
	time.Sleep(time.Second)

	// Verify immediate availability of zero-delay item
	v, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "zero-delay", v, "Zero delay item should be available immediately")
}

func TestDelayingQueueImpl_NegativeDelay(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	// Test negative delay
	err := q.PutWithDelay("negative-delay", -100)
	assert.NoError(t, err, "Put with negative delay should not return an error")

	// Test very short delay
	time.Sleep(time.Second)

	// Verify immediate availability
	v, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "negative-delay", v, "Negative delay item should be available immediately")
}

func TestDelayingQueueImpl_DuplicateItems(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	// Put same item multiple times with different delays
	err := q.PutWithDelay("duplicate", 100)
	assert.NoError(t, err)

	err = q.PutWithDelay("duplicate", 50)
	assert.NoError(t, err)

	err = q.PutWithDelay("duplicate", 150)
	assert.NoError(t, err)

	// Test very short delay
	time.Sleep(time.Second)

	// All items should be available
	assert.Equal(t, 3, q.Len(), "Queue should contain all duplicate items")
}

func TestDelayingQueueImpl_ConcurrentShutdown(t *testing.T) {
	q := NewDelayingQueue(nil)

	var wg sync.WaitGroup
	wg.Add(100)

	// Start goroutines that try to put items
	for i := 0; i < 100; i++ {
		go func(index int) {
			defer wg.Done()
			_ = q.PutWithDelay(fmt.Sprintf("item-%d", index), DELAYDUCRATION)
		}(i)
	}

	// Shutdown while items are being added
	time.Sleep(10 * time.Millisecond)
	q.Shutdown()

	wg.Wait()

	// Try to add item after shutdown
	err := q.PutWithDelay("after-shutdown", DELAYDUCRATION)
	assert.ErrorIs(t, err, ErrQueueIsClosed, "Put after shutdown should return ErrQueueIsClosed")
}
