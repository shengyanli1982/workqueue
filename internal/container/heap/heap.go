package heap

import lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"

// RBTree 是一个结构体，表示一个红黑树
// RBTree is a struct that represents a RED-BLACK tree
type RBTree struct {
	// root 是一个节点，表示红黑树的树根
	// root is a node, representing the root of the RED-BLACK tree
	root *lst.Node

	// head 是一个节点，表示红黑树中的最小节点
	// head is a node, representing the smallest node in the RED-BLACK tree
	head *lst.Node

	// tail 是一个节点，表示红黑树中的最大节点
	// tail is a node, representing the largest node in the RED-BLACK tree
	tail *lst.Node

	// count 是一个整数，表示红黑树中的节点数量
	// count is an integer, representing the number of nodes in the RED-BLACK tree
	count int64
}

// New 函数用于创建一个新的红黑树, 返回一个指向红黑树的指针
// The New function is used to create a new RED-BLACK tree, returning a pointer to the RED-BLACK tree
func New() *RBTree { return &RBTree{} }

// leftRotate 函数用于对红黑树进行左旋转操作
// The leftRotate function is used to perform a left rotation operation on the RED-BLACK tree
func leftRotate(tree *RBTree, node *lst.Node) {

	// rightChild 是 node 的右子节点
	// rightChild is the right child of node
	rightChild := node.Right

	// 将 node 的右子节点设置为 rightChild 的左子节点
	// Set the right child of node to the left child of rightChild
	node.Right = rightChild.Left

	// 如果 rightChild 的左子节点不为空
	// If the left child of rightChild is not nil
	if rightChild.Left != nil {

		// 将 rightChild 的左子节点的父节点设置为 node
		// Set the parent of the left child of rightChild to node
		rightChild.Left.Parent = node

	}

	// 将 rightChild 的父节点设置为 node 的父节点
	// Set the parent of rightChild to the parent of node
	rightChild.Parent = node.Parent

	// 如果 node 的父节点为空
	// If the parent of node is nil
	if node.Parent == nil {

		// 将树根设置为 rightChild
		// Set the root of the tree to rightChild
		tree.root = rightChild

	} else if node == node.Parent.Left { // 如果 node 是其父节点的左子节点

		// 将 node 的父节点的左子节点设置为 rightChild
		// Set the left child of the parent of node to rightChild
		node.Parent.Left = rightChild

	} else { // 如果 node 是其父节点的右子节点

		// 将 node 的父节点的右子节点设置为 rightChild
		// Set the right child of the parent of node to rightChild
		node.Parent.Right = rightChild

	}

	// 将 rightChild 的左子节点设置为 node
	// Set the left child of rightChild to node
	rightChild.Left = node

	// 将 node 的父节点设置为 rightChild
	// Set the parent of node to rightChild
	node.Parent = rightChild

}

// rightRotate 函数用于对红黑树进行右旋转操作
// The rightRotate function is used to perform a right rotation operation on the RED-BLACK tree
func rightRotate(tree *RBTree, node *lst.Node) {

	// leftChild 是 node 的左子节点
	// leftChild is the left child of node
	leftChild := node.Left

	// 将 node 的左子节点设置为 leftChild 的右子节点
	// Set the left child of node to the right child of leftChild
	node.Left = leftChild.Right

	// 如果 leftChild 的右子节点不为空
	// If the right child of leftChild is not nil
	if leftChild.Right != nil {

		// 将 leftChild 的右子节点的父节点设置为 node
		// Set the parent of the right child of leftChild to node
		leftChild.Right.Parent = node

	}

	// 将 leftChild 的父节点设置为 node 的父节点
	// Set the parent of leftChild to the parent of node
	leftChild.Parent = node.Parent

	// 如果 node 的父节点为空
	// If the parent of node is nil
	if node.Parent == nil {

		// 将树根设置为 leftChild
		// Set the root of the tree to leftChild
		tree.root = leftChild

	} else if node == node.Parent.Right { // 如果 node 是其父节点的右子节点

		// 将 node 的父节点的右子节点设置为 leftChild
		// Set the right child of the parent of node to leftChild
		node.Parent.Right = leftChild

	} else { // 如果 node 是其父节点的左子节点

		// 将 node 的父节点的左子节点设置为 leftChild
		// Set the left child of the parent of node to leftChild
		node.Parent.Left = leftChild

	}

	// 将 leftChild 的右子节点设置为 node
	// Set the right child of leftChild to node
	leftChild.Right = node

	// 将 node 的父节点设置为 leftChild
	// Set the parent of node to leftChild
	node.Parent = leftChild

}

