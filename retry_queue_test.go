package workqueue

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryQueue_RetryAndRequeue(t *testing.T) {
	config := NewRetryQueueConfig().WithPolicy(NewExponentialRetryPolicy(50*time.Millisecond, 50*time.Millisecond, 3))
	q := NewRetryQueue(config)
	defer q.Shutdown()

	err := q.Put("task")
	assert.NoError(t, err)

	value, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "task", value)

	err = q.Retry(value, errors.New("failed"))
	assert.NoError(t, err)
	assert.Equal(t, 1, q.NumRequeues("task"))

	assert.Eventually(t, func() bool {
		v, getErr := q.Get()
		if getErr != nil {
			return false
		}
		return v == "task"
	}, 2*time.Second, 20*time.Millisecond)
}

func TestRetryQueue_RetryExhausted(t *testing.T) {
	config := NewRetryQueueConfig().WithPolicy(NewExponentialRetryPolicy(0, 0, 1))
	q := NewRetryQueue(config)
	defer q.Shutdown()

	err := q.Put("task")
	assert.NoError(t, err)

	value, err := q.Get()
	assert.NoError(t, err)

	err = q.Retry(value, errors.New("first failed"))
	assert.NoError(t, err)

	var ok bool
	assert.Eventually(t, func() bool {
		value, err = q.Get()
		ok = err == nil
		return ok
	}, 2*time.Second, 20*time.Millisecond)
	assert.True(t, ok)

	err = q.Retry(value, errors.New("second failed"))
	assert.ErrorIs(t, err, ErrRetryExhausted)
	assert.Equal(t, 0, q.NumRequeues("task"))
}

func TestRetryQueue_Forget(t *testing.T) {
	config := NewRetryQueueConfig().WithPolicy(NewExponentialRetryPolicy(0, 0, 3))
	q := NewRetryQueue(config)
	defer q.Shutdown()

	err := q.Put("task")
	assert.NoError(t, err)

	value, err := q.Get()
	assert.NoError(t, err)

	err = q.Retry(value, errors.New("failed"))
	assert.NoError(t, err)
	assert.Equal(t, 1, q.NumRequeues("task"))

	q.Forget("task")
	assert.Equal(t, 0, q.NumRequeues("task"))
}

func TestRetryQueue_EmptyRetryKey(t *testing.T) {
	config := NewRetryQueueConfig().
		WithKeyFunc(func(any) string { return "" }).
		WithPolicy(NewExponentialRetryPolicy(0, 0, 1))
	q := NewRetryQueue(config)
	defer q.Shutdown()

	err := q.Put("task")
	assert.NoError(t, err)

	value, err := q.Get()
	assert.NoError(t, err)

	err = q.Retry(value, errors.New("failed"))
	assert.ErrorIs(t, err, ErrRetryKeyEmpty)
}

// immediateRetryPolicy 始终返回零延迟，使 Retry 走立即路径（delay < 1ms → 直接 Put）。
type immediateRetryPolicy struct {
	maxRetries int
}

func (p *immediateRetryPolicy) NextDelay(_ any, attempt int, _ error) (time.Duration, bool) {
	if p.maxRetries >= 0 && attempt > p.maxRetries {
		return 0, false
	}
	return 0, true
}

// fixedDelayRetryPolicy 返回固定延迟，用于延迟路径（PutWithDelay）测试。
type fixedDelayRetryPolicy struct {
	delay      time.Duration
	maxRetries int
}

func (p *fixedDelayRetryPolicy) NextDelay(_ any, attempt int, _ error) (time.Duration, bool) {
	if p.maxRetries >= 0 && attempt > p.maxRetries {
		return 0, false
	}
	return p.delay, true
}

