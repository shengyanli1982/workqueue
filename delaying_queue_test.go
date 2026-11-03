package workqueue

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// waitForDelivery 轮询 Get 直到取出元素或超时，供事件驱动唤醒相关测试复用。
func waitForDelivery(t *testing.T, q DelayingQueue, timeout time.Duration) (any, error) {
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

var DELAYDUCRATION = int64(150)

func TestDelayingQueueImpl_PutWithDelay(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	err := q.PutWithDelay("test1", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test2", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test3", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	time.Sleep(time.Second)

	assert.Equal(t, 3, q.Len(), "Queue length should be 3")
	assert.Equal(t, []any{"test1", "test2", "test3"}, q.Values(), "Queue values should be [test1, test2, test3]")
}

func TestDelayingQueueImpl_PutWithDelay_Closed(t *testing.T) {
	q := NewDelayingQueue(nil)
	q.Shutdown()

	err := q.PutWithDelay("test", 0)
	assert.ErrorIs(t, err, ErrQueueIsClosed, "Put should return ErrQueueIsClosed")

	time.Sleep(time.Second)
}

func TestDelayingQueueImpl_PutWithDelay_Nil(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	err := q.PutWithDelay(nil, 0)
	assert.ErrorIs(t, err, ErrElementIsNil, "Put should return ErrElementIsNil")

	time.Sleep(time.Second)
}

func TestDelayingQueueImpl_PutWithDelay_Parallel(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	count := 1000

	wg := sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			err := q.PutWithDelay("test", DELAYDUCRATION)
			assert.NoError(t, err, "Put should not return an error")
		}()
	}
	wg.Wait()

	time.Sleep(time.Second)

	assert.Equal(t, count, q.Len(), "Queue length should be 1000")
}

