package workqueue

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestDelayingQueueImpl_PutWithDelay_DrainWindow_RejectedUnderLock 钉住 P1-1：
// drain 窗口内 PutWithDelay 必须被锁内 draining 复查拒绝，不得静默接受后丢弃。
func TestDelayingQueueImpl_PutWithDelay_DrainWindow_RejectedUnderLock(t *testing.T) {
	config := NewDelayingQueueConfig()
	config.WithDrainTracking()
	q := NewDelayingQueue(config).(*delayingQueueImpl)

	// blocker：已 Get 未 Done 的在途项，drainTracking 下 drain 判定持续阻塞，
	// 保证“drain 已开始但未完成”的时序由测试掌控。
	assert.NoError(t, q.Put("blocker"))
	blocker, err := q.Get()
	assert.NoError(t, err)

	// 持有堆锁，冻结生产者在锁外检查之后、锁内复查之前。
	q.lock.Lock()

	putErr := make(chan error, 1)
	go func() {
		putErr <- q.PutWithDelay("victim", 3600_000)
	}()
	// 等待生产者通过锁外检查并阻塞在 q.lock 上。
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	drainDone := make(chan error, 1)
	go func() {
		drainDone <- q.ShutdownWithDrain(ctx)
	}()
	// 等待 draining 置位。
	time.Sleep(100 * time.Millisecond)
	assert.True(t, q.IsDraining(), "drain must be in progress (blocked by in-flight blocker)")

	q.lock.Unlock()

	assert.ErrorIs(t, <-putErr, ErrQueueIsClosed,
		"PutWithDelay acquiring the lock inside the drain window must be rejected by the under-lock draining recheck, not silently accepted")

	q.Done(blocker)
	assert.NoError(t, <-drainDone)

	assert.Zero(t, q.DiscardedDelayed(),
		"window-rejected item must never reach the heap and be silently discarded at close")
	leaked := 0
	q.HeapRange(func(any, int64) bool {
		leaked++
		return true
	})
	assert.Zero(t, leaked)
	assert.True(t, q.IsClosed())
}

