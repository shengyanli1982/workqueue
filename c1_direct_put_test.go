package workqueue

// c1_direct_put_test.go 钉住 C-1 优化（PutWithDelay delay≤0 直通内层队列）的边缘语义。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// waitDraining 轮询等待 draining 置位，超时判定失败。
func waitDraining(t *testing.T, q DrainingQueue) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for !q.IsDraining() {
		if time.Now().After(deadline) {
			t.Fatal("draining flag was not set within 2s")
		}
		time.Sleep(time.Millisecond)
	}
}

// TestC1_PutWithDelay_ZeroDelay_ImmediatelyGettable delay=0 直通后立即可 Get，不经延迟堆。
func TestC1_PutWithDelay_ZeroDelay_ImmediatelyGettable(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	assert.NoError(t, q.PutWithDelay("zero", 0))

	heapCount := 0
	q.HeapRange(func(any, int64) bool {
		heapCount++
		return true
	})
	assert.Zero(t, heapCount, "direct-path item must never enter the delay heap")

	v, err := q.Get()
	assert.NoError(t, err, "delay=0 item must be gettable immediately, without waiting for scheduler transport")
	assert.Equal(t, "zero", v)
	q.Done(v)
}

// TestC1_PutWithDelay_NegativeDelay_ImmediatelyGettable 负 delay 同 0 语义，直通后立即可 Get。
func TestC1_PutWithDelay_NegativeDelay_ImmediatelyGettable(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	assert.NoError(t, q.PutWithDelay("negative", -100))

	v, err := q.Get()
	assert.NoError(t, err, "negative-delay item must be gettable immediately, without waiting for scheduler transport")
	assert.Equal(t, "negative", v)
	q.Done(v)
}

// TestC1_PutWithDelay_ZeroDelay_RejectedWhileDraining drain 期间直通分支同样被拒绝。
func TestC1_PutWithDelay_ZeroDelay_RejectedWhileDraining(t *testing.T) {
	q := NewDelayingQueue(nil)

	// 保留在队项不消费，使队列停留在 draining 状态。
	assert.NoError(t, q.Put("blocker"))

	errCh := make(chan error, 1)
	dq := drainableOf(t, q)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		errCh <- dq.ShutdownWithDrain(ctx)
	}()

	waitDraining(t, q.(DrainingQueue))

	assert.ErrorIs(t, q.PutWithDelay("zero-during-drain", 0), ErrQueueIsClosed,
		"delay=0 put must be rejected by the entry precheck while draining")
	assert.ErrorIs(t, q.PutWithDelay("neg-during-drain", -5), ErrQueueIsClosed,
		"negative-delay put must be rejected by the entry precheck while draining")

	// 消费阻塞项，drain 应干净完成。
	v, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "blocker", v)
	q.Done(v)

	assert.NoError(t, <-errCh, "drain must complete once the blocker item is consumed")
}

// TestC1_PutWithDelay_ZeroDelay_DeliveredUnderDrain drain 前 accepted 的直通项必须被投递。
func TestC1_PutWithDelay_ZeroDelay_DeliveredUnderDrain(t *testing.T) {
	q := NewDelayingQueue(nil)

	assert.NoError(t, q.PutWithDelay("accepted-zero", 0))

	errCh := make(chan error, 1)
	dq := drainableOf(t, q)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		errCh <- dq.ShutdownWithDrain(ctx)
	}()

	// 项已同步位于内层队列，drain 期间 Get 即命中。
	v, err := q.Get()
	assert.NoError(t, err, "accepted 0-delay item must be delivered during drain")
	assert.Equal(t, "accepted-zero", v)
	q.Done(v)

	assert.NoError(t, <-errCh, "drain must complete after the accepted item is consumed")
}

// TestC1_CancelDelay_ZeroDelayDirectItem_NotFound 直通项不进堆，CancelDelay 返回 false。
func TestC1_CancelDelay_ZeroDelayDirectItem_NotFound(t *testing.T) {
	q := NewDelayingQueue(nil)
	defer q.Shutdown()

	assert.NoError(t, q.PutWithDelay("direct-zero", 0))

	assert.NotPanics(t, func() {
		assert.False(t, q.CancelDelay("direct-zero"),
			"direct-path item never enters the heap; CancelDelay must report not-found")
	})

	v, err := q.Get()
	assert.NoError(t, err, "cancel attempt must not consume or corrupt the direct-path item")
	assert.Equal(t, "direct-zero", v)
	q.Done(v)
}

// TestC1_PutWithDelay_ZeroDelay_OnPutExactlyOnce 直通项 OnPut 恰好一次，OnDelay 不触发。
func TestC1_PutWithDelay_ZeroDelay_OnPutExactlyOnce(t *testing.T) {
	callback := &testDelayingQueueCallback{}
	config := NewDelayingQueueConfig().WithCallback(callback)
	q := NewDelayingQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.PutWithDelay("cb-zero", 0))

	v, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "cb-zero", v)
	q.Done(v)

	// 静默窗口：若 scheduler 被错误卷入可观测。
	time.Sleep(50 * time.Millisecond)

	callback.Lock()
	puts := append([]any(nil), callback.puts...)
	delays := append([]any(nil), callback.delays...)
	callback.Unlock()

	assert.Equal(t, []any{"cb-zero"}, puts, "OnPut must fire exactly once for the direct-path item")
	assert.Empty(t, delays,
		"direct-path (delay≤0) item bypasses heap acceptance; OnDelay must not fire (callback surface ≡ Put)")
}

// TestC1_PutWithDelay_ZeroDelay_IdempotentDedup 幂等模式下直通路径去重与搬运路径一致。
func TestC1_PutWithDelay_ZeroDelay_IdempotentDedup(t *testing.T) {
	config := NewDelayingQueueConfig()
	config.WithValueIdempotent()
	q := NewDelayingQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.PutWithDelay("dup", 0))

	assert.ErrorIs(t, q.PutWithDelay("dup", 0), ErrElementAlreadyExist,
		"direct path must surface the inner idempotent dedup synchronously")

	assert.Equal(t, 1, q.Len(), "queue must hold exactly one copy")

	v, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "dup", v)
	q.Done(v)

	_, err = q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty, "deduped copy must never be delivered")
}