func TestDelayingQueueImpl_HeapRange(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	err := q.PutWithDelay("test1", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test2", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test3", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	values := []any{}
	q.HeapRange(func(value any, _ int64) bool {
		values = append(values, value)
		return true
	})

	time.Sleep(time.Second)

	assert.Equal(t, []any{"test1", "test2", "test3"}, values, "Queue values should be [test1, test2, test3]")
}

func TestDelayingQueueImpl_HeapRange_Closed(t *testing.T) {
	q := NewDelayingQueue(nil)

	err := q.PutWithDelay("test1", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test2", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test3", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	q.Shutdown()

	values := []any{}
	q.HeapRange(func(value any, _ int64) bool {
		values = append(values, value)
		return true
	})

	assert.Equal(t, []any{}, values, "Values should be []")
}

type testDelayingQueueCallback struct {
	sync.Mutex
	puts, gets, dones, delays, errors []any
}

func (c *testDelayingQueueCallback) OnPut(value any) {
	c.Lock()
	defer c.Unlock()
	c.puts = append(c.puts, value)
}

func (c *testDelayingQueueCallback) OnGet(value any) {
	c.Lock()
	defer c.Unlock()
	c.gets = append(c.gets, value)
}

func (c *testDelayingQueueCallback) OnDone(value any) {
	c.Lock()
	defer c.Unlock()
	c.dones = append(c.dones, value)
}

func (c *testDelayingQueueCallback) OnDelay(value any, delay int64) {
	c.Lock()
	defer c.Unlock()
	c.delays = append(c.delays, value)
}

func (c *testDelayingQueueCallback) OnPullError(value any, err error) {
	c.Lock()
	defer c.Unlock()
	c.errors = append(c.errors, value)
}

func TestDelayingQueueImpl_Callback(t *testing.T) {
	callback := &testDelayingQueueCallback{}
	config := NewDelayingQueueConfig().WithCallback(callback)
	q := NewDelayingQueue(config)
	defer q.Shutdown()

	err := q.PutWithDelay("test1", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test2", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay("test3", DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	time.Sleep(time.Second)

	err = q.Put("test4")
	assert.NoError(t, err, "Put should not return an error")

	v, err := q.Get()
	assert.NoError(t, err, "Get should not return an error")
	assert.Equal(t, "test1", v, "Get value should be test1")

	q.Done(v)

	callback.Lock()
	delays := callback.delays
	puts := callback.puts
	gets := callback.gets
	dones := callback.dones
	errors := callback.errors
	callback.Unlock()

	assert.Equal(t, []any{"test1", "test2", "test3"}, delays, "Callback delays should be [test1, test2, test3]")
	assert.Equal(t, []any{"test1", "test2", "test3", "test4"}, puts, "Callback puts should be [test1, test2, test3, test4]")
	assert.Equal(t, []any{"test1"}, gets, "Callback gets should be [test1]")
	assert.Equal(t, []any{"test1"}, dones, "Callback dones should be [test1]")
	assert.Equal(t, []any(nil), errors, "Callback errors should be []")
}

type testAccNode struct {
	value any
	ts    int64
}

func TestDelayingQueueImpl_Accuracy(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	err := q.PutWithDelay(&testAccNode{value: "test1", ts: time.Now().UnixMilli()}, DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay(&testAccNode{value: "test2", ts: time.Now().UnixMilli()}, DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	err = q.PutWithDelay(&testAccNode{value: "test3", ts: time.Now().UnixMilli()}, DELAYDUCRATION)
	assert.NoError(t, err, "Put should not return an error")

	time.Sleep(time.Second)

	assert.Equal(t, 3, q.Len(), "Queue length should be 3")

	values := q.Values()
	for i, v := range values {
		node := v.(*testAccNode)
		assert.Equal(t, fmt.Sprintf("test%d", i+1), node.value, fmt.Sprintf("Value should be test%d", i+1))
		assert.True(t, time.Now().UnixMilli()-node.ts > DELAYDUCRATION, "Delay duration should be greater than 150ms")
	}
}

func TestDelayingQueueImpl_ExtremeDelays(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	err := q.PutWithDelay("zero-delay", 0)
	assert.NoError(t, err, "Put with zero delay should not return an error")

	err = q.PutWithDelay("long-delay", 24*60*60*1000)
	assert.NoError(t, err, "Put with long delay should not return an error")

	time.Sleep(time.Second)

	v, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "zero-delay", v, "Zero delay item should be available immediately")
}

func TestDelayingQueueImpl_NegativeDelay(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	err := q.PutWithDelay("negative-delay", -100)
	assert.NoError(t, err, "Put with negative delay should not return an error")

	time.Sleep(time.Second)

	v, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "negative-delay", v, "Negative delay item should be available immediately")
}

func TestDelayingQueueImpl_DuplicateItems(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	err := q.PutWithDelay("duplicate", 100)
	assert.NoError(t, err)

	err = q.PutWithDelay("duplicate", 50)
	assert.NoError(t, err)

	err = q.PutWithDelay("duplicate", 150)
	assert.NoError(t, err)

	time.Sleep(time.Second)

	assert.Equal(t, 3, q.Len(), "Queue should contain all duplicate items")
}

func TestDelayingQueueImpl_ConcurrentShutdown(t *testing.T) {
	q := NewDelayingQueue(nil)

	var wg sync.WaitGroup
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func(index int) {
			defer wg.Done()
			_ = q.PutWithDelay(fmt.Sprintf("item-%d", index), DELAYDUCRATION)
		}(i)
	}

	time.Sleep(10 * time.Millisecond)
	q.Shutdown()

	wg.Wait()

	err := q.PutWithDelay("after-shutdown", DELAYDUCRATION)
	assert.ErrorIs(t, err, ErrQueueIsClosed, "Put after shutdown should return ErrQueueIsClosed")
}

// TestDelayingQueueImpl_ShutdownLatency 验证 Shutdown 能立即唤醒 puller，无须等待心跳。
func TestDelayingQueueImpl_ShutdownLatency(t *testing.T) {
	q := NewDelayingQueue(nil)

	// 等待 puller 进入心跳阻塞状态。
	time.Sleep(20 * time.Millisecond)

	start := time.Now()
	q.Shutdown()
	elapsed := time.Since(start)

	// Shutdown 应在远低于心跳周期内完成。
	assert.Less(t, elapsed, 100*time.Millisecond,
		"Shutdown should complete quickly without waiting for the 300ms heartbeat ticker")
}

// TestDelayingQueueImpl_PutWithDelay_RejectsAfterShutdown Shutdown 后 PutWithDelay 返回 ErrQueueIsClosed，堆无残留。
func TestDelayingQueueImpl_PutWithDelay_RejectsAfterShutdown(t *testing.T) {
	q := NewDelayingQueue(nil)

	assert.NoError(t, q.PutWithDelay("before", 3600_000))
	q.Shutdown()

	err := q.PutWithDelay("after", 3600_000)
	assert.ErrorIs(t, err, ErrQueueIsClosed, "PutWithDelay after Shutdown should return ErrQueueIsClosed")

	leaked := 0
	q.HeapRange(func(value any, delay int64) bool {
		leaked++
		return true
	})
	assert.Zero(t, leaked, "heap should not retain nodes after Shutdown")
}

// TestDelayingQueueImpl_PutWithDelay_ShutdownRace_NoLeakedNodes 并发压力测试（D3）：
// PutWithDelay × Shutdown 竞态下验证零丢项、堆无幽灵节点。
func TestDelayingQueueImpl_PutWithDelay_ShutdownRace_NoLeakedNodes(t *testing.T) {
	const (
		rounds      = 1024
		producers   = 16
		longDelayMs = int64(3600_000) // 1 小时：确保堆状态可用于事后检验
	)

	for round := 0; round < rounds; round++ {
		q := NewDelayingQueue(nil)

		var wg sync.WaitGroup
		errCh := make([]error, producers)
		wg.Add(producers)
		for i := 0; i < producers; i++ {
			go func(index int) {
				defer wg.Done()
				errCh[index] = q.PutWithDelay(fmt.Sprintf("round-%d-item-%d", round, index), longDelayMs)
			}(i)
		}

		// 与 producers 并发触发 Shutdown，制造竞态窗口。
		q.Shutdown()
		wg.Wait()

		for i, err := range errCh {
			if err != nil {
				assert.ErrorIs(t, err, ErrQueueIsClosed,
					"round %d producer %d: PutWithDelay must return nil or ErrQueueIsClosed", round, i)
			}
		}

		leaked := 0
		q.HeapRange(func(value any, delay int64) bool {
			leaked++
			return true
		})
		assert.Zero(t, leaked,
			"round %d: ghost nodes pushed after Shutdown drained the heap (silent item loss)", round)

		assert.ErrorIs(t, q.PutWithDelay("post-round", longDelayMs), ErrQueueIsClosed)
	}
}

// TestDelayingQueueImpl_DelayPrecision 验证 G10 精确 timer 延迟偏差：
// 20ms 延迟项在 200ms 内必达，偏差 < 100ms。
func TestDelayingQueueImpl_DelayPrecision(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	const declared = int64(20)

	// 等待搬运协程进入空闲状态。
	time.Sleep(50 * time.Millisecond)

	start := time.Now()
	assert.NoError(t, q.PutWithDelay("precise", declared))

	value, err := waitForDelivery(t, q, 200*time.Millisecond)
	assert.NoError(t, err, "20ms delay item must be delivered within 200ms")
	assert.Equal(t, "precise", value)

	overrun := time.Since(start) - time.Duration(declared)*time.Millisecond
	assert.Less(t, overrun, 100*time.Millisecond,
		"delivery overrun beyond declared delay must be far below the legacy 300ms polling bound")
	q.Done(value)
}

// TestDelayingQueueImpl_InstantWake 验证 G10 事件驱动：
// 空闲队列被 PutWithDelay 经 notifyWake 即时唤醒搬运。
func TestDelayingQueueImpl_InstantWake(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	// 确保搬运协程已进入空堆等待。
	time.Sleep(100 * time.Millisecond)

	start := time.Now()
	assert.NoError(t, q.PutWithDelay("wake-me", 0))

	value, err := waitForDelivery(t, q, 150*time.Millisecond)
	assert.NoError(t, err, "zero-delay item must be moved instantly via notifyWake, not the 300ms ticker")
	assert.Equal(t, "wake-me", value)
	assert.Less(t, time.Since(start), 150*time.Millisecond)
	q.Done(value)
}

// TestDelayingQueueImpl_CancelDelay 堆内未到期项被移除，重复取消返回 false。
func TestDelayingQueueImpl_CancelDelay(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	assert.NoError(t, q.PutWithDelay("cancel-me", 3600_000))
	assert.Equal(t, 1, q.Len())

	assert.True(t, q.CancelDelay("cancel-me"), "in-heap item must be cancelled")

	assert.Equal(t, 0, q.Len())
	assert.False(t, q.CancelDelay("cancel-me"), "second cancel on the same item must miss")
}

// TestDelayingQueueImpl_CancelDelay_NotFound 未入堆的 value 返回 false。
func TestDelayingQueueImpl_CancelDelay_NotFound(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	assert.False(t, q.CancelDelay("never-inserted"), "cancel on empty queue must return false")

	assert.NoError(t, q.PutWithDelay("other", 3600_000))
	assert.False(t, q.CancelDelay("never-inserted"), "cancel of an absent value must return false")
	assert.Equal(t, 1, q.Len(), "unrelated in-heap item must survive")
}

// TestDelayingQueueImpl_CancelDelay_NoDelivery 取消后不发生投递。
func TestDelayingQueueImpl_CancelDelay_NoDelivery(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	assert.NoError(t, q.PutWithDelay("cancel-me", 50))
	assert.True(t, q.CancelDelay("cancel-me"))

	// 等待跨越到期窗口。
	time.Sleep(300 * time.Millisecond)

	_, err := q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty, "cancelled item must never be delivered")
	assert.Equal(t, 0, q.Len())
}

// TestDelayingQueueImpl_CancelDelay_AfterDelivery 已搬运的项不可再取消。
func TestDelayingQueueImpl_CancelDelay_AfterDelivery(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	assert.NoError(t, q.PutWithDelay("delivered", 10))

	value, err := waitForDelivery(t, q, time.Second)
	assert.NoError(t, err)
	assert.Equal(t, "delivered", value)

	assert.False(t, q.CancelDelay("delivered"),
		"item already moved into the inner queue can no longer be cancelled")
	q.Done(value)
}

// TestDelayingQueueImpl_CancelDelay_AfterShutdown 关停后取消返回 false。
func TestDelayingQueueImpl_CancelDelay_AfterShutdown(t *testing.T) {
	q := NewDelayingQueue(nil)

	assert.NoError(t, q.PutWithDelay("pending", 3600_000))
	q.Shutdown()

	assert.False(t, q.CancelDelay("pending"),
		"cancel must return false after Shutdown discarded the heap")
}

// TestDelayingQueueImpl_CancelDelay_NonComparableValue 不可比较值走 reflect.DeepEqual 匹配。
func TestDelayingQueueImpl_CancelDelay_NonComparableValue(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	assert.NoError(t, q.PutWithDelay([]int{1, 2, 3}, 3600_000))
	assert.True(t, q.CancelDelay([]int{1, 2, 3}), "DeepEqual-matching value must be cancelled")
	assert.Equal(t, 0, q.Len())
}

// TestDelayingQueueImpl_CancelDelay_Concurrent 高并发竞态压力测试，验证关停后堆零残留。
func TestDelayingQueueImpl_CancelDelay_Concurrent(t *testing.T) {
	const (
		rounds  = 64
		workers = 8
		items   = 50
	)

	for round := 0; round < rounds; round++ {
		q := NewDelayingQueue(nil)

		var wg sync.WaitGroup
		wg.Add(workers * 2)
		for i := 0; i < workers; i++ {
			go func(index int) {
				defer wg.Done()
				for j := 0; j < items; j++ {
					_ = q.PutWithDelay(fmt.Sprintf("item-%d-%d", index, j), int64(1+j%5))
				}
			}(i)
			go func(index int) {
				defer wg.Done()
				for j := 0; j < items; j++ {
					_ = q.CancelDelay(fmt.Sprintf("item-%d-%d", index, j))
				}
			}(i)
		}

		time.Sleep(10 * time.Millisecond)
		q.Shutdown()
		wg.Wait()

		leaked := 0
		q.HeapRange(func(value any, delay int64) bool {
			leaked++
			return true
		})
		assert.Zero(t, leaked, "round %d: heap must be drained after Shutdown", round)
	}
}

// TestDelayingQueueImpl_ShutdownDiscardsUndueItems Q4a 关停丢弃契约：
// Shutdown 丢弃未到期延迟项，不可被 Get、HeapRange 不可见。
func TestDelayingQueueImpl_ShutdownDiscardsUndueItems(t *testing.T) {
	q := NewDelayingQueue(nil)

	assert.NoError(t, q.PutWithDelay("discard-1", 3600_000))
	assert.NoError(t, q.PutWithDelay("discard-2", 3600_000))
	assert.NoError(t, q.PutWithDelay("discard-3", 3600_000))

	q.Shutdown()

	_, err := q.Get()
	assert.ErrorIs(t, err, ErrQueueIsClosed, "undue items must not be retrievable after Shutdown")

	// 堆清空，Len 归零。
	assert.Equal(t, 0, q.Len())
	leaked := 0
	q.HeapRange(func(value any, delay int64) bool {
		leaked++
		return true
	})
	assert.Zero(t, leaked)

	// discardedDelayed 计数丢弃的延迟项。
	impl := q.(*delayingQueueImpl)
	assert.Equal(t, int64(3), impl.discardedDelayed.Load(),
		"discardedDelayed must count exactly the undue items dropped at Shutdown")
}

// TestDelayingQueueImpl_DiscardedDelayed #15b：DiscardedDelayed() 读取关停时丢弃的延迟项计数。
func TestDelayingQueueImpl_DiscardedDelayed(t *testing.T) {
	q := NewDelayingQueue(nil)

	// 可选接口断言。
	dq, ok := q.(interface{ DiscardedDelayed() int64 })
	assert.True(t, ok, "delaying queue must expose DiscardedDelayed via optional interface")

	// 关停前为 0。
	assert.Zero(t, dq.DiscardedDelayed(), "fresh queue must report zero discarded")

	// 放入未到期项后 Shutdown，按 Q4a 契约丢弃并计数。
	assert.NoError(t, q.PutWithDelay("undue", 3600_000))
	q.Shutdown()

	assert.GreaterOrEqual(t, dq.DiscardedDelayed(), int64(1),
		"DiscardedDelayed must count the undue item dropped at Shutdown")
}

// TestDelayingQueueImpl_BatchDelivery 验证超 128 项同到期时分批搬运，不丢不重。
func TestDelayingQueueImpl_BatchDelivery(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	const total = 300
	for i := 0; i < total; i++ {
		assert.NoError(t, q.PutWithDelay(i, 10))
	}

	deadline := time.Now().Add(3 * time.Second)
	seen := make(map[any]struct{}, total)
	for len(seen) < total && time.Now().Before(deadline) {
		value, err := q.Get()
		if err != nil {
			time.Sleep(2 * time.Millisecond)
			continue
		}
		seen[value] = struct{}{}
		q.Done(value)
	}

	assert.Equal(t, total, len(seen),
		"300 items due at the same instant must all be delivered across multiple 128-item batches")
}

// TestDelayingQueueImpl_Scheduler_SilentExitOnClosedInner 钉住 P2-2：
// scheduler 搬运时内层已关停应静默退出，不触发 OnPullError。
func TestDelayingQueueImpl_Scheduler_SilentExitOnClosedInner(t *testing.T) {
	callback := &testDelayingQueueCallback{}
	config := NewDelayingQueueConfig().WithCallback(callback)
	q := NewDelayingQueue(config).(*delayingQueueImpl)
	defer q.Shutdown()

	q.Queue.Shutdown()

	node := q.elementpool.Get()
	node.Value = "victim"
	node.Priority = toDelay(0)
	q.lock.Lock()
	q.sorting.Push(node)
	q.lock.Unlock()
	q.notifyWake()

	exited := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(exited)
	}()

	select {
	case <-exited:
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler must exit silently after inner Put fails with ErrQueueIsClosed")
	}

	callback.Lock()
	pullErrors := len(callback.errors)
	callback.Unlock()
	assert.Zero(t, pullErrors, "shutdown-window transport failure must not fire OnPullError")
}

