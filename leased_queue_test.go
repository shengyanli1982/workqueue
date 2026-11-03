package workqueue

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
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

// recordingLeasedQueueCallback 记录全部回调事件，用于验证 OnNack 及内嵌路径基础事件。
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

// TestLeasedQueue_Nack_OnNackCallback 验证 Nack 的 reason 真实传递到 OnNack 回调（D4a）。
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
	// 基础回调经内嵌路径正常：初始 Put + Nack 重入队 = 2 次 OnPut。
	assert.Equal(t, []any{"job-nack", "job-nack"}, puts)
	assert.Equal(t, []any{"job-nack"}, gets)
	// 非幂等模式 Done 为 best-effort 完成信号：Nack 内部先 Put 重入队再 Done 释放租约。
	assert.Equal(t, []any{"job-nack"}, dones)
}

// TestLeasedQueue_Callback_BaseEventsViaEmbeddedPath 验证 Ack 路径基础事件正常且不误触发 OnNack。
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

// TestLeasedQueue_Nack_NopCallbackDefault 验证未配置回调时 Nack 正常且不 panic。
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
// 幂等模式下租约过期后元素应重新入队可再次 Get。
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
// 幂等模式下 Nack 应返回 nil，元素重新入队可再次 Get。
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

// TestLeasedQueue_Idempotent_Ack_NoRequeue 验证幂等模式下 Ack 后元素不再重入队。
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

	// 跨过 reaper 扫描窗口，确认无误重入队。
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

// TestLeasedQueue_LeaseInfos 验证 G5：LeaseInfos 暴露在租租约的只读快照。
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

	// Ack 后租约从快照消失。
	assert.NoError(t, q.Ack(leaseA))
	infos = q.LeaseInfos()
	assert.Len(t, infos, 1)
	assert.Equal(t, leaseB, infos[0].LeaseID)
}

// TestLeasedQueue_LeaseInfos_Empty 验证无在租租约时返回空非 nil 切片。
func TestLeasedQueue_LeaseInfos_Empty(t *testing.T) {
	config := NewLeasedQueueConfig().WithScanInterval(time.Hour)
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	infos := q.LeaseInfos()
	assert.NotNil(t, infos)
	assert.Empty(t, infos)
}

// TestLeasedQueue_LeaseInfos_CopyMutation 验证快照为副本，篡改不影响内部状态。
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

// TestLeasedQueue_LeaseInfos_SnapshotDetached 验证快照脱锁，并发操作不被阻塞。
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

// TestLeasedQueue_Reaper_AckRace_NoResurrectNoStolenDone 复现 P1 #2 竞态链：
// reaper 陈旧快照不得复活已 Ack 的元素或窃取在途计数。手动驱动交错，不依赖时序。
func TestLeasedQueue_Reaper_AckRace_NoResurrectNoStolenDone(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(time.Minute).
		WithScanInterval(time.Hour) // 关闭后台 reaper 干扰，由测试手动驱动交错
	config.WithDrainTracking()
	q := NewLeasedQueue(config).(*leasedQueueImpl)
	defer q.Shutdown()

	assert.NoError(t, q.Put("v1"))
	_, lease1, err := q.GetWithLease(time.Minute)
	assert.NoError(t, err)

	// reaper 持锁快照（假“未来时间”使租约立即判定过期），随后释放锁。
	expired := q.collectExpired(time.Now().Add(2*time.Minute), nil)
	assert.Len(t, expired, 1)

	// 快照窗口内消费者 Ack 抢先完整执行：removeLease+Done，返回 nil 即承诺终态。
	assert.NoError(t, q.Ack(lease1))

	// 另一元素在租：其在途计数是“误调 Done”的可观测受害者。
	assert.NoError(t, q.Put("v2"))
	_, lease2, err := q.GetWithLease(time.Minute)
	assert.NoError(t, err)
	inner := q.Queue.(*queueImpl)
	assert.Equal(t, int64(1), inner.inFlight.Load())

	// reaper 处理陈旧快照。
	q.requeueCollected(expired)

	assert.Equal(t, 0, q.Len(), "acked value must not be resurrected by a stale reaper snapshot")
	assert.Equal(t, int64(1), inner.inFlight.Load(), "stale snapshot must not steal v2's in-flight count")
	infos := q.LeaseInfos()
	assert.Len(t, infos, 1)
	assert.Equal(t, lease2, infos[0].LeaseID)
}

