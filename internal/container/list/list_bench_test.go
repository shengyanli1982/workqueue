package list

import (
	"container/list"
	"testing"
)

func BenchmarkList_PushBack(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PushBack(nodes[i])
	}
}

func BenchmarkList_PushFront(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PushFront(nodes[i])
	}
}

func BenchmarkList_PopBack(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
		l.PushBack(nodes[i])
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PopBack()
	}
}

func BenchmarkList_PopFront(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
		l.PushBack(nodes[i])
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PopFront()
	}
}

func BenchmarkList_InsertBefore(b *testing.B) {
	l := New()
	n := &Node{Value: int64(-1)}
	l.PushBack(n)
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.InsertBefore(nodes[i], n)
	}
}

func BenchmarkList_InsertAfter(b *testing.B) {
	l := New()
	n := &Node{Value: int64(-1)}
	l.PushBack(n)
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.InsertAfter(nodes[i], n)
	}
}

func BenchmarkList_Remove(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
		l.PushBack(nodes[i])
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Remove(nodes[i])
	}
}

func BenchmarkList_MoveToFront(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
		l.PushBack(nodes[i])
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.MoveToFront(nodes[i])
	}
}

func BenchmarkList_MoveToBack(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
		l.PushBack(nodes[i])
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.MoveToBack(nodes[i])
	}
}

func BenchmarkList_Swap(b *testing.B) {
	l := New()
	nodes := make([]*Node, b.N)

	for i := 0; i < b.N; i++ {
		nodes[i] = &Node{Value: int64(i)}
		l.PushBack(nodes[i])
	}

	b.ResetTimer()

	for i := 0; i < b.N-1; i++ {
		l.Swap(nodes[i], nodes[i+1])
	}
}

func BenchmarkCompareGoStdList_PushBack(b *testing.B) {
	l := list.New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PushBack(int64(i))
	}
}

func BenchmarkCompareGoStdList_PushFront(b *testing.B) {
	l := list.New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PushFront(int64(i))
	}
}

func BenchmarkCompareGoStdList_PopBack(b *testing.B) {
	l := list.New()

	for i := 0; i < b.N; i++ {
		l.PushBack(int64(i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Remove(l.Back())
	}
}

func BenchmarkCompareGoStdList_PopFront(b *testing.B) {
	l := list.New()

	for i := 0; i < b.N; i++ {
		l.PushBack(int64(i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Remove(l.Front())
	}
}

func BenchmarkCompareGoStdList_InsertBefore(b *testing.B) {
	l := list.New()
	e := l.PushBack(int64(-1))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.InsertBefore(int64(i), e)
	}
}

func BenchmarkCompareGoStdList_InsertAfter(b *testing.B) {
	l := list.New()
	e := l.PushBack(int64(-1))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.InsertAfter(int64(i), e)
	}
}

func BenchmarkCompareWQList_PushBack(b *testing.B) {
	l := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PushBack(&Node{Value: int64(i)})
	}
}

func BenchmarkCompareWQList_PushFront(b *testing.B) {
	l := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PushFront(&Node{Value: int64(i)})
	}
}

func BenchmarkCompareWQList_PopBack(b *testing.B) {
	l := New()

	for i := 0; i < b.N; i++ {
		l.PushBack(&Node{Value: int64(i)})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PopBack()
	}
}

func BenchmarkCompareWQList_PopFront(b *testing.B) {
	l := New()

	for i := 0; i < b.N; i++ {
		l.PushBack(&Node{Value: int64(i)})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PopFront()
	}
}

func BenchmarkCompareWQList_InsertBefore(b *testing.B) {
	l := New()
	n := &Node{Value: int64(-1)}
	l.PushBack(n)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.InsertBefore(&Node{Value: int64(i)}, n)
	}
}

func BenchmarkCompareWQList_InsertAfter(b *testing.B) {
	l := New()
	n := &Node{Value: int64(-1)}
	l.PushBack(n)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.InsertAfter(&Node{Value: int64(i)}, n)
	}
}
