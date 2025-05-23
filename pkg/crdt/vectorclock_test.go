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

func TestCompareClocks(t *testing.T) {
	tests := []struct {
		name     string
		a        VectorClock
		b        VectorClock
		expected ClockComparison
	}{
		{
			name:     "equal clocks",
			a:        VectorClock{"1": 1},
			b:        VectorClock{"1": 1},
			expected: ClockEqual,
		},
		{
			name:     "a < b",
			a:        VectorClock{"1": 1},
			b:        VectorClock{"1": 2},
			expected: ClockIsDominated,
		},
		{
			name:     "a > b",
			a:        VectorClock{"1": 3},
			b:        VectorClock{"1": 2},
			expected: ClockDominates,
		},
		{
			name:     "a < b (different clients)",
			a:        VectorClock{"1": 1},
			b:        VectorClock{"1": 1, "2": 1},
			expected: ClockIsDominated,
		},
		{
			name:     "a > b (different clients)",
			a:        VectorClock{"1": 1, "2": 2},
			b:        VectorClock{"1": 1, "2": 1},
			expected: ClockDominates,
		},
		{
			name:     "concurrent clocks",
			a:        VectorClock{"1": 2},
			b:        VectorClock{"2": 2},
			expected: ClockConcurrent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareClocks(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("compareClocks(%q) = %v, want %v", tt.name, result, tt.expected)
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
		ownerA ClientID
		ownerB ClientID
		append bool
		want   VectorClock
	}{
		{
			name:   "a wins (newer)",
			a:      VectorClock{"c1": 3},
			b:      VectorClock{"c1": 2},
			ownerA: "c1",
			ownerB: "c1",
			append: false,
			want:   VectorClock{"c1": 3},
		},
		{
			name:   "b wins (newer)",
			a:      VectorClock{"c1": 2},
			b:      VectorClock{"c1": 3},
			ownerA: "c1",
			ownerB: "c1",
			append: false,
			want:   VectorClock{"c1": 3},
		},
		{
			name:   "append merge",
			a:      VectorClock{"c1": 2},
			b:      VectorClock{"c2": 3},
			ownerA: "c1",
			ownerB: "c2",
			append: true,
			want:   VectorClock{"c1": 2, "c2": 3},
		},
		{
			name:   "append merge (b wins)",
			a:      VectorClock{"c1": 1, "c2": 2},
			b:      VectorClock{"c1": 2, "c2": 2},
			ownerA: "c1",
			ownerB: "c2",
			append: false,
			want:   VectorClock{"c1": 2, "c2": 2},
		},
		{
			name:   "tie, lowest client id wins (a lower)",
			a:      VectorClock{"0001": 2},
			b:      VectorClock{"0002": 2},
			ownerA: "0001",
			ownerB: "0002",
			append: false,
			want:   VectorClock{"0001": 2},
		},
		{
			name:   "tie, lowest client id wins (b lower)",
			a:      VectorClock{"0003": 2},
			b:      VectorClock{"0002": 2},
			ownerA: "0003",
			ownerB: "0002",
			append: false,
			want:   VectorClock{"0002": 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClock, _ := resolveConflict(tt.a, tt.b, tt.ownerA, tt.ownerB, tt.append)
			if !clocksEqual(gotClock, tt.want) {
				t.Errorf("resolveConflict(%v) = %v, want %v", tt.name, gotClock, tt.want)
			}
		})
	}
}
