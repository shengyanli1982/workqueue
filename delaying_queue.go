package workqueue

import (
	"sync"
	"time"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// toDelay 将相对毫秒延迟转换为绝对 Unix 毫秒时间戳。
func toDelay(duration int64) int64 {
	return time.Now().Add(time.Millisecond * time.Duration(duration)).UnixMilli()
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
	closed      bool
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
		q.closed = true
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
	// 心跳轮询用于从延迟树中搬运已到期元素。
	heartbeat := time.NewTicker(time.Millisecond * 300)
	defer func() {
		heartbeat.Stop()
		q.wg.Done()
	}()

	for !q.IsClosed() {
		q.lock.Lock()

		if q.sorting.Len() > 0 && q.sorting.Front().Priority <= time.Now().UnixMilli() {
			top := q.sorting.Pop()
			value := top.Value
			q.lock.Unlock()

			q.elementpool.Put(top)
			if err := q.Queue.Put(value); err != nil {
				q.config.callback.OnPullError(value, err)
			}
			continue
		}
		q.lock.Unlock()
		<-heartbeat.C
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
	count := int(q.sorting.Len() + q.Queue.(*queueImpl).list.Len())
	q.lock.Unlock()
	return count
}
