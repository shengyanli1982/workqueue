package workqueue

import (
	"context"
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// blockingGetOf 断言队列实现了可选接口 BlockingGetQueue。
func blockingGetOf(t *testing.T, q any) BlockingGetQueue {
	t.Helper()

	bq, ok := q.(BlockingGetQueue)
	if !ok {
		t.Fatal("queue does not implement BlockingGetQueue")
	}

	return bq
}

// awaitBlocking 在独立 goroutine 中执行 GetWithContext 并把结果送回，
// 用 2s 兜底超时防止测试因唤醒丢失而悬挂。
func awaitBlocking(t *testing.T, q BlockingGetQueue, ctx context.Context) (any, error) {
	t.Helper()

	type result struct {
		value any
		err   error
	}

	ch := make(chan result, 1)
	go func() {
		v, err := q.GetWithContext(ctx)
		ch <- result{v, err}
	}()

	select {
	case r := <-ch:
		return r.value, r.err
	case <-time.After(2 * time.Second):
		t.Fatal("GetWithContext blocked longer than 2s: suspected lost wakeup")
		return nil, nil
	}
}

// TestQueue_GetWithContext_BlocksUntilPut 验证空队列上 GetWithContext 阻塞，
// 直到 Put 入队后立即唤醒（宽松阈值 5ms，设计目标 p99<200µs）。
func TestQueue_GetWithContext_BlocksUntilPut(t *testing.T) {
	q := NewQueue(NewQueueConfig())
	defer q.Shutdown()

	bq := blockingGetOf(t, q)

	started := make(chan struct{})
	type result struct {
		value any
		err   error
		at    time.Time
	}
	ch := make(chan result, 1)
	go func() {
		close(started)
		v, err := bq.GetWithContext(context.Background())
		ch <- result{v, err, time.Now()}
	}()

	<-started
	// 确保等待者已就位再 Put，构造“先等待后入队”的唤醒路径。
	time.Sleep(20 * time.Millisecond)

	beforePut := time.Now()
	if err := q.Put("hello"); err != nil {
		t.Fatalf("put: %v", err)
	}

	select {
	case r := <-ch:
		if r.err != nil {
			t.Fatalf("GetWithContext error: %v", r.err)
		}
		if r.value != "hello" {
			t.Fatalf("value = %v, want hello", r.value)
		}
		if latency := r.at.Sub(beforePut); latency > 5*time.Millisecond {
			t.Errorf("wakeup latency %v exceeds 5ms", latency)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("GetWithContext did not wake after Put: lost wakeup")
	}
}

// TestQueue_GetWithContext_PrePutReturnsImmediately 验证已有在队项时立即返回。
func TestQueue_GetWithContext_PrePutReturnsImmediately(t *testing.T) {
	q := NewQueue(NewQueueConfig())
	defer q.Shutdown()

	if err := q.Put(42); err != nil {
		t.Fatalf("put: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	v, err := awaitBlocking(t, blockingGetOf(t, q), ctx)
	if err != nil {
		t.Fatalf("GetWithContext error: %v", err)
	}
	if v != 42 {
		t.Fatalf("value = %v, want 42", v)
	}
}

// TestQueue_GetWithContext_ContextCanceled 验证 ctx 取消返回 ctx.Err()，
// 且队列在取消后仍可继续使用。
func TestQueue_GetWithContext_ContextCanceled(t *testing.T) {
	q := NewQueue(NewQueueConfig())
	defer q.Shutdown()

	bq := blockingGetOf(t, q)

	ctx, cancel := context.WithCancel(context.Background())

	started := make(chan struct{})
	ch := make(chan error, 1)
	go func() {
		close(started)
		_, err := bq.GetWithContext(ctx)
		ch <- err
	}()

	<-started
	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case err := <-ch:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("GetWithContext did not return after ctx cancel")
	}

	// 取消后队列仍可服务新的阻塞消费。
	if err := q.Put("after-cancel"); err != nil {
		t.Fatalf("put: %v", err)
	}
	v, err := awaitBlocking(t, bq, context.Background())
	if err != nil || v != "after-cancel" {
		t.Fatalf("after cancel: v=%v err=%v", v, err)
	}
}

// TestQueue_GetWithContext_ContextTimeout 验证 ctx 超时返回 DeadlineExceeded。
func TestQueue_GetWithContext_ContextTimeout(t *testing.T) {
	q := NewQueue(NewQueueConfig())
	defer q.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := awaitBlocking(t, blockingGetOf(t, q), ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
}

// TestQueue_GetWithContext_NeverReturnsEmpty 验证阻塞路径永不返回 ErrQueueIsEmpty：
// 有值返回值，无值等到 ctx 完成。
func TestQueue_GetWithContext_NeverReturnsEmpty(t *testing.T) {
	q := NewQueue(NewQueueConfig())
	defer q.Shutdown()

	bq := blockingGetOf(t, q)

	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		_, err := bq.GetWithContext(ctx)
		cancel()
		if errors.Is(err, ErrQueueIsEmpty) {
			t.Fatal("GetWithContext must never return ErrQueueIsEmpty")
		}
	}
}

// TestQueue_GetWithContext_ShutdownWakesWaiter 验证 Shutdown 立即唤醒等待者
// 并返回 ErrQueueIsClosed。
func TestQueue_GetWithContext_ShutdownWakesWaiter(t *testing.T) {
	q := NewQueue(NewQueueConfig())
	bq := blockingGetOf(t, q)

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
		if !errors.Is(err, ErrQueueIsClosed) {
			t.Fatalf("err = %v, want ErrQueueIsClosed", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("GetWithContext did not return after Shutdown")
	}
}

// TestQueue_GetWithContext_MultipleWaitersAllReleased 验证关闭时全部等待者被释放，
// 且单值入队只唤醒一个等待者（其余继续等待）。
func TestQueue_GetWithContext_MultipleWaitersAllReleased(t *testing.T) {
	t.Run("AllReleasedOnShutdown", func(t *testing.T) {
		q := NewQueue(NewQueueConfig())
		bq := blockingGetOf(t, q)

		const waiters = 10
		var wg sync.WaitGroup
		errs := make([]error, waiters)
		for i := 0; i < waiters; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				_, errs[idx] = bq.GetWithContext(context.Background())
			}(i)
		}

		time.Sleep(20 * time.Millisecond)
		q.Shutdown()

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("not all waiters released after Shutdown")
		}

		for i, err := range errs {
			if !errors.Is(err, ErrQueueIsClosed) {
				t.Errorf("waiter %d err = %v, want ErrQueueIsClosed", i, err)
			}
		}
	})

	t.Run("SingleValueWakesOne", func(t *testing.T) {
		q := NewQueue(NewQueueConfig())
		defer q.Shutdown()
		bq := blockingGetOf(t, q)

		const waiters = 5
		var got atomic.Int64
		var wg sync.WaitGroup
		for i := 0; i < waiters; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				v, err := bq.GetWithContext(context.Background())
				if err == nil && v == "one" {
					got.Add(1)
				}
			}()
		}

		time.Sleep(20 * time.Millisecond)
		if err := q.Put("one"); err != nil {
			t.Fatalf("put: %v", err)
		}

		deadline := time.Now().Add(time.Second)
		for got.Load() == 0 && time.Now().Before(deadline) {
			time.Sleep(5 * time.Millisecond)
		}
		if got.Load() != 1 {
			t.Fatalf("got = %d, want exactly one consumer served", got.Load())
		}
	})
}

