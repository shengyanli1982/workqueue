package workqueue

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueueImpl_Put(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test2")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test3")
	assert.NoError(t, err, "Put should not return an error")

	assert.Equal(t, 3, q.Len(), "Queue length should be 3")
	assert.Equal(t, []interface{}{"test1", "test2", "test3"}, q.Values(), "Queue values should be [test1, test2, test3]")
}

func TestQueueImpl_Put_Closed(t *testing.T) {
	q := NewQueue(nil)
	q.Shutdown()

	err := q.Put("test1")
	assert.ErrorIs(t, err, ErrQueueIsClosed, "Put should return ErrQueueIsClosed")
}

func TestQueueImpl_Put_Nil(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	err := q.Put(nil)
	assert.ErrorIs(t, err, ErrElementIsNil, "Put should return ErrElementIsNil")
}

func TestQueueImpl_Put_Parallel(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	count := 1000

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

	assert.Equal(t, count, q.Len(), "Queue length should be 1000")
}

func TestQueueImpl_Get(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test2")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test3")
	assert.NoError(t, err, "Put should not return an error")

	v, err := q.Get()
	assert.NoError(t, err, "Get should not return an error")
	assert.Equal(t, "test1", v, "Get value should be test1")

	assert.Equal(t, 2, q.Len(), "Queue length should be 2")
	assert.Equal(t, []interface{}{"test2", "test3"}, q.Values(), "Queue values should be [test2, test3]")
}

func TestQueueImpl_Get_Closed(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test2")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test3")
	assert.NoError(t, err, "Put should not return an error")

	q.Shutdown()
	v, err := q.Get()
	assert.Error(t, err, "Get should return an error")
	assert.Nil(t, v, "Get value should be nil")

	assert.Equal(t, 0, q.Len(), "Queue length should be 0")
	assert.Equal(t, []interface{}{}, q.Values(), "Queue values should be []")
}

func TestQueueImpl_Get_Empty(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	v, err := q.Get()
	assert.Error(t, err, "Get should return an error")
	assert.Nil(t, v, "Get value should be nil")
}

func TestQueueImpl_Get_Parallel(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	count := 1000

	for i := 0; i < count; i++ {
		err := q.Put("test")
		assert.NoError(t, err, "Put should not return an error")
	}

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

	assert.Equal(t, 0, q.Len(), "Queue length should be 0")
}

func TestQueueImpl_PutAndGet_Parallel(t *testing.T) {
	q := NewQueue(nil)

	count := 1000

	wg := sync.WaitGroup{}
	wg.Add(count * 2)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			err := q.Put("test")
			assert.NoError(t, err, "Put should not return an error")
		}()
		go func() {
			defer wg.Done()
			for {
				if _, err := q.Get(); err != nil {
					if errors.Is(err, ErrQueueIsEmpty) {
						time.Sleep(50 * time.Millisecond)
						continue
					}
					if !errors.Is(err, ErrQueueIsClosed) {
						assert.NoError(t, err, "Get should not return an error")
					}
					break
				}
			}
		}()
	}

	time.Sleep(time.Second)

	q.Shutdown()
	wg.Wait()

	assert.Equal(t, 0, q.Len(), "Queue length should be 0")
}

func TestQueueImpl_Len(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test2")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test3")
	assert.NoError(t, err, "Put should not return an error")

	length := q.Len()
	assert.Equal(t, 3, length, "Queue length should be 3")
}

func TestQueueImpl_Len_Closed(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test2")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test3")
	assert.NoError(t, err, "Put should not return an error")

	q.Shutdown()
	length := q.Len()
	assert.Equal(t, 0, length, "Queue length should be 0")
}

func TestQueueImpl_Len_Empty(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	length := q.Len()
	assert.Equal(t, 0, length, "Queue length should be 0")
}

func TestQueueImpl_Range(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test2")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test3")
	assert.NoError(t, err, "Put should not return an error")

	values := make([]interface{}, 0)
	q.Range(func(value interface{}) bool {
		values = append(values, value)
		return true
	})

	assert.Equal(t, []interface{}{"test1", "test2", "test3"}, values, "Queue values should be [test1, test2, test3]")
}

