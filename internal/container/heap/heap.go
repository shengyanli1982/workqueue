package heap

import (
	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
	"github.com/shengyanli1982/workqueue/v2/internal/ternary"
)

// RBTree represents a Red-Black Tree data structure
// RBTree 表示红黑树数据结构
type RBTree struct {
	count int64     // number of nodes in the tree / 树中节点的数量
	root  *lst.Node // root node of the tree / 树的根节点
	head  *lst.Node // node with minimum priority / 具有最小优先级的节点
	tail  *lst.Node // node with maximum priority / 具有最大优先级的节点
}

// New creates a new RBTree
// New 创建一个新的红黑树
func New() *RBTree { return &RBTree{} }

// leftRotate performs a left rotation on the given node
// When performing a left rotation:
// 1. The right child becomes the new root of the subtree
// 2. The original node becomes the left child of the new root
// 3. The left child of the right child becomes the right child of the original node
//
// leftRotate 对给定节点执行左旋转操作
// 执行左旋转时：
// 1. 右子节点成为子树的新根
// 2. 原节点成为新根的左子节点
// 3. 右子节点的左子点成为原节点的右子节点
func leftRotate(tree *RBTree, node *lst.Node) {
	if node == nil || node.Right == nil {
		return
	}
	// 保存当前节点的右子节点
	// Save the right child of current node
	rightChild := node.Right

	// 将右子节点的左子树设置为当前节点的右子树
	// Set right child's left subtree as current node's right subtree
	node.Right = rightChild.Left
	if rightChild.Left != nil {
		rightChild.Left.Parent = node
	}

	// 更新父节点关系
	// Update parent relationships
	rightChild.Parent = node.Parent
	if node.Parent == nil {
		// 如果当前节点是根节点，更新树的根
		// If current node is root, update tree's root
		tree.root = rightChild
	} else {
		// 将右子节点连接到当前节点的父节点
		// Connect right child to current node's parent
		if node == node.Parent.Left {
			node.Parent.Left = rightChild
		} else {
			node.Parent.Right = rightChild
		}
	}

	// 完成旋转
	// Complete the rotation
	rightChild.Left = node
	node.Parent = rightChild
}

// rightRotate performs a right rotation on the given node
// rightRotate 对给定节点执行右旋转操作
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

// insertFixUp maintains Red-Black tree properties after insertion
// The main properties to maintain are:
// 1. Root is always black
// 2. No two adjacent red nodes
// 3. Black height must be the same for all paths
//
// insertFixUp 在插入后维护红黑树的性质
// 需要维护的主要性质：
// 1. 根节点始终为黑色
// 2. 不能有两个相邻的红色节点
// 3. 所有路径的黑色节点高度必须相同
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

// insert adds a new node to the tree while maintaining Red-Black properties
// insert 向树中添加新节点，同时保持红黑树的性质
func (tree *RBTree) insert(node *lst.Node) {
	if node == nil {
		return
	}

	// 寻找插入位置
	// Find insertion position
	var parent *lst.Node
	current := tree.root

	// 遍历找到合适的插入位置
	// Traverse to find proper insertion position
	for current != nil {
		parent = current
		current = ternary.If(node.Priority < current.Priority, current.Left, current.Right)
	}

	// 设置新节点的父节点
	// Set new node's parent
	node.Parent = parent
	if parent == nil {
		// 如果是空树，设置为根节点
		// If tree is empty, set as root
		tree.root = node
	} else {
		// 根据优先级决定是左子节点还是右子节点
		// Determine left or right child based on priority
		if node.Priority < parent.Priority {
			parent.Left = node
		} else {
			parent.Right = node
		}
	}

	// 初始化新节点
	// Initialize new node
	node.Left = nil
	node.Right = nil
	node.Color = lst.RED

	// 修复红黑树性质
	// Fix Red-Black tree properties
	insertFixUp(tree, node)
	tree.count++

	// 更新头尾节点
	// Update head and tail nodes
	if tree.head == nil || node.Priority < tree.head.Priority {
		tree.head = node
	}
	if tree.tail == nil || node.Priority > tree.tail.Priority {
		tree.tail = node
	}
}