// TestRetryQueue_Idempotent_RetryImmediate 复现 D2 [P0] 立即路径：
// 幂等模式下 Get 走元素后元素仍留在 state，Retry 的 “先 Put 后 Done” 顺序使
// Put 撞 ErrElementAlreadyExist 直接返回，Done 永不执行，元素不在 list、永留 state，
// 永久丢失且无回调通知。
// 修复前：Retry 返回 ErrElementAlreadyExist，元素无法再次 Get；
// 修复后：Put 命中 processing 挂起返回 nil，Done 触发重入队，Retry 返回 nil 且元素可再次 Get。
func TestRetryQueue_Idempotent_RetryImmediate(t *testing.T) {
	config := NewRetryQueueConfig().WithPolicy(&immediateRetryPolicy{maxRetries: 3})
	config.WithValueIdempotent()
	q := NewRetryQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("pay-callback"))

	value, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "pay-callback", value)

	err = q.Retry(value, errors.New("process failed"))
	assert.NoError(t, err, "Retry must succeed in idempotent mode (D2 immediate path)")

	requeued, err := waitQueueGet(t, q, time.Second)
	assert.NoError(t, err, "retried element must be gettable again")
	assert.Equal(t, "pay-callback", requeued)
}

// TestRetryQueue_Idempotent_RetryDelayed 验证 D2 延迟路径在新语义下的完整链路：
// PutWithDelay 不碰内层 state，Done 时 state 不含该值故不误重入队；
// 到期后 puller 转运正常入队，元素可再次 Get。
func TestRetryQueue_Idempotent_RetryDelayed(t *testing.T) {
	config := NewRetryQueueConfig().WithPolicy(&fixedDelayRetryPolicy{delay: 30 * time.Millisecond, maxRetries: 3})
	config.WithValueIdempotent()
	q := NewRetryQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("pay-callback-delay"))

	value, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "pay-callback-delay", value)

	err = q.Retry(value, errors.New("process failed"))
	assert.NoError(t, err)

	// puller 以 300ms 心跳扫描，转运延迟至多约 330ms。
	requeued, err := waitQueueGet(t, q, 2*time.Second)
	assert.NoError(t, err, "delayed-retried element must be transported by puller and gettable again")
	assert.Equal(t, "pay-callback-delay", requeued)
}

// TestRetryQueue_Idempotent_RetryDelayed_DedupWithUserPut 验证到期前用户重新 Put 的去重语义：
// 若同值在延迟到期前已被用户 Put 回队列（占据 state），puller 到期转运延迟项时
// 撞 ErrElementAlreadyExist 而丢弃 —— 这是正确去重（元素已在队），不是丢元素：
// 队列中最终恰好一份，不产生重复副本。
func TestRetryQueue_Idempotent_RetryDelayed_DedupWithUserPut(t *testing.T) {
	config := NewRetryQueueConfig().WithPolicy(&fixedDelayRetryPolicy{delay: 100 * time.Millisecond, maxRetries: 3})
	config.WithValueIdempotent()
	q := NewRetryQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("dedup-case"))

	value, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "dedup-case", value)

	// 延迟路径：进入延迟堆，不碰内层 state。
	assert.NoError(t, q.Retry(value, errors.New("process failed")))

	// 到期前用户重新 Put 同值。
	assert.NoError(t, q.Put("dedup-case"))

	requeued, err := waitQueueGet(t, q, 2*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, "dedup-case", requeued)

	// 等待 puller 越过到期点（100ms 延迟 + 300ms 心跳上限），确认延迟项被去重丢弃、无第二份副本。
	time.Sleep(600 * time.Millisecond)
	_, err = q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty, "discarded duplicate must not produce a second copy")
}

type testRetryQueueCallback struct {
	mu sync.Mutex

	retries   []any
	exhausted []any
	forgets   []any
}

func (c *testRetryQueueCallback) OnPut(any) {}

func (c *testRetryQueueCallback) OnGet(any) {}

func (c *testRetryQueueCallback) OnDone(any) {}

func (c *testRetryQueueCallback) OnDelay(any, int64) {}

