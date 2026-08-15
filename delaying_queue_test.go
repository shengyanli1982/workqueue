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
	assert.Equal(t, []any(nil), dones, "Callback dones should be nil")
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

// TestDelayingQueueImpl_ShutdownLatency 验证 puller goroutine 在 heartbeat ticker 阻塞期间
// Shutdown 能够立即唤醒 puller 而无须等待最多 300ms。
func TestDelayingQueueImpl_ShutdownLatency(t *testing.T) {
	q := NewDelayingQueue(nil)

	// 等待 puller 进入 heartbeat.C 阻塞状态（队列为空，无 expired 元素）。
	time.Sleep(20 * time.Millisecond)

	start := time.Now()
	q.Shutdown()
	elapsed := time.Since(start)

	// Shutdown 应在远低于 300ms 内完成；以 100ms 作为安全阈值。
	assert.Less(t, elapsed, 100*time.Millisecond,
		"Shutdown should complete quickly without waiting for the 300ms heartbeat ticker")
}

// TestDelayingQueueImpl_PutWithDelay_RejectsAfterShutdown 回归测试：
// Shutdown 完成后 PutWithDelay 必须返回 ErrQueueIsClosed，且堆中无任何残留节点。
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
// PutWithDelay × Shutdown 竞态下验证零丢项可观测契约：
//  1. 每次 PutWithDelay 调用要么成功入堆（返回 nil），要么显式拒绝（ErrQueueIsClosed），
//     不允许出现"返回成功但节点进入已清空的堆"的静默丢失；
//  2. Shutdown 完成后堆中不得残留任何幽灵节点（锁外检查通过后持锁 Push 的竞态产物）。
//
// 修复前：锁外 IsClosed 检查与持锁 Push 之间无复查， producer 可能在 Shutdown
// 清空堆之后把节点 Push 进死堆——HeapRange 可观测到泄漏节点（RED）。
// 修复后：锁内复查 closed，命中即归还节点并返回 ErrQueueIsClosed，泄漏恒为零（GREEN）。
func TestDelayingQueueImpl_PutWithDelay_ShutdownRace_NoLeakedNodes(t *testing.T) {
	const (
		rounds      = 1024
		producers   = 16
		longDelayMs = int64(3600_000) // 1 小时：确保到期项不会被 puller 搬走，堆状态可用于事后检验
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

		// 与 producers 并发触发 Shutdown，制造"锁外检查通过 → Shutdown 清空 → 持锁 Push"竞态窗口。
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

// TestDelayingQueueImpl_DelayPrecision 验证 G10 精确 timer 的延迟偏差显著小于旧 300ms 轮询上限。
// 宽松断言防 flaky：20ms 延迟项在 200ms 内必达，且实际投递超出声明延迟的偏差 < 100ms。
// 修复前：搬运依赖 300ms 心跳轮询，到期后平均还要等待 ~150ms，偏差断言失败（RED）；
// 修复后：堆顶精确 timer 到期即搬运，偏差为毫秒级（GREEN）。
func TestDelayingQueueImpl_DelayPrecision(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	const declared = int64(20)

	// 先让搬运协程进入等待状态，避免与队列启动时刻重叠。
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

// TestDelayingQueueImpl_InstantWake 验证 G10 事件驱动的核心可观测收益：
// 空闲（空堆）队列被 PutWithDelay 经 notifyWake 即时唤醒并搬运，
// 不依赖 300ms 兜底 ticker 到期。
// 修复前：puller 停在 heartbeat.C 上，零延迟项要等到下一个 300ms tick 才被搬运（RED）；
// 修复后：PutWithDelay 唤醒 scheduler 立即搬运（GREEN）。
func TestDelayingQueueImpl_InstantWake(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	// 确保搬运协程已停在空堆等待（无活动 timer，仅 wake/ticker/closed）。
	time.Sleep(100 * time.Millisecond)

	start := time.Now()
	assert.NoError(t, q.PutWithDelay("wake-me", 0))

	value, err := waitForDelivery(t, q, 150*time.Millisecond)
	assert.NoError(t, err, "zero-delay item must be moved instantly via notifyWake, not the 300ms ticker")
	assert.Equal(t, "wake-me", value)
	assert.Less(t, time.Since(start), 150*time.Millisecond)
	q.Done(value)
}

// TestDelayingQueueImpl_CancelDelay 命中：堆内未到期项被移除，重复取消返回 false。
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

// TestDelayingQueueImpl_CancelDelay_NoDelivery 取消后跨越到期窗口不发生投递。
func TestDelayingQueueImpl_CancelDelay_NoDelivery(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	assert.NoError(t, q.PutWithDelay("cancel-me", 50))
	assert.True(t, q.CancelDelay("cancel-me"))

	// 等待跨越到期窗口：精确 timer 下若未取消，毫秒级即被搬运。
	time.Sleep(300 * time.Millisecond)

	_, err := q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty, "cancelled item must never be delivered")
	assert.Equal(t, 0, q.Len())
}

// TestDelayingQueueImpl_CancelDelay_AfterDelivery 已搬运进内层的项不可再取消（边界语义明确）。
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

// TestDelayingQueueImpl_CancelDelay_AfterShutdown 关停清空堆后取消一律返回 false。
func TestDelayingQueueImpl_CancelDelay_AfterShutdown(t *testing.T) {
	q := NewDelayingQueue(nil)

	assert.NoError(t, q.PutWithDelay("pending", 3600_000))
	q.Shutdown()

	assert.False(t, q.CancelDelay("pending"),
		"cancel must return false after Shutdown discarded the heap")
}

// TestDelayingQueueImpl_CancelDelay_NonComparableValue 不可比较值走 reflect.DeepEqual 匹配，
// 与 TimerQueue.Cancel 的 value 匹配语义对齐。
func TestDelayingQueueImpl_CancelDelay_NonComparableValue(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	assert.NoError(t, q.PutWithDelay([]int{1, 2, 3}, 3600_000))
	assert.True(t, q.CancelDelay([]int{1, 2, 3}), "DeepEqual-matching value must be cancelled")
	assert.Equal(t, 0, q.Len())
}

// TestDelayingQueueImpl_CancelDelay_Concurrent 高并发 PutWithDelay/CancelDelay/Shutdown
// 竞态安全压力测试（配合 -race 运行）：无竞态、关停后堆零残留。
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
// Shutdown 丢弃堆中未搬运的延迟项——不可被 Get、HeapRange 不可见、
// 内部 discardedDelayed 精确计数（本里程碑只计数不暴露，G4/G5 里程碑暴露）。
func TestDelayingQueueImpl_ShutdownDiscardsUndueItems(t *testing.T) {
	q := NewDelayingQueue(nil)

	assert.NoError(t, q.PutWithDelay("discard-1", 3600_000))
	assert.NoError(t, q.PutWithDelay("discard-2", 3600_000))
	assert.NoError(t, q.PutWithDelay("discard-3", 3600_000))

	q.Shutdown()

	// 未到期项不可被 Get（内层已关闭且堆已清空）。
	_, err := q.Get()
	assert.ErrorIs(t, err, ErrQueueIsClosed, "undue items must not be retrievable after Shutdown")

	// 堆清空，DelayingQueue.Len 归零。
	assert.Equal(t, 0, q.Len())
	leaked := 0
	q.HeapRange(func(value any, delay int64) bool {
		leaked++
		return true
	})
	assert.Zero(t, leaked)

	// discardedDelayed 计数丢弃的未搬运延迟项数量。
	impl := q.(*delayingQueueImpl)
	assert.Equal(t, int64(3), impl.discardedDelayed.Load(),
		"discardedDelayed must count exactly the undue items dropped at Shutdown")
}

// TestDelayingQueueImpl_BatchDelivery 验证单批 128 上限语义在 scheduler 重写后保留：
// 同一到期时刻超过 128 个元素时分多批全部搬运，不丢不重。
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
