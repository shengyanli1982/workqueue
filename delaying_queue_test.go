package workqueue

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type delayingcallback struct {
	a0, g0, d0, r0 []any
}

func (c *delayingcallback) OnAdd(item any) {
	c.a0 = append(c.a0, item)
}

func (c *delayingcallback) OnGet(item any) {
	c.g0 = append(c.g0, item)
}

func (c *delayingcallback) OnDone(item any) {
	c.d0 = append(c.d0, item)
}

func (c *delayingcallback) OnAddAfter(item any, _ time.Duration) {
	c.r0 = append(c.r0, item)
}

func TestDelayingQueue(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Stop()
	_ = q.AddAfter(time.Now().Local().UnixMilli(), 100*time.Millisecond)
	time.Sleep(time.Second)
	item, err := q.Get()
	assert.Equal(t, nil, err)
	q.Done(item)
	if item.(int64) > time.Now().UnixMilli() {
		assert.Error(t, errors.New("item should not be ready yet"))
	} else {
		assert.Equal(t, (time.Now().UnixMilli()-item.(int64))/1000, int64(1))
		return
	}
}

func TestDelayingQueue_TwoFireEarly(t *testing.T) {
	first := "foo"
	second := "bar"
	third := "baz"
	q := NewDelayingQueue(nil)
	defer q.Stop()
	_ = q.AddAfter(first, 300*time.Millisecond)
	_ = q.AddAfter(second, 100*time.Millisecond)
	time.Sleep(600 * time.Millisecond)
	item, err := q.Get()
	assert.Equal(t, nil, err)
	assert.Equal(t, item, second)
	q.Done(item)
	_ = q.AddAfter(third, 300*time.Millisecond)
	time.Sleep(600 * time.Millisecond)
	item, err = q.Get()
	assert.Equal(t, nil, err)
	assert.Equal(t, item, first)
	q.Done(item)
	time.Sleep(600 * time.Millisecond)
	item, err = q.Get()
	assert.Equal(t, nil, err)
	assert.Equal(t, item, third)
	q.Done(item)
}

func TestDelayingQueue_CallbackFuncs(t *testing.T) {
	conf := NewDelayingQConfig()
	conf.WithCallback(&delayingcallback{})

	q := NewDelayingQueue(conf)

	_ = q.Add("foo")
	_ = q.Add("bar")
	_ = q.Add("baz")

	// Start processing
	i, err := q.Get()
	assert.Equal(t, "foo", i)
	assert.Equal(t, nil, err)
	q.Done(i)

	assert.Equal(t, []any{"foo", "bar", "baz"}, q.config.callback.(*delayingcallback).a0)
	assert.Equal(t, []any{"foo"}, q.config.callback.(*delayingcallback).g0)
	assert.Equal(t, []any{"foo"}, q.config.callback.(*delayingcallback).d0)

	_ = q.AddAfter("cat", 100*time.Millisecond)
	time.Sleep(600 * time.Millisecond)

	assert.Equal(t, []any{"foo", "bar", "baz", "cat"}, q.config.callback.(*delayingcallback).a0)
	assert.Equal(t, []any{"cat"}, q.config.callback.(*delayingcallback).r0)

	// Stop the queue
	q.Stop()
}

func TestDelayingQueue_WithCustomQueue(t *testing.T) {
	conf := NewDelayingQConfig()
	conf.WithCallback(&delayingcallback{})

	queue := NewSimpleQueue(&conf.QConfig)
	q := NewDelayingQueueWithCustomQueue(conf, queue)

	_ = q.Add("foo")
	_ = q.Add("bar")
	_ = q.Add("baz")

	// Start processing
	i, err := q.Get()
	assert.Equal(t, "foo", i)
	assert.Equal(t, nil, err)
	q.Done(i)

	assert.Equal(t, []any{"foo", "bar", "baz"}, q.config.callback.(*delayingcallback).a0)
	assert.Equal(t, []any{"foo"}, q.config.callback.(*delayingcallback).g0)
	assert.Equal(t, []any{"foo"}, q.config.callback.(*delayingcallback).d0)

	_ = q.AddAfter("cat", 100*time.Millisecond)
	time.Sleep(600 * time.Millisecond)

	assert.Equal(t, []any{"foo", "bar", "baz", "cat"}, q.config.callback.(*delayingcallback).a0)
	assert.Equal(t, []any{"cat"}, q.config.callback.(*delayingcallback).r0)

	// Stop the queue
	q.Stop()
}
