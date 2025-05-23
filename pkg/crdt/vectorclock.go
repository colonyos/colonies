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
	case ClockDominates:
		return copyClock(a), ownerA
	case ClockIsDominated:
		return copyClock(b), ownerB
	case ClockEqual, ClockConcurrent:
		if append {
			// Merge both clocks if appending (e.g., arrays)
			merged := mergeClocks(a, b)
			return merged, "" // No definitive winner in append mode
		}

		// LAST WRITER WINS fallback:
		aVersion := a[ownerA] // defaults to 0 if not present
		bVersion := b[ownerB] // defaults to 0 if not present

		if aVersion > bVersion {
			return copyClock(a), ownerA
		} else if bVersion > aVersion {
			return copyClock(b), ownerB
		}

		// Tie-break on ClientID
		if ownerA < ownerB {
			return copyClock(a), ownerA
		}
		return copyClock(b), ownerB
	}

	// Should not reach here
	return copyClock(a), ownerA
}