func (c *testRetryQueueCallback) OnPullError(any, error) {}

func (c *testRetryQueueCallback) OnRetry(value any, _ int, _ time.Duration, _ error) {
	c.mu.Lock()
	c.retries = append(c.retries, value)
	c.mu.Unlock()
}

func (c *testRetryQueueCallback) OnRetryExhausted(value any, _ int, _ error) {
	c.mu.Lock()
	c.exhausted = append(c.exhausted, value)
	c.mu.Unlock()
}

func (c *testRetryQueueCallback) OnForget(value any) {
	c.mu.Lock()
	c.forgets = append(c.forgets, value)
	c.mu.Unlock()
}

// TestRetryQueue_DeadLetterBridge_Exhausted 验证 G9：配置死信队列后，
// 重试耗尽时自动 PutDead，letter 各字段完整且正确。
func TestRetryQueue_DeadLetterBridge_Exhausted(t *testing.T) {
	dlq := NewDeadLetterQueue(nil)
	defer dlq.Shutdown()

	config := NewRetryQueueConfig().
		WithPolicy(NewExponentialRetryPolicy(0, 0, 1)).
		WithDeadLetterQueue(dlq, "retry-main")
	q := NewRetryQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("task"))

	value, err := q.Get()
	assert.NoError(t, err)
	assert.NoError(t, q.Retry(value, errors.New("first failed")))

	var ok bool
	assert.Eventually(t, func() bool {
		value, err = q.Get()
		ok = err == nil
		return ok
	}, 2*time.Second, 20*time.Millisecond)
	assert.True(t, ok)

	err = q.Retry(value, errors.New("second failed"))
	assert.ErrorIs(t, err, ErrRetryExhausted)

	letter, err := dlq.GetDead()
	assert.NoError(t, err, "exhausted value must land in the dead letter queue")
	assert.Equal(t, "task", letter.Payload)
	assert.Equal(t, "retry-main", letter.SourceQueue)
	assert.Equal(t, 2, letter.Attempts)
	assert.Equal(t, "second failed", letter.LastError)
	assert.NotEmpty(t, letter.ID)
	assert.False(t, letter.FailedAt.IsZero())
	assert.Nil(t, letter.Meta)
}

// TestRetryQueue_DeadLetterBridge_IDMonotonic 验证同值多次耗尽产生的死信
// ID 为 base36 编码且严格单调递增（重试队列侧原子序号）。
func TestRetryQueue_DeadLetterBridge_IDMonotonic(t *testing.T) {
	dlq := NewDeadLetterQueue(nil)
	defer dlq.Shutdown()

	config := NewRetryQueueConfig().
		WithPolicy(&fixedDelayRetryPolicy{delay: 5 * time.Millisecond, maxRetries: 1}).
		WithDeadLetterQueue(dlq, "retry-main")
	q := NewRetryQueue(config)
	defer q.Shutdown()

	const rounds = 3
	prev := uint64(0)
	for r := 0; r < rounds; r++ {
		assert.NoError(t, q.Put("task"))

		value, err := q.Get()
		assert.NoError(t, err)
		assert.NoError(t, q.Retry(value, errors.New("transient")))

		var ok bool
		assert.Eventually(t, func() bool {
			value, err = q.Get()
			ok = err == nil
			return ok
		}, 2*time.Second, 5*time.Millisecond)
		assert.True(t, ok)

		assert.ErrorIs(t, q.Retry(value, errors.New("fatal")), ErrRetryExhausted)

		letter, err := dlq.GetDead()
		assert.NoError(t, err)

		id, err := strconv.ParseUint(letter.ID, 36, 64)
		assert.NoError(t, err, "dead letter ID must be base36 encoded")
		assert.Greater(t, id, prev, "IDs must be strictly increasing across exhaustions")
		prev = id
	}
}

