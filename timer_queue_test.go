package workqueue

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimerQueue_Order(t *testing.T) {
	q := NewTimerQueue(nil)
	defer q.Shutdown()

	now := time.Now()
	assert.NoError(t, q.PutAt("late", now.Add(60*time.Millisecond)))
	assert.NoError(t, q.PutAt("early", now.Add(20*time.Millisecond)))

	v1, err := waitQueueGet(t, q, 200*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "early", v1)

	v2, err := waitQueueGet(t, q, 200*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "late", v2)
}

func TestTimerQueue_Cancel(t *testing.T) {
	q := NewTimerQueue(nil)
	defer q.Shutdown()

	assert.NoError(t, q.PutAfter("cancel-me", 40*time.Millisecond))
	assert.True(t, q.Cancel("cancel-me"))

	time.Sleep(60 * time.Millisecond)
	_, err := q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty)
}

func TestTimerQueue_PutAfter_Immediate(t *testing.T) {
	q := NewTimerQueue(nil)
	defer q.Shutdown()

	assert.NoError(t, q.PutAfter("immediate", -10*time.Millisecond))

	value, err := waitQueueGet(t, q, 80*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "immediate", value)
}

func TestTimerQueue_Close(t *testing.T) {
	q := NewTimerQueue(nil)
	q.Shutdown()

	assert.ErrorIs(t, q.PutAfter("x", time.Second), ErrQueueIsClosed)
}

func TestTimerQueue_CancelNotFound(t *testing.T) {
	q := NewTimerQueue(nil)
	defer q.Shutdown()

	assert.False(t, q.Cancel("not-found"))
}

func TestTimerQueue_Cancel_DuplicateComparableValue(t *testing.T) {
	q := NewTimerQueue(nil)
	defer q.Shutdown()

	now := time.Now()
	assert.NoError(t, q.PutAt("dup", now.Add(80*time.Millisecond)))
	assert.NoError(t, q.PutAt("dup", now.Add(20*time.Millisecond)))

	assert.True(t, q.Cancel("dup"))

	time.Sleep(40 * time.Millisecond)
	_, err := q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty)

	value, err := waitQueueGet(t, q, 200*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "dup", value)
}

func TestTimerQueue_Cancel_NonComparableValue(t *testing.T) {
	q := NewTimerQueue(nil)
	defer q.Shutdown()

	assert.NoError(t, q.PutAfter([]int{1, 2, 3}, 40*time.Millisecond))
	assert.True(t, q.Cancel([]int{1, 2, 3}))

	time.Sleep(60 * time.Millisecond)
	_, err := q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty)
}

func TestTimerQueue_HeapRange(t *testing.T) {
	q := NewTimerQueue(nil)
	defer q.Shutdown()

	assert.NoError(t, q.PutAfter("a", 40*time.Millisecond))
	assert.NoError(t, q.PutAfter("b", 20*time.Millisecond))
	assert.NoError(t, q.PutAfter("c", 30*time.Millisecond))

	items := make([]interface{}, 0, 3)
	q.HeapRange(func(value interface{}, _ int64) bool {
		items = append(items, value)
		return true
	})

	assert.Len(t, items, 3)
}

func waitTimerQueueGet(t *testing.T, q Queue, timeout time.Duration) (interface{}, error) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		value, err := q.Get()
		if err == nil {
			return value, nil
		}
		if !errors.Is(err, ErrQueueIsEmpty) {
			return nil, err
		}
		time.Sleep(2 * time.Millisecond)
	}

	return nil, ErrQueueIsEmpty
}
