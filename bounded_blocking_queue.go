package workqueue

import (
	"context"
	"sync"
)

type boundedBlockingQueueImpl struct {
	Queue
	config *BoundedBlockingQueueConfig

	slots chan struct{}
	items chan struct{}

	closed chan struct{}
	once   sync.Once
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

func (q *boundedBlockingQueueImpl) Cap() int {
	return q.config.capacity
}

func (q *boundedBlockingQueueImpl) Put(value interface{}) error {
	if q.IsClosed() {
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

	err := q.Queue.Put(value)
	if err != nil {
		q.releaseSlot()
		return err
	}

	q.items <- struct{}{}
	return nil
}

func (q *boundedBlockingQueueImpl) Get() (value interface{}, err error) {
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

func (q *boundedBlockingQueueImpl) PutWithContext(ctx context.Context, value interface{}) error {
	if q.IsClosed() {
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

	err := q.Queue.Put(value)
	if err != nil {
		q.releaseSlot()
		return err
	}

	q.items <- struct{}{}
	return nil
}

func (q *boundedBlockingQueueImpl) GetWithContext(ctx context.Context) (interface{}, error) {
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
		return nil, err
	}

	q.releaseSlot()
	return value, nil
}

func (q *boundedBlockingQueueImpl) Shutdown() {
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
