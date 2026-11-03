package workqueue

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// drainableOf 断言队列实现 DrainableQueue。
func drainableOf(t *testing.T, q Queue) DrainableQueue {
	t.Helper()

	dq, ok := q.(DrainableQueue)
	if !ok {
		t.Fatal("queue does not implement DrainableQueue")
	}

	return dq
}

// consumeUntilClosed 持续消费并 Done，直到队列关闭，返回消费计数。
func consumeUntilClosed(q Queue, consumed *atomic.Int64) {
	for {
		v, err := q.Get()
		if err != nil {
			if errors.Is(err, ErrQueueIsClosed) {
				return
			}
			time.Sleep(time.Millisecond)
			continue
		}
		consumed.Add(1)
		q.Done(v)
	}
}

// TestQueueImpl_ShutdownWithDrain_Idempotent_ZeroLoss 幂等模式 drain 零丢失。
func TestQueueImpl_ShutdownWithDrain_Idempotent_ZeroLoss(t *testing.T) {
	q := NewQueue(NewQueueConfig().WithValueIdempotent())

	const total = 203
	const inFlight = 3

	for i := 0; i < total; i++ {
		if err := q.Put(i); err != nil {
			t.Fatalf("put: %v", err)
		}
	}

	inFlightValues := make([]any, inFlight)
	for i := 0; i < inFlight; i++ {
		v, err := q.Get()
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		inFlightValues[i] = v
	}

	var consumed atomic.Int64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumeUntilClosed(q, &consumed)
	}()

	go func() {
		time.Sleep(20 * time.Millisecond)
		for _, v := range inFlightValues {
			q.Done(v)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := drainableOf(t, q).ShutdownWithDrain(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}

	wg.Wait()

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
	if got := consumed.Load(); got != total-inFlight {
		t.Errorf("consumed = %d, want %d (zero loss)", got, total-inFlight)
	}
	if q.Len() != 0 {
		t.Errorf("len = %d, want 0", q.Len())
	}
}

// TestQueueImpl_ShutdownWithDrain_DrainTracking_ZeroLoss 非幂等 + drainTracking 零丢失。
func TestQueueImpl_ShutdownWithDrain_DrainTracking_ZeroLoss(t *testing.T) {
	q := NewQueue(NewQueueConfig().WithDrainTracking())

	const total = 203
	const inFlight = 3

	for i := 0; i < total; i++ {
		if err := q.Put(i); err != nil {
			t.Fatalf("put: %v", err)
		}
	}

	inFlightValues := make([]any, inFlight)
	for i := 0; i < inFlight; i++ {
		v, err := q.Get()
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		inFlightValues[i] = v
	}

	var consumed atomic.Int64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumeUntilClosed(q, &consumed)
	}()

	go func() {
		time.Sleep(20 * time.Millisecond)
		for _, v := range inFlightValues {
			q.Done(v)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := drainableOf(t, q).ShutdownWithDrain(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}

	wg.Wait()

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
	if got := consumed.Load(); got != total-inFlight {
		t.Errorf("consumed = %d, want %d (zero loss)", got, total-inFlight)
	}
}

// TestQueueImpl_ShutdownWithDrain_Timeout R5：无消费者时 drain 超时强制关闭。
func TestQueueImpl_ShutdownWithDrain_Timeout(t *testing.T) {
	q := NewQueue(nil)
	for i := 0; i < 5; i++ {
		if err := q.Put(i); err != nil {
			t.Fatalf("put: %v", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := drainableOf(t, q).ShutdownWithDrain(ctx)
	elapsed := time.Since(start)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	if elapsed >= time.Second {
		t.Errorf("drain should return right after deadline, took %v", elapsed)
	}
	if !q.IsClosed() {
		t.Error("queue must be force-closed after timeout")
	}
	if q.Len() != 0 {
		t.Errorf("len = %d, want 0 after force close", q.Len())
	}
	if err := q.Put(99); !errors.Is(err, ErrQueueIsClosed) {
		t.Errorf("put after force close = %v, want ErrQueueIsClosed", err)
	}
}

// TestQueueImpl_ShutdownWithDrain_Canceled 主动取消返回 context.Canceled。
func TestQueueImpl_ShutdownWithDrain_Canceled(t *testing.T) {
	q := NewQueue(nil)
	if err := q.Put(1); err != nil {
		t.Fatalf("put: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	dq := drainableOf(t, q)
	go func() {
		errCh <- dq.ShutdownWithDrain(ctx)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("drain should return after cancel")
	}

	if !q.IsClosed() {
		t.Error("queue must be force-closed after cancel")
	}
}

// TestQueueImpl_ShutdownWithDrain_DrainSemantics drain 期间 Put 被拒绝，Get/Done 正常。
func TestQueueImpl_ShutdownWithDrain_DrainSemantics(t *testing.T) {
	q := NewQueue(NewQueueConfig().WithValueIdempotent())
	if err := q.Put(1); err != nil {
		t.Fatalf("put: %v", err)
	}
	if err := q.Put(2); err != nil {
		t.Fatalf("put: %v", err)
	}

	v1, err := q.Get()
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	errCh := make(chan error, 1)
	dq := drainableOf(t, q)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		errCh <- dq.ShutdownWithDrain(ctx)
	}()

	// 等待 drain 启动后再探测。
	time.Sleep(20 * time.Millisecond)

	deadline := time.Now().Add(time.Second)
	for {
		if err := q.Put(3); errors.Is(err, ErrQueueIsClosed) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("drain did not start rejecting Put in time")
		}
		time.Sleep(time.Millisecond)
	}

	if q.IsClosed() {
		t.Fatal("IsClosed must stay false while draining")
	}

	v2, err := q.Get()
	if err != nil {
		t.Fatalf("get during drain: %v", err)
	}
	if v2 != 2 {
		t.Fatalf("get = %v, want 2", v2)
	}

	q.Done(v1)
	q.Done(v2)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("drain should complete after all in-flight done: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("drain did not complete")
	}

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
}

// TestQueueImpl_ShutdownWithDrain_NoTrackingInFlightUnknown 无 tracking 时 drain 仅等在队项清空。
func TestQueueImpl_ShutdownWithDrain_NoTrackingInFlightUnknown(t *testing.T) {
	q := NewQueue(nil)
	if err := q.Put(1); err != nil {
		t.Fatalf("put: %v", err)
	}

	if _, err := q.Get(); err != nil {
		t.Fatalf("get: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := drainableOf(t, q).ShutdownWithDrain(ctx); err != nil {
		t.Fatalf("drain should return nil once list is empty: %v", err)
	}
	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
}

// TestQueueImpl_ShutdownWithDrain_ConcurrentShutdown Shutdown 与 drain 并发交错。
func TestQueueImpl_ShutdownWithDrain_ConcurrentShutdown(t *testing.T) {
	for i := 0; i < 50; i++ {
		q := NewQueue(NewQueueConfig().WithValueIdempotent())
		for j := 0; j < 10; j++ {
			if err := q.Put(j); err != nil {
				t.Fatalf("put: %v", err)
			}
		}

		var wg sync.WaitGroup
		dq := drainableOf(t, q)
		wg.Add(2)
		go func() {
			defer wg.Done()
			q.Shutdown()
		}()
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_ = dq.ShutdownWithDrain(ctx)
		}()
		wg.Wait()

		if !q.IsClosed() {
			t.Fatal("queue must be closed")
		}
	}
}

// TestQueueImpl_ShutdownWithDrain_AfterShutdown Shutdown 后 drain 直接返回 nil。
func TestQueueImpl_ShutdownWithDrain_AfterShutdown(t *testing.T) {
	q := NewQueue(nil)
	if err := q.Put(1); err != nil {
		t.Fatalf("put: %v", err)
	}
	q.Shutdown()

	if err := drainableOf(t, q).ShutdownWithDrain(context.Background()); err != nil {
		t.Fatalf("drain after shutdown = %v, want nil", err)
	}
}

// TestQueueImpl_ShutdownWithDrain_EmptyQueue 覆盖空队列 drain 立即完成。
func TestQueueImpl_ShutdownWithDrain_EmptyQueue(t *testing.T) {
	q := NewQueue(NewQueueConfig().WithValueIdempotent())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	start := time.Now()
	if err := drainableOf(t, q).ShutdownWithDrain(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Errorf("empty queue drain took %v, want immediate", elapsed)
	}
	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
}
