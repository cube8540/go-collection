package rbtree

import (
	"cmp"
	"fmt"
	"math/rand"
	"testing"
)

func TestRBInsertMaintainsInvariants(t *testing.T) {
	testcases := []struct {
		name   string
		values []int
	}{
		{
			name:   "빈 트리",
			values: nil,
		},
		{
			name:   "노드 1개",
			values: []int{10},
		},
		{
			name:   "오름차순 삽입",
			values: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			name:   "내림차순 삽입",
			values: []int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		},
		{
			name:   "좌우 회전이 섞이는 삽입",
			values: []int{20, 15, 3, 12, 5, 11, 7, 30, 25, 40, 35},
		},
		{
			name:   "복잡한 삽입 순서",
			values: []int{50, 25, 75, 10, 30, 60, 80, 5, 15, 27, 35, 55, 65, 77, 90, 1, 6},
		},
		{
			name:   "음수 포함",
			values: []int{0, -10, 10, -20, -5, 5, 20, -30, -15, 15, 30},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tree := New[int, int]()

			for i, v := range tc.values {
				inserted := tree.Insert(v, v)
				if inserted == nil {
					t.Fatalf("%d번째 삽입 값 %d가 중복이 아닌데 nil을 반환했습니다", i, v)
				}

				assertRBTreeInvariants(t, tree)
			}

			if tree.Size() != len(tc.values) {
				t.Fatalf("트리 크기가 일치하지 않습니다. 기대값: %d, 실제값: %d", len(tc.values), tree.Size())
			}

			assertInOrderValues(t, tree, sortedCopy(tc.values))
		})
	}
}

func TestRBInsertDuplicateValue(t *testing.T) {
	tree := New[int, int]()

	inserted := tree.Insert(10, 10)
	if inserted == nil {
		t.Fatal("첫 번째 삽입이 nil을 반환했습니다")
	}

	duplicated := tree.Insert(10, 10)
	if duplicated != nil {
		t.Fatal("중복 값 삽입 시 nil을 반환해야 합니다")
	}

	if tree.Size() != 1 {
		t.Fatalf("중복 삽입 후 size가 증가하면 안 됩니다. 기대값: %d, 실제값: %d", 1, tree.Size())
	}

	if tree.root == nil {
		t.Fatal("루트 노드가 nil입니다")
	}

	if tree.root.value != 10 {
		t.Fatalf("루트 노드 값이 변경되면 안 됩니다. 기대값: %d, 실제값: %d", 10, tree.root.value)
	}

	assertRBTreeInvariants(t, tree)
}

func TestRBInsertDuplicateValuesMixed(t *testing.T) {
	tree := New[int, int]()

	values := []int{10, 5, 20, 5, 10, 20, 15, 15}
	expectedUniqueValues := []int{5, 10, 15, 20}

	insertedCount := 0
	for _, v := range values {
		inserted := tree.Insert(v, v)
		if inserted != nil {
			insertedCount++
		}

		assertRBTreeInvariants(t, tree)
	}

	if insertedCount != len(expectedUniqueValues) {
		t.Fatalf("실제 삽입된 노드 수가 일치하지 않습니다. 기대값: %d, 실제값: %d", len(expectedUniqueValues), insertedCount)
	}

	if tree.Size() != len(expectedUniqueValues) {
		t.Fatalf("중복을 제외한 size가 일치하지 않습니다. 기대값: %d, 실제값: %d", len(expectedUniqueValues), tree.Size())
	}

	assertInOrderValues(t, tree, expectedUniqueValues)
}

func TestRBInsertRootIsAlwaysBlack(t *testing.T) {
	tree := New[int, int]()

	values := []int{41, 38, 31, 12, 19, 8, 50, 60, 55, 54, 53}
	for _, v := range values {
		tree.Insert(v, v)

		if tree.root == nil {
			t.Fatal("삽입 후 루트 노드가 nil입니다")
		}

		if tree.root.color != black {
			t.Fatalf("루트 노드는 항상 Black이어야 합니다. 삽입 값: %d, 루트 값: %d, 루트 색상: %d", v, tree.root.value, tree.root.color)
		}

		assertRBTreeInvariants(t, tree)
	}
}

