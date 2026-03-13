package workqueue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBoundedBlockingQueue_PutWithContext_BlockUntilGet(t *testing.T) {
	q := NewBoundedBlockingQueue(NewBoundedBlockingQueueConfig().WithCapacity(1))
	defer q.Shutdown()

	assert.NoError(t, q.Put("first"))

	result := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		result <- q.PutWithContext(ctx, "second")
	}()

	time.Sleep(30 * time.Millisecond)
	select {
	case err := <-result:
		t.Fatalf("put should block but returned: %v", err)
	default:
	}

	v, err := q.GetWithContext(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "first", v)

	assert.NoError(t, <-result)

	v, err = q.GetWithContext(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "second", v)
}

func TestBoundedBlockingQueue_GetWithContext_Timeout(t *testing.T) {
	q := NewBoundedBlockingQueue(NewBoundedBlockingQueueConfig().WithCapacity(1))
	defer q.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	_, err := q.GetWithContext(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestBoundedBlockingQueue_PutWithContext_Timeout(t *testing.T) {
	q := NewBoundedBlockingQueue(NewBoundedBlockingQueueConfig().WithCapacity(1))
	defer q.Shutdown()

	assert.NoError(t, q.Put("first"))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err := q.PutWithContext(ctx, "second")
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestBoundedBlockingQueue_ShutdownWakeup(t *testing.T) {
	q := NewBoundedBlockingQueue(NewBoundedBlockingQueueConfig().WithCapacity(1))

	result := make(chan error, 1)
	go func() {
		_, err := q.GetWithContext(context.Background())
		result <- err
	}()

	time.Sleep(20 * time.Millisecond)
	q.Shutdown()

	err := <-result
	assert.ErrorIs(t, err, ErrQueueIsClosed)
}

func TestBoundedBlockingQueue_GetWithContext_ReleasesSlotOnGetError(t *testing.T) {
	q := NewBoundedBlockingQueue(NewBoundedBlockingQueueConfig().WithCapacity(1)).(*boundedBlockingQueueImpl)
	defer q.Shutdown()

	// 模拟异常路径：items 信号已就绪，但底层队列为空。
	<-q.slots
	q.items <- struct{}{}

	_, err := q.GetWithContext(context.Background())
	assert.ErrorIs(t, err, ErrQueueIsEmpty)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	assert.NoError(t, q.PutWithContext(ctx, "ok"))
}
