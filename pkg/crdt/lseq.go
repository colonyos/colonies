package crdt

import (
	"fmt"
	"math/rand"
	"sort"
)

// Constants for LSEQ allocation
const (
	Base  = 16 // Base interval range
	Bound = 10 // Controls probability of depth increase
)

// Position is a variable-length array of ints
type Position []int

// Compare lexicographically
func (p Position) Less(other Position) bool {
	for i := 0; i < len(p) && i < len(other); i++ {
		if p[i] < other[i] {
			return true
		} else if p[i] > other[i] {
			return false
		}
	}
	return len(p) < len(other)
}

// Element represents a character with a position and a unique NodeID
type Element struct {
	Value    rune
	Position Position
	NodeID   string // Unique identifier per client insertion
}

// Document holds LSEQ elements
type Document struct {
	Elements []Element
	SiteID   int
}

// NewDocument creates a new document
func NewDocument(siteID int) *Document {
	return &Document{SiteID: siteID}
}

// InsertBetween inserts a new character between two positions with NodeID
func (doc *Document) InsertBetween(left, right Position, value rune, nodeID string) {
	newPos := generatePositionBetweenLSEQ(left, right)
	element := Element{Value: value, Position: newPos, NodeID: nodeID}
	doc.Elements = append(doc.Elements, element)
	doc.sort()
}

func (doc *Document) sort() {
	sort.Slice(doc.Elements, func(i, j int) bool {
		if doc.Elements[i].Position.Less(doc.Elements[j].Position) {
			return true
		}
		if doc.Elements[j].Position.Less(doc.Elements[i].Position) {
			return false
		}
		return doc.Elements[i].NodeID < doc.Elements[j].NodeID // Tie-breaker
	})
}

func (doc *Document) Print() {
	for _, el := range doc.Elements {
		fmt.Printf("%c", el.Value)
	}
	fmt.Println()
}

func (p Position) String() string {
	return fmt.Sprint([]int(p))
}

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

		gap := r - l - 1
		if gap > 0 {
			newDigit := l + 1 + rand.Intn(gap)
			pos = append(pos, newDigit)
			return pos
		}

		// No gap, append l and move to deeper level
		pos = append(pos, l)
		level++

		// Occasionally stop depth increase to avoid long positions
		if rand.Intn(Bound) == 0 {
			// Fallback if still no gap: add a middle digit arbitrarily
			pos = append(pos, Base/2)
			return pos
		}
	}
}

//	func generatePositionBetweenLSEQ(left, right Position) Position {
//		pos := Position{}
//		level := 0
//
//		for {
//			var l, r int
//			if level < len(left) {
//				l = left[level]
//			}
//			if level < len(right) {
//				r = right[level]
//			} else {
//				r = Base
//			}
//
//			if r-l > 1 {
//				// Allocate randomly within the gap
//				newDigit := rand.Intn(r-l-1) + l + 1
//				pos = append(pos, newDigit)
//				break
//			} else {
//				pos = append(pos, l)
//				level++
//			}
//
//			// Control depth increase probability
//			if rand.Intn(Bound) == 0 {
//				break
//			}
//		}
//
//		return pos
//	}
func sortEdgesByLSEQ(edges []*Edge) {
	sort.SliceStable(edges, func(i, j int) bool {
		p1 := edges[i].LSEQPosition
		p2 := edges[j].LSEQPosition

		// Lexicographic comparison of Position slices
		for k := 0; k < len(p1) && k < len(p2); k++ {
			if p1[k] < p2[k] {
				return true
			}
			if p1[k] > p2[k] {
				return false
			}
		}

		if len(p1) != len(p2) {
			return len(p1) < len(p2)
		}

		// Tie-breaker: use NodeID to guarantee deterministic ordering
		return edges[i].To < edges[j].To
	})
}

func main() {
	doc := NewDocument(1)

	// Initial dummy left/right bounds
	doc.InsertBetween(Position{}, Position{Base}, 'A', "clientA-001")
	doc.InsertBetween(doc.Elements[0].Position, Position{Base * 2}, 'B', "clientA-002")
	doc.InsertBetween(doc.Elements[1].Position, Position{Base * 3}, 'C', "clientA-003")

	// Two clients inserting at same location
	doc.InsertBetween(doc.Elements[0].Position, doc.Elements[1].Position, 'X', "clientB-001")
	doc.InsertBetween(doc.Elements[0].Position, doc.Elements[1].Position, 'Y', "clientA-004")

	doc.Print() // Output: A X Y B C (deterministic order)
}
