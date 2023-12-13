package workqueue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueueStandard(t *testing.T) {
	q := NewQueue(nil)
	err := q.Add("foo")
	assert.Equal(t, nil, err)
	err = q.Add("bar")
	assert.Equal(t, nil, err)
	err = q.Add("baz")
	assert.Equal(t, nil, err)
	assert.Equal(t, 3, q.Len())
	item, err := q.Get()
	assert.Equal(t, "foo", item)
	assert.Equal(t, nil, err)
	q.Done(item)
	item, err = q.Get()
	assert.Equal(t, "bar", item)
	assert.Equal(t, nil, err)
	q.Done(item)
	item, err = q.Get()
	assert.Equal(t, "baz", item)
	assert.Equal(t, nil, err)
	q.Done(item)
	assert.Equal(t, 0, q.Len())
	q.Stop()
}

func TestQueueItemExist(t *testing.T) {
	q := NewQueue(nil)
	err := q.Add("foo")
	assert.Equal(t, nil, err)
	err = q.Add("foo")
	assert.Equal(t, ErrorQueueElementExist, err)
	q.Stop()
}

func TestQueueShutDown(t *testing.T) {
	q := NewQueue(nil)
	err := q.Add("foo")
	assert.Equal(t, nil, err)
	err = q.Add("bar")
	assert.Equal(t, nil, err)
	err = q.Add("baz")
	assert.Equal(t, nil, err)
	assert.Equal(t, 3, q.Len())
	q.Stop()
	assert.Equal(t, 0, q.Len())
	assert.True(t, q.IsClosed())
}

func TestQueueItemEmptyGet(t *testing.T) {
	q := NewQueue(nil)
	item, err := q.Get()
	assert.Equal(t, nil, item)
	assert.Equal(t, ErrorQueueEmpty, err)
	q.Stop()
}

func TestQueueBlockingGet(t *testing.T) {
	q := NewQueue(nil)
	err := q.Add("foo")
	assert.Equal(t, nil, err)
	err = q.Add("bar")
	assert.Equal(t, nil, err)
	err = q.Add("baz")
	assert.Equal(t, nil, err)
	assert.Equal(t, 3, q.Len())
	item, err := q.GetWithBlock()
	assert.Equal(t, "foo", item)
	assert.Equal(t, nil, err)
	q.Done(item)
	item, err = q.GetWithBlock()
	assert.Equal(t, "bar", item)
	assert.Equal(t, nil, err)
	q.Done(item)
	item, err = q.GetWithBlock()
	assert.Equal(t, "baz", item)
	assert.Equal(t, nil, err)
	q.Done(item)
	assert.Equal(t, 0, q.Len())
	q.Stop()
}

func TestQueueAddItemFull(t *testing.T) {
	conf := NewQConfig()
	conf.WithCap(defaultQueueCap + 1)
	q := NewQueue(conf)
	for i := 0; i < defaultQueueCap+1; i++ {
		err := q.Add(i)
		assert.Equal(t, nil, err)
	}
	err := q.Add("foo")
	assert.Equal(t, ErrorQueueFull, err)
	q.Stop()
}

func TestQueueReInsert(t *testing.T) {
	q := NewQueue(nil)

	_ = q.Add("foo")

	// Start processing
	i, err := q.Get()
	assert.Equal(t, "foo", i)
	assert.Equal(t, nil, err)

	// Add it back while processing
	_ = q.Add(i)

	// Finish it up
	q.Done(i)

	// It can not get it again
	_, err = q.Get()
	assert.Equal(t, ErrorQueueEmpty, err)

	// Finish that one up
	q.Done(i)
	assert.Equal(t, 0, q.Len())

	q.Stop()
}

func TestQueueAddInProcessing(t *testing.T) {
	q := NewQueue(nil)

	_ = q.Add("foo")

	// Start processing
	i, err := q.Get()
	assert.Equal(t, "foo", i)
	assert.Equal(t, nil, err)

	// Add it back while processing
	_ = q.Add(i)

	// Add it back while processing
	_ = q.Add(i)

	// Finish it up
	q.Done(i)

	// It can not get it again
	_, err = q.Get()
	assert.Equal(t, ErrorQueueEmpty, err)

	// Finish that one up
	q.Done(i)
	assert.Equal(t, 0, q.Len())

	q.Stop()
}

type callback struct {
	a0, g0, d0 []any
}

func (cb *callback) OnAdd(item any) {
	cb.a0 = append(cb.a0, item)
}
func (cb *callback) OnGet(item any) {
	cb.g0 = append(cb.g0, item)
}
func (cb *callback) OnDone(item any) {
	cb.d0 = append(cb.d0, item)
}

func TestQueueCallbackFuncs(t *testing.T) {
	conf := NewQConfig()
	conf.WithCallback(&callback{})

	q := NewQueue(conf)

	_ = q.Add("foo")
	_ = q.Add("bar")
	_ = q.Add("baz")

	// Start processing
	i, err := q.Get()
	assert.Equal(t, "foo", i)
	assert.Equal(t, nil, err)
	q.Done(i)

	assert.Equal(t, []any{"foo", "bar", "baz"}, q.config.cb.(*callback).a0)
	assert.Equal(t, []any{"foo"}, q.config.cb.(*callback).g0)
	assert.Equal(t, []any{"foo"}, q.config.cb.(*callback).d0)

	// Stop the queue
	q.Stop()
}
