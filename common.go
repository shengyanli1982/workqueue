package workqueue

import "time"

// defaultQueueSortWin 是默认的队列排序窗口大小
// defaultQueueSortWin is the default queue sort window size
const defaultQueueSortWin = 500

// defaultQueueHeartbeat 是默认的队列心跳间隔（毫秒）
// defaultQueueHeartbeat is the default queue heartbeat interval (milliseconds)
const defaultQueueHeartbeat = 500

// defaultQueueRateLimit 是默认的队列速率限制
// defaultQueueRateLimit is the default queue rate limit
const defaultQueueRateLimit = 100

// defaultQueueRateBurst 是默认的队列速率突发值
// defaultQueueRateBurst is the default queue rate burst value
const defaultQueueRateBurst = 100

// defaultQueueExpFailureBase 是默认的队列指数失败基数
// defaultQueueExpFailureBase is the default queue exponential failure base
const defaultQueueExpFailureBase = 100

// defaultQueueExpFailureMax 是默认的队列指数失败最大值
// defaultQueueExpFailureMax is the default queue exponential failure maximum value
const defaultQueueExpFailureMax = 500

// emptyCallback 是一个空的回调结构体，用于在没有提供回调函数时使用
// emptyCallback is an empty callback structure, used when no callback function is provided
type emptyCallback struct{}

// OnDone 是一个空的完成回调函数，不执行任何操作
// OnDone is an empty done callback function, it does not perform any operations
func (emptyCallback) OnDone(_ any) {}

// OnAdd 是一个空的添加回调函数，不执行任何操作
// OnAdd is an empty add callback function, it does not perform any operations
func (emptyCallback) OnAdd(_ any) {}

// OnGet 是一个空的获取回调函数，不执行任何操作
// OnGet is an empty get callback function, it does not perform any operations
func (emptyCallback) OnGet(_ any) {}

// OnAddAfter 是一个空的添加后回调函数，不执行任何操作
// OnAddAfter is an empty add after callback function, it does not perform any operations
func (emptyCallback) OnAddAfter(_ any, _ time.Duration) {}

// OnAddWeight 是一个空的添加权重回调函数，不执行任何操作
// OnAddWeight is an empty add weight callback function, it does not perform any operations
func (emptyCallback) OnAddWeight(_ any, _ int) {}

// OnAddLimited 是一个空的添加限制回调函数，不执行任何操作
// OnAddLimited is an empty add limited callback function, it does not perform any operations
func (emptyCallback) OnAddLimited(_ any) {}

// OnForget 是一个空的忘记回调函数，不执行任何操作
// OnForget is an empty forget callback function, it does not perform any operations
func (emptyCallback) OnForget(_ any) {}

// OnGetTimes 是一个空的获取次数回调函数，不执行任何操作
// OnGetTimes is an empty get times callback function, it does not perform any operations
func (emptyCallback) OnGetTimes(_ any, _ int) {}

// newEmptyCallback 是一个创建新的空回调结构体的函数
// newEmptyCallback is a function to create a new empty callback structure
func newEmptyCallback() *emptyCallback {
	return &emptyCallback{}
}
