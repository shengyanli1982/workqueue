package workqueue

import "time"

// Queue 接口定义了一个队列应该具备的基本操作。
// The Queue interface defines the basic operations that a queue should have.
type Queue = interface {
	// Put 方法用于将元素放入队列。
	// The Put method is used to put an element into the queue.
	Put(interface{}) error

	// Get 方法用于从队列中获取元素。
	// The Get method is used to get an element from the queue.
	Get() (interface{}, error)

	// Done 方法用于标记元素处理完成。
	// The Done method is used to mark the element as done.
	Done(interface{})

	// Len 方法用于获取队列的长度。
	// The Len method is used to get the length of the queue.
	Len() int

	// Values 方法用于获取队列中的所有元素。
	// The Values method is used to get all the elements in the queue.
	Values() []interface{}

	// Shutdown 方法用于关闭队列。
	// The Shutdown method is used to shut down the queue.
	Shutdown()

	// IsClosed 方法用于检查队列是否已关闭。
	// The IsClosed method is used to check if the queue is closed.
	IsClosed() bool
}

// DelayingQueue 接口继承了 Queue 接口，并添加了一个 PutWithDelay 方法，用于将元素延迟放入队列。
// The DelayingQueue interface inherits from the Queue interface and adds a PutWithDelay method to put an element into the queue with delay.
type DelayingQueue = interface {
	Queue

	// PutWithDelay 方法用于将元素延迟放入队列。
	// The PutWithDelay method is used to put an element into the queue with delay.
	PutWithDelay(interface{}, int64) error
}

// PriorityQueue 接口继承了 Queue 接口，并添加了一个 PutWithPriority 方法，用于将元素按优先级放入队列。
// The PriorityQueue interface inherits from the Queue interface and adds a PutWithPriority method to put an element into the queue with priority.
type PriorityQueue = interface {
	Queue

	// PutWithPriority 方法用于将元素按优先级放入队列。
	// The PutWithPriority method is used to put an element into the queue with priority.
	PutWithPriority(interface{}, int64) error
}

// RateLimitingQueue 接口继承了 DelayingQueue 接口，并添加了一个 PutWithLimited 方法，用于将元素按速率限制放入队列。
// The RateLimitingQueue interface inherits from the DelayingQueue interface and adds a PutWithLimited method to put an element into the queue with rate limiting.
type RateLimitingQueue = interface {
	DelayingQueue

	// PutWithLimited 方法用于将元素按速率限制放入队列。
	// The PutWithLimited method is used to put an element into the queue with rate limiting.
	PutWithLimited(interface{}) error
}

// QueueCallback 接口定义了队列回调应该具备的基本操作。
// The QueueCallback interface defines the basic operations that a queue callback should have.
type QueueCallback = interface {
	// OnPut 方法在将元素放入队列时被调用。
	// The OnPut method is called when an element is put into the queue.
	OnPut(interface{})

	// OnGet 方法在从队列中获取元素时被调用。
	// The OnGet method is called when an element is gotten from the queue.
	OnGet(interface{})

	// OnDone 方法在元素处理完成后被调用。
	// The OnDone method is called when the element is done processing.
	OnDone(interface{})
}

// DelayingQueueCallback 接口继承了 QueueCallback 接口，并添加了 OnDelay 和 OnPullError 方法。
// The DelayingQueueCallback interface inherits from the QueueCallback interface and adds OnDelay and OnPullError methods.
type DelayingQueueCallback = interface {
	QueueCallback

	// OnDelay 方法在元素被延迟放入队列时被调用。
	// The OnDelay method is called when an element is put into the queue with delay.
	OnDelay(interface{}, int64)

	// OnPullError 方法在从队列中获取元素时出错被调用。
	// The OnPullError method is called when an error occurs while getting an element from the queue.
	OnPullError(interface{}, error)
}

// PriorityQueueCallback 接口继承了 QueueCallback 接口，并添加了 OnPriority 方法。
// The PriorityQueueCallback interface inherits from the QueueCallback interface and adds the OnPriority method.
type PriorityQueueCallback = interface {
	QueueCallback

	// OnPriority 方法在元素被按优先级放入队列时被调用。
	// The OnPriority method is called when an element is put into the queue with priority.
	OnPriority(interface{}, int64)
}

// RateLimitingQueueCallback 接口继承了 DelayingQueueCallback 接口，并添加了 OnLimited 方法。
// The RateLimitingQueueCallback interface inherits from the DelayingQueueCallback interface and adds the OnLimited method.
type RateLimitingQueueCallback = interface {
	DelayingQueueCallback

	// OnLimited 方法在元素被按速率限制放入队列时被调用。
	// The OnLimited method is called when an element is put into the queue with rate limiting.
	OnLimited(interface{})
}

// Limiter 接口定义了一个限制器应该具备的基本操作。
// The Limiter interface defines the basic operations that a limiter should have.
type Limiter = interface {
	// When 方法用于获取元素应该被放入队列的时间。
	// The When method is used to get the time when the element should be put into the queue.
	When(interface{}) time.Duration
}
