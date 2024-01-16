package workqueue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue_Standard(t *testing.T) {
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

func TestQueue_ItemExist(t *testing.T) {
	q := NewQueue(nil)
	err := q.Add("foo")
	assert.Equal(t, nil, err)
	err = q.Add("foo")
	assert.Equal(t, ErrorQueueElementExist, err)
	q.Stop()
}

func TestQueue_ShutDown(t *testing.T) {
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

func TestQueue_ItemEmptyGet(t *testing.T) {
	q := NewQueue(nil)
	item, err := q.Get()
	assert.Equal(t, nil, item)
	assert.Equal(t, ErrorQueueEmpty, err)
	q.Stop()
}

func TestQueue_BlockingGet(t *testing.T) {
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

func TestQueue_ReInsert(t *testing.T) {
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

func TestQueue_AddInProcessing(t *testing.T) {
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

func TestQueue_CallbackFuncs(t *testing.T) {
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

func TestQueue_GetItemAfterStop(t *testing.T) {
	q := NewQueue(nil)
	err := q.Add("foo")
	assert.Equal(t, nil, err)
	q.Stop()
	_, err = q.Get()
	assert.Equal(t, ErrorQueueClosed, err)
}