// dlqInspectingCallback 在 OnRetryExhausted 时刻记录死信队列长度，
// 用于断言 PutDead 相对回调的发生顺序。
type dlqInspectingCallback struct {
	dlq          Queue
	lenAtExhaust int
	exhaustCalls int
}

func (c *dlqInspectingCallback) OnPut(any)                              {}
func (c *dlqInspectingCallback) OnGet(any)                              {}
func (c *dlqInspectingCallback) OnDone(any)                             {}
func (c *dlqInspectingCallback) OnDelay(any, int64)                     {}
func (c *dlqInspectingCallback) OnPullError(any, error)                 {}
func (c *dlqInspectingCallback) OnRetry(any, int, time.Duration, error) {}
func (c *dlqInspectingCallback) OnForget(any)                           {}

func (c *dlqInspectingCallback) OnRetryExhausted(any, int, error) {
	c.exhaustCalls++
	c.lenAtExhaust = c.dlq.Len()
}

// TestRetryQueue_DeadLetterBridge_AfterExhaustedCallback 验证 G9 时序契约：
// PutDead 发生在 OnRetryExhausted 回调之后——回调触发时刻死信队列尚为空，
// Retry 返回后 letter 才入队。
func TestRetryQueue_DeadLetterBridge_AfterExhaustedCallback(t *testing.T) {
	dlq := NewDeadLetterQueue(nil)
	defer dlq.Shutdown()

	callback := &dlqInspectingCallback{dlq: dlq}
	config := NewRetryQueueConfig().
		WithPolicy(&fixedDelayRetryPolicy{delay: 5 * time.Millisecond, maxRetries: 1}).
		WithCallback(callback).
		WithDeadLetterQueue(dlq, "retry-main")
	q := NewRetryQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("task"))

	value, err := q.Get()
	assert.NoError(t, err)
	assert.NoError(t, q.Retry(value, errors.New("transient")))

	var ok bool
	assert.Eventually(t, func() bool {
		value, err = q.Get()
		ok = err == nil
		return ok
	}, 2*time.Second, 5*time.Millisecond)
	assert.True(t, ok)

	err = q.Retry(value, errors.New("fatal"))
	assert.ErrorIs(t, err, ErrRetryExhausted)

	assert.Equal(t, 1, callback.exhaustCalls)
	assert.Equal(t, 0, callback.lenAtExhaust, "letter must not be in DLQ before OnRetryExhausted returns")
	assert.Equal(t, 1, dlq.Len(), "letter must be in DLQ after Retry returns")
}

// TestRetryQueue_DeadLetterBridge_PutDeadFailure 验证 PutDead 失败语义：
// 死信队列已关停（PutDead 返回 ErrQueueIsClosed）时，桥不崩溃、不改变
// Retry 返回值（仍为 ErrRetryExhausted），耗尽事件仍由 OnRetryExhausted 报告。
func TestRetryQueue_DeadLetterBridge_PutDeadFailure(t *testing.T) {
	dlq := NewDeadLetterQueue(nil)
	dlq.Shutdown() // 使后续 PutDead 失败

	callback := &testRetryQueueCallback{}
	config := NewRetryQueueConfig().
		WithPolicy(&fixedDelayRetryPolicy{delay: 5 * time.Millisecond, maxRetries: 1}).
		WithCallback(callback).
		WithDeadLetterQueue(dlq, "retry-main")
	q := NewRetryQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("task"))

	value, err := q.Get()
	assert.NoError(t, err)
	assert.NoError(t, q.Retry(value, errors.New("transient")))

	var ok bool
	assert.Eventually(t, func() bool {
		value, err = q.Get()
		ok = err == nil
		return ok
	}, 2*time.Second, 5*time.Millisecond)
	assert.True(t, ok)

	err = q.Retry(value, errors.New("fatal"))
	assert.ErrorIs(t, err, ErrRetryExhausted, "Retry must keep returning ErrRetryExhausted when PutDead fails")

	callback.mu.Lock()
	assert.Equal(t, []any{"task"}, callback.exhausted, "OnRetryExhausted must fire even when PutDead fails")
	callback.mu.Unlock()
	assert.Equal(t, 0, dlq.Len(), "no letter can land in a closed DLQ")
}

