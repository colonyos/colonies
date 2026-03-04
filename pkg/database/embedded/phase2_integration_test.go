package embedded

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/database/embedded/diskstore"
	"github.com/colonyos/colonies/pkg/database/embedded/flusher"
	"github.com/colonyos/colonies/pkg/database/embedded/index"
	"github.com/colonyos/colonies/pkg/database/embedded/store"
	"github.com/colonyos/colonies/pkg/database/embedded/wal"
	"github.com/stretchr/testify/assert"
)

type P2Record struct {
	ID        string `json:"id"`
	Colony    string `json:"colony"`
	State     int    `json:"state"`
	Priority  int64  `json:"priority"`
	Value     int    `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

func identity2(s string) string { return s }

// TestFullLifecycleP2 tests: Initialize -> Put -> indexes updated -> flusher writes
// to disk -> Stop -> Reinitialize from disk -> indexes rebuilt -> queries return same results
func TestFullLifecycleP2(t *testing.T) {
	dir := t.TempDir()

	// --- Phase A: Create, populate, let flusher persist ---

	w, err := wal.NewFileWAL(filepath.Join(dir, "wal"), wal.SyncAlways)
	mustNoErr(t, err)

	ds, err := diskstore.NewDiskStore[P2Record](filepath.Join(dir, "records"), "data")
	mustNoErr(t, err)

	s := store.NewStore[string, P2Record](store.Config[string, P2Record]{
		Disk:       ds,
		WAL:        w,
		EntityName: "data",
		KeyToStr:   identity2,
		StrToKey:   identity2,
	})

	// Secondary indexes
	colonyIdx := index.NewMapIndex[string, string]()
	stateIdx := index.NewCompoundIndex()

	// Helper to add record with index maintenance
	addRecord := func(r *P2Record) {
		mustNoErr(t, s.Put(r.ID, r))
		colonyIdx.Add(r.ID, r.Colony)
		stateIdx.Add(r.Colony, r.State, index.IndexEntry[string]{
			SortKey:    r.Priority,
			PrimaryKey: r.ID,
		})
	}

	// Populate
	addRecord(&P2Record{ID: "p1", Colony: "col-a", State: 0, Priority: 100, Value: 10})
	addRecord(&P2Record{ID: "p2", Colony: "col-a", State: 0, Priority: 50, Value: 20})
	addRecord(&P2Record{ID: "p3", Colony: "col-a", State: 1, Priority: 200, Value: 30})
	addRecord(&P2Record{ID: "p4", Colony: "col-b", State: 0, Priority: 10, Value: 40})

	// Start flusher
	f := flusher.NewFlusher(50 * time.Millisecond)
	f.AddStore(s)
	f.Start()

	// Wait for flush
	time.Sleep(200 * time.Millisecond)
	f.Stop()

	// Verify indexes before shutdown
	var keys []string
	stateIdx.AscendFirst("col-a", 0, 10, func(e index.IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	assert.Equal(t, []string{"p2", "p1"}, keys) // sorted by priority: 50, 100

	// Truncate WAL since everything is flushed
	mustNoErr(t, w.Truncate())
	mustNoErr(t, w.Close())

	// --- Phase B: Reopen from disk, rebuild indexes ---

	w2, err := wal.NewFileWAL(filepath.Join(dir, "wal"), wal.SyncNone)
	mustNoErr(t, err)
	defer w2.Close()

	ds2, err := diskstore.NewDiskStore[P2Record](filepath.Join(dir, "records"), "data")
	mustNoErr(t, err)

	s2 := store.NewStore[string, P2Record](store.Config[string, P2Record]{
		Disk:       ds2,
		WAL:        w2,
		EntityName: "data",
		KeyToStr:   identity2,
		StrToKey:   identity2,
	})

	mustNoErr(t, s2.LoadAll())

	// Rebuild indexes from loaded data
	colonyIdx2 := index.NewMapIndex[string, string]()
	stateIdx2 := index.NewCompoundIndex()

	for _, r := range s2.All() {
		colonyIdx2.Add(r.ID, r.Colony)
		stateIdx2.Add(r.Colony, r.State, index.IndexEntry[string]{
			SortKey:    r.Priority,
			PrimaryKey: r.ID,
		})
	}

	// Verify same query results
	var keys2 []string
	stateIdx2.AscendFirst("col-a", 0, 10, func(e index.IndexEntry[string]) bool {
		keys2 = append(keys2, e.PrimaryKey)
		return true
	})
	assert.Equal(t, []string{"p2", "p1"}, keys2)

	assert.Equal(t, 4, s2.Len())

	got, ok := s2.Get("p3")
	assert.True(t, ok)
	assert.Equal(t, 30, got.Value)
	assert.Equal(t, 1, got.State)
}

// TestCrashRecoveryP2 simulates: Put records -> crash before flush ->
// restart -> WAL replay -> indexes rebuilt -> all records recovered
func TestCrashRecoveryP2(t *testing.T) {
	dir := t.TempDir()

	// Write records with WAL but don't flush to disk
	w, err := wal.NewFileWAL(filepath.Join(dir, "wal"), wal.SyncAlways)
	mustNoErr(t, err)

	s := store.NewStore[string, P2Record](store.Config[string, P2Record]{
		WAL:        w,
		EntityName: "data",
		KeyToStr:   identity2,
		StrToKey:   identity2,
	})

	mustNoErr(t, s.Put("r1", &P2Record{ID: "r1", Colony: "col-a", State: 0, Priority: 100, Value: 10}))
	mustNoErr(t, s.Put("r2", &P2Record{ID: "r2", Colony: "col-a", State: 0, Priority: 50, Value: 20}))
	mustNoErr(t, s.Put("r3", &P2Record{ID: "r3", Colony: "col-b", State: 1, Priority: 200, Value: 30}))

	// "Crash" - close WAL, lose in-memory state
	mustNoErr(t, w.Close())

	// Recovery
	w2, err := wal.NewFileWAL(filepath.Join(dir, "wal"), wal.SyncNone)
	mustNoErr(t, err)
	defer w2.Close()

	s2 := store.NewStore[string, P2Record](store.Config[string, P2Record]{
		WAL:        w2,
		EntityName: "data",
		KeyToStr:   identity2,
		StrToKey:   identity2,
	})

	// Replay WAL - collect entries first to avoid deadlock (Replay holds WAL mutex)
	var entries []wal.Entry
	err = w2.Replay(func(e wal.Entry) error {
		if e.Entity == "data" {
			entries = append(entries, e)
		}
		return nil
	})
	mustNoErr(t, err)

	// Apply replayed entries to store
	s2.Lock()
	for _, e := range entries {
		switch e.Op {
		case wal.OpPut:
			var r P2Record
			if err := json.Unmarshal(e.Data, &r); err != nil {
				t.Fatalf("unmarshal during replay: %v", err)
			}
			s2.ReplayPut(e.Key, &r)
		case wal.OpDelete:
			s2.ReplayDelete(e.Key)
		}
	}
	s2.Unlock()

	assert.Equal(t, 3, s2.Len())

	// Rebuild indexes
	stateIdx := index.NewCompoundIndex()
	for _, r := range s2.All() {
		stateIdx.Add(r.Colony, r.State, index.IndexEntry[string]{
			SortKey:    r.Priority,
			PrimaryKey: r.ID,
		})
	}

	// Verify scheduler query works
	var keys []string
	stateIdx.AscendFirst("col-a", 0, 1, func(e index.IndexEntry[string]) bool {
		keys = append(keys, e.PrimaryKey)
		return true
	})
	assert.Equal(t, []string{"r2"}, keys) // lowest priority = 50
}

// TestStoreWithIndexConsistency tests that index stays consistent after
// a sequence of puts and deletes.
func TestStoreWithIndexConsistency(t *testing.T) {
	s := store.NewStore[string, P2Record](store.Config[string, P2Record]{
		EntityName: "data",
		KeyToStr:   identity2,
		StrToKey:   identity2,
	})

	colonyIdx := index.NewMapIndex[string, string]()
	stateIdx := index.NewCompoundIndex()

	add := func(r *P2Record) {
		mustNoErr(t, s.Put(r.ID, r))
		colonyIdx.Add(r.ID, r.Colony)
		stateIdx.Add(r.Colony, r.State, index.IndexEntry[string]{
			SortKey:    r.Priority,
			PrimaryKey: r.ID,
		})
	}

	remove := func(id string) {
		r, ok := s.Get(id)
		if !ok {
			return
		}
		colonyIdx.Remove(id, r.Colony)
		stateIdx.Remove(r.Colony, r.State, index.IndexEntry[string]{
			SortKey:    r.Priority,
			PrimaryKey: r.ID,
		})
		mustNoErr(t, s.Delete(id))
	}

	// Add records
	for i := 0; i < 20; i++ {
		colony := fmt.Sprintf("col-%d", i%3)
		state := i % 2
		add(&P2Record{
			ID:       fmt.Sprintf("p%d", i),
			Colony:   colony,
			State:    state,
			Priority: int64(i * 10),
			Value:    i,
		})
	}

	assert.Equal(t, 20, s.Len())

	// Delete some
	remove("p3")
	remove("p7")
	remove("p15")

	assert.Equal(t, 17, s.Len())

	// Verify colony index
	col0 := colonyIdx.Lookup("col-0")
	for _, id := range col0 {
		r, ok := s.Get(id)
		assert.True(t, ok, "index references non-existent record %s", id)
		assert.Equal(t, "col-0", r.Colony)
	}

	// Verify compound index: query col-0, state=0
	var found []string
	stateIdx.AscendFirst("col-0", 0, 100, func(e index.IndexEntry[string]) bool {
		found = append(found, e.PrimaryKey)
		return true
	})
	// Every found key should exist in the store with correct colony and state
	for _, id := range found {
		r, ok := s.Get(id)
		assert.True(t, ok, "compound index references non-existent record %s", id)
		assert.Equal(t, "col-0", r.Colony)
		assert.Equal(t, 0, r.State)
	}
}

// TestFlusherWithStoreAndWAL tests the complete flusher integration:
// writes -> flusher persists -> WAL truncated -> restart -> data intact
func TestFlusherWithStoreAndWAL(t *testing.T) {
	dir := t.TempDir()

	w, err := wal.NewFileWAL(filepath.Join(dir, "wal"), wal.SyncAlways)
	mustNoErr(t, err)

	ds, err := diskstore.NewDiskStore[P2Record](filepath.Join(dir, "records"), "data")
	mustNoErr(t, err)

	s := store.NewStore[string, P2Record](store.Config[string, P2Record]{
		Disk:       ds,
		WAL:        w,
		EntityName: "data",
		KeyToStr:   identity2,
		StrToKey:   identity2,
	})

	f := flusher.NewFlusher(30 * time.Millisecond)
	f.AddStore(s)
	f.Start()

	// Write records
	for i := 0; i < 50; i++ {
		mustNoErr(t, s.Put(fmt.Sprintf("r%d", i), &P2Record{
			ID:    fmt.Sprintf("r%d", i),
			Value: i,
		}))
	}

	// Wait for flusher to persist
	time.Sleep(200 * time.Millisecond)

	// Stop flusher (triggers final flush)
	f.Stop()

	// Truncate WAL
	mustNoErr(t, w.Truncate())
	mustNoErr(t, w.Close())

	// Reopen from disk only (WAL is empty)
	ds2, err := diskstore.NewDiskStore[P2Record](filepath.Join(dir, "records"), "data")
	mustNoErr(t, err)

	s2 := store.NewStore[string, P2Record](store.Config[string, P2Record]{
		Disk:       ds2,
		EntityName: "data",
		KeyToStr:   identity2,
		StrToKey:   identity2,
	})
	mustNoErr(t, s2.LoadAll())

	assert.Equal(t, 50, s2.Len())

	for i := 0; i < 50; i++ {
		got, ok := s2.Get(fmt.Sprintf("r%d", i))
		assert.True(t, ok, "missing record r%d", i)
		if ok {
			assert.Equal(t, i, got.Value)
		}
	}
}
