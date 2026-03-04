package index

import (
	"sync"

	"github.com/google/btree"
)

// IndexEntry is a btree entry pairing a sort key with a primary key.
type IndexEntry[K comparable] struct {
	SortKey    int64
	PrimaryKey K
}

// OrderedIndex provides sorted access to primary keys via a btree.
// Thread-safe via internal RWMutex.
type OrderedIndex[K comparable] struct {
	mu   sync.RWMutex
	tree *btree.BTreeG[IndexEntry[K]]
	less func(a, b IndexEntry[K]) bool
}

// NewOrderedIndex creates a new OrderedIndex. Entries are sorted by SortKey
// ascending, with ties broken by primary key string representation.
func NewOrderedIndex[K comparable](keyLess func(a, b K) bool) *OrderedIndex[K] {
	less := func(a, b IndexEntry[K]) bool {
		if a.SortKey != b.SortKey {
			return a.SortKey < b.SortKey
		}
		return keyLess(a.PrimaryKey, b.PrimaryKey)
	}
	return &OrderedIndex[K]{
		tree: btree.NewG(32, less),
		less: less,
	}
}

// NewStringOrderedIndex creates an OrderedIndex with string primary keys.
func NewStringOrderedIndex() *OrderedIndex[string] {
	return NewOrderedIndex[string](func(a, b string) bool { return a < b })
}

// Add inserts an entry into the index.
func (o *OrderedIndex[K]) Add(entry IndexEntry[K]) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.tree.ReplaceOrInsert(entry)
}

// Remove removes an entry from the index.
func (o *OrderedIndex[K]) Remove(entry IndexEntry[K]) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.tree.Delete(entry)
}

// AscendFirst calls fn for the first n entries in ascending order.
// Iteration stops when fn returns false or n entries have been visited.
func (o *OrderedIndex[K]) AscendFirst(n int, fn func(IndexEntry[K]) bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	count := 0
	o.tree.Ascend(func(item IndexEntry[K]) bool {
		if count >= n {
			return false
		}
		count++
		return fn(item)
	})
}

// DescendFirst calls fn for the first n entries in descending order.
// Iteration stops when fn returns false or n entries have been visited.
func (o *OrderedIndex[K]) DescendFirst(n int, fn func(IndexEntry[K]) bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	count := 0
	o.tree.Descend(func(item IndexEntry[K]) bool {
		if count >= n {
			return false
		}
		count++
		return fn(item)
	})
}

// AscendRange calls fn for entries with SortKey in [min, max] (inclusive).
// Iteration stops when fn returns false.
func (o *OrderedIndex[K]) AscendRange(min, max int64, fn func(IndexEntry[K]) bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	var zeroK K
	pivot := IndexEntry[K]{SortKey: min, PrimaryKey: zeroK}
	o.tree.AscendGreaterOrEqual(pivot, func(item IndexEntry[K]) bool {
		if item.SortKey > max {
			return false
		}
		return fn(item)
	})
}

// AscendGreaterThan calls fn for entries with SortKey > threshold.
// Iteration stops when fn returns false.
func (o *OrderedIndex[K]) AscendGreaterThan(threshold int64, fn func(IndexEntry[K]) bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	var zeroK K
	// Start from threshold+1 to be strictly greater
	pivot := IndexEntry[K]{SortKey: threshold + 1, PrimaryKey: zeroK}
	o.tree.AscendGreaterOrEqual(pivot, func(item IndexEntry[K]) bool {
		return fn(item)
	})
}

// Len returns the number of entries in the index.
func (o *OrderedIndex[K]) Len() int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.tree.Len()
}

// lenUnlocked returns the count without acquiring the lock.
func (o *OrderedIndex[K]) lenUnlocked() int {
	return o.tree.Len()
}

// Clear removes all entries.
func (o *OrderedIndex[K]) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.tree.Clear(false)
}
