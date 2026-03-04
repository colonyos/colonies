package index

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ===== MapIndex Tests =====

func TestMapIndexAddAndLookup(t *testing.T) {
	m := NewMapIndex[string, string]()
	m.Add("p1", "colony-a")
	m.Add("p2", "colony-a")
	m.Add("p3", "colony-b")

	result := m.Lookup("colony-a")
	sort.Strings(result)
	assert.Equal(t, []string{"p1", "p2"}, result)

	result = m.Lookup("colony-b")
	assert.Equal(t, []string{"p3"}, result)
}

func TestMapIndexAddSameKeyMultipleSecondary(t *testing.T) {
	m := NewMapIndex[string, string]()
	m.Add("p1", "tag-a")
	m.Add("p1", "tag-b")

	assert.Equal(t, []string{"p1"}, m.Lookup("tag-a"))
	assert.Equal(t, []string{"p1"}, m.Lookup("tag-b"))
}

func TestMapIndexRemove(t *testing.T) {
	m := NewMapIndex[string, string]()
	m.Add("p1", "colony-a")
	m.Add("p2", "colony-a")

	m.Remove("p1", "colony-a")
	assert.Equal(t, []string{"p2"}, m.Lookup("colony-a"))
}

func TestMapIndexRemoveNonExistent(t *testing.T) {
	m := NewMapIndex[string, string]()
	m.Remove("nope", "nope") // should not panic
}

func TestMapIndexLookupNonExistent(t *testing.T) {
	m := NewMapIndex[string, string]()
	result := m.Lookup("nope")
	assert.Empty(t, result)
}

func TestMapIndexLookupAfterAllRemoved(t *testing.T) {
	m := NewMapIndex[string, string]()
	m.Add("p1", "colony-a")
	m.Remove("p1", "colony-a")

	result := m.Lookup("colony-a")
	assert.Empty(t, result)
}

func TestMapIndexAddDuplicate(t *testing.T) {
	m := NewMapIndex[string, string]()
	m.Add("p1", "colony-a")
	m.Add("p1", "colony-a") // duplicate

	result := m.Lookup("colony-a")
	assert.Len(t, result, 1)
}

func TestMapIndexLargeScale(t *testing.T) {
	m := NewMapIndex[string, string]()
	n := 10000
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key-%d", i)
		secondary := fmt.Sprintf("group-%d", i%10)
		m.Add(key, secondary)
	}

	for g := 0; g < 10; g++ {
		result := m.Lookup(fmt.Sprintf("group-%d", g))
		assert.Len(t, result, n/10)
	}
}

func TestMapIndexCount(t *testing.T) {
	m := NewMapIndex[string, string]()
	m.Add("p1", "colony-a")
	m.Add("p2", "colony-a")
	m.Add("p3", "colony-b")

	assert.Equal(t, 2, m.Count("colony-a"))
	assert.Equal(t, 1, m.Count("colony-b"))
	assert.Equal(t, 0, m.Count("colony-c"))
}

func TestMapIndexClear(t *testing.T) {
	m := NewMapIndex[string, string]()
	m.Add("p1", "colony-a")
	m.Clear()
	assert.Empty(t, m.Lookup("colony-a"))
}

func TestMapIndexIntSecondaryKey(t *testing.T) {
	m := NewMapIndex[string, int]()
	m.Add("p1", 0) // WAITING
	m.Add("p2", 0)
	m.Add("p3", 1) // RUNNING

	waiting := m.Lookup(0)
	sort.Strings(waiting)
	assert.Equal(t, []string{"p1", "p2"}, waiting)
	assert.Equal(t, []string{"p3"}, m.Lookup(1))
}

// ===== OrderedIndex Tests =====