func TestRBInsertParentPointers(t *testing.T) {
	tree := New[int, int]()

	values := []int{20, 15, 3, 12, 5, 11, 7, 6, 8, 30, 25, 40, 35, 45}
	for _, v := range values {
		tree.Insert(v, v)
		assertParentPointers(t, tree.root, nil)
	}
}

func TestRBInsertWithStringValues(t *testing.T) {
	tree := New[string, string]()

	values := []string{"d", "b", "f", "a", "c", "e", "g"}
	for _, v := range values {
		inserted := tree.Insert(v, v)
		if inserted == nil {
			t.Fatalf("문자열 값 %q 삽입이 nil을 반환했습니다", v)
		}
	}

	if tree.Size() != len(values) {
		t.Fatalf("트리 크기가 일치하지 않습니다. 기대값: %d, 실제값: %d", len(values), tree.Size())
	}

	assertRBTreeInvariants(t, tree)
	assertInOrderValues(t, tree, []string{"a", "b", "c", "d", "e", "f", "g"})
}

func TestRBInsertManySequentialValues(t *testing.T) {
	tree := New[int, int]()

	const count = 1000
	for i := 1; i <= count; i++ {
		inserted := tree.Insert(i, i)
		if inserted == nil {
			t.Fatalf("%d 삽입이 nil을 반환했습니다", i)
		}
	}

	if tree.Size() != count {
		t.Fatalf("트리 크기가 일치하지 않습니다. 기대값: %d, 실제값: %d", count, tree.Size())
	}

	assertRBTreeInvariants(t, tree)
}

func TestRBRemoveMaintainsInvariants(t *testing.T) {
	testcases := []struct {
		name    string
		values  []int
		removes []int
	}{
		{
			name:    "빈 트리에서 삭제",
			values:  nil,
			removes: []int{10},
		},
		{
			name:    "노드 1개 삭제",
			values:  []int{10},
			removes: []int{10},
		},
		{
			name:    "leaf 삭제",
			values:  []int{10, 5, 20, 3, 7, 15, 30},
			removes: []int{3, 7, 15, 30},
		},
		{
			name:    "자식 하나를 가진 노드 삭제",
			values:  []int{10, 5, 20, 3, 15, 30, 25},
			removes: []int{30, 5},
		},
		{
			name:    "자식 둘을 가진 노드 삭제",
			values:  []int{10, 5, 20, 3, 7, 15, 30, 12, 17},
			removes: []int{10, 20, 5},
		},
		{
			name:    "복잡한 순서 삭제",
			values:  []int{50, 25, 75, 10, 30, 60, 80, 5, 15, 27, 35, 55, 65, 77, 90, 1, 6},
			removes: []int{50, 1, 90, 25, 75, 10, 30, 60, 80, 5, 15, 27, 35, 55, 65, 77, 6},
		},
		{
			name:    "오름차순 삭제",
			values:  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			removes: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			name:    "내림차순 삭제",
			values:  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			removes: []int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		},
		{
			name:    "루트를 반복해서 삭제",
			values:  []int{40, 20, 60, 10, 30, 50, 70, 5, 15, 25, 35, 45, 55, 65, 75},
			removes: []int{40, 35, 30, 25, 20, 15, 10, 5},
		},
		{
			name:    "predecessor가 왼쪽 자식을 가진 노드 삭제",
			values:  []int{100, 50, 150, 25, 75, 125, 175, 60, 90, 80, 85, 82},
			removes: []int{100, 75, 85, 50},
		},
		{
			name:    "Black leaf 삭제로 fixup이 연쇄되는 케이스",
			values:  []int{41, 38, 31, 12, 19, 8, 50, 60, 55, 54, 53, 52, 51},
			removes: []int{8, 12, 19, 31, 38, 41},
		},
		{
			name:    "음수와 양수가 섞인 트리 삭제",
			values:  []int{0, -20, 20, -30, -10, 10, 30, -25, -5, 5, 15, 25, 35, -1, 1},
			removes: []int{0, -20, 20, -30, 30, -1, 1, -25, 25},
		},
		{
			name:    "없는 값과 존재하는 값을 섞어서 삭제",
			values:  []int{30, 15, 45, 10, 20, 40, 50, 5, 12, 18, 22, 35, 42, 48, 55},
			removes: []int{999, 30, -999, 15, 45, 777, 5, 55},
		},
		{
			name: "큰 트리에서 지그재그 순서 삭제",
			values: []int{
				64, 32, 96, 16, 48, 80, 112, 8, 24, 40, 56, 72, 88, 104, 120,
				4, 12, 20, 28, 36, 44, 52, 60, 68, 76, 84, 92, 100, 108, 116, 124,
			},
			removes: []int{
				64, 4, 124, 32, 96, 12, 116, 48, 80, 20, 108, 56, 72, 28, 100,
				40, 88, 36, 92, 44, 84, 52, 76, 60, 68, 8, 120, 16, 112, 24, 104,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tree := New[int, int]()

			for _, v := range tc.values {
				tree.Insert(v, v)
			}

			assertRBTreeInvariants(t, tree)

			remaining := append([]int(nil), tc.values...)
			for _, removed := range tc.removes {
				tree.Remove(removed)
				remaining = removeValue(remaining, removed)

				assertRBTreeInvariants(t, tree)
				assertInOrderValues(t, tree, sortedCopy(remaining))

				if tree.Find(removed) != nil {
					t.Fatalf("삭제한 값 %d가 여전히 트리에 존재합니다", removed)
				}
			}
		})
	}
}

