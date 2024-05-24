package workqueue

import (
	"sync"
	"time"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

func toDelay(duration time.Duration) int64 {
	return time.Now().Add(time.Millisecond * duration).UnixMilli()
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
		q.wg.Wait()

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

	return nil
}

func (q *DelayingQueueImpl) puller() {
	heartbeat := time.NewTicker(time.Millisecond * 300)
	defer func() {
		heartbeat.Stop()
		q.wg.Done()
	}()

	// now := time.Now().UnixMilli()

}
