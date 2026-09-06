package heap

import (
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// RBTree 是按 Priority 排序的红黑树实现，用作优先级堆的底层数据结构。
// 树中节点按 Priority 升序排列，head 指向最小优先级节点，tail 指向最大优先级节点，
// 支持 O(log n) 的插入与删除，以及 O(1) 的最小/最大节点访问。
type RBTree struct {
	count int64
	root  *lst.Node
	head  *lst.Node
	tail  *lst.Node
}

// New 创建一个空的红黑树。
func New() *RBTree { return &RBTree{} }

// leftRotate 对节点执行左旋操作：将节点的右子节点提升为该位置的新根，
// 原节点下沉为其左子节点。左旋用于红黑树插入和删除后的平衡修复。
func leftRotate(tree *RBTree, node *lst.Node) {
	if node == nil || node.Right == nil {
		return
	}

	rightChild := node.Right

	node.Right = rightChild.Left
	if rightChild.Left != nil {
		rightChild.Left.Parent = node
	}

	rightChild.Parent = node.Parent
	if node.Parent == nil {

		tree.root = rightChild
	} else {

		if node == node.Parent.Left {
			node.Parent.Left = rightChild
		} else {
			node.Parent.Right = rightChild
		}
	}

	rightChild.Left = node
	node.Parent = rightChild
}

// rightRotate 对节点执行右旋操作：将节点的左子节点提升为该位置的新根，
// 原节点下沉为其右子节点。右旋用于红黑树插入和删除后的平衡修复。
func rightRotate(tree *RBTree, node *lst.Node) {
	if node == nil || node.Left == nil {
		return
	}
	leftChild := node.Left
	node.Left = leftChild.Right
	if leftChild.Right != nil {
		leftChild.Right.Parent = node
	}
	leftChild.Parent = node.Parent
	if node.Parent == nil {
		tree.root = leftChild
	} else {
		if node == node.Parent.Right {
			node.Parent.Right = leftChild
		} else {
			node.Parent.Left = leftChild
		}
	}
	leftChild.Right = node
	node.Parent = leftChild
}

// insertFixUp 在 BST 插入新节点后修复红黑树性质。新插入的节点初始为红色，
// 当其父节点也是红色时违反"红节点不能有红子节点"性质，需通过叔叔节点着色
// 和旋转操作逐层向上修复，最终确保根节点为黑色。
func insertFixUp(tree *RBTree, node *lst.Node) {
	for node.Parent != nil && node.Parent.Color == lst.RED {
		if node.Parent == node.Parent.Parent.Left {
			uncle := node.Parent.Parent.Right
			if uncle != nil && uncle.Color == lst.RED {
				node.Parent.Color = lst.BLACK
				uncle.Color = lst.BLACK
				node.Parent.Parent.Color = lst.RED
				node = node.Parent.Parent
			} else {
				if node == node.Parent.Right {
					node = node.Parent
					leftRotate(tree, node)
				}
				node.Parent.Color = lst.BLACK
				node.Parent.Parent.Color = lst.RED
				rightRotate(tree, node.Parent.Parent)
			}
		} else {
			uncle := node.Parent.Parent.Left
			if uncle != nil && uncle.Color == lst.RED {
				node.Parent.Color = lst.BLACK
				uncle.Color = lst.BLACK
				node.Parent.Parent.Color = lst.RED
				node = node.Parent.Parent
			} else {
				if node == node.Parent.Left {
					node = node.Parent
					rightRotate(tree, node)
				}
				node.Parent.Color = lst.BLACK
				node.Parent.Parent.Color = lst.RED
				leftRotate(tree, node.Parent.Parent)
			}
		}
	}
	tree.root.Color = lst.BLACK
}

// insert 将节点插入红黑树。流程：BST 定位插入位置 → 着色为红 → insertFixUp 修复红黑性质 →
// 更新 head（最小优先级）和 tail（最大优先级）指针。同优先级节点沿右链插入，新节点成为最右。
func (tree *RBTree) insert(node *lst.Node) {
	if node == nil {
		return
	}

	var parent *lst.Node
	current := tree.root

	for current != nil {
		parent = current
		if node.Priority < current.Priority {
			current = current.Left
		} else {
			current = current.Right
		}
	}

	node.Parent = parent
	if parent == nil {

		tree.root = node
	} else {

		if node.Priority < parent.Priority {
			parent.Left = node
		} else {
			parent.Right = node
		}
	}

	node.Left = nil
	node.Right = nil
	node.Color = lst.RED

	insertFixUp(tree, node)
	tree.count++

	if tree.head == nil || node.Priority < tree.head.Priority {
		tree.head = node
	}
	// 同优先级节点沿右链插入，新节点成为最右节点，tail 更新必须取等。
	if tree.tail == nil || node.Priority >= tree.tail.Priority {
		tree.tail = node
	}
}

func deleteFixUp(tree *RBTree, node *lst.Node, parent *lst.Node) {
	// 删除黑节点后可能破坏红黑树性质，这里执行标准修复流程。
	// node 可能为 nil（被删黑节点的子节点为空），此时无法经 node.Parent 回溯，
	// 必须显式携带 parent 定位兄弟节点；否则修复被跳过，红黑性质（含根必为黑）
	// 遭破坏，后续 insert/delete 会解引用损坏的父链。
	for node != tree.root && (node == nil || node.Color == lst.BLACK) {
		if parent == nil {
			break
		}

		isLeftChild := node == parent.Left
		var sibling *lst.Node
		if isLeftChild {
			sibling = parent.Right
		} else {
			sibling = parent.Left
		}

		if sibling == nil {
			// 兄弟为空等价于“兄弟两子皆黑”：兄弟侧无可借节点，
			// 将黑缺口上移至父节点继续修复。
			node = parent
			parent = parent.Parent
			continue
		}

		if sibling.Color == lst.RED {
			sibling.Color = lst.BLACK
			parent.Color = lst.RED
			if isLeftChild {
				leftRotate(tree, parent)
				sibling = parent.Right
			} else {
				rightRotate(tree, parent)
				sibling = parent.Left
			}
			// 旋转后的新兄弟可能为空（nil 表示下红兄弟的子节点可为空树），同样上移修复。
			if sibling == nil {
				node = parent
				parent = parent.Parent
				continue
			}
		}

		siblingLeftBlack := sibling.Left == nil || sibling.Left.Color == lst.BLACK
		siblingRightBlack := sibling.Right == nil || sibling.Right.Color == lst.BLACK

		if siblingLeftBlack && siblingRightBlack {
			sibling.Color = lst.RED
			node = parent
			parent = parent.Parent
		} else {
			if isLeftChild {
				if siblingRightBlack {
					if sibling.Left != nil {
						sibling.Left.Color = lst.BLACK
					}
					sibling.Color = lst.RED
					rightRotate(tree, sibling)
					sibling = parent.Right
				}
			} else {
				if siblingLeftBlack {
					if sibling.Right != nil {
						sibling.Right.Color = lst.BLACK
					}
					sibling.Color = lst.RED
					leftRotate(tree, sibling)
					sibling = parent.Left
				}
			}

			sibling.Color = parent.Color
			parent.Color = lst.BLACK

			if isLeftChild && sibling.Right != nil {
				sibling.Right.Color = lst.BLACK
				leftRotate(tree, parent)
			} else if !isLeftChild && sibling.Left != nil {
				sibling.Left.Color = lst.BLACK
				rightRotate(tree, parent)
			}

			node = tree.root
		}
	}

	if node != nil {
		node.Color = lst.BLACK
	}
}

func (tree *RBTree) delete(node *lst.Node) {
	if node == nil {
		return
	}

	if node != tree.root && (node.Parent == nil ||
		(node.Parent.Left != node && node.Parent.Right != node)) {
		return
	}

	updateHead := tree.head == node
	updateTail := tree.tail == node
	var nextHead, nextTail *lst.Node
	if updateHead {
		nextHead = tree.successor(node)
	}
	if updateTail {
		nextTail = tree.predecessor(node)
	}

	var child *lst.Node
	var parent *lst.Node
	removedColor := node.Color

	if node.Left != nil && node.Right != nil {
		// 双子节点：将后继节点物理换位到 node 的位置，再解链 node 本身，
		// 确保调用者给定的节点真正脱离树。调用方（Cancel/CancelDelay）随后
		// 会把该节点归还节点池并 Reset 其指针；若沿用“复制后继值”实现，
		// 该节点仍留在树中，Reset 会破坏树结构（且 tail 可能滞留失效节点）。
		target := tree.successor(node)
		removedColor = target.Color
		child = target.Right
		parent = target.Parent

		// 先把 target 从原位置摘下，其右子树（可为空）上移填补。
		if child != nil {
			child.Parent = parent
		}
		if parent.Left == target {
			parent.Left = child
		} else {
			parent.Right = child
		}

		// 让 target 占据 node 的位置，node 原位颜色保留在该位置。
		target.Parent = node.Parent
		if node.Parent == nil {
			tree.root = target
		} else if node.Parent.Left == node {
			node.Parent.Left = target
		} else {
			node.Parent.Right = target
		}
		target.Left = node.Left
		node.Left.Parent = target
		target.Right = node.Right
		if node.Right != nil {
			node.Right.Parent = target
		}
		target.Color = node.Color
		if parent == node {
			// 后继是 node 的直接右子时，换位后 child 的父位置由 target 接管，
			// 修复阶段必须从 target 而非已解链的 node 开始回溯。
			parent = target
		}
	} else {
		if node.Left != nil {
			child = node.Left
		} else {
			child = node.Right
		}
		parent = node.Parent

		if child != nil {
			child.Parent = parent
		}

		if parent == nil {
			tree.root = child
		} else if parent.Left == node {
			parent.Left = child
		} else {
			parent.Right = child
		}
	}

	if removedColor == lst.BLACK {
		deleteFixUp(tree, child, parent)
	}

	tree.count--

	if tree.count == 0 {
		tree.head = nil
		tree.tail = nil
	} else {
		if updateHead {
			if nextHead != nil {
				tree.head = nextHead
			} else {
				tree.head = tree.minimum(tree.root)
			}
		}
		if updateTail {
			if nextTail != nil {
				tree.tail = nextTail
			} else {
				tree.tail = tree.maximum(tree.root)
			}
		}
	}
}

// minimum 返回以 node 为根的子树中优先级最小的节点（最左节点），节点为空时返回 nil。
func (tree *RBTree) minimum(node *lst.Node) *lst.Node {
	if node == nil {
		return nil
	}

	for node.Left != nil {
		node = node.Left
	}
	return node
}

// maximum 返回以 node 为根的子树中优先级最大的节点（最右节点），节点为空时返回 nil。
func (tree *RBTree) maximum(node *lst.Node) *lst.Node {
	if node == nil {
		return nil
	}

	for node.Right != nil {
		node = node.Right
	}
	return node
}

// successor 返回 node 的中序后继节点（优先级大于当前节点的最小节点）。
// 若右子树非空，后继为右子树的最小节点；否则沿父链向上找第一个作为左子节点的祖先。
func (tree *RBTree) successor(node *lst.Node) *lst.Node {
	if node.Right != nil {
		return tree.minimum(node.Right)
	}
	parent := node.Parent
	for parent != nil && node == parent.Right {
		node = parent
		parent = parent.Parent
	}
	return parent
}

// predecessor 返回 node 的中序前驱节点（优先级小于当前节点的最大节点）。
// 若左子树非空，前驱为左子树的最大节点；否则沿父链向上找第一个作为右子节点的祖先。
func (tree *RBTree) predecessor(node *lst.Node) *lst.Node {
	if node.Left != nil {
		return tree.maximum(node.Left)
	}
	parent := node.Parent
	for parent != nil && node == parent.Left {
		node = parent
		parent = parent.Parent
	}
	return parent
}

// popMin 弹出并返回最小优先级节点（head）。通过 O(1) 访问 head 获取最小节点，
// 再执行删除修复（deleteFixUp）维持红黑树性质，最后更新 head 指针。
func (tree *RBTree) popMin() *lst.Node {
	node := tree.head
	if node == nil {
		return nil
	}

	nextHead := tree.successor(node)
	parent := node.Parent
	child := node.Right

	if child != nil {
		child.Parent = parent
	}

	if parent == nil {
		tree.root = child
	} else {
		parent.Left = child
	}

	if node.Color == lst.BLACK {
		deleteFixUp(tree, child, parent)
	}

	tree.count--
	if tree.count == 0 {
		tree.head = nil
		tree.tail = nil
	} else {
		if nextHead != nil {
			tree.head = nextHead
		} else {
			tree.head = tree.minimum(tree.root)
		}
	}

	return node
}

// Len 返回红黑树中的节点数量。
func (tree *RBTree) Len() int64 { return tree.count }

// Root 返回红黑树的根节点。
func (tree *RBTree) Root() *lst.Node { return tree.root }

// Front 返回最小优先级节点（head），不将其从树中移除。
func (tree *RBTree) Front() *lst.Node { return tree.head }

// Back 返回最大优先级节点（tail），不将其从树中移除。
func (tree *RBTree) Back() *lst.Node { return tree.tail }

// Remove 从红黑树中删除指定节点。
func (tree *RBTree) Remove(node *lst.Node) { tree.delete(node) }

// inOrderTraverse 以中序遍历（左→根→右）访问红黑树节点，按优先级升序回调 fn。
// fn 返回 false 时提前终止遍历。
func inOrderTraverse(node *lst.Node, fn func(*lst.Node) bool) bool {
	if node == nil {
		return true
	}
	if !inOrderTraverse(node.Left, fn) {
		return false
	}
	if !fn(node) {
		return false
	}
	return inOrderTraverse(node.Right, fn)
}

// Range 以中序遍历（优先级升序）访问所有节点，fn 返回 false 时提前终止。
func (tree *RBTree) Range(fn func(*lst.Node) bool) {
	inOrderTraverse(tree.root, fn)
}

// Slice 将红黑树中所有节点的值按优先级升序收集到切片并返回。
func (tree *RBTree) Slice() []any {
	if tree.count == 0 {
		return nil
	}
	nodes := make([]any, 0, tree.count)
	tree.Range(func(node *lst.Node) bool {
		nodes = append(nodes, node.Value)
		return true
	})
	return nodes
}

// Cleanup 清空红黑树的所有节点引用，将树重置为空状态。
func (tree *RBTree) Cleanup() {
	tree.root = nil
	tree.head = nil
	tree.tail = nil
	tree.count = 0
}

// Push 将一个节点插入红黑树，节点为空时不执行操作。
func (tree *RBTree) Push(node *lst.Node) {
	if node != nil {
		tree.insert(node)
	}
}

// Pop 弹出并返回最小优先级节点，树为空时返回 nil。
func (tree *RBTree) Pop() *lst.Node {
	return tree.popMin()
}
