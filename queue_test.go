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
	assert.Equal(t, []any{"test1", "test2", "test3"}, q.Values(), "Queue values should be [test1, test2, test3]")
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
	assert.Equal(t, []any{"test2", "test3"}, q.Values(), "Queue values should be [test2, test3]")
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
	assert.Equal(t, []any{}, q.Values(), "Queue values should be []")
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

	values := make([]any, 0)
	q.Range(func(value any) bool {
		values = append(values, value)
		return true
	})

	assert.Equal(t, []any{"test1", "test2", "test3"}, values, "Queue values should be [test1, test2, test3]")
}

func TestQueueImpl_Range_Empty(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	values := make([]any, 0)
	q.Range(func(value any) bool {
		values = append(values, value)
		return true
	})

	assert.Equal(t, []any{}, values, "Queue values should be []")
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
	values := make([]any, 0)
	q.Range(func(value any) bool {
		values = append(values, value)
		return true
	})

	assert.Equal(t, []any{}, values, "Queue values should be []")

}

func TestQueueImpl_IsClosed(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	assert.False(t, q.IsClosed(), "Queue should not be closed initially")

	q.Shutdown()

	assert.True(t, q.IsClosed(), "Queue should be closed")
}

type testQueueCallback struct {
	puts, gets, dones []any
}

func (c *testQueueCallback) OnPut(value any) {
	c.puts = append(c.puts, value)
}

func (c *testQueueCallback) OnGet(value any) {
	c.gets = append(c.gets, value)
}

func (c *testQueueCallback) OnDone(value any) {
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

	assert.Equal(t, []any{"test1", "test2", "test3"}, callback.puts, "Callback puts should be [test1, test2, test3]")
	assert.Equal(t, []any{"test1"}, callback.gets, "Callback gets should be [test1]")
	assert.Equal(t, []any(nil), callback.dones, "Callback dones should be [test1]")
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
	assert.Equal(t, q.Values(), []any{"test2"}, "Queue values should be [test2]")

	queue := q.(*queueImpl)
	assert.Equal(t, queue.state.List(), []any{"test2"}, "Queue state should be [test2]")
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

	assert.Equal(t, []any{"test1", "test2", "test3"}, callback.puts, "Callback puts should be [test1, test2, test3]")
	assert.Equal(t, []any{"test1"}, callback.gets, "Callback gets should be [test1]")
	assert.Equal(t, []any{"test1"}, callback.dones, "Callback dones should be [test1]")
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
		Field3 map[string]any
	}

	complexData := complexStruct{
		Field1: "test",
		Field2: []int{1, 2, 3},
		Field3: map[string]any{"key": "value"},
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
	q.Range(func(value any) bool {
		count++
		return count < 5
	})

	assert.Equal(t, 5, count, "Range should have processed exactly 5 items")
	assert.Equal(t, 10, q.Len(), "Queue length should remain unchanged")
}

// TestQueueImpl_Idempotent_PutWhileProcessing 验证 G2 挂起 Put 语义：
// 元素被 Get 走但尚未 Done（处理中）时，再次 Put 同值必须：
//  1. 返回 nil（接受入队请求，而非 ErrElementAlreadyExist）；
//  2. 不立即进入 list（仅挂起标记，元素仍在处理中）；
//  3. Done 时真正重新入队。
//
// 该语义是 D1/D2 的结构性修复基础：包装队列（reaper/Retry）的
// “先 Put 后 Done” 顺序在新语义下自动正确。
func TestQueueImpl_Idempotent_PutWhileProcessing(t *testing.T) {
	config := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("job"))

	value, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "job", value)
	assert.Equal(t, 0, q.Len())

	// 处理中再次 Put → 接受（挂起标记），但尚不入队。
	assert.NoError(t, q.Put("job"), "Put while processing must be accepted")
	assert.Equal(t, 0, q.Len(), "pending element must not enter list before Done")

	// Done 触发重新入队。
	q.Done("job")
	assert.Equal(t, 1, q.Len(), "Done must requeue the pending element")

	value, err = q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "job", value)

	// 生命周期正常结束后，同值可再次完整入队。
	q.Done("job")
	assert.Equal(t, 0, q.Len())
	assert.NoError(t, q.Put("job"))
	assert.Equal(t, 1, q.Len())
}

