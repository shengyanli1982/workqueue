package workqueue

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
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

// TestTimerQueue_Cancel_DuplicateComparableValue 同值重复调度后 Cancel 取消唯一节点（D4b）。
func TestTimerQueue_Cancel_DuplicateComparableValue(t *testing.T) {
	q := NewTimerQueue(nil)
	defer q.Shutdown()

	now := time.Now()
	assert.NoError(t, q.PutAt("dup", now.Add(80*time.Millisecond)))
	assert.NoError(t, q.PutAt("dup", now.Add(20*time.Millisecond)))

	assert.True(t, q.Cancel("dup"))

	time.Sleep(120 * time.Millisecond)
	_, err := q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty)
}

// TestTimerQueue_PutAt_Duplicate_RescheduleEarlierDeliversOnce 同值提前重排后恰投递一次（D4b）。
func TestTimerQueue_PutAt_Duplicate_RescheduleEarlierDeliversOnce(t *testing.T) {
	q := NewTimerQueue(nil)
	defer q.Shutdown()

	now := time.Now()
	assert.NoError(t, q.PutAt("dup", now.Add(80*time.Millisecond)))
	assert.NoError(t, q.PutAt("dup", now.Add(20*time.Millisecond)))

	time.Sleep(60 * time.Millisecond)
	value, err := q.Get()
	assert.NoError(t, err, "value should be delivered at the rescheduled 20ms deadline")
	assert.Equal(t, "dup", value)

	// 跨过原到期点后不得有第二次投递。
	time.Sleep(60 * time.Millisecond)
	_, err = q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty, "reschedule must not deliver a second copy at the old deadline")
}

// TestTimerQueue_PutAt_Duplicate_RescheduleLaterDeliversOnce 同值推迟到期后恰投递一次（D4b）。
func TestTimerQueue_PutAt_Duplicate_RescheduleLaterDeliversOnce(t *testing.T) {
	q := NewTimerQueue(nil)
	defer q.Shutdown()

	now := time.Now()
	assert.NoError(t, q.PutAt("dup", now.Add(30*time.Millisecond)))
	assert.NoError(t, q.PutAt("dup", now.Add(120*time.Millisecond)))

	time.Sleep(60 * time.Millisecond)
	_, err := q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty, "old earlier deadline must have been replaced, not delivered")

	value, err := waitQueueGet(t, q, 300*time.Millisecond)
	assert.NoError(t, err, "value should be delivered at the rescheduled 120ms deadline")
	assert.Equal(t, "dup", value)

	_, err = q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty, "reschedule must deliver exactly once")
}

// TestTimerQueue_PutAfter_Immediate_ReplacesPending 立即投递替换待投递同值节点（D4b）。
func TestTimerQueue_PutAfter_Immediate_ReplacesPending(t *testing.T) {
	q := NewTimerQueue(nil)
	defer q.Shutdown()

	assert.NoError(t, q.PutAfter("dup", 250*time.Millisecond))
	assert.NoError(t, q.PutAfter("dup", -10*time.Millisecond))

	value, err := waitQueueGet(t, q, 100*time.Millisecond)
	assert.NoError(t, err, "immediate PutAfter should deliver right away")
	assert.Equal(t, "dup", value)

	// 跨过原到期点后不得有第二次投递。
	time.Sleep(300 * time.Millisecond)
	_, err = q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty, "pending node must be replaced by the immediate delivery")
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

	items := make([]any, 0, 3)
	q.HeapRange(func(value any, _ int64) bool {
		items = append(items, value)
		return true
	})

	assert.Len(t, items, 3)
}

// TestTimerQueue_PutAt_ShutdownRace_NoLeakedNodes 并发压力测试（D3 同族）：
// PutAt/PutAfter × Shutdown 竞态下验证零丢项、堆无幽灵节点。
func TestTimerQueue_PutAt_ShutdownRace_NoLeakedNodes(t *testing.T) {
	const (
		rounds    = 1024
		producers = 16
	)

	for round := 0; round < rounds; round++ {
		q := NewTimerQueue(nil)

		var wg sync.WaitGroup
		errCh := make([]error, producers)
		wg.Add(producers)
		for i := 0; i < producers; i++ {
			go func(index int) {
				defer wg.Done()
				value := fmt.Sprintf("round-%d-item-%d", round, index)
				// 远到期：确保堆状态可用于事后检验。交替走 PutAt/PutAfter。
				if index%2 == 0 {
					errCh[index] = q.PutAt(value, time.Now().Add(time.Hour))
				} else {
					errCh[index] = q.PutAfter(value, time.Hour)
				}
			}(i)
		}

		// 与 producers 并发触发 Shutdown，制造竞态窗口。
		q.Shutdown()
		wg.Wait()

		for i, err := range errCh {
			if err != nil {
				assert.ErrorIs(t, err, ErrQueueIsClosed,
					"round %d producer %d: PutAt/PutAfter must return nil or ErrQueueIsClosed", round, i)
			}
		}

		leaked := 0
		q.HeapRange(func(value any, at int64) bool {
			leaked++
			return true
		})
		assert.Zero(t, leaked,
			"round %d: ghost nodes pushed after Shutdown drained the heap (silent item loss)", round)
		assert.Zero(t, q.Len(), "round %d: queue must be empty after Shutdown", round)

		assert.ErrorIs(t, q.PutAt("post-round", time.Now().Add(time.Hour)), ErrQueueIsClosed,
			"round %d: PutAt after Shutdown must return ErrQueueIsClosed", round)
		assert.ErrorIs(t, q.PutAfter("post-round", time.Hour), ErrQueueIsClosed,
			"round %d: PutAfter after Shutdown must return ErrQueueIsClosed", round)
	}
}

