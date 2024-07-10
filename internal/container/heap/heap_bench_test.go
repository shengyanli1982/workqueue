package heap

import (
	"testing"

	"container/heap"
)

type heapNodes struct {
	nodes []*Node
}

func (h *heapNodes) Len() int           { return len(h.nodes) }
func (h *heapNodes) Less(i, j int) bool { return h.nodes[i].priority < h.nodes[j].priority }
func (h *heapNodes) Swap(i, j int)      { h.nodes[i], h.nodes[j] = h.nodes[j], h.nodes[i] }

func (h *heapNodes) Push(x any) { h.nodes = append(h.nodes, x.(*Node)) }
func (h *heapNodes) Pop() any {
	n := h.nodes[len(h.nodes)-1]
	h.nodes = h.nodes[:len(h.nodes)-1]
	return n
}

func BenchmarkHeap_Push(b *testing.B) {
	h := New()
	nodes := make([]*Node, b.N)

	// Push the nodes
	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{priority: int64(b.N - i - 1)}
	}

	b.ResetTimer()

	// Push the nodes
	for i := 0; i < b.N; i++ {
		h.Push(nodes[i])
	}
}

func BenchmarkHeap_Pop(b *testing.B) {
	h := New()

	// Push the nodes
	for i := 0; i < b.N; i++ {
		h.Push(&Node{priority: int64(i)})
	}

	b.ResetTimer()

	// Pop the nodes
	for i := 0; i < b.N; i++ {
		h.Pop()
	}
}

func BenchmarkHeap_Remove(b *testing.B) {
	h := New()
	nodes := make([]*Node, b.N)

	// Push the nodes
	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{priority: int64(i)}
	}

	b.ResetTimer()

	// Remove the nodes
	for i := 0; i < b.N; i++ {
		h.Remove(nodes[i])
	}
}

func BenchmarkCompareGoStdHeap_Push(b *testing.B) {
	h := &heapNodes{}
	heap.Init(h)

	b.ResetTimer()

	// Push the nodes
	for i := 0; i < b.N; i++ {
		heap.Push(h, &Node{priority: int64(b.N - i - 1)})
	}
}

func BenchmarkCompareGoStdHeap_Pop(b *testing.B) {
	h := &heapNodes{}
	heap.Init(h)

	// Push the nodes
	for i := 0; i < b.N; i++ {
		heap.Push(h, &Node{priority: int64(i)})
	}

	b.ResetTimer()

	// Pop the nodes
	for i := 0; i < b.N; i++ {
		heap.Pop(h)
	}
}

func BenchmarkCompareWQHeap_Push(b *testing.B) {
	h := New()
	b.ResetTimer()

	// Push the nodes
	for i := 0; i < b.N; i++ {
		h.Push(&Node{priority: int64(b.N - i - 1)})
	}
}

func BenchmarkCompareWQHeap_Pop(b *testing.B) {
	h := New()

	// Push the nodes
	for i := 0; i < b.N; i++ {
		h.Push(&Node{priority: int64(i)})
	}

	b.ResetTimer()

	// Pop the nodes
	for i := 0; i < b.N; i++ {
		h.Pop()
	}
}
