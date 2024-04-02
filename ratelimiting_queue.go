package workqueue

import (
	"sync"
)

// RateLimitingQInterface 是 Queue 方法的接口的限速版本
// RateLimitingQInterface is the rate-limited version of the Queue method interface
type RateLimitingQInterface interface {

	// 继承 DelayingQInterface 接口，包含了延迟队列的一些基本操作
	// Inherits DelayingQInterface, includes some basic operations of the delayed queue
	DelayingQInterface

	// AddLimited 方法用于添加一个元素到队列中，该元素会被限速处理
	// The AddLimited method is used to add an element to the queue, which will be rate-limited
	AddLimited(element any) error

	// Forget 方法用于从队列中移除一个元素，该元素不再被限速处理
	// The Forget method is used to remove an element from the queue, which will no longer be rate-limited
	Forget(element any)

	// NumLimitTimes 方法返回一个元素被限速的次数
	// The NumLimitTimes method returns the number of times an element has been rate-limited
	NumLimitTimes(element any) int
}

// RateLimitingQCallback 是 Queue 的回调接口的限速版本
// RateLimitingQCallback is the rate-limited version of the Queue callback interface
type RateLimitingQCallback interface {
	// 继承 DelayingQCallback 接口，包含了延迟队列的一些基本操作的回调
	// Inherits DelayingQCallback interface, includes callbacks for some basic operations of the delayed queue
	DelayingQCallback

	// OnAddLimited 是添加元素后的回调，参数 any 是添加的元素
	// OnAddLimited is the callback after adding an element, the parameter any is the added element
	OnAddLimited(any)

	// OnForget 是忘记元素后的回调，参数 any 是被忘记的元素
	// OnForget is the callback after forgetting an element, the parameter any is the forgotten element
	OnForget(any)

	// OnGetTimes 是获取元素被限速次数的回调，参数 any 是被限速的元素，int 是元素被限速的次数
	// OnGetTimes is the callback to get the number of times an element has been rate-limited, the parameter any is the rate-limited element, and int is the number of times the element has been rate-limited
	OnGetTimes(any, int)
}

// RateLimitingQConfig 结构体定义了限速队列的配置信息
// The RateLimitingQConfig struct defines the configuration information of the rate-limited queue
type RateLimitingQConfig struct {
	// DelayingQConfig 是延迟队列的基本配置，包含了队列的一些通用配置信息
	// DelayingQConfig is the basic configuration of the delayed queue, containing some common configuration information of the queue
	DelayingQConfig

	// callback 是一个限速队列回调接口，用于实现队列元素的处理
	// callback is a rate-limited queue callback interface, used to implement the processing of queue elements
	callback RateLimitingQCallback

	// limiter 是一个限速器，用于控制队列的处理速度
	// limiter is a rate limiter used to control the processing speed of the queue
	limiter RateLimiter
}

// NewRateLimitingQConfig 创建一个新的限速队列的配置
// NewRateLimitingQConfig creates a new configuration for a rate-limited queue
func NewRateLimitingQConfig() *RateLimitingQConfig {
	return &RateLimitingQConfig{}
}

// WithCallback 设置限速队列的回调接口
// WithCallback sets the callback interface for the rate-limited queue
func (c *RateLimitingQConfig) WithCallback(cb RateLimitingQCallback) *RateLimitingQConfig {
	c.callback = cb
	return c
}

// WithLimiter 设置限速器的实例
// WithLimiter sets the instance of the rate limiter
func (c *RateLimitingQConfig) WithLimiter(limiter RateLimiter) *RateLimitingQConfig {
	c.limiter = limiter
	return c
}

// isRateLimitingQConfigValid 函数用于验证限速队列的配置是否有效
// The isRateLimitingQConfigValid function is used to verify whether the configuration of the rate-limited queue is valid
func isRateLimitingQConfigValid(conf *RateLimitingQConfig) *RateLimitingQConfig {
	// 如果配置为空，则创建一个新的配置，并设置默认的限速器和回调
	// If the configuration is nil, create a new configuration and set the default rate limiter and callback
	if conf == nil {
		conf = NewRateLimitingQConfig().WithLimiter(DefaultBucketRateLimiter()).WithCallback(newEmptyCallback())
	} else {
		// 如果回调为空，则设置一个空的回调
		// If the callback is nil, set an empty callback
		if conf.callback == nil {
			conf.callback = newEmptyCallback()
		}

		// 如果限速器为空，则设置默认的限速器
		// If the rate limiter is nil, set the default rate limiter
		if conf.limiter == nil {
			conf.limiter = DefaultBucketRateLimiter()
		}
	}

	// 返回经过验证和可能的修改后的配置
	// Return the configuration after verification and possible modification
	return conf
}

