package workqueue

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLeasedQueue_GetWithLease_ExpireRequeue(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(20 * time.Millisecond).
		WithScanInterval(5 * time.Millisecond)
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("job-1"))

	value, leaseID, err := q.GetWithLease(15 * time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "job-1", value)
	assert.NotEmpty(t, leaseID)

	requeued, err := waitQueueGet(t, q, 200*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "job-1", requeued)
}

func TestLeasedQueue_Ack_NoRequeue(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(20 * time.Millisecond).
		WithScanInterval(5 * time.Millisecond)
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("job-2"))

	_, leaseID, err := q.GetWithLease(15 * time.Millisecond)
	assert.NoError(t, err)
	assert.NoError(t, q.Ack(leaseID))

	time.Sleep(40 * time.Millisecond)

	_, err = q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty)
}

func TestLeasedQueue_Nack_Requeue(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(100 * time.Millisecond).
		WithScanInterval(5 * time.Millisecond)
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("job-3"))

	_, leaseID, err := q.GetWithLease(80 * time.Millisecond)
	assert.NoError(t, err)
	assert.NoError(t, q.Nack(leaseID, errors.New("retry")))

	requeued, err := waitQueueGet(t, q, 100*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "job-3", requeued)
}

func TestLeasedQueue_ExtendLease(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(20 * time.Millisecond).
		WithScanInterval(5 * time.Millisecond)
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("job-4"))

	_, leaseID, err := q.GetWithLease(15 * time.Millisecond)
	assert.NoError(t, err)
	assert.NoError(t, q.ExtendLease(leaseID, 40*time.Millisecond))

	time.Sleep(20 * time.Millisecond)
	_, err = q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty)

	requeued, err := waitQueueGet(t, q, 120*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "job-4", requeued)
}

func TestLeasedQueue_InvalidLease(t *testing.T) {
	q := NewLeasedQueue(nil)
	defer q.Shutdown()

	assert.ErrorIs(t, q.Ack("missing"), ErrLeaseNotFound)
	assert.ErrorIs(t, q.Nack("missing", nil), ErrLeaseNotFound)
	assert.ErrorIs(t, q.ExtendLease("missing", 10*time.Millisecond), ErrLeaseNotFound)
	assert.ErrorIs(t, q.ExtendLease("missing", 0), ErrInvalidLeaseDuration)
}

type nackRecord struct {
	value  any
	reason error
}

// recordingLeasedQueueCallback 记录全部回调事件，用于验证
// LeasedQueueCallback 的 OnNack 以及经内嵌 QueueCallback 路径的基础事件。
type recordingLeasedQueueCallback struct {
	mu    sync.Mutex
	puts  []any
	gets  []any
	dones []any
	nacks []nackRecord
}

func (c *recordingLeasedQueueCallback) OnPut(value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.puts = append(c.puts, value)
}

func (c *recordingLeasedQueueCallback) OnGet(value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gets = append(c.gets, value)
}

func (c *recordingLeasedQueueCallback) OnDone(value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dones = append(c.dones, value)
}

func (c *recordingLeasedQueueCallback) OnNack(value any, reason error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nacks = append(c.nacks, nackRecord{value: value, reason: reason})
}

func (c *recordingLeasedQueueCallback) snapshot() (puts []any, gets []any, dones []any, nacks []nackRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]any(nil), c.puts...),
		append([]any(nil), c.gets...),
		append([]any(nil), c.dones...),
		append([]nackRecord(nil), c.nacks...)
}

// TestLeasedQueue_Nack_OnNackCallback 验证 Nack 的 reason 参数真实传递到 OnNack 回调（D4a）。
// 修复前：Nack(leaseID, _ error) 直接丢弃 reason，且 LeasedQueueCallback 不存在（编译失败 RED）；
// 修复后：OnNack 恰好触发一次，收到正确的 value 与 reason。
func TestLeasedQueue_Nack_OnNackCallback(t *testing.T) {
	cb := &recordingLeasedQueueCallback{}
	config := NewLeasedQueueConfig().
		WithLeaseDuration(100 * time.Millisecond).
		WithScanInterval(5 * time.Millisecond).
		WithCallback(cb)
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("job-nack"))
	_, leaseID, err := q.GetWithLease(80 * time.Millisecond)
	assert.NoError(t, err)

	nackReason := errors.New("downstream timeout")
	assert.NoError(t, q.Nack(leaseID, nackReason))

	puts, gets, dones, nacks := cb.snapshot()
	if assert.Len(t, nacks, 1, "Nack should trigger OnNack exactly once") {
		assert.Equal(t, "job-nack", nacks[0].value)
		assert.Same(t, nackReason, nacks[0].reason, "OnNack must receive the exact reason error")
	}
	// 基础回调经内嵌 QueueConfig 路径保持正常：初始 Put + Nack 重入队 = 2 次 OnPut。
	assert.Equal(t, []any{"job-nack", "job-nack"}, puts)
	assert.Equal(t, []any{"job-nack"}, gets)
	// 非幂等模式下 Done 为 no-op，不触发 OnDone（既有语义，见 queueImpl.Done）。
	assert.Empty(t, dones)
}

