package workqueue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleQueue(t *testing.T) {
	q := NewSimpleQueue(nil)
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

func TestSimpleQueue_ItemExist(t *testing.T) {
	q := NewSimpleQueue(nil)
	err := q.Add("foo")
	assert.Equal(t, nil, err)
	err = q.Add("foo")
	assert.Equal(t, nil, err)
	q.Stop()
}

func TestSimpleQueue_ShutDown(t *testing.T) {
	q := NewSimpleQueue(nil)
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

func TestSimpleQueue_ItemEmptyGet(t *testing.T) {
	q := NewSimpleQueue(nil)
	item, err := q.Get()
	assert.Equal(t, nil, item)
	assert.Equal(t, ErrorQueueEmpty, err)
	q.Stop()
}

func TestSimpleQueue_CallbackFuncs(t *testing.T) {
	conf := NewQConfig()
	conf.WithCallback(&callback{})

	q := NewSimpleQueue(conf)

	_ = q.Add("foo")
	_ = q.Add("bar")
	_ = q.Add("baz")

	// Start processing
	i, err := q.Get()
	assert.Equal(t, "foo", i)
	assert.Equal(t, nil, err)
	q.Done(i)

	assert.Equal(t, []any{"foo", "bar", "baz"}, q.config.callback.(*callback).a0)
	assert.Equal(t, []any{"foo"}, q.config.callback.(*callback).g0)
	assert.Equal(t, []any{"foo"}, q.config.callback.(*callback).d0)

	// Stop the queue
	q.Stop()
}

func TestSimpleQueue_GetItemAfterStop(t *testing.T) {
	q := NewSimpleQueue(nil)
	err := q.Add("foo")
	assert.Equal(t, nil, err)
	q.Stop()
	_, err = q.Get()
	assert.Equal(t, ErrorQueueClosed, err)
}

func TestSimpleQueue_GetStoreValues(t *testing.T) {
	q := NewSimpleQueue(nil)
	err := q.Add("foo")
	assert.Equal(t, nil, err)
	err = q.Add("bar")
	assert.Equal(t, nil, err)
	err = q.Add("baz")
	assert.Equal(t, nil, err)

	values := q.GetValues()
	assert.Equal(t, []any{"foo", "bar", "baz"}, values)

	q.Stop()
}
func TestSimpleQueue_Range(t *testing.T) {
	q := NewSimpleQueue(nil)
	err := q.Add("foo")
	assert.Equal(t, nil, err)
	err = q.Add("bar")
	assert.Equal(t, nil, err)
	err = q.Add("baz")
	assert.Equal(t, nil, err)

	var result []any
	q.Range(func(element any) bool {
		result = append(result, element)
		return true
	})

	expected := []any{"foo", "bar", "baz"}
	assert.Equal(t, expected, result)

	q.Stop()
}