// deleteFixUp maintains Red-Black tree properties after deletion
// This function is called when a black node is deleted or moved, which can violate:
// 1. The black height property
// 2. The root property
// The function fixes these violations by recoloring and rotating nodes
//
// deleteFixUp 在删除后维护红黑树的性质
// 当删除或移动黑色节点时调用此函数，可能违反：
// 1. 黑色高度性质
// 2. 根节点性质
// 该函数通过重新着色和旋转节点来修复这些违规
func deleteFixUp(tree *RBTree, node *lst.Node) {
	for node != tree.root && (node == nil || node.Color == lst.BLACK) {
		if node == nil || node.Parent == nil {
			break
		}

		isLeftChild := node == node.Parent.Left
		sibling := ternary.If(isLeftChild, node.Parent.Right, node.Parent.Left)

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

// delete removes a node from the tree while maintaining Red-Black properties
// delete 从树中删除节点，同时保持红黑树的性质
func (tree *RBTree) delete(node *lst.Node) {
	if node == nil {
		return
	}

	// 验证节点是否在树中
	// Verify if node is in the tree
	if node != tree.root && (node.Parent == nil ||
		(node.Parent.Left != node && node.Parent.Right != node)) {
		return
	}

	// 找到实际要删除的节点
	// Find the actual node to delete
	target := ternary.If(node.Left == nil || node.Right == nil, node, tree.successor(node))
	child := ternary.If(target.Left != nil, target.Left, target.Right)

	// 处理子节点的父指针
	// Handle child's parent pointer
	if child != nil {
		child.Parent = target.Parent
	}

	// 更新父节点的子指针
	// Update parent's child pointer
	if target.Parent == nil {
		tree.root = child
	} else {
		if target == target.Parent.Left {
			target.Parent.Left = child
		} else {
			target.Parent.Right = child
		}
	}

	// 如果删除的是后继节点，复制值到原始节点
	// If deleting successor node, copy values to original node
	if target != node {
		node.Value = target.Value
		node.Priority = target.Priority
	}

	// 如果删除的是黑色节点，需要修复红黑树性质
	// If deleted node is black, fix Red-Black tree properties
	if target.Color == lst.BLACK {
		deleteFixUp(tree, child)
	}

	tree.count--

	// 更新头尾节点
	// Update head and tail nodes
	if tree.count == 0 {
		tree.head = nil
		tree.tail = nil
	} else {
		if tree.head == node {
			tree.head = tree.minimum(tree.root)
		}
		if tree.tail == node {
			tree.tail = tree.maximum(tree.root)
		}
	}
}

// minimum finds the node with the smallest priority in the subtree
// minimum 查找子树中具有最小优先级的节点
func (tree *RBTree) minimum(node *lst.Node) *lst.Node {
	if node == nil {
		return nil
	}

	for node.Left != nil {
		node = node.Left
	}
	return node
}

// maximum finds the node with the largest priority in the subtree
// maximum 查找子树中具有最大优先级的节点
func (tree *RBTree) maximum(node *lst.Node) *lst.Node {
	if node == nil {
		return nil
	}

	for node.Right != nil {
		node = node.Right
	}
	return node
}

// successor finds the node with the next larger priority
// For a given node:
// 1. If right subtree exists, return minimum node in right subtree
// 2. Otherwise, go up the tree until we find a parent that contains the node in its left subtree
//
// successor 查找具有下一个更大优先级的节点
// 对于给定节点：
// 1. 如果存在右子树，返回右子树中的最小节点
// 2. 否则，向上遍历树，直到找到一个包该节点在其左子树中的父节点
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

// Len returns the number of nodes in the tree
// Len 返回树中节点的数量
func (tree *RBTree) Len() int64 { return tree.count }

// Root returns the root node of the tree
// Root 返回树的根节点
func (tree *RBTree) Root() *lst.Node { return tree.root }

// Front returns the node with the minimum priority
// Front 返回具有最小优先级的节点
func (tree *RBTree) Front() *lst.Node { return tree.head }

// Back returns the node with the maximum priority
// Back 返回具有最大优先级的节点
func (tree *RBTree) Back() *lst.Node { return tree.tail }

// Remove removes a node from the tree
// Remove 从树中删除节点
func (tree *RBTree) Remove(node *lst.Node) { tree.delete(node) }

// Range traverses the tree in-order and applies the given function to each node
// The traversal order is: left subtree -> current node -> right subtree
// Returns false if the provided function returns false, true otherwise
//
// Range 按中序遍历树，并对每个节点应用给定的函数
// 遍历顺序为左子树 -> 当前节点 -> 右子树
// 如果提供的函数返回 false 则返回 false，否则返回 true
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

// Slice returns a slice containing all values in the tree in sorted order
// If the tree is empty, returns nil
// Slice 返回一个包含树中所有值的有序切片
// 如果树为空，则返回 nil
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

// Cleanup removes all nodes from the tree and resets it to initial state
// Cleanup 移除树中的所有节点并将其重置为初始状态
func (tree *RBTree) Cleanup() {
	tree.root = nil
	tree.head = nil
	tree.tail = nil
	tree.count = 0
}

// Push adds a new node to the tree
// Push 向树中添加新节点
func (tree *RBTree) Push(node *lst.Node) {
	if node != nil {
		tree.insert(node)
	}
}

// Pop removes and returns the node with the minimum priority
// Pop 移除并返回具有最小优先级的节点
func (tree *RBTree) Pop() *lst.Node {
	if tree.head == nil {
		return nil
	}
	node := tree.head
	tree.delete(node)
	return node
}
