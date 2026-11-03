package workqueue

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// boundedBlockingQueueImpl 有界阻塞队列，基于双 channel 信号量：
// slots 控制可用槽位，items 控制已入队信号，closed 广播关停事件。
type boundedBlockingQueueImpl struct {
	Queue
	config *BoundedBlockingQueueConfig

	slots chan struct{}
	items chan struct{}

	closed chan struct{}
	once   sync.Once

	// draining drain 开始时置位，拒绝新的 Put/PutWithContext。
	draining atomic.Bool
}

// NewBoundedBlockingQueue 创建有界阻塞队列。
//
// 不支持 WithValueIdempotent：幂等标记与 Done 重入队会破坏双 channel 信号
// 与元素的 1:1 配对，导致 Get 永久阻塞。构造期 fail-fast 拒绝该组合；
// 幂等语义请使用 Queue/RetryQueue/LeasedQueue/DeadLetterQueue/TimerQueue。
func NewBoundedBlockingQueue(config *BoundedBlockingQueueConfig) BoundedBlockingQueue {
	config = isBoundedBlockingQueueConfigEffective(config)

	if config.idempotent {
		panic("workqueue: BoundedBlockingQueue does not support WithValueIdempotent: " +
			"idempotent pending-marker and Done-requeue break the 1:1 pairing " +
			"between items tokens and inner elements, hanging Get forever; " +
			"use Queue, RetryQueue, LeasedQueue, DeadLetterQueue or TimerQueue " +
			"for idempotent semantics")
	}

	capacity := config.capacity
	if capacity <= 0 {
		capacity = 1024
		config.capacity = capacity
	}

	q := &boundedBlockingQueueImpl{
		Queue:  NewQueue(&config.QueueConfig),
		config: config,
		slots:  make(chan struct{}, capacity),
		items:  make(chan struct{}, capacity),
		closed: make(chan struct{}),
	}

	for i := 0; i < capacity; i++ {
		q.slots <- struct{}{}
	}

	return q
}

// Cap 返回队列容量上限。
func (q *boundedBlockingQueueImpl) Cap() int {
	return q.config.capacity
}

// Put 阻塞式入队：获取槽位后投入内层队列并发送 items 信号。
// 队列满时阻塞，关闭或 drain 中返回 ErrQueueIsClosed。
func (q *boundedBlockingQueueImpl) Put(value any) error {
	if q.IsClosed() || q.draining.Load() {
		return ErrQueueIsClosed
	}
	if value == nil {
		return ErrElementIsNil
	}

	select {
	case <-q.closed:
		return ErrQueueIsClosed
	case <-q.slots:
	}

	// 获取槽位后复查：等待期间的 drain/Shutdown 不得再接收新元素。
	if q.draining.Load() || q.IsClosed() {
		q.releaseSlot()
		return ErrQueueIsClosed
	}

	err := q.Queue.Put(value)
	if err != nil {
		q.releaseSlot()
		return err
	}

	q.items <- struct{}{}
	return nil
}

// Get 阻塞式出队：获取 items 信号后取出元素并归还槽位。
// 队列空时阻塞，关闭时返回 ErrQueueIsClosed。
func (q *boundedBlockingQueueImpl) Get() (value any, err error) {
	if q.IsClosed() {
		return nil, ErrQueueIsClosed
	}

	select {
	case <-q.closed:
		return nil, ErrQueueIsClosed
	case <-q.items:
	}

	value, err = q.Queue.Get()
	if err != nil {
		q.releaseSlot()
		// 对齐为 ErrQueueIsClosed：保证阻塞消费入口永不返回 ErrQueueIsEmpty。
		if errors.Is(err, ErrQueueIsEmpty) {
			return nil, ErrQueueIsClosed
		}
		return nil, err
	}

	q.releaseSlot()
	return value, nil
}

// PutWithContext 是 Put 的上下文变体：支持 ctx 取消，语义与 Put 一致。
func (q *boundedBlockingQueueImpl) PutWithContext(ctx context.Context, value any) error {
	if q.IsClosed() || q.draining.Load() {
		return ErrQueueIsClosed
	}
	if value == nil {
		return ErrElementIsNil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-q.closed:
		return ErrQueueIsClosed
	case <-q.slots:
	}

	if q.draining.Load() || q.IsClosed() {
		q.releaseSlot()
		return ErrQueueIsClosed
	}

	err := q.Queue.Put(value)
	if err != nil {
		q.releaseSlot()
		return err
	}

	q.items <- struct{}{}
	return nil
}

// GetWithContext 是 Get 的上下文变体：支持 ctx 取消，语义与 Get 一致。
func (q *boundedBlockingQueueImpl) GetWithContext(ctx context.Context) (any, error) {
	if q.IsClosed() {
		return nil, ErrQueueIsClosed
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-q.closed:
		return nil, ErrQueueIsClosed
	case <-q.items:
	}

	value, err := q.Queue.Get()
	if err != nil {
		q.releaseSlot()
		if errors.Is(err, ErrQueueIsEmpty) {
			return nil, ErrQueueIsClosed
		}
		return nil, err
	}

	q.releaseSlot()
	return value, nil
}

func (q *boundedBlockingQueueImpl) Shutdown() {
	q.closeNow()
}

// ShutdownWithDrain 优雅关停：置位 draining 拒绝新入队，等待内层队列 drained；
// 超时或取消时强制关闭并返回 ctx.Err()。
func (q *boundedBlockingQueueImpl) ShutdownWithDrain(ctx context.Context) error {

	if q.IsClosed() {
		return nil
	}

	q.draining.Store(true)

	inner := q.Queue.(*queueImpl)

	err := waitForDrain(ctx, inner.IsClosed, inner.isDrained)

	q.closeNow()

	return err
}

// IsDraining 返回是否处于排空阶段。三态语义见 DrainingQueue 接口。
func (q *boundedBlockingQueueImpl) IsDraining() bool {
	return q.draining.Load()
}

func (q *boundedBlockingQueueImpl) closeNow() {
	q.once.Do(func() { close(q.closed) })
	q.Queue.Shutdown()

	// 排空 items 和 slots channel，释放残留信号。
	for {
		select {
		case <-q.items:
		default:
			goto drainSlots
		}
	}

drainSlots:
	for {
		select {
		case <-q.slots:
		default:
			return
		}
	}
}

func (q *boundedBlockingQueueImpl) releaseSlot() {
	select {
	case q.slots <- struct{}{}:
	default:
	}
}
