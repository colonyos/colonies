package crdt

import (
	"math/rand"
	"sort"
)

// Constants for LSEQ allocation
const (
	Base  = 16 // Base interval range
	Bound = 10 // Controls probability of depth increase
)

type Position []int

func generatePositionBetweenLSEQ(left, right Position) Position {
	pos := Position{}
	level := 0

	for {
		var l, r int
		if level < len(left) {
			l = left[level]
		}
		if level < len(right) {
			r = right[level]
		} else {
			r = Base
		}

		if r-l > 1 {
			// Space exists, choose randomly in the gap
			newDigit := l + 1 + rand.Intn(r-l-1)
			return append(pos, newDigit)
		}

		// No room, copy l and go deeper
		pos = append(pos, l)
		level++
	}
}

func sortEdgesByLSEQ(edges []*Edge) {
	sort.SliceStable(edges, func(i, j int) bool {
		p1 := edges[i].LSEQPosition
		p2 := edges[j].LSEQPosition

		// Lexicographic comparison
		for k := 0; k < len(p1) && k < len(p2); k++ {
			if p1[k] < p2[k] {
				return true
			}
			if p1[k] > p2[k] {
				return false
			}
		}

		// If one is prefix of the other, shorter one is smaller
		if len(p1) != len(p2) {
			return len(p1) < len(p2)
		}

		// Tie-breaker: use Node To ID to guarantee deterministic sort
		return edges[i].To < edges[j].To
	})
}
