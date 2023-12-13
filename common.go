package workqueue

import "time"

// Queue 容量
// Queue capacity
const defaultQueueCap = 2048

// Queue 排序窗口
// Queue sort window
const defaultQueueSortWin = 500 // ms

// Queue 限速设置
// Queue rate limit setting
const defaultQueueRateLimit = 100 // per second
const defaultQueueRateBurst = 100 // per second
// const defaultQueueExpFailureBase = 100 // milliseconds
// const defaultQueueExpFailureMax = 500  // seconds

// 空实现
// empty implementation
type emptyCallback struct{}

func (emptyCallback) OnDone(_ any)                   {}
func (emptyCallback) OnAdd(_ any)                    {}
func (emptyCallback) OnGet(_ any)                    {}
func (emptyCallback) OnAfter(_ any, _ time.Duration) {}
func (emptyCallback) OnWeight(_ any, _ int)          {}
func (emptyCallback) OnLimited(_ any)                {}
func (emptyCallback) OnForget(_ any)                 {}
func (emptyCallback) OnGetTimes(_ any, _ int)        {}