func TestQueueImpl_Range_Empty(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	values := make([]interface{}, 0)
	q.Range(func(value interface{}) bool {
		values = append(values, value)
		return true
	})

	assert.Equal(t, []interface{}{}, values, "Queue values should be []")
}

func TestQueueImpl_Range_Closed(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test2")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test3")
	assert.NoError(t, err, "Put should not return an error")

	q.Shutdown()
	values := make([]interface{}, 0)
	q.Range(func(value interface{}) bool {
		values = append(values, value)
		return true
	})

	assert.Equal(t, []interface{}{}, values, "Queue values should be []")

}

func TestQueueImpl_IsClosed(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	assert.False(t, q.IsClosed(), "Queue should not be closed initially")

	q.Shutdown()

	assert.True(t, q.IsClosed(), "Queue should be closed")
}

type testQueueCallback struct {
	puts, gets, dones []interface{}
}

func (c *testQueueCallback) OnPut(value interface{}) {
	c.puts = append(c.puts, value)
}

func (c *testQueueCallback) OnGet(value interface{}) {
	c.gets = append(c.gets, value)
}

func (c *testQueueCallback) OnDone(value interface{}) {
	c.dones = append(c.dones, value)
}

func TestQueueImpl_Callback(t *testing.T) {
	callback := &testQueueCallback{}
	config := NewQueueConfig().WithCallback(callback)
	q := NewQueue(config)
	defer q.Shutdown()

	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test2")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test3")
	assert.NoError(t, err, "Put should not return an error")

	v, err := q.Get()
	assert.NoError(t, err, "Get should not return an error")
	assert.Equal(t, "test1", v, "Get value should be test1")

	q.Done(v)

	assert.Equal(t, []interface{}{"test1", "test2", "test3"}, callback.puts, "Callback puts should be [test1, test2, test3]")
	assert.Equal(t, []interface{}{"test1"}, callback.gets, "Callback gets should be [test1]")
	assert.Equal(t, []interface{}(nil), callback.dones, "Callback dones should be [test1]")
}

func TestQueueImpl_Idempotent_Put(t *testing.T) {
	callback := &testQueueCallback{}
	config := NewQueueConfig().WithCallback(callback).WithValueIdempotent()
	q := NewQueue(config)
	defer q.Shutdown()

	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test1")
	assert.ErrorIs(t, err, ErrElementAlreadyExist, "Put should return ErrElementAlreadyExist")
}

func TestQueueImpl_Idempotent_Put_ParallelSameValue(t *testing.T) {
	config := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(config)
	defer q.Shutdown()

	const count = 200
	var okCount int64
	var duplicateCount int64

	start := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			<-start
			err := q.Put("same-value")
			switch {
			case err == nil:
				atomic.AddInt64(&okCount, 1)
			case errors.Is(err, ErrElementAlreadyExist):
				atomic.AddInt64(&duplicateCount, 1)
			default:
				assert.NoError(t, err, "unexpected error from Put")
			}
		}()
	}

	close(start)
	wg.Wait()

	assert.Equal(t, int64(1), okCount, "only one goroutine should put successfully")
	assert.Equal(t, int64(count-1), duplicateCount, "others should report duplicate")
	assert.Equal(t, 1, q.Len(), "queue should contain exactly one element")
}

func TestQueueImpl_Idempotent_Get(t *testing.T) {
	callback := &testQueueCallback{}
	config := NewQueueConfig().WithCallback(callback).WithValueIdempotent()
	q := NewQueue(config)
	defer q.Shutdown()

	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test2")
	assert.NoError(t, err, "Put should not return an error")

	v, err := q.Get()
	assert.NoError(t, err, "Get should not return an error")
	assert.Equal(t, "test1", v, "Get value should be test1")

	q.Done(v)

	assert.Equal(t, 1, q.Len(), "Queue length should be 1")
	assert.Equal(t, q.Values(), []interface{}{"test2"}, "Queue values should be [test2]")

	queue := q.(*queueImpl)
	assert.Equal(t, queue.dirty.List(), []interface{}{"test2"}, "Queue dirty should be [test2]")
	assert.Equal(t, queue.processing.List(), []interface{}{}, "Queue processing should be []")
}

