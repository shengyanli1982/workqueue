package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	wkq "github.com/shengyanli1982/workqueue/v2"
)

func main() {
	q := wkq.NewQueue(
		wkq.NewQueueConfig().WithValueIdempotent(),
	)
	defer q.Shutdown()

	_ = q.Put("job-1")
	if err := q.Put("job-1"); errors.Is(err, wkq.ErrElementAlreadyExist) {
		fmt.Println("dedup works: duplicate job ignored")
	}

	value, err := q.Get()
	if err != nil {
		fmt.Println("get failed:", err)
		return
	}
	fmt.Println("consumed:", value)
	q.Done(value)

	_, err = q.Get()
	if errors.Is(err, wkq.ErrQueueIsEmpty) {
		fmt.Println("queue is empty now")
	}

	demoIsDraining()
}

// demoIsDraining 演示 DrainingQueue 可选接口的三态语义：
//   - drain 进行中：IsDraining=true ∧ IsClosed=false，Put 被拒绝；
//   - drain 完成后：IsDraining 与 IsClosed 均为 true；
//   - Shutdown() 立即关闭：IsDraining 恒为 false。
//
// 时序由 channel 同步保证确定性，不依赖 sleep。
func demoIsDraining() {
	q := wkq.NewQueue(wkq.NewQueueConfig().WithValueIdempotent())

	// DrainingQueue 为可选接口，调用方经类型断言使用。
	dq, ok := q.(wkq.DrainingQueue)
	if !ok {
		fmt.Println("queue does not implement DrainingQueue")
		return
	}
	// 阻塞消费与优雅关停均为可选接口，标准队列实现均支持。
	bq := q.(wkq.BlockingGetQueue)
	drainable := q.(wkq.DrainableQueue)

	seedTaken := make(chan struct{})
	release := make(chan struct{})
	workerDone := make(chan struct{})

	// worker：消费种子项后阻塞等待 release，释放后继续消费助力 drain 完成。
	go func() {
		defer close(workerDone)
		ctx := context.Background()

		value, err := bq.GetWithContext(ctx)
		if err != nil {
			return
		}
		q.Done(value)
		close(seedTaken)

		<-release

		for {
			value, err := bq.GetWithContext(ctx)
			if err != nil {
				return
			}
			q.Done(value)
		}
	}()

	if err := q.Put("drain-seed"); err != nil {
		fmt.Println("put seed failed:", err)
		return
	}
	// 等待 worker 消费种子项，此时队列与处理中集合均为空。
	<-seedTaken

	// 待处理项须在 ShutdownWithDrain 之前入队，drain 才不会立即完成。
	if err := q.Put("drain-pending"); err != nil {
		fmt.Println("put pending failed:", err)
		return
	}

	drainCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	drainDone := make(chan error, 1)
	go func() {
		drainDone <- drainable.ShutdownWithDrain(drainCtx)
	}()

	// 轮询等待进入 drain 窗口，带超时上限。
	deadline := time.Now().Add(2 * time.Second)
	for !(dq.IsDraining() && !q.IsClosed()) && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if !dq.IsDraining() || q.IsClosed() {
		fmt.Println("timeout waiting for drain window")
		close(release)
		cancel()
		<-drainDone
		return
	}

	fmt.Printf("during drain: IsDraining=%v, IsClosed=%v\n", dq.IsDraining(), q.IsClosed())
	if err := q.Put("drain-rejected"); errors.Is(err, wkq.ErrQueueIsClosed) {
		fmt.Println("put during drain rejected with ErrQueueIsClosed")
	}

	// 释放 worker 消费待处理项，队列清空后 drain 完成并关闭。
	close(release)
	if err := <-drainDone; err != nil {
		fmt.Println("drain failed:", err)
	} else {
		fmt.Println("drain completed without error")
	}
	<-workerDone
	fmt.Printf("after drain: IsDraining=%v, IsClosed=%v\n", dq.IsDraining(), q.IsClosed())

	// Shutdown() 立即关闭：IsDraining 恒为 false。
	immediate := wkq.NewQueue(wkq.NewQueueConfig())
	idq, ok := immediate.(wkq.DrainingQueue)
	if !ok {
		fmt.Println("queue does not implement DrainingQueue")
		return
	}
	immediate.Shutdown()
	fmt.Printf("after immediate Shutdown: IsDraining=%v, IsClosed=%v\n", idq.IsDraining(), immediate.IsClosed())
}
