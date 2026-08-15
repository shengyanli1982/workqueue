package heap

import (
	"fmt"
	"math/rand"
	"testing"

	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
	"github.com/stretchr/testify/assert"
)

func PrintRootIndexs(h *RBTree) {
	fmt.Printf("# root: %v\n", h.root)
}

func PrintOrderTraversalIndexs(n *lst.Node) {
	if n != nil {
		PrintOrderTraversalIndexs(n.Left)
		fmt.Printf(">> priority: %d, value: %v, left: %v, right: %v\n", n.Priority, n.Value, n.Left, n.Right)
		PrintOrderTraversalIndexs(n.Right)
	}
}

func PrintNodeIndexs(nodes []*lst.Node) {
	for _, n := range nodes {
		fmt.Printf("# priority: %v\n", n.Priority)
	}
}

func TestHeap_Push(t *testing.T) {
	h := New()
	count := 4

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(i), Value: i})
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front priority should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))

	assert.Equal(t, int64(0), h.Front().Priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(1), h.Root().Priority, "root priority should be 1")
	assert.Equal(t, int64(0), h.Root().Left.Priority, "root left priority should be 0")
	assert.Equal(t, int64(2), h.Root().Right.Priority, "root right priority should be 2")
	assert.Equal(t, int64(3), h.Root().Right.Right.Priority, "root right right priority should be 3")
}

func TestHeap_Push_Reverse(t *testing.T) {
	h := New()
	count := 4

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(count - i - 1)})
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front priority should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))

	assert.Equal(t, int64(0), h.Front().Priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(2), h.Root().Priority, "root priority should be 2")
	assert.Equal(t, int64(1), h.Root().Left.Priority, "root left priority should be 1")
	assert.Equal(t, int64(0), h.Root().Left.Left.Priority, "root left left priority should be 0")
	assert.Equal(t, int64(3), h.Root().Right.Priority, "root right priority should be 0")
}

func TestHeap_Push_Random(t *testing.T) {
	h := New()
	count := 4

	h.Push(&lst.Node{Priority: int64(2)})
	h.Push(&lst.Node{Priority: int64(0)})
	h.Push(&lst.Node{Priority: int64(1)})
	h.Push(&lst.Node{Priority: int64(3)})

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	assert.Equal(t, int64(0), h.Front().Priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(1), h.Root().Priority, "root priority should be 1")
	assert.Equal(t, int64(0), h.Root().Left.Priority, "root left priority should be 0")
	assert.Equal(t, int64(2), h.Root().Right.Priority, "root right priority should be 2")
	assert.Equal(t, int64(3), h.Root().Right.Right.Priority, "root right right priority should be 3")
}

func TestHeap_Push_Duplicate(t *testing.T) {
	h := New()
	count := 4

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(count - i - 1)})
	}

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(i)})
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count*2), h.Len(), fmt.Sprintf("heap length should be %d", count*2))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	assert.Equal(t, int64(0), h.Front().Priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(2), h.Root().Priority, "root priority should be 2")
	assert.Equal(t, int64(0), h.Root().Left.Priority, "root left priority should be 0")
	assert.Equal(t, int64(3), h.Root().Right.Priority, "root right priority should be 3")
	assert.Equal(t, int64(0), h.Root().Left.Left.Priority, "root left left priority should be 0")
	assert.Equal(t, int64(1), h.Root().Left.Right.Priority, "root left right priority should be 1")
	assert.Equal(t, int64(2), h.Root().Right.Left.Priority, "root right left priority should be 2")
	assert.Equal(t, int64(3), h.Root().Right.Right.Priority, "root right right priority should be 3")
	assert.Equal(t, int64(1), h.Root().Left.Right.Right.Priority, "root left right priority should be 1")
}

func TestHeap_Push_Nil(t *testing.T) {
	h := New()

	for i := 0; i < 10; i++ {
		h.Push(nil)
	}

	assert.Equal(t, int64(0), h.Len(), "heap length should be 0")
	assert.Nil(t, h.Front(), "front value should be nil")
	assert.Nil(t, h.Back(), "back value should be nil")
}

func TestHeap_Pop(t *testing.T) {
	h := New()
	count := 4

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(count - i - 1)})
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	fristNode := h.Front()

	popNode := h.Pop()

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.NotNil(t, fristNode, "first node should not be nil")
	assert.NotNil(t, popNode, "pop node should not be nil")
	assert.Equal(t, fristNode, popNode, "first node should be equal to pop node")
}