// insertFixUp 函数用于在插入节点后调整红黑树，以保持红黑树的性质
// The insertFixUp function is used to adjust the RED-BLACK treeafter inserting a node to maintain the properties of the RED-BLACK tree
func insertFixUp(tree *RBTree, node *lst.Node) {

	// 当 node 的父节点不为空且颜色为红色时，进行循环
	// Loop when the parent of node is not nil and its color is RED
	for node.Parent != nil && node.Parent.Color == lst.RED {

		// 如果 node 的父节点是其父父节点的左子节点
		// If the parent of node is the left child of its grandparent
		if node.Parent == node.Parent.Parent.Left {

			// uncle 是 node 的叔叔节点
			// uncle is the uncle of node
			uncle := node.Parent.Parent.Right

			// 如果叔叔节点不为空且颜色为红色
			// If the uncle node is not nil and its color is RED
			if uncle != nil && uncle.Color == lst.RED {

				// 将 node 的父节点和叔叔节点的颜色都设置为黑色
				// Set the color of the parent and uncle of node to BLACK
				node.Parent.Color = lst.BLACK
				uncle.Color = lst.BLACK

				// 将 node 的父父节点的颜色设置为红色
				// Set the color of the grandparent of node to RED
				node.Parent.Parent.Color = lst.RED

				// 将 node 设置为其父父节点
				// Set node to its grandparent
				node = node.Parent.Parent

			} else { // 如果叔叔节点为空或颜色为黑色

				// 如果 node 是其父节点的右子节点
				// If node is the right child of its parent
				if node == node.Parent.Right {

					// 将 node 设置为其父节点
					// Set node to its parent
					node = node.Parent

					// 对 node 进行左旋转
					// Perform a left rotation on node
					leftRotate(tree, node)

				}

				// 将 node 的父节点的颜色设置为黑色
				// Set the color of the parent of node to BLACK
				node.Parent.Color = lst.BLACK

				// 将 node 的父父节点的颜色设置为红色
				// Set the color of the grandparent of node to RED
				node.Parent.Parent.Color = lst.RED

				// 对 node 的父父节点进行右旋转
				// Perform a right rotation on the grandparent of node
				rightRotate(tree, node.Parent.Parent)

			}

		} else { // 如果 node 的父节点是其父父节点的右子节点

			// uncle 是 node 的叔叔节点
			// uncle is the uncle of node
			uncle := node.Parent.Parent.Left

			// 如果叔叔节点不为空且颜色为红色
			// If the uncle node is not nil and its color is RED
			if uncle != nil && uncle.Color == lst.RED {

				// 将 node 的父节点和叔叔节点的颜色都设置为黑色
				// Set the color of the parent and uncle of node to BLACK
				node.Parent.Color = lst.BLACK

				uncle.Color = lst.BLACK

				// 将 node 的父父节点的颜色设置为红色
				// Set the color of the grandparent of node to RED
				node.Parent.Parent.Color = lst.RED

				// 将 node 设置为其父父节点
				// Set node to its grandparent
				node = node.Parent.Parent

			} else { // 如果叔叔节点为空或颜色为黑色

				// 如果 node 是其父节点的左子节点
				// If node is the left child of its parent
				if node == node.Parent.Left {

					// 将 node 设置为其父节点
					// Set node to its parent
					node = node.Parent

					// 对 node 进行右旋转
					// Perform a right rotation on node
					rightRotate(tree, node)

				}

				// 将 node 的父节点的颜色设置为黑色
				// Set the color of the parent of node to BLACK
				node.Parent.Color = lst.BLACK

				// 将 node 的父父节点的颜色设置为红色
				// Set the color of the grandparent of node to RED
				node.Parent.Parent.Color = lst.RED

				// 对 node 的父父节点进行左旋转
				// Perform a left rotation on the grandparent of node
				leftRotate(tree, node.Parent.Parent)

			}

		}

	}

	// 将树根的颜色设置为黑色
	// Set the color of the root to BLACK
	tree.root.Color = lst.BLACK

}

