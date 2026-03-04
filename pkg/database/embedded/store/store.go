package store

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/colonyos/colonies/pkg/database/embedded/diskstore"
	"github.com/colonyos/colonies/pkg/database/embedded/wal"
)

// Store is a generic in-memory cache backed by a DiskStore and WAL.
// All reads are served from memory. Writes go to WAL first, then memory.
// Dirty records are flushed to disk periodically by the background flusher.
type Store[K comparable, V any] struct {
	mu         sync.RWMutex
	records    map[K]*V
	dirty      map[K]struct{}
	deleted    map[K]struct{}
	disk       *diskstore.DiskStore[V]
	wal        wal.WAL
	entityName string
	keyToStr   func(K) string
	strToKey   func(string) K
}

// Config holds the configuration for creating a new Store.
type Config[K comparable, V any] struct {
	Disk       *diskstore.DiskStore[V]
	WAL        wal.WAL
	EntityName string
	KeyToStr   func(K) string // convert key to string for WAL/disk
	StrToKey   func(string) K // convert string back to key
}

// NewStore creates a new Store. Call LoadAll() after creation to populate
// from disk, or start empty.
func NewStore[K comparable, V any](cfg Config[K, V]) *Store[K, V] {
	return &Store[K, V]{
		records:    make(map[K]*V),
		dirty:      make(map[K]struct{}),
		deleted:    make(map[K]struct{}),
		disk:       cfg.Disk,
		wal:        cfg.WAL,
		entityName: cfg.EntityName,
		keyToStr:   cfg.KeyToStr,
		strToKey:   cfg.StrToKey,
	}
}

// Put adds or updates a record. The change is written to WAL first,
// then applied to memory and marked dirty.
func (s *Store[K, V]) Put(key K, value *V) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Write to WAL
	if s.wal != nil {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal for WAL: %w", err)
		}
		if err := s.wal.Append(wal.Entry{
			Op:     wal.OpPut,
			Entity: s.entityName,
			Key:    s.keyToStr(key),
			Data:   data,
		}); err != nil {
			return fmt.Errorf("failed to write WAL: %w", err)
		}
	}

	// Apply to memory
	s.records[key] = value
	s.dirty[key] = struct{}{}
	delete(s.deleted, key)

	return nil
}

// Get retrieves a record by primary key. Returns nil, false if not found.
func (s *Store[K, V]) Get(key K) (*V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.records[key]
	return v, ok
}

// Delete removes a record. The change is written to WAL first,
// then applied to memory.
func (s *Store[K, V]) Delete(key K) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.records[key]; !ok {
		return nil // no-op for non-existent keys
	}

	// Write to WAL
	if s.wal != nil {
		if err := s.wal.Append(wal.Entry{
			Op:     wal.OpDelete,
			Entity: s.entityName,
			Key:    s.keyToStr(key),
		}); err != nil {
			return fmt.Errorf("failed to write delete to WAL: %w", err)
		}
	}

	delete(s.records, key)
	delete(s.dirty, key)
	s.deleted[key] = struct{}{}

	return nil
}

// All returns all records in the store. The returned slice contains
// pointers to the actual stored records.
func (s *Store[K, V]) All() []*V {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*V, 0, len(s.records))
	for _, v := range s.records {
		result = append(result, v)
	}
	return result
}

// Filter returns all records matching the predicate.
func (s *Store[K, V]) Filter(fn func(*V) bool) []*V {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*V
	for _, v := range s.records {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Count returns the number of records matching the predicate.
func (s *Store[K, V]) Count(fn func(*V) bool) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, v := range s.records {
		if fn(v) {
			count++
		}
	}
	return count
}

// Len returns the total number of records in the store.
func (s *Store[K, V]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.records)
}

type dirtyEntry[K comparable, V any] struct {
	key   K
	value V
}

// FlushDirty writes all dirty records to disk and removes tombstoned
// records from disk. Clears the dirty and deleted sets afterward.
func (s *Store[K, V]) FlushDirty() error {
	s.mu.Lock()

	if s.disk == nil {
		s.dirty = make(map[K]struct{})
		s.deleted = make(map[K]struct{})
		s.mu.Unlock()
		return nil
	}

	// Snapshot dirty records as a slice (avoid map allocation)
	dirtyEntries := make([]dirtyEntry[K, V], 0, len(s.dirty))
	for k := range s.dirty {
		if v, ok := s.records[k]; ok {
			dirtyEntries = append(dirtyEntries, dirtyEntry[K, V]{key: k, value: *v})
		}
	}

	// Snapshot deleted keys as a slice
	deletedKeys := make([]K, 0, len(s.deleted))
	for k := range s.deleted {
		deletedKeys = append(deletedKeys, k)
	}

	// Clear sets while holding the lock
	s.dirty = make(map[K]struct{})
	s.deleted = make(map[K]struct{})

	s.mu.Unlock()

	// Perform disk I/O outside the lock
	for i := range dirtyEntries {
		if err := s.disk.Write(s.keyToStr(dirtyEntries[i].key), dirtyEntries[i].value); err != nil {
			return fmt.Errorf("failed to flush record to disk: %w", err)
		}
	}

	for _, k := range deletedKeys {
		s.disk.Delete(s.keyToStr(k))
	}

	return nil
}