// TestLeasedQueue_Reaper_StaleSnapshotAfterNack_NoDuplicate 复现陈旧快照与 Nack 交错：
// reaper 不得对已 Nack 归还的元素二次 Put 产生重复。
func TestLeasedQueue_Reaper_StaleSnapshotAfterNack_NoDuplicate(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(time.Minute).
		WithScanInterval(time.Hour)
	q := NewLeasedQueue(config).(*leasedQueueImpl)
	defer q.Shutdown()

	assert.NoError(t, q.Put("v1"))
	_, lease1, err := q.GetWithLease(time.Minute)
	assert.NoError(t, err)

	expired := q.collectExpired(time.Now().Add(2*time.Minute), nil)
	assert.Len(t, expired, 1)

	// 快照窗口内 Nack 完整执行。
	assert.NoError(t, q.Nack(lease1, errors.New("retry")))
	assert.Equal(t, 1, q.Len())

	q.requeueCollected(expired)

	assert.Equal(t, 1, q.Len(), "stale reaper snapshot must not duplicate a nacked value")
	assert.Equal(t, []any{"v1"}, q.Values())
	assert.Empty(t, q.LeaseInfos())
}

// TestLeasedQueue_Reaper_ExtendLeaseRace_NoStaleRequeue 复现陈旧快照与 ExtendLease 交错：
// reaper 不得用陈旧快照重入队已续期的租约。手动驱动交错，不依赖时序。
func TestLeasedQueue_Reaper_ExtendLeaseRace_NoStaleRequeue(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(time.Minute).
		WithScanInterval(time.Hour) // 关闭后台 reaper 干扰，由测试手动驱动交错
	q := NewLeasedQueue(config).(*leasedQueueImpl)
	defer q.Shutdown()

	assert.NoError(t, q.Put("v1"))
	_, lease1, err := q.GetWithLease(time.Minute)
	assert.NoError(t, err)

	// reaper 持锁快照（假“未来时间”使租约立即判定过期），随后释放锁。
	expired := q.collectExpired(time.Now().Add(2*time.Minute), nil)
	assert.Len(t, expired, 1)

	// 快照窗口内消费者 ExtendLease 续期：新 deadline 落在未来，租约不再应被视为过期。
	assert.NoError(t, q.ExtendLease(lease1, time.Hour))
	extended := q.LeaseInfos()[0].Deadline

	// reaper 恢复执行，处理陈旧快照。
	q.requeueCollected(expired)

	assert.Equal(t, 0, q.Len(), "an extended lease must not be requeued by a stale reaper snapshot")
	infos := q.LeaseInfos()
	if assert.Len(t, infos, 1, "the extended lease must survive the stale snapshot") {
		assert.Equal(t, lease1, infos[0].LeaseID)
		assert.Equal(t, "v1", infos[0].Value)
		assert.True(t, infos[0].Deadline.Equal(extended), "lease must keep its extended deadline")
	}
}

// nackWindowReaperCallback 在 OnPut 内触发 reaper 处理陈旧快照，模拟 Nack 窗口交错。
// fired 保证钩子只触发一次，避免递归。
type nackWindowReaperCallback struct {
	leasedQueueCallbackImpl
	q       *leasedQueueImpl
	expired []expiredLease
	fired   bool
}

func (c *nackWindowReaperCallback) OnPut(value any) {
	if c.fired || c.expired == nil {
		return
	}
	c.fired = true
	c.q.requeueCollected(c.expired)
}

