package store

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/colonyos/colonies/pkg/database/embedded/diskstore"
	"github.com/colonyos/colonies/pkg/database/embedded/wal"
	"github.com/stretchr/testify/assert"
)

type TestRecord struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func mustNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func identity(s string) string { return s }

func newTestStore(t *testing.T) (*Store[string, TestRecord], string) {
	t.Helper()
	dir := t.TempDir()

	ds, err := diskstore.NewDiskStore[TestRecord](filepath.Join(dir, "records"), "data")
	mustNoErr(t, err)

	w, err := wal.NewFileWAL(filepath.Join(dir, "wal"), wal.SyncNone)
	mustNoErr(t, err)

	t.Cleanup(func() { w.Close() })

	s := NewStore[string, TestRecord](Config[string, TestRecord]{
		Disk:       ds,
		WAL:        w,
		EntityName: "data",
		KeyToStr:   identity,
		StrToKey:   identity,
	})

	return s, dir
}

func newTestStoreNoWAL(t *testing.T) *Store[string, TestRecord] {
	t.Helper()
	dir := t.TempDir()

	ds, err := diskstore.NewDiskStore[TestRecord](filepath.Join(dir, "records"), "data")
	mustNoErr(t, err)

	return NewStore[string, TestRecord](Config[string, TestRecord]{
		Disk:       ds,
		EntityName: "data",
		KeyToStr:   identity,
		StrToKey:   identity,
	})
}

func newTestStoreNoDisk(t *testing.T) *Store[string, TestRecord] {
	t.Helper()
	return NewStore[string, TestRecord](Config[string, TestRecord]{
		EntityName: "data",
		KeyToStr:   identity,
		StrToKey:   identity,
	})
}

func rec(id string, value int) *TestRecord {
	return &TestRecord{ID: id, Name: "name-" + id, Value: value}
}

// --- Put/Get round-trip ---

func TestPutGet(t *testing.T) {
	s, _ := newTestStore(t)
	r := rec("r1", 10)
	mustNoErr(t, s.Put("r1", r))

	got, ok := s.Get("r1")
	assert.True(t, ok)
	assert.Equal(t, r, got)
}

func TestPutOverwrite(t *testing.T) {
	s, _ := newTestStore(t)
	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	mustNoErr(t, s.Put("r1", rec("r1", 20)))

	got, ok := s.Get("r1")
	assert.True(t, ok)
	assert.Equal(t, 20, got.Value)
}

func TestGetNonExistent(t *testing.T) {
	s, _ := newTestStore(t)
	_, ok := s.Get("nope")
	assert.False(t, ok)
}

// --- Delete ---

func TestDelete(t *testing.T) {
	s, _ := newTestStore(t)
	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	mustNoErr(t, s.Delete("r1"))

	_, ok := s.Get("r1")
	assert.False(t, ok)
}

func TestDeleteNonExistentNoOp(t *testing.T) {
	s, _ := newTestStore(t)
	err := s.Delete("nope")
	assert.NoError(t, err)
}

func TestDeleteThenPut(t *testing.T) {
	s, _ := newTestStore(t)
	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	mustNoErr(t, s.Delete("r1"))
	mustNoErr(t, s.Put("r1", rec("r1", 20)))

	got, ok := s.Get("r1")
	assert.True(t, ok)
	assert.Equal(t, 20, got.Value)
}

// --- All ---

func TestAll(t *testing.T) {
	s, _ := newTestStore(t)
	mustNoErr(t, s.Put("r1", rec("r1", 1)))
	mustNoErr(t, s.Put("r2", rec("r2", 2)))
	mustNoErr(t, s.Put("r3", rec("r3", 3)))

	all := s.All()
	assert.Len(t, all, 3)
}

func TestAllEmpty(t *testing.T) {
	s, _ := newTestStore(t)
	all := s.All()
	assert.Empty(t, all)
}

// --- Filter ---

func TestFilter(t *testing.T) {
	s, _ := newTestStore(t)
	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	mustNoErr(t, s.Put("r2", rec("r2", 20)))
	mustNoErr(t, s.Put("r3", rec("r3", 30)))

	result := s.Filter(func(r *TestRecord) bool { return r.Value > 15 })
	assert.Len(t, result, 2)
}

func TestFilterNoMatch(t *testing.T) {
	s, _ := newTestStore(t)
	mustNoErr(t, s.Put("r1", rec("r1", 10)))

	result := s.Filter(func(r *TestRecord) bool { return r.Value > 100 })
	assert.Empty(t, result)
}

// --- Count ---