func TestRBRemoveMissingValueDoesNotChangeTree(t *testing.T) {
	tree := New[int, int]()

	values := []int{10, 5, 20, 3, 7, 15, 30}
	for _, v := range values {
		tree.Insert(v, v)
	}

	tree.Remove(999)

	if tree.Size() != len(values) {
		t.Fatalf("없는 값 삭제 후 size가 변경되면 안 됩니다. 기대값: %d, 실제값: %d", len(values), tree.Size())
	}

	assertRBTreeInvariants(t, tree)
	assertInOrderValues(t, tree, sortedCopy(values))
}

func TestRBRemoveInterleavedOperationsMaintainInvariants(t *testing.T) {
	type operation struct {
		kind  string
		value int
	}

	testcases := []struct {
		name       string
		initial    []int
		operations []operation
	}{
		{
			name:    "삭제와 재삽입이 반복되는 복잡한 시나리오",
			initial: []int{50, 25, 75, 10, 30, 60, 80, 5, 15, 27, 35, 55, 65, 77, 90},
			operations: []operation{
				{kind: "remove", value: 50},
				{kind: "remove", value: 25},
				{kind: "insert", value: 50},
				{kind: "remove", value: 75},
				{kind: "insert", value: 25},
				{kind: "remove", value: 5},
				{kind: "remove", value: 90},
				{kind: "insert", value: 5},
				{kind: "insert", value: 90},
				{kind: "remove", value: 60},
				{kind: "insert", value: 62},
				{kind: "remove", value: 55},
				{kind: "insert", value: 58},
				{kind: "remove", value: 80},
				{kind: "remove", value: 77},
				{kind: "insert", value: 79},
				{kind: "insert", value: 81},
				{kind: "remove", value: 10},
				{kind: "remove", value: 15},
				{kind: "insert", value: 12},
			},
		},
		{
			name:    "없는 값 삭제와 기존 값 삭제를 교차 수행",
			initial: []int{100, 50, 150, 25, 75, 125, 175, 10, 35, 60, 90, 110, 140, 160, 190},
			operations: []operation{
				{kind: "remove", value: 999},
				{kind: "remove", value: 100},
				{kind: "remove", value: -999},
				{kind: "remove", value: 50},
				{kind: "insert", value: 100},
				{kind: "remove", value: 175},
				{kind: "remove", value: 175},
				{kind: "insert", value: 175},
				{kind: "remove", value: 25},
				{kind: "insert", value: 27},
				{kind: "remove", value: 10},
				{kind: "remove", value: 190},
				{kind: "insert", value: 5},
				{kind: "insert", value: 195},
				{kind: "remove", value: 150},
				{kind: "remove", value: 125},
				{kind: "remove", value: 140},
			},
		},
		{
			name: "큰 트리에서 삭제와 삽입이 번갈아 발생",
			initial: []int{
				64, 32, 96, 16, 48, 80, 112, 8, 24, 40, 56, 72, 88, 104, 120,
				4, 12, 20, 28, 36, 44, 52, 60, 68, 76, 84, 92, 100, 108, 116, 124,
			},
			operations: []operation{
				{kind: "remove", value: 64},
				{kind: "remove", value: 32},
				{kind: "insert", value: 63},
				{kind: "insert", value: 31},
				{kind: "remove", value: 96},
				{kind: "remove", value: 16},
				{kind: "insert", value: 97},
				{kind: "remove", value: 112},
				{kind: "insert", value: 113},
				{kind: "remove", value: 4},
				{kind: "remove", value: 124},
				{kind: "insert", value: 2},
				{kind: "insert", value: 126},
				{kind: "remove", value: 48},
				{kind: "remove", value: 80},
				{kind: "insert", value: 49},
				{kind: "insert", value: 81},
				{kind: "remove", value: 24},
				{kind: "remove", value: 104},
				{kind: "remove", value: 56},
				{kind: "insert", value: 57},
				{kind: "remove", value: 72},
				{kind: "insert", value: 73},
				{kind: "remove", value: 88},
				{kind: "insert", value: 89},
			},
		},
		{
			name:    "음수 양수 혼합 트리에서 삭제 재삽입 반복",
			initial: []int{0, -40, 40, -60, -20, 20, 60, -70, -50, -30, -10, 10, 30, 50, 70},
			operations: []operation{
				{kind: "remove", value: 0},
				{kind: "remove", value: -40},
				{kind: "remove", value: 40},
				{kind: "insert", value: 0},
				{kind: "insert", value: -40},
				{kind: "insert", value: 40},
				{kind: "remove", value: -70},
				{kind: "remove", value: 70},
				{kind: "insert", value: -65},
				{kind: "insert", value: 65},
				{kind: "remove", value: -20},
				{kind: "remove", value: 20},
				{kind: "insert", value: -25},
				{kind: "insert", value: 25},
				{kind: "remove", value: -10},
				{kind: "remove", value: 10},
				{kind: "remove", value: 999},
				{kind: "remove", value: -999},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tree := New[int, int]()
			expected := make(map[int]bool)

			for _, v := range tc.initial {
				inserted := tree.Insert(v, v)
				if inserted == nil {
					t.Fatalf("초기 삽입 값 %d가 중복이 아닌데 nil을 반환했습니다", v)
				}
				expected[v] = true

				assertRBTreeInvariants(t, tree)
				assertInOrderValues(t, tree, sortedKeys(expected))
			}

			for i, op := range tc.operations {
				switch op.kind {
				case "insert":
					inserted := tree.Insert(op.value, op.value)
					if expected[op.value] {
						if inserted != nil {
							t.Fatalf("%d번째 연산: 중복 값 %d 삽입 시 nil을 반환해야 합니다", i, op.value)
						}
					} else {
						if inserted == nil {
							t.Fatalf("%d번째 연산: 새 값 %d 삽입이 nil을 반환했습니다", i, op.value)
						}
						expected[op.value] = true
					}

				case "remove":
					tree.Remove(op.value)
					delete(expected, op.value)

				default:
					t.Fatalf("%d번째 연산: 알 수 없는 연산 종류 %q", i, op.kind)
				}

				if tree.Size() != len(expected) {
					t.Fatalf(
						"%d번째 연산 후 size가 일치하지 않습니다. 연산: %s %d, 기대값: %d, 실제값: %d",
						i,
						op.kind,
						op.value,
						len(expected),
						tree.Size(),
					)
				}

				assertRBTreeInvariants(t, tree)
				assertInOrderValues(t, tree, sortedKeys(expected))

				for v := range expected {
					if tree.Find(v) == nil {
						t.Fatalf("%d번째 연산 후 값 %d를 찾을 수 없습니다", i, v)
					}
				}

				if !expected[op.value] && tree.Find(op.value) != nil {
					t.Fatalf("%d번째 연산 후 존재하지 않아야 할 값 %d가 트리에 남아 있습니다", i, op.value)
				}
			}
		})
	}
}

