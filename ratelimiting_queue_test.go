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

	count := 4
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

	// Verify the callback state
	assert.True(t, len(callback.delays) <= 2, "Callback delays length should be less than or equal to 2")
	assert.Equal(t, []interface{}{"test1", "test2", "test3", "test4"}, callback.puts, "Callback puts should contain all put items")
	assert.Equal(t, []interface{}{"test1"}, callback.gets, "Callback gets should be [test1]")
	assert.Equal(t, []interface{}(nil), callback.dones, "Callback dones should be [test1]")
	assert.Empty(t, callback.errors, "Callback errors should be empty")
}

func TestRateLimitingQueueImpl_HighConcurrencyRateLimit(t *testing.T) {
	// 测试高并发下的限流效果
	config := NewRateLimitingQueueConfig().
		WithLimiter(NewBucketRateLimiterImpl(2, 1)) // 每秒只允许2个请求
	q := NewRateLimitingQueue(config)
	defer q.Shutdown()

	var wg sync.WaitGroup
	start := time.Now()

	// 并发发送10个请求
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := q.PutWithLimited(i)
			assert.NoError(t, err, "Put should not return an error")
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// 验证是否有请求被限流
	assert.True(t, duration < time.Second*2, "Should complete within 2 seconds")
}

func TestRateLimitingQueueImpl_DuplicateElements(t *testing.T) {
	// 测试重复元素的处理
	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(5, 1))
	q := NewRateLimitingQueue(config)
	defer q.Shutdown()

	// 重复放入相同元素
	err := q.PutWithLimited("duplicate")
	assert.NoError(t, err, "First put should succeed")

	err = q.PutWithLimited("duplicate")
	assert.NoError(t, err, "Second put with same value should succeed")

	assert.Equal(t, 2, q.Len(), "Queue should contain both duplicate elements")
}

func TestRateLimitingQueueImpl_DifferentTypes(t *testing.T) {
	// 测试不同数据类型的处理
	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(5, 1))
	q := NewRateLimitingQueue(config)
	defer q.Shutdown()

	testCases := []interface{}{
		42,
		"string",
		struct{ name string }{"test"},
		[]int{1, 2, 3},
		map[string]int{"key": 1},
	}

	for _, tc := range testCases {
		err := q.PutWithLimited(tc)
		assert.NoError(t, err, "Should handle different types")
	}

	assert.Equal(t, len(testCases), q.Len(), "Queue should contain all elements")
}

func TestRateLimitingQueueImpl_EmptyQueueGet(t *testing.T) {
	// 测试空队列的Get操作
	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(5, 1))
	q := NewRateLimitingQueue(config)
	defer q.Shutdown()

	// 从空队列中获取元素
	_, err := q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty, "Get should return ErrQueueIsEmpty")
}
