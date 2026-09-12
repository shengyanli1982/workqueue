package workqueue

import (
	"context"
	"errors"
	"sync"
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

	// G3 契约：阻塞消费永不返回 ErrQueueIsEmpty——令牌与元素配对下该窗口
	// 仅存在于关停竞态，透传的空队列错误对齐为 ErrQueueIsClosed。
	_, err := q.GetWithContext(context.Background())
	assert.ErrorIs(t, err, ErrQueueIsClosed)

	// 本测试的核心意图不变：错误路径必须释放槽位，后续容量 1 的入队不阻塞。
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	assert.NoError(t, q.PutWithContext(ctx, "ok"))
}

// TestBoundedBlockingQueue_RejectsIdempotentConfig 钉住 P0 修复：幂等配置必须在
// 构造时 fail-fast（panic）。修复前该组合确定性挂死（单线程可复现）：
// Put(v)→Get()→Put(v) 打挂起标记（内层 list 仍空但 bounded 层已发 items 信号）
// →Get() 吞掉信号返回 ErrQueueIsEmpty→Done(v) 内层重入队 list=[v]，而 bounded
// 层无 Done 覆写、无人补发 items 信号→此后 Get 永久阻塞而 Len()==1。
// 根因：双 channel 信号与内层元素的 1:1 配对假设被幂等挂起标记/Done 重入队破坏。
func TestBoundedBlockingQueue_RejectsIdempotentConfig(t *testing.T) {
	config := NewBoundedBlockingQueueConfig().WithCapacity(8)
	config.WithValueIdempotent()

	assert.PanicsWithValue(t,
		"workqueue: BoundedBlockingQueue does not support WithValueIdempotent: "+
			"idempotent pending-marker and Done-requeue break the 1:1 pairing "+
			"between items tokens and inner elements, hanging Get forever; "+
			"use Queue, RetryQueue, LeasedQueue, DeadLetterQueue or TimerQueue "+
			"for idempotent semantics",
		func() { NewBoundedBlockingQueue(config) },
	)
}

// TestBoundedBlockingQueue_ConcurrentPutShutdown_RejectAfterClose 为 #6(P2) 钉住测试：
// 并发 Put/PutWithContext + Shutdown 压力（-race 运行），断言：
//   - 运行期间 Put 只可能返回 nil 或 ErrQueueIsClosed（不透传其他错误）；
//   - Shutdown 返回后，所有 Put/PutWithContext 必须返回 ErrQueueIsClosed 而非 nil。
//
// Shutdown 本身允许丢弃已入队元素，故不断言丢失为零；本修复在获取槽位后复查
// IsClosed，提前拒绝关停后的入队，缩小“Put 成功后元素被静默丢弃”的窗口。
func TestBoundedBlockingQueue_ConcurrentPutShutdown_RejectAfterClose(t *testing.T) {
	const rounds = 50
	for r := 0; r < rounds; r++ {
		q := NewBoundedBlockingQueue(NewBoundedBlockingQueueConfig().WithCapacity(4))

		var wg sync.WaitGroup
		// 生产者持续 Put 直到观察到关停错误，保证与 Shutdown 存在真实并发重叠。
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 10000; j++ {
					value := id*10000 + j
					var err error
					if id%2 == 0 {
						err = q.Put(value)
					} else {
						err = q.PutWithContext(context.Background(), value)
					}
					if err != nil {
						if !errors.Is(err, ErrQueueIsClosed) {
							t.Errorf("round %d: Put returned unexpected error: %v", r, err)
						}
						return
					}
				}
			}(i)
		}
		// 消费者加速槽位周转，触发 Get 后 releaseSlot 回填 slots 的路径。
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					if _, err := q.Get(); err != nil {
						return
					}
				}
			}()
		}

		// 变化关停相位，覆盖“select 阻塞中 / 刚通过 select / 复查点前后”等交错。
		time.Sleep(time.Duration(r%5) * time.Millisecond)
		q.Shutdown()
		wg.Wait()

		// 钉住：close 后所有 Put 必须返回 ErrQueueIsClosed 而非 nil。
		for i := 0; i < 16; i++ {
			assert.ErrorIs(t, q.Put(i), ErrQueueIsClosed, "Put after Shutdown must be rejected")
			assert.ErrorIs(t, q.PutWithContext(context.Background(), i), ErrQueueIsClosed,
				"PutWithContext after Shutdown must be rejected")
		}
	}
}
