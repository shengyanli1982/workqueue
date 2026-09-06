package workqueue

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// boundedBlockingQueueImpl 实现有界阻塞队列，基于双 channel 信号量机制：
// slots channel 控制可用槽位（初始填满，Put 消费、Get/Done 归还），
// items channel 控制已入队元素信号（Put 生产、Get 消费），
// closed channel 用于广播关停事件唤醒所有阻塞 goroutine。
type boundedBlockingQueueImpl struct {
	Queue
	config *BoundedBlockingQueueConfig

	slots chan struct{}
	items chan struct{}

	closed chan struct{}
	once   sync.Once

	// draining 在 drain 开始时置位：拒绝新的 Put/PutWithContext，
	// 阻塞等待者的唤醒沿用现有 closed channel 关闭路径。
	draining atomic.Bool
}

// NewBoundedBlockingQueue 创建有界阻塞队列。
func NewBoundedBlockingQueue(config *BoundedBlockingQueueConfig) BoundedBlockingQueue {
	config = isBoundedBlockingQueueConfigEffective(config)

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

// Put 阻塞式入队：从 slots channel 获取槽位后将元素投入内层队列，
// 同时在 items channel 投放已入队信号。队列满时阻塞等待槽位释放，
// 队列关闭或 drain 中返回 ErrQueueIsClosed。
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

	// 获取槽位后复查 draining：等待期间开始的 drain 不得再接收新元素。
	if q.draining.Load() {
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

// Get 阻塞式出队：从 items channel 获取已入队信号后从内层队列取出元素，
// 同时归还槽位到 slots channel。队列空时阻塞等待新元素入队，
// 队列关闭时返回 ErrQueueIsClosed。
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
		return nil, err
	}

	q.releaseSlot()
	return value, nil
}

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

	// 获取槽位后复查 draining：等待期间开始的 drain 不得再接收新元素。
	if q.draining.Load() {
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
		// items 令牌与内层元素一一配对，正常路径 Get 不会遇空队列；
		// 仅关停清理竞态可能透传 ErrQueueIsEmpty，将其对齐为 ErrQueueIsClosed，
		// 保证阻塞消费永不返回 ErrQueueIsEmpty。
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

// ShutdownWithDrain 优雅关停：置位 draining 拒绝新入队，等待内层队列 drained
// 后关闭；超时或取消时强制关闭并返回 ctx.Err()。
// 强制关闭经 closeNow 关闭 closed channel，阻塞在 Put/Get 上的等待者
// 被唤醒并收到 ErrQueueIsClosed（沿用现有关闭路径）。
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

func (q *boundedBlockingQueueImpl) closeNow() {
	// 1. 关闭 closed channel，唤醒所有阻塞在 Put/Get select 上的 goroutine。
	//    这些 goroutine 会收到 ErrQueueIsClosed 并退出。
	q.once.Do(func() {
		close(q.closed)
	})

	// 2. 关闭底层队列，阻止新的 Put/Get 操作。
	q.Queue.Shutdown()

	// 3. 排空 items channel：释放已入队信号，确保不会因为 channel 满而阻塞。
	for {
		select {
		case <-q.items:
		default:
			goto drainSlots
		}
	}

	// 4. 排空 slots channel：释放所有可用槽位，避免 Shutdown 后残留信号。
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