// TestTimerQueue_PutAt_Cancel_Shutdown_Concurrent 三路并发压力测试：
// 每值至多投递一次，Shutdown 后堆零残留，无数据竞争。
func TestTimerQueue_PutAt_Cancel_Shutdown_Concurrent(t *testing.T) {
	const (
		rounds  = 128
		workers = 8
		items   = 100
	)

	for round := 0; round < rounds; round++ {
		q := NewTimerQueue(nil)

		var (
			mu           sync.Mutex
			delivered    = make(map[string]int)  // value -> 观测到的投递次数
			cancelledHit = make(map[string]bool) // value -> Cancel 曾返回 true
		)

		// 消费者：关停前持续取出投递项并计数；队列关闭后 Get 返回 ErrQueueIsClosed 即退出。
		consumerDone := make(chan struct{})
		go func() {
			defer close(consumerDone)
			for {
				value, err := q.Get()
				if err != nil {
					if errors.Is(err, ErrQueueIsClosed) {
						return
					}
					time.Sleep(time.Millisecond)
					continue
				}
				mu.Lock()
				delivered[value.(string)]++
				mu.Unlock()
				q.Done(value)
			}
		}()

		var wg sync.WaitGroup
		wg.Add(workers * 2)
		for i := 0; i < workers; i++ {
			go func(index int) {
				defer wg.Done()
				for j := 0; j < items; j++ {
					value := fmt.Sprintf("round-%d-item-%d-%d", round, index, j)
					if j%2 == 0 {
						// 远到期：节点滞留堆中，参与 Cancel/Shutdown 竞态。
						_ = q.PutAt(value, time.Now().Add(time.Hour))
					} else {
						// 短到期：参与搬运路径，与 Cancel/Shutdown 交织。
						_ = q.PutAfter(value, time.Duration(1+j%5)*time.Millisecond)
					}
					if j%10 == 0 {
						runtime.Gosched() // 让出调度，拉长与 Shutdown 的重叠窗口
					}
				}
			}(i)
			go func(index int) {
				defer wg.Done()
				for j := 0; j < items; j++ {
					value := fmt.Sprintf("round-%d-item-%d-%d", round, index, j)
					if q.Cancel(value) {
						mu.Lock()
						cancelledHit[value] = true
						mu.Unlock()
					}
				}
			}(i)
		}

		// 与活跃的生产者/取消者并发触发 Shutdown（偶数轮立即、奇数轮 1ms 抖动），
		// 交错覆盖"putter 先于/后于 closeNow 清空"两类时序。
		if round%2 == 1 {
			time.Sleep(time.Millisecond)
		}
		q.Shutdown()
		wg.Wait()
		<-consumerDone

		mu.Lock()
		for value, hit := range cancelledHit {
			if !hit {
				continue
			}
			assert.Zero(t, delivered[value],
				"round %d: value %q delivered after Cancel returned true (double delivery)", round, value)
		}
		for value, count := range delivered {
			assert.LessOrEqual(t, count, 1,
				"round %d: value %q delivered %d times (must be at most once)", round, value, count)
		}
		mu.Unlock()

		leaked := 0
		q.HeapRange(func(value any, at int64) bool {
			leaked++
			return true
		})
		assert.Zero(t, leaked, "round %d: heap must be drained after Shutdown", round)
		assert.Zero(t, q.Len(), "round %d: queue must be empty after Shutdown", round)

		assert.False(t, q.Cancel(fmt.Sprintf("round-%d-item-0-0", round)),
			"round %d: Cancel after Shutdown must return false", round)
		assert.False(t, q.Cancel("never-inserted"),
			"round %d: Cancel of absent value must return false", round)
	}
}

func waitTimerQueueGet(t *testing.T, q Queue, timeout time.Duration) (any, error) {
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
