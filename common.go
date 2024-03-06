package workqueue

import "time"

// defaultQueueSortWin 定义了排序窗口的默认值，单位是毫秒
// defaultQueueSortWin defines the default value of the sort window, in milliseconds
const defaultQueueSortWin = 500 // milliseconds

// defaultQueueHeartbeat 定义了心跳间隔的默认值，单位是毫秒
// defaultQueueHeartbeat defines the default value of the heartbeat interval, in milliseconds
const defaultQueueHeartbeat = 500 // milliseconds

// defaultQueueRateLimit 定义了限速设置的默认值，单位是每秒
// defaultQueueRateLimit defines the default value of the rate limit setting, per second
const defaultQueueRateLimit = 100 // per second

// defaultQueueRateBurst 定义了限速突发的默认值，单位是每秒
// defaultQueueRateBurst defines the default value of the rate burst, per second
const defaultQueueRateBurst = 100 // per second

// defaultQueueExpFailureBase 定义了失败指数基数的默认值，单位是毫秒
// defaultQueueExpFailureBase defines the default value of the failure exponential base, in milliseconds
const defaultQueueExpFailureBase = 100 // milliseconds

// defaultQueueExpFailureMax 定义了失败指数最大值的默认值，单位是秒
// defaultQueueExpFailureMax defines the default value of the maximum failure exponential, in seconds
const defaultQueueExpFailureMax = 500 // seconds

// emptyCallback 是一个空实现，它实现了 Callback 接口的所有方法，但这些方法的实现都是空的
// emptyCallback is an empty implementation that implements all methods of the Callback interface, but the implementations of these methods are empty
type emptyCallback struct{}

func (emptyCallback) OnDone(_ any)                      {} // OnDone 方法的空实现
func (emptyCallback) OnAdd(_ any)                       {} // OnAdd 方法的空实现
func (emptyCallback) OnGet(_ any)                       {} // OnGet 方法的空实现
func (emptyCallback) OnAddAfter(_ any, _ time.Duration) {} // OnAddAfter 方法的空实现
func (emptyCallback) OnAddWeight(_ any, _ int)          {} // OnAddWeight 方法的空实现
func (emptyCallback) OnAddLimited(_ any)                {} // OnAddLimited 方法的空实现
func (emptyCallback) OnForget(_ any)                    {} // OnForget 方法的空实现
func (emptyCallback) OnGetTimes(_ any, _ int)           {} // OnGetTimes 方法的空实现

// newEmptyCallback 函数创建并返回一个新的 emptyCallback 实例
// The newEmptyCallback function creates and returns a new instance of emptyCallback
func newEmptyCallback() *emptyCallback {
	return &emptyCallback{}
}
