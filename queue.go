package workqueue

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// queueImpl 基础队列实现，支持可选幂等语义。
type queueImpl struct {
	closed      atomic.Bool
	draining    atomic.Bool  // drain 开始时置位，drain 期间 IsClosed 保持 false
	once        sync.Once
	lock        sync.Mutex
	config      *QueueConfig
	list        container
	elementpool *lst.NodePool
	state       Set          // 幂等去重集合
	processing  Set          // 幂等模式：已 Get 未 Done 的处理中元素
	trackInFlight bool       // 非幂等 + WithDrainTracking 时启用 in-flight 计数
	inFlight    atomic.Int64
	broadcast   chan struct{} // GetWithContext 唤醒信号，惰性创建，未使用时为 nil
	waiters     atomic.Int64  // 当前阻塞等待的消费者数量
	closedCh    chan struct{} // 关停时关闭，唤醒全部等待者
}

// NewQueue 创建基础队列。
func NewQueue(config *QueueConfig) Queue {
	return newQueue(&wrapInternalList{List: lst.New()}, lst.NewNodePool(), config)
}

// newQueue 内部构造函数，接受容器与节点池，完成配置校验与字段初始化。
func newQueue(list container, elementpool *lst.NodePool, config *QueueConfig) *queueImpl {

	q := &queueImpl{
		config:      isQueueConfigEffective(config),
		list:        list,
		elementpool: elementpool,
		closedCh:    make(chan struct{}),
	}

	if q.config.idempotent {
		q.state = q.config.setCreator()
		q.processing = q.config.setCreator()
	} else {
		q.trackInFlight = q.config.drainTracking
	}

	return q
}

// Shutdown 立即关停队列，经 sync.Once 保证 closeNow 恰好执行一次。
func (q *queueImpl) Shutdown() {
	q.once.Do(q.closeNow)
}

// ShutdownWithDrain 优雅关停：等待在队项清空且 in-flight 完成后关闭；
// ctx 超时或取消时强制关闭并返回 ctx.Err()。
// 与 Shutdown 共用同一 sync.Once；并发调用时立即关闭优先生效。
// drain 期间新 Put 被拒绝，Get/Done 正常。
func (q *queueImpl) ShutdownWithDrain(ctx context.Context) error {

	if q.IsClosed() {
		return nil
	}

	q.draining.Store(true)

	err := waitForDrain(ctx, q.IsClosed, q.isDrained)

	q.once.Do(q.closeNow)

	return err
}

// closeNow 执行关停清理，由 Shutdown 与 ShutdownWithDrain 共用。
func (q *queueImpl) closeNow() {

	q.closed.Store(true)
	close(q.closedCh) // 唤醒全部 GetWithContext 等待者

	q.lock.Lock()

	nodes := make([]*lst.Node, 0, q.list.Len())
	q.list.Range(func(value any) bool {
		nodes = append(nodes, value.(*lst.Node))
		return true
	})
	q.list.Cleanup()

	if q.config.idempotent {
		q.state.Cleanup()
		q.processing.Cleanup()
	}
	q.lock.Unlock()

	for _, node := range nodes {
		q.elementpool.Put(node) // 锁外归还池，缩短临界区
	}
}

// drainPollInterval drain 轮询间隔，低频关停操作无需引入广播机制。
const drainPollInterval = 5 * time.Millisecond

// waitForDrain 轮询直到 drain 完成或 ctx 完成。
func waitForDrain(ctx context.Context, isClosed, drained func() bool) error {

	if drained() {
		return nil
	}

	ticker := time.NewTicker(drainPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		if isClosed() {
			return nil
		}

		if drained() {
			return nil
		}
	}
}

// isDrained 判定 drain 完成条件，与 Get/Put 共用 q.lock 串行化。
func (q *queueImpl) isDrained() bool {

	q.lock.Lock()
	defer q.lock.Unlock()

	if q.list.Len() != 0 {
		return false
	}

	if q.config.idempotent {
		return q.processing.Len() == 0
	}

	if q.trackInFlight {
		return q.inFlight.Load() == 0
	}

	return true
}