// insert 方法用于向红黑树中插入一个节点
// The insert method is used to insert a node into the RED-BLACK tree
func (tree *RBTree) insert(node *lst.Node) {

	// parent 用于保存当前节点的父节点
	// parent is used to save the parent of the current node
	var parent *lst.Node

	// current 用于保存当前正在处理的节点，初始为树根
	// current is used to save the node currently being processed, initially the root of the tree
	current := tree.root

	// 当 current 不为空时，进行循环
	// Loop when current is not nil
	for current != nil {

		// 将 parent 设置为 current
		// Set parent to current
		parent = current

		// 如果 node 的优先级小于 current 的优先级
		// If the priority of node is less than the priority of current
		if node.Priority < current.Priority {

			// 将 current 设置为其左子节点
			// Set current to its left child
			current = current.Left

		} else {

			// 将 current 设置为其右子节点
			// Set current to its right child
			current = current.Right

		}

	}

	// 将 node 的父节点设置为 parent
	// Set the parent of node to parent
	node.Parent = parent

	// 如果 parent 为空
	// If parent is nil
	if parent == nil {

		// 将树根设置为 node
		// Set the root of the tree to node
		tree.root = node

	} else if node.Priority < parent.Priority { // 如果 node 的优先级小于 parent 的优先级

		// 将 parent 的左子节点设置为 node
		// Set the left child of parent to node
		parent.Left = node

	} else {

		// 将 parent 的右子节点设置为 node
		// Set the right child of parent to node
		parent.Right = node

	}

	// 将 node 的左子节点和右子节点都设置为 nil
	// Set both the left child and right child of node to nil
	node.Left = nil
	node.Right = nil

	// 将 node 的颜色设置为红色
	// Set the color of node to RED
	node.Color = lst.RED

	// 调用 insertFixUp 方法对树进行调整
	// Call the insertFixUp method to adjust the tree
	insertFixUp(tree, node)

	// 将树的节点数量加一
	// Increase the number of nodes in the tree by one
	tree.count++

	// 如果树的头节点为空或 node 的优先级小于头节点的优先级
	// If the head of the tree is nil or the priority of node is less than the priority of the head
	if tree.head == nil || node.Priority < tree.head.Priority {

		// 将树的头节点设置为 node
		// Set the head of the tree to node
		tree.head = node

	}

	// 如果树的尾节点为空或 node 的优先级大于尾节点的优先级
	// If the tail of the tree is nil or the priority of node is greater than the priority of the tail
	if tree.tail == nil || node.Priority > tree.tail.Priority {

		// 将树的尾节点设置为 node
		// Set the tail of the tree to node
		tree.tail = node

	}

}

