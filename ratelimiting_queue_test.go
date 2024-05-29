package workqueue

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimitingQueueImpl_PutWithLimited(t *testing.T) {
	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(5, 1))
	q := NewRateLimitingQueue(config)
	defer q.Shutdown()

	// Put content into queue
	err := q.PutWithLimited("test1")
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithLimited("test2")
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithLimited("test3")
	assert.NoError(t, err, "Put should not return an error")

	time.Sleep(time.Second)

	// Verify the queue state
	assert.Equal(t, 3, q.Len(), "Queue length should be 3")
	assert.Equal(t, []interface{}{"test1", "test2", "test3"}, q.Values(), "Queue values should be [test1, test2, test3]")
}

func TestRateLimitingQueueImpl_PutWithLimited_Closed(t *testing.T) {
	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(5, 1))
	q := NewRateLimitingQueue(config)
	q.Shutdown()

	// Put nil content into queue
	err := q.PutWithLimited("test")
	assert.ErrorIs(t, err, ErrQueueIsClosed, "Put should return ErrQueueIsClosed")

	time.Sleep(time.Second)
}

func TestRateLimitingQueueImpl_PutWithLimited_Nil(t *testing.T) {
	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(5, 1))
	q := NewRateLimitingQueue(config)
	defer q.Shutdown()

	// Put nil content into queue
	err := q.PutWithLimited(nil)
	assert.ErrorIs(t, err, ErrElementIsNil, "Put should return ErrElementIsNil")

	time.Sleep(time.Second)
}

func TestRateLimitingQueueImpl_PutWithLimited_Parallel(t *testing.T) {
	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(5, 1))
	q := NewRateLimitingQueue(config)
	defer q.Shutdown()

	count := 1000
	var wg sync.WaitGroup
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()

			err := q.PutWithLimited(i)
			assert.NoError(t, err, "Put should not return an error")
		}(i)
	}

	wg.Wait()

	time.Sleep(time.Second)

	// Verify the queue state
	assert.Equal(t, count, q.Len(), "Queue length should be 1000")
}

type testRateLimitingQueueCallback struct {
	puts, gets, dones, delays, errors, limits []interface{}
}

func (c *testRateLimitingQueueCallback) OnPut(value interface{}) {
	c.puts = append(c.puts, value)
}

func (c *testRateLimitingQueueCallback) OnGet(value interface{}) {
	c.gets = append(c.gets, value)
}

func (c *testRateLimitingQueueCallback) OnDone(value interface{}) {
	c.dones = append(c.dones, value)
}

func (c *testRateLimitingQueueCallback) OnDelay(value interface{}, delay int64) {
	c.delays = append(c.delays, value)
}

func (c *testRateLimitingQueueCallback) OnPullError(value interface{}, err error) {
	c.errors = append(c.errors, value)
}

func (c *testRateLimitingQueueCallback) OnLimited(value interface{}) {
	c.limits = append(c.limits, value)
}

func TestRateLimitingQueueImpl_Callback(t *testing.T) {
	callback := &testRateLimitingQueueCallback{}
	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(5, 1)).WithCallback(callback)
	q := NewRateLimitingQueue(config)
	defer q.Shutdown()

	// Put content into queue
	err := q.PutWithLimited("test1")
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithLimited("test2")
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithLimited("test3")
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

	// Verify the callback state, test1 delay is 0, so it should not be in the delays
	assert.Equal(t, []interface{}{"test2", "test3"}, callback.delays, "Callback puts should be [test2, test3]")
	assert.Equal(t, []interface{}{"test1", "test2", "test3", "test4"}, callback.puts, "Callback puts should be [test1, test2, test3, test4]")
	assert.Equal(t, []interface{}{"test1"}, callback.gets, "Callback gets should be [test1]")
	assert.Equal(t, []interface{}{"test1"}, callback.dones, "Callback dones should be [test1]")
	assert.Equal(t, []interface{}(nil), callback.errors, "Callback errors should be []")
}