// TestLeasedQueue_Nack_ReaperRaceInPutWindow_NoDuplicate 复现 Nack 与 reaper 的
// 第二个交错窗口：经 OnPut 钩子注入"Nack Put 后、removeLease 前"的抢占。
func TestLeasedQueue_Nack_ReaperRaceInPutWindow_NoDuplicate(t *testing.T) {
	cb := &nackWindowReaperCallback{}
	config := NewLeasedQueueConfig().
		WithLeaseDuration(time.Minute).
		WithScanInterval(time.Hour).
		WithCallback(cb)
	q := NewLeasedQueue(config).(*leasedQueueImpl)
	cb.q = q
	defer q.Shutdown()

	assert.NoError(t, q.Put("v1")) // expired==nil，钩子不动作
	_, lease1, err := q.GetWithLease(time.Minute)
	assert.NoError(t, err)

	expired := q.collectExpired(time.Now().Add(2*time.Minute), nil)
	assert.Len(t, expired, 1)
	cb.expired = expired // 装填钩子：Nack 归还入队时 reaper 抢占

	assert.NoError(t, q.Nack(lease1, errors.New("retry")))
	assert.True(t, cb.fired, "OnPut hook must have run the stale-snapshot reaper")

	assert.Equal(t, 1, q.Len(), "reaper must not duplicate the value Nack is returning")
	assert.Equal(t, []any{"v1"}, q.Values())
	assert.Empty(t, q.LeaseInfos())
}

// TestLeasedQueue_Reaper_PutFail_ReregistersLease 钉住 reaper 的“入队失败不丢元素”
// 语义：claim 后 Put 失败（内层已关停）时，租约按原 deadline 重新登记，待下轮重试。
func TestLeasedQueue_Reaper_PutFail_ReregistersLease(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(time.Minute).
		WithScanInterval(time.Hour)
	q := NewLeasedQueue(config).(*leasedQueueImpl)
	defer q.Shutdown()

	assert.NoError(t, q.Put("v1"))
	_, lease1, err := q.GetWithLease(time.Minute)
	assert.NoError(t, err)

	before := q.LeaseInfos()[0].Deadline
	expired := q.collectExpired(time.Now().Add(2*time.Minute), nil)
	assert.Len(t, expired, 1)

	// 内层队列关停，requeueCollected 的 Put 必然失败。
	q.Queue.Shutdown()

	q.requeueCollected(expired)

	infos := q.LeaseInfos()
	if assert.Len(t, infos, 1, "lease must survive a failed requeue") {
		assert.Equal(t, lease1, infos[0].LeaseID)
		assert.Equal(t, "v1", infos[0].Value)
		assert.True(t, infos[0].Deadline.Equal(before), "lease must be re-registered with its original deadline")
	}
}

// TestLeasedQueue_Nack_PutFail_KeepsLease 钉住 Nack 的“入队失败不丢元素”语义：
// Put 失败时返回错误且租约按原 deadline 保留，元素可经重试或 reaper 回收。
func TestLeasedQueue_Nack_PutFail_KeepsLease(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(time.Minute).
		WithScanInterval(time.Hour)
	q := NewLeasedQueue(config).(*leasedQueueImpl)
	defer q.Shutdown()

	assert.NoError(t, q.Put("v1"))
	_, lease1, err := q.GetWithLease(time.Minute)
	assert.NoError(t, err)

	before := q.LeaseInfos()[0].Deadline

	// 内层队列关停：Nack 的 Put 必然失败（ErrQueueIsClosed）。
	q.Queue.Shutdown()

	err = q.Nack(lease1, errors.New("retry"))
	assert.ErrorIs(t, err, ErrQueueIsClosed)

	infos := q.LeaseInfos()
	if assert.Len(t, infos, 1, "lease must survive a failed Nack requeue") {
		assert.Equal(t, lease1, infos[0].LeaseID)
		assert.Equal(t, "v1", infos[0].Value)
		assert.True(t, infos[0].Deadline.Equal(before), "lease must keep its original deadline")
	}
}

// onGetReentrantCallback 在 OnGet 内重入租约队列 API，验证外层持锁时重入不死锁。
type onGetReentrantCallback struct {
	leasedQueueCallbackImpl
	q      LeasedQueue
	probed chan struct{}
	once   sync.Once
}

func (c *onGetReentrantCallback) OnGet(value any) {
	_ = c.q.LeaseInfos()
	_ = c.q.Ack("missing")
	_ = c.q.ExtendLease("missing", time.Second)
	c.once.Do(func() { close(c.probed) })
}

