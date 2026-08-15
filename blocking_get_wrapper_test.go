package workqueue

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// leasedBlockingOf 断言租约队列提供阻塞租约消费变体（opt-in 方法，
// 外部使用者经类型断言获取，与 BlockingGetQueue 的 opt-in 风格一致）。
func leasedBlockingOf(t *testing.T, q LeasedQueue) interface {
	GetWithLeaseWithContext(ctx context.Context, timeout time.Duration) (value any, leaseID string, err error)
} {
	t.Helper()

	lq, ok := q.(interface {
		GetWithLeaseWithContext(ctx context.Context, timeout time.Duration) (value any, leaseID string, err error)
	})
	if !ok {
		t.Fatal("leased queue does not implement GetWithLeaseWithContext")
	}

	return lq
}

type fixedDelayPolicy struct{ delay time.Duration }

func (p fixedDelayPolicy) NextDelay(value any, attempt int, reason error) (time.Duration, bool) {
	return p.delay, attempt <= 3
}

type fixedLimiter struct{ delay time.Duration }

func (l fixedLimiter) When(value any) time.Duration { return l.delay }

// TestDelayingQueue_GetWithContext_WakesOnDue 验证到期项由 scheduler 搬入内层后
// 唤醒阻塞消费者。
func TestDelayingQueue_GetWithContext_WakesOnDue(t *testing.T) {
	q := NewDelayingQueue(NewDelayingQueueConfig())
	defer q.Shutdown()

	if err := q.PutWithDelay("late", 80); err != nil {
		t.Fatalf("put with delay: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	v, err := awaitBlocking(t, blockingGetOf(t, q), ctx)
	if err != nil {
		t.Fatalf("GetWithContext error: %v", err)
	}
	if v != "late" {
		t.Fatalf("value = %v, want late", v)
	}
	if elapsed := time.Since(start); elapsed < 70*time.Millisecond {
		t.Errorf("elapsed %v shorter than delay 80ms: item delivered too early", elapsed)
	}
}

// TestTimerQueue_GetWithContext_WakesOnDue 验证定时项到期投递后唤醒阻塞消费者；
// 取消调度后等待超时。
func TestTimerQueue_GetWithContext_WakesOnDue(t *testing.T) {
	t.Run("WakeOnDue", func(t *testing.T) {
		q := NewTimerQueue(NewTimerQueueConfig())
		defer q.Shutdown()

		if err := q.PutAfter("later", 80*time.Millisecond); err != nil {
			t.Fatalf("put after: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		start := time.Now()
		v, err := awaitBlocking(t, blockingGetOf(t, q), ctx)
		if err != nil {
			t.Fatalf("GetWithContext error: %v", err)
		}
		if v != "later" {
			t.Fatalf("value = %v, want later", v)
		}
		if elapsed := time.Since(start); elapsed < 70*time.Millisecond {
			t.Errorf("elapsed %v shorter than 80ms: delivered too early", elapsed)
		}
	})

	t.Run("CanceledItemNeverWakes", func(t *testing.T) {
		q := NewTimerQueue(NewTimerQueueConfig())
		defer q.Shutdown()

		if err := q.PutAfter("doomed", 200*time.Millisecond); err != nil {
			t.Fatalf("put after: %v", err)
		}
		if !q.Cancel("doomed") {
			t.Fatal("cancel failed")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := awaitBlocking(t, blockingGetOf(t, q), ctx)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("err = %v, want DeadlineExceeded after cancel", err)
		}
	})
}

// TestRetryQueue_GetWithContext_WakesAfterRetryDelay 验证重试延迟到期搬运后
// 唤醒阻塞消费者（RetryQueue 委托 DelayingQueue 的 GetWithContext）。
func TestRetryQueue_GetWithContext_WakesAfterRetryDelay(t *testing.T) {
	cfg := NewRetryQueueConfig().WithPolicy(fixedDelayPolicy{delay: 60 * time.Millisecond})
	q := NewRetryQueue(cfg)
	defer q.Shutdown()

	bq := blockingGetOf(t, q)

	if err := q.Put("task"); err != nil {
		t.Fatalf("put: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	v, err := bq.GetWithContext(ctx)
	cancel()
	if err != nil || v != "task" {
		t.Fatalf("first get: v=%v err=%v", v, err)
	}

	if err := q.Retry("task", errors.New("boom")); err != nil {
		t.Fatalf("retry: %v", err)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	start := time.Now()
	v, err = bq.GetWithContext(ctx2)
	if err != nil || v != "task" {
		t.Fatalf("retry get: v=%v err=%v", v, err)
	}
	if elapsed := time.Since(start); elapsed < 50*time.Millisecond {
		t.Errorf("elapsed %v shorter than retry delay 60ms", elapsed)
	}
	q.Done("task")
}

// TestRateLimitingQueue_GetWithContext_WakesAfterLimitDelay 验证限流延迟到期后
// 唤醒阻塞消费者（委托 DelayingQueue 的 GetWithContext）。
func TestRateLimitingQueue_GetWithContext_WakesAfterLimitDelay(t *testing.T) {
	cfg := NewRateLimitingQueueConfig().WithLimiter(fixedLimiter{delay: 60 * time.Millisecond})
	q := NewRateLimitingQueue(cfg)
	defer q.Shutdown()

	if err := q.PutWithLimited("limited"); err != nil {
		t.Fatalf("put with limited: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	v, err := awaitBlocking(t, blockingGetOf(t, q), ctx)
	if err != nil || v != "limited" {
		t.Fatalf("get: v=%v err=%v", v, err)
	}
	if elapsed := time.Since(start); elapsed < 50*time.Millisecond {
		t.Errorf("elapsed %v shorter than limit delay 60ms", elapsed)
	}
}

// TestPriorityQueue_GetWithContext 验证优先级队列：Put/PutWithPriority 唤醒
// 阻塞消费者（堆即内层存储，复用内层广播），且按优先级顺序出队。
func TestPriorityQueue_GetWithContext(t *testing.T) {
	t.Run("WakeOnPut", func(t *testing.T) {
		q := NewPriorityQueue(NewPriorityQueueConfig())
		defer q.Shutdown()
		bq := blockingGetOf(t, q)

		started := make(chan struct{})
		ch := make(chan any, 1)
		go func() {
			close(started)
			v, err := bq.GetWithContext(context.Background())
			if err == nil {
				ch <- v
			}
		}()

		<-started
		time.Sleep(10 * time.Millisecond)
		if err := q.PutWithPriority("urgent", PRIORITY_HIGH); err != nil {
			t.Fatalf("put with priority: %v", err)
		}

		select {
		case v := <-ch:
			if v != "urgent" {
				t.Fatalf("value = %v, want urgent", v)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("waiter not woken by PutWithPriority: lost wakeup")
		}
	})

	t.Run("PriorityOrderPreserved", func(t *testing.T) {
		q := NewPriorityQueue(NewPriorityQueueConfig())
		defer q.Shutdown()
		bq := blockingGetOf(t, q)

		if err := q.Put("normal"); err != nil {
			t.Fatalf("put: %v", err)
		}
		if err := q.PutWithPriority("high", PRIORITY_HIGH); err != nil {
			t.Fatalf("put with priority: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		v, err := bq.GetWithContext(ctx)
		if err != nil || v != "high" {
			t.Fatalf("first get: v=%v err=%v, want high", v, err)
		}
		v, err = bq.GetWithContext(ctx)
		if err != nil || v != "normal" {
			t.Fatalf("second get: v=%v err=%v, want normal", v, err)
		}
	})
}

// TestLeasedQueue_GetWithLeaseWithContext 验证阻塞租约消费变体：
// 阻塞唤醒、租约登记与 Ack、timeout<=0 回退 config.leaseDuration、ctx 取消。
func TestLeasedQueue_GetWithLeaseWithContext(t *testing.T) {
	t.Run("BlockThenWakeAndLease", func(t *testing.T) {
		q := NewLeasedQueue(NewLeasedQueueConfig().WithLeaseDuration(500 * time.Millisecond))
		defer q.Shutdown()
		lq := leasedBlockingOf(t, q)

		started := make(chan struct{})
		type result struct {
			value   any
			leaseID string
			err     error
		}
		ch := make(chan result, 1)
		go func() {
			close(started)
			v, id, err := lq.GetWithLeaseWithContext(context.Background(), 0)
			ch <- result{v, id, err}
		}()

		<-started
		time.Sleep(10 * time.Millisecond)
		if err := q.Put("leased"); err != nil {
			t.Fatalf("put: %v", err)
		}

		select {
		case r := <-ch:
			if r.err != nil {
				t.Fatalf("GetWithLeaseWithContext error: %v", r.err)
			}
			if r.value != "leased" || r.leaseID == "" {
				t.Fatalf("value=%v leaseID=%q, want leased + non-empty lease", r.value, r.leaseID)
			}
			if err := q.Ack(r.leaseID); err != nil {
				t.Fatalf("ack: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("GetWithLeaseWithContext not woken after Put: lost wakeup")
		}
	})

	t.Run("TimeoutFallbackToConfig", func(t *testing.T) {
		q := NewLeasedQueue(NewLeasedQueueConfig().WithLeaseDuration(500 * time.Millisecond))
		defer q.Shutdown()
		lq := leasedBlockingOf(t, q)

		if err := q.Put("cfg"); err != nil {
			t.Fatalf("put: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		v, leaseID, err := lq.GetWithLeaseWithContext(ctx, 0)
		if err != nil || v != "cfg" || leaseID == "" {
			t.Fatalf("get: v=%v leaseID=%q err=%v", v, leaseID, err)
		}
		if err := q.Ack(leaseID); err != nil {
			t.Fatalf("ack: %v", err)
		}
	})

	t.Run("ContextCanceledNoLeaseLeak", func(t *testing.T) {
		q := NewLeasedQueue(NewLeasedQueueConfig().WithLeaseDuration(500 * time.Millisecond))
		defer q.Shutdown()
		lq := leasedBlockingOf(t, q)

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		_, leaseID, err := lq.GetWithLeaseWithContext(ctx, 0)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("err = %v, want DeadlineExceeded", err)
		}
		if leaseID != "" {
			t.Fatalf("leaseID = %q, want empty on cancel", leaseID)
		}

		// 取消不影响后续消费：元素仍在队。
		if err := q.Put("still-here"); err != nil {
			t.Fatalf("put: %v", err)
		}
		v, leaseID, err := lq.GetWithLeaseWithContext(context.Background(), 0)
		if err != nil || v != "still-here" || leaseID == "" {
			t.Fatalf("after cancel: v=%v leaseID=%q err=%v", v, leaseID, err)
		}
		if err := q.Ack(leaseID); err != nil {
			t.Fatalf("ack: %v", err)
		}
	})

	t.Run("ShutdownWakesWaiter", func(t *testing.T) {
		q := NewLeasedQueue(NewLeasedQueueConfig())
		lq := leasedBlockingOf(t, q)

		started := make(chan struct{})
		ch := make(chan error, 1)
		go func() {
			close(started)
			_, _, err := lq.GetWithLeaseWithContext(context.Background(), 0)
			ch <- err
		}()

		<-started
		time.Sleep(10 * time.Millisecond)
		q.Shutdown()

		select {
		case err := <-ch:
			if !errors.Is(err, ErrQueueIsClosed) {
				t.Fatalf("err = %v, want ErrQueueIsClosed", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("GetWithLeaseWithContext not woken after Shutdown")
		}
	})
}

// TestBoundedBlockingQueue_GetWithContext_NeverReturnsEmpty 验证 BBQ 阻塞消费
// 永不返回 ErrQueueIsEmpty：空队列等待 ctx 超时，关闭唤醒返回 ErrQueueIsClosed。
func TestBoundedBlockingQueue_GetWithContext_NeverReturnsEmpty(t *testing.T) {
	q := NewBoundedBlockingQueue(NewBoundedBlockingQueueConfig().WithCapacity(16))
	bq := blockingGetOf(t, q)

	for i := 0; i < 10; i++ {
		if err := q.Put(i); err != nil {
			t.Fatalf("put %d: %v", i, err)
		}
	}

	var wg sync.WaitGroup
	var sawEmpty atomic.Bool
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
				_, err := bq.GetWithContext(ctx)
				cancel()
				if err == nil {
					continue
				}
				if errors.Is(err, ErrQueueIsEmpty) {
					sawEmpty.Store(true)
				}
				return
			}
		}()
	}

	wg.Wait()
	if sawEmpty.Load() {
		t.Fatal("GetWithContext returned ErrQueueIsEmpty")
	}

	// 空队列上阻塞的消费者：ctx 超时与关闭唤醒均不得返回 ErrQueueIsEmpty。
	started := make(chan struct{})
	ch := make(chan error, 1)
	go func() {
		close(started)
		_, err := bq.GetWithContext(context.Background())
		ch <- err
	}()

	<-started
	time.Sleep(10 * time.Millisecond)
	q.Shutdown()

	select {
	case err := <-ch:
		if errors.Is(err, ErrQueueIsEmpty) {
			t.Fatal("GetWithContext returned ErrQueueIsEmpty on shutdown wake")
		}
		if !errors.Is(err, ErrQueueIsClosed) && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("err = %v, want ErrQueueIsClosed", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("waiter not released on Shutdown")
	}
}

// TestDeadLetterQueue_GetWithContext 验证死信队列阻塞消费委托内层并返回 *DeadLetter。
func TestDeadLetterQueue_GetWithContext(t *testing.T) {
	q := NewDeadLetterQueue(NewDeadLetterQueueConfig())
	defer q.Shutdown()
	bq := blockingGetOf(t, q)

	started := make(chan struct{})
	type result struct {
		value any
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		close(started)
		v, err := bq.GetWithContext(context.Background())
		ch <- result{v, err}
	}()

	<-started
	time.Sleep(10 * time.Millisecond)

	letter := &DeadLetter{Payload: "dead"}
	if err := q.PutDead(letter); err != nil {
		t.Fatalf("put dead: %v", err)
	}

	select {
	case r := <-ch:
		if r.err != nil {
			t.Fatalf("GetWithContext error: %v", r.err)
		}
		got, ok := r.value.(*DeadLetter)
		if !ok {
			t.Fatalf("value type = %T, want *DeadLetter", r.value)
		}
		if got.Payload != "dead" {
			t.Fatalf("payload = %v, want dead", got.Payload)
		}
		if err := q.AckDead(got); err != nil {
			t.Fatalf("ack dead: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("waiter not woken by PutDead: lost wakeup")
	}
}
