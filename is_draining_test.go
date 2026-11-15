package workqueue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// drainingOf 断言队列实现了可选接口 DrainingQueue（io.Closer 风格契约）。
func drainingOf(t *testing.T, q Queue) DrainingQueue {
	t.Helper()

	dq, ok := q.(DrainingQueue)
	require.True(t, ok, "queue must implement DrainingQueue")

	return dq
}

// isDrainingCase 描述一种队列类型的排空状态可见性场景：
// newQ 构造队列；seed 制造使 drain 判定悬挂的阻塞窗口；
// release 消费被阻塞项使 drain 判定可完成。
type isDrainingCase struct {
	name    string
	newQ    func() Queue
	seed    func(t *testing.T, q Queue)
	release func(t *testing.T, q Queue)
}

// isDrainingCases 枚举九种队列类型的构造与阻塞窗口方案：
// 多数队列为“在队项未消费”窗口；LeasedQueue 额外经“在租未 Ack”窗口
// 覆盖其 leases 表判定；DeadLetterQueue 以 *DeadLetter 值入队。
func isDrainingCases() []isDrainingCase {
	putOne := func(t *testing.T, q Queue) {
		require.NoError(t, q.Put("item"))
	}
	getOne := func(t *testing.T, q Queue) {
		_, err := q.Get()
		require.NoError(t, err)
	}

	return []isDrainingCase{
		{name: "Queue", newQ: func() Queue { return NewQueue(nil) }, seed: putOne, release: getOne},
		{name: "DelayingQueue", newQ: func() Queue { return NewDelayingQueue(nil) }, seed: putOne, release: getOne},
		{name: "PriorityQueue", newQ: func() Queue { return NewPriorityQueue(nil) }, seed: putOne, release: getOne},
		{name: "RateLimitingQueue", newQ: func() Queue { return NewRateLimitingQueue(nil) }, seed: putOne, release: getOne},
		{name: "RetryQueue", newQ: func() Queue { return NewRetryQueue(nil) }, seed: putOne, release: getOne},
		{
			name: "DeadLetterQueue",
			newQ: func() Queue { return NewDeadLetterQueue(nil) },
			seed: func(t *testing.T, q Queue) {
				require.NoError(t, q.Put(&DeadLetter{Payload: "item"}))
			},
			release: getOne,
		},
		{
			name: "LeasedQueue",
			newQ: func() Queue { return NewLeasedQueue(nil) },
			seed: putOne,
			release: func(t *testing.T, q Queue) {
				lq := q.(LeasedQueue)
				_, leaseID, err := lq.GetWithLease(time.Second)
				require.NoError(t, err)
				require.NoError(t, lq.Ack(leaseID))
			},
		},
		{name: "BoundedBlockingQueue", newQ: func() Queue { return NewBoundedBlockingQueue(nil) }, seed: putOne, release: getOne},
		{name: "TimerQueue", newQ: func() Queue { return NewTimerQueue(nil) }, seed: putOne, release: getOne},
	}
}

// TestIsDraining_DuringDrain 覆盖九种队列类型的排空状态可见性：
// a) 初始 IsDraining=false；b) ShutdownWithDrain 悬挂期间 IsDraining=true
// 且 IsClosed 保持 false（生产者据此区分“排空中”与“已关闭”）；
// c) drain 完成后两者皆 true（排空完成已关闭）；d) 经 DrainingQueue
// 类型断言可用（io.Closer 惯例）。
func TestIsDraining_DuringDrain(t *testing.T) {
	for _, c := range isDrainingCases() {
		t.Run(c.name, func(t *testing.T) {
			q := c.newQ()

			// d) 可选接口类型断言。
			dq := drainingOf(t, q)

			// a) 初始未在排空。
			assert.False(t, dq.IsDraining(), "fresh queue must not be draining")
			assert.False(t, q.IsClosed())

			// 制造 drain 判定悬挂的阻塞窗口。
			c.seed(t, q)

			errCh := make(chan error, 1)
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				errCh <- drainableOf(t, q).ShutdownWithDrain(ctx)
			}()

			// b) drain 进行中：IsDraining=true ∧ IsClosed=false。
			require.Eventually(t, dq.IsDraining, 2*time.Second, time.Millisecond,
				"IsDraining must become true once ShutdownWithDrain starts")
			assert.False(t, q.IsClosed(), "IsClosed must stay false while draining")

			// 解除阻塞，drain 判定应可完成。
			c.release(t, q)

			select {
			case err := <-errCh:
				require.NoError(t, err, "drain must complete after release")
			case <-time.After(3 * time.Second):
				t.Fatal("drain did not complete after release")
			}

			// 排空完成：两者皆 true。
			assert.True(t, dq.IsDraining(), "IsDraining stays true after drain completes")
			assert.True(t, q.IsClosed(), "queue must be closed after drain")
		})
	}
}

// TestIsDraining_FalseAfterImmediateShutdown 覆盖经 Shutdown() 立即关闭的路径：
// draining 从未置位，IsDraining 恒为 false 且 IsClosed 为 true。
func TestIsDraining_FalseAfterImmediateShutdown(t *testing.T) {
	for _, c := range isDrainingCases() {
		t.Run(c.name, func(t *testing.T) {
			q := c.newQ()
			dq := drainingOf(t, q)

			q.Shutdown()

			assert.False(t, dq.IsDraining(), "immediate Shutdown must not set draining")
			assert.True(t, q.IsClosed())
		})
	}
}