// IsClosed 返回队列是否已关停。
func (q *queueImpl) IsClosed() bool {
	return q.closed.Load()
}

// IsDraining 返回队列是否处于排空阶段。状态语义见 DrainingQueue 接口文档。
func (q *queueImpl) IsDraining() bool {
	return q.draining.Load()
}

// Len 返回当前在队元素数量。
func (q *queueImpl) Len() (count int) {

	q.lock.Lock()
	count = int(q.list.Len())
	q.lock.Unlock()
	return
}

// Values 返回当前在队元素的快照副本，副本可安全修改。
func (q *queueImpl) Values() []any {

	q.lock.Lock()
	items := q.list.Slice()
	q.lock.Unlock()
	return items
}

// InFlight 返回当前处理中元素（已 Get 未 Done）的快照副本。
// 仅幂等模式维护 per-value 集合，非幂等模式返回 nil。
func (q *queueImpl) InFlight() []any {
	if !q.config.idempotent {
		return nil
	}

	q.lock.Lock()
	items := q.processing.List()
	q.lock.Unlock()
	return items
}

// Range 持锁遍历在队元素，fn 返回 false 时提前终止。
// fn 在锁内执行，禁止在 fn 中回调本队列任何方法（死锁风险）。
func (q *queueImpl) Range(fn func(any) bool) {

	if fn == nil {
		return
	}

	q.lock.Lock()
	q.list.Range(func(value any) bool {
		node := value.(*lst.Node)
		return fn(node.Value)
	})
	q.lock.Unlock()
}

// notifyWaitersLocked 唤醒所有阻塞的 GetWithContext 消费者，调用方须持有 q.lock。
func (q *queueImpl) notifyWaitersLocked() {
	if q.broadcast != nil && q.waiters.Load() > 0 {
		close(q.broadcast)
		q.broadcast = make(chan struct{})
	}
}

// Put 将元素入队。幂等模式通过 state 集合去重，处理中元素打挂起标记（Done 时重入队）；
// 非幂等模式直接入队。锁内复查 closed 防止关停窗口丢失。
func (q *queueImpl) Put(value any) error {

	if q.IsClosed() || q.draining.Load() {
		return ErrQueueIsClosed
	}

	if value == nil {
		return ErrElementIsNil
	}

	if q.config.idempotent {
		// 锁外取节点，拒绝路径在锁外归还池，缩短临界区。
		last := q.elementpool.Get()
		last.Value = value
		q.lock.Lock()
		if q.closed.Load() {
			q.lock.Unlock()
			q.elementpool.Put(last)
			return ErrQueueIsClosed
		}
		if q.processing.Contains(value) {
			// 元素处理中：打挂起标记，真实重入队推迟到 Done 时完成。
			// 挂起标记不产生立即可消费项，不唤醒等待者。
			q.state.TryAdd(value)
			q.lock.Unlock()
			q.elementpool.Put(last)
		} else {
			if !q.state.TryAdd(value) {
				q.lock.Unlock()
				q.elementpool.Put(last)
				return ErrElementAlreadyExist
			}
			q.list.Push(last)
			q.notifyWaitersLocked()
			q.lock.Unlock()
		}
	} else {
		last := q.elementpool.Get()
		last.Value = value
		q.lock.Lock()
		if q.closed.Load() {
			q.lock.Unlock()
			q.elementpool.Put(last)
			return ErrQueueIsClosed
		}
		q.list.Push(last)
		q.notifyWaitersLocked()
		q.lock.Unlock()
	}

	q.config.callback.OnPut(value)
	return nil
}

