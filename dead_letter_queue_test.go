package workqueue

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeadLetterQueue_PutDeadAndGetDead(t *testing.T) {
	q := NewDeadLetterQueue(nil)
	defer q.Shutdown()

	err := q.PutDead(&DeadLetter{
		Payload:     "task-1",
		SourceQueue: "orders",
		Attempts:    3,
		LastError:   "timeout",
		Meta: map[string]string{
			"tenant": "acme",
		},
	})
	assert.NoError(t, err)

	letter, err := q.GetDead()
	assert.NoError(t, err)
	assert.Equal(t, "task-1", letter.Payload)
	assert.Equal(t, "orders", letter.SourceQueue)
	assert.Equal(t, 3, letter.Attempts)
	assert.Equal(t, "timeout", letter.LastError)
	assert.NotEmpty(t, letter.ID)
	assert.False(t, letter.FailedAt.IsZero())
	assert.Equal(t, "acme", letter.Meta["tenant"])
}

func TestDeadLetterQueue_PutDead_NilPayload(t *testing.T) {
	q := NewDeadLetterQueue(nil)
	defer q.Shutdown()

	err := q.PutDead(&DeadLetter{
		Payload: nil,
	})
	assert.ErrorIs(t, err, ErrElementIsNil)
}

func TestDeadLetterQueue_Put_InvalidType(t *testing.T) {
	q := NewDeadLetterQueue(nil)
	defer q.Shutdown()

	err := q.Put("not-dead-letter")
	assert.ErrorIs(t, err, ErrInvalidDeadLetter)
}

func TestDeadLetterQueue_AckAndRequeue(t *testing.T) {
	dlq := NewDeadLetterQueue(nil)
	defer dlq.Shutdown()

	target := NewQueue(nil)
	defer target.Shutdown()

	err := dlq.PutDead(&DeadLetter{
		ID:        "dlq-1",
		Payload:   "recover-task",
		LastError: "panic",
		FailedAt:  time.Now(),
	})
	assert.NoError(t, err)

	letter, err := dlq.GetDead()
	assert.NoError(t, err)

	err = dlq.RequeueDead(letter, target)
	assert.NoError(t, err)

	v, err := target.Get()
	assert.NoError(t, err)
	assert.Equal(t, "recover-task", v)
}

func TestDeadLetterQueue_Requeue_InvalidTarget(t *testing.T) {
	dlq := NewDeadLetterQueue(nil)
	defer dlq.Shutdown()

	letter := &DeadLetter{
		ID:       "dlq-2",
		Payload:  "recover-task",
		FailedAt: time.Now(),
	}

	err := dlq.RequeueDead(letter, nil)
	assert.ErrorIs(t, err, ErrInvalidTargetQueue)
}

func TestDeadLetterQueue_RangeDead(t *testing.T) {
	dlq := NewDeadLetterQueue(nil)
	defer dlq.Shutdown()

	assert.NoError(t, dlq.PutDead(&DeadLetter{Payload: "a", FailedAt: time.Now()}))
	assert.NoError(t, dlq.PutDead(&DeadLetter{Payload: "b", FailedAt: time.Now()}))

	got := make([]interface{}, 0, 2)
	dlq.RangeDead(func(letter *DeadLetter) bool {
		got = append(got, letter.Payload)
		return true
	})

	assert.ElementsMatch(t, []interface{}{"a", "b"}, got)
}

type testDeadLetterQueueCallback struct {
	mu sync.Mutex

	deads    []string
	acks     []string
	requeues []string
}

func (c *testDeadLetterQueueCallback) OnPut(interface{}) {}

func (c *testDeadLetterQueueCallback) OnGet(interface{}) {}

func (c *testDeadLetterQueueCallback) OnDone(interface{}) {}

func (c *testDeadLetterQueueCallback) OnDead(letter *DeadLetter) {
	c.mu.Lock()
	c.deads = append(c.deads, letter.ID)
	c.mu.Unlock()
}

func (c *testDeadLetterQueueCallback) OnAckDead(letter *DeadLetter) {
	c.mu.Lock()
	c.acks = append(c.acks, letter.ID)
	c.mu.Unlock()
}

func (c *testDeadLetterQueueCallback) OnRequeueDead(letter *DeadLetter, _ Queue) {
	c.mu.Lock()
	c.requeues = append(c.requeues, letter.ID)
	c.mu.Unlock()
}

func TestDeadLetterQueue_Callback(t *testing.T) {
	callback := &testDeadLetterQueueCallback{}
	dlq := NewDeadLetterQueue(NewDeadLetterQueueConfig().WithCallback(callback))
	defer dlq.Shutdown()

	target := NewQueue(nil)
	defer target.Shutdown()

	assert.NoError(t, dlq.PutDead(&DeadLetter{
		ID:       "dlq-cb-1",
		Payload:  "payload",
		FailedAt: time.Now(),
	}))

	letter, err := dlq.GetDead()
	assert.NoError(t, err)

	assert.NoError(t, dlq.RequeueDead(letter, target))

	callback.mu.Lock()
	defer callback.mu.Unlock()
	assert.Equal(t, []string{"dlq-cb-1"}, callback.deads)
	assert.Equal(t, []string{"dlq-cb-1"}, callback.acks)
	assert.Equal(t, []string{"dlq-cb-1"}, callback.requeues)
}