// LoadAll loads all records from disk into memory. This is called during
// startup to populate the cache. Existing in-memory records are preserved
// (WAL replay may have already added some).
func (s *Store[K, V]) LoadAll() error {
	if s.disk == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.disk.Scan(func(keyStr string, value V) error {
		key := s.strToKey(keyStr)
		// Don't overwrite records already in memory (e.g. from WAL replay)
		if _, exists := s.records[key]; !exists {
			v := value // copy
			s.records[key] = &v
		}
		return nil
	})
}

// Clear removes all records from memory without writing to WAL or disk.
// Used for testing and Drop operations.
func (s *Store[K, V]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = make(map[K]*V)
	s.dirty = make(map[K]struct{})
	s.deleted = make(map[K]struct{})
}

// Lock acquires the write lock. Use with Unlock for atomic multi-step
// operations (e.g. SelectAndAssign). The caller must call Unlock.
func (s *Store[K, V]) Lock() {
	s.mu.Lock()
}

// Unlock releases the write lock.
func (s *Store[K, V]) Unlock() {
	s.mu.Unlock()
}

// RLock acquires the read lock.
func (s *Store[K, V]) RLock() {
	s.mu.RLock()
}

// RUnlock releases the read lock.
func (s *Store[K, V]) RUnlock() {
	s.mu.RUnlock()
}

// PutUnlocked adds a record without acquiring the lock. The caller must
// hold the write lock. Used for atomic multi-step operations.
func (s *Store[K, V]) PutUnlocked(key K, value *V) error {
	if s.wal != nil {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal for WAL: %w", err)
		}
		if err := s.wal.Append(wal.Entry{
			Op:     wal.OpPut,
			Entity: s.entityName,
			Key:    s.keyToStr(key),
			Data:   data,
		}); err != nil {
			return fmt.Errorf("failed to write WAL: %w", err)
		}
	}

	s.records[key] = value
	s.dirty[key] = struct{}{}
	delete(s.deleted, key)
	return nil
}

// GetUnlocked retrieves a record without acquiring the lock. The caller
// must hold at least the read lock.
func (s *Store[K, V]) GetUnlocked(key K) (*V, bool) {
	v, ok := s.records[key]
	return v, ok
}

// DeleteUnlocked removes a record without acquiring the lock. The caller
// must hold the write lock.
func (s *Store[K, V]) DeleteUnlocked(key K) error {
	if _, ok := s.records[key]; !ok {
		return nil
	}

	if s.wal != nil {
		if err := s.wal.Append(wal.Entry{
			Op:     wal.OpDelete,
			Entity: s.entityName,
			Key:    s.keyToStr(key),
		}); err != nil {
			return fmt.Errorf("failed to write delete to WAL: %w", err)
		}
	}

	delete(s.records, key)
	delete(s.dirty, key)
	s.deleted[key] = struct{}{}
	return nil
}

// FilterUnlocked returns matching records without acquiring the lock.
// The caller must hold at least the read lock.
func (s *Store[K, V]) FilterUnlocked(fn func(*V) bool) []*V {
	var result []*V
	for _, v := range s.records {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// CountUnlocked returns count without acquiring the lock.
func (s *Store[K, V]) CountUnlocked(fn func(*V) bool) int {
	count := 0
	for _, v := range s.records {
		if fn(v) {
			count++
		}
	}
	return count
}

// ForEach calls fn for each (key, value) pair in the store.
func (s *Store[K, V]) ForEach(fn func(K, *V)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.records {
		fn(k, v)
	}
}

// SetWAL sets or replaces the WAL for this store. Useful for setting
// the WAL after replay to avoid double-logging.
func (s *Store[K, V]) SetWAL(w wal.WAL) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wal = w
}

// ReplayPut inserts a record directly into memory without writing to WAL.
// The caller must hold the write lock. Used during WAL replay to avoid
// double-logging and deadlocks.
func (s *Store[K, V]) ReplayPut(key K, value *V) {
	s.records[key] = value
	s.dirty[key] = struct{}{}
	delete(s.deleted, key)
}

// ReplayDelete removes a record directly from memory without writing to WAL.
// The caller must hold the write lock. Used during WAL replay.
func (s *Store[K, V]) ReplayDelete(key K) {
	delete(s.records, key)
	delete(s.dirty, key)
	s.deleted[key] = struct{}{}
}

// DirtyCount returns the number of dirty records pending flush.
func (s *Store[K, V]) DirtyCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.dirty)
}

// DeletedCount returns the number of tombstoned records pending flush.
func (s *Store[K, V]) DeletedCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.deleted)
}
