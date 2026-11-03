package workqueue

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"
)

// deadLetterQueueImpl 死信队列实现，组合模式内嵌基础 Queue，
// 原子计数器生成进程内单调递增 ID，draining 标记控制写入拒绝。
type deadLetterQueueImpl struct {
	Queue
	config *DeadLetterQueueConfig
	seed   atomic.Uint64
	// draining drain 开始时置位，拒绝新的 PutDead/Put。
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

// Put 适配 Queue 接口：转换为 *DeadLetter 后委托 PutDead，非死信类型返回错误。
func (q *deadLetterQueueImpl) Put(value any) error {
	letter, ok := toDeadLetter(value)
	if !ok {
		return ErrInvalidDeadLetter
	}

	return q.PutDead(letter)
}

// Get 适配 Queue 接口：委托 GetDead 返回 *DeadLetter。
func (q *deadLetterQueueImpl) Get() (any, error) {
	return q.GetDead()
}

// Done 适配 Queue 接口：转换为 *DeadLetter 后委托 AckDead，非死信类型静默忽略。
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

// GetWithContext 阻塞直到消费到死信、ctx 完成或队列关闭，永不返回 ErrQueueIsEmpty。
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

// ShutdownWithDrain 优雅关停：置位 draining 拒绝新写入，等待内层队列 drained；
// 超时或取消时强制关闭并返回 ctx.Err()。
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

// IsDraining 返回是否处于排空阶段。三态语义见 DrainingQueue 接口。
func (q *deadLetterQueueImpl) IsDraining() bool {
	return q.draining.Load()
}

// AckDead 确认死信处理完成，触发 OnAckDead。
func (q *deadLetterQueueImpl) AckDead(letter *DeadLetter) error {
	if letter == nil {
		return ErrInvalidDeadLetter
	}

	q.Queue.Done(letter)
	q.config.callback.OnAckDead(letter)
	return nil
}

// RequeueDead 将死信重新放回目标队列：先 AckDead 再投入 target。
// target.Put 失败时自动回退到本队列（PutDead），成功后触发 OnRequeueDead。
// 契约：letter 必须来自 GetDead（在途）。
// 已知窗口：target.Put 失败时的回退 PutDead 错误会被吞掉，调用方须自行补救。
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

// RangeDead 持锁遍历在队全部死信，fn 返回 false 时提前终止。
// fn 在锁内执行，禁止调用本队列任何方法，否则死锁。
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
func (q *deadLetterQueueImpl) nextID() string {
	return strconv.FormatUint(q.seed.Add(1), 36)
}

// toDeadLetter 安全转换为 *DeadLetter：指针直接返回，值类型取地址副本，其他返回 (nil, false)。
func toDeadLetter(value any) (*DeadLetter, bool) {
	switch v := value.(type) {
	case *DeadLetter:
		if v == nil {
			return nil, false
		}
		return v, true
	case DeadLetter:
		letterCopy := v
		return &letterCopy, true
	default:
		return nil, false
	}
}
