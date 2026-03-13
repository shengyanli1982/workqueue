package list

import "testing"

func BenchmarkListPprof_MixedOps(b *testing.B) {
	const workingSet = 8192

	l := New()
	pool := NewNodePool()
	for i := 0; i < workingSet; i++ {
		node := pool.Get()
		node.Value = i
		l.PushBack(node)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n := l.PopFront()
		l.PushBack(n)

		if i&1 == 0 {
			l.MoveToFront(l.Back())
		}
		if i&3 == 0 {
			l.MoveToBack(l.Front())
		}
		if i&15 == 0 {
			front := l.Front()
			back := l.Back()
			if front != nil && back != nil && front != back {
				l.Swap(front, back)
			}
		}
	}

	b.StopTimer()
	for n := l.PopFront(); n != nil; n = l.PopFront() {
		pool.Put(n)
	}
}

func BenchmarkListPprof_AllocChurn(b *testing.B) {
	l := New()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n := NewNode()
		l.PushBack(n)
		_ = l.PopFront()
	}
}
