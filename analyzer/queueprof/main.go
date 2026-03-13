package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"

	wq "github.com/shengyanli1982/workqueue/v2"
)

type immediateRetryPolicy struct{}

func (p *immediateRetryPolicy) NextDelay(interface{}, int, error) (time.Duration, bool) {
	return 0, true
}

func main() {
	var (
		queueType  = flag.String("queue", "all", "queue type: deadletter|retry|leased|bounded|timer|all")
		workers    = flag.Int("workers", runtime.GOMAXPROCS(0), "worker count")
		duration   = flag.Duration("duration", 10*time.Second, "stress duration")
		cpuProfile = flag.String("cpuprofile", "", "cpu profile output file")
		memProfile = flag.String("memprofile", "", "heap profile output file")
	)
	flag.Parse()

	if *workers <= 0 {
		*workers = 1
	}
	if *duration <= 0 {
		*duration = 10 * time.Second
	}

	stopCPU, err := startCPUProfile(*cpuProfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start cpu profile failed: %v\n", err)
		os.Exit(1)
	}
	if stopCPU != nil {
		defer stopCPU()
	}

	runners := map[string]func(context.Context, int) uint64{
		"deadletter": runDeadLetter,
		"retry":      runRetry,
		"leased":     runLeased,
		"bounded":    runBoundedBlocking,
		"timer":      runTimer,
	}

	if *queueType == "all" {
		var total uint64
		for name, fn := range runners {
			ctx, cancel := context.WithTimeout(context.Background(), *duration)
			ops := fn(ctx, *workers)
			cancel()
			total += ops
			fmt.Printf("[%s] ops=%d\n", name, ops)
		}
		fmt.Printf("total_ops=%d\n", total)
	} else {
		fn, ok := runners[*queueType]
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown queue type: %s\n", *queueType)
			os.Exit(2)
		}
		ctx, cancel := context.WithTimeout(context.Background(), *duration)
		ops := fn(ctx, *workers)
		cancel()
		fmt.Printf("[%s] ops=%d\n", *queueType, ops)
	}

	if err = writeMemProfile(*memProfile); err != nil {
		fmt.Fprintf(os.Stderr, "write mem profile failed: %v\n", err)
		os.Exit(1)
	}
}

func startCPUProfile(path string) (func(), error) {
	if path == "" {
		return nil, nil
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if err = pprof.StartCPUProfile(file); err != nil {
		_ = file.Close()
		return nil, err
	}
	return func() {
		pprof.StopCPUProfile()
		_ = file.Close()
	}, nil
}

func writeMemProfile(path string) error {
	if path == "" {
		return nil
	}
	runtime.GC()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return pprof.WriteHeapProfile(file)
}

func runDeadLetter(ctx context.Context, workers int) uint64 {
	queue := wq.NewDeadLetterQueue(nil)
	defer queue.Shutdown()
	var ops atomic.Uint64

	execWorkers(ctx, workers, func(seq int) {
		letter := &wq.DeadLetter{Payload: seq}
		if err := queue.PutDead(letter); err != nil {
			return
		}
		got, err := queue.GetDead()
		if err != nil {
			return
		}
		if err = queue.AckDead(got); err != nil {
			return
		}
		ops.Add(1)
	})

	return ops.Load()
}

func runRetry(ctx context.Context, workers int) uint64 {
	cfg := wq.NewRetryQueueConfig().
		WithPolicy(&immediateRetryPolicy{}).
		WithKeyFunc(func(interface{}) string { return "retry" })
	queue := wq.NewRetryQueue(cfg)
	defer queue.Shutdown()
	var ops atomic.Uint64

	benchErr := errors.New("bench")
	execWorkers(ctx, workers, func(seq int) {
		if err := queue.Put(seq); err != nil {
			return
		}
		value, err := queue.Get()
		if err != nil {
			return
		}
		if err = queue.Retry(value, benchErr); err != nil {
			return
		}
		retried, err := queue.Get()
		if err != nil {
			return
		}
		queue.Done(retried)
		queue.Forget(retried)
		ops.Add(1)
	})

	return ops.Load()
}

func runLeased(ctx context.Context, workers int) uint64 {
	cfg := wq.NewLeasedQueueConfig().
		WithLeaseDuration(time.Second).
		WithScanInterval(time.Hour)
	queue := wq.NewLeasedQueue(cfg)
	defer queue.Shutdown()
	var ops atomic.Uint64

	execWorkers(ctx, workers, func(seq int) {
		if err := queue.Put(seq); err != nil {
			return
		}
		_, leaseID, err := queue.GetWithLease(time.Second)
		if err != nil {
			return
		}
		if err = queue.Ack(leaseID); err != nil {
			return
		}
		ops.Add(1)
	})

	return ops.Load()
}

func runBoundedBlocking(ctx context.Context, workers int) uint64 {
	cfg := wq.NewBoundedBlockingQueueConfig().WithCapacity(4096)
	queue := wq.NewBoundedBlockingQueue(cfg)
	defer queue.Shutdown()
	var ops atomic.Uint64

	execWorkers(ctx, workers, func(seq int) {
		if err := queue.Put(seq); err != nil {
			return
		}
		value, err := queue.Get()
		if err != nil {
			return
		}
		queue.Done(value)
		ops.Add(1)
	})

	return ops.Load()
}

func runTimer(ctx context.Context, workers int) uint64 {
	queue := wq.NewTimerQueue(nil)
	defer queue.Shutdown()
	var ops atomic.Uint64

	execWorkers(ctx, workers, func(seq int) {
		if err := queue.PutAt(seq, time.Now()); err != nil {
			return
		}
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			value, err := queue.Get()
			if err == nil {
				queue.Done(value)
				ops.Add(1)
				return
			}
			if errors.Is(err, wq.ErrQueueIsEmpty) {
				runtime.Gosched()
				continue
			}
			return
		}
	})

	return ops.Load()
}

func execWorkers(ctx context.Context, workers int, fn func(seq int)) {
	var (
		wg  sync.WaitGroup
		seq atomic.Uint64
	)

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				n := seq.Add(1)
				fn(int(n))
			}
		}()
	}
	wg.Wait()
}
