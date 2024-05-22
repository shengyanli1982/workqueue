package list

import "testing"

func BenchmarkNodePool_Get(b *testing.B) {
	pool := NewNodePool()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pool.Get()
	}
}

func BenchmarkNodePool_Put(b *testing.B) {
	pool := NewNodePool()
	node := NewNode()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pool.Put(node)
	}
}

func BenchmarkNodePool_PutAndGet(b *testing.B) {
	pool := NewNodePool()
	node := pool.Get()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pool.Put(node)
		node = pool.Get()
	}
}