// deleteFixUp 函数用于在删除节点后调整红黑树，以保持红黑树的性质
// The deleteFixUp function is used to adjust the RED-BLACK treeafter deleting a node to maintain the properties of the RED-BLACK tree
func deleteFixUp(tree *RBTree, node *lst.Node) {

	// 当 node 不是树根且颜色为黑色时，进行循环
	// Loop when node is not the root and its color is BLACK
	for node != tree.root && (node == nil || node.Color == lst.BLACK) {

		// 如果 node 不为空且有父节点且是其父节点的左子节点
		// If node is not nil and has a parent and is the left child of its parent
		if node != nil && node.Parent != nil && node == node.Parent.Left {

			// sibling 是 node 的兄弟节点
			// sibling is the sibling of node
			sibling := node.Parent.Right

			// 如果兄弟节点的颜色为红色
			// If the color of the sibling node is RED
			if sibling.Color == lst.RED {

				// 将兄弟节点的颜色设置为黑色
				// Set the color of the sibling node to BLACK
				sibling.Color = lst.BLACK

				// 将 node 的父节点的颜色设置为红色
				// Set the color of the parent of node to RED
				node.Parent.Color = lst.RED

				// 对 node 的父节点进行左旋转
				// Perform a left rotation on the parent of node
				leftRotate(tree, node.Parent)

				// 将 sibling 设置为 node 的父节点的右子节点
				// Set sibling to the right child of the parent of node
				sibling = node.Parent.Right

			}

			// 如果兄弟节点的左子节点为空或颜色为黑色且右子节点为空或颜色为黑色
			// If the left child of the sibling node is nil or BLACK and the right child is nil or BLACK
			if (sibling.Left == nil || sibling.Left.Color == lst.BLACK) && (sibling.Right == nil || sibling.Right.Color == lst.BLACK) {

				// 将兄弟节点的颜色设置为红色
				// Set the color of the sibling node to RED
				sibling.Color = lst.RED

				// 将 node 设置为其父节点
				// Set node to its parent
				node = node.Parent

			} else {

				// 如果兄弟节点的右子节点为空或颜色为黑色
				// If the right child of the sibling node is nil or BLACK
				if sibling.Right == nil || sibling.Right.Color == lst.BLACK {

					// 如果兄弟节点的左子节点不为空
					// If the left child of the sibling node is not nil
					if sibling.Left != nil {

						// 将兄弟节点的左子节点的颜色设置为黑色
						// Set the color of the left child of the sibling node to BLACK
						sibling.Left.Color = lst.BLACK

					}

					// 将兄弟节点的颜色设置为红色
					// Set the color of the sibling node to RED
					sibling.Color = lst.RED

					// 对兄弟节点进行右旋转
					// Perform a right rotation on the sibling node
					rightRotate(tree, sibling)

					// 将 sibling 设置为 node 的父节点的右子节点
					// Set sibling to the right child of the parent of node
					sibling = node.Parent.Right

				}

				// 将兄弟节点的颜色设置为 node 的父节点的颜色
				// Set the color of the sibling node to the color of the parent of node
				sibling.Color = node.Parent.Color

				// 将 node 的父节点的颜色设置为黑色
				// Set the color of the parent of node to BLACK
				node.Parent.Color = lst.BLACK

				// 如果兄弟节点的右子节点不为空
				// If the right child of the sibling node is not nil
				if sibling.Right != nil {

					// 将兄弟节点的右子节点的颜色设置为黑色
					// Set the color of the right child of the sibling node to BLACK
					sibling.Right.Color = lst.BLACK

				}

				// 对 node 的父节点进行左旋转
				// Perform a left rotation on the parent of node
				leftRotate(tree, node.Parent)

				// 将 node 设置为树根
				// Set node to the root of the tree
				node = tree.root

			}

		} else if node != nil && node.Parent != nil { // 如果 node 不为空且有父节点

			// sibling 是 node 的兄弟节点
			// sibling is the sibling of node
			sibling := node.Parent.Left

			// 如果兄弟节点的颜色为红色
			// If the color of the sibling node is RED
			if sibling.Color == lst.RED {

				// 将兄弟节点的颜色设置为黑色
				// Set the color of the sibling node to BLACK
				sibling.Color = lst.BLACK

				// 将 node 的父节点的颜色设置为红色
				// Set the color of the parent of node to RED
				node.Parent.Color = lst.RED

				// 对 node 的父节点进行右旋转
				// Perform a right rotation on the parent of node
				rightRotate(tree, node.Parent)

				// 将 sibling 设置为 node 的父节点的左子节点
				// Set sibling to the left child of the parent of node
				sibling = node.Parent.Left

			}

			// 如果兄弟节点的左子节点为空或颜色为黑色且右子节点为空或颜色为黑色
			// If the left child of the sibling node is nil or BLACK and the right child is nil or BLACK
			if (sibling.Left == nil || sibling.Left.Color == lst.BLACK) && (sibling.Right == nil || sibling.Right.Color == lst.BLACK) {

				// 将兄弟节点的颜色设置为红色
				// Set the color of the sibling node to RED
				sibling.Color = lst.RED

				// 将 node 设置为其父节点
				// Set node to its parent
				node = node.Parent

			} else {

				// 如果兄弟节点的左子节点为空或颜色为黑色
				// If the left child of the sibling node is nil or BLACK
				if sibling.Left == nil || sibling.Left.Color == lst.BLACK {

					// 如果兄弟节点的右子节点不为空
					// If the right child of the sibling node is not nil
					if sibling.Right != nil {

						// 将兄弟节点的右子节点的颜色设置为黑色
						// Set the color of the right child of the sibling node to BLACK
						sibling.Right.Color = lst.BLACK

					}

					// 将兄弟节点的颜色设置为红色
					// Set the color of the sibling node to RED
					sibling.Color = lst.RED

					// 对兄弟节点进行左旋转
					// Perform a left rotation on the sibling node
					leftRotate(tree, sibling)

					// 将 sibling 设置为 node 的父节点的左子节点
					// Set sibling to the left child of the parent of node
					sibling = node.Parent.Left

				}

				// 将兄弟节点的颜色设置为 node 的父节点的颜色
				// Set the color of the sibling node to the color of the parent of node
				sibling.Color = node.Parent.Color

				// 将 node 的父节点的颜色设置为黑色
				// Set the color of the parent of node to BLACK
				node.Parent.Color = lst.BLACK

				// 如果兄弟节点的左子节点不为空
				// If the left child of the sibling node is not nil
				if sibling.Left != nil {

					// 将兄弟节点的左子节点的颜色设置为黑色
					// Set the color of the left child of the sibling node to BLACK
					sibling.Left.Color = lst.BLACK

				}

				// 对 node 的父节点进行右旋转
				// Perform a right rotation on the parent of node
				rightRotate(tree, node.Parent)

				// 将 node 设置为树根
				// Set node to the root of the tree
				node = tree.root

			}

		} else {

			// 跳出循环
			// Break the loop
			break

		}

	}

	// 如果 node 不为空
	// If node is not nil
	if node != nil {

		// 将 node 的颜色设置为黑色
		// Set the color of node to BLACK
		node.Color = lst.BLACK

	}

}