// TestTimerQueueImpl_PutAt_DrainWindow_RejectedUnderLock P1-1 timer 侧钉住：
// drain 窗口内 PutAt 必须被锁内 draining 复查拒绝。
func TestTimerQueueImpl_PutAt_DrainWindow_RejectedUnderLock(t *testing.T) {
	config := NewTimerQueueConfig()
	config.WithDrainTracking()
	q := NewTimerQueue(config).(*timerQueueImpl)

	assert.NoError(t, q.Put("blocker"))
	blocker, err := q.Get()
	assert.NoError(t, err)

	q.lock.Lock()

	putErr := make(chan error, 1)
	go func() {
		putErr <- q.PutAt("victim", time.Now().Add(time.Hour))
	}()
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	drainDone := make(chan error, 1)
	go func() {
		drainDone <- q.ShutdownWithDrain(ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	assert.True(t, q.IsDraining(), "drain must be in progress (blocked by in-flight blocker)")

	q.lock.Unlock()

	assert.ErrorIs(t, <-putErr, ErrQueueIsClosed,
		"PutAt acquiring the lock inside the drain window must be rejected by the under-lock draining recheck, not silently accepted")

	q.Done(blocker)
	assert.NoError(t, <-drainDone)

	leaked := 0
	q.HeapRange(func(any, int64) bool {
		leaked++
		return true
	})
	assert.Zero(t, leaked)
	assert.True(t, q.IsClosed())
}

// TestDelayingQueueImpl_PutWithDelay_ShutdownWithDrain_Concurrent P1-1 并发压力钉住：
// 多生产者 PutWithDelay(delay=0) 与 ShutdownWithDrain 并发，断言 accepted==delivered
// 且 DiscardedDelayed==0。防空转屏障保证每轮有真实基数。
func TestDelayingQueueImpl_PutWithDelay_ShutdownWithDrain_Concurrent(t *testing.T) {
	const (
		rounds    = 64
		producers = 8
		perProd   = 50
	)

	acceptedTotal := 0
	minDrain := time.Duration(0)

	for round := 0; round < rounds; round++ {
		config := NewDelayingQueueConfig()
		config.WithDrainTracking()
		q := NewDelayingQueue(config).(*delayingQueueImpl)

		// blocker：在途项，drain 首次判定与放行由主 goroutine 掌控。
		assert.NoError(t, q.Put(fmt.Sprintf("round-%d-blocker", round)))
		blocker, err := q.Get()
		assert.NoError(t, err)

		var mu sync.Mutex
		delivered := 0
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
				delivered++
				mu.Unlock()
				q.Done(value)
			}
		}()

		accepted := make([]int, producers)
		firstAccepted := make(chan struct{}, producers)
		var wg sync.WaitGroup
		wg.Add(producers)
		for i := 0; i < producers; i++ {
			go func(idx int) {
				defer wg.Done()
				for j := 0; j < perProd; j++ {
					err := q.PutWithDelay(fmt.Sprintf("round-%d-item-%d-%d", round, idx, j), 0)
					if err != nil {
						if !errors.Is(err, ErrQueueIsClosed) {
							t.Errorf("round %d producer %d: PutWithDelay returned unexpected error: %v", round, idx, err)
						}
						return
					}
					if j == 0 {
						firstAccepted <- struct{}{}
					}
					accepted[idx]++
				}
			}(i)
		}

		// 屏障：等待全部生产者首项已接受，保证核算有真实基数。
		for i := 0; i < producers; i++ {
			<-firstAccepted
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		drainErr := make(chan error, 1)
		drainDur := make(chan time.Duration, 1)
		go func() {
			start := time.Now()
			err := q.ShutdownWithDrain(ctx)
			drainDur <- time.Since(start)
			drainErr <- err
		}()

		// 让 drain 首次判定命中 blocker 后再放行。
		time.Sleep(2 * drainPollInterval)
		q.Done(blocker)

		err = <-drainErr
		dur := <-drainDur
		cancel()
		assert.NoError(t, err, "round %d: drain must complete within timeout", round)

		wg.Wait()
		<-consumerDone

		total := 0
		for _, n := range accepted {
			total += n
		}
		mu.Lock()
		got := delivered
		mu.Unlock()

		assert.Positive(t, total,
			"round %d: barrier must guarantee real accepted items (vacuous 0==0 accounting is a defect)", round)
		assert.Equal(t, total, got,
			"round %d: every nil-return PutWithDelay must be delivered under drain semantics (no silent loss)", round)
		assert.Zero(t, q.DiscardedDelayed(),
			"round %d: due (delay=0) items must never be discarded at close", round)

		acceptedTotal += total
		if round == 0 || dur < minDrain {
			minDrain = dur
		}
		if dur < drainPollInterval {
			t.Logf("round %d: WARNING drain short-circuited: %v < drainPollInterval %v (ticker path not exercised)",
				round, dur, drainPollInterval)
		}
	}

	assert.Positive(t, acceptedTotal,
		"concurrent accounting must cover real accepted items across rounds (idle-spin rounds cannot pass)")
	t.Logf("rounds=%d acceptedTotal=%d minDrain=%v (drainPollInterval=%v)", rounds, acceptedTotal, minDrain, drainPollInterval)
}

// TestTimerQueueImpl_PutAt_ShutdownWithDrain_Concurrent P1-1 timer 侧并发压力：
// PutAt/PutAfter 与 ShutdownWithDrain 并发，钉住窗口闭合后不变量：
// 无幽灵投递、堆零残留、无数据竞争。
func TestTimerQueueImpl_PutAt_ShutdownWithDrain_Concurrent(t *testing.T) {
	const (
		rounds    = 32
		producers = 8
		perProd   = 30
	)

	acceptedTotal := 0
	minDrain := time.Duration(0)

	for round := 0; round < rounds; round++ {
		config := NewTimerQueueConfig()
		config.WithDrainTracking()
		q := NewTimerQueue(config).(*timerQueueImpl)

		// blocker：在途项，drain 首次判定与放行由主 goroutine 掌控。
		assert.NoError(t, q.Put(fmt.Sprintf("round-%d-blocker", round)))
		blocker, err := q.Get()
		assert.NoError(t, err)

		var mu sync.Mutex
		deliveredSet := make(map[string]bool)
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
				deliveredSet[value.(string)] = true
				mu.Unlock()
				q.Done(value)
			}
		}()

		acceptedVals := make([][]string, producers)
		firstAccepted := make(chan struct{}, producers)
		var wg sync.WaitGroup
		wg.Add(producers)
		for i := 0; i < producers; i++ {
			go func(idx int) {
				defer wg.Done()
				for j := 0; j < perProd; j++ {
					value := fmt.Sprintf("round-%d-item-%d-%d", round, idx, j)
					var err error
					if j%2 == 0 {
						err = q.PutAt(value, time.Now().Add(2*time.Millisecond))
					} else {
						err = q.PutAfter(value, 2*time.Millisecond)
					}
					if err != nil {
						if !errors.Is(err, ErrQueueIsClosed) {
							t.Errorf("round %d producer %d: PutAt/PutAfter returned unexpected error: %v", round, idx, err)
						}
						return
					}
					if j == 0 {
						firstAccepted <- struct{}{}
					}
					acceptedVals[idx] = append(acceptedVals[idx], value)
				}
			}(i)
		}

		// 屏障：等待全部生产者首项已接受。
		for i := 0; i < producers; i++ {
			<-firstAccepted
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		drainErr := make(chan error, 1)
		drainDur := make(chan time.Duration, 1)
		go func() {
			start := time.Now()
			err := q.ShutdownWithDrain(ctx)
			drainDur <- time.Since(start)
			drainErr <- err
		}()

		// 让 drain 首次判定命中 blocker 后放行。
		time.Sleep(2 * drainPollInterval)
		q.Done(blocker)

		err = <-drainErr
		dur := <-drainDur
		cancel()
		assert.NoError(t, err, "round %d: drain must complete within timeout", round)

		wg.Wait()
		<-consumerDone

		acceptedSet := make(map[string]bool)
		total := 0
		for _, vals := range acceptedVals {
			total += len(vals)
			for _, v := range vals {
				acceptedSet[v] = true
			}
		}
		mu.Lock()
		for v := range deliveredSet {
			if !acceptedSet[v] {
				t.Errorf("round %d: delivered value %q was never accepted (ghost delivery)", round, v)
			}
		}
		mu.Unlock()

		assert.Positive(t, total,
			"round %d: barrier must guarantee real accepted items (idle-spin rounds are a defect)", round)

		leaked := 0
		q.HeapRange(func(any, int64) bool {
			leaked++
			return true
		})
		assert.Zero(t, leaked, "round %d: heap must be empty after ShutdownWithDrain", round)
		assert.True(t, q.IsClosed())

		acceptedTotal += total
		if round == 0 || dur < minDrain {
			minDrain = dur
		}
		if dur < drainPollInterval {
			t.Logf("round %d: WARNING drain short-circuited: %v < drainPollInterval %v (ticker path not exercised)",
				round, dur, drainPollInterval)
		}
	}

	assert.Positive(t, acceptedTotal,
		"concurrent stress must cover real accepted items across rounds (idle-spin rounds cannot pass)")
	t.Logf("rounds=%d acceptedTotal=%d minDrain=%v (drainPollInterval=%v)", rounds, acceptedTotal, minDrain, drainPollInterval)
}
