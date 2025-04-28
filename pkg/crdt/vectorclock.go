package crdt

type ClientID string
type VectorClock map[ClientID]int

func copyClock(clock VectorClock) VectorClock {
	newClock := make(VectorClock)
	for k, v := range clock {
		newClock[k] = v
	}
	return newClock
}

func compareClocks(a, b VectorClock) int {
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
	case less && !greater:
		return -1
	case greater && !less:
		return 1
	default:
		return 0
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

func resolveConflict(a, b VectorClock, append bool) VectorClock {
	cmp := compareClocks(a, b)

	switch {
	case cmp == 1:
		return copyClock(a)
	case cmp == -1:
		return copyClock(b)
	default:
		if append {
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
		} else {
			if lowestClientIDAFirst(a, b) {
				return copyClock(a)
			} else {
				return copyClock(b)
			}
		}
	}
}
