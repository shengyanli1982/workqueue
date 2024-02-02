package workqueue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type rateLimitingcallback struct {
	a0, g0, d0, l0, f0, r0, t0 []any
}

func (c *rateLimitingcallback) OnAdd(item any) {
	c.a0 = append(c.a0, item)
}

func (c *rateLimitingcallback) OnGet(item any) {
	c.g0 = append(c.g0, item)
}

func (c *rateLimitingcallback) OnDone(item any) {
	c.d0 = append(c.d0, item)
}

func (c *rateLimitingcallback) OnAddAfter(item any, _ time.Duration) {
	c.t0 = append(c.t0, item)
}

func (c *rateLimitingcallback) OnAddLimited(item any) {
	c.l0 = append(c.l0, item)
}

func (c *rateLimitingcallback) OnForget(item any) {
	c.f0 = append(c.f0, item)
}

func (c *rateLimitingcallback) OnGetTimes(item any, _ int) {
	c.r0 = append(c.r0, item)
}

func TestRatelimitingQueue(t *testing.T) {
	conf := NewRateLimitingQConfig().WithLimiter(NewBucketRateLimiter(float64(3), 1))
	q := NewRateLimitingQueue(conf)
	defer q.Stop()
	_ = q.Add("foo")
	item, err := q.Get()
	assert.Equal(t, nil, err)
	assert.Equal(t, item, "foo")
	q.Done(item)
	_ = q.AddAfter("bar", 100*time.Millisecond)
	time.Sleep(600 * time.Millisecond)
	item, err = q.Get()
	assert.Equal(t, nil, err)
	assert.Equal(t, item, "bar")
	q.Done(item)
	for i := 0; i < 10; i++ {
		_ = q.AddLimited(i)
	}
	time.Sleep(3500 * time.Millisecond)
	assert.Equal(t, 10, q.Len())
	for i := 0; i < 10; i++ {
		item, err = q.Get()
		assert.Equal(t, nil, err)
		assert.Equal(t, item, i)
		q.Done(item)
	}
}

func TestRatelimitingQueue_CallbackFuncs(t *testing.T) {
	conf := NewRateLimitingQConfig().WithLimiter(NewBucketRateLimiter(float64(3), 1)).WithCallback(&rateLimitingcallback{})
	q := NewRateLimitingQueue(conf)
	defer q.Stop()

	_ = q.Add("foo")
	_ = q.Add("bar")
	_ = q.Add("baz")
	assert.Equal(t, 3, len(q.config.callback.(*rateLimitingcallback).a0))

	// Start processing
	i, err := q.Get()
	assert.Equal(t, "foo", i)
	assert.Equal(t, nil, err)
	q.Done(i)
	assert.Equal(t, 1, len(q.config.callback.(*rateLimitingcallback).g0))
	assert.Equal(t, 1, len(q.config.callback.(*rateLimitingcallback).d0))

	// Add element delay 100ms
	_ = q.AddAfter("x4", 100*time.Millisecond)
	assert.Equal(t, 1, len(q.config.callback.(*rateLimitingcallback).t0))
	time.Sleep(600 * time.Millisecond)
	assert.Equal(t, 4, len(q.config.callback.(*rateLimitingcallback).a0))

	// Add element ratelimit
	for i := 0; i < 10; i++ {
		_ = q.AddLimited(i)
	}
	assert.Equal(t, 10, len(q.config.callback.(*rateLimitingcallback).l0))

	q.Forget("x5")
	assert.Equal(t, 1, len(q.config.callback.(*rateLimitingcallback).f0))

	q.NumLimitTimes("x6")
	assert.Equal(t, 1, len(q.config.callback.(*rateLimitingcallback).r0))

}