func TestHeap_PopAll(t *testing.T) {
	h := New()
	count := 4

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(count - i - 1)})
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	for i := 0; i < count; i++ {
		n := h.Pop()
		assert.NotNil(t, n, "pop value should not be nil")
		assert.Equal(t, int64(i), n.Priority, fmt.Sprintf("pop value should be %d", i))
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(0), h.Len(), "heap length should be 0")
	assert.Nil(t, h.Front(), "front value should be nil")
	assert.Nil(t, h.Back(), "back value should be nil")
}

func TestHeap_PopEmpty(t *testing.T) {
	h := New()

	n := h.Pop()
	assert.Nil(t, n, "pop value should be nil")
}

func TestHeap_PutAndPop_Intersect(t *testing.T) {
	h := New()
	count := 4

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(count - i - 1)})
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	assert.Equal(t, int64(0), h.Front().Priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(2), h.Root().Priority, "root priority should be 2")
	assert.Equal(t, int64(1), h.Root().Left.Priority, "root left priority should be 1")
	assert.Equal(t, int64(0), h.Root().Left.Left.Priority, "root left left priority should be 0")
	assert.Equal(t, int64(3), h.Root().Right.Priority, "root right priority should be 0")

	n := h.Pop()
	assert.NotNil(t, n, "pop value should not be nil")
	assert.Equal(t, int64(0), n.Priority, "pop value should be 0")

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count-1), h.Len(), fmt.Sprintf("heap length should be %d", count-1))
	assert.Equal(t, int64(1), h.Front().Priority, fmt.Sprintf("front value should be %d", 1))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	h.Push(&lst.Node{Priority: int64(0)})

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	assert.Equal(t, int64(0), h.Root().Left.Left.Priority, "root left left priority should be 0")

	h.Push(&lst.Node{Priority: int64(2)})

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count+1), h.Len(), fmt.Sprintf("heap length should be %d", count+1))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	assert.Equal(t, int64(2), h.Root().Right.Left.Priority, "root right left priority should be 2")

	h.Push(&lst.Node{Priority: int64(count)})

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count+2), h.Len(), fmt.Sprintf("heap length should be %d", count+2))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count), h.Back().Priority, fmt.Sprintf("back value should be %d", count))

	assert.Equal(t, int64(count), h.Root().Right.Right.Priority, "root right right priority should be 4")
}

func TestHeap_Remove(t *testing.T) {
	h := New()
	count := 4
	nodes := make([]*lst.Node, count)

	for i := 0; i < count; i++ {
		n := &lst.Node{Priority: int64(count - i - 1)}
		nodes[i] = n
		h.Push(n)
	}

	PrintRootIndexs(h)
	PrintOrderTraversalIndexs(h.root)

	assert.Equal(t, int64(count), h.Len(), fmt.Sprintf("heap length should be %d", count))
	assert.Equal(t, int64(0), h.Front().Priority, fmt.Sprintf("front value should be %d", 0))
	assert.Equal(t, int64(count-1), h.Back().Priority, fmt.Sprintf("back value should be %d", count-1))

	assert.Equal(t, int64(0), h.Front().Priority, "front priority should be 0")
	assert.Equal(t, int64(3), h.Back().Priority, fmt.Sprintf("back priority should be %d", count-1))
	assert.Equal(t, int64(2), h.Root().Priority, "root priority should be 2")
	assert.Equal(t, int64(1), h.Root().Left.Priority, "root left priority should be 1")
	assert.Equal(t, int64(0), h.Root().Left.Left.Priority, "root left left priority should be 0")
	assert.Equal(t, int64(3), h.Root().Right.Priority, "root right priority should be 0")

	for i := 0; i < count; i++ {
		h.Remove(nodes[i])
	}

	assert.Equal(t, int64(0), h.Len(), "heap length should be 0")
	assert.Nil(t, h.Front(), "front value should be nil")
	assert.Nil(t, h.Back(), "back value should be nil")
}

func TestHeap_ExtremeValues(t *testing.T) {
	h := New()

	h.Push(&lst.Node{Priority: int64(9223372036854775807)})
	h.Push(&lst.Node{Priority: int64(-9223372036854775808)})
	h.Push(&lst.Node{Priority: 0})

	assert.Equal(t, int64(-9223372036854775808), h.Front().Priority, "front should be MinInt64")
	assert.Equal(t, int64(9223372036854775807), h.Back().Priority, "back should be MaxInt64")
	assert.Equal(t, int64(3), h.Len(), "heap should contain 3 elements")
}

