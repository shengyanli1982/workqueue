package workqueue

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestDelayingQueue_ShutdownWithDrain_ZeroLoss_DiscardUndue 覆盖延迟队列 drain 编排：
// drain 等待期间到期的项被搬运并消费（零丢失），堆中未到期项按 Q4a 契约
// 在最终关闭时丢弃并计入 discardedDelayed。
// 用 1 个 in-flight（WithDrainTracking）项阻塞 drain 判定，确保 20ms 延迟项
// 在判定完成前到期并被搬运。
func TestDelayingQueue_ShutdownWithDrain_ZeroLoss_DiscardUndue(t *testing.T) {
	cfg := NewDelayingQueueConfig()
	cfg.WithDrainTracking()
	q := NewDelayingQueue(cfg)

	const immediate = 5
	const dueSoon = 5
	const undue = 3

	for i := 0; i < immediate; i++ {
		if err := q.Put(i); err != nil {
			t.Fatalf("put: %v", err)
		}
	}
	for i := 0; i < dueSoon; i++ {
		if err := q.PutWithDelay(100+i, 20); err != nil {
			t.Fatalf("put with delay: %v", err)
		}
	}
	for i := 0; i < undue; i++ {
		if err := q.PutWithDelay(200+i, 10_000); err != nil {
			t.Fatalf("put with delay: %v", err)
		}
	}

	// in-flight 项：阻塞 drain 判定 ~80ms，给 20ms 延迟项留出到期搬运窗口。
	inFlightValue, err := q.Get()
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	var consumed atomic.Int64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumeUntilClosed(q, &consumed)
	}()

	go func() {
		time.Sleep(80 * time.Millisecond)
		q.Done(inFlightValue)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := drainableOf(t, q).ShutdownWithDrain(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}

	wg.Wait()

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
	// in-flight 项由测试 goroutine 直接 Done，不经消费者计数。
	if got, want := consumed.Load(), int64(immediate+dueSoon-1); got != want {
		t.Errorf("consumed = %d, want %d", got, want)
	}
	if got := q.(*delayingQueueImpl).discardedDelayed.Load(); got != undue {
		t.Errorf("discardedDelayed = %d, want %d", got, undue)
	}
}

