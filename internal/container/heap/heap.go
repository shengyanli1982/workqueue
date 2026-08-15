package heap

import (
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
)

// RBTree 是按 Priority 排序的红黑树实现。
type RBTree struct {
	count int64
	root  *lst.Node
	head  *lst.Node
	tail  *lst.Node
}

func New() *RBTree { return &RBTree{} }

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

func (tree *RBTree) minimum(node *lst.Node) *lst.Node {
	if node == nil {
		return nil
	}

	for node.Left != nil {
		node = node.Left
	}
	return node
}

func (tree *RBTree) maximum(node *lst.Node) *lst.Node {
	if node == nil {
		return nil
	}

	for node.Right != nil {
		node = node.Right
	}
	return node
}

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

func (tree *RBTree) Len() int64 { return tree.count }

func (tree *RBTree) Root() *lst.Node { return tree.root }

func (tree *RBTree) Front() *lst.Node { return tree.head }

func (tree *RBTree) Back() *lst.Node { return tree.tail }

func (tree *RBTree) Remove(node *lst.Node) { tree.delete(node) }

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

func (tree *RBTree) Range(fn func(*lst.Node) bool) {
	inOrderTraverse(tree.root, fn)
}

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

func (tree *RBTree) Cleanup() {
	tree.root = nil
	tree.head = nil
	tree.tail = nil
	tree.count = 0
}

func (tree *RBTree) Push(node *lst.Node) {
	if node != nil {
		tree.insert(node)
	}
}

func (tree *RBTree) Pop() *lst.Node {
	return tree.popMin()
}
