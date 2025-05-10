package crdt

type ClientID string
type VectorClock map[ClientID]int

type ClockComparison int

const (
	ClockEqual ClockComparison = iota
	ClockDominates
	ClockIsDominated
	ClockConcurrent
)

func copyClock(clock VectorClock) VectorClock {
	newClock := make(VectorClock)
	for k, v := range clock {
		newClock[k] = v
	}
	return newClock
}

func compareClocks(a, b VectorClock) ClockComparison {
	less, greater := false, false
	keys := make(map[ClientID]struct{})
	for k := range a {
		keys[k] = struct{}{}
	}
	for k := range b {
		keys[k] = struct{}{}
	}
	for k := range keys {
		av, aok := a[k]
		bv, bok := b[k]
		if !aok {
			av = 0
		}
		if !bok {
			bv = 0
		}
		if av < bv {
			less = true
		}
		if av > bv {
			greater = true
		}
	}
	switch {
	case !less && !greater:
		return ClockEqual
	case less && !greater:
		return ClockIsDominated
	case greater && !less:
		return ClockDominates
	default:
		return ClockConcurrent
	}
}

//	func compareClocks(a, b VectorClock) ClockComparison {
//		less, greater := false, false
//		keys := make(map[ClientID]struct{})
//		for k := range a {
//			keys[k] = struct{}{}
//		}
//		for k := range b {
//			keys[k] = struct{}{}
//		}
//		for k := range keys {
//			av, aok := a[k]
//			bv, bok := b[k]
//			if !aok {
//				av = 0
//			}
//			if !bok {
//				bv = 0
//			}
//			if av < bv {
//				less = true
//			}
//			if av > bv {
//				greater = true
//			}
//		}
//		switch {
//		case !less && !greater:
//			return ClockEqual
//		case less && !greater:
//			return ClockIsDominated
//		case greater && !less:
//			return ClockDominates
//		default:
//			return ClockConcurrent
//		}
//	}
//
//	func compareClocks(a, b VectorClock) int {
//		less, greater := false, false
//		keys := make(map[ClientID]struct{})
//		for k := range a {
//			keys[k] = struct{}{}
//		}
//		for k := range b {
//			keys[k] = struct{}{}
//		}
//		for k := range keys {
//			av, aok := a[k]
//			bv, bok := b[k]
//			if !aok {
//				av = 0
//			}
//			if !bok {
//				bv = 0
//			}
//			if av < bv {
//				less = true
//			}
//			if av > bv {
//				greater = true
//			}
//		}
//		switch {
//		case less && !greater:
//			return -1
//		case greater && !less:
//			return 1
//		default:
//			return 0
//		}
//	}
func clocksEqual(a, b VectorClock) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}

func mergeClocks(a, b VectorClock) VectorClock {
	merged := make(VectorClock)
	for k, v := range a {
		merged[k] = v
	}
	for k, v := range b {
		if mv, ok := merged[k]; !ok || v > mv {
			merged[k] = v
		}
	}
	return merged
}

func lowestClientIDAFirst(a, b VectorClock) bool {
	minA := findLowestClientID(a)
	minB := findLowestClientID(b)

	return minA < minB
}

func findLowestClientID(clock VectorClock) ClientID {
	var minID ClientID
	for id := range clock {
		if minID == "" || id < minID {
			minID = id
		}
	}
	return minID
}

func resolveConflict(a, b VectorClock, ownerA, ownerB ClientID, append bool) (VectorClock, ClientID) {
	cmp := compareClocks(a, b)

	switch cmp {
	case 1:
		// A > B
		return copyClock(a), ownerA
	case -1:
		// B > A
		return copyClock(b), ownerB
	default:
		// clocks are concurrent or equal
		if append {
			// Merge in append mode
			merged := mergeClocks(a, b)
			return merged, "" // optional, since no single winner
		}

		// LAST WRITER WINS LOGIC:
		aVersion := a[ownerA]
		bVersion := b[ownerB]
		if aVersion > bVersion {
			return copyClock(a), ownerA
		} else if bVersion > aVersion {
			return copyClock(b), ownerB
		} else {
			// Tie-break by ID if needed
			if ownerA < ownerB {
				return copyClock(a), ownerA
			} else {
				return copyClock(b), ownerB
			}
		}
	}
}

// func resolveConflict(a, b VectorClock, ownerA, ownerB ClientID, append bool) (VectorClock, ClientID) {
// 	switch compareClocks(a, b) {
// 	case ClockDominates:
// 		fmt.Println("Clock A dominates B")
// 		return copyClock(a), ownerA
// 	case ClockIsDominated:
// 		fmt.Println("Clock B dominates A")
// 		return copyClock(b), ownerB
// 	case ClockEqual:
// 		fmt.Println("Clocks are exactly equal")
// 		return copyClock(a), ownerA // or b, doesn’t matter
// 	case ClockConcurrent:
// 		fmt.Println("Clocks are concurrent")
// 		if append {
// 			merged := mergeClocks(a, b)
// 			return merged, "" // no owner wins
// 		}
// 		if ownerA < ownerB {
// 			return copyClock(a), ownerA
// 		}
// 		return copyClock(b), ownerB
// 	}
// 	panic("unreachable")
// }
//
//
// func resolveConflict(a, b VectorClock, ownerA, ownerB ClientID, append bool) (VectorClock, ClientID) {
// 	cmp := compareClocks(a, b)
//
// 	switch {
// 	case cmp == 1:
// 		fmt.Println("Clock A is greater than Clock B")
// 		return copyClock(a), ownerA
// 	case cmp == -1:
// 		fmt.Println("Clock B is greater than Clock A")
// 		return copyClock(b), ownerB
// 	default: // 0,  are concurrent or identical
// 		fmt.Println("Clocks are concurrent or identical")
// 		if append {
// 			fmt.Println("Appending to the clock")
// 			merged := make(VectorClock)
// 			for k, v := range a {
// 				merged[k] = v
// 			}
// 			for k, v := range b {
// 				if mv, ok := merged[k]; !ok || v > mv {
// 					merged[k] = v
// 				}
// 			}
// 			return merged, "" // merged case, no clear owner winner
// 		} else {
// 			fmt.Println("Not appending to the clock, tie-breaking by owner ID")
// 			// Tie-breaker: lowest owner ID wins
// 			if ownerA < ownerB {
// 				return copyClock(a), ownerA
// 			} else {
// 				return copyClock(b), ownerB
// 			}
// 		}
// 	}
// }