// TestDelayingQueue_ShutdownWithDrain_RejectNewPut 覆盖 drain 期间新入队被拒绝、
// IsClosed 保持 false、Get/Done 正常。
func TestDelayingQueue_ShutdownWithDrain_RejectNewPut(t *testing.T) {
	cfg := NewDelayingQueueConfig()
	cfg.WithDrainTracking()
	q := NewDelayingQueue(cfg)

	for i := 1; i <= 2; i++ {
		if err := q.Put(i); err != nil {
			t.Fatalf("put: %v", err)
		}
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

	// draining 在 drain 入口同步置位；等待 goroutine 调度完成后再开始探测。
	time.Sleep(20 * time.Millisecond)

	if err := q.PutWithDelay(9, 10); !errors.Is(err, ErrQueueIsClosed) {
		t.Fatalf("PutWithDelay during drain = %v, want ErrQueueIsClosed", err)
	}
	if err := q.Put(9); !errors.Is(err, ErrQueueIsClosed) {
		t.Fatalf("Put during drain = %v, want ErrQueueIsClosed", err)
	}
	if q.IsClosed() {
		t.Fatal("IsClosed must stay false while draining")
	}

	// Get/Done 正常：消费剩余在队项并结束 in-flight，drain 应完成。
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
			t.Fatalf("drain should complete: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("drain did not complete")
	}

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
}

// TestDelayingQueue_ShutdownWithDrain_Timeout 覆盖无消费者时 drain 超时：
// 返回 DeadlineExceeded，强制关闭，堆中项计入 discardedDelayed。
func TestDelayingQueue_ShutdownWithDrain_Timeout(t *testing.T) {
	q := NewDelayingQueue(NewDelayingQueueConfig())
	if err := q.PutWithDelay(1, 10_000); err != nil {
		t.Fatalf("put with delay: %v", err)
	}
	if err := q.Put(2); err != nil {
		t.Fatalf("put: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := drainableOf(t, q).ShutdownWithDrain(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	if !q.IsClosed() {
		t.Error("queue must be force-closed after timeout")
	}
	if got := q.(*delayingQueueImpl).discardedDelayed.Load(); got != 1 {
		t.Errorf("discardedDelayed = %d, want 1", got)
	}
}

// TestDelayingQueue_ShutdownWithDrain_ConcurrentShutdown 覆盖并发交错。
func TestDelayingQueue_ShutdownWithDrain_ConcurrentShutdown(t *testing.T) {
	for i := 0; i < 30; i++ {
		q := NewDelayingQueue(NewDelayingQueueConfig())
		for j := 0; j < 5; j++ {
			if err := q.Put(j); err != nil {
				t.Fatalf("put: %v", err)
			}
		}
		if err := q.PutWithDelay(100, 5); err != nil {
			t.Fatalf("put with delay: %v", err)
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

// TestTimerQueue_ShutdownWithDrain_ZeroLoss_DiscardFuture 覆盖定时队列 drain 编排：
// drain 等待期间到期的调度项被搬运并消费，未到期项在最终关闭时丢弃
// （TimerQueue 无丢弃计数器，契约见 ShutdownWithDrain 文档）。
func TestTimerQueue_ShutdownWithDrain_ZeroLoss_DiscardFuture(t *testing.T) {
	cfg := NewTimerQueueConfig()
	cfg.WithDrainTracking()
	q := NewTimerQueue(cfg)

	const immediate = 2
	const dueSoon = 2

	for i := 0; i < immediate; i++ {
		if err := q.Put(i); err != nil {
			t.Fatalf("put: %v", err)
		}
	}
	for i := 0; i < dueSoon; i++ {
		if err := q.PutAfter(100+i, 20*time.Millisecond); err != nil {
			t.Fatalf("put after: %v", err)
		}
	}
	if err := q.PutAfter(200, 10*time.Second); err != nil {
		t.Fatalf("put after: %v", err)
	}

	// in-flight 项：阻塞 drain 判定，给 20ms 调度项留出到期搬运窗口。
	inFlightValue, err := q.Get()
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	var consumed atomic.Int64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumeUntilClosed(q, &consumed)
	}()

	go func() {
		time.Sleep(80 * time.Millisecond)
		q.Done(inFlightValue)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := drainableOf(t, q).ShutdownWithDrain(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}

	wg.Wait()

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
	// in-flight 项由测试 goroutine 直接 Done，不经消费者计数。
	if got, want := consumed.Load(), int64(immediate+dueSoon-1); got != want {
		t.Errorf("consumed = %d, want %d", got, want)
	}
}

// TestTimerQueue_ShutdownWithDrain_RejectNewPut 覆盖 drain 期间 PutAt/PutAfter/Put
// 被拒绝、IsClosed 保持 false、Get/Done 正常。
func TestTimerQueue_ShutdownWithDrain_RejectNewPut(t *testing.T) {
	cfg := NewTimerQueueConfig()
	cfg.WithDrainTracking()
	q := NewTimerQueue(cfg)

	for i := 1; i <= 2; i++ {
		if err := q.Put(i); err != nil {
			t.Fatalf("put: %v", err)
		}
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

	time.Sleep(20 * time.Millisecond)

	if err := q.PutAfter(9, time.Second); !errors.Is(err, ErrQueueIsClosed) {
		t.Fatalf("PutAfter during drain = %v, want ErrQueueIsClosed", err)
	}
	if err := q.PutAt(9, time.Now().Add(time.Second)); !errors.Is(err, ErrQueueIsClosed) {
		t.Fatalf("PutAt during drain = %v, want ErrQueueIsClosed", err)
	}
	if err := q.Put(9); !errors.Is(err, ErrQueueIsClosed) {
		t.Fatalf("Put during drain = %v, want ErrQueueIsClosed", err)
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
			t.Fatalf("drain should complete: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("drain did not complete")
	}

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
}

// TestTimerQueue_ShutdownWithDrain_Timeout 覆盖无消费者时在队项未被消费：
// drain 悬挂到 ctx 超时强制关闭（堆中未到期项一并丢弃）。
func TestTimerQueue_ShutdownWithDrain_Timeout(t *testing.T) {
	q := NewTimerQueue(NewTimerQueueConfig())
	if err := q.PutAfter(1, 10*time.Second); err != nil {
		t.Fatalf("put after: %v", err)
	}
	if err := q.Put(2); err != nil {
		t.Fatalf("put: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := drainableOf(t, q).ShutdownWithDrain(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	if !q.IsClosed() {
		t.Error("queue must be force-closed after timeout")
	}
	if q.Len() != 0 {
		t.Errorf("len = %d, want 0 after force close", q.Len())
	}
}

// TestTimerQueue_ShutdownWithDrain_ConcurrentShutdown 覆盖并发交错。
func TestTimerQueue_ShutdownWithDrain_ConcurrentShutdown(t *testing.T) {
	for i := 0; i < 30; i++ {
		q := NewTimerQueue(NewTimerQueueConfig())
		for j := 0; j < 5; j++ {
			if err := q.Put(j); err != nil {
				t.Fatalf("put: %v", err)
			}
		}
		if err := q.PutAfter(100, 5*time.Millisecond); err != nil {
			t.Fatalf("put after: %v", err)
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

// leaseConsumer 持续以租约方式消费并 Ack，直到队列关闭，返回消费计数。
func leaseConsumer(q LeasedQueue, consumed *atomic.Int64) {
	for {
		v, leaseID, err := q.GetWithLease(time.Second)
		if err != nil {
			if errors.Is(err, ErrQueueIsClosed) {
				return
			}
			time.Sleep(time.Millisecond)
			continue
		}
		consumed.Add(1)
		_ = q.Ack(leaseID)
		_ = v
	}
}

// TestLeasedQueue_ShutdownWithDrain_LeasesCleared 覆盖租约队列 drain 编排：
// drain 条件 = 内层 drained ∧ leases 清空；reaper 照常回收过期租约并重入队，
// 在租项归还消费后 drain 完成，零丢失。
func TestLeasedQueue_ShutdownWithDrain_LeasesCleared(t *testing.T) {
	cfg := NewLeasedQueueConfig().
		WithLeaseDuration(30 * time.Millisecond).
		WithScanInterval(10 * time.Millisecond)
	q := NewLeasedQueue(cfg)

	const total = 5
	for i := 0; i < total; i++ {
		if err := q.Put(i); err != nil {
			t.Fatalf("put: %v", err)
		}
	}

	// 持有两个租约不 Ack：等待 reaper 过期回收 → 重入队 → 被消费。
	for i := 0; i < 2; i++ {
		if _, _, err := q.GetWithLease(30 * time.Millisecond); err != nil {
			t.Fatalf("get with lease: %v", err)
		}
	}

	var consumed atomic.Int64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		leaseConsumer(q, &consumed)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := drainableOf(t, q).ShutdownWithDrain(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}

	wg.Wait()

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
	if got := consumed.Load(); got != total {
		t.Errorf("consumed = %d, want %d (zero loss)", got, total)
	}
}

// TestLeasedQueue_ShutdownWithDrain_RejectNewPut 覆盖 drain 期间新 Put 被拒绝、
// IsClosed 保持 false、GetWithLease/Ack 正常。
func TestLeasedQueue_ShutdownWithDrain_RejectNewPut(t *testing.T) {
	q := NewLeasedQueue(NewLeasedQueueConfig())
	if err := q.Put(1); err != nil {
		t.Fatalf("put: %v", err)
	}
	if err := q.Put(2); err != nil {
		t.Fatalf("put: %v", err)
	}

	v1, lease1, err := q.GetWithLease(time.Second)
	if err != nil {
		t.Fatalf("get with lease: %v", err)
	}
	_ = v1

	errCh := make(chan error, 1)
	dq := drainableOf(t, q)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		errCh <- dq.ShutdownWithDrain(ctx)
	}()

	time.Sleep(20 * time.Millisecond)

	if err := q.Put(9); !errors.Is(err, ErrQueueIsClosed) {
		t.Fatalf("Put during drain = %v, want ErrQueueIsClosed", err)
	}
	if q.IsClosed() {
		t.Fatal("IsClosed must stay false while draining")
	}

	// GetWithLease/Ack 正常：消费剩余在队项并 Ack 全部租约，drain 应完成。
	v2, lease2, err := q.GetWithLease(time.Second)
	if err != nil {
		t.Fatalf("get with lease during drain: %v", err)
	}
	if v2 != 2 {
		t.Fatalf("get = %v, want 2", v2)
	}

	if err := q.Ack(lease1); err != nil {
		t.Fatalf("ack during drain: %v", err)
	}
	if err := q.Ack(lease2); err != nil {
		t.Fatalf("ack during drain: %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("drain should complete: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("drain did not complete")
	}

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
}

// TestLeasedQueue_ShutdownWithDrain_TimeoutDiscardsResidualLeases 覆盖超时契约：
// 残余租约随强制关闭丢弃（与现 Shutdown leases=nil 语义一致）。
func TestLeasedQueue_ShutdownWithDrain_TimeoutDiscardsResidualLeases(t *testing.T) {
	cfg := NewLeasedQueueConfig().
		WithLeaseDuration(time.Hour).
		WithScanInterval(10 * time.Millisecond)
	q := NewLeasedQueue(cfg)

	if err := q.Put(1); err != nil {
		t.Fatalf("put: %v", err)
	}
	if _, _, err := q.GetWithLease(0); err != nil {
		t.Fatalf("get with lease: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := drainableOf(t, q).ShutdownWithDrain(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	if !q.IsClosed() {
		t.Error("queue must be force-closed after timeout")
	}
}

// TestLeasedQueue_ShutdownWithDrain_ConcurrentShutdown 覆盖并发交错。
func TestLeasedQueue_ShutdownWithDrain_ConcurrentShutdown(t *testing.T) {
	for i := 0; i < 30; i++ {
		q := NewLeasedQueue(NewLeasedQueueConfig())
		for j := 0; j < 5; j++ {
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

// TestBoundedBlockingQueue_ShutdownWithDrain_ZeroLoss 覆盖有界阻塞队列 drain：
// 在队项与 in-flight 项全部完成后关闭，零丢失。
func TestBoundedBlockingQueue_ShutdownWithDrain_ZeroLoss(t *testing.T) {
	cfg := NewBoundedBlockingQueueConfig().WithCapacity(16)
	cfg.WithDrainTracking()
	q := NewBoundedBlockingQueue(cfg)

	const total = 5
	for i := 0; i < total; i++ {
		if err := q.Put(i); err != nil {
			t.Fatalf("put: %v", err)
		}
	}

	inFlightValue, err := q.Get()
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	var consumed atomic.Int64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumeUntilClosed(q, &consumed)
	}()

	go func() {
		time.Sleep(50 * time.Millisecond)
		q.Done(inFlightValue)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := drainableOf(t, q).ShutdownWithDrain(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}

	wg.Wait()

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
	// in-flight 项由测试 goroutine 直接 Done，不经消费者计数。
	if got, want := consumed.Load(), int64(total-1); got != want {
		t.Errorf("consumed = %d, want %d", got, want)
	}
}

// TestBoundedBlockingQueue_ShutdownWithDrain_RejectNewPut 覆盖 drain 期间
// Put/PutWithContext 被拒绝、IsClosed 保持 false、Get/GetWithContext 正常。
func TestBoundedBlockingQueue_ShutdownWithDrain_RejectNewPut(t *testing.T) {
	cfg := NewBoundedBlockingQueueConfig().WithCapacity(16)
	cfg.WithDrainTracking()
	q := NewBoundedBlockingQueue(cfg)

	for i := 1; i <= 2; i++ {
		if err := q.Put(i); err != nil {
			t.Fatalf("put: %v", err)
		}
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

	time.Sleep(20 * time.Millisecond)

	if err := q.Put(9); !errors.Is(err, ErrQueueIsClosed) {
		t.Fatalf("Put during drain = %v, want ErrQueueIsClosed", err)
	}
	if err := q.PutWithContext(context.Background(), 9); !errors.Is(err, ErrQueueIsClosed) {
		t.Fatalf("PutWithContext during drain = %v, want ErrQueueIsClosed", err)
	}
	if q.IsClosed() {
		t.Fatal("IsClosed must stay false while draining")
	}

	v2, err := q.GetWithContext(context.Background())
	if err != nil {
		t.Fatalf("GetWithContext during drain: %v", err)
	}
	if v2 != 2 {
		t.Fatalf("get = %v, want 2", v2)
	}

	q.Done(v1)
	q.Done(v2)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("drain should complete: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("drain did not complete")
	}

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
}

// TestBoundedBlockingQueue_ShutdownWithDrain_WakesBlockedWaiters 覆盖超时强制
// 关闭唤醒阻塞在槽位上的生产者（现有 closed channel 关闭路径）。
func TestBoundedBlockingQueue_ShutdownWithDrain_WakesBlockedWaiters(t *testing.T) {
	q := NewBoundedBlockingQueue(NewBoundedBlockingQueueConfig().WithCapacity(2))
	if err := q.Put(1); err != nil {
		t.Fatalf("put: %v", err)
	}
	if err := q.Put(2); err != nil {
		t.Fatalf("put: %v", err)
	}

	prodErr := make(chan error, 1)
	go func() {
		prodErr <- q.PutWithContext(context.Background(), 3)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := drainableOf(t, q).ShutdownWithDrain(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}

	select {
	case err := <-prodErr:
		if !errors.Is(err, ErrQueueIsClosed) {
			t.Fatalf("blocked producer err = %v, want ErrQueueIsClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("blocked producer not woken after force close")
	}

	if !q.IsClosed() {
		t.Error("queue must be force-closed after timeout")
	}
}

// TestBoundedBlockingQueue_ShutdownWithDrain_Timeout 覆盖无消费者时 drain 超时。
func TestBoundedBlockingQueue_ShutdownWithDrain_Timeout(t *testing.T) {
	q := NewBoundedBlockingQueue(NewBoundedBlockingQueueConfig().WithCapacity(8))
	for i := 0; i < 3; i++ {
		if err := q.Put(i); err != nil {
			t.Fatalf("put: %v", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := drainableOf(t, q).ShutdownWithDrain(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	if !q.IsClosed() {
		t.Error("queue must be force-closed after timeout")
	}
	if q.Len() != 0 {
		t.Errorf("len = %d, want 0 after force close", q.Len())
	}
}

// TestBoundedBlockingQueue_ShutdownWithDrain_ConcurrentShutdown 覆盖并发交错。
func TestBoundedBlockingQueue_ShutdownWithDrain_ConcurrentShutdown(t *testing.T) {
	for i := 0; i < 30; i++ {
		q := NewBoundedBlockingQueue(NewBoundedBlockingQueueConfig().WithCapacity(8))
		for j := 0; j < 5; j++ {
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

// TestPriorityQueue_ShutdownWithDrain_ZeroLoss 覆盖优先队列 drain：
// 堆中项与内层 list 统一看待，全部可被 Get 消费后关闭，零丢失。
func TestPriorityQueue_ShutdownWithDrain_ZeroLoss(t *testing.T) {
	cfg := NewPriorityQueueConfig()
	cfg.WithDrainTracking()
	q := NewPriorityQueue(cfg)

	const total = 6
	priorities := []int64{PRIORITY_HIGH, PRIORITY_NORMAL, PRIORITY_LOW, PRIORITY_FASTEST, PRIORITY_NORMAL, PRIORITY_SLOWEST}
	for i := 0; i < total; i++ {
		if err := q.PutWithPriority(i, priorities[i]); err != nil {
			t.Fatalf("put with priority: %v", err)
		}
	}

	inFlightValue, err := q.Get()
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	var consumed atomic.Int64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumeUntilClosed(q, &consumed)
	}()

	go func() {
		time.Sleep(50 * time.Millisecond)
		q.Done(inFlightValue)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := drainableOf(t, q).ShutdownWithDrain(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}

	wg.Wait()

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
	// in-flight 项由测试 goroutine 直接 Done，不经消费者计数。
	if got, want := consumed.Load(), int64(total-1); got != want {
		t.Errorf("consumed = %d, want %d", got, want)
	}
}

// TestPriorityQueue_ShutdownWithDrain_RejectNewPut 覆盖 drain 期间
// PutWithPriority/Put 被拒绝、Get/Done 正常。
func TestPriorityQueue_ShutdownWithDrain_RejectNewPut(t *testing.T) {
	cfg := NewPriorityQueueConfig()
	cfg.WithDrainTracking()
	q := NewPriorityQueue(cfg)

	if err := q.PutWithPriority(1, PRIORITY_HIGH); err != nil {
		t.Fatalf("put: %v", err)
	}
	if err := q.PutWithPriority(2, PRIORITY_LOW); err != nil {
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

	time.Sleep(20 * time.Millisecond)

	if err := q.PutWithPriority(9, PRIORITY_HIGH); !errors.Is(err, ErrQueueIsClosed) {
		t.Fatalf("PutWithPriority during drain = %v, want ErrQueueIsClosed", err)
	}
	if err := q.Put(9); !errors.Is(err, ErrQueueIsClosed) {
		t.Fatalf("Put during drain = %v, want ErrQueueIsClosed", err)
	}
	if q.IsClosed() {
		t.Fatal("IsClosed must stay false while draining")
	}

	v2, err := q.Get()
	if err != nil {
		t.Fatalf("get during drain: %v", err)
	}
	if v2 != 2 {
		t.Fatalf("get = %v, want 2 (1 is in-flight)", v2)
	}

	q.Done(v1)
	q.Done(v2)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("drain should complete: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("drain did not complete")
	}

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
}

// TestPriorityQueue_ShutdownWithDrain_Timeout 覆盖无消费者时 drain 超时。
func TestPriorityQueue_ShutdownWithDrain_Timeout(t *testing.T) {
	q := NewPriorityQueue(NewPriorityQueueConfig())
	if err := q.PutWithPriority(1, PRIORITY_NORMAL); err != nil {
		t.Fatalf("put with priority: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := drainableOf(t, q).ShutdownWithDrain(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	if !q.IsClosed() {
		t.Error("queue must be force-closed after timeout")
	}
}

// TestDeadLetterQueue_ShutdownWithDrain_ZeroLoss 覆盖死信队列 drain：
// 主队列（死信存储区）drained 即视为完成。
func TestDeadLetterQueue_ShutdownWithDrain_ZeroLoss(t *testing.T) {
	q := NewDeadLetterQueue(NewDeadLetterQueueConfig())

	const total = 3
	for i := 0; i < total; i++ {
		if err := q.PutDead(&DeadLetter{Payload: i, SourceQueue: "test"}); err != nil {
			t.Fatalf("put dead: %v", err)
		}
	}

	var consumed atomic.Int64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			letter, err := q.GetDead()
			if err != nil {
				if errors.Is(err, ErrQueueIsClosed) {
					return
				}
				time.Sleep(time.Millisecond)
				continue
			}
			consumed.Add(1)
			_ = q.AckDead(letter)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := drainableOf(t, q).ShutdownWithDrain(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}

	wg.Wait()

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
	if got := consumed.Load(); got != total {
		t.Errorf("consumed = %d, want %d", got, total)
	}
}

// TestDeadLetterQueue_ShutdownWithDrain_RejectNewPut 覆盖 drain 期间
// PutDead/Put 被拒绝、GetDead/AckDead 正常；drain 不等待死信被 Ack。
func TestDeadLetterQueue_ShutdownWithDrain_RejectNewPut(t *testing.T) {
	q := NewDeadLetterQueue(NewDeadLetterQueueConfig())

	for i := 1; i <= 2; i++ {
		if err := q.PutDead(&DeadLetter{Payload: i, SourceQueue: "test"}); err != nil {
			t.Fatalf("put dead: %v", err)
		}
	}

	errCh := make(chan error, 1)
	dq := drainableOf(t, q)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		errCh <- dq.ShutdownWithDrain(ctx)
	}()

	time.Sleep(20 * time.Millisecond)

	if err := q.PutDead(&DeadLetter{Payload: 9, SourceQueue: "test"}); !errors.Is(err, ErrQueueIsClosed) {
		t.Fatalf("PutDead during drain = %v, want ErrQueueIsClosed", err)
	}
	if err := q.Put(&DeadLetter{Payload: 9, SourceQueue: "test"}); !errors.Is(err, ErrQueueIsClosed) {
		t.Fatalf("Put during drain = %v, want ErrQueueIsClosed", err)
	}
	if q.IsClosed() {
		t.Fatal("IsClosed must stay false while draining")
	}

	// GetDead/AckDead 正常：消费全部死信使存储区清空，drain 随之完成。
	for i := 0; i < 2; i++ {
		letter, err := q.GetDead()
		if err != nil {
			t.Fatalf("get dead during drain: %v", err)
		}
		if err := q.AckDead(letter); err != nil {
			t.Fatalf("ack dead during drain: %v", err)
		}
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("drain should complete once storage empties: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("drain did not complete")
	}

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
}

// TestDeadLetterQueue_ShutdownWithDrain_Timeout 覆盖无消费者时 drain 超时。
func TestDeadLetterQueue_ShutdownWithDrain_Timeout(t *testing.T) {
	q := NewDeadLetterQueue(NewDeadLetterQueueConfig())
	if err := q.PutDead(&DeadLetter{Payload: 1, SourceQueue: "test"}); err != nil {
		t.Fatalf("put dead: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := drainableOf(t, q).ShutdownWithDrain(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	if !q.IsClosed() {
		t.Error("queue must be force-closed after timeout")
	}
}

// TestRetryQueue_ShutdownWithDrain_Delegate 覆盖重试队列 drain 委托：
// 内层 DelayingQueue drained 后关闭并重置 attempts；drain 期间 Retry 被拒绝。
func TestRetryQueue_ShutdownWithDrain_Delegate(t *testing.T) {
	q := NewRetryQueue(NewRetryQueueConfig())

	const total = 3
	for i := 0; i < total; i++ {
		if err := q.Put(i); err != nil {
			t.Fatalf("put: %v", err)
		}
	}

	var consumed atomic.Int64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumeUntilClosed(q, &consumed)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := drainableOf(t, q).ShutdownWithDrain(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}

	wg.Wait()

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
	if got := consumed.Load(); got != total {
		t.Errorf("consumed = %d, want %d", got, total)
	}
	if n := q.NumRequeues(0); n != 0 {
		t.Errorf("attempts must be reset after drain, got %d", n)
	}
}

// TestRetryQueue_ShutdownWithDrain_RejectRetry 覆盖 drain 期间 Retry
// （含 in-flight 重入队）被拒绝：与 Shutdown 语义一致，由调用方兜底处理。
func TestRetryQueue_ShutdownWithDrain_RejectRetry(t *testing.T) {
	cfg := NewRetryQueueConfig()
	cfg.WithDrainTracking()
	q := NewRetryQueue(cfg)
	if err := q.Put(1); err != nil {
		t.Fatalf("put: %v", err)
	}

	v, err := q.Get()
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

	time.Sleep(20 * time.Millisecond)

	if err := q.Retry(v, errors.New("fail")); !errors.Is(err, ErrQueueIsClosed) {
		t.Fatalf("Retry during drain = %v, want ErrQueueIsClosed", err)
	}
	if q.IsClosed() {
		t.Fatal("IsClosed must stay false while draining")
	}

	// Retry 被拒后由调用方 Done 释放处理中追踪，drain 随之完成。
	q.Done(v)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("drain should complete: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("drain did not complete")
	}

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
}

// TestRateLimitingQueue_ShutdownWithDrain_Delegate 覆盖限流队列 drain 委托。
func TestRateLimitingQueue_ShutdownWithDrain_Delegate(t *testing.T) {
	q := NewRateLimitingQueue(NewRateLimitingQueueConfig())

	const total = 3
	for i := 0; i < total; i++ {
		if err := q.PutWithLimited(i); err != nil {
			t.Fatalf("put with limited: %v", err)
		}
	}

	var consumed atomic.Int64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumeUntilClosed(q, &consumed)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := drainableOf(t, q).ShutdownWithDrain(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}

	wg.Wait()

	if !q.IsClosed() {
		t.Error("queue must be closed after drain")
	}
	if got := consumed.Load(); got != total {
		t.Errorf("consumed = %d, want %d", got, total)
	}
}