func TestRBRandomOperations(t *testing.T) {
	tree := New[int, int]()
	expected := map[int]bool{}

	for i := 0; i < 10000; i++ {
		v := rand.Intn(1000)

		if rand.Intn(2) == 0 {
			inserted := tree.Insert(v, v)
			if expected[v] {
				if inserted != nil {
					t.Fatalf("중복 삽입인데 nil이 아닙니다: %d", v)
				}
			} else {
				if inserted == nil {
					t.Fatalf("새 값 삽입인데 nil입니다: %d", v)
				}
				expected[v] = true
			}
		} else {
			tree.Remove(v)
			delete(expected, v)
		}

		if tree.Size() != len(expected) {
			t.Fatalf("size 불일치")
		}

		assertRBTreeInvariants(t, tree)
		assertInOrderValues(t, tree, sortedKeys(expected))
	}
}

func removeValue(values []int, target int) []int {
	for i, v := range values {
		if v == target {
			return append(values[:i], values[i+1:]...)
		}
	}
	return values
}

func sortedCopy(values []int) []int {
	copied := append([]int(nil), values...)

	for i := 1; i < len(copied); i++ {
		key := copied[i]
		j := i - 1

		for j >= 0 && copied[j] > key {
			copied[j+1] = copied[j]
			j--
		}

		copied[j+1] = key
	}

	return copied
}