// RateLimitingQ 结构体定义了一个限速队列的数据结构
// The RateLimitingQ struct defines a data structure for a rate-limited queue
type RateLimitingQ struct {
	// DelayingQInterface 是延迟队列的接口，定义了队列的基本操作
	// DelayingQInterface is the interface of the delayed queue, defining the basic operations of the queue
	DelayingQInterface

	// once 是一个 sync.Once 对象，用于确保某个操作只执行一次
	// once is a sync.Once object used to ensure that an operation is performed only once
	once sync.Once

	// rlock 是一个互斥锁，用于保护限速器的并发访问
	// rlock is a mutex used to protect concurrent access to the rate limiter
	rlock *sync.Mutex

	// limiter 是一个限速器，用于控制队列的处理速度
	// limiter is a rate limiter used to control the processing speed of the queue
	limiter RateLimiter

	// config 是一个指向 RateLimitingQConfig 结构体的指针，用于存储队列的配置信息和回调接口
	// config is a pointer to a RateLimitingQConfig struct used to store the configuration information and callback interface of the queue
	config *RateLimitingQConfig
}

// newRateLimitingQueue 函数用于创建一个新的 RateLimitingQueue 实例
// The newRateLimitingQueue function is used to create a new RateLimitingQueue instance
func newRateLimitingQueue(conf *RateLimitingQConfig, queue DelayingQInterface) *RateLimitingQ {
	// 如果传入的队列为空，则直接返回 nil
	// If the passed in queue is nil, return nil directly
	if queue == nil {
		return nil
	}

	// 验证并修正传入的配置，确保配置是有效的
	// Verify and correct the passed in configuration to ensure that the configuration is valid
	conf = isRateLimitingQConfigValid(conf)

	// 将回调接口设置到延迟队列的配置中
	// Set the callback interface to the configuration of the delayed queue
	conf.DelayingQConfig.callback = conf.callback

	// 创建一个新的 RateLimitingQueue 实例
	// Create a new RateLimitingQueue instance
	q := &RateLimitingQ{
		// 设置延迟队列接口
		// Set the delayed queue interface
		DelayingQInterface: queue,

		// 初始化 sync.Once 对象
		// Initialize the sync.Once object
		once: sync.Once{},

		// 初始化互斥锁
		// Initialize the mutex
		rlock: &sync.Mutex{},

		// 设置配置
		// Set the configuration
		config: conf,
	}

	// 设置限速器
	// Set the rate limiter
	q.limiter = q.config.limiter

	// 返回创建的 RateLimitingQueue 实例
	// Return the created RateLimitingQueue instance
	return q
}

// NewRateLimitingQueue 函数用于创建一个新的 RateLimitingQueue 实例
// The NewRateLimitingQueue function is used to create a new RateLimitingQueue instance
func NewRateLimitingQueue(conf *RateLimitingQConfig) *RateLimitingQ {
	// 验证并修正传入的配置，确保配置是有效的
	// Verify and correct the passed in configuration to ensure that the configuration is valid
	conf = isRateLimitingQConfigValid(conf)

	// 将回调接口设置到延迟队列的配置中
	// Set the callback interface to the configuration of the delayed queue
	conf.DelayingQConfig.callback = conf.callback

	// 使用配置和新的延迟队列创建一个新的 RateLimitingQueue 实例
	// Create a new RateLimitingQueue instance with the configuration and a new delayed queue
	return newRateLimitingQueue(conf, NewDelayingQueue(&conf.DelayingQConfig))
}

