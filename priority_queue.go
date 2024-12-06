package workqueue

import (
	"math"
	"sync"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// Priority constants define the priority levels for queue items
// 优先级常量定义了队列项目的优先级级别
const (
	PRIORITY_SLOWEST = math.MaxInt64 // Lowest priority (slowest) / 最低优先级（最慢）
	PRIORITY_LOW     = math.MaxInt32 // Low priority / 低优先级
	PRIORITY_NORMAL  = 0             // Normal priority / 正常优先级
	PRIORITY_HIGH    = math.MinInt32 // High priority / 高优先级
	PRIORITY_FASTEST = math.MinInt64 // Highest priority (fastest) / 最高优先级（最快）
)

// priorityQueueImpl implements priority queue functionality
// priorityQueueImpl 实现优先级队列功能
type priorityQueueImpl struct {
	Queue                            // 基础队列 / Base queue
	config      *PriorityQueueConfig // 优先级队列配置 / Priority queue configuration
	sorting     *hp.RBTree           // 用于优先级排序的红黑树 / Red-black tree for priority sorting
	elementpool *lst.NodePool        // 节点对象池 / Node object pool
	lock        *sync.Mutex          // 互斥锁 / Mutex for thread safety
}

// NewPriorityQueue creates a new priority queue with the given configuration
// NewPriorityQueue 创建一个新的优先级队列，使用给定的配置
func NewPriorityQueue(config *PriorityQueueConfig) PriorityQueue {
	// Validate and set default configuration if needed
	// 验证配置并设置默认值（如果需要）
	config = isPriorityQueueConfigEffective(config)

	q := &priorityQueueImpl{
		config:      config,
		sorting:     hp.New(),
		elementpool: lst.NewNodePool(),
		lock:        &sync.Mutex{},
	}

	// Initialize base queue with wrapped heap
	// 使用包装的堆初始化基础队列
	q.Queue = newQueue(&wrapInternalHeap{RBTree: q.sorting}, q.elementpool, &config.QueueConfig)

	return q
}

// Shutdown stops the queue
// Shutdown 停止队列
func (q *priorityQueueImpl) Shutdown() {
	q.Queue.Shutdown()
}

// Put adds an item to the queue with normal priority
// Put 使用正常优先级将项目添加到队列中
func (q *priorityQueueImpl) Put(value interface{}) error {
	return q.PutWithPriority(value, PRIORITY_NORMAL)
}

// PutWithPriority adds an item to the queue with specified priority
// PutWithPriority 将项目添加到队列中，并指定优先级
func (q *priorityQueueImpl) PutWithPriority(value interface{}, priority int64) error {
	// Check queue status and input validity
	// 检查队列状态和输入有效性
	if q.IsClosed() {
		return ErrQueueIsClosed
	}

	if value == nil {
		return ErrElementIsNil
	}

	// Prepare and add the prioritized item
	// 准备并添加优先级项目
	last := q.elementpool.Get()
	last.Value = value
	last.Priority = priority

	// Thread-safe operation for adding item
	// 线程安全的添加操作
	q.lock.Lock()
	q.sorting.Push(last)
	q.lock.Unlock()

	// Notify through callback
	// 通过回调通知
	q.config.callback.OnPriority(value, priority)

	return nil
}

// HeapRange iterates over items in the priority queue
// HeapRange 遍历优先级队列中的所有项目
func (q *priorityQueueImpl) HeapRange(fn func(value interface{}, delay int64) bool) {
	q.lock.Lock()
	q.sorting.Range(func(node *lst.Node) bool {
		return fn(node.Value, node.Priority)
	})
	q.lock.Unlock()
}
