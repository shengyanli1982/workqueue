package workqueue

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMetricsRecorder_Counting 直接驱动各回调方法，验证计数语义映射：
// OnPut→Adds、OnGet→Gets、OnDone→Dones、OnRetry→Retries、
// OnRetryExhausted→RetryExhausteds、OnPullError→PullErrors、
// OnScheduleError→ScheduleErrors、Unfinished=Gets-Dones。
func TestMetricsRecorder_Counting(t *testing.T) {
	rec := NewMetricsRecorder()

	rec.OnPut("v1")
	rec.OnPut("v2")
	rec.OnGet("v1")
	rec.OnDone("v1")
	rec.OnRetry("v1", 1, time.Second, errors.New("transient"))
	rec.OnRetryExhausted("v1", 2, errors.New("fatal"))
	rec.OnPullError("v1", errors.New("pull failed"))
	rec.OnScheduleError("v1", errors.New("schedule failed"))
	// 空实现方法不得影响计数。
	rec.OnDelay("v1", 100)
	rec.OnPriority("v1", 7)
	rec.OnLimited("v1")
	rec.OnForget("v1")
	rec.OnDead(&DeadLetter{})
	rec.OnAckDead(&DeadLetter{})
	rec.OnRequeueDead(&DeadLetter{}, nil)
	rec.OnNack("v1", errors.New("nack"))
	rec.OnSchedule("v1", time.Now().UnixMilli())

	s := rec.Snapshot()
	assert.EqualValues(t, 2, s.Adds)
	assert.EqualValues(t, 1, s.Gets)
	assert.EqualValues(t, 1, s.Dones)
	assert.EqualValues(t, 1, s.Retries)
	assert.EqualValues(t, 1, s.RetryExhausteds)
	assert.EqualValues(t, 1, s.PullErrors)
	assert.EqualValues(t, 1, s.ScheduleErrors)
	assert.EqualValues(t, 0, s.Unfinished, "Unfinished must equal Gets-Dones")

	// 再来一次 Get 且不再 Done：Unfinished 应增长为 1。
	rec.OnGet("v2")
	s = rec.Snapshot()
	assert.EqualValues(t, 2, s.Gets)
	assert.EqualValues(t, 1, s.Unfinished)
}

// TestMetricsRecorder_HotPath_ZeroAlloc 验证热路径零分配：
// 各计数方法仅原子累加，Snapshot 仅原子读，均不得产生堆分配。
func TestMetricsRecorder_HotPath_ZeroAlloc(t *testing.T) {
	rec := NewMetricsRecorder()

	var value any = "value"
	reason := errors.New("boom")

	assert.Zero(t, testing.AllocsPerRun(1000, func() { rec.OnPut(value) }), "OnPut must be zero-alloc")
	assert.Zero(t, testing.AllocsPerRun(1000, func() { rec.OnGet(value) }), "OnGet must be zero-alloc")
	assert.Zero(t, testing.AllocsPerRun(1000, func() { rec.OnDone(value) }), "OnDone must be zero-alloc")
	assert.Zero(t, testing.AllocsPerRun(1000, func() { rec.OnRetry(value, 3, time.Millisecond, reason) }), "OnRetry must be zero-alloc")
	assert.Zero(t, testing.AllocsPerRun(1000, func() { rec.OnRetryExhausted(value, 3, reason) }), "OnRetryExhausted must be zero-alloc")
	assert.Zero(t, testing.AllocsPerRun(1000, func() { rec.OnPullError(value, reason) }), "OnPullError must be zero-alloc")
	assert.Zero(t, testing.AllocsPerRun(1000, func() { rec.OnScheduleError(value, reason) }), "OnScheduleError must be zero-alloc")
	assert.Zero(t, testing.AllocsPerRun(1000, func() { rec.Snapshot() }), "Snapshot must be zero-alloc")
}

// TestMetricsRecorder_QueueIntegration 验证挂到基础队列（幂等模式）后，
// 一轮 Put/Get/Done 操作的计数正确；被拒绝的重复 Put 不计入 Adds。
func TestMetricsRecorder_QueueIntegration(t *testing.T) {
	rec := NewMetricsRecorder()
	q := NewQueue(NewQueueConfig().WithCallback(rec).WithValueIdempotent())
	defer q.Shutdown()

	assert.NoError(t, q.Put("task-a"))
	assert.ErrorIs(t, q.Put("task-a"), ErrElementAlreadyExist, "duplicate Put must be rejected")
	assert.NoError(t, q.Put("task-b"))

	value, err := q.Get()
	assert.NoError(t, err)
	assert.Equal(t, "task-a", value)
	q.Done(value)

	s := rec.Snapshot()
	assert.EqualValues(t, 2, s.Adds, "rejected duplicate Put must not be counted")
	assert.EqualValues(t, 1, s.Gets)
	assert.EqualValues(t, 1, s.Dones)
	assert.EqualValues(t, 0, s.Unfinished)
	assert.EqualValues(t, 0, s.Retries)
	assert.EqualValues(t, 0, s.RetryExhausteds)
	assert.EqualValues(t, 0, s.PullErrors)
	assert.EqualValues(t, 0, s.ScheduleErrors)
}

// TestMetricsRecorder_RetryQueueIntegration 验证挂到重试队列后，
// 一轮“入队→消费→重试→再消费→耗尽”的计数正确：
// 延迟到期搬运产生的内层 Put 同样计入 Adds（Adds 语义 = OnPut 事件数）。
func TestMetricsRecorder_RetryQueueIntegration(t *testing.T) {
	rec := NewMetricsRecorder()
	config := NewRetryQueueConfig().
		WithCallback(rec).
		WithPolicy(NewExponentialRetryPolicy(0, 0, 1)) // baseDelay 钳制为 100ms，走延迟路径
	q := NewRetryQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("task")) // Adds=1

	value, err := q.Get() // Gets=1
	assert.NoError(t, err)

	assert.NoError(t, q.Retry(value, errors.New("transient"))) // Retries=1；100ms 后搬运 → Adds=2

	var got bool
	assert.Eventually(t, func() bool {
		value, err = q.Get() // Gets=2
		got = err == nil
		return got
	}, 2*time.Second, 5*time.Millisecond)
	assert.True(t, got)

	err = q.Retry(value, errors.New("fatal")) // RetryExhausteds=1
	assert.ErrorIs(t, err, ErrRetryExhausted)

	s := rec.Snapshot()
	assert.EqualValues(t, 2, s.Adds, "initial Put + delayed transport Put")
	assert.EqualValues(t, 2, s.Gets)
	assert.EqualValues(t, 0, s.Dones)
	assert.EqualValues(t, 1, s.Retries)
	assert.EqualValues(t, 1, s.RetryExhausteds)
	assert.EqualValues(t, 0, s.PullErrors)
	assert.EqualValues(t, 0, s.ScheduleErrors)
	assert.EqualValues(t, 2, s.Unfinished, "both Gets never Done")
}