func TestHeap_NegativePriorities(t *testing.T) {
	h := New()

	priorities := []int64{-1, -5, -3, -2, -4}
	for _, p := range priorities {
		h.Push(&lst.Node{Priority: p})
	}

	assert.Equal(t, int64(-5), h.Front().Priority, "front should be -5")
	assert.Equal(t, int64(-1), h.Back().Priority, "back should be -1")

	expected := []int64{-5, -4, -3, -2, -1}
	for _, exp := range expected {
		node := h.Pop()
		assert.Equal(t, exp, node.Priority, fmt.Sprintf("expected priority %d", exp))
	}
}

func TestHeap_DuplicatePriorities(t *testing.T) {
	h := New()

	count := 5
	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: 1, Value: i})
	}

	assert.Equal(t, int64(count), h.Len(), "heap should contain 5 elements")
	assert.Equal(t, int64(1), h.Front().Priority, "front priority should be 1")
	assert.Equal(t, int64(1), h.Back().Priority, "back priority should be 1")

	for i := 0; i < count; i++ {
		node := h.Pop()
		assert.Equal(t, int64(1), node.Priority, "all nodes should have priority 1")
	}

	assert.Equal(t, int64(0), h.Len(), "heap should be empty")
}

func TestHeap_RemoveNilAndInvalid(t *testing.T) {
	h := New()

	h.Remove(nil)
	assert.Equal(t, int64(0), h.Len(), "heap should be empty after removing nil")

	validNode := &lst.Node{Priority: 1}
	h.Push(validNode)
	assert.Equal(t, int64(1), h.Len(), "heap should contain one node")

	nonExistentNode := &lst.Node{Priority: 999}
	h.Remove(nonExistentNode)
	assert.Equal(t, int64(1), h.Len(), "heap should still contain the valid node")

	h.Remove(validNode)
	assert.Equal(t, int64(0), h.Len(), "heap should be empty after removing valid node")
}

// TestHeap_TailCorrectness_EqualPriorities 同优先级节点沿右链插入，
// Back() 必须始终指向最后一个插入的同优先级节点（tail 更新需取等）。
// 修复前：tail 比较使用严格大于，等优先级插入不更新 tail，Back() 滞留失效节点。
func TestHeap_TailCorrectness_EqualPriorities(t *testing.T) {
	h := New()

	nodes := make([]*lst.Node, 5)
	for i := 0; i < 5; i++ {
		nodes[i] = &lst.Node{Priority: 1, Value: i}
		h.Push(nodes[i])
		assert.Equal(t, nodes[i], h.Back(), "back must be the latest inserted equal-priority node")
		assert.Equal(t, nodes[0], h.Front(), "front must stay the earliest inserted equal-priority node")
	}

	// 全部弹出后 Front/Back 归 nil。
	for i := 0; i < 5; i++ {
		n := h.Pop()
		assert.Equal(t, nodes[i], n, "equal-priority nodes must pop in insertion order")
	}
	assert.Nil(t, h.Front())
	assert.Nil(t, h.Back())
}

// TestHeap_Remove_TwoChildren_UnlinksGivenNode 验证 Remove 的双子节点路径
// 物理解链的是调用者给定的节点本身（而非后继节点）：调用方（如队列的
// Cancel/CancelDelay）在 Remove 后会把该节点归还节点池并 Reset 其指针，
// 若该节点仍留在树中，Reset 将破坏树结构。
func TestHeap_Remove_TwoChildren_UnlinksGivenNode(t *testing.T) {
	for round := 0; round < 100; round++ {
		h := New()
		// 构造 root 带双子节点的稳定形状：2 为根，1/3 为左右子，4 为最右。
		root := &lst.Node{Priority: 2, Value: "root"}
		h.Push(root)
		h.Push(&lst.Node{Priority: 1, Value: "left"})
		h.Push(&lst.Node{Priority: 3, Value: "right"})
		h.Push(&lst.Node{Priority: 4, Value: "rightmost"})

		h.Remove(root)
		// 模拟归还节点池：Reset 归还不该影响树结构。
		root.Reset()

		assert.Equal(t, int64(3), h.Len())
		assert.Equal(t, int64(1), h.Front().Priority)
		assert.Equal(t, int64(4), h.Back().Priority)

		// 剩余节点仍可按序全部弹出。
		expected := []int64{1, 3, 4}
		for _, p := range expected {
			n := h.Pop()
			assert.NotNil(t, n)
			assert.Equal(t, p, n.Priority)
		}
		assert.Equal(t, int64(0), h.Len())
	}
}

