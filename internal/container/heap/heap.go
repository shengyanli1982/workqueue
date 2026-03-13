package heap

import (
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
	"github.com/shengyanli1982/workqueue/v2/internal/ternary"
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
		current = ternary.If(node.Priority < current.Priority, current.Left, current.Right)
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
	if tree.tail == nil || node.Priority > tree.tail.Priority {
		tree.tail = node
	}
}

func deleteFixUp(tree *RBTree, node *lst.Node) {
	// 删除黑节点后可能破坏红黑树性质，这里执行标准修复流程。
	for node != tree.root && (node == nil || node.Color == lst.BLACK) {
		if node == nil || node.Parent == nil {
			break
		}

		isLeftChild := node == node.Parent.Left
		var sibling *lst.Node
		if isLeftChild {
			sibling = node.Parent.Right
		} else {
			sibling = node.Parent.Left
		}

		if sibling == nil {
			break
		}

		if sibling.Color == lst.RED {
			sibling.Color = lst.BLACK
			node.Parent.Color = lst.RED
			if isLeftChild {
				leftRotate(tree, node.Parent)
				sibling = node.Parent.Right
			} else {
				rightRotate(tree, node.Parent)
				sibling = node.Parent.Left
			}
		}

		siblingLeftBlack := sibling.Left == nil || sibling.Left.Color == lst.BLACK
		siblingRightBlack := sibling.Right == nil || sibling.Right.Color == lst.BLACK

		if siblingLeftBlack && siblingRightBlack {
			sibling.Color = lst.RED
			node = node.Parent
		} else {
			if isLeftChild {
				if siblingRightBlack {
					if sibling.Left != nil {
						sibling.Left.Color = lst.BLACK
					}
					sibling.Color = lst.RED
					rightRotate(tree, sibling)
					sibling = node.Parent.Right
				}
			} else {
				if siblingLeftBlack {
					if sibling.Right != nil {
						sibling.Right.Color = lst.BLACK
					}
					sibling.Color = lst.RED
					leftRotate(tree, sibling)
					sibling = node.Parent.Left
				}
			}

			sibling.Color = node.Parent.Color
			node.Parent.Color = lst.BLACK

			if isLeftChild && sibling.Right != nil {
				sibling.Right.Color = lst.BLACK
				leftRotate(tree, node.Parent)
			} else if !isLeftChild && sibling.Left != nil {
				sibling.Left.Color = lst.BLACK
				rightRotate(tree, node.Parent)
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

	var target *lst.Node
	if node.Left == nil || node.Right == nil {
		target = node
	} else {
		target = tree.successor(node)
	}

	var child *lst.Node
	if target.Left != nil {
		child = target.Left
	} else {
		child = target.Right
	}

	if child != nil {
		child.Parent = target.Parent
	}

	if target.Parent == nil {
		tree.root = child
	} else {
		if target == target.Parent.Left {
			target.Parent.Left = child
		} else {
			target.Parent.Right = child
		}
	}

	if target != node {
		node.Value = target.Value
		node.Priority = target.Priority
	}

	if target.Color == lst.BLACK {
		deleteFixUp(tree, child)
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
		deleteFixUp(tree, child)
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

func (tree *RBTree) Slice() []interface{} {
	if tree.count == 0 {
		return nil
	}
	nodes := make([]interface{}, 0, tree.count)
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
