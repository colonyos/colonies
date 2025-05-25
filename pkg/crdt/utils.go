package crdt

func lowestClientID(a, b ClientID) ClientID {
	if a < b {
		return a
	}
	return b
}

func normalizeNumber(v interface{}) interface{} {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return v
	}
}
