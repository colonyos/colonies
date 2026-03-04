package index

import "sync"

// MapIndex maps secondary keys to sets of primary keys for equality lookups.
// Thread-safe via internal RWMutex.
type MapIndex[K comparable, SK comparable] struct {
	mu    sync.RWMutex
	index map[SK]map[K]struct{}
}

// NewMapIndex creates a new MapIndex.
func NewMapIndex[K comparable, SK comparable]() *MapIndex[K, SK] {
	return &MapIndex[K, SK]{
		index: make(map[SK]map[K]struct{}),
	}
}

// Add associates a primary key with a secondary key.
func (m *MapIndex[K, SK]) Add(key K, secondaryKey SK) {
	m.mu.Lock()
	defer m.mu.Unlock()
	set := m.index[secondaryKey]
	if set == nil {
		set = make(map[K]struct{})
		m.index[secondaryKey] = set
	}
	set[key] = struct{}{}
}

// Remove disassociates a primary key from a secondary key.
func (m *MapIndex[K, SK]) Remove(key K, secondaryKey SK) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if set, ok := m.index[secondaryKey]; ok {
		delete(set, key)
		if len(set) == 0 {
			delete(m.index, secondaryKey)
		}
	}
}

// Lookup returns all primary keys associated with the given secondary key.
// Returns nil if none found.
func (m *MapIndex[K, SK]) Lookup(secondaryKey SK) []K {
	m.mu.RLock()
	defer m.mu.RUnlock()
	set := m.index[secondaryKey]
	if len(set) == 0 {
		return nil
	}
	result := make([]K, 0, len(set))
	for k := range set {
		result = append(result, k)
	}
	return result
}

// Count returns the number of primary keys associated with the given secondary key.
func (m *MapIndex[K, SK]) Count(secondaryKey SK) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.index[secondaryKey])
}

// Clear removes all entries from the index.
func (m *MapIndex[K, SK]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.index = make(map[SK]map[K]struct{})
}