// TestRetryQueue_RetryExhausted_WithoutDLQ 是 G9 的未配置回归：未通过
// WithDeadLetterQueue 接线时，耗尽路径语义保持不变（返回 ErrRetryExhausted、
// attempts 归零），队列在耗尽后继续可用。
func TestRetryQueue_RetryExhausted_WithoutDLQ(t *testing.T) {
	config := NewRetryQueueConfig().WithPolicy(NewExponentialRetryPolicy(0, 0, 1))
	q := NewRetryQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("task"))

	value, err := q.Get()
	assert.NoError(t, err)
	assert.NoError(t, q.Retry(value, errors.New("first failed")))

	var ok bool
	assert.Eventually(t, func() bool {
		value, err = q.Get()
		ok = err == nil
		return ok
	}, 2*time.Second, 20*time.Millisecond)
	assert.True(t, ok)

	err = q.Retry(value, errors.New("second failed"))
	assert.ErrorIs(t, err, ErrRetryExhausted)
	assert.Equal(t, 0, q.NumRequeues("task"))

	// 耗尽重置计数后，同值可重新走完整重试周期。
	assert.NoError(t, q.Put("task"))
	value, err = q.Get()
	assert.NoError(t, err)
	assert.NoError(t, q.Retry(value, errors.New("new round")))
	assert.Equal(t, 1, q.NumRequeues("task"))
}

// zeroDelayPolicy 是 D5 复现专用零延迟策略：attempt 不超过 2 时返回零延迟
// 并要求继续重试，超过后拒绝重试（走耗尽分支）。零延迟使 Retry 必然命中
// immediate 路径（delay < 1ms → 直接 Put）。
type zeroDelayPolicy struct{}

func (p *zeroDelayPolicy) NextDelay(_ any, attempt int, _ error) (time.Duration, bool) {
	if attempt > 2 {
		return 0, false
	}
	return 0, true
}

// TestRetryQueue_ZeroDelayPolicy_AttemptsAccumulate 复现 D5：零延迟策略下
// immediate 路径若在 increment 后立即 delete，attempts 永不累积——
// NumRequeues 恒为 0 且 Retry 永不耗尽。修复后 attempts 逐次累积 1/2，
// 第 3 次 Retry 因策略拒绝而耗尽并重置计数。
func TestRetryQueue_ZeroDelayPolicy_AttemptsAccumulate(t *testing.T) {
	config := NewRetryQueueConfig().WithPolicy(&zeroDelayPolicy{})
	q := NewRetryQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("task"))

	value, err := q.Get()
	assert.NoError(t, err)

	assert.NoError(t, q.Retry(value, errors.New("fail-1")))
	assert.Equal(t, 1, q.NumRequeues("task"), "attempts must accumulate on the immediate path")

	value, err = q.Get()
	assert.NoError(t, err)

	assert.NoError(t, q.Retry(value, errors.New("fail-2")))
	assert.Equal(t, 2, q.NumRequeues("task"))

	value, err = q.Get()
	assert.NoError(t, err)

	err = q.Retry(value, errors.New("fail-3"))
	assert.ErrorIs(t, err, ErrRetryExhausted)
	assert.Equal(t, 0, q.NumRequeues("task"), "exhaustion must reset attempts")
}

