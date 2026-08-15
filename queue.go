package workqueue

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// queueImpl 是基础队列实现，支持可选幂等语义。
type queueImpl struct {
	closed atomic.Bool
	// draining 在 drain 开始时置位：Get/Done 继续正常工作，新 Put 被拒绝返回
	// ErrQueueIsClosed；drain 期间 IsClosed 保持 false。
	draining    atomic.Bool
	once        sync.Once
	lock        sync.Mutex
	config      *QueueConfig
	list        container
	elementpool *lst.NodePool
	state       Set
	// processing 追踪已被 Get 取走但尚未 Done 的处理中元素，仅幂等模式初始化。
	// 与 state 配合区分“在队”与“处理中”两种状态：
	// 不变式：state 含 v ⟺ v 在 list ∨ (v 处理中且被再次 Put)。
	processing Set
	// trackInFlight 仅在非幂等且启用 WithDrainTracking 时为 true，
	// Get +1 / Done -1（负值钳制为 0），为 drain 判定提供 in-flight 可见性。
	trackInFlight bool
	inFlight      atomic.Int64
	// broadcast 是 GetWithContext 的唤醒信号，受 q.lock 保护。
	// 首次 GetWithContext 才在锁内惰性创建，未使用阻塞消费时保持 nil：
	// Put/Done 路径仅增加锁内一次指针判空，默认非阻塞路径零开销。
	broadcast chan struct{}
	// waiters 统计当前阻塞在 GetWithContext 等待循环中的消费者数量：
	// 注册等待时（锁内）+1，离开等待（唤醒/关闭/ctx 完成）时 -1，收支恒等。
	// Put/Done 仅在 waiters>0 时执行 close+重建广播。
	waiters atomic.Int64
	// closedCh 在关停时关闭（sync.Once 保证恰好一次），
	// 用于唤醒全部 GetWithContext 等待者返回 ErrQueueIsClosed。
	closedCh chan struct{}
}

// NewQueue 创建基础队列。
func NewQueue(config *QueueConfig) Queue {
	return newQueue(&wrapInternalList{List: lst.New()}, lst.NewNodePool(), config)
}

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

func (q *queueImpl) Shutdown() {
	q.once.Do(q.closeNow)
}

// ShutdownWithDrain 优雅关停：等待“在队项清空 ∧ in-flight 完成”后关闭；
// ctx 超时或取消时强制关闭并返回 ctx.Err()。
//
// 语义要点：
//   - 与 Shutdown 共用同一 sync.Once 内的 closeNow；并发调用时立即关闭优先生效，
//     drain 观察到已关闭后返回 nil（关停目标已达成）。
//   - drain 期间新 Put 返回 ErrQueueIsClosed，IsClosed 保持 false，Get/Done 正常；
//     若此时仍有在途元素，调用方应继续 Done 以释放处理中追踪。
//   - 幂等模式判定“list 空 ∧ processing 空”；非幂等模式启用 WithDrainTracking
//     时为“list 空 ∧ inFlight 归零”；未启用 tracking 的非幂等队列仅等在队项
//     清空（in-flight 不可知，为文档化契约）。
func (q *queueImpl) ShutdownWithDrain(ctx context.Context) error {

	if q.IsClosed() {
		return nil
	}

	q.draining.Store(true)

	err := waitForDrain(ctx, q.IsClosed, q.isDrained)

	q.once.Do(q.closeNow)

	return err
}

// closeNow 执行真实的关停清理，由 Shutdown 与 ShutdownWithDrain 共用。
func (q *queueImpl) closeNow() {

	q.closed.Store(true)

	// 关停即唤醒全部 GetWithContext 等待者：等待者重查 closed 后返回
	// ErrQueueIsClosed。经 sync.Once 保证恰好关闭一次。
	close(q.closedCh)

	q.lock.Lock()

	// 先收集所有节点，避免遍历中归还池会 Reset 指针破坏遍历。
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

	// 锁外统一归还池，缩短临界区。
	for _, node := range nodes {
		q.elementpool.Put(node)
	}
}

// drainPollInterval 是 drain 判定的轮询间隔。drain 为低频关停操作，
// 小间隔轮询保持实现简单，避免为此引入新的广播机制。
const drainPollInterval = 5 * time.Millisecond

// waitForDrain 轮询直到 drain 完成或 ctx 完成。
// 已 drained、或队列被并发 Shutdown 关闭时返回 nil；
// ctx 超时/取消时返回 ctx.Err()。
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

// isDrained 判定 drain 完成条件，判定与 Get/Put 共用 q.lock 串行化，
// 避免“弹出未计数”窗口导致 drain 提前完成。
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

func (q *queueImpl) IsClosed() bool {
	return q.closed.Load()
}

func (q *queueImpl) Len() (count int) {

	q.lock.Lock()
	count = int(q.list.Len())
	q.lock.Unlock()
	return
}

func (q *queueImpl) Values() []any {

	q.lock.Lock()
	items := q.list.Slice()
	q.lock.Unlock()
	return items
}

// InFlight 返回当前处理中元素（已 Get 未 Done）的快照副本，锁内拷贝、
// 锁外返回；副本可安全修改，不影响内部状态。
//
// 仅幂等模式（WithValueIdempotent）维护 per-value 的 processing 集合，
// 非幂等模式返回 nil（此时 in-flight 仅有计数而无值可见性）。
// 经 InFlightQueue 可选接口暴露（io.Closer 风格类型断言），不并入 Queue 接口。
func (q *queueImpl) InFlight() []any {
	if !q.config.idempotent {
		return nil
	}

	q.lock.Lock()
	items := q.processing.List()
	q.lock.Unlock()
	return items
}

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