// TestLeasedQueue_Callback_BaseEventsViaEmbeddedPath 验证 WithCallback shadow 布线后
// Ack 路径的基础事件正常（幂等模式下 OnDone 经内嵌路径触发）且不误触发 OnNack。
func TestLeasedQueue_Callback_BaseEventsViaEmbeddedPath(t *testing.T) {
	cb := &recordingLeasedQueueCallback{}
	config := NewLeasedQueueConfig().
		WithLeaseDuration(100 * time.Millisecond).
		WithScanInterval(5 * time.Millisecond).
		WithCallback(cb)
	config.WithValueIdempotent()
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("job-ack"))
	_, leaseID, err := q.GetWithLease(80 * time.Millisecond)
	assert.NoError(t, err)
	assert.NoError(t, q.Ack(leaseID))

	puts, gets, dones, nacks := cb.snapshot()
	assert.Equal(t, []any{"job-ack"}, puts)
	assert.Equal(t, []any{"job-ack"}, gets)
	assert.Equal(t, []any{"job-ack"}, dones)
	assert.Empty(t, nacks, "Ack must not trigger OnNack")
}

// TestLeasedQueue_Nack_NopCallbackDefault 验证未配置回调（nil config / 默认 Nop）时
// Nack 正常工作且不 panic。
func TestLeasedQueue_Nack_NopCallbackDefault(t *testing.T) {
	q := NewLeasedQueue(nil)
	defer q.Shutdown()

	assert.NoError(t, q.Put("job-nop"))
	_, leaseID, err := q.GetWithLease(80 * time.Millisecond)
	assert.NoError(t, err)
	assert.NoError(t, q.Nack(leaseID, errors.New("retry")))

	requeued, err := waitQueueGet(t, q, 100*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "job-nop", requeued)
}

// TestLeasedQueue_Idempotent_RequeueAfterLeaseExpiry 复现 D1 [P0] 租约过期路径：
// 幂等模式下 GetWithLease 取走元素后，元素离开 list 但仍留在幂等 state 集合中。
// 租约过期时 reaper 的 Put 撞 ErrElementAlreadyExist → continue 跳过 removeLease，
// 每个扫描周期空转一次，元素永不回队。
// 修复前：等待窗口内元素无法再次 Get（waitQueueGet 超时）；
// 修复后：reaper 的 Put 命中 processing 挂起返回 nil，Put→removeLease→Done 顺序下
// Done 触发重入队，元素可再次 Get。
func TestLeasedQueue_Idempotent_RequeueAfterLeaseExpiry(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(20 * time.Millisecond).
		WithScanInterval(5 * time.Millisecond)
	config.WithValueIdempotent()
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("lease-expire"))

	value, _, err := q.GetWithLease(10 * time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "lease-expire", value)

	requeued, err := waitQueueGet(t, q, time.Second)
	assert.NoError(t, err, "expired lease must be requeued and the element gettable again (D1)")
	assert.Equal(t, "lease-expire", requeued)
}

// TestLeasedQueue_Idempotent_NackRequeue 复现 D1 [P0] 主动 Nack 路径：
// 幂等模式下 Nack 先 Put 后释放租约，Put 撞 ErrElementAlreadyExist 直接返回错误，
// 租约不释放，元素随后落入过期循环永久卡死。
// 修复前：Nack 返回 ErrElementAlreadyExist；
// 修复后：Nack 返回 nil，元素重新入队可再次 Get。
func TestLeasedQueue_Idempotent_NackRequeue(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(100 * time.Millisecond).
		WithScanInterval(5 * time.Millisecond)
	config.WithValueIdempotent()
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("lease-nack"))

	value, leaseID, err := q.GetWithLease(80 * time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "lease-nack", value)

	assert.NoError(t, q.Nack(leaseID, errors.New("downstream timeout")), "Nack must succeed in idempotent mode (D1)")

	requeued, err := waitQueueGet(t, q, 100*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "lease-nack", requeued)
}