// TestDelayingQueueImpl_Scheduler_EarlyExitFinalizesBatchResidue 钉住 P2-1：
// scheduler 早退时必须收尾批次剩余节点（递减 inTransit、归还节点池、计入 discardedDelayed）。
func TestDelayingQueueImpl_Scheduler_EarlyExitFinalizesBatchResidue(t *testing.T) {
	callback := &testDelayingQueueCallback{}
	config := NewDelayingQueueConfig().WithCallback(callback)
	q := NewDelayingQueue(config).(*delayingQueueImpl)
	defer q.Shutdown()

	q.Queue.Shutdown()

	const batch = 4
	q.lock.Lock()
	for i := 0; i < batch; i++ {
		node := q.elementpool.Get()
		node.Value = i
		node.Priority = toDelay(0)
		q.sorting.Push(node)
	}
	q.lock.Unlock()
	q.notifyWake()

	exited := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(exited)
	}()

	select {
	case <-exited:
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler must exit after inner Put fails with ErrQueueIsClosed")
	}

	assert.Zero(t, q.inTransit.Load(),
		"batch residue must be finalized (inTransit decremented) on ErrQueueIsClosed early exit")
	assert.Equal(t, int64(batch), q.discardedDelayed.Load(),
		"residue nodes must be counted in discardedDelayed (aligned with closeNow cleanup semantics)")

	callback.Lock()
	pullErrors := len(callback.errors)
	callback.Unlock()
	assert.Zero(t, pullErrors, "shutdown-window transport failure must not fire OnPullError")

	leaked := 0
	q.HeapRange(func(any, int64) bool {
		leaked++
		return true
	})
	assert.Zero(t, leaked, "heap must be empty after the whole batch was popped")
}
