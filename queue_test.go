package workqueue

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueueImpl_Put(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	// Put content into queue
	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test2")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test3")
	assert.NoError(t, err, "Put should not return an error")

	// Verify the queue state
	assert.Equal(t, 3, q.Len(), "Queue length should be 3")
	assert.Equal(t, []interface{}{"test1", "test2", "test3"}, q.Values(), "Queue values should be [test1, test2, test3]")
}

func TestQueueImpl_Put_Closed(t *testing.T) {
	q := NewQueue(nil)
	q.Shutdown()

	// Put content into closed queue
	err := q.Put("test1")
	assert.Error(t, err, "Put should return an error")
}

func TestQueueImpl_Put_Nil(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	// Put nil content into queue
	err := q.Put(nil)
	assert.Error(t, err, "Put should return an error")
}

func TestQueueImpl_Put_Parallel(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	count := 1000

	// Put content into queue in parallel
	wg := sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			err := q.Put("test")
			assert.NoError(t, err, "Put should not return an error")
		}()
	}
	wg.Wait()

	// Verify the queue state
	assert.Equal(t, count, q.Len(), "Queue length should be 1000")
}

func TestQueueImpl_Get(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	// Put content into queue
	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test2")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test3")
	assert.NoError(t, err, "Put should not return an error")

	// Get content from queue
	v, err := q.Get()
	assert.NoError(t, err, "Get should not return an error")
	assert.Equal(t, "test1", v, "Get value should be test1")

	// Verify the queue state
	assert.Equal(t, 2, q.Len(), "Queue length should be 2")
	assert.Equal(t, []interface{}{"test2", "test3"}, q.Values(), "Queue values should be [test2, test3]")
}

func TestQueueImpl_Get_Closed(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	// Put content into queue
	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test2")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test3")
	assert.NoError(t, err, "Put should not return an error")

	// Get content from closed queue
	q.Shutdown()
	v, err := q.Get()
	assert.Error(t, err, "Get should return an error")
	assert.Nil(t, v, "Get value should be nil")

	//Verify the queue state
	assert.Equal(t, 0, q.Len(), "Queue length should be 0")
	assert.Equal(t, []interface{}{}, q.Values(), "Queue values should be []")
}

func TestQueueImpl_Get_Empty(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	// Get content from empty queue
	v, err := q.Get()
	assert.Error(t, err, "Get should return an error")
	assert.Nil(t, v, "Get value should be nil")
}

func TestQueueImpl_Get_Parallel(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	count := 1000

	// Put content into queue
	for i := 0; i < count; i++ {
		err := q.Put("test")
		assert.NoError(t, err, "Put should not return an error")
	}

	// Get content from queue in parallel
	wg := sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			_, err := q.Get()
			assert.NoError(t, err, "Get should not return an error")
		}()
	}
	wg.Wait()

	// Verify the queue state
	assert.Equal(t, 0, q.Len(), "Queue length should be 0")
}

func TestQueueImpl_PutAndGet_Parallel(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	count := 1000

	// Put and get content from queue in parallel
	wg := sync.WaitGroup{}
	wg.Add(count * 2)
	closed := atomic.Bool{}
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			err := q.Put("test")
			assert.NoError(t, err, "Put should not return an error")
		}()
		go func() {
			defer wg.Done()
			for {
				if closed.Load() {
					break
				}
				v, err := q.Get()
				if v != nil {
					assert.NoError(t, err, "Get should not return an error")
				}
			}
		}()
		if i == count-1 {
			time.Sleep(time.Second)
			closed.Store(true)
		}
	}
	wg.Wait()

	// Verify the queue state
	assert.Equal(t, 0, q.Len(), "Queue length should be 0")

}

func TestQueueImpl_Len(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	// Put content into queue
	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test2")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test3")
	assert.NoError(t, err, "Put should not return an error")

	// Verify the queue length
	length := q.Len()
	assert.Equal(t, 3, length, "Queue length should be 3")
}

func TestQueueImpl_Len_Closed(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	// Put content into queue
	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test2")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test3")
	assert.NoError(t, err, "Put should not return an error")

	// Verify the queue length when closed
	q.Shutdown()
	length := q.Len()
	assert.Equal(t, 0, length, "Queue length should be 0")
}

func TestQueueImpl_Len_Empty(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	// Verify the queue length when empty
	length := q.Len()
	assert.Equal(t, 0, length, "Queue length should be 0")
}

func TestQueueImpl_IsClosed(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	// Verify that the queue is not closed initially
	assert.False(t, q.IsClosed(), "Queue should not be closed initially")

	// Close the queue
	q.Shutdown()

	// Verify that the queue is closed
	assert.True(t, q.IsClosed(), "Queue should be closed")
}
