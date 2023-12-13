package bennchmark

import (
	"sync"
	"testing"

	wq "k8s.io/client-go/util/workqueue"
)

func BenchmarkClientgoAdd(b *testing.B) {
	q := wq.New()
	defer q.ShutDown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Add(i)
	}
}

func BenchmarkClientgoGet(b *testing.B) {
	q := wq.New()
	defer q.ShutDown()

	for i := 0; i < b.N; i++ {
		q.Add(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = q.Get()
		q.Done(i)
	}
}

func BenchmarkClientgoAddAndGet(b *testing.B) {
	q := wq.New()
	defer q.ShutDown()
	wg := sync.WaitGroup{}
	wg.Add(2)
	b.ResetTimer()

	go func() {
		for i := 0; i < b.N; i++ {
			q.Add(i)
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < b.N; i++ {
			if ele, shutdown := q.Get(); !shutdown {
				q.Done(ele)
			}
		}
		wg.Done()
	}()

	wg.Wait()
}
