package workqueue

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type leasedItem struct {
	value    interface{}
	deadline time.Time
}

type leasedQueueImpl struct {
	Queue
	config *LeasedQueueConfig

	lock    sync.Mutex
	leases  map[string]leasedItem
	leaseID atomic.Uint64

	closed chan struct{}
	once   sync.Once
	wg     sync.WaitGroup
}

// NewLeasedQueue 创建租约队列。
func NewLeasedQueue(config *LeasedQueueConfig) LeasedQueue {
	config = isLeasedQueueConfigEffective(config)

	q := &leasedQueueImpl{
		Queue:  NewQueue(&config.QueueConfig),
		config: config,
		leases: make(map[string]leasedItem),
		closed: make(chan struct{}),
	}

	q.wg.Add(1)
	go q.requeueExpiredLeases()

	return q
}

func (q *leasedQueueImpl) GetWithLease(timeout time.Duration) (value interface{}, leaseID string, err error) {
	if timeout <= 0 {
		timeout = q.config.leaseDuration
	}
	if timeout <= 0 {
		return nil, "", ErrInvalidLeaseDuration
	}

	value, err = q.Queue.Get()
	if err != nil {
		return nil, "", err
	}

	seq := q.leaseID.Add(1)
	var raw [16]byte
	leaseID = string(strconv.AppendUint(raw[:0], seq, 36))
	deadline := time.Now().Add(timeout)

	q.lock.Lock()
	q.leases[leaseID] = leasedItem{
		value:    value,
		deadline: deadline,
	}
	q.lock.Unlock()

	return value, leaseID, nil
}

func (q *leasedQueueImpl) Ack(leaseID string) error {
	value, ok := q.removeLease(leaseID)
	if !ok {
		return ErrLeaseNotFound
	}

	q.Queue.Done(value)
	return nil
}

func (q *leasedQueueImpl) Nack(leaseID string, _ error) error {
	value, ok := q.removeLease(leaseID)
	if !ok {
		return ErrLeaseNotFound
	}

	q.Queue.Done(value)
	return q.Queue.Put(value)
}

func (q *leasedQueueImpl) ExtendLease(leaseID string, timeout time.Duration) error {
	if timeout <= 0 {
		return ErrInvalidLeaseDuration
	}

	q.lock.Lock()
	item, ok := q.leases[leaseID]
	if ok {
		item.deadline = time.Now().Add(timeout)
		q.leases[leaseID] = item
	}
	q.lock.Unlock()

	if !ok {
		return ErrLeaseNotFound
	}

	return nil
}

func (q *leasedQueueImpl) Shutdown() {
	q.once.Do(func() {
		close(q.closed)
		q.wg.Wait()

		q.lock.Lock()
		q.leases = nil
		q.lock.Unlock()
	})

	q.Queue.Shutdown()
}

func (q *leasedQueueImpl) removeLease(leaseID string) (value interface{}, ok bool) {
	if leaseID == "" {
		return nil, false
	}

	q.lock.Lock()
	item, ok := q.leases[leaseID]
	if ok {
		delete(q.leases, leaseID)
	}
	q.lock.Unlock()

	if !ok {
		return nil, false
	}
	return item.value, true
}

func (q *leasedQueueImpl) requeueExpiredLeases() {
	ticker := time.NewTicker(q.config.scanInterval)
	defer func() {
		ticker.Stop()
		q.wg.Done()
	}()

	for {
		select {
		case <-q.closed:
			return
		case <-ticker.C:
			now := time.Now()
			expired := q.collectExpired(now)
			for _, value := range expired {
				q.Queue.Done(value)
				_ = q.Queue.Put(value)
			}
		}
	}
}

func (q *leasedQueueImpl) collectExpired(now time.Time) []interface{} {
	q.lock.Lock()
	var expired []interface{}
	for id, item := range q.leases {
		if !item.deadline.After(now) {
			expired = append(expired, item.value)
			delete(q.leases, id)
		}
	}
	q.lock.Unlock()
	return expired
}
