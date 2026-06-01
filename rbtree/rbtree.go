package rbtree

import "cmp"

type color int

const (
	red   color = 1
	black color = 0
)

type rotation int

const (
	right rotation = 1
	left  rotation = 0
)

func (r rotation) reverse() rotation {
	return (r - 1) * -1
}

type Node[K cmp.Ordered, T any] struct {
	color color
	key   K
	value T

	parent, left, right *Node[K, T]
}

func newNode[K cmp.Ordered, T any](key K, value T) *Node[K, T] {
	return &Node[K, T]{
		key:   key,
		value: value,
		color: red,
	}
}

func (node *Node[K, T]) Key() K {
	return node.key
}

func (node *Node[K, T]) Value() T {
	return node.value
}

func (node *Node[K, T]) insert(n *Node[K, T]) {
	if n.key < node.key {
		node.left = n
	} else {
		node.right = n
	}
	n.parent = node
}

func (node *Node[K, T]) replace(o, n *Node[K, T]) {
	if node.left == o {
		node.left = n
	}
	if node.right == o {
		node.right = n
	}
}

func (node *Node[K, T]) sibling() *Node[K, T] {
	parent := node.parent
	if parent == nil {
		return nil
	}
	if parent.left == node {
		return parent.right
	}
	return parent.left
}

func (node *Node[K, T]) findPos(n *Node[K, T]) rotation {
	if node.left == n {
		return left
	}
	return right
}

type extraBlack[K cmp.Ordered, T any] struct {
	original *Node[K, T]

	parent *Node[K, T]
	pos    rotation
	col    color
}

func addExtraBlack[K cmp.Ordered, T any](node *Node[K, T], col color) *extraBlack[K, T] {
	return &extraBlack[K, T]{
		original: node,
		col:      col,
	}
}

func nilExtraBlack[K cmp.Ordered, T any](parent *Node[K, T], pos rotation, col color) *extraBlack[K, T] {
	return &extraBlack[K, T]{
		parent: parent,
		pos:    pos,
		col:    col,
	}
}

func (ex *extraBlack[K, T]) getParent() *Node[K, T] {
	if ex.original != nil {
		return ex.original.parent
	}
	return ex.parent
}

func (ex *extraBlack[K, T]) getPos() rotation {
	if ex.original != nil {
		return ex.original.parent.findPos(ex.original)
	}
	return ex.pos
}

func (ex *extraBlack[K, T]) isRoot() bool {
	return ex.original != nil && ex.original.parent == nil
}

func (ex *extraBlack[K, T]) isRedAndBlack() bool {
	return ex.col == red
}

func (ex *extraBlack[K, T]) changeColor(col color) {
	if ex.original != nil {
		ex.original.color = col
	}
}

func (ex *extraBlack[K, T]) sibling() *Node[K, T] {
	if ex.original != nil {
		return ex.original.sibling()
	}
	if ex.pos == left {
		return ex.parent.right
	}
	return ex.parent.left
}

type Tree[K cmp.Ordered, T any] struct {
	root *Node[K, T]
	size int
}

func New[K cmp.Ordered, T any]() *Tree[K, T] {
	return &Tree[K, T]{}
}

func (tree *Tree[K, T]) Root() *Node[K, T] {
	return tree.root
}

func (tree *Tree[K, T]) Size() int {
	return tree.size
}

func (tree *Tree[K, T]) Find(key K) *Node[K, T] {
	node := tree.root
	for node != nil {
		if node.key == key {
			return node
		} else if key < node.key {
			node = node.left
		} else {
			node = node.right
		}
	}
	return nil
}

func (tree *Tree[K, T]) Insert(key K, value T) *Node[K, T] {
	node := newNode(key, value)
	parent, exists := tree.findNewPoint(node)
	if exists {
		return nil
	}
	if parent == nil {
		tree.root = node
	} else {
		parent.insert(node)
	}

	tree.size++
	tree.insertFixup(node)

	return node
}

func (tree *Tree[K, T]) Remove(key K) {
	node := tree.Find(key)
	if node == nil {
		return
	}
	if tree.size == 1 {
		tree.root, tree.size = nil, 0
		return
	}

	var replace *Node[K, T]
	var extraBK *extraBlack[K, T]

	if node.left != nil && node.right != nil {
		replace = tree.findPredecessor(node)
		if child := replace.left; child != nil {
			moved := tree.move(child, replace)
			extraBK = addExtraBlack(moved, moved.color)
		} else {
			parent := replace.parent
			if parent == node {
				parent = replace
			}
			extraBK = nilExtraBlack(parent, replace.parent.findPos(replace), replace.color)
		}
	} else {
		if node.left != nil {
			replace = node.left
		} else if node.right != nil {
			replace = node.right
		}

		if replace != nil {
			extraBK = addExtraBlack(replace, replace.color)
		} else {
			extraBK = nilExtraBlack(node.parent, node.parent.findPos(node), node.color)
		}
	}

	tree.move(replace, node)
	if replace != nil {
		replace.color = node.color
	}
	tree.size--
	tree.removeFixup(extraBK)
}

