package workqueue

import (
	"sync"
	"time"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

type DelayingQueueImpl struct {
	Queue
	config      *DelayingQueueConfig
	sorting     *hp.Heap
	elementpool *lst.NodePool
	lock        sync.Mutex
	once        sync.Once
	wg          sync.WaitGroup
}

func NewDelayingQueue(config *DelayingQueueConfig) DelayingQueue {
	config = isDelayingQueueConfigEffective(config)

	q := &DelayingQueueImpl{
		config:      config,
		Queue:       NewQueue(&config.QueueConfig),
		sorting:     hp.New(),
		elementpool: lst.NewNodePool(),
		lock:        sync.Mutex{},
		once:        sync.Once{},
		wg:          sync.WaitGroup{},
	}

	q.wg.Add(1)
	go q.puller()

	return q
}

func (q *DelayingQueueImpl) Shutdown() {
	q.Queue.Shutdown()

	q.once.Do(func() {
		q.lock.Lock()
		q.sorting.Cleanup()
		q.lock.Unlock()
	})
}

func (q *DelayingQueueImpl) PutWithDelay(value interface{}, delay time.Duration) error {
	if q.IsClosed() {
		return ErrQueueIsClosed
	}

	if value == nil {
		return ErrElementIsNil
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	q.lock.Lock()
	defer q.lock.Unlock()

	e := q.elementpool.Get()
	e.Value = value
	e.Index = int64(delay)

	q.sorting.Push(e)

	q.config.callback.OnDelay(value, delay)

	return nil
}

func (q *DelayingQueueImpl) puller() {
	ticker := time.NewTicker(time.Millisecond * 500)
	defer func() {
		ticker.Stop()
		q.wg.Done()
	}()

	for {
		if q.IsClosed() {
			break
		}

		if q.sorting.Front() != nil && q.Len() > 0 {
			e := q.sorting.Pop()
			value := e.Value
			q.elementpool.Put(e)
			_ = q.Put(value)
		} else {
			<-ticker.C
		}
	}
}