// delete 函数用于删除红黑树中的节点
// The delete function is used to delete a node in the RED-BLACK tree
func (tree *RBTree) delete(node *lst.Node) {

	// 定义 child 和 target 节点
	// Define child and target nodes
	var child, target *lst.Node

	// 如果 node 的左子节点为空或右子节点为空
	// If the left child of node is nil or the right child is nil
	if node.Left == nil || node.Right == nil {

		// target 设置为 node
		// Set target to node
		target = node

	} else {

		// target 设置为 node 的后继节点
		// Set target to the successor of node
		target = tree.successor(node)

	}

	// 如果 target 的左子节点不为空
	// If the left child of target is not nil
	if target.Left != nil {

		// child 设置为 target 的左子节点
		// Set child to the left child of target
		child = target.Left

	} else {

		// child 设置为 target 的右子节点
		// Set child to the right child of target
		child = target.Right

	}

	// 如果 child 不为空
	// If child is not nil
	if child != nil {

		// 将 child 的父节点设置为 target 的父节点
		// Set the parent of child to the parent of target
		child.Parent = target.Parent

	}

	// 如果 target 的父节点为空
	// If the parent of target is nil
	if target.Parent == nil {

		// 将树根设置为 child
		// Set the root of the tree to child
		tree.root = child

	} else if target == target.Parent.Left { // 如果 target 是其父节点的左子节点

		// 将 target 的父节点的左子节点设置为 child
		// Set the left child of the parent of target to child
		target.Parent.Left = child

	} else { // 如果 target 是其父节点的右子节点

		// 将 target 的父节点的右子节点设置为 child
		// Set the right child of the parent of target to child
		target.Parent.Right = child

	}

	// 如果 target 不等于 node
	// If target is not equal to node
	if target != node {

		// 将 node 的值设置为 target 的值
		// Set the value of node to the value of target
		node.Value = target.Value

		// 将 node 的优先级设置为 target 的优先级
		// Set the priority of node to the priority of target
		node.Priority = target.Priority

	}

	// 如果 target 的颜色为黑色
	// If the color of target is BLACK
	if target.Color == lst.BLACK {

		// 调用 deleteFixUp 函数调整红黑树
		// Call the deleteFixUp function to adjust the RED-BLACK tree
		deleteFixUp(tree, child)

	}

	// 将树的节点数量减一
	// Decrease the number of nodes in the tree by one
	tree.count--

	// 如果树的节点数量为零
	// If the number of nodes in the tree is zero
	if tree.count == 0 {

		// 将树的头节点和尾节点设置为 nil
		// Set the head and tail nodes of the tree to nil
		tree.head = nil
		tree.tail = nil

	} else {

		// 如果树的头节点等于 node
		// If the head node of the tree is equal to node
		if tree.head == node {

			// 将树的头节点设置为树根的最小节点
			// Set the head node of the tree to the minimum node of the root
			tree.head = tree.minimum(tree.root)

		}

		// 如果树的尾节点等于 node
		// If the tail node of the tree is equal to node
		if tree.tail == node {

			// 将树的尾节点设置为树根的最大节点
			// Set the tail node of the tree to the maximum node of the root
			tree.tail = tree.maximum(tree.root)

		}

	}

}