// TestRetryQueue_ZeroDelayPolicy_ExhaustsInsteadOfRetryingForever 以有界循环
// 固化“零延迟策略必然耗尽”：策略允许至多 2 次重试，3 次 Retry 调用内必须耗尽。
// 修复前 attempts 永不累积、Retry 无限重入队，该循环会跑满上限而失败；
// 上限将“无限重试”转化为确定性可观察的断言失败，测试自身不会挂死。
func TestRetryQueue_ZeroDelayPolicy_ExhaustsInsteadOfRetryingForever(t *testing.T) {
	config := NewRetryQueueConfig().WithPolicy(&zeroDelayPolicy{})
	q := NewRetryQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("task"))

	const maxCalls = 10
	exhausted := false
	for i := 0; i < maxCalls; i++ {
		value, err := q.Get()
		assert.NoError(t, err)

		err = q.Retry(value, errors.New("keep failing"))
		if errors.Is(err, ErrRetryExhausted) {
			exhausted = true
			assert.Equal(t, 3, i+1, "must exhaust exactly at the third Retry call")
			break
		}
		assert.NoError(t, err)
	}
	assert.True(t, exhausted, "zero-delay policy must exhaust; infinite retry detected")
}

// TestRetryQueue_ZeroDelayPolicy_DeadLetterBridge 验证 D5 修复后 G9 死信桥对
// 零延迟策略同样生效：耗尽时 letter 自动进入死信队列，Attempts 为累积值 3。
// 修复前 attempts 不累积导致耗尽永不发生，死信桥对零延迟策略永不触发。
func TestRetryQueue_ZeroDelayPolicy_DeadLetterBridge(t *testing.T) {
	dlq := NewDeadLetterQueue(nil)
	defer dlq.Shutdown()

	config := NewRetryQueueConfig().
		WithPolicy(&zeroDelayPolicy{}).
		WithDeadLetterQueue(dlq, "retry-main")
	q := NewRetryQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("task"))

	for i := 0; i < 2; i++ {
		value, err := q.Get()
		assert.NoError(t, err)
		assert.NoError(t, q.Retry(value, errors.New("transient")))
	}

	value, err := q.Get()
	assert.NoError(t, err)
	assert.ErrorIs(t, q.Retry(value, errors.New("fatal")), ErrRetryExhausted)

	letter, err := dlq.GetDead()
	if !assert.NoError(t, err, "zero-delay exhaustion must reach the dead letter bridge") {
		return
	}
	assert.Equal(t, "task", letter.Payload)
	assert.Equal(t, "retry-main", letter.SourceQueue)
	assert.Equal(t, 3, letter.Attempts)
	assert.Equal(t, "fatal", letter.LastError)
}

func TestRetryQueue_Callback(t *testing.T) {
	callback := &testRetryQueueCallback{}
	config := NewRetryQueueConfig().
		WithCallback(callback).
		WithPolicy(NewExponentialRetryPolicy(0, 0, 1))
	q := NewRetryQueue(config)
	defer q.Shutdown()

	err := q.Put("task")
	assert.NoError(t, err)

	value, err := q.Get()
	assert.NoError(t, err)

	err = q.Retry(value, errors.New("failed"))
	assert.NoError(t, err)

	var ok bool
	assert.Eventually(t, func() bool {
		value, err = q.Get()
		ok = err == nil
		return ok
	}, 2*time.Second, 20*time.Millisecond)
	assert.True(t, ok)
	err = q.Retry(value, errors.New("failed-again"))
	assert.ErrorIs(t, err, ErrRetryExhausted)

	q.Forget("task")

	callback.mu.Lock()
	defer callback.mu.Unlock()
	assert.Equal(t, []any{"task"}, callback.retries)
	assert.Equal(t, []any{"task"}, callback.exhausted)
	assert.Equal(t, []any{"task"}, callback.forgets)
}