func TestQueueImpl_Idempotent_Callback(t *testing.T) {
	callback := &testQueueCallback{}
	config := NewQueueConfig().WithCallback(callback).WithValueIdempotent()
	q := NewQueue(config)
	defer q.Shutdown()

	err := q.Put("test1")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test2")
	assert.NoError(t, err, "Put should not return an error")
	err = q.Put("test3")
	assert.NoError(t, err, "Put should not return an error")

	v, err := q.Get()
	assert.NoError(t, err, "Get should not return an error")
	assert.Equal(t, "test1", v, "Get value should be test1")

	q.Done(v)

	assert.Equal(t, []interface{}{"test1", "test2", "test3"}, callback.puts, "Callback puts should be [test1, test2, test3]")
	assert.Equal(t, []interface{}{"test1"}, callback.gets, "Callback gets should be [test1]")
	assert.Equal(t, []interface{}{"test1"}, callback.dones, "Callback dones should be [test1]")
}

func TestQueueImpl_LargeCapacity(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	count := 1000000
	for i := 0; i < count; i++ {
		err := q.Put(i)
		assert.NoError(t, err, "Put should not return an error")
	}

	assert.Equal(t, count, q.Len(), "Queue length should match input count")

	for i := 0; i < count; i++ {
		v, err := q.Get()
		assert.NoError(t, err)
		assert.Equal(t, i, v)
	}
}

func TestQueueImpl_ComplexDataTypes(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	specialStr := "!@#$%^&*()"
	err := q.Put(specialStr)
	assert.NoError(t, err)

	type complexStruct struct {
		Field1 string
		Field2 []int
		Field3 map[string]interface{}
	}

	complexData := complexStruct{
		Field1: "test",
		Field2: []int{1, 2, 3},
		Field3: map[string]interface{}{"key": "value"},
	}

	err = q.Put(complexData)
	assert.NoError(t, err)

	v1, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, specialStr, v1)

	v2, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, complexData, v2)
}

func TestQueueImpl_DuplicateDone(t *testing.T) {
	callback := &testQueueCallback{}
	config := NewQueueConfig().WithCallback(callback)
	q := NewQueue(config)
	defer q.Shutdown()

	err := q.Put("test")
	assert.NoError(t, err)

	v, err := q.Get()
	assert.NoError(t, err)

	q.Done(v)
	q.Done(v)
	q.Done(v)

	assert.Equal(t, 0, len(callback.dones), "Done callback should only be called once")
}

func TestQueueImpl_Idempotent_DuplicateDone(t *testing.T) {
	callback := &testQueueCallback{}
	config := NewQueueConfig().WithCallback(callback).WithValueIdempotent()
	q := NewQueue(config)
	defer q.Shutdown()

	err := q.Put("test")
	assert.NoError(t, err)

	v, err := q.Get()
	assert.NoError(t, err)

	q.Done(v)
	q.Done(v)
	q.Done(v)

	assert.Equal(t, 1, len(callback.dones), "Done callback should only be called once")
}

func TestQueueImpl_ShutdownDuringProcessing(t *testing.T) {
	q := NewQueue(nil)

	for i := 0; i < 100; i++ {
		err := q.Put(i)
		assert.NoError(t, err)
	}

	for i := 0; i < 50; i++ {
		_, err := q.Get()
		assert.NoError(t, err)
	}

	q.Shutdown()

	assert.Equal(t, 0, q.Len(), "Queue should be empty after shutdown")
	assert.True(t, q.IsClosed(), "Queue should be closed")
}

func TestQueueImpl_RangeEarlyExit(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	for i := 0; i < 10; i++ {
		err := q.Put(i)
		assert.NoError(t, err)
	}

	count := 0
	q.Range(func(value interface{}) bool {
		count++
		return count < 5
	})

	assert.Equal(t, 5, count, "Range should have processed exactly 5 items")
	assert.Equal(t, 10, q.Len(), "Queue length should remain unchanged")
}