// notifyWaitersLocked 唤醒所有阻塞在 GetWithContext 上的消费者：
// 关闭当前 broadcast 并重建新实例。调用方必须持有 q.lock —— close 因此
// 天然串行（不存在对同一 channel 的二次关闭）；无等待者时成本仅为一次
// 指针判空 + 一次原子读，默认路径零变化。
func (q *queueImpl) notifyWaitersLocked() {
	if q.broadcast != nil && q.waiters.Load() > 0 {
		close(q.broadcast)
		q.broadcast = make(chan struct{})
	}
}

func (q *queueImpl) Put(value any) error {

	if q.IsClosed() || q.draining.Load() {
		return ErrQueueIsClosed
	}

	if value == nil {
		return ErrElementIsNil
	}

	if q.config.idempotent {
		q.lock.Lock()
		if q.closed.Load() {
			q.lock.Unlock()
			return ErrQueueIsClosed
		}
		if q.processing.Contains(value) {
			// 元素处理中：打上挂起标记并接受本次 Put，真实重入队推迟到 Done 时完成。
			// TryAdd 失败仅说明标记已存在（重复挂起 Put），自然幂等。
			// 这是 D1/D2 的结构性修复点：包装队列（reaper/Retry）的“先 Put 后 Done”
			// 顺序在新语义下自动正确，leased_queue.go / retry_queue.go 零改动。
			// 挂起标记不产生立即可消费项，不唤醒等待者（由 Done 重入队时唤醒）。
			q.state.TryAdd(value)
			q.lock.Unlock()
		} else {
			if !q.state.TryAdd(value) {
				q.lock.Unlock()
				return ErrElementAlreadyExist
			}
			last := q.elementpool.Get()
			last.Value = value
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

// popLocked 持 q.lock 弹出队首元素并完成配套簿记（幂等模式双集合搬运 /
// WithDrainTracking 在途计数）：Get 与 GetWithContext 共享此代码路径，
// 保证两种消费入口的 pop 语义永不漂移。调用方负责解锁后归还节点与触发回调。
func (q *queueImpl) popLocked() (front *lst.Node, value any) {

	front = q.list.Pop().(*lst.Node)
	value = front.Value

	if q.config.idempotent {
		// 双集合搬运与 pop 同在一个临界区：元素离开“在队”状态，进入“处理中”状态。
		q.state.TryRemove(value)
		q.processing.TryAdd(value)
	} else if q.trackInFlight {
		// 与 pop 同在一个临界区计数，与 isDrained 判定串行化。
		q.inFlight.Add(1)
	}

	return front, value
}

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

// GetWithContext 阻塞直到消费到一个值、ctx 完成或队列关闭，永不返回
// ErrQueueIsEmpty。Get/Done 在 drain 期间正常工作，因此阻塞消费者可助力 drain。
//
// 机制（惰性广播 channel）：
//   - broadcast 首次调用才在锁内创建；从未使用则保持 nil，Put/Done 路径
//     仅多一次指针判空，默认非阻塞路径零变化。
//   - 无丢失唤醒的关键：空检查、broadcast 快照获取与 waiters 注册必须在
//     同一临界区完成——此后任何 Put 都被 q.lock 串行化，必然 close 该快照。
//     因此本方法必须是 queueImpl 内部实现，不能由公共 Get() 组合而来。
//   - 唤醒后的循环重入会重新检查 closed/非空/继续等待三态，
//     waiters 在每次注册 +1、每次离开等待 -1，收支恒等。
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

		// 空队列：惰性创建 + 快照获取 + 等待者注册在同一临界区完成。
		if q.broadcast == nil {
			q.broadcast = make(chan struct{})
		}
		signal := q.broadcast
		q.waiters.Add(1)
		q.lock.Unlock()

		select {
		case <-signal:
			// 被唤醒后可能有值也可能已被他人取走：重入循环重新判定。
			q.waiters.Add(-1)
		case <-q.closedCh:
			q.waiters.Add(-1)
			return nil, ErrQueueIsClosed
		case <-ctx.Done():
			q.waiters.Add(-1)
			return nil, ctx.Err()
		}
	}
}

func (q *queueImpl) Done(value any) {

	if q.IsClosed() {
		return
	}

	if q.trackInFlight {
		q.decrementInFlight()
	}

	if q.config.idempotent {
		q.lock.Lock()

		// 锁内复查 closed：Shutdown 并发清空后不得重入队（与 Put 的锁内复查对齐）。
		if q.closed.Load() {
			q.lock.Unlock()
			return
		}

		if !q.processing.TryRemove(value) {
			// 非处理中元素（含仍在队元素）的 Done 为安全 no-op，
			// 不再误清在队项的去重标记。
			q.lock.Unlock()
			return
		}

		if q.state.Contains(value) {
			// 处理期间有 Put 留下了挂起标记：重新入队，value 保留在 state
			//（与不变式一致）。重入队由 Done 驱动，不重复触发 OnPut，
			// 并唤醒可能阻塞在 GetWithContext 上的消费者。
			last := q.elementpool.Get()
			last.Value = value
			q.list.Push(last)
			q.notifyWaitersLocked()
		}

		q.lock.Unlock()
		q.config.callback.OnDone(value)
	}
}

// decrementInFlight 递减在途计数并钳制负值为 0：
// Done 对未计数项（或重复调用）不得把计数打成负数。
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