// heapValidate 校验红黑树结构不变式：父链一致、根为黑、无连续红节点、
// 黑高一致、计数正确、Front/Back 与中序首尾一致。
func heapValidate(t *testing.T, tree *RBTree, op string) {
	t.Helper()

	root := tree.Root()
	if root == nil {
		if tree.Len() != 0 {
			t.Fatalf("[%s] nil root but Len=%d", op, tree.Len())
		}
		if tree.Front() != nil || tree.Back() != nil {
			t.Fatalf("[%s] empty tree with non-nil head/tail", op)
		}
		return
	}
	if root.Parent != nil {
		t.Fatalf("[%s] root.Parent != nil", op)
	}
	if root.Color != lst.BLACK {
		t.Fatalf("[%s] root is RED", op)
	}

	count := int64(0)
	blackHeight := -1
	var first, last *lst.Node
	var walk func(n *lst.Node, parent *lst.Node, blacks int)
	walk = func(n *lst.Node, parent *lst.Node, blacks int) {
		if n == nil {
			if blackHeight == -1 {
				blackHeight = blacks
			} else if blackHeight != blacks {
				t.Fatalf("[%s] black height violation: %d vs %d", op, blackHeight, blacks)
			}
			return
		}
		if n.Parent != parent {
			t.Fatalf("[%s] broken parent link at node prio=%d value=%v", op, n.Priority, n.Value)
		}
		if n.Color == lst.RED && n.Parent != nil && n.Parent.Color == lst.RED {
			t.Fatalf("[%s] consecutive RED nodes at prio=%d", op, n.Priority)
		}
		next := blacks
		if n.Color == lst.BLACK {
			next++
		}
		walk(n.Left, n, next)
		// 中序位置记录首尾节点与计数。
		count++
		if first == nil {
			first = n
		}
		last = n
		walk(n.Right, n, next)
	}
	walk(root, nil, 0)

	if count != tree.Len() {
		t.Fatalf("[%s] walk count %d != Len %d", op, count, tree.Len())
	}
	if tree.Front() != first {
		t.Fatalf("[%s] head mismatch: got %p want %p", op, tree.Front(), first)
	}
	if tree.Back() != last {
		t.Fatalf("[%s] tail mismatch: got %p want %p", op, tree.Back(), last)
	}
}

// TestHeap_FuzzOps 以固定种子回放 Push/Remove/Pop 随机序列并逐步校验全部
// 结构不变式；Remove/Pop 后对节点执行 Reset 以模拟队列归还节点池的行为。
// 回归覆盖：deleteFixUp 空子节点修复、双子节点物理换位解链、同优先级 tail 更新。
func TestHeap_FuzzOps(t *testing.T) {
	for seed := int64(0); seed < 200; seed++ {
		rng := rand.New(rand.NewSource(seed))
		tree := New()
		inTree := map[int]*lst.Node{}

		for step := 0; step < 1500; step++ {
			key := rng.Intn(60)
			prio := int64(rng.Intn(5)) // 高碰撞优先级，模拟毫秒时间戳撞车
			ctx := fmt.Sprintf("seed=%d step=%d", seed, step)

			switch rng.Intn(3) {
			case 0: // push（同 key 不重复入树）
				if _, ok := inTree[key]; !ok {
					n := &lst.Node{Value: key, Priority: prio}
					tree.Push(n)
					inTree[key] = n
					heapValidate(t, tree, ctx+" push")
				}
			case 1: // remove + Reset（模拟 Cancel 后归还节点池）
				if n, ok := inTree[key]; ok {
					tree.Remove(n)
					n.Reset()
					delete(inTree, key)
					heapValidate(t, tree, ctx+" remove")
				}
			case 2: // pop + Reset（模拟到期搬运后归还节点池）
				if m := tree.Pop(); m != nil {
					if v, ok := m.Value.(int); ok {
						delete(inTree, v)
					}
					m.Reset()
					heapValidate(t, tree, ctx+" pop")
				}
			}
		}
	}
}

func TestHeap_LargeDataSet(t *testing.T) {
	h := New()
	count := 1000

	for i := 0; i < count; i++ {
		h.Push(&lst.Node{Priority: int64(i)})
	}

	assert.Equal(t, int64(count), h.Len(), "heap should contain 1000 elements")
	assert.Equal(t, int64(0), h.Front().Priority, "front should have minimum priority")
	assert.Equal(t, int64(count-1), h.Back().Priority, "back should have maximum priority")

	prev := h.Pop()
	for i := 1; i < count; i++ {
		current := h.Pop()
		assert.True(t, prev.Priority <= current.Priority, "heap order property should be maintained")
		prev = current
	}
}
