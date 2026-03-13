package workqueue

import (
	"strconv"
	"sync/atomic"
	"time"
)

type deadLetterQueueImpl struct {
	Queue
	config *DeadLetterQueueConfig
	seed   atomic.Uint64
}

// NewDeadLetterQueue 创建死信队列。
func NewDeadLetterQueue(config *DeadLetterQueueConfig) DeadLetterQueue {
	config = isDeadLetterQueueConfigEffective(config)

	return &deadLetterQueueImpl{
		Queue:  NewQueue(&config.QueueConfig),
		config: config,
	}
}

func (q *deadLetterQueueImpl) Put(value interface{}) error {
	letter, ok := toDeadLetter(value)
	if !ok {
		return ErrInvalidDeadLetter
	}

	return q.PutDead(letter)
}

func (q *deadLetterQueueImpl) Get() (interface{}, error) {
	return q.GetDead()
}

func (q *deadLetterQueueImpl) Done(value interface{}) {
	letter, ok := toDeadLetter(value)
	if !ok {
		return
	}

	_ = q.AckDead(letter)
}

func (q *deadLetterQueueImpl) PutDead(letter *DeadLetter) error {
	if q.IsClosed() {
		return ErrQueueIsClosed
	}
	if letter == nil {
		return ErrInvalidDeadLetter
	}
	if letter.Payload == nil {
		return ErrElementIsNil
	}

	normalized := q.normalize(letter)
	if err := q.Queue.Put(normalized); err != nil {
		return err
	}

	q.config.callback.OnDead(normalized)
	return nil
}

func (q *deadLetterQueueImpl) GetDead() (*DeadLetter, error) {
	value, err := q.Queue.Get()
	if err != nil {
		return nil, err
	}

	letter, ok := toDeadLetter(value)
	if !ok {
		return nil, ErrInvalidDeadLetter
	}

	return letter, nil
}

func (q *deadLetterQueueImpl) AckDead(letter *DeadLetter) error {
	if letter == nil {
		return ErrInvalidDeadLetter
	}

	q.Queue.Done(letter)
	q.config.callback.OnAckDead(letter)
	return nil
}

func (q *deadLetterQueueImpl) RequeueDead(letter *DeadLetter, target Queue) error {
	if letter == nil {
		return ErrInvalidDeadLetter
	}
	if target == nil {
		return ErrInvalidTargetQueue
	}

	if err := target.Put(letter.Payload); err != nil {
		return err
	}

	if err := q.AckDead(letter); err != nil {
		return err
	}

	q.config.callback.OnRequeueDead(letter, target)
	return nil
}

func (q *deadLetterQueueImpl) RangeDead(fn func(letter *DeadLetter) bool) {
	if fn == nil {
		return
	}

	q.Queue.Range(func(value interface{}) bool {
		letter, ok := toDeadLetter(value)
		if !ok {
			return true
		}
		return fn(letter)
	})
}

func (q *deadLetterQueueImpl) normalize(letter *DeadLetter) *DeadLetter {
	if letter.ID == "" {
		letter.ID = q.nextID()
	}
	if letter.FailedAt.IsZero() {
		letter.FailedAt = time.Now()
	}
	return letter
}

func (q *deadLetterQueueImpl) nextID() string {
	// 使用进程内单调序列，避免时间戳拼接带来的额外开销。
	var raw [16]byte
	buf := raw[:0]
	seq := q.seed.Add(1)
	buf = strconv.AppendUint(buf, seq, 36)
	return string(buf)
}

func toDeadLetter(value interface{}) (*DeadLetter, bool) {
	switch v := value.(type) {
	case *DeadLetter:
		if v == nil {
			return nil, false
		}
		return v, true
	case DeadLetter:
		copy := v
		return &copy, true
	default:
		return nil, false
	}
}