func sortedKeys(values map[int]bool) []int {
	keys := make([]int, 0, len(values))
	for v := range values {
		keys = append(keys, v)
	}

	for i := 1; i < len(keys); i++ {
		key := keys[i]
		j := i - 1

		for j >= 0 && keys[j] > key {
			keys[j+1] = keys[j]
			j--
		}

		keys[j+1] = key
	}

	return keys
}

func assertRBTreeInvariants[K cmp.Ordered, V any](t *testing.T, tree *Tree[K, V]) {
	t.Helper()

	if tree.root == nil {
		if tree.Size() != 0 {
			t.Fatalf("루트가 nil인데 size가 0이 아닙니다. 실제값: %d", tree.Size())
		}
		return
	}

	if tree.root.parent != nil {
		t.Fatalf("루트의 parent는 nil이어야 합니다. 루트 값: %s", fmt.Sprintf("%#v", tree.root.key))
	}

	if tree.root.color != black {
		t.Fatalf("루트는 Black이어야 합니다. 루트 값: %s, 실제 색상: %d", fmt.Sprintf("%#v", tree.root.key), tree.root.color)
	}

	count, _ := validateRBNode(t, tree.root, nil, nil, nil)
	if count != tree.Size() {
		t.Fatalf("실제 노드 수와 size가 일치하지 않습니다. 실제 노드 수: %d, size: %d", count, tree.Size())
	}
}