func TestOrderedIndexAscendFirst(t *testing.T) {
	idx := NewStringOrderedIndex()
	idx.Add(IndexEntry[string]{SortKey: 30, PrimaryKey: "c"})
	idx.Add(IndexEntry[string]{SortKey: 10, PrimaryKey: "a"})
	idx.Add(IndexEntry[string]{SortKey: 20, PrimaryKey: "b"})

	var keys []string
	idx.AscendFirst(10, func(e IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	assert.Equal(t, []string{"a", "b", "c"}, keys)
}

func TestOrderedIndexDescendFirst(t *testing.T) {
	idx := NewStringOrderedIndex()
	idx.Add(IndexEntry[string]{SortKey: 30, PrimaryKey: "c"})
	idx.Add(IndexEntry[string]{SortKey: 10, PrimaryKey: "a"})
	idx.Add(IndexEntry[string]{SortKey: 20, PrimaryKey: "b"})

	var keys []string
	idx.DescendFirst(10, func(e IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	assert.Equal(t, []string{"c", "b", "a"}, keys)
}

func TestOrderedIndexAscendRange(t *testing.T) {
	idx := NewStringOrderedIndex()
	for i := 0; i < 10; i++ {
		idx.Add(IndexEntry[string]{SortKey: int64(i * 10), PrimaryKey: fmt.Sprintf("k%d", i)})
	}

	var keys []string
	idx.AscendRange(20, 50, func(e IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	assert.Equal(t, []string{"k2", "k3", "k4", "k5"}, keys)
}

func TestOrderedIndexAscendRangeEmpty(t *testing.T) {
	idx := NewStringOrderedIndex()
	idx.Add(IndexEntry[string]{SortKey: 10, PrimaryKey: "a"})

	var keys []string
	idx.AscendRange(20, 30, func(e IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	assert.Empty(t, keys)
}

func TestOrderedIndexAscendFirstLimited(t *testing.T) {
	idx := NewStringOrderedIndex()
	for i := 0; i < 100; i++ {
		idx.Add(IndexEntry[string]{SortKey: int64(i), PrimaryKey: fmt.Sprintf("k%03d", i)})
	}

	var keys []string
	idx.AscendFirst(5, func(e IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	assert.Len(t, keys, 5)
	assert.Equal(t, "k000", keys[0])
	assert.Equal(t, "k004", keys[4])
}

func TestOrderedIndexDescendFirstLimited(t *testing.T) {
	idx := NewStringOrderedIndex()
	for i := 0; i < 100; i++ {
		idx.Add(IndexEntry[string]{SortKey: int64(i), PrimaryKey: fmt.Sprintf("k%03d", i)})
	}

	var keys []string
	idx.DescendFirst(3, func(e IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	assert.Len(t, keys, 3)
	assert.Equal(t, "k099", keys[0])
	assert.Equal(t, "k097", keys[2])
}

func TestOrderedIndexRemove(t *testing.T) {
	idx := NewStringOrderedIndex()
	idx.Add(IndexEntry[string]{SortKey: 10, PrimaryKey: "a"})
	idx.Add(IndexEntry[string]{SortKey: 20, PrimaryKey: "b"})
	idx.Add(IndexEntry[string]{SortKey: 30, PrimaryKey: "c"})

	idx.Remove(IndexEntry[string]{SortKey: 20, PrimaryKey: "b"})

	var keys []string
	idx.AscendFirst(10, func(e IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	assert.Equal(t, []string{"a", "c"}, keys)
}

func TestOrderedIndexRemoveNonExistent(t *testing.T) {
	idx := NewStringOrderedIndex()
	idx.Add(IndexEntry[string]{SortKey: 10, PrimaryKey: "a"})
	idx.Remove(IndexEntry[string]{SortKey: 99, PrimaryKey: "nope"}) // no-op

	assert.Equal(t, 1, idx.Len())
}

func TestOrderedIndexSameSortKeyTiebreak(t *testing.T) {
	idx := NewStringOrderedIndex()
	idx.Add(IndexEntry[string]{SortKey: 100, PrimaryKey: "c"})
	idx.Add(IndexEntry[string]{SortKey: 100, PrimaryKey: "a"})
	idx.Add(IndexEntry[string]{SortKey: 100, PrimaryKey: "b"})

	var keys []string
	idx.AscendFirst(10, func(e IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	// Tie-broken by PrimaryKey alphabetically
	assert.Equal(t, []string{"a", "b", "c"}, keys)
}

func TestOrderedIndexLargeScale(t *testing.T) {
	idx := NewStringOrderedIndex()
	n := 100000
	for i := 0; i < n; i++ {
		idx.Add(IndexEntry[string]{SortKey: int64(n - i), PrimaryKey: fmt.Sprintf("k%06d", i)})
	}

	var top10 []string
	idx.AscendFirst(10, func(e IndexEntry[string]) bool {
		top10 = append(top10, e.PrimaryKey)
		return true
	})
	assert.Len(t, top10, 10)
	// SortKey 1 should be first (that's i=n-1)
	assert.Equal(t, fmt.Sprintf("k%06d", n-1), top10[0])
}

func TestOrderedIndexInterleavedAddRemove(t *testing.T) {
	idx := NewStringOrderedIndex()

	// Add 20 entries
	for i := 0; i < 20; i++ {
		idx.Add(IndexEntry[string]{SortKey: int64(i), PrimaryKey: fmt.Sprintf("k%d", i)})
	}

	// Remove even entries
	for i := 0; i < 20; i += 2 {
		idx.Remove(IndexEntry[string]{SortKey: int64(i), PrimaryKey: fmt.Sprintf("k%d", i)})
	}

	assert.Equal(t, 10, idx.Len())

	var keys []string
	idx.AscendFirst(100, func(e IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	assert.Len(t, keys, 10)
	// Should be k1, k3, k5, ..., k19
	assert.Equal(t, "k1", keys[0])
	assert.Equal(t, "k19", keys[9])
}

func TestOrderedIndexAscendGreaterThan(t *testing.T) {
	idx := NewStringOrderedIndex()
	for i := 0; i < 10; i++ {
		idx.Add(IndexEntry[string]{SortKey: int64(i * 10), PrimaryKey: fmt.Sprintf("k%d", i)})
	}

	var keys []string
	idx.AscendGreaterThan(25, func(e IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	// Should get k3 (30), k4 (40), ..., k9 (90)
	assert.Len(t, keys, 7)
	assert.Equal(t, "k3", keys[0])
}

func TestOrderedIndexClear(t *testing.T) {
	idx := NewStringOrderedIndex()
	idx.Add(IndexEntry[string]{SortKey: 10, PrimaryKey: "a"})
	idx.Add(IndexEntry[string]{SortKey: 20, PrimaryKey: "b"})
	idx.Clear()
	assert.Equal(t, 0, idx.Len())
}

// ===== CompoundIndex Tests =====

func TestCompoundIndexBasic(t *testing.T) {
	ci := NewCompoundIndex()
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 100, PrimaryKey: "p1"}) // WAITING
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 50, PrimaryKey: "p2"})
	ci.Add("colony-a", 1, IndexEntry[string]{SortKey: 200, PrimaryKey: "p3"}) // RUNNING
	ci.Add("colony-b", 0, IndexEntry[string]{SortKey: 10, PrimaryKey: "p4"})

	var keys []string
	ci.AscendFirst("colony-a", 0, 10, func(e IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	// Ordered by SortKey: p2 (50), p1 (100)
	assert.Equal(t, []string{"p2", "p1"}, keys)
}

func TestCompoundIndexNonExistentGroup(t *testing.T) {
	ci := NewCompoundIndex()
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 10, PrimaryKey: "p1"})

	var keys []string
	ci.AscendFirst("nope", 0, 10, func(e IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	assert.Empty(t, keys)
}

func TestCompoundIndexNonExistentSubgroup(t *testing.T) {
	ci := NewCompoundIndex()
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 10, PrimaryKey: "p1"})

	var keys []string
	ci.AscendFirst("colony-a", 99, 10, func(e IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	assert.Empty(t, keys)
}

func TestCompoundIndexIsolation(t *testing.T) {
	ci := NewCompoundIndex()
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 10, PrimaryKey: "p1"})
	ci.Add("colony-b", 0, IndexEntry[string]{SortKey: 20, PrimaryKey: "p2"})

	var keysA []string
	ci.AscendFirst("colony-a", 0, 10, func(e IndexEntry[string]) bool {
		keysA = append(keysA, e.PrimaryKey)
		return true
	})
	assert.Equal(t, []string{"p1"}, keysA)

	var keysB []string
	ci.AscendFirst("colony-b", 0, 10, func(e IndexEntry[string]) bool {
		keysB = append(keysB, e.PrimaryKey)
		return true
	})
	assert.Equal(t, []string{"p2"}, keysB)
}

func TestCompoundIndexRemove(t *testing.T) {
	ci := NewCompoundIndex()
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 10, PrimaryKey: "p1"})
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 20, PrimaryKey: "p2"})

	ci.Remove("colony-a", 0, IndexEntry[string]{SortKey: 10, PrimaryKey: "p1"})

	var keys []string
	ci.AscendFirst("colony-a", 0, 10, func(e IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	assert.Equal(t, []string{"p2"}, keys)
}

func TestCompoundIndexRemoveLastCleansUp(t *testing.T) {
	ci := NewCompoundIndex()
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 10, PrimaryKey: "p1"})
	ci.Remove("colony-a", 0, IndexEntry[string]{SortKey: 10, PrimaryKey: "p1"})

	// Internal maps should be cleaned up
	assert.Nil(t, ci.Get("colony-a", 0))
}

func TestCompoundIndexDescend(t *testing.T) {
	ci := NewCompoundIndex()
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 10, PrimaryKey: "p1"})
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 30, PrimaryKey: "p3"})
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 20, PrimaryKey: "p2"})

	var keys []string
	ci.DescendFirst("colony-a", 0, 2, func(e IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	assert.Equal(t, []string{"p3", "p2"}, keys)
}

func TestCompoundIndexCount(t *testing.T) {
	ci := NewCompoundIndex()
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 10, PrimaryKey: "p1"})
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 20, PrimaryKey: "p2"})
	ci.Add("colony-a", 1, IndexEntry[string]{SortKey: 30, PrimaryKey: "p3"})
	ci.Add("colony-b", 0, IndexEntry[string]{SortKey: 40, PrimaryKey: "p4"})

	assert.Equal(t, 2, ci.Count("colony-a", 0))
	assert.Equal(t, 1, ci.Count("colony-a", 1))
	assert.Equal(t, 1, ci.Count("colony-b", 0))
	assert.Equal(t, 0, ci.Count("colony-c", 0))
}

func TestCompoundIndexCountAll(t *testing.T) {
	ci := NewCompoundIndex()
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 10, PrimaryKey: "p1"})
	ci.Add("colony-a", 1, IndexEntry[string]{SortKey: 20, PrimaryKey: "p2"})
	ci.Add("colony-a", 2, IndexEntry[string]{SortKey: 30, PrimaryKey: "p3"})

	assert.Equal(t, 3, ci.CountAll("colony-a"))
	assert.Equal(t, 0, ci.CountAll("colony-b"))
}

func TestCompoundIndexCountBySubgroup(t *testing.T) {
	ci := NewCompoundIndex()
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 10, PrimaryKey: "p1"})
	ci.Add("colony-b", 0, IndexEntry[string]{SortKey: 20, PrimaryKey: "p2"})
	ci.Add("colony-c", 1, IndexEntry[string]{SortKey: 30, PrimaryKey: "p3"})

	assert.Equal(t, 2, ci.CountBySubgroup(0))
	assert.Equal(t, 1, ci.CountBySubgroup(1))
	assert.Equal(t, 0, ci.CountBySubgroup(99))
}

func TestCompoundIndexClear(t *testing.T) {
	ci := NewCompoundIndex()
	ci.Add("colony-a", 0, IndexEntry[string]{SortKey: 10, PrimaryKey: "p1"})
	ci.Add("colony-b", 1, IndexEntry[string]{SortKey: 20, PrimaryKey: "p2"})
	ci.Clear()

	assert.Equal(t, 0, ci.Count("colony-a", 0))
	assert.Equal(t, 0, ci.Count("colony-b", 1))
}