// TestQueue_GetWithContext_DrainConsumesInQueueItems 验证 drain 期间
// GetWithContext 正常消费在队项（助力 drain），drain 完成后返回 ErrQueueIsClosed。
func TestQueue_GetWithContext_DrainConsumesInQueueItems(t *testing.T) {
	q := NewQueue(NewQueueConfig())
	bq := blockingGetOf(t, q)

	const total = 5
	for i := 1; i <= total; i++ {
		if err := q.Put(i); err != nil {
			t.Fatalf("put %d: %v", i, err)
		}
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- q.(DrainableQueue).ShutdownWithDrain(
			context.Background())
	}()

	var got []any
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		v, err := bq.GetWithContext(ctx)
		cancel()
		if err != nil {
			if !errors.Is(err, ErrQueueIsClosed) {
				t.Fatalf("err = %v, want ErrQueueIsClosed after drain", err)
			}
			break
		}
		got = append(got, v)
	}

	if len(got) != total {
		t.Fatalf("consumed %d items during drain, want %d", len(got), total)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("ShutdownWithDrain: %v", err)
	}
}

// TestQueue_GetWithContext_Idempotent_TransferAndRequeue 验证幂等模式下
// pop 的 state/processing 搬运与普通 Get 一致，且 Done 驱动的重入队能唤醒等待者。
func TestQueue_GetWithContext_Idempotent_TransferAndRequeue(t *testing.T) {
	q := NewQueue(NewQueueConfig().WithValueIdempotent())
	defer q.Shutdown()

	bq := blockingGetOf(t, q)

	if err := q.Put("job"); err != nil {
		t.Fatalf("put: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	v, err := bq.GetWithContext(ctx)
	if err != nil || v != "job" {
		t.Fatalf("first get: v=%v err=%v", v, err)
	}

	// 处理中 Put：接受并入挂起标记，不重复入队。
	if err := q.Put("job"); err != nil {
		t.Fatalf("put while processing: %v", err)
	}
	if q.Len() != 0 {
		t.Fatalf("len = %d, want 0 (pending mark only)", q.Len())
	}

	// Done 驱动重入队：等待者应被唤醒拿到同一值。
	blocking := make(chan struct{})
	ch := make(chan any, 1)
	go func() {
		close(blocking)
		v, err := bq.GetWithContext(context.Background())
		if err == nil {
			ch <- v
		}
	}()

	<-blocking
	time.Sleep(10 * time.Millisecond)
	q.Done("job")

	select {
	case v := <-ch:
		if v != "job" {
			t.Fatalf("requeued value = %v, want job", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("waiter not woken by Done requeue: lost wakeup")
	}

	q.Done("job")
	if q.Len() != 0 {
		t.Fatalf("len = %d, want 0 after final Done", q.Len())
	}
}

// TestQueue_GetWithContext_WakeupLatency 采样 Put→唤醒延迟分布：
// 断言宽松阈值 p99<5ms 防 flaky，并记录实测 p99（设计目标 p99<200µs）。
func TestQueue_GetWithContext_WakeupLatency(t *testing.T) {
	q := NewQueue(NewQueueConfig())
	defer q.Shutdown()
	bq := blockingGetOf(t, q)

	const samples = 200
	latencies := make([]time.Duration, 0, samples)

	for i := 0; i < samples; i++ {
		started := make(chan struct{})
		ch := make(chan time.Time, 1)
		go func() {
			close(started)
			_, err := bq.GetWithContext(context.Background())
			if err == nil {
				ch <- time.Now()
			}
		}()

		<-started
		// 留出时间让等待者完成注册。
		time.Sleep(2 * time.Millisecond)

		before := time.Now()
		if err := q.Put(i); err != nil {
			t.Fatalf("put: %v", err)
		}

		select {
		case at := <-ch:
			latencies = append(latencies, at.Sub(before))
		case <-time.After(time.Second):
			t.Fatalf("iteration %d: lost wakeup", i)
		}
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	p99 := latencies[len(latencies)*99/100]
	t.Logf("wakeup latency p50=%v p99=%v max=%v (design target p99<200µs)",
		latencies[len(latencies)/2], p99, latencies[len(latencies)-1])

	if p99 > 5*time.Millisecond {
		t.Errorf("wakeup latency p99=%v exceeds 5ms", p99)
	}
}

// TestQueue_GetWithContext_Stress_MultiProducerMultiConsumer 多生产者并发 Put +
// 多消费者 GetWithContext 压力测试（-race 下运行），校验零丢失与无丢失唤醒。
func TestQueue_GetWithContext_Stress_MultiProducerMultiConsumer(t *testing.T) {
	cases := []struct {
		name   string
		config *QueueConfig
	}{
		{"Plain", NewQueueConfig()},
		{"Idempotent", NewQueueConfig().WithValueIdempotent()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQueue(tc.config)
			bq := blockingGetOf(t, q)

			const producers = 4
			const perProducer = 500
			const consumers = 4
			const total = producers * perProducer

			var consumed atomic.Int64
			var sum atomic.Int64
			var wgConsumer sync.WaitGroup
			for i := 0; i < consumers; i++ {
				wgConsumer.Add(1)
				go func() {
					defer wgConsumer.Done()
					for {
						v, err := bq.GetWithContext(context.Background())
						if err != nil {
							return
						}
						sum.Add(int64(v.(int)))
						consumed.Add(1)
						q.Done(v)
					}
				}()
			}

			var wgProducer sync.WaitGroup
			for p := 0; p < producers; p++ {
				wgProducer.Add(1)
				go func(base int) {
					defer wgProducer.Done()
					for i := 0; i < perProducer; i++ {
						if err := q.Put(base + i); err != nil {
							t.Errorf("put: %v", err)
							return
						}
					}
				}(p * perProducer)
			}

			wgProducer.Wait()

			deadline := time.Now().Add(5 * time.Second)
			for consumed.Load() < total && time.Now().Before(deadline) {
				time.Sleep(5 * time.Millisecond)
			}
			if consumed.Load() != total {
				t.Fatalf("consumed = %d, want %d", consumed.Load(), total)
			}

			q.Shutdown()
			done := make(chan struct{})
			go func() {
				wgConsumer.Wait()
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("consumers did not exit after Shutdown")
			}

			wantSum := int64(total * (total - 1) / 2)
			if sum.Load() != wantSum {
				t.Fatalf("sum = %d, want %d", sum.Load(), wantSum)
			}
		})
	}
}

// TestQueue_GetWithContext_Stress_PutInterleaveWaitRegistration 构造
// “Put 与等待建立交错”的高频窗口：短 ctx 消费者与即时生产者并发对打，
// 最终校验放入总数 = 消费数 + 剩余在队数（无丢失唤醒、无元素丢失）。
func TestQueue_GetWithContext_Stress_PutInterleaveWaitRegistration(t *testing.T) {
	q := NewQueue(NewQueueConfig())
	defer q.Shutdown()
	bq := blockingGetOf(t, q)

	const rounds = 300
	var consumed atomic.Int64

	var wg sync.WaitGroup
	for i := 0; i < rounds; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
			defer cancel()
			v, err := bq.GetWithContext(ctx)
			if err == nil {
				consumed.Add(1)
				q.Done(v)
			}
		}()

		go func(v int) {
			defer wg.Done()
			if err := q.Put(v); err != nil {
				t.Errorf("put: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// 消费者全部退出后，清理可能残留的在队项并核对总数。
	remaining := 0
	for {
		v, err := q.Get()
		if err != nil {
			break
		}
		q.Done(v)
		remaining++
	}

	if got := consumed.Load() + int64(remaining); got != rounds {
		t.Fatalf("consumed(%d) + remaining(%d) = %d, want %d: lost wakeup or lost item",
			consumed.Load(), remaining, got, rounds)
	}
}
