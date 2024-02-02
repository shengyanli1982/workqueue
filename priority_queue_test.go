package workqueue

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type prioritycallback struct {
	a0, g0, d0, p0 []any
}

func (c *prioritycallback) OnAdd(item any) {
	c.a0 = append(c.a0, item)
}

func (c *prioritycallback) OnGet(item any) {
	c.g0 = append(c.g0, item)
}

func (c *prioritycallback) OnDone(item any) {
	c.d0 = append(c.d0, item)
}

func (c *prioritycallback) OnAddWeight(item any, _ int) {
	c.p0 = append(c.p0, item)
}

func TestPriorityQueue(t *testing.T) {
	q := NewPriorityQueue(nil)
	defer q.Stop()
	_ = q.AddWeight(time.Now().Local().UnixMilli(), 10)
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

func TestPriorityQueue_TwoFireEarly(t *testing.T) {
	first := "foo"
	second := "bar"
	third := "baz"
	q := NewPriorityQueue(nil)
	defer q.Stop()
	_ = q.AddWeight(first, 30)
	_ = q.AddWeight(second, 10)
	time.Sleep(600 * time.Millisecond)
	item, err := q.Get()
	assert.Equal(t, nil, err)
	assert.Equal(t, item, second)
	q.Done(item)
	_ = q.AddWeight(third, 5)
	time.Sleep(600 * time.Millisecond)
	item, err = q.Get()
	assert.Equal(t, nil, err)
	assert.Equal(t, item, first)
	q.Done(item)
	item, err = q.Get()
	assert.Equal(t, nil, err)
	assert.Equal(t, item, third)
	q.Done(item)
}

func TestPriorityQueue_CallbackFuncs(t *testing.T) {
	conf := NewPriorityQConfig()
	conf.WithCallback(&prioritycallback{})

	q := NewPriorityQueue(conf)

	_ = q.Add("foo")
	_ = q.Add("bar")
	_ = q.Add("baz")

	// Start processing
	i, err := q.Get()
	assert.Equal(t, "foo", i)
	assert.Equal(t, nil, err)
	q.Done(i)

	assert.Equal(t, []any{"foo", "bar", "baz"}, q.config.callback.(*prioritycallback).a0)
	assert.Equal(t, []any{"foo"}, q.config.callback.(*prioritycallback).g0)
	assert.Equal(t, []any{"foo"}, q.config.callback.(*prioritycallback).d0)

	_ = q.AddWeight("cat", 100)
	time.Sleep(600 * time.Millisecond)

	assert.Equal(t, []any{"foo", "bar", "baz", "cat"}, q.config.callback.(*prioritycallback).a0)
	assert.Equal(t, []any{"cat"}, q.config.callback.(*prioritycallback).p0)

	// Stop the queue
	q.Stop()
}