// TestQueueImpl_Idempotent_PendingPut_CallbackCounts 验证挂起 Put 与 Done 重入队的回调计数语义：
// 成功 Put（含挂起 Put）每次触发 OnPut；Done 驱动的重新入队不重复触发 OnPut；
// 处理完成（processing 移除成功）触发 OnDone，重入队场景同样触发。
func TestQueueImpl_Idempotent_PendingPut_CallbackCounts(t *testing.T) {
	callback := &testQueueCallback{}
	config := NewQueueConfig().WithCallback(callback).WithValueIdempotent()
	q := NewQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("job")) // OnPut ×1

	value, err := q.Get() // OnGet ×1
	assert.NoError(t, err)

	assert.NoError(t, q.Put(value)) // 挂起 Put → OnPut ×1（共 2 次）

	q.Done(value) // 重入队 → OnDone ×1，且不额外触发 OnPut

	assert.Equal(t, []any{"job", "job"}, callback.puts, "pending Put triggers OnPut, requeue must not")
	assert.Equal(t, []any{"job"}, callback.gets)
	assert.Equal(t, []any{"job"}, callback.dones)

	value, err = q.Get() // OnGet ×2
	assert.NoError(t, err)

	q.Done(value) // 生命周期结束 → OnDone ×2

	assert.Equal(t, []any{"job", "job"}, callback.puts)
	assert.Equal(t, []any{"job", "job"}, callback.gets)
	assert.Equal(t, []any{"job", "job"}, callback.dones)
}

// TestQueueImpl_Idempotent_DoneOnQueuedElementIsNoOp 验证 R3 语义变更：
// 旧语义下 Done 对 state 做 TryRemove，会把仍在队列中（未被 Get）元素的在队去重标记误清；
// 新语义下 Done 对非处理中元素为安全 no-op，在队去重保持有效。
func TestQueueImpl_Idempotent_DoneOnQueuedElementIsNoOp(t *testing.T) {
	config := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("job"))

	// 元素仍在队列中（未被 Get），Done 必须无副作用。
	q.Done("job")

	assert.Equal(t, 1, q.Len(), "Done must not remove a queued element from list")
	assert.ErrorIs(t, q.Put("job"), ErrElementAlreadyExist, "dedup must remain effective after Done on queued element")

	value, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "job", value)
}

// TestQueueImpl_Idempotent_PendingPutOnce 验证挂起标记的幂等性：
// 处理中同值多次 Put 均返回 nil，最终只重入队一份。
func TestQueueImpl_Idempotent_PendingPutOnce(t *testing.T) {
	config := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("job"))

	value, err := q.Get()
	assert.NoError(t, err)

	assert.NoError(t, q.Put(value))
	assert.NoError(t, q.Put(value), "repeated pending Put must stay idempotent")
	assert.NoError(t, q.Put(value))

	q.Done(value)
	assert.Equal(t, 1, q.Len(), "pending marker must collapse to a single requeue")
}

// TestQueueImpl_Idempotent_DoneAfterShutdown 验证 Closed+Done 边界：
// Get 后挂起 Put 已在 state 留下标记，Shutdown 之后调用 Done 不得重新入队（元素丢弃），
// 与 Shutdown 清空语义一致，且不 panic。
func TestQueueImpl_Idempotent_DoneAfterShutdown(t *testing.T) {
	config := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(config)

	assert.NoError(t, q.Put("job"))

	value, err := q.Get()
	assert.NoError(t, err)

	// 挂起标记：元素处理中且被再次 Put（此时 state 含 "job"）。
	assert.NoError(t, q.Put(value))

	q.Shutdown()

	q.Done(value) // 不得 panic，不得重入队。

	assert.Equal(t, 0, q.Len(), "no requeue after shutdown")
	assert.True(t, q.IsClosed())
}

// TestQueueImpl_Idempotent_ProcessingInvariant 逐阶段验证双集合不变式：
// state 含 v ⟺ v 在 list ∨ (v 处理中且被再次 Put)。
func TestQueueImpl_Idempotent_ProcessingInvariant(t *testing.T) {
	config := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(config)
	defer q.Shutdown()

	impl := q.(*queueImpl)

	// 在队：state 含 v，processing 空。
	assert.NoError(t, q.Put("job"))
	assert.Equal(t, []any{"job"}, impl.state.List())
	assert.Equal(t, 0, impl.processing.Len())

	// Get 后：双集合搬运，state 空，processing 含 v。
	v, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, 0, impl.state.Len())
	assert.Equal(t, 1, impl.processing.Len())

	// 挂起 Put：v 回到 state（标记），仍在 processing。
	assert.NoError(t, q.Put(v))
	assert.Equal(t, []any{"job"}, impl.state.List())
	assert.Equal(t, 1, impl.processing.Len())

	// Done：重入队，v 留在 state，离开 processing。
	q.Done(v)
	assert.Equal(t, []any{"job"}, impl.state.List())
	assert.Equal(t, 0, impl.processing.Len())
	assert.Equal(t, 1, q.Len())

	// 完整结束：双集合皆空。
	v, err = q.Get()
	assert.NoError(t, err)
	q.Done(v)
	assert.Equal(t, 0, impl.state.Len())
	assert.Equal(t, 0, impl.processing.Len())
	assert.Equal(t, 0, q.Len())
}

