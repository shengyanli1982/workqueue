package workqueue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var DELAYDUCRATION = int64(150)

func TestDelayingQueueImpl_PutWithDelay(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	// Put content into queue
	err := q.PutWithDelay("test1", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test2", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test3", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	time.Sleep(time.Second)

	// Verify the queue state
	assert.Equal(t, 3, q.Len(), "Queue length should be 3")
	assert.Equal(t, []interface{}{"test1", "test2", "test3"}, q.Values(), "Queue values should be [test1, test2, test3]")
}

func TestDelayingQueueImpl_PutWithDelay_Closed(t *testing.T) {
	q := NewDelayingQueue(nil)
	q.Shutdown()

	// Put nil content into queue
	err := q.PutWithDelay(nil, 0)
	assert.Error(t, err, "Put should return an error")
}

func TestDelayingQueueImpl_PutWithDelay_Nil(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	// Put nil content into queue
	err := q.PutWithDelay(nil, 0)
	assert.Error(t, err, "Put should return an error")
}

func TestDelayingQueueImpl_PutWithDelay_Parallel(t *testing.T) {

}

func TestDelayingQueueImpl_Callback(t *testing.T) {

}
