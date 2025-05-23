package crdt

func lowestClientID(a, b ClientID) ClientID {
	if a < b {
		return a
	}
	return b
}