// NewRateLimitingQueueWithCustomQueue 函数用于创建一个新的 RateLimitingQueue 实例，使用自定义的 DelayingQInterface 队列
// The NewRateLimitingQueueWithCustomQueue function is used to create a new RateLimitingQueue instance, using a custom DelayingQInterface queue
func NewRateLimitingQueueWithCustomQueue(conf *RateLimitingQConfig, queue DelayingQInterface) *RateLimitingQ {
	// 验证并修正传入的配置，确保配置是有效的
	// Verify and correct the passed in configuration to ensure that the configuration is valid
	conf = isRateLimitingQConfigValid(conf)

	// 将回调接口设置到延迟队列的配置中
	// Set the callback interface to the configuration of the delayed queue
	conf.DelayingQConfig.callback = conf.callback

	// 将回调接口设置到队列的配置中
	// Set the callback interface to the configuration of the queue
	conf.DelayingQConfig.QConfig.callback = conf.callback

	// 使用配置和自定义的 DelayingQInterface 队列创建一个新的 RateLimitingQueue 实例
	// Create a new RateLimitingQueue instance with the configuration and the custom DelayingQInterface queue
	return newRateLimitingQueue(conf, queue)
}

// DefaultRateLimitingQueue 函数用于创建一个默认的 RateLimitingQueue 实例
// The DefaultRateLimitingQueue function is used to create a default RateLimitingQueue instance
func DefaultRateLimitingQueue() RateLimitingQInterface {
	// 使用 nil 配置创建一个新的 RateLimitingQueue 实例
	// Create a new RateLimitingQueue instance with nil configuration
	return NewRateLimitingQueue(nil)
}

// AddLimited 方法用于将元素添加到队列中，如果元素有 token，则直接添加到队列中，否则添加到等待队列中
// The AddLimited method is used to add an element to the queue. If the element has a token, it is directly added to the queue, otherwise it is added to the waiting queue
func (q *RateLimitingQ) AddLimited(element any) error {
	// 如果队列已经关闭，则返回 ErrorQueueClosed 错误
	// If the queue is already closed, return the ErrorQueueClosed error
	if q.IsClosed() {
		return ErrorQueueClosed
	}

	// 将元素添加到等待队列中，等待时间由 limiter 的 When 方法决定
	// Add the element to the waiting queue, the waiting time is determined by the When method of the limiter
	err := q.AddAfter(element, q.limiter.When(element))

	// 调用回调函数 OnAddLimited
	// Call the callback function OnAddLimited
	q.config.callback.OnAddLimited(element)

	// 返回错误
	// Return error
	return err
}

// Forget 方法用于忘记一个元素，即不对该元素进行限速处理
// The Forget method is used to forget an element, that is, not to rate limit the element
func (q *RateLimitingQ) Forget(element any) {
	// 调用 limiter 的 Forget 方法忘记一个元素
	// Call the Forget method of the limiter to forget an element
	q.limiter.Forget(element)

	// 调用回调函数 OnForget
	// Call the callback function OnForget
	q.config.callback.OnForget(element)
}

// NumLimitTimes 方法返回一个元素被限速的次数
// The NumLimitTimes method returns the number of times an element has been rate-limited
func (q *RateLimitingQ) NumLimitTimes(element any) int {
	// 调用 limiter 的 NumLimitTimes 方法获取一个元素被限速的次数
	// Call the NumLimitTimes method of the limiter to get the number of times an element has been rate-limited
	count := q.limiter.NumLimitTimes(element)

	// 调用回调函数 OnGetTimes
	// Call the callback function OnGetTimes
	q.config.callback.OnGetTimes(element, count)

	// 返回次数
	// Return count
	return count
}

// Stop 方法用于停止限速队列的操作
// The Stop method is used to stop the operations of the rate-limiting queue
func (q *RateLimitingQ) Stop() {
	// 调用 DelayingQInterface 的 Stop 方法，停止延迟队列的操作
	// Call the Stop method of DelayingQInterface to stop the operations of the delaying queue
	q.DelayingQInterface.Stop()

	// 使用 sync.Once 的 Do 方法确保以下操作只执行一次
	// Use the Do method of sync.Once to ensure that the following operations are only performed once
	q.once.Do(func() {
		// 调用 limiter 的 Stop 方法，停止限速器的操作
		// Call the Stop method of limiter to stop the operations of the rate limiter
		q.limiter.Stop()
	})
}
