package crdt

import (
	"testing"
)

func TestCopyClock(t *testing.T) {
	orig := VectorClock{
		"client1": 1,
		"client2": 2,
	}
	copy := copyClock(orig)

	// Should be equal initially
	if !clocksEqual(orig, copy) {
		t.Errorf("Copy of clock not equal to original")
	}

	// Modify copy, original should not change
	copy["client1"] = 42
	if orig["client1"] == 42 {
		t.Errorf("Original clock was modified when copy changed")
	}
}

// Unit tests for VectorClock comparison and equality functions.
//
// In these tests, we verify the behavior of:
// 1. compareClocks(a, b): Determines if vector clock `a` is less than, greater than, or concurrent with vector clock `b`.
//   - Returns -1 if `a` is strictly less than `b`
//   - Returns 1 if `a` is strictly greater than `b`
//   - Returns 0 if `a` and `b` are concurrent or identical
//
// 2. clocksEqual(a, b): Checks if two vector clocks are exactly equal (same keys and values).
//
// Vector clocks are used in CRDTs (Conflict-Free Replicated Data Types) to track causality
// between updates made by different clients in distributed systems.
// Correct comparison and equality checking are essential to detect conflicts and merge changes correctly.
func TestCompareClocks(t *testing.T) {
	tests := []struct {
		name     string
		a        VectorClock
		b        VectorClock
		expected int
	}{
		{
			name:     "equal clocks", // a and b are identical
			a:        VectorClock{"1": 1},
			b:        VectorClock{"1": 1},
			expected: 0,
		},
		{
			name:     "a < b", // a is strictly older than b, a happened before b
			a:        VectorClock{"1": 1},
			b:        VectorClock{"1": 2},
			expected: -1,
		},
		{
			name:     "a > b", // a is strictly greater than b, a happened after b
			a:        VectorClock{"1": 3},
			b:        VectorClock{"1": 2},
			expected: 1,
		},
		{
			name:     "a < b (different clients)", // "a < b (different clients)" means all updates in 'a' are included in 'b', but 'b' has extra updates from other clients.
			a:        VectorClock{"1": 1},
			b:        VectorClock{"1": 1, "2": 1},
			expected: -1,
		},
		{
			name:     "a > b (different clients)", // "a > b (different clients)" means all updates in 'b' are included in 'a', but 'a' has extra updates from other clients.
			a:        VectorClock{"1": 1, "2": 2},
			b:        VectorClock{"1": 1, "2": 1},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareClocks(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("compareClocks() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClocksEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        VectorClock
		b        VectorClock
		expected bool
	}{
		{
			name:     "equal clocks",
			a:        VectorClock{"1": 1, "2": 2},
			b:        VectorClock{"2": 2, "1": 1},
			expected: true,
		},
		{
			name:     "different lengths",
			a:        VectorClock{"1": 1},
			b:        VectorClock{"1": 1, "2": 2},
			expected: false,
		},
		{
			name:     "different values",
			a:        VectorClock{"1": 1},
			b:        VectorClock{"1": 2},
			expected: false,
		},
		{
			name:     "different keys",
			a:        VectorClock{"1": 1},
			b:        VectorClock{"3": 1},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clocksEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("clocksEqual() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLowestClientIDAFirst(t *testing.T) {
	tests := []struct {
		name string
		a    VectorClock
		b    VectorClock
		want bool
	}{
		{
			name: "a first",
			a:    VectorClock{"0001": 1, "abcd": 2},
			b:    VectorClock{"0002": 1, "abcd": 2},
			want: true,
		},
		{
			name: "b first",
			a:    VectorClock{"abcd": 2},
			b:    VectorClock{"0001": 1, "abcd": 2},
			want: false,
		},
		{
			name: "equal lowest ID",
			a:    VectorClock{"0001": 1, "abcd": 2},
			b:    VectorClock{"0001": 2, "efgh": 3},
			want: false, // same lowest ID, so a is NOT strictly before b
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lowestClientIDAFirst(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("lowestClientIDAFirst() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveConflict(t *testing.T) {
	tests := []struct {
		name   string
		a      VectorClock
		b      VectorClock
		append bool
		want   VectorClock
	}{
		{
			name:   "a wins (newer)",
			a:      VectorClock{"c1": 3},
			b:      VectorClock{"c1": 2},
			append: false,
			want:   VectorClock{"c1": 3},
		},
		{
			name:   "b wins (newer)",
			a:      VectorClock{"c1": 2},
			b:      VectorClock{"c1": 3},
			append: false,
			want:   VectorClock{"c1": 3},
		},
		{
			name:   "append merge",
			a:      VectorClock{"c1": 2},
			b:      VectorClock{"c2": 3},
			append: true,
			want:   VectorClock{"c1": 2, "c2": 3},
		},
		{
			name:   "tie, lowest client id wins (a lower)",
			a:      VectorClock{"0001": 2},
			b:      VectorClock{"0002": 2},
			append: false,
			want:   VectorClock{"0001": 2},
		},
		{
			name:   "tie, lowest client id wins (b lower)",
			a:      VectorClock{"0003": 2},
			b:      VectorClock{"0002": 2},
			append: false,
			want:   VectorClock{"0002": 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveConflict(tt.a, tt.b, tt.append)
			if !clocksEqual(got, tt.want) {
				t.Errorf("resolveConflict() = %v, want %v", got, tt.want)
			}
		})
	}
}