// minimum 函数返回给定节点的最小子节点
// The minimum function returns the smallest child node of the given node
func (tree *RBTree) minimum(node *lst.Node) *lst.Node {

	// 循环直到找到最左边的节点，也就是最小的节点
	// Loop until the leftmost node is found, which is the smallest node
	for node.Left != nil {
		node = node.Left
	}

	// 返回最小节点
	// Return the smallest node
	return node

}

// maximum 函数返回给定节点的最大子节点
// The maximum function returns the largest child node of the given node
func (tree *RBTree) maximum(node *lst.Node) *lst.Node {

	// 循环直到找到最右边的节点，也就是最大的节点
	// Loop until the rightmost node is found, which is the largest node
	for node.Right != nil {
		node = node.Right
	}

	// 返回最大节点
	// Return the largest node
	return node

}

// successor 函数用于找到给定节点的后继节点
// The successor function is used to find the successor of a given node
func (tree *RBTree) successor(node *lst.Node) *lst.Node {

	// 如果节点的右子节点不为空
	// If the right child of the node is not nil
	if node.Right != nil {

		// 返回节点右子树中的最小节点，即后继节点
		// Return the minimum node in the right subtree of the node, which is the successor
		return tree.minimum(node.Right)
	}

	// 定义 parent 为节点的父节点
	// Define parent as the parent of the node
	parent := node.Parent

	// 当 parent 不为空且节点是其父节点的右子节点时
	// When parent is not nil and the node is the right child of its parent
	for parent != nil && node == parent.Right {

		// 将节点更新为其父节点
		// Update the node to its parent
		node = parent

		// 更新 parent 为其父节点
		// Update parent to its parent
		parent = parent.Parent
	}

	// 返回 parent，即后继节点
	// Return parent, which is the successor
	return parent

}

// Len 函数返回红黑树的节点数量
// The Len function returns the number of nodes in the RED-BLACK tree
func (tree *RBTree) Len() int64 { return tree.count }

// Root 函数返回红黑树的根节点
// The Root function returns the root node of the RED-BLACK tree
func (tree *RBTree) Root() *lst.Node { return tree.root }

