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

	err := q.PutWithLimited("test1")
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithLimited("test2")
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithLimited("test3")
	assert.NoError(t, err, "Put should not return an error")

	time.Sleep(time.Second)

	assert.Equal(t, 3, q.Len(), "Queue length should be 3")
	assert.Equal(t, []any{"test1", "test2", "test3"}, q.Values(), "Queue values should be [test1, test2, test3]")
}

func TestRateLimitingQueueImpl_PutWithLimited_Closed(t *testing.T) {
	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(5, 1))
	q := NewRateLimitingQueue(config)
	q.Shutdown()

	err := q.PutWithLimited("test")
	assert.ErrorIs(t, err, ErrQueueIsClosed, "Put should return ErrQueueIsClosed")

	time.Sleep(time.Second)
}

func TestRateLimitingQueueImpl_PutWithLimited_Nil(t *testing.T) {
	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(5, 1))
	q := NewRateLimitingQueue(config)
	defer q.Shutdown()

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

	assert.Equal(t, count, q.Len(), "Queue length should be 1000")
}

type testRateLimitingQueueCallback struct {
	sync.Mutex
	puts, gets, dones, delays, errors, limits []any
}

func (c *testRateLimitingQueueCallback) OnPut(value any) {
	c.Lock()
	defer c.Unlock()
	c.puts = append(c.puts, value)
}

func (c *testRateLimitingQueueCallback) OnGet(value any) {
	c.Lock()
	defer c.Unlock()
	c.gets = append(c.gets, value)
}

func (c *testRateLimitingQueueCallback) OnDone(value any) {
	c.Lock()
	defer c.Unlock()
	c.dones = append(c.dones, value)
}

func (c *testRateLimitingQueueCallback) OnDelay(value any, delay int64) {
	c.Lock()
	defer c.Unlock()
	c.delays = append(c.delays, value)
}

func (c *testRateLimitingQueueCallback) OnPullError(value any, err error) {
	c.Lock()
	defer c.Unlock()
	c.errors = append(c.errors, value)
}

func (c *testRateLimitingQueueCallback) OnLimited(value any) {
	c.Lock()
	defer c.Unlock()
	c.limits = append(c.limits, value)
}

func TestRateLimitingQueueImpl_Callback(t *testing.T) {
	callback := &testRateLimitingQueueCallback{}
	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(5, 1)).WithCallback(callback)
	q := NewRateLimitingQueue(config)
	defer q.Shutdown()

	err := q.PutWithLimited("test1")
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithLimited("test2")
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithLimited("test3")
	assert.NoError(t, err, "Put should not return an error")

	time.Sleep(time.Second)

	err = q.Put("test4")
	assert.NoError(t, err, "Put should not return an error")

	v, err := q.Get()
	assert.NoError(t, err, "Get should not return an error")
	assert.Equal(t, "test1", v, "Get value should be test1")

	q.Done(v)

	callback.Lock()
	delaysLen := len(callback.delays)
	puts := callback.puts
	gets := callback.gets
	dones := callback.dones
	errors := callback.errors
	callback.Unlock()

	assert.True(t, delaysLen <= 2, "Callback delays length should be less than or equal to 2")
	assert.Equal(t, []any{"test1", "test2", "test3", "test4"}, puts, "Callback puts should contain all put items")
	assert.Equal(t, []any{"test1"}, gets, "Callback gets should be [test1]")
	assert.Equal(t, []any(nil), dones, "Callback dones should be [test1]")
	assert.Empty(t, errors, "Callback errors should be empty")
}

func TestRateLimitingQueueImpl_HighConcurrencyRateLimit(t *testing.T) {

	config := NewRateLimitingQueueConfig().
		WithLimiter(NewBucketRateLimiterImpl(2, 1))
	q := NewRateLimitingQueue(config)
	defer q.Shutdown()

	var wg sync.WaitGroup
	start := time.Now()

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

	assert.True(t, duration < time.Second*2, "Should complete within 2 seconds")
}

func TestRateLimitingQueueImpl_DuplicateElements(t *testing.T) {

	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(5, 1))
	q := NewRateLimitingQueue(config)
	defer q.Shutdown()

	err := q.PutWithLimited("duplicate")
	assert.NoError(t, err, "First put should succeed")

	err = q.PutWithLimited("duplicate")
	assert.NoError(t, err, "Second put with same value should succeed")

	assert.Equal(t, 2, q.Len(), "Queue should contain both duplicate elements")
}

func TestRateLimitingQueueImpl_DifferentTypes(t *testing.T) {

	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(5, 1))
	q := NewRateLimitingQueue(config)
	defer q.Shutdown()

	testCases := []any{
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

	config := NewRateLimitingQueueConfig().WithLimiter(NewBucketRateLimiterImpl(5, 1))
	q := NewRateLimitingQueue(config)
	defer q.Shutdown()

	_, err := q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty, "Get should return ErrQueueIsEmpty")
}

