package workqueue

import (
	"sync"
	"time"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

func toDelay(duration int64) int64 {
	return time.Now().UnixMilli() + duration
}

// delayingQueueImpl 通过排序树维护尚未到期的元素。
type delayingQueueImpl struct {
	Queue
	config      *DelayingQueueConfig
	sorting     *hp.RBTree
	elementpool *lst.NodePool
	lock        sync.Mutex
	once        sync.Once
	wg          sync.WaitGroup
}

// NewDelayingQueue 创建延迟队列并启动搬运协程。
func NewDelayingQueue(config *DelayingQueueConfig) DelayingQueue {
	config = isDelayingQueueConfigEffective(config)
	q := &delayingQueueImpl{
		config:      config,
		sorting:     hp.New(),
		elementpool: lst.NewNodePool(),
		once:        sync.Once{},
		wg:          sync.WaitGroup{},
	}

	q.Queue = newQueue(&wrapInternalList{List: lst.New()}, q.elementpool, &config.QueueConfig)
	q.wg.Add(1)
	go q.puller()
	return q
}

func (q *delayingQueueImpl) Shutdown() {
	q.Queue.Shutdown()
	q.once.Do(func() {
		q.lock.Lock()
		q.sorting.Range(func(node *lst.Node) bool {
			q.elementpool.Put(node)
			return true
		})
		q.sorting.Cleanup()
		q.lock.Unlock()
		q.wg.Wait()
	})
}

func (q *delayingQueueImpl) PutWithDelay(value interface{}, delay int64) error {

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

func (q *delayingQueueImpl) puller() {
	heartbeat := time.NewTicker(time.Millisecond * 300)
	defer func() {
		heartbeat.Stop()
		q.wg.Done()
	}()

	var expired []*lst.Node

	for !q.IsClosed() {
		now := time.Now().UnixMilli()
		q.lock.Lock()
		expired = expired[:0]
		for q.sorting.Len() > 0 && q.sorting.Front().Priority <= now {
			expired = append(expired, q.sorting.Pop())
			if len(expired) >= 128 {
				break
			}
		}
		q.lock.Unlock()

		for _, node := range expired {
			value := node.Value
			q.elementpool.Put(node)
			if err := q.Queue.Put(value); err != nil {
				q.config.callback.OnPullError(value, err)
			}
		}

		if len(expired) == 0 {
			<-heartbeat.C
		}
	}
}

func (q *delayingQueueImpl) HeapRange(fn func(value interface{}, delay int64) bool) {
	q.lock.Lock()
	q.sorting.Range(func(n *lst.Node) bool {
		return fn(n.Value, n.Priority)
	})
	q.lock.Unlock()
}

func (q *delayingQueueImpl) Len() int {
	q.lock.Lock()
	count := int(q.sorting.Len())
	q.lock.Unlock()
	return count + q.Queue.Len()
}
