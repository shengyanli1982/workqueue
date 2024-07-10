package workqueue

import (
	"time"

	hp "github.com/shengyanli1982/workqueue/v2/internal/container/heap"
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// Queue 接口定义了一个队列应该具备的基本操作。
// The Queue interface defines the basic operations that a queue should have.
type Queue = interface {
	// Put 方法用于将元素放入队列。
	// The Put method is used to put an element into the queue.
	Put(value interface{}) error

	// Get 方法用于从队列中获取元素。
	// The Get method is used to get an element from the queue.
	Get() (value interface{}, err error)

	// Done 方法用于标记元素处理完成。
	// The Done method is used to mark the element as done.
	Done(value interface{})

	// Len 方法用于获取队列的长度。
	// The Len method is used to get the length of the queue.
	Len() int

	// Values 方法用于获取队列中的所有元素。
	// The Values method is used to get all the elements in the queue.
	Values() []interface{}

	// Range 方法用于遍历队列中的所有元素。
	// The Range method is used to traverse all elements in the queue.
	Range(fn func(value interface{}) bool)

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
	PutWithDelay(value interface{}, delay int64) error

	// HeapRange 方法用于遍历 sorted 堆中的所有元素。
	// The HeapRange method is used to traverse all elements in the sorted heap.
	HeapRange(fn func(value interface{}, delay int64) bool)
}

// PriorityQueue 接口继承了 Queue 接口，并添加了一个 PutWithPriority 方法，用于将元素按优先级放入队列。
// The PriorityQueue interface inherits from the Queue interface and adds a PutWithPriority method to put an element into the queue with priority.
type PriorityQueue = interface {
	Queue

	// PutWithPriority 方法用于将元素按优先级放入队列。
	// The PutWithPriority method is used to put an element into the queue with priority.
	PutWithPriority(value interface{}, priority int64) error
}

// RateLimitingQueue 接口继承了 DelayingQueue 接口，并添加了一个 PutWithLimited 方法，用于将元素按速率限制放入队列。
// The RateLimitingQueue interface inherits from the DelayingQueue interface and adds a PutWithLimited method to put an element into the queue with rate limiting.
type RateLimitingQueue = interface {
	DelayingQueue

	// PutWithLimited 方法用于将元素按速率限制放入队列。
	// The PutWithLimited method is used to put an element into the queue with rate limiting.
	PutWithLimited(value interface{}) error
}

// QueueCallback 接口定义了队列回调应该具备的基本操作。
// The QueueCallback interface defines the basic operations that a queue callback should have.
type QueueCallback = interface {
	// OnPut 方法在将元素放入队列时被调用。
	// The OnPut method is called when an element is put into the queue.
	OnPut(value interface{})

	// OnGet 方法在从队列中获取元素时被调用。
	// The OnGet method is called when an element is gotten from the queue.
	OnGet(value interface{})

	// OnDone 方法在元素处理完成后被调用。
	// The OnDone method is called when the element is done processing.
	OnDone(value interface{})
}

// DelayingQueueCallback 接口继承了 QueueCallback 接口，并添加了 OnDelay 和 OnPullError 方法。
// The DelayingQueueCallback interface inherits from the QueueCallback interface and adds OnDelay and OnPullError methods.
type DelayingQueueCallback = interface {
	QueueCallback

	// OnDelay 方法在元素被延迟放入队列时被调用。
	// The OnDelay method is called when an element is put into the queue with delay.
	OnDelay(value interface{}, delay int64)

	// OnPullError 方法在从队列中获取元素时出错被调用。
	// The OnPullError method is called when an error occurs while getting an element from the queue.
	OnPullError(value interface{}, reason error)
}

// PriorityQueueCallback 接口继承了 QueueCallback 接口，并添加了 OnPriority 方法。
// The PriorityQueueCallback interface inherits from the QueueCallback interface and adds the OnPriority method.
type PriorityQueueCallback = interface {
	QueueCallback

	// OnPriority 方法在元素被按优先级放入队列时被调用。
	// The OnPriority method is called when an element is put into the queue with priority.
	OnPriority(value interface{}, priority int64)
}

// RateLimitingQueueCallback 接口继承了 DelayingQueueCallback 接口，并添加了 OnLimited 方法。
// The RateLimitingQueueCallback interface inherits from the DelayingQueueCallback interface and adds the OnLimited method.
type RateLimitingQueueCallback = interface {
	DelayingQueueCallback

	// OnLimited 方法在元素被按速率限制放入队列时被调用。
	// The OnLimited method is called when an element is put into the queue with rate limiting.
	OnLimited(value interface{})
}

// Limiter 接口定义了一个限制器应该具备的基本操作。
// The Limiter interface defines the basic operations that a limiter should have.
type Limiter = interface {
	// When 方法用于获取元素应该被放入队列的时间。
	// The When method is used to get the time when the element should be put into the queue.
	When(value interface{}) time.Duration
}

// Set 是一个接口，定义了一组方法，用于操作集合
// Set is an interface that defines a set of methods for operating on a set
type Set = interface {
	// Add 方法用于向集合中添加一个元素
	// The Add method is used to add an element to the set
	Add(item interface{})

	// Remove 方法用于从集合中移除一个元素
	// The Remove method is used to remove an element from the set
	Remove(item interface{})

	// Contains 方法用于检查集合中是否包含一个元素，如果包含则返回 true，否则返回 false
	// The Contains method is used to check whether an element is in the set. If it is, true is returned; otherwise, false is returned
	Contains(item interface{}) bool

	// List 方法用于返回集合中所有元素的列表
	// The List method is used to return a list of all elements in the set
	List() []interface{}

	// Len 方法用于返回集合中元素的数量
	// The Len method is used to return the number of elements in the set
	Len() int

	// Cleanup 方法用于清理集合，移除所有元素
	// The Cleanup method is used to clean up the set, removing all elements
	Cleanup()
}

// elementStorage 是一个接口，定义了一组操作列表的方法
// elementStorage is an interface that defines a set of methods for operating on lists
type elementStorage = interface {
	// Push 方法用于向列表中添加一个元素
	// The Push method is used to add an element to the list
	Push(value interface{})

	// Pop 方法用于从列表中弹出一个元素
	// The Pop method is used to pop an element from the list
	Pop() interface{}

	// Slice 方法用于将列表转换为切片
	// The Slice method is used to convert the list to a slice
	Slice() []interface{}

	// Range 方法用于遍历列表中的所有元素
	// The Range method is used to traverse all elements in the list
	Range(fn func(value interface{}) bool)

	// Len 方法用于获取列表的长度
	// The Len method is used to get the length of the list
	Len() int64

	// Cleanup 方法用于清理列表
	// The Cleanup method is used to clean up the list
	Cleanup()
}

// wrapInternalList 结构体用于包装内部的 List
// The wrapInternalList struct is used to wrap the internal List
type wrapInternalList struct {
	*lst.List
}

// Push 方法用于向 WrapInternalList 中添加一个元素
// The Push method is used to add an element to the WrapInternalList
func (sl *wrapInternalList) Push(value interface{}) { sl.List.PushBack(value.(*lst.Node)) }

// Pop 方法用于从 WrapInternalList 中弹出一个元素
// The Pop method is used to pop an element from the WrapInternalList
func (sl *wrapInternalList) Pop() interface{} { return sl.List.PopFront() }

// Range 方法用于遍历 WrapInternalList 中的所有元素
// The Range method is used to traverse all elements in the WrapInternalList
func (sl *wrapInternalList) Range(fn func(value interface{}) bool) {
	sl.List.Range(func(node *lst.Node) bool { return fn(node) })
}

// wrapInternalHeap 结构体用于包装内部的 RBTree
// The wrapInternalHeap struct is used to wrap the internal RBTree
type wrapInternalHeap struct {
	*hp.RBTree
}

// Push 方法用于向 WrapInternalHeap 中添加一个元素
// The Push method is used to add an element to the WrapInternalHeap
func (sh *wrapInternalHeap) Push(value interface{}) { sh.RBTree.Push(value.(*lst.Node)) }

// Pop 方法用于从 WrapInternalHeap 中弹出一个元素
// The Pop method is used to pop an element from the WrapInternalHeap
func (sh *wrapInternalHeap) Pop() interface{} { return sh.RBTree.Pop() }

// Range 方法用于遍历 WrapInternalHeap 中的所有元素
// The Range method is used to traverse all elements in the WrapInternalHeap
func (sh *wrapInternalHeap) Range(fn func(value interface{}) bool) {
	sh.RBTree.Range(func(node *lst.Node) bool { return fn(node) })
}
