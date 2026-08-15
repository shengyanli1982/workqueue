package workqueue

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"
)

type deadLetterQueueImpl struct {
	Queue
	config *DeadLetterQueueConfig
	seed   atomic.Uint64
	// draining 在 drain 开始时置位：拒绝新的 PutDead/Put。
	draining atomic.Bool
}

// NewDeadLetterQueue 创建死信队列。
func NewDeadLetterQueue(config *DeadLetterQueueConfig) DeadLetterQueue {
	config = isDeadLetterQueueConfigEffective(config)

	return &deadLetterQueueImpl{
		Queue:  NewQueue(&config.QueueConfig),
		config: config,
	}
}

func (q *deadLetterQueueImpl) Put(value any) error {
	letter, ok := toDeadLetter(value)
	if !ok {
		return ErrInvalidDeadLetter
	}

	return q.PutDead(letter)
}

func (q *deadLetterQueueImpl) Get() (any, error) {
	return q.GetDead()
}

func (q *deadLetterQueueImpl) Done(value any) {
	letter, ok := toDeadLetter(value)
	if !ok {
		return
	}

	_ = q.AckDead(letter)
}

func (q *deadLetterQueueImpl) PutDead(letter *DeadLetter) error {
	if q.IsClosed() || q.draining.Load() {
		return ErrQueueIsClosed
	}
	if letter == nil {
		return ErrInvalidDeadLetter
	}
	if letter.Payload == nil {
		return ErrElementIsNil
	}

	// normalize 原地修改（填充 ID/FailedAt），返回同一指针。
	_ = q.normalize(letter)
	if err := q.Queue.Put(letter); err != nil {
		return err
	}

	q.config.callback.OnDead(letter)
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

// GetWithContext 阻塞直到消费到一个死信、ctx 完成或队列关闭，永不返回
// ErrQueueIsEmpty。委托内层队列并将结果还原为 *DeadLetter（与 GetDead 一致）。
func (q *deadLetterQueueImpl) GetWithContext(ctx context.Context) (any, error) {
	value, err := q.Queue.(BlockingGetQueue).GetWithContext(ctx)
	if err != nil {
		return nil, err
	}

	letter, ok := toDeadLetter(value)
	if !ok {
		return nil, ErrInvalidDeadLetter
	}

	return letter, nil
}

// ShutdownWithDrain 优雅关停：置位 draining 拒绝新死信写入，主队列（死信存储区）
// drained 即视为完成；超时或取消时强制关闭并返回 ctx.Err()。
// 契约：死信侧为存储区，drain 仅按基础队列语义等待在队项清空与
// Get/AckDead（Done）的 in-flight 追踪，不额外追踪死信的 Ack 状态。
func (q *deadLetterQueueImpl) ShutdownWithDrain(ctx context.Context) error {

	if q.IsClosed() {
		return nil
	}

	q.draining.Store(true)

	inner := q.Queue.(*queueImpl)

	err := waitForDrain(ctx, inner.IsClosed, inner.isDrained)

	q.Queue.Shutdown()

	return err
}

// AckDead 确认死信处理完成。
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

	if err := q.AckDead(letter); err != nil {
		return err
	}

	if err := target.Put(letter.Payload); err != nil {
		_ = q.PutDead(letter)
		return err
	}

	q.config.callback.OnRequeueDead(letter, target)
	return nil
}

func (q *deadLetterQueueImpl) RangeDead(fn func(letter *DeadLetter) bool) {
	if fn == nil {
		return
	}

	q.Queue.Range(func(value any) bool {
		letter, ok := toDeadLetter(value)
		if !ok {
			return true
		}
		return fn(letter)
	})
}

// normalize 原地填充缺失字段（ID、FailedAt），不分配新对象。
//
// 注意：调用方直接使用传入的 letter 即可，无需关心返回值；
// 保留返回签名以兼容未来需要拷贝的场景。
func (q *deadLetterQueueImpl) normalize(letter *DeadLetter) *DeadLetter {
	if letter.ID == "" {
		letter.ID = q.nextID()
	}
	if letter.FailedAt.IsZero() {
		letter.FailedAt = time.Now()
	}
	return letter
}

// nextID 生成进程内单调递增的 base-36 ID。
//
// 使用 strconv.FormatUint 直接产出 string，避免
// 手写 [N]byte + strconv.AppendUint 后再 string(buf) 的两步
// 转换。FormatUint 内部仍使用栈上缓冲区并返回一个新分配的
// string，堆分配次数与旧实现一致（1 次），但函数体更小，
// 编译期成本（~35）远低于内联预算（80），可被调用方内联，
// 从而消除一次间接调用开销。
func (q *deadLetterQueueImpl) nextID() string {
	return strconv.FormatUint(q.seed.Add(1), 36)
}

func toDeadLetter(value any) (*DeadLetter, bool) {
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
