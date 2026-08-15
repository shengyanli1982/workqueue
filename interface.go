package workqueue

import (
	"context"
	"time"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// DrainableQueue 描述支持优雅关停的队列。仿照 io.Closer 惯例以可选接口提供：
// 不并入 Queue 接口（保护外部 Queue 实现者的兼容性），调用方经类型断言使用。
type DrainableQueue = interface {
	ShutdownWithDrain(ctx context.Context) error
}

// BlockingGetQueue 描述支持阻塞式消费的队列。仿照 io.Closer 惯例以可选接口提供：
// 不并入 Queue 接口（保护外部 Queue 实现者的兼容性），调用方经类型断言使用。
// GetWithContext 阻塞直到“有值 / ctx 完成 / 队列关闭”三者之一，永不返回
// ErrQueueIsEmpty；未使用阻塞消费的队列默认路径零开销（惰性广播机制）。
type BlockingGetQueue = interface {
	GetWithContext(ctx context.Context) (value any, err error)
}

// InFlightQueue 描述能暴露处理中元素（已 Get 未 Done）的队列。仿照 io.Closer
// 惯例以可选接口提供：不并入 Queue 接口（保护外部 Queue 实现者的兼容性），
// 调用方经类型断言使用。仅幂等模式（WithValueIdempotent）维护 per-value 处理中
// 集合；非幂等实现的 InFlight 返回 nil。
type InFlightQueue = interface {
	InFlight() []any
}

// Queue 定义基础队列语义：消费端 Get 成功后应调用 Done。
type Queue = interface {
	Put(value any) error

	Get() (value any, err error)

	Done(value any)

	Len() int

	Values() []any

	Range(fn func(value any) bool)

	Shutdown()

	IsClosed() bool
}

// DelayingQueue 在普通队列基础上支持按延迟时间入队。
type DelayingQueue = interface {
	Queue

	PutWithDelay(value any, delay int64) error

	// CancelDelay 取消堆中尚未搬运的单个延迟项，命中并移除返回 true；
	// 未命中（未入堆、已搬运、已取消或已关停清空）返回 false。
	CancelDelay(value any) bool

	HeapRange(fn func(value any, delay int64) bool)
}

// PriorityQueue 在普通队列基础上支持按优先级入队。
type PriorityQueue = interface {
	Queue

	PutWithPriority(value any, priority int64) error

	HeapRange(fn func(value any, priority int64) bool)
}

// RateLimitingQueue 在 DelayingQueue 基础上提供限流入队能力。
type RateLimitingQueue = interface {
	DelayingQueue

	PutWithLimited(value any) error
}

// RetryQueue 在 DelayingQueue 基础上提供失败重试能力。
type RetryQueue = interface {
	DelayingQueue

	Retry(value any, reason error) error

	Forget(value any)

	NumRequeues(value any) int

	// RequeueCounts 返回当前重试计数表的快照副本，
	// map key 为 RetryKeyFunc 产出的重试 key，详见实现文档。
	RequeueCounts() map[string]int
}

// DeadLetter 保存失败终态任务及其诊断元数据。
type DeadLetter struct {
	ID          string
	Payload     any
	SourceQueue string
	Attempts    int
	LastError   string
	FailedAt    time.Time
	Meta        map[string]string
}

// DeadLetterQueue 在 Queue 基础上提供死信治理能力。
type DeadLetterQueue = interface {
	Queue

	PutDead(letter *DeadLetter) error

	GetDead() (*DeadLetter, error)

	AckDead(letter *DeadLetter) error

	RequeueDead(letter *DeadLetter, target Queue) error

	RangeDead(fn func(letter *DeadLetter) bool)
}

// LeasedQueue 在基础队列上提供租约消费语义。
type LeasedQueue = interface {
	Queue

	GetWithLease(timeout time.Duration) (value any, leaseID string, err error)

	Ack(leaseID string) error

	Nack(leaseID string, reason error) error

	ExtendLease(leaseID string, timeout time.Duration) error

	// LeaseInfos 返回当前全部在租租约的只读快照副本，
	// 锁内拷贝、锁外返回；副本可安全修改，不影响内部状态。
	LeaseInfos() []LeaseInfo
}

// BoundedBlockingQueue 在基础队列上提供容量限制和阻塞读写。
type BoundedBlockingQueue = interface {
	Queue

	Cap() int

	PutWithContext(ctx context.Context, value any) error

	GetWithContext(ctx context.Context) (value any, err error)
}

// TimerQueue 在基础队列上提供按绝对时间调度入队。
type TimerQueue = interface {
	Queue

	PutAt(value any, at time.Time) error

	PutAfter(value any, after time.Duration) error

	Cancel(value any) bool

	HeapRange(fn func(value any, at int64) bool)
}

// QueueCallback 定义基础队列生命周期回调。
type QueueCallback = interface {
	OnPut(value any)

	OnGet(value any)

	OnDone(value any)
}

// DelayingQueueCallback 扩展延迟队列回调。
type DelayingQueueCallback = interface {
	QueueCallback

	OnDelay(value any, delay int64)

	OnPullError(value any, reason error)
}

// TimerQueueCallback 扩展定时队列回调。
type TimerQueueCallback = interface {
	QueueCallback

	OnSchedule(value any, at int64)

	OnScheduleError(value any, reason error)
}

// PriorityQueueCallback 扩展优先队列回调。
type PriorityQueueCallback = interface {
	QueueCallback

	OnPriority(value any, priority int64)
}

// RateLimitingQueueCallback 扩展限流队列回调。
type RateLimitingQueueCallback = interface {
	DelayingQueueCallback

	OnLimited(value any)
}

// RetryQueueCallback 扩展重试队列回调。
type RetryQueueCallback = interface {
	DelayingQueueCallback

	OnRetry(value any, attempt int, delay time.Duration, reason error)

	OnRetryExhausted(value any, attempt int, reason error)

	OnForget(value any)
}

// DeadLetterQueueCallback 扩展死信队列回调。
type DeadLetterQueueCallback = interface {
	QueueCallback

	OnDead(letter *DeadLetter)

	OnAckDead(letter *DeadLetter)

	OnRequeueDead(letter *DeadLetter, target Queue)
}

// LeasedQueueCallback 扩展租约队列回调。
type LeasedQueueCallback = interface {
	QueueCallback

	OnNack(value any, reason error)
}

// Limiter 决定元素下一次允许入队的等待时长。
type Limiter = interface {
	When(value any) time.Duration
}

// LimiterForgetter 是 Limiter 的可选扩展接口（io.Closer 风格）：
// 持有 per-item 状态的 Limiter 可实现该接口，由持有方通过类型断言联动清理。
// item 处理成功后调用 Forget 移除其退避状态，防止内部表无限增长。
type LimiterForgetter = interface {
	Forget(value any)
}

// RetryPolicy 决定元素下一次重试的等待时长以及是否继续重试。
type RetryPolicy = interface {
	NextDelay(value any, attempt int, reason error) (delay time.Duration, retry bool)
}

// RetryKeyFunc 生成重试计数所使用的稳定 key。
type RetryKeyFunc = func(value any) string

// Set 抽象了幂等模式下使用的集合能力。
type Set = interface {
	Add(item any)

	Remove(item any)

	Contains(item any) bool

	TryAdd(item any) bool

	TryRemove(item any) bool

	List() []any

	Len() int

	Cleanup()
}

// container 统一了列表和堆两种内部容器的最小行为集。
type container = interface {
	Push(value any)

	Pop() any

	Slice() []any

	Range(fn func(value any) bool)

	Len() int64

	Cleanup()
}

// wrapInternalList 适配 list.List 到 container。
type wrapInternalList struct {
	*lst.List
}

func (sl *wrapInternalList) Push(value any) { sl.List.PushBack(value.(*lst.Node)) }

func (sl *wrapInternalList) Pop() any { return sl.List.PopFront() }

func (sl *wrapInternalList) Range(fn func(value any) bool) {
	sl.List.Range(func(node *lst.Node) bool { return fn(node) })
}

// wrapInternalHeap 适配 heap.RBTree 到 container。
type wrapInternalHeap struct {
	*hp.RBTree
}

func (sh *wrapInternalHeap) Push(value any) { sh.RBTree.Push(value.(*lst.Node)) }

func (sh *wrapInternalHeap) Pop() any { return sh.RBTree.Pop() }

func (sh *wrapInternalHeap) Range(fn func(value any) bool) {
	sh.RBTree.Range(func(node *lst.Node) bool { return fn(node) })
}