// TestRetryQueue_RequeueCounts 验证 G5：RequeueCounts 暴露 attempts 的快照副本，
// 默认 key 语义 = 默认 RetryKeyFunc 产出（fmt.Sprintf("%T:%#v", value, value)）。
func TestRetryQueue_RequeueCounts(t *testing.T) {
	config := NewRetryQueueConfig().WithPolicy(NewExponentialRetryPolicy(0, 0, 3))
	q := NewRetryQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("task-a"))
	assert.NoError(t, q.Put("task-b"))

	value, err := q.Get()
	assert.NoError(t, err)
	assert.NoError(t, q.Retry(value, errors.New("fail")))

	value, err = q.Get()
	assert.NoError(t, err)
	assert.NoError(t, q.Retry(value, errors.New("fail")))
	assert.NoError(t, q.Retry(value, errors.New("fail-again")))

	counts := q.RequeueCounts()
	keyA := fmt.Sprintf("%T:%#v", "task-a", "task-a")
	keyB := fmt.Sprintf("%T:%#v", "task-b", "task-b")
	assert.Len(t, counts, 2)
	assert.Equal(t, 1, counts[keyA], "key must be the default RetryKeyFunc output, not the raw value")
	assert.Equal(t, 2, counts[keyB])
}

// TestRetryQueue_RequeueCounts_CustomKeyFunc 验证自定义 key 函数时，
// 返回的 map key 跟随 RetryKeyFunc 的输出。
func TestRetryQueue_RequeueCounts_CustomKeyFunc(t *testing.T) {
	config := NewRetryQueueConfig().
		WithPolicy(NewExponentialRetryPolicy(0, 0, 3)).
		WithKeyFunc(func(value any) string { return "custom-" + value.(string) })
	q := NewRetryQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("task"))

	value, err := q.Get()
	assert.NoError(t, err)
	assert.NoError(t, q.Retry(value, errors.New("fail")))

	counts := q.RequeueCounts()
	assert.Len(t, counts, 1)
	assert.Equal(t, 1, counts["custom-task"])
}

// TestRetryQueue_RequeueCounts_EmptyAndCopyMutation 验证空队列返回非 nil 可写副本，
// 且篡改副本不影响内部 attempts。
func TestRetryQueue_RequeueCounts_EmptyAndCopyMutation(t *testing.T) {
	config := NewRetryQueueConfig().WithPolicy(NewExponentialRetryPolicy(0, 0, 3))
	q := NewRetryQueue(config)
	defer q.Shutdown()

	empty := q.RequeueCounts()
	assert.NotNil(t, empty)
	empty["probe"] = 42 // 副本可写，不得 panic

	assert.NoError(t, q.Put("task"))

	value, err := q.Get()
	assert.NoError(t, err)
	assert.NoError(t, q.Retry(value, errors.New("fail")))

	counts := q.RequeueCounts()
	for key := range counts {
		counts[key] = 999
	}
	counts["injected"] = 7

	fresh := q.RequeueCounts()
	key := fmt.Sprintf("%T:%#v", "task", "task")
	assert.Len(t, fresh, 1)
	assert.Equal(t, 1, fresh[key], "mutating the snapshot must not affect internal attempts")
}

// TestRetryQueue_RequeueCounts_SnapshotDetached 验证快照脱锁：取快照后并发
// Retry/Forget 不被阻塞、不死锁。
func TestRetryQueue_RequeueCounts_SnapshotDetached(t *testing.T) {
	config := NewRetryQueueConfig().WithPolicy(NewExponentialRetryPolicy(0, 0, -1))
	q := NewRetryQueue(config)
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
				if value, err := q.Get(); err == nil {
					_ = q.Retry(value, errors.New("transient"))
					q.Forget(value)
				}
			}
		}
	}()

	snapDone := make(chan struct{})
	go func() {
		defer close(snapDone)
		for i := 0; i < 200; i++ {
			_ = q.RequeueCounts()
		}
	}()

	select {
	case <-snapDone:
	case <-time.After(3 * time.Second):
		close(stop)
		<-trafficDone
		t.Fatal("RequeueCounts blocked while concurrent Retry/Forget traffic is running")
	}

	close(stop)
	<-trafficDone
}
