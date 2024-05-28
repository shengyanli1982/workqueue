package workqueue

import (
	"sync"
	"time"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

func toDelay(duration int64) int64 {
	return time.Now().Add(time.Millisecond * time.Duration(duration)).UnixMilli()
}

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
		sorting:     hp.New(),
		elementpool: lst.NewNodePool(),
		lock:        sync.Mutex{},
		once:        sync.Once{},
		wg:          sync.WaitGroup{},
	}

	q.Queue = newQueue(lst.New(), q.elementpool, &config.QueueConfig)

	q.wg.Add(1)
	go q.puller()

	return q
}

func (q *DelayingQueueImpl) Shutdown() {
	q.Queue.Shutdown()

	q.once.Do(func() {
		q.wg.Wait()

		q.lock.Lock()
		q.sorting.Cleanup()
		q.lock.Unlock()
	})
}

func (q *DelayingQueueImpl) PutWithDelay(value interface{}, delay int64) error {
	if q.IsClosed() {
		return ErrQueueIsClosed
	}

	if value == nil {
		return ErrElementIsNil
	}

	last := q.elementpool.Get()
	last.Value = value
	last.Priority = toDelay(delay)

	q.lock.Lock()

	q.sorting.Push(last)

	q.lock.Unlock()

	q.config.callback.OnDelay(value, delay)

	return nil
}

func (q *DelayingQueueImpl) puller() {
	heartbeat := time.NewTicker(time.Millisecond * 300)
	defer func() {
		heartbeat.Stop()
		q.wg.Done()
	}()

	for {
		if q.IsClosed() {
			break
		}

		q.lock.Lock()

		if q.sorting.Len() > 0 {
			top := q.sorting.Pop()
			value := top.Value
			q.lock.Unlock()

			q.elementpool.Put(top)

			if err := q.Queue.Put(value); err != nil {
				q.config.callback.OnPullError(value, err)
			}
		} else {
			q.lock.Unlock()
			<-heartbeat.C
		}
	}
}
