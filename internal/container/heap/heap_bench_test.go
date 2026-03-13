package heap

import (
	"testing"

	"container/heap"

	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

const stableHeapBatchSize = 8192

type heapNodes struct {
	nodes []*lst.Node
}

func (h *heapNodes) Len() int           { return len(h.nodes) }
func (h *heapNodes) Less(i, j int) bool { return h.nodes[i].Priority < h.nodes[j].Priority }
func (h *heapNodes) Swap(i, j int)      { h.nodes[i], h.nodes[j] = h.nodes[j], h.nodes[i] }

func (h *heapNodes) Push(x any) { h.nodes = append(h.nodes, x.(*lst.Node)) }
func (h *heapNodes) Pop() any {
	n := h.nodes[len(h.nodes)-1]
	h.nodes = h.nodes[:len(h.nodes)-1]
	return n
}

func BenchmarkHeap_Push(b *testing.B) {
	h := New()
	nodes := make([]*lst.Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &lst.Node{Priority: int64(b.N - i - 1)}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h.Push(nodes[i])
	}
}

func BenchmarkHeap_Pop(b *testing.B) {
	h := New()

	for i := 0; i < b.N; i++ {
		h.Push(&lst.Node{Priority: int64(i)})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h.Pop()
	}
}

func BenchmarkHeap_Remove(b *testing.B) {
	h := New()
	nodes := make([]*lst.Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &lst.Node{Priority: int64(i)}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h.Remove(nodes[i])
	}
}

func BenchmarkCompareGoStdHeap_Push(b *testing.B) {
	h := &heapNodes{}
	heap.Init(h)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		heap.Push(h, &lst.Node{Priority: int64(b.N - i - 1)})
	}
}

func BenchmarkCompareGoStdHeap_Pop(b *testing.B) {
	h := &heapNodes{}
	heap.Init(h)

	for i := 0; i < b.N; i++ {
		heap.Push(h, &lst.Node{Priority: int64(i)})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		heap.Pop(h)
	}
}

func BenchmarkCompareWQHeap_Push(b *testing.B) {
	h := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h.Push(&lst.Node{Priority: int64(b.N - i - 1)})
	}
}

func BenchmarkCompareWQHeap_Pop(b *testing.B) {
	h := New()

	for i := 0; i < b.N; i++ {
		h.Push(&lst.Node{Priority: int64(i)})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h.Pop()
	}
}

func BenchmarkHeap_PopBatchStable(b *testing.B) {
	b.ReportAllocs()
	ops := 0

	for ops < b.N {
		b.StopTimer()
		h := New()
		for i := 0; i < stableHeapBatchSize; i++ {
			h.Push(&lst.Node{Priority: int64(i)})
		}
		b.StartTimer()

		for i := 0; i < stableHeapBatchSize && ops < b.N; i++ {
			_ = h.Pop()
			ops++
		}
	}
}

func BenchmarkHeap_DeleteMinReinsertStable(b *testing.B) {
	b.ReportAllocs()
	h := New()
	for i := 0; i < stableHeapBatchSize; i++ {
		h.Push(&lst.Node{Priority: int64(i)})
	}

	nextPriority := int64(stableHeapBatchSize)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n := h.Pop()
		n.Priority = nextPriority
		nextPriority++
		h.Push(n)
	}
}