// popLocked 持锁弹出队首元素并完成配套簿记（幂等双集合搬运 / drainTracking 计数），
// Get 与 GetWithContext 共享此路径。调用方负责解锁后归还节点与触发回调。
func (q *queueImpl) popLocked() (front *lst.Node, value any) {

	front = q.list.Pop().(*lst.Node)
	value = front.Value

	if q.config.idempotent {
		q.state.TryRemove(value)
		q.processing.TryAdd(value)
	} else if q.trackInFlight {
		q.inFlight.Add(1)
	}

	return front, value
}

// Get 非阻塞弹出队首元素。队列空返回 ErrQueueIsEmpty，关停返回 ErrQueueIsClosed。
func (q *queueImpl) Get() (any, error) {

	if q.IsClosed() {
		return nil, ErrQueueIsClosed
	}

	q.lock.Lock()

	if q.list.Len() == 0 {
		q.lock.Unlock()
		return nil, ErrQueueIsEmpty
	}

	front, value := q.popLocked()

	q.lock.Unlock()

	q.elementpool.Put(front)

	q.config.callback.OnGet(value)

	return value, nil
}

// GetWithContext 阻塞直到消费到值、ctx 完成或队列关闭，永不返回 ErrQueueIsEmpty。
// 使用惰性 broadcast channel 实现无丢失唤醒：空检查、快照获取与 waiters 注册
// 在同一临界区完成，此后任何 Put 被 q.lock 串行化必然 close 该快照。
func (q *queueImpl) GetWithContext(ctx context.Context) (value any, err error) {

	for {
		if q.IsClosed() {
			return nil, ErrQueueIsClosed
		}

		q.lock.Lock()

		if q.closed.Load() {
			q.lock.Unlock()
			return nil, ErrQueueIsClosed
		}

		if q.list.Len() > 0 {
			front, value := q.popLocked()

			q.lock.Unlock()

			q.elementpool.Put(front)

			q.config.callback.OnGet(value)

			return value, nil
		}

		// 惰性创建 + 快照获取 + 等待者注册在同一临界区完成。
		if q.broadcast == nil {
			q.broadcast = make(chan struct{})
		}
		signal := q.broadcast
		q.waiters.Add(1)
		q.lock.Unlock()

		select {
		case <-signal:
			q.waiters.Add(-1) // 重入循环重新判定
		case <-q.closedCh:
			q.waiters.Add(-1)
			return nil, ErrQueueIsClosed
		case <-ctx.Done():
			q.waiters.Add(-1)
			return nil, ctx.Err()
		}
	}
}

// Done 标记元素处理完成。关停时直接返回，不触发回调。
// 必须与 Get 返回的值配对调用。
//
// 幂等模式：从 processing 移除，若有挂起标记则重新入队并唤醒等待者；
// 非处理中元素的 Done 为安全 no-op。
// 非幂等模式：best-effort 完成信号，递减 inFlight（启用 drainTracking 时）并触发 OnDone。
func (q *queueImpl) Done(value any) {

	if q.IsClosed() {
		return
	}

	if q.trackInFlight {
		q.decrementInFlight()
	}

	if q.config.idempotent {
		q.lock.Lock()

		// 锁内复查 closed：关停后不得重入队。
		if q.closed.Load() {
			q.lock.Unlock()
			return
		}

		if !q.processing.TryRemove(value) {
			q.lock.Unlock() // 非处理中元素的 Done 为安全 no-op
			return
		}

		if q.state.Contains(value) {
			// 处理期间有挂起标记：重新入队并唤醒等待者。
			last := q.elementpool.Get()
			last.Value = value
			q.list.Push(last)
			q.notifyWaitersLocked()
		}

		q.lock.Unlock()
		q.config.callback.OnDone(value)
	} else {
		q.config.callback.OnDone(value) // best-effort 完成信号，无簿记校验
	}
}

// decrementInFlight 递减在途计数，钳制负值为 0。
func (q *queueImpl) decrementInFlight() {
	for {
		current := q.inFlight.Load()
		if current == 0 {
			return
		}
		if q.inFlight.CompareAndSwap(current, current-1) {
			return
		}
	}
}
