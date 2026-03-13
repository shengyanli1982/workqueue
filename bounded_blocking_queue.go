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

	select {
	case <-q.closed:
		q.releaseSlot()
		return ErrQueueIsClosed
	case q.items <- struct{}{}:
		return nil
	}
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

	select {
	case <-q.closed:
		q.releaseSlot()
		return ErrQueueIsClosed
	case q.items <- struct{}{}:
		return nil
	}
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
	q.once.Do(func() {
		close(q.closed)
	})
	q.Queue.Shutdown()
}

func (q *boundedBlockingQueueImpl) releaseSlot() {
	select {
	case q.slots <- struct{}{}:
	default:
	}
}