// Front 函数返回红黑树的头节点
// The Front function returns the head node of the RED-BLACK tree
func (tree *RBTree) Front() *lst.Node { return tree.head }

// Back 函数返回红黑树的尾节点
// The Back function returns the tail node of the RED-BLACK tree
func (tree *RBTree) Back() *lst.Node { return tree.tail }

// Remove 函数删除红黑树中的给定节点
// The Remove function deletes the given node in the RED-BLACK tree
func (tree *RBTree) Remove(node *lst.Node) { tree.delete(node) }

// inOrderTraverse 函数用于中序遍历节点
// The inOrderTraverse function is used to traverse nodes in order
func inOrderTraverse(node *lst.Node, fn func(*lst.Node) bool) bool {

	// 如果节点为空
	// If the node is nil
	if node == nil {

		// 返回 true
		// Return true
		return true

	}

	// 对节点的左子节点进行中序遍历
	// Traverse the left child of the node in order
	if !inOrderTraverse(node.Left, fn) {

		// 如果遍历返回 false，则返回 false
		// If the traversal returns false, return false
		return false

	}

	// 对节点执行函数 fn
	// Execute function fn on the node
	if !fn(node) {

		// 如果函数 fn 返回 false，则返回 false
		// If function fn returns false, return false
		return false

	}

	// 对节点的右子节点进行中序遍历
	// Traverse the right child of the node in order
	return inOrderTraverse(node.Right, fn)

}

// Range 函数用于对红黑树进行范围操作
// The Range function is used to perform range operations on the RED-BLACK tree
func (tree *RBTree) Range(fn func(*lst.Node) bool) {

	// 对红黑树的根节点进行中序遍历
	// Traverse the root of the RED-BLACK treein order
	inOrderTraverse(tree.root, fn)

}

// Slice 函数返回红黑树的节点值的切片
// The Slice function returns a slice of node values in the RED-BLACK tree
func (tree *RBTree) Slice() []interface{} {

	// 定义 nodes 为节点切片
	// Define nodes as a slice of nodes
	nodes := make([]interface{}, 0, tree.count)

	// 我们遍历链表，将每个节点的 Value 添加到切片中。
	// We traverse the list and add the Value of each node to the slice.
	tree.Range(func(node *lst.Node) bool {
		nodes = append(nodes, node.Value)
		return true
	})

	// 返回节点切片
	// Return the slice of nodes
	return nodes

}

// Cleanup 函数用于清理红黑树
// The Cleanup function is used to clean up the RED-BLACK tree
func (tree *RBTree) Cleanup() {

	// 将树的根节点设置为 nil
	// Set the root of the tree to nil
	tree.root = nil

	// 将树的头节点设置为 nil
	// Set the head of the tree to nil
	tree.head = nil

	// 将树的尾节点设置为 nil
	// Set the tail of the tree to nil
	tree.tail = nil

	// 将树的节点数量设置为 0
	// Set the count of nodes in the tree to 0
	tree.count = 0

}

// Push 函数用于向红黑树中插入节点
// The Push function is used to insert a node into the RED-BLACK tree
func (tree *RBTree) Push(node *lst.Node) {

	// 如果节点不为 nil
	// If the node is not nil
	if node != nil {

		// 调用 insert 函数插入节点
		// Call the insert function to insert the node
		tree.insert(node)

	}

}

// Pop 函数用于从红黑树中弹出节点
// The Pop function is used to pop a node from the RED-BLACK tree
func (tree *RBTree) Pop() *lst.Node {

	// 如果树的头节点为 nil
	// If the head of the tree is nil
	if tree.head == nil {

		// 返回 nil
		// Return nil
		return nil

	}

	// 定义 node 为树的头节点
	// Define node as the head of the tree
	node := tree.head

	// 调用 delete 函数删除节点
	// Call the delete function to delete the node
	tree.delete(node)

	// 返回节点
	// Return the node
	return node

}
