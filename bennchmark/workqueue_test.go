package bennchmark

import (
	"sync"
	"testing"

	wq "github.com/shengyanli1982/workqueue"
)

func BenchmarkWorkqueueAdd(b *testing.B) {
	q := wq.NewQueue(nil)
	defer q.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Add(i)
	}
}

func BenchmarkWorkqueueGet(b *testing.B) {
	q := wq.NewQueue(nil)
	defer q.Stop()

	for i := 0; i < b.N; i++ {
		_ = q.Add(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = q.Get()
		q.Done(i)
	}
}

func BenchmarkWorkqueueAddAndGet(b *testing.B) {
	q := wq.NewQueue(nil)
	defer q.Stop()
	wg := sync.WaitGroup{}
	wg.Add(2)
	b.ResetTimer()

	go func() {
		for i := 0; i < b.N; i++ {
			_ = q.Add(i)
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < b.N; i++ {
			if ele, err := q.Get(); err == nil {
				q.Done(ele)
			}
		}
		wg.Done()
	}()

	wg.Wait()
}