func TestCount(t *testing.T) {
	s, _ := newTestStore(t)
	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	mustNoErr(t, s.Put("r2", rec("r2", 20)))
	mustNoErr(t, s.Put("r3", rec("r3", 30)))

	count := s.Count(func(r *TestRecord) bool { return r.Value >= 20 })
	assert.Equal(t, 2, count)
}

func TestLen(t *testing.T) {
	s, _ := newTestStore(t)
	assert.Equal(t, 0, s.Len())

	mustNoErr(t, s.Put("r1", rec("r1", 1)))
	assert.Equal(t, 1, s.Len())

	mustNoErr(t, s.Put("r2", rec("r2", 2)))
	assert.Equal(t, 2, s.Len())

	mustNoErr(t, s.Delete("r1"))
	assert.Equal(t, 1, s.Len())
}

// --- Dirty tracking ---

func TestPutMarksDirty(t *testing.T) {
	s, _ := newTestStore(t)
	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	assert.Equal(t, 1, s.DirtyCount())
}

func TestFlushClearsDirty(t *testing.T) {
	s, _ := newTestStore(t)
	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	assert.Equal(t, 1, s.DirtyCount())

	mustNoErr(t, s.FlushDirty())
	assert.Equal(t, 0, s.DirtyCount())
}

func TestUnmodifiedNotFlushed(t *testing.T) {
	s, _ := newTestStore(t)
	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	mustNoErr(t, s.FlushDirty())

	// No modifications - dirty should be 0
	assert.Equal(t, 0, s.DirtyCount())
	mustNoErr(t, s.FlushDirty()) // should be a no-op
	assert.Equal(t, 0, s.DirtyCount())
}

func TestDeleteMarksTombstone(t *testing.T) {
	s, _ := newTestStore(t)
	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	mustNoErr(t, s.FlushDirty()) // flush to disk

	mustNoErr(t, s.Delete("r1"))
	assert.Equal(t, 1, s.DeletedCount())

	mustNoErr(t, s.FlushDirty()) // should remove from disk
	assert.Equal(t, 0, s.DeletedCount())
}

// --- FlushDirty writes to disk ---

func TestFlushWritesToDisk(t *testing.T) {
	s, dir := newTestStore(t)
	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	mustNoErr(t, s.Put("r2", rec("r2", 20)))
	mustNoErr(t, s.FlushDirty())

	// Create a new DiskStore pointing to the same directory and verify
	ds2, err := diskstore.NewDiskStore[TestRecord](filepath.Join(dir, "records"), "data")
	mustNoErr(t, err)

	got, err := ds2.Read("r1")
	mustNoErr(t, err)
	assert.Equal(t, "r1", got.ID)
	assert.Equal(t, 10, got.Value)

	got, err = ds2.Read("r2")
	mustNoErr(t, err)
	assert.Equal(t, "r2", got.ID)
}

func TestFlushDeleteRemovesFromDisk(t *testing.T) {
	s, dir := newTestStore(t)
	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	mustNoErr(t, s.FlushDirty())

	mustNoErr(t, s.Delete("r1"))
	mustNoErr(t, s.FlushDirty())

	ds2, err := diskstore.NewDiskStore[TestRecord](filepath.Join(dir, "records"), "data")
	mustNoErr(t, err)

	_, err = ds2.Read("r1")
	assert.Error(t, err) // should be gone
}

// --- LoadAll ---

func TestLoadAll(t *testing.T) {
	s, dir := newTestStore(t)
	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	mustNoErr(t, s.Put("r2", rec("r2", 20)))
	mustNoErr(t, s.FlushDirty())

	// Create a new store pointing to the same dir and load
	ds2, err := diskstore.NewDiskStore[TestRecord](filepath.Join(dir, "records"), "data")
	mustNoErr(t, err)

	s2 := NewStore[string, TestRecord](Config[string, TestRecord]{
		Disk:       ds2,
		EntityName: "data",
		KeyToStr:   identity,
		StrToKey:   identity,
	})
	mustNoErr(t, s2.LoadAll())

	got, ok := s2.Get("r1")
	assert.True(t, ok)
	assert.Equal(t, 10, got.Value)

	got, ok = s2.Get("r2")
	assert.True(t, ok)
	assert.Equal(t, 20, got.Value)
}

func TestLoadAllEmpty(t *testing.T) {
	s, _ := newTestStore(t)
	mustNoErr(t, s.LoadAll())
	assert.Equal(t, 0, s.Len())
}

