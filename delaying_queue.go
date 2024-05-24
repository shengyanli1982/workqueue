package workqueue

import (
	"fmt"
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

	e := q.elementpool.Get()
	e.Value = value
	e.Index = toDelay(delay)

	fmt.Printf("Value: %v, Time: %v\n", value, e.Index)

	q.sorting.Push(e)

	q.config.callback.OnDelay(value, delay)

	return nil
}

func (q *DelayingQueueImpl) puller() {
	heartbeat := time.NewTicker(time.Millisecond * 300)
	defer func() {
		heartbeat.Stop()
		q.wg.Done()
	}()

	now := time.Now().UnixMilli()

	for {
		if q.IsClosed() {
			break
		}

		select {
		case t := <-heartbeat.C:
			now = t.UnixMilli()
		default:
			q.lock.Lock()
			if q.sorting.Len() > 0 {
				e := q.sorting.Front()
				if e.Index <= now {
					fmt.Printf("Heap: %v\n", q.sorting.Slice())
					v := e.Value
					q.sorting.Pop()
					q.elementpool.Put(e)
					q.lock.Unlock()
					fmt.Printf("Value: %v\n", v)
					if err := q.Queue.Put(v); err != nil {
						q.config.callback.OnPullError(v, err)
					}
					continue
				}
			}
			q.lock.Unlock()
		}
	}
}