// TestQueueImpl_Idempotent_ShutdownClearsProcessing 验证 Shutdown 一并清理 processing 集合
// （与 list/state 清理对齐），Get 后的处理中残留项不泄漏。
func TestQueueImpl_Idempotent_ShutdownClearsProcessing(t *testing.T) {
	config := NewQueueConfig().WithValueIdempotent()
	q := NewQueue(config)

	assert.NoError(t, q.Put("job"))
	_, err := q.Get()
	assert.NoError(t, err)

	impl := q.(*queueImpl)
	assert.Equal(t, 1, impl.processing.Len())

	q.Shutdown()

	assert.Equal(t, 0, impl.processing.Len(), "processing must be cleaned on shutdown")
	assert.Equal(t, 0, impl.state.Len(), "state must be cleaned on shutdown")
}

// TestQueueImpl_InFlight 验证 G5：幂等模式下 InFlight 暴露“已 Get 未 Done”的
// 处理中元素快照，且随 Get/Done 实时增减。
func TestQueueImpl_InFlight(t *testing.T) {
	q := NewQueue(NewQueueConfig().WithValueIdempotent())
	defer q.Shutdown()

	iq, ok := q.(InFlightQueue)
	assert.True(t, ok, "idempotent queue must satisfy InFlightQueue")

	assert.Empty(t, iq.InFlight(), "no in-flight element before any Get")

	assert.NoError(t, q.Put("task-a"))
	assert.NoError(t, q.Put("task-b"))

	value, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "task-a", value)
	assert.Equal(t, []any{"task-a"}, iq.InFlight())

	value, err = q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "task-b", value)
	assert.Len(t, iq.InFlight(), 2)

	q.Done("task-a")
	assert.Equal(t, []any{"task-b"}, iq.InFlight())

	q.Done("task-b")
	assert.Empty(t, iq.InFlight())
}

// TestQueueImpl_InFlight_NonIdempotentReturnsNil 验证非幂等模式不维护 per-value
// 处理中集合，InFlight 恒返回 nil；queueImpl 仍满足 InFlightQueue，
// 使调用方无需区分配置即可统一类型断言。
func TestQueueImpl_InFlight_NonIdempotentReturnsNil(t *testing.T) {
	q := NewQueue(nil)
	defer q.Shutdown()

	iq, ok := q.(InFlightQueue)
	assert.True(t, ok, "queueImpl must satisfy InFlightQueue regardless of config")
	assert.Nil(t, iq.InFlight())

	assert.NoError(t, q.Put("task"))
	_, err := q.Get()
	assert.NoError(t, err)
	assert.Nil(t, iq.InFlight(), "non-idempotent queue has no per-value in-flight tracking")
}

// TestQueueImpl_InFlight_CopyMutation 验证快照为副本：篡改返回值不影响内部状态。
func TestQueueImpl_InFlight_CopyMutation(t *testing.T) {
	q := NewQueue(NewQueueConfig().WithValueIdempotent())
	defer q.Shutdown()

	assert.NoError(t, q.Put("task"))
	_, err := q.Get()
	assert.NoError(t, err)

	iq := q.(InFlightQueue)
	snapshot := iq.InFlight()
	snapshot[0] = "tampered"

	assert.Equal(t, []any{"task"}, iq.InFlight(), "mutating the snapshot must not affect internal state")
}

// TestQueueImpl_InFlight_SnapshotDetached 验证快照脱锁：取快照后并发
// Put/Get/Done 不被阻塞、不死锁。
func TestQueueImpl_InFlight_SnapshotDetached(t *testing.T) {
	q := NewQueue(NewQueueConfig().WithValueIdempotent())
	defer q.Shutdown()

	assert.NoError(t, q.Put("seed"))
	_, err := q.Get() // seed 进入处理中，保证快照非空
	assert.NoError(t, err)

	assert.NotEmpty(t, q.(InFlightQueue).InFlight())

	stop := make(chan struct{})
	trafficDone := make(chan struct{})
	go func() {
		defer close(trafficDone)
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			value := i
			if q.Put(value) == nil {
				if got, err := q.Get(); err == nil {
					q.Done(got)
				}
			}
		}
	}()

	snapDone := make(chan struct{})
	go func() {
		defer close(snapDone)
		for i := 0; i < 200; i++ {
			_ = q.(InFlightQueue).InFlight()
		}
	}()

	select {
	case <-snapDone:
	case <-time.After(3 * time.Second):
		close(stop)
		<-trafficDone
		t.Fatal("InFlight blocked while concurrent Put/Get/Done traffic is running")
	}

	close(stop)
	<-trafficDone
}
