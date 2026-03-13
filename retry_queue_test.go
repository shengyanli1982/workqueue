package workqueue

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryQueue_RetryAndRequeue(t *testing.T) {
	config := NewRetryQueueConfig().WithPolicy(NewExponentialRetryPolicy(50*time.Millisecond, 50*time.Millisecond, 3))
	q := NewRetryQueue(config)
	defer q.Shutdown()

	err := q.Put("task")
	assert.NoError(t, err)

	value, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "task", value)

	err = q.Retry(value, errors.New("failed"))
	assert.NoError(t, err)
	assert.Equal(t, 1, q.NumRequeues("task"))

	assert.Eventually(t, func() bool {
		v, getErr := q.Get()
		if getErr != nil {
			return false
		}
		return v == "task"
	}, 2*time.Second, 20*time.Millisecond)
}

func TestRetryQueue_RetryExhausted(t *testing.T) {
	config := NewRetryQueueConfig().WithPolicy(NewExponentialRetryPolicy(0, 0, 1))
	q := NewRetryQueue(config)
	defer q.Shutdown()

	err := q.Put("task")
	assert.NoError(t, err)

	value, err := q.Get()
	assert.NoError(t, err)

	err = q.Retry(value, errors.New("first failed"))
	assert.NoError(t, err)

	var ok bool
	assert.Eventually(t, func() bool {
		value, err = q.Get()
		ok = err == nil
		return ok
	}, 2*time.Second, 20*time.Millisecond)
	assert.True(t, ok)

	err = q.Retry(value, errors.New("second failed"))
	assert.ErrorIs(t, err, ErrRetryExhausted)
	assert.Equal(t, 0, q.NumRequeues("task"))
}

func TestRetryQueue_Forget(t *testing.T) {
	config := NewRetryQueueConfig().WithPolicy(NewExponentialRetryPolicy(0, 0, 3))
	q := NewRetryQueue(config)
	defer q.Shutdown()

	err := q.Put("task")
	assert.NoError(t, err)

	value, err := q.Get()
	assert.NoError(t, err)

	err = q.Retry(value, errors.New("failed"))
	assert.NoError(t, err)
	assert.Equal(t, 1, q.NumRequeues("task"))

	q.Forget("task")
	assert.Equal(t, 0, q.NumRequeues("task"))
}

func TestRetryQueue_EmptyRetryKey(t *testing.T) {
	config := NewRetryQueueConfig().
		WithKeyFunc(func(interface{}) string { return "" }).
		WithPolicy(NewExponentialRetryPolicy(0, 0, 1))
	q := NewRetryQueue(config)
	defer q.Shutdown()

	err := q.Put("task")
	assert.NoError(t, err)

	value, err := q.Get()
	assert.NoError(t, err)

	err = q.Retry(value, errors.New("failed"))
	assert.ErrorIs(t, err, ErrRetryKeyEmpty)
}

type testRetryQueueCallback struct {
	mu sync.Mutex

	retries   []interface{}
	exhausted []interface{}
	forgets   []interface{}
}

func (c *testRetryQueueCallback) OnPut(interface{}) {}

func (c *testRetryQueueCallback) OnGet(interface{}) {}

func (c *testRetryQueueCallback) OnDone(interface{}) {}

func (c *testRetryQueueCallback) OnDelay(interface{}, int64) {}

func (c *testRetryQueueCallback) OnPullError(interface{}, error) {}

func (c *testRetryQueueCallback) OnRetry(value interface{}, _ int, _ time.Duration, _ error) {
	c.mu.Lock()
	c.retries = append(c.retries, value)
	c.mu.Unlock()
}

func (c *testRetryQueueCallback) OnRetryExhausted(value interface{}, _ int, _ error) {
	c.mu.Lock()
	c.exhausted = append(c.exhausted, value)
	c.mu.Unlock()
}

func (c *testRetryQueueCallback) OnForget(value interface{}) {
	c.mu.Lock()
	c.forgets = append(c.forgets, value)
	c.mu.Unlock()
}

func TestRetryQueue_Callback(t *testing.T) {
	callback := &testRetryQueueCallback{}
	config := NewRetryQueueConfig().
		WithCallback(callback).
		WithPolicy(NewExponentialRetryPolicy(0, 0, 1))
	q := NewRetryQueue(config)
	defer q.Shutdown()

	err := q.Put("task")
	assert.NoError(t, err)

	value, err := q.Get()
	assert.NoError(t, err)

	err = q.Retry(value, errors.New("failed"))
	assert.NoError(t, err)

	var ok bool
	assert.Eventually(t, func() bool {
		value, err = q.Get()
		ok = err == nil
		return ok
	}, 2*time.Second, 20*time.Millisecond)
	assert.True(t, ok)
	err = q.Retry(value, errors.New("failed-again"))
	assert.ErrorIs(t, err, ErrRetryExhausted)

	q.Forget("task")

	callback.mu.Lock()
	defer callback.mu.Unlock()
	assert.Equal(t, []interface{}{"task"}, callback.retries)
	assert.Equal(t, []interface{}{"task"}, callback.exhausted)
	assert.Equal(t, []interface{}{"task"}, callback.forgets)
}
