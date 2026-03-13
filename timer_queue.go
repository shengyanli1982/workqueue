package workqueue

import (
	"reflect"
	"sync"
	"time"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

type timerQueueImpl struct {
	Queue
	config      *TimerQueueConfig
	sorting     *hp.RBTree
	elementpool *lst.NodePool
	lock        sync.Mutex
	once        sync.Once
	wg          sync.WaitGroup
	wake        chan struct{}
	closed      chan struct{}
}

// NewTimerQueue 创建定时队列。
func NewTimerQueue(config *TimerQueueConfig) TimerQueue {
	config = isTimerQueueConfigEffective(config)

	q := &timerQueueImpl{
		config:      config,
		sorting:     hp.New(),
		elementpool: lst.NewNodePool(),
		wake:        make(chan struct{}, 1),
		closed:      make(chan struct{}),
	}

	q.Queue = newQueue(&wrapInternalList{List: lst.New()}, q.elementpool, &config.QueueConfig)
	q.wg.Add(1)
	go q.scheduler()

	return q
}

func (q *timerQueueImpl) PutAt(value interface{}, at time.Time) error {
	if q.IsClosed() {
		return ErrQueueIsClosed
	}
	if value == nil {
		return ErrElementIsNil
	}

	atMillis := at.UnixMilli()
	if atMillis <= time.Now().UnixMilli() {
		return q.Queue.Put(value)
	}

	node := q.elementpool.Get()
	node.Value = value
	node.Priority = atMillis

	q.lock.Lock()
	front := q.sorting.Front()
	q.sorting.Push(node)
	shouldWake := front == nil || atMillis < front.Priority
	q.lock.Unlock()

	if shouldWake {
		q.notifyWake()
	}
	return nil
}

func (q *timerQueueImpl) PutAfter(value interface{}, after time.Duration) error {
	return q.PutAt(value, time.Now().Add(after))
}

func (q *timerQueueImpl) Cancel(value interface{}) bool {
	q.lock.Lock()
	var target *lst.Node
	if front := q.sorting.Front(); front != nil && matchTimerValue(front.Value, value) {
		target = q.sorting.Pop()
	} else {
		target = q.findNodeLocked(value)
		if target != nil {
			q.sorting.Remove(target)
		}
	}
	q.lock.Unlock()

	if target == nil {
		return false
	}

	q.elementpool.Put(target)
	q.notifyWake()
	return true
}

func (q *timerQueueImpl) HeapRange(fn func(value interface{}, at int64) bool) {
	if fn == nil {
		return
	}

	q.lock.Lock()
	q.sorting.Range(func(node *lst.Node) bool {
		return fn(node.Value, node.Priority)
	})
	q.lock.Unlock()
}

func (q *timerQueueImpl) Shutdown() {
	q.Queue.Shutdown()

	q.once.Do(func() {
		close(q.closed)
		q.notifyWake()
		q.wg.Wait()

		q.lock.Lock()
		q.sorting.Range(func(node *lst.Node) bool {
			q.elementpool.Put(node)
			return true
		})
		q.sorting.Cleanup()
		q.lock.Unlock()
	})
}

func (q *timerQueueImpl) Len() int {
	q.lock.Lock()
	count := int(q.sorting.Len())
	q.lock.Unlock()
	return count + q.Queue.Len()
}

func (q *timerQueueImpl) scheduler() {
	defer q.wg.Done()

	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()

	for {
		wait, due, ok := q.nextWait()
		if !ok {
			return
		}

		if due != nil {
			value := due.Value
			q.elementpool.Put(due)
			_ = q.Queue.Put(value)
			continue
		}

		if !q.wait(timer, wait) {
			return
		}
	}
}

func (q *timerQueueImpl) nextWait() (time.Duration, *lst.Node, bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	select {
	case <-q.closed:
		return 0, nil, false
	default:
	}

	front := q.sorting.Front()
	if front == nil {
		return 0, nil, true
	}

	now := time.Now().UnixMilli()
	if front.Priority <= now {
		return 0, q.sorting.Pop(), true
	}

	return time.Duration(front.Priority-now) * time.Millisecond, nil, true
}

func (q *timerQueueImpl) wait(timer *time.Timer, d time.Duration) bool {
	if d <= 0 {
		select {
		case <-q.closed:
			return false
		case <-q.wake:
			return true
		}
	}

	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(d)

	select {
	case <-q.closed:
		return false
	case <-q.wake:
		return true
	case <-timer.C:
		return true
	}
}

func (q *timerQueueImpl) notifyWake() {
	select {
	case q.wake <- struct{}{}:
	default:
	}
}

func (q *timerQueueImpl) findNodeLocked(value interface{}) *lst.Node {
	var target *lst.Node
	q.sorting.Range(func(node *lst.Node) bool {
		if matchTimerValue(node.Value, value) {
			target = node
			return false
		}
		return true
	})
	return target
}

func matchTimerValue(candidate, target interface{}) bool {
	switch tv := target.(type) {
	case int:
		cv, ok := candidate.(int)
		return ok && cv == tv
	case int8:
		cv, ok := candidate.(int8)
		return ok && cv == tv
	case int16:
		cv, ok := candidate.(int16)
		return ok && cv == tv
	case int32:
		cv, ok := candidate.(int32)
		return ok && cv == tv
	case int64:
		cv, ok := candidate.(int64)
		return ok && cv == tv
	case uint:
		cv, ok := candidate.(uint)
		return ok && cv == tv
	case uint8:
		cv, ok := candidate.(uint8)
		return ok && cv == tv
	case uint16:
		cv, ok := candidate.(uint16)
		return ok && cv == tv
	case uint32:
		cv, ok := candidate.(uint32)
		return ok && cv == tv
	case uint64:
		cv, ok := candidate.(uint64)
		return ok && cv == tv
	case uintptr:
		cv, ok := candidate.(uintptr)
		return ok && cv == tv
	case string:
		cv, ok := candidate.(string)
		return ok && cv == tv
	default:
		return reflect.DeepEqual(candidate, target)
	}
}
