package workqueue

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLeasedQueue_GetWithLease_ExpireRequeue(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(20 * time.Millisecond).
		WithScanInterval(5 * time.Millisecond)
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("job-1"))

	value, leaseID, err := q.GetWithLease(15 * time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "job-1", value)
	assert.NotEmpty(t, leaseID)

	requeued, err := waitQueueGet(t, q, 200*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "job-1", requeued)
}

func TestLeasedQueue_Ack_NoRequeue(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(20 * time.Millisecond).
		WithScanInterval(5 * time.Millisecond)
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("job-2"))

	_, leaseID, err := q.GetWithLease(15 * time.Millisecond)
	assert.NoError(t, err)
	assert.NoError(t, q.Ack(leaseID))

	time.Sleep(40 * time.Millisecond)

	_, err = q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty)
}

func TestLeasedQueue_Nack_Requeue(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(100 * time.Millisecond).
		WithScanInterval(5 * time.Millisecond)
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("job-3"))

	_, leaseID, err := q.GetWithLease(80 * time.Millisecond)
	assert.NoError(t, err)
	assert.NoError(t, q.Nack(leaseID, errors.New("retry")))

	requeued, err := waitQueueGet(t, q, 100*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "job-3", requeued)
}

func TestLeasedQueue_ExtendLease(t *testing.T) {
	config := NewLeasedQueueConfig().
		WithLeaseDuration(20 * time.Millisecond).
		WithScanInterval(5 * time.Millisecond)
	q := NewLeasedQueue(config)
	defer q.Shutdown()

	assert.NoError(t, q.Put("job-4"))

	_, leaseID, err := q.GetWithLease(15 * time.Millisecond)
	assert.NoError(t, err)
	assert.NoError(t, q.ExtendLease(leaseID, 40*time.Millisecond))

	time.Sleep(20 * time.Millisecond)
	_, err = q.Get()
	assert.ErrorIs(t, err, ErrQueueIsEmpty)

	requeued, err := waitQueueGet(t, q, 120*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "job-4", requeued)
}

func TestLeasedQueue_InvalidLease(t *testing.T) {
	q := NewLeasedQueue(nil)
	defer q.Shutdown()

	assert.ErrorIs(t, q.Ack("missing"), ErrLeaseNotFound)
	assert.ErrorIs(t, q.Nack("missing", nil), ErrLeaseNotFound)
	assert.ErrorIs(t, q.ExtendLease("missing", 10*time.Millisecond), ErrLeaseNotFound)
	assert.ErrorIs(t, q.ExtendLease("missing", 0), ErrInvalidLeaseDuration)
}

func waitQueueGet(t *testing.T, q Queue, timeout time.Duration) (interface{}, error) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		value, err := q.Get()
		if err == nil {
			return value, nil
		}
		if !errors.Is(err, ErrQueueIsEmpty) {
			return nil, err
		}
		time.Sleep(2 * time.Millisecond)
	}

	return nil, ErrQueueIsEmpty
}
