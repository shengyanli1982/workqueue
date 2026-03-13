package workqueue

import (
	"context"
	"time"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// Queue 定义基础队列语义：消费端 Get 成功后应调用 Done。
type Queue = interface {
	Put(value interface{}) error

	Get() (value interface{}, err error)

	Done(value interface{})

	Len() int

	Values() []interface{}

	Range(fn func(value interface{}) bool)

	Shutdown()

	IsClosed() bool
}

// DelayingQueue 在普通队列基础上支持按延迟时间入队。
type DelayingQueue = interface {
	Queue

	PutWithDelay(value interface{}, delay int64) error

	HeapRange(fn func(value interface{}, delay int64) bool)
}

// PriorityQueue 在普通队列基础上支持按优先级入队。
type PriorityQueue = interface {
	Queue

	PutWithPriority(value interface{}, priority int64) error

	HeapRange(fn func(value interface{}, delay int64) bool)
}

// RateLimitingQueue 在 DelayingQueue 基础上提供限流入队能力。
type RateLimitingQueue = interface {
	DelayingQueue

	PutWithLimited(value interface{}) error
}

// RetryQueue 在 DelayingQueue 基础上提供失败重试能力。
type RetryQueue = interface {
	DelayingQueue

	Retry(value interface{}, reason error) error

	Forget(value interface{})

	NumRequeues(value interface{}) int
}

// DeadLetter 保存失败终态任务及其诊断元数据。
type DeadLetter struct {
	ID          string
	Payload     interface{}
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

	GetWithLease(timeout time.Duration) (value interface{}, leaseID string, err error)

	Ack(leaseID string) error

	Nack(leaseID string, reason error) error

	ExtendLease(leaseID string, timeout time.Duration) error
}

// BoundedBlockingQueue 在基础队列上提供容量限制和阻塞读写。
type BoundedBlockingQueue = interface {
	Queue

	Cap() int

	PutWithContext(ctx context.Context, value interface{}) error

	GetWithContext(ctx context.Context) (value interface{}, err error)
}

// TimerQueue 在基础队列上提供按绝对时间调度入队。
type TimerQueue = interface {
	Queue

	PutAt(value interface{}, at time.Time) error

	PutAfter(value interface{}, after time.Duration) error

	Cancel(value interface{}) bool

	HeapRange(fn func(value interface{}, at int64) bool)
}

// QueueCallback 定义基础队列生命周期回调。
type QueueCallback = interface {
	OnPut(value interface{})

	OnGet(value interface{})

	OnDone(value interface{})
}

// DelayingQueueCallback 扩展延迟队列回调。
type DelayingQueueCallback = interface {
	QueueCallback

	OnDelay(value interface{}, delay int64)

	OnPullError(value interface{}, reason error)
}

// PriorityQueueCallback 扩展优先队列回调。
type PriorityQueueCallback = interface {
	QueueCallback

	OnPriority(value interface{}, priority int64)
}

// RateLimitingQueueCallback 扩展限流队列回调。
type RateLimitingQueueCallback = interface {
	DelayingQueueCallback

	OnLimited(value interface{})
}

// RetryQueueCallback 扩展重试队列回调。
type RetryQueueCallback = interface {
	DelayingQueueCallback

	OnRetry(value interface{}, attempt int, delay time.Duration, reason error)

	OnRetryExhausted(value interface{}, attempt int, reason error)

	OnForget(value interface{})
}

// DeadLetterQueueCallback 扩展死信队列回调。
type DeadLetterQueueCallback = interface {
	QueueCallback

	OnDead(letter *DeadLetter)

	OnAckDead(letter *DeadLetter)

	OnRequeueDead(letter *DeadLetter, target Queue)
}

// Limiter 决定元素下一次允许入队的等待时长。
type Limiter = interface {
	When(value interface{}) time.Duration
}

// RetryPolicy 决定元素下一次重试的等待时长以及是否继续重试。
type RetryPolicy = interface {
	NextDelay(value interface{}, attempt int, reason error) (delay time.Duration, retry bool)
}

// RetryKeyFunc 生成重试计数所使用的稳定 key。
type RetryKeyFunc = func(value interface{}) string

// Set 抽象了幂等模式下使用的集合能力。
type Set = interface {
	Add(item interface{})

	Remove(item interface{})

	Contains(item interface{}) bool

	List() []interface{}

	Len() int

	Cleanup()
}

// container 统一了列表和堆两种内部容器的最小行为集。
type container = interface {
	Push(value interface{})

	Pop() interface{}

	Slice() []interface{}

	Range(fn func(value interface{}) bool)

	Len() int64

	Cleanup()
}

// wrapInternalList 适配 list.List 到 container。
type wrapInternalList struct {
	*lst.List
}

func (sl *wrapInternalList) Push(value interface{}) { sl.List.PushBack(value.(*lst.Node)) }

func (sl *wrapInternalList) Pop() interface{} { return sl.List.PopFront() }

func (sl *wrapInternalList) Range(fn func(value interface{}) bool) {
	sl.List.Range(func(node *lst.Node) bool { return fn(node) })
}

// wrapInternalHeap 适配 heap.RBTree 到 container。
type wrapInternalHeap struct {
	*hp.RBTree
}

func (sh *wrapInternalHeap) Push(value interface{}) { sh.RBTree.Push(value.(*lst.Node)) }

func (sh *wrapInternalHeap) Pop() interface{} { return sh.RBTree.Pop() }

func (sh *wrapInternalHeap) Range(fn func(value interface{}) bool) {
	sh.RBTree.Range(func(node *lst.Node) bool { return fn(node) })
}