func validateRBNode[K cmp.Ordered, V any](t *testing.T, node *Node[K, V], parent *Node[K, V], min *K, max *K) (int, int) {
	t.Helper()

	if node == nil {
		return 0, 1
	}

	if node.parent != parent {
		if parent == nil {
			t.Fatalf("노드 %s의 parent가 nil이어야 합니다", fmt.Sprintf("%#v", node.key))
		}
		v1, v2 := fmt.Sprintf("%#v", node.key), fmt.Sprintf("%#v", parent.key)
		t.Fatalf("노드 %s의 parent가 올바르지 않습니다. 기대 parent: %s", v1, v2)
	}

	if min != nil && node.key <= *min {
		v1, v2 := fmt.Sprintf("%#v", node.value), fmt.Sprintf("%#v", *min)
		t.Fatalf("BST 조건 위반: 노드 값 %s는 하한 %s보다 커야 합니다", v1, v2)
	}

	if max != nil && node.key >= *max {
		v1, v2 := fmt.Sprintf("%#v", node.key), fmt.Sprintf("%#v", *max)
		t.Fatalf("BST 조건 위반: 노드 값 %s는 상한 %s보다 작아야 합니다", v1, v2)
	}

	if node.color == red {
		if node.left != nil && node.left.color == red {
			v1, v2 := fmt.Sprintf("%#v", node.key), fmt.Sprintf("%#v", node.left.key)
			t.Fatalf("Red 노드 %s의 왼쪽 자식 %s도 Red입니다", v1, v2)
		}

		if node.right != nil && node.right.color == red {
			v1, v2 := fmt.Sprintf("%#v", node.key), fmt.Sprintf("%#v", node.right.key)
			t.Fatalf("Red 노드 %s의 오른쪽 자식 %s도 Red입니다", v1, v2)
		}
	}

	leftCount, leftBlackHeight := validateRBNode(t, node.left, node, min, &node.key)
	rightCount, rightBlackHeight := validateRBNode(t, node.right, node, &node.key, max)

	if leftBlackHeight != rightBlackHeight {
		t.Fatalf(
			"black-height가 일치하지 않습니다. 노드: %s, 왼쪽 black-height: %s, 오른쪽 black-height: %s",
			fmt.Sprintf("%#v", node.key),
			fmt.Sprintf("%#v", leftBlackHeight),
			fmt.Sprintf("%#v", rightBlackHeight),
		)
	}

	blackHeight := leftBlackHeight
	if node.color == black {
		blackHeight++
	}

	return leftCount + rightCount + 1, blackHeight
}

func assertParentPointers[K cmp.Ordered, V any](t *testing.T, node *Node[K, V], parent *Node[K, V]) {
	t.Helper()

	if node == nil {
		return
	}

	if node.parent != parent {
		v1 := fmt.Sprintf("%#v", node.key)
		if parent == nil {
			t.Fatalf("노드 %s의 parent가 nil이어야 합니다", v1)
		}
		v2 := fmt.Sprintf("%#v", parent.key)
		t.Fatalf("노드 %s의 parent가 올바르지 않습니다. 기대 parent: %s", v1, v2)
	}

	assertParentPointers(t, node.left, node)
	assertParentPointers(t, node.right, node)
}

func assertInOrderValues[K cmp.Ordered, V any](t *testing.T, tree *Tree[K, V], expected []K) {
	t.Helper()

	actual := make([]K, 0, tree.Size())
	collectInOrderValues(tree.root, &actual)

	if len(actual) != len(expected) {
		t.Fatalf("중위 순회 결과 길이가 일치하지 않습니다. 기대값: %d, 실제값: %d", len(expected), len(actual))
	}

	for i := range expected {
		if actual[i] != expected[i] {
			v1, v2 := fmt.Sprintf("%#v", expected[i]), fmt.Sprintf("%#v", actual[i])
			t.Fatalf("중위 순회 결과가 일치하지 않습니다. index: %d, 기대값: %s, 실제값: %s", i, v1, v2)
		}
	}
}

func collectInOrderValues[K cmp.Ordered, V any](node *Node[K, V], values *[]K) {
	if node == nil {
		return
	}

	collectInOrderValues(node.left, values)
	*values = append(*values, node.key)
	collectInOrderValues(node.right, values)
}