// TestLeasedQueue_Idempotent_Ack_NoRequeue 验证幂等模式下 Ack 的正常生命周期出口：
// Ack 后元素不得重新入队（新语义下 processing.TryRemove 成功且 state 无挂起标记 → 生命周期完成）。
func TestLeasedQueue_Idempotent_Ack_NoRequeue(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(20 * time.Millisecond).
		WithScanInterval(5 * time.Millisecond)
	config.WithValueIdempotent()
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("lease-ack"))

	value, leaseID, err := q.GetWithLease(15 * time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "lease-ack", value)

	assert.NoError(t, q.Ack(leaseID))

	// 跨过 reaper 扫描窗口，确认没有被误重入队。
	time.Sleep(40 * time.Millisecond)
	_, err = q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty, "Ack must not requeue the element")
}

func waitQueueGet(t *testing.T, q Queue, timeout time.Duration) (any, error) {
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

// TestLeasedQueue_LeaseInfos 验证 G5：LeaseInfos 暴露全部在租租约的只读快照，
// LeaseID/Value/Deadline 三类盲区数据均正确可见。
func TestLeasedQueue_LeaseInfos(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(time.Minute).
		WithScanInterval(time.Hour)
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("task-a"))
	assert.NoError(t, q.Put("task-b"))

	before := time.Now()
	_, leaseA, err := q.GetWithLease(2 * time.Minute)
	assert.NoError(t, err)
	_, leaseB, err := q.GetWithLease(0) // timeout<=0 回退 leaseDuration=time.Minute
	assert.NoError(t, err)

	infos := q.LeaseInfos()
	assert.Len(t, infos, 2)

	byID := make(map[string]LeaseInfo, len(infos))
	for _, info := range infos {
		byID[info.LeaseID] = info
	}

	infoA, ok := byID[leaseA]
	assert.True(t, ok, "lease A must be visible")
	assert.Equal(t, "task-a", infoA.Value)
	assert.True(t, infoA.Deadline.After(before.Add(time.Minute)), "deadline must honor the explicit 2m timeout")
	assert.True(t, infoA.Deadline.Before(before.Add(3*time.Minute)))

	infoB, ok := byID[leaseB]
	assert.True(t, ok, "lease B must be visible")
	assert.Equal(t, "task-b", infoB.Value)
	assert.True(t, infoB.Deadline.After(before.Add(30*time.Second)), "deadline must honor the default 1m duration")
	assert.True(t, infoB.Deadline.Before(before.Add(2*time.Minute)))

	// Ack 释放租约后立即从快照消失。
	assert.NoError(t, q.Ack(leaseA))
	infos = q.LeaseInfos()
	assert.Len(t, infos, 1)
	assert.Equal(t, leaseB, infos[0].LeaseID)
}

// TestLeasedQueue_LeaseInfos_Empty 验证无在租租约时返回空且非 nil 的切片。
func TestLeasedQueue_LeaseInfos_Empty(t *testing.T) {
	config := NewLeasedQueueConfig().WithScanInterval(time.Hour)
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	infos := q.LeaseInfos()
	assert.NotNil(t, infos)
	assert.Empty(t, infos)
}

// TestLeasedQueue_LeaseInfos_CopyMutation 验证快照为副本：篡改返回值不影响内部租约表。
func TestLeasedQueue_LeaseInfos_CopyMutation(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(time.Minute).
		WithScanInterval(time.Hour)
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("task"))
	value, leaseID, err := q.GetWithLease(0)
	assert.NoError(t, err)
	assert.Equal(t, "task", value)

	snapshot := q.LeaseInfos()
	snapshot[0].LeaseID = "tampered"
	snapshot[0].Value = "tampered"
	snapshot[0].Deadline = time.Time{}

	fresh := q.LeaseInfos()
	assert.Len(t, fresh, 1)
	assert.Equal(t, leaseID, fresh[0].LeaseID)
	assert.Equal(t, "task", fresh[0].Value)
	assert.False(t, fresh[0].Deadline.IsZero())
}

// TestLeasedQueue_LeaseInfos_SnapshotDetached 验证快照脱锁：取快照后并发
// Put/GetWithLease/Ack 不被阻塞、不死锁。
func TestLeasedQueue_LeaseInfos_SnapshotDetached(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(time.Minute).
		WithScanInterval(time.Hour)
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	stop := make(chan struct{})
	trafficDone := make(chan struct{})
	go func() {
		defer close(trafficDone)
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			if q.Put(i) == nil {
				if _, leaseID, err := q.GetWithLease(time.Minute); err == nil {
					_ = q.Ack(leaseID)
				}
			}
		}
	}()

	snapDone := make(chan struct{})
	go func() {
		defer close(snapDone)
		for i := 0; i < 200; i++ {
			_ = q.LeaseInfos()
		}
	}()

	select {
	case <-snapDone:
	case <-time.After(3 * time.Second):
		close(stop)
		<-trafficDone
		t.Fatal("LeaseInfos blocked while concurrent Put/GetWithLease/Ack traffic is running")
	}

	close(stop)
	<-trafficDone
}
