package workqueue

import (
	"sync/atomic"
	"time"
)

// MetricsRecorder 是零第三方依赖的内置指标记录器：以超集实现满足队列族全部
// 回调接口，可直接传给任意队列的 WithCallback（Queue/DelayingQueue/TimerQueue/
// PriorityQueue/RateLimitingQueue/RetryQueue/DeadLetterQueue/LeasedQueue 的配置
// 入口均接受）；热路径仅原子累加、零分配。
//
// 设计要点：recorder 不持有队列引用，Snapshot 仅返回累计计数——队列深度
// （depth）等状态类指标为拉取式，由使用者结合 queue.Len() 自行组合。
// 这使得单个 recorder 既可服务单个队列，也可同时挂到多个队列上做聚合统计。
// 累计语义以回调事件为准：例如延迟项到期搬运触发的内层 Put 会计入 Adds，
// 被拒绝的入队（重复值/已关停）不触发回调故不计数。
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

// 编译期断言：MetricsRecorder 满足队列族全部回调接口，
// 可直接传给任意队列的 WithCallback。
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

// MetricsSnapshot 是 MetricsRecorder 累计计数的只读快照。
type MetricsSnapshot struct {
	// Adds 累计 OnPut（成功入队）事件数。
	Adds int64

	// Gets 累计 OnGet（消费取走）事件数。
	Gets int64

	// Dones 累计 OnDone（处理完成）事件数。
	Dones int64

	// Retries 累计 OnRetry（重入队成功）事件数。
	Retries int64

	// RetryExhausteds 累计 OnRetryExhausted（重试耗尽）事件数。
	RetryExhausteds int64

	// PullErrors 累计 OnPullError（延迟搬运失败）事件数。
	PullErrors int64

	// ScheduleErrors 累计 OnScheduleError（定时调度失败）事件数。
	ScheduleErrors int64

	// Unfinished 为已 Get 未 Done 的在途数量（Gets - Dones）。
	Unfinished int64
}

// Snapshot 返回当前累计计数快照，仅原子读、零分配。
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

// OnDone 累计一次完成事件。
func (r *MetricsRecorder) OnDone(any) { r.dones.Add(1) }

// OnDelay 为延迟入队事件，搬运入队时另行计入 Adds，此处空实现。
func (r *MetricsRecorder) OnDelay(any, int64) {}

// OnPullError 累计一次延迟搬运失败事件。
func (r *MetricsRecorder) OnPullError(any, error) { r.pullErrors.Add(1) }

// OnSchedule 为定时调度事件，投递入队时另行计入 Adds，此处空实现。
func (r *MetricsRecorder) OnSchedule(any, int64) {}

// OnScheduleError 累计一次定时调度失败事件。
func (r *MetricsRecorder) OnScheduleError(any, error) { r.scheduleErrors.Add(1) }

// OnPriority 为优先级入队事件，此处空实现。
func (r *MetricsRecorder) OnPriority(any, int64) {}

// OnLimited 为限流入队事件，此处空实现。
func (r *MetricsRecorder) OnLimited(any) {}

// OnRetry 累计一次重试事件。
func (r *MetricsRecorder) OnRetry(any, int, time.Duration, error) { r.retries.Add(1) }

// OnRetryExhausted 累计一次重试耗尽事件。
func (r *MetricsRecorder) OnRetryExhausted(any, int, error) { r.retryExhausteds.Add(1) }

// OnForget 为 forget 事件，此处空实现。
func (r *MetricsRecorder) OnForget(any) {}

// OnDead 为死信入队事件，此处空实现。
func (r *MetricsRecorder) OnDead(*DeadLetter) {}

// OnAckDead 为死信确认事件，此处空实现。
func (r *MetricsRecorder) OnAckDead(*DeadLetter) {}

// OnRequeueDead 为死信重入队事件，此处空实现。
func (r *MetricsRecorder) OnRequeueDead(*DeadLetter, Queue) {}

// OnNack 为租约 nack 事件，此处空实现。
func (r *MetricsRecorder) OnNack(any, error) {}
