package workqueue

import (
	"sync/atomic"
	"time"
)

// MetricsRecorder 零依赖内置指标记录器：满足队列族全部回调接口，
// 可直接传给任意队列的 WithCallback；热路径仅原子累加、零分配。
// recorder 不持有队列引用，单个 recorder 可服务单个或多个队列做聚合统计。
// 累计语义以回调事件为准：被拒绝的入队不触发回调故不计数。
type MetricsRecorder struct {
	adds            atomic.Int64
	gets            atomic.Int64
	dones           atomic.Int64
	retries         atomic.Int64
	retryExhausteds atomic.Int64
	pullErrors      atomic.Int64
	scheduleErrors  atomic.Int64
}

// NewMetricsRecorder 创建指标记录器。
func NewMetricsRecorder() *MetricsRecorder {
	return &MetricsRecorder{}
}

// 编译期断言：MetricsRecorder 满足队列族全部回调接口。
var (
	_ QueueCallback             = (*MetricsRecorder)(nil)
	_ DelayingQueueCallback     = (*MetricsRecorder)(nil)
	_ TimerQueueCallback        = (*MetricsRecorder)(nil)
	_ PriorityQueueCallback     = (*MetricsRecorder)(nil)
	_ RateLimitingQueueCallback = (*MetricsRecorder)(nil)
	_ RetryQueueCallback        = (*MetricsRecorder)(nil)
	_ DeadLetterQueueCallback   = (*MetricsRecorder)(nil)
	_ LeasedQueueCallback       = (*MetricsRecorder)(nil)
)

// MetricsSnapshot 累计计数的只读快照。
type MetricsSnapshot struct {
	// Adds 累计 OnPut 事件数。
	Adds int64

	// Gets 累计 OnGet 事件数。
	Gets int64

	// Dones 累计 OnDone 事件数。
	// 幂等模式下经簿记校验；非幂等模式下为 best-effort，重复 Done 重复计数。
	Dones int64

	// Retries 累计 OnRetry 事件数。
	Retries int64

	// RetryExhausteds 累计 OnRetryExhausted 事件数。
	RetryExhausteds int64

	// PullErrors 累计 OnPullError 事件数。
	PullErrors int64

	// ScheduleErrors 累计 OnScheduleError 事件数。
	ScheduleErrors int64

	// Unfinished 已 Get 未 Done 的在途数量（Gets - Dones）。
	// 幂等模式下等于真实在途；非幂等模式下为 best-effort，需结合消费方纪律解读。
	Unfinished int64
}

// Snapshot 返回累计计数快照，仅原子读、零分配。
func (r *MetricsRecorder) Snapshot() MetricsSnapshot {
	gets := r.gets.Load()
	dones := r.dones.Load()

	return MetricsSnapshot{
		Adds:            r.adds.Load(),
		Gets:            gets,
		Dones:           dones,
		Retries:         r.retries.Load(),
		RetryExhausteds: r.retryExhausteds.Load(),
		PullErrors:      r.pullErrors.Load(),
		ScheduleErrors:  r.scheduleErrors.Load(),
		Unfinished:      gets - dones,
	}
}

// OnPut 累计一次入队事件。
func (r *MetricsRecorder) OnPut(any) { r.adds.Add(1) }

// OnGet 累计一次消费事件。
func (r *MetricsRecorder) OnGet(any) { r.gets.Add(1) }

// OnDone 累计完成事件。幂等模式下经簿记校验；非幂等模式下调用即触发。
func (r *MetricsRecorder) OnDone(any) { r.dones.Add(1) }

// OnDelay 延迟入队事件，搬运入队时计入 Adds，此处空实现。
func (r *MetricsRecorder) OnDelay(any, int64) {}

// OnPullError 累计一次延迟搬运失败事件。
func (r *MetricsRecorder) OnPullError(any, error) { r.pullErrors.Add(1) }

// OnSchedule 定时调度事件，投递入队时计入 Adds，此处空实现。
func (r *MetricsRecorder) OnSchedule(any, int64) {}

// OnScheduleError 累计一次定时调度失败事件。
func (r *MetricsRecorder) OnScheduleError(any, error) { r.scheduleErrors.Add(1) }

// OnPriority 优先级入队事件，空实现。
func (r *MetricsRecorder) OnPriority(any, int64) {}

// OnLimited 限流入队事件，空实现。
func (r *MetricsRecorder) OnLimited(any) {}

// OnRetry 累计一次重试事件。
func (r *MetricsRecorder) OnRetry(any, int, time.Duration, error) { r.retries.Add(1) }

// OnRetryExhausted 累计一次重试耗尽事件。
func (r *MetricsRecorder) OnRetryExhausted(any, int, error) { r.retryExhausteds.Add(1) }

// OnForget forget 事件，空实现。
func (r *MetricsRecorder) OnForget(any) {}

// OnDead 死信入队事件，空实现。
func (r *MetricsRecorder) OnDead(*DeadLetter) {}

// OnAckDead 死信确认事件，空实现。
func (r *MetricsRecorder) OnAckDead(*DeadLetter) {}

// OnRequeueDead 死信重入队事件，空实现。
func (r *MetricsRecorder) OnRequeueDead(*DeadLetter, Queue) {}

// OnNack 租约 nack 事件，空实现。
func (r *MetricsRecorder) OnNack(any, error) {}