func TestLoadAllDoesNotOverwriteExisting(t *testing.T) {
	s, dir := newTestStore(t)
	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	mustNoErr(t, s.FlushDirty())

	// New store, pre-populate r1 with different value (simulates WAL replay)
	ds2, err := diskstore.NewDiskStore[TestRecord](filepath.Join(dir, "records"), "data")
	mustNoErr(t, err)

	s2 := NewStore[string, TestRecord](Config[string, TestRecord]{
		Disk:       ds2,
		EntityName: "data",
		KeyToStr:   identity,
		StrToKey:   identity,
	})
	mustNoErr(t, s2.Put("r1", rec("r1", 99)))
	mustNoErr(t, s2.LoadAll())

	got, ok := s2.Get("r1")
	assert.True(t, ok)
	assert.Equal(t, 99, got.Value) // WAL replay value preserved
}

// --- WAL integration ---

func TestWALAppendOnPut(t *testing.T) {
	dir := t.TempDir()
	w, err := wal.NewFileWAL(filepath.Join(dir, "wal"), wal.SyncNone)
	mustNoErr(t, err)

	s := NewStore[string, TestRecord](Config[string, TestRecord]{
		WAL:        w,
		EntityName: "data",
		KeyToStr:   identity,
		StrToKey:   identity,
	})

	mustNoErr(t, s.Put("r1", rec("r1", 10)))

	// Verify WAL has the entry
	var entries []wal.Entry
	err = w.Replay(func(e wal.Entry) error {
		entries = append(entries, e)
		return nil
	})
	mustNoErr(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, wal.OpPut, entries[0].Op)
	assert.Equal(t, "r1", entries[0].Key)
	assert.Equal(t, "data", entries[0].Entity)

	var r TestRecord
	mustNoErr(t, json.Unmarshal(entries[0].Data, &r))
	assert.Equal(t, 10, r.Value)

	w.Close()
}

func TestWALAppendOnDelete(t *testing.T) {
	dir := t.TempDir()
	w, err := wal.NewFileWAL(filepath.Join(dir, "wal"), wal.SyncNone)
	mustNoErr(t, err)

	s := NewStore[string, TestRecord](Config[string, TestRecord]{
		WAL:        w,
		EntityName: "data",
		KeyToStr:   identity,
		StrToKey:   identity,
	})

	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	mustNoErr(t, s.Delete("r1"))

	var entries []wal.Entry
	err = w.Replay(func(e wal.Entry) error {
		entries = append(entries, e)
		return nil
	})
	mustNoErr(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, wal.OpDelete, entries[1].Op)

	w.Close()
}

func TestWALRecovery(t *testing.T) {
	dir := t.TempDir()
	walDir := filepath.Join(dir, "wal")
	recordsDir := filepath.Join(dir, "records")

	// Write records with WAL, don't flush
	w, err := wal.NewFileWAL(walDir, wal.SyncAlways)
	mustNoErr(t, err)

	ds, err := diskstore.NewDiskStore[TestRecord](recordsDir, "data")
	mustNoErr(t, err)

	s := NewStore[string, TestRecord](Config[string, TestRecord]{
		Disk:       ds,
		WAL:        w,
		EntityName: "data",
		KeyToStr:   identity,
		StrToKey:   identity,
	})

	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	mustNoErr(t, s.Put("r2", rec("r2", 20)))
	// Don't flush - simulate crash
	w.Close()

	// Recovery: replay WAL into new store
	w2, err := wal.NewFileWAL(walDir, wal.SyncNone)
	mustNoErr(t, err)
	defer w2.Close()

	ds2, err := diskstore.NewDiskStore[TestRecord](recordsDir, "data")
	mustNoErr(t, err)

	s2 := NewStore[string, TestRecord](Config[string, TestRecord]{
		Disk:       ds2,
		WAL:        w2,
		EntityName: "data",
		KeyToStr:   identity,
		StrToKey:   identity,
	})

	// Load from disk first (empty since we never flushed)
	mustNoErr(t, s2.LoadAll())

	// Replay WAL
	err = w2.Replay(func(e wal.Entry) error {
		if e.Entity != "data" {
			return nil
		}
		switch e.Op {
		case wal.OpPut:
			var r TestRecord
			if err := json.Unmarshal(e.Data, &r); err != nil {
				return err
			}
			s2.records[s2.strToKey(e.Key)] = &r
		case wal.OpDelete:
			delete(s2.records, s2.strToKey(e.Key))
		}
		return nil
	})
	mustNoErr(t, err)

	// Verify recovered
	got, ok := s2.Get("r1")
	assert.True(t, ok)
	assert.Equal(t, 10, got.Value)

	got, ok = s2.Get("r2")
	assert.True(t, ok)
	assert.Equal(t, 20, got.Value)
}

// --- Concurrent access ---

func TestConcurrentPuts(t *testing.T) {
	s := newTestStoreNoDisk(t)

	var wg sync.WaitGroup
	n := 100
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("r%d", i)
			err := s.Put(key, rec(key, i))
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()
	assert.Equal(t, n, s.Len())
}