func (tree *Tree[K, T]) insertFixup(node *Node[K, T]) {
	parent := node.parent
	// 루트 노드인 경우
	if parent == nil {
		node.color = black
		return
	}

	// 부모 노드가 Black인 경우 종료
	if parent.color == black {
		return
	}

	grandParent, uncle := parent.parent, parent.sibling()
	pos := parent.findPos(node)

	switch {
	// Case 1. 삼촌 노드가 레드인 경우
	// - 부모와 그 형제의 색을 블랙으로 변경한다.
	// - 조부모를 기준으로 재배열을 반복한다.
	case uncle != nil && uncle.color == red:
		parent.color, uncle.color, grandParent.color = black, black, red
		tree.insertFixup(grandParent)

	// Case 2. 삼촌 노드가 블랙이고, 꺾인 형태인 경우
	// - 부모를 기준으로 삽입된 노드의 반대 방향으로 회전 한다.
	// - 부모를 기준으로 Case 3를 진행한다.
	case isZigzag(grandParent, parent, node):
		tree.rotate(parent, pos.reverse())
		tree.insertFixup(parent)

	// Case 3. 삼촌 노드가 블랙이고, 일직선 형태인 경우
	// - 조부모와 부모의 색을 교환한다.
	// - 조부모를 기준으로 삽입된 노드의 반대 방향으로 회전한다.
	default:
		parent.color, grandParent.color = grandParent.color, parent.color
		tree.rotate(grandParent, pos.reverse())
	}
}

func (tree *Tree[K, T]) removeFixup(extraBK *extraBlack[K, T]) {
	if extraBK.isRedAndBlack() || extraBK.isRoot() {
		extraBK.changeColor(black)
		return
	}

	parent, pos := extraBK.getParent(), extraBK.getPos()
	sibling := extraBK.sibling()
	if sibling == nil {
		return
	}

	var near, far *Node[K, T]
	if pos == left {
		near, far = sibling.left, sibling.right
	} else {
		near, far = sibling.right, sibling.left
	}

	switch {
	// Case 1. 형제가 레드인 경우
	// - 형제를 블랙으로 바꾸고 부모를 레드로 변경
	// - 부모를 기준으로 엑스트라 블랙 방향으로 회전
	// - Case 2, 3, 4로 이동
	case sibling.color == red:
		sibling.color, parent.color = black, red
		tree.rotate(parent, pos)
		tree.removeFixup(extraBK)

	// Case 2. 형제가 블랙이고, 형제의 자식들이 모두 블랙인 경우
	// - 형제를 레드로 만들고 엑스트라 블랙을 부모에게 위임
	// - 엑스트라 블랙이 부여된 부모를 기준으로 Case 1, 3, 4로 이동
	case nodeColor(near) == black && nodeColor(far) == black:
		sibling.color = red
		tree.removeFixup(addExtraBlack(parent, parent.color))

	// Case 3. 형제가 블랙이고, 형제의 레드 자식이 꺾인선으로 연결되어 있는 경우
	// - 형제를 레드로, 꺾인 노드의 색을 블랙으로 변경
	// - Case 4로 이동
	case nodeColor(far) == black:
		sibling.color, near.color = red, black
		tree.rotate(sibling, pos.reverse())
		tree.removeFixup(extraBK)

	// Case 4. 형제가 블랙이고, 형제의 레드 자식이 직선으로 연결되어 있는 경우
	// - 형제의 색을 부모으 색으로, 부모의 색을 블랙으로 변경인
	// - 부모를 기준으로 엑스트라 블랙 방향으로 회전
	default:
		sibling.color, parent.color = parent.color, black
		far.color = black
		tree.rotate(parent, pos)
	}
}

func (tree *Tree[K, T]) replace(o, n *Node[K, T]) {
	if o.parent == nil {
		tree.root = n
	} else {
		if o.parent.left == o {
			o.parent.left = n
		} else {
			o.parent.right = n
		}
	}
	if n != nil {
		n.parent = o.parent
	}
}

func (tree *Tree[K, T]) move(src, dest *Node[K, T]) *Node[K, T] {
	if src != nil {
		if parent := src.parent; parent != nil {
			parent.replace(src, nil)
		}
		tree.replace(dest, src)
		if src.left == nil && dest.left != nil {
			src.insert(dest.left)
		}
		if src.right == nil && dest.right != nil {
			src.insert(dest.right)
		}
	} else {
		tree.replace(dest, nil)
	}
	return src
}

func (tree *Tree[K, T]) rotate(src *Node[K, T], d rotation) {
	var child *Node[K, T]
	var orphan *Node[K, T]

	if d == left {
		child = src.right
		orphan = child.left

		child.left, src.right = src, orphan
	} else {
		child = src.left
		orphan = child.right

		child.right, src.left = src, orphan
	}

	tree.replace(src, child)
	src.parent = child
	if orphan != nil {
		orphan.parent = src
	}

	if child.parent == nil {
		tree.root = child
	}
}

func (tree *Tree[K, T]) findNewPoint(insertedNode *Node[K, T]) (*Node[K, T], bool) {
	for node := tree.root; node != nil; {
		if insertedNode.key == node.key {
			return nil, true
		}

		next := node.right
		if insertedNode.key < node.key {
			next = node.left
		}
		if next == nil {
			return node, false
		}
		node = next
	}
	return nil, false
}

func (tree *Tree[K, T]) findPredecessor(node *Node[K, T]) *Node[K, T] {
	successor := node.left
	for successor.right != nil {
		successor = successor.right
	}
	return successor
}

func nodeColor[K cmp.Ordered, T any](node *Node[K, T]) color {
	if node == nil {
		return black
	}
	return node.color
}

func isZigzag[K cmp.Ordered, T any](g, p, n *Node[K, T]) bool {
	return (g.left == p && p.right == n) || (g.right == p && p.left == n)
}