// TestLeasedQueue_GetWithLease_OnGetReentrant_NoDeadlock 复现 P1 #3：
// OnGet 回调内重入 LeaseInfos/Ack 不得死锁。以超时保护判定挂死。
func TestLeasedQueue_GetWithLease_OnGetReentrant_NoDeadlock(t *testing.T) {
	cb := &onGetReentrantCallback{probed: make(chan struct{})}
	config := NewLeasedQueueConfig().
		WithLeaseDuration(time.Minute).
		WithScanInterval(time.Hour).
		WithCallback(cb)
	q := NewLeasedQueue(config)
	cb.q = q

	assert.NoError(t, q.Put("job"))

	type getResult struct {
		value   any
		leaseID string
		err     error
	}
	done := make(chan getResult, 1)
	go func() {
		value, leaseID, err := q.GetWithLease(time.Minute)
		done <- getResult{value: value, leaseID: leaseID, err: err}
	}()

	select {
	case r := <-done:
		assert.NoError(t, r.err)
		assert.Equal(t, "job", r.value)
		assert.NotEmpty(t, r.leaseID)
		select {
		case <-cb.probed:
		default:
			t.Fatal("OnGet reentrant probe did not run")
		}
	case <-time.After(3 * time.Second):
		// 死锁态下调 Shutdown 会挂死测试进程。
		t.Fatal("GetWithLease deadlocked: OnGet re-entered lease APIs while q.lock was held")
	}

	q.Shutdown()
}

// TestLeasedQueue_Concurrent_MixedOps_ShutdownWithDrain -race 并发回归：
// 混合操作 + ShutdownWithDrain，断言无死锁、零丢失。
func TestLeasedQueue_Concurrent_MixedOps_ShutdownWithDrain(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(5 * time.Millisecond).
		WithScanInterval(time.Millisecond)
	config.WithDrainTracking()
	q := NewLeasedQueue(config)

	const total = 200
	for i := 0; i < total; i++ {
		assert.NoError(t, q.Put(i))
	}

	var acked atomic.Int64
	var workers sync.WaitGroup

	// 非阻塞消费者：随机 Ack/Nack/持有过期，关停后退出。
	for w := 0; w < 4; w++ {
		workers.Add(1)
		go func(seed int64) {
			defer workers.Done()
			rnd := rand.New(rand.NewSource(seed))
			for {
				_, leaseID, err := q.GetWithLease(5 * time.Millisecond)
				if err != nil {
					if errors.Is(err, ErrQueueIsClosed) {
						return
					}
					// 队列暂空，短暂等待后重试。
					time.Sleep(time.Millisecond)
					continue
				}
				switch rnd.Intn(10) {
				case 0, 1: // 持有不 Ack，由 reaper 回收
				case 2, 3:
					_ = q.Nack(leaseID, errors.New("stress"))
				default:
					if q.Ack(leaseID) == nil {
						acked.Add(1)
					}
				}
			}
		}(int64(w))
	}

	// 阻塞消费者：覆盖 GetWithLeaseWithContext 关停退出路径。
	blocking := leasedBlockingOf(t, q)
	for w := 0; w < 2; w++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for {
				_, leaseID, err := blocking.GetWithLeaseWithContext(context.Background(), 5*time.Millisecond)
				if err != nil {
					if errors.Is(err, ErrQueueIsClosed) {
						return
					}
					continue
				}
				if q.Ack(leaseID) == nil {
					acked.Add(1)
				}
			}
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	assert.NoError(t, drainableOf(t, q).ShutdownWithDrain(ctx), "drain must complete: no deadlock")

	workersDone := make(chan struct{})
	go func() {
		workers.Wait()
		close(workersDone)
	}()
	select {
	case <-workersDone:
	case <-time.After(10 * time.Second):
		t.Fatal("workers did not exit after ShutdownWithDrain: goroutine leak or deadlock")
	}

	assert.Equal(t, int64(total), acked.Load(), "every element must be acked exactly once (zero loss, no premature drain)")
	assert.True(t, q.IsClosed())
	assert.Equal(t, 0, q.Len())
	assert.Empty(t, q.LeaseInfos())
}
