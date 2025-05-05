package crdt

func copyFields(original map[string]VersionedField) map[string]VersionedField {
	newFields := make(map[string]VersionedField)
	for k, v := range original {
		newFields[k] = v
	}
	return newFields
}

func edgeExists(edges []*Edge, candidate *Edge) bool {
	for _, e := range edges {
		if e.From == candidate.From && e.To == candidate.To && e.Label == candidate.Label && e.Position == candidate.Position {
			return true
		}
	}
	return false
}

func lowestClientID(a, b ClientID) ClientID {
	if a < b {
		return a
	}
	return b
}