func TestConcurrentPutAndGet(t *testing.T) {
	s := newTestStoreNoDisk(t)

	// Seed
	for i := 0; i < 10; i++ {
		mustNoErr(t, s.Put(fmt.Sprintf("r%d", i), rec(fmt.Sprintf("r%d", i), i)))
	}

	var wg sync.WaitGroup

	// Writers
	wg.Add(50)
	for i := 0; i < 50; i++ {
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("r%d", i%10)
			s.Put(key, rec(key, i))
		}(i)
	}

	// Readers
	wg.Add(50)
	for i := 0; i < 50; i++ {
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("r%d", i%10)
			got, ok := s.Get(key)
			if ok {
				assert.Equal(t, key, got.ID)
			}
		}(i)
	}

	wg.Wait()
}

func TestConcurrentPutAndDelete(t *testing.T) {
	s := newTestStoreNoDisk(t)

	var wg sync.WaitGroup
	n := 100

	// Writers
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("r%d", i)
			s.Put(key, rec(key, i))
		}(i)
	}

	// Deleters (delete even keys)
	wg.Add(n / 2)
	for i := 0; i < n; i += 2 {
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("r%d", i)
			s.Delete(key)
		}(i)
	}

	wg.Wait()
	// State is non-deterministic due to race between puts and deletes,
	// but should not panic or corrupt
}

func TestConcurrentFilter(t *testing.T) {
	s := newTestStoreNoDisk(t)

	for i := 0; i < 50; i++ {
		mustNoErr(t, s.Put(fmt.Sprintf("r%d", i), rec(fmt.Sprintf("r%d", i), i)))
	}

	var wg sync.WaitGroup

	// Concurrent filters
	wg.Add(20)
	for i := 0; i < 20; i++ {
		go func() {
			defer wg.Done()
			result := s.Filter(func(r *TestRecord) bool { return r.Value > 25 })
			// Should be roughly 24 entries (26-49), but concurrent puts may vary
			_ = result
		}()
	}

	// Concurrent puts
	wg.Add(20)
	for i := 50; i < 70; i++ {
		go func(i int) {
			defer wg.Done()
			s.Put(fmt.Sprintf("r%d", i), rec(fmt.Sprintf("r%d", i), i))
		}(i)
	}

	wg.Wait()
}

// --- Full round-trip: Put -> Flush -> Clear -> LoadAll ---

func TestPutFlushClearLoadAll(t *testing.T) {
	s, dir := newTestStore(t)

	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	mustNoErr(t, s.Put("r2", rec("r2", 20)))
	mustNoErr(t, s.Put("r3", rec("r3", 30)))
	mustNoErr(t, s.FlushDirty())

	s.Clear()
	assert.Equal(t, 0, s.Len())

	// LoadAll does not use the WAL, so create a fresh store pointing to disk
	ds2, err := diskstore.NewDiskStore[TestRecord](filepath.Join(dir, "records"), "data")
	mustNoErr(t, err)

	s2 := NewStore[string, TestRecord](Config[string, TestRecord]{
		Disk:       ds2,
		EntityName: "data",
		KeyToStr:   identity,
		StrToKey:   identity,
	})
	mustNoErr(t, s2.LoadAll())

	assert.Equal(t, 3, s2.Len())
	got, ok := s2.Get("r1")
	assert.True(t, ok)
	assert.Equal(t, 10, got.Value)
}

// --- No disk, no WAL (pure in-memory) ---

func TestPureInMemory(t *testing.T) {
	s := newTestStoreNoDisk(t)

	mustNoErr(t, s.Put("r1", rec("r1", 10)))
	got, ok := s.Get("r1")
	assert.True(t, ok)
	assert.Equal(t, 10, got.Value)

	mustNoErr(t, s.FlushDirty()) // no-op without disk
	mustNoErr(t, s.LoadAll())    // no-op without disk
}

// --- Unlocked operations ---

func TestUnlockedOperations(t *testing.T) {
	s := newTestStoreNoDisk(t)

	s.Lock()
	mustNoErr(t, s.PutUnlocked("r1", rec("r1", 10)))
	mustNoErr(t, s.PutUnlocked("r2", rec("r2", 20)))

	got, ok := s.GetUnlocked("r1")
	assert.True(t, ok)
	assert.Equal(t, 10, got.Value)

	mustNoErr(t, s.DeleteUnlocked("r2"))
	_, ok = s.GetUnlocked("r2")
	assert.False(t, ok)

	results := s.FilterUnlocked(func(r *TestRecord) bool { return true })
	assert.Len(t, results, 1)

	count := s.CountUnlocked(func(r *TestRecord) bool { return true })
	assert.Equal(t, 1, count)
	s.Unlock()
}
