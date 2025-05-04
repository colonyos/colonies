package crdt

import (
	"testing"
)

// LexCompare returns -1 if a < b, 0 if equal, 1 if a > b
func LexCompare(a, b Position) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	} else if len(a) > len(b) {
		return 1
	}
	return 0
}

func TestGeneratePositionBetweenLSEQ(t *testing.T) {
	left := Position{5}
	right := Position{10}

	for i := 0; i < 100; i++ {
		pos := generatePositionBetweenLSEQ(left, right)

		if LexCompare(pos, left) <= 0 {
			t.Errorf("Generated position is not greater than left: got %v", pos)
		}
		if LexCompare(pos, right) >= 0 {
			t.Errorf("Generated position is not less than right: got %v", pos)
		}
		if LexCompare(pos, left) == 0 || LexCompare(pos, right) == 0 {
			t.Errorf("Generated position should not be equal to bounds: got %v", pos)
		}
	}
}

func TestGeneratePositionDepth(t *testing.T) {
	left := Position{1}
	right := Position{2}

	maxDepth := 0
	for i := 0; i < 1000; i++ {
		pos := generatePositionBetweenLSEQ(left, right)
		if len(pos) > maxDepth {
			maxDepth = len(pos)
		}
	}

	if maxDepth > 5 {
		t.Errorf("Depth too large: got max depth %d", maxDepth)
	}
}
