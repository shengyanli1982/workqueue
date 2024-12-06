package workqueue

import (
	"sync"
	"time"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// toDelay converts duration in milliseconds to Unix timestamp for delay calculation
// toDelay 将毫秒级持续时间转换为延迟计算用的Unix时间戳
func toDelay(duration int64) int64 {
	return time.Now().Add(time.Millisecond * time.Duration(duration)).UnixMilli()
}

// delayingQueueImpl implements DelayingQueue interface with delay functionality
// delayingQueueImpl 实现了带有延迟功能的队列接口
type delayingQueueImpl struct {
	Queue                            // 基础队列 / Base queue
	config      *DelayingQueueConfig // 延迟队列配置 / Configuration for delaying queue
	sorting     *hp.RBTree           // 用于排序延迟项的红黑树 / Red-black tree for sorting delayed items
	elementpool *lst.NodePool        // 节点对象池 / Node object pool
	lock        *sync.Mutex          // 互斥锁 / Mutex for thread safety
	once        sync.Once            // 确保只执行一次的控制器 / Ensures single execution
	wg          sync.WaitGroup       // 等待组 / Wait group for goroutines
	closed      bool                 // 队列关闭状态 / Queue closure status
}

// NewDelayingQueue creates a new DelayingQueue with the given configuration
// NewDelayingQueue 创建一个新的延迟队列，使用给定的配置
func NewDelayingQueue(config *DelayingQueueConfig) DelayingQueue {
	config = isDelayingQueueConfigEffective(config)
	q := &delayingQueueImpl{
		config:      config,
		sorting:     hp.New(),
		elementpool: lst.NewNodePool(),
		once:        sync.Once{},
		wg:          sync.WaitGroup{},
		lock:        &sync.Mutex{},
	}

	// Initialize base queue and start puller goroutine
	// 初始化基础队列并启动拉取器协程
	q.Queue = newQueue(&wrapInternalList{List: lst.New()}, q.elementpool, &config.QueueConfig)
	q.wg.Add(1)
	go q.puller()
	return q
}

// Shutdown stops the queue and cleans up resources
// Shutdown 停止队列并清理资源
func (q *delayingQueueImpl) Shutdown() {
	q.Queue.Shutdown()
	q.once.Do(func() {
		q.lock.Lock()
		q.closed = true
		q.sorting.Cleanup()
		q.lock.Unlock()
		q.wg.Wait()
	})
}

// PutWithDelay adds an item to the queue with specified delay
// PutWithDelay 添加一个带有指定延迟时间的元素到队列中
func (q *delayingQueueImpl) PutWithDelay(value interface{}, delay int64) error {
	// Check queue status and input validity
	// 检查队列状态和输入有效性
	if q.IsClosed() {
		return ErrQueueIsClosed
	}
	if value == nil {
		return ErrElementIsNil
	}

	// Prepare and add the delayed item
	// 准备并添加延迟项
	last := q.elementpool.Get()
	last.Value = value
	last.Priority = toDelay(delay)

	q.lock.Lock()
	q.sorting.Push(last)
	q.lock.Unlock()

	q.config.callback.OnDelay(value, delay)
	return nil
}

// puller continuously checks and moves ready items from sorting tree to main queue
// puller 持续检查并将已到期的项目从排序树移动到主队列
func (q *delayingQueueImpl) puller() {
	// Create heartbeat ticker for periodic checks
	// 创建心跳计时器用于定期检查
	heartbeat := time.NewTicker(time.Millisecond * 300)
	defer func() {
		heartbeat.Stop()
		q.wg.Done()
	}()

	for !q.IsClosed() {
		q.lock.Lock()
		// Check if there are items ready to be moved to main queue
		// 检查是否有项目准备好移动到主队列
		if q.sorting.Len() > 0 && q.sorting.Front().Priority <= time.Now().UnixMilli() {
			top := q.sorting.Pop()
			value := top.Value
			q.lock.Unlock()

			q.elementpool.Put(top)
			if err := q.Queue.Put(value); err != nil {
				q.config.callback.OnPullError(value, err)
			}
			continue
		}
		q.lock.Unlock()
		<-heartbeat.C
	}
}

// HeapRange iterates over items in the sorting tree
// HeapRange 遍历排序树中的所有项目
func (q *delayingQueueImpl) HeapRange(fn func(value interface{}, delay int64) bool) {
	q.lock.Lock()
	q.sorting.Range(func(n *lst.Node) bool {
		return fn(n.Value, n.Priority)
	})
	q.lock.Unlock()
}

// Len returns the total number of items in both sorting tree and main queue
// Len 返回排序树和主队列中的总项目数
func (q *delayingQueueImpl) Len() int {
	q.lock.Lock()
	count := int(q.sorting.Len() + q.Queue.(*queueImpl).list.Len())
	q.lock.Unlock()
	return count
}
