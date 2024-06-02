package workqueue

import (
	"sync"
	"time"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// toDelay 函数用于计算延迟时间，它接受一个持续时间（毫秒），并返回当前时间加上持续时间的 Unix 毫秒时间戳。
// The toDelay function is used to calculate the delay time. It takes a duration (in milliseconds) and returns the Unix millisecond timestamp of the current time plus the duration.
func toDelay(duration int64) int64 {
	// 使用 time.Now().Add 方法计算延迟时间，并使用 UnixMilli 方法将结果转换为 Unix 毫秒时间戳
	// Use the time.Now().Add method to calculate the delay time, and use the UnixMilli method to convert the result to a Unix millisecond timestamp
	return time.Now().Add(time.Millisecond * time.Duration(duration)).UnixMilli()
}

// delayingQueueImpl 结构体实现了 DelayingQueue 接口，它是一个支持延迟的队列。
// The delayingQueueImpl structure implements the DelayingQueue interface, it is a queue that supports delay.
type delayingQueueImpl struct {
	// Queue 是一个队列接口
	// Queue is a queue interface
	Queue

	// config 是队列的配置
	// config is the configuration of the queue
	config *DelayingQueueConfig

	// sorting 是一个堆，用于排序队列中的元素
	// sorting is a heap, used to sort the elements in the queue
	sorting *hp.Heap

	// elementpool 是元素内存池，用于存储队列中的元素
	// elementpool is the element memory pool, used to store the elements in the queue
	elementpool *lst.NodePool

	// lock 是一个互斥锁，用于保护队列的并发操作
	// lock is a mutex, used to protect the concurrent operations of the queue
	lock *sync.Mutex

	// once 是一个 sync.Once，用于保证某些操作只执行一次
	// once is a sync.Once, used to ensure that certain operations are only performed once
	once sync.Once

	// wg 是一个 sync.WaitGroup，用于等待队列中的所有操作完成
	// wg is a sync.WaitGroup, used to wait for all operations in the queue to complete
	wg sync.WaitGroup
}

// NewDelayingQueue 函数用于创建一个新的 DelayingQueue。
// The NewDelayingQueue function is used to create a new DelayingQueue.
func NewDelayingQueue(config *DelayingQueueConfig) DelayingQueue {
	// 检查配置是否有效，如果无效，使用默认配置
	// Check if the configuration is valid, if not, use the default configuration
	config = isDelayingQueueConfigEffective(config)

	// 创建一个新的 DelayingQueueImpl
	// Create a new DelayingQueueImpl
	q := &delayingQueueImpl{
		// 设置配置
		// Set the configuration
		config: config,

		// 创建一个新的排序堆，用于存储延迟元素
		// Create a new sorting heap for storing delayed elements
		sorting: hp.New(),

		// 创建一个新的元素内存池，用于存储队列元素，减少内存分配
		// Create a new element memory pool for storing queue elements, reducing memory allocation
		elementpool: lst.NewNodePool(),

		// 创建一个新的 sync.Once，用于确保某个操作只执行一次
		// Create a new sync.Once to ensure that an operation is performed only once
		once: sync.Once{},

		// 创建一个新的 sync.WaitGroup，用于等待所有 goroutine 完成
		// Create a new sync.WaitGroup for waiting for all goroutines to complete
		wg: sync.WaitGroup{},
	}

	// 使用 lst.New 创建一个新的队列，并将其赋值给 q.Queue
	// Use lst.New to create a new queue, and assign it to q.Queue
	q.Queue = newQueue(lst.New(), q.elementpool, &config.QueueConfig)

	// 将 q.Queue 的锁赋值给 q.lock
	// Assign the lock of q.Queue to q.lock
	q.lock = q.Queue.(*queueImpl).lock

	// 增加 wg 的计数
	// Increase the count of wg
	q.wg.Add(1)

	// 启动一个新的 goroutine，用于从队列中拉取元素
	// Start a new goroutine to pull elements from the queue
	go q.puller()

	// 返回新创建的 DelayingQueue
	// Return the newly created DelayingQueue
	return q
}

// Shutdown 方法用于关闭 DelayingQueue。
// The Shutdown method is used to shut down the DelayingQueue.
func (q *delayingQueueImpl) Shutdown() {
	// 关闭内部的 Queue
	// Shut down the internal Queue
	q.Queue.Shutdown()

	// 使用 sync.Once 确保以下操作只执行一次
	// Use sync.Once to ensure the following operations are performed only once
	q.once.Do(func() {
		// 等待所有 goroutine 完成
		// Wait for all goroutines to complete
		q.wg.Wait()

		// 加锁，保护排序堆的并发操作
		// Lock, to protect the concurrent operations of the sorting heap
		q.lock.Lock()

		// 清理排序堆
		// Clean up the sorting heap
		q.sorting.Cleanup()

		// 解锁
		// Unlock
		q.lock.Unlock()
	})
}

// PutWithDelay 方法用于将一个元素放入 DelayingQueue，并设置其延迟时间。
// The PutWithDelay method is used to put an element into the DelayingQueue and set its delay time.
func (q *delayingQueueImpl) PutWithDelay(value interface{}, delay int64) error {
	// 如果 DelayingQueue 已关闭，返回错误
	// If the DelayingQueue is closed, return an error
	if q.IsClosed() {
		return ErrQueueIsClosed
	}

	// 如果元素值为 nil，返回错误
	// If the element value is nil, return an error
	if value == nil {
		return ErrElementIsNil
	}

	// 从元素内存池中获取一个元素
	// Get an element from the element memory pool
	last := q.elementpool.Get()

	// 设置元素的值
	// Set the value of the element
	last.Value = value

	// 将延迟时间转换为优先级，并设置元素的优先级
	// Convert the delay time to priority, and set the priority of the element
	last.Priority = toDelay(delay)

	// 加锁，保护排序堆的并发操作
	// Lock, to protect the concurrent operations of the sorting heap
	q.lock.Lock()

	// 将元素放入排序堆
	// Put the element into the sorting heap
	q.sorting.Push(last)

	// 解锁
	// Unlock
	q.lock.Unlock()

	// 调用回调函数，通知元素已被放入并设置了延迟时间
	// Call the callback function to notify that the element has been put and the delay time has been set
	q.config.callback.OnDelay(value, delay)

	// 返回 nil 错误
	// Return a nil error
	return nil
}

// puller 方法用于从 DelayingQueue 中拉取元素。
// The puller method is used to pull elements from the DelayingQueue.
func (q *delayingQueueImpl) puller() {
	// 创建一个新的定时器，每 300 毫秒触发一次
	// Create a new ticker that triggers every 300 milliseconds
	heartbeat := time.NewTicker(time.Millisecond * 300)

	// 使用 defer 确保以下操作在函数返回时执行
	// Use defer to ensure the following operations are performed when the function returns
	defer func() {
		// 停止定时器
		// Stop the ticker
		heartbeat.Stop()

		// 减少 wg 的计数
		// Decrease the count of wg
		q.wg.Done()
	}()

	// 使用 for 循环不断从 DelayingQueue 中拉取元素
	// Use a for loop to continuously pull elements from the DelayingQueue
	for {
		// 如果 DelayingQueue 已关闭，跳出循环
		// If the DelayingQueue is closed, break the loop
		if q.IsClosed() {
			break
		}

		// 加锁，保护排序堆的并发操作
		// Lock, to protect the concurrent operations of the sorting heap
		q.lock.Lock()

		// 如果排序堆中有元素, 并且元素的执行时间戳小于等于当前时间戳
		// If there are elements in the sorting heap and the execution timestamp of the element is less than or equal to the current timestamp
		if q.sorting.Len() > 0 && q.sorting.Front().Priority <= time.Now().UnixMilli() {
			// 从排序堆中弹出一个元素
			// Pop an element from the sorting heap
			top := q.sorting.Pop()

			// 获取元素的值
			// Get the value of the element
			value := top.Value

			// 解锁
			// Unlock
			q.lock.Unlock()

			// 将元素放回元素内存池
			// Put the element back into the element memory pool
			q.elementpool.Put(top)

			// 将元素放入 Queue
			// Put the element into the Queue
			if err := q.Queue.Put(value); err != nil {
				// 如果放入失败，调用回调函数，通知元素拉取失败
				// If the put fails, call the callback function to notify that the element pull failed
				q.config.callback.OnPullError(value, err)
			}
		} else {
			// 如果排序堆中没有元素，解锁并等待下一次定时器触发
			// If there are no elements in the sorting heap, unlock and wait for the next ticker trigger
			q.lock.Unlock()
			<-heartbeat.C
		}
	}
}