// delayRecordingCallback 只记录 OnDelay 的延迟值，用于观测限流入堆的实际退避。
type delayRecordingCallback struct {
	mu     sync.Mutex
	delays []int64
}

func (c *delayRecordingCallback) OnPut(any)              {}
func (c *delayRecordingCallback) OnGet(any)              {}
func (c *delayRecordingCallback) OnDone(any)             {}
func (c *delayRecordingCallback) OnPullError(any, error) {}
func (c *delayRecordingCallback) OnLimited(any)          {}

func (c *delayRecordingCallback) OnDelay(_ any, delay int64) {
	c.mu.Lock()
	c.delays = append(c.delays, delay)
	c.mu.Unlock()
}

func (c *delayRecordingCallback) snapshot() []int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]int64, len(c.delays))
	copy(out, c.delays)
	return out
}

// TestRateLimitingQueueImpl_Forget 验证队列级 Forget 与 limiter 的联动：
// 持有 LimiterForgetter 限流器时，Forget 清理该 item 的退避状态（下次入堆
// 延迟回到 base）；持有不支持 Forget 的限流器时，Forget 为安全 no-op。
func TestRateLimitingQueueImpl_Forget(t *testing.T) {
	callback := &delayRecordingCallback{}
	limiter := NewItemExponentialFailureRateLimiter(50*time.Millisecond, time.Second)
	config := NewRateLimitingQueueConfig().WithLimiter(limiter).WithCallback(callback)
	q := NewRateLimitingQueue(config)
	defer q.Shutdown()

	forgetter, ok := q.(LimiterForgetter)
	if !ok {
		t.Fatal("rate limiting queue must expose Forget for its limiter")
	}

	assert.NoError(t, q.PutWithLimited("k")) // 退避 50ms
	assert.NoError(t, q.PutWithLimited("k")) // 退避 100ms
	assert.Equal(t, []int64{50, 100}, callback.snapshot())

	forgetter.Forget("k")

	assert.NoError(t, q.PutWithLimited("k")) // 计数已清理，退避回到 50ms
	assert.Equal(t, []int64{50, 100, 50}, callback.snapshot())

	// nil item 为安全 no-op。
	forgetter.Forget(nil)
	assert.Equal(t, []int64{50, 100, 50}, callback.snapshot())
}

// TestRateLimitingQueueImpl_PutWithLimited_ItemExponential 验证 G8 集成：
// per-item 指数退避限流器驱动 PutWithLimited 按退避延迟入延迟堆，
// 延迟到期前不可 Get，到期后由 scheduler 搬运、按入队顺序可 Get。
func TestRateLimitingQueueImpl_PutWithLimited_ItemExponential(t *testing.T) {
	callback := &delayRecordingCallback{}
	limiter := NewItemExponentialFailureRateLimiter(100*time.Millisecond, time.Second)
	q := NewRateLimitingQueue(NewRateLimitingQueueConfig().WithLimiter(limiter).WithCallback(callback))
	defer q.Shutdown()

	assert.NoError(t, q.PutWithLimited("k"))
	assert.NoError(t, q.PutWithLimited("k"))

	// 两次入堆的退避延迟为 100ms、200ms（指数增长）。
	assert.Equal(t, []int64{100, 200}, callback.snapshot())
	assert.Equal(t, 2, q.Len(), "delayed items must be tracked in the delaying heap")

	// 延迟未到期：不可 Get。
	_, err := q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty)

	// 延迟到期并被搬运后，按入队顺序可依次 Get。
	var got []any
	assert.Eventually(t, func() bool {
		for len(got) < 2 {
			v, err := q.Get()
			if err != nil {
				return false
			}
			q.Done(v)
			got = append(got, v)
		}
		return true
	}, 3*time.Second, 20*time.Millisecond)
	assert.Equal(t, []any{"k", "k"}, got)
}

// TestRateLimitingQueueImpl_Forget_UnsupportedLimiter 验证限流器不支持 Forget
// （Nop/Bucket）时，队列级 Forget 为安全 no-op，不 panic 也不影响后续入队。
func TestRateLimitingQueueImpl_Forget_UnsupportedLimiter(t *testing.T) {
	for _, limiter := range []Limiter{NewNopRateLimiterImpl(), NewBucketRateLimiterImpl(100, 10)} {
		q := NewRateLimitingQueue(NewRateLimitingQueueConfig().WithLimiter(limiter))

		forgetter, ok := q.(LimiterForgetter)
		if !ok {
			q.Shutdown()
			t.Fatal("rate limiting queue must expose Forget even when limiter does not support it")
		}

		assert.NoError(t, q.PutWithLimited("k"))
		forgetter.Forget("k") // no-op，不应 panic
		assert.NoError(t, q.PutWithLimited("k"))

		q.Shutdown()
	}
}
