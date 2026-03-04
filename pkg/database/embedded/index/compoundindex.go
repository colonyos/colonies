package index

import "sync"

// CompoundIndex provides nested lookups: group -> subgroup -> OrderedIndex.
// Designed for queries like: WHERE colony=$1 AND state=$2 ORDER BY priorityTime.
// Thread-safe via internal RWMutex.
type CompoundIndex struct {
	mu    sync.RWMutex
	index map[string]map[int]*OrderedIndex[string]
}

// NewCompoundIndex creates a new CompoundIndex.
func NewCompoundIndex() *CompoundIndex {
	return &CompoundIndex{
		index: make(map[string]map[int]*OrderedIndex[string]),
	}
}

func (c *CompoundIndex) getOrCreate(group string, subgroup int) *OrderedIndex[string] {
	if c.index[group] == nil {
		c.index[group] = make(map[int]*OrderedIndex[string])
	}
	if c.index[group][subgroup] == nil {
		c.index[group][subgroup] = NewStringOrderedIndex()
	}
	return c.index[group][subgroup]
}

// Add inserts an entry into the index at the given group and subgroup.
func (c *CompoundIndex) Add(group string, subgroup int, entry IndexEntry[string]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.getOrCreate(group, subgroup).Add(entry)
}

// Remove removes an entry from the index at the given group and subgroup.
func (c *CompoundIndex) Remove(group string, subgroup int, entry IndexEntry[string]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if groups, ok := c.index[group]; ok {
		if idx, ok := groups[subgroup]; ok {
			idx.Remove(entry)
			if idx.Len() == 0 {
				delete(groups, subgroup)
				if len(groups) == 0 {
					delete(c.index, group)
				}
			}
		}
	}
}

// Get returns the OrderedIndex for a given group and subgroup, or nil if not found.
func (c *CompoundIndex) Get(group string, subgroup int) *OrderedIndex[string] {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if groups, ok := c.index[group]; ok {
		return groups[subgroup]
	}
	return nil
}

// AscendFirst calls fn for the first n entries in the given group/subgroup
// in ascending order. Returns immediately if group/subgroup doesn't exist.
func (c *CompoundIndex) AscendFirst(group string, subgroup int, n int, fn func(IndexEntry[string]) bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if groups, ok := c.index[group]; ok {
		if idx, ok := groups[subgroup]; ok {
			idx.AscendFirst(n, fn)
		}
	}
}

// DescendFirst calls fn for the first n entries in the given group/subgroup
// in descending order.
func (c *CompoundIndex) DescendFirst(group string, subgroup int, n int, fn func(IndexEntry[string]) bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if groups, ok := c.index[group]; ok {
		if idx, ok := groups[subgroup]; ok {
			idx.DescendFirst(n, fn)
		}
	}
}

// Count returns the number of entries in the given group/subgroup.
func (c *CompoundIndex) Count(group string, subgroup int) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if groups, ok := c.index[group]; ok {
		if idx, ok := groups[subgroup]; ok {
			return idx.Len()
		}
	}
	return 0
}

// CountAll returns the total count across all subgroups for a group.
func (c *CompoundIndex) CountAll(group string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	groups, ok := c.index[group]
	if !ok {
		return 0
	}
	total := 0
	for _, idx := range groups {
		total += idx.Len()
	}
	return total
}

// CountBySubgroup returns the total count for a subgroup across all groups.
func (c *CompoundIndex) CountBySubgroup(subgroup int) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	total := 0
	for _, groups := range c.index {
		if idx, ok := groups[subgroup]; ok {
			total += idx.Len()
		}
	}
	return total
}

// Clear removes all entries.
func (c *CompoundIndex) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.index = make(map[string]map[int]*OrderedIndex[string])
}
