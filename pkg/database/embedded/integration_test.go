package embedded

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// TestWALFlushToDiskStore verifies the intended workflow:
// write to WAL -> flush to DiskStore -> truncate WAL -> data on disk
func TestWALFlushToDiskStore(t *testing.T) {
	dir := t.TempDir()

	w, err := wal.NewFileWAL(filepath.Join(dir, "wal"), wal.SyncNone)
	mustNoErr(t, err)
	defer w.Close()

	ds, err := diskstore.NewDiskStore[TestRecord](filepath.Join(dir, "records"), "testentity")
	mustNoErr(t, err)

	// Write records through WAL
	records := []TestRecord{
		{ID: "r1", Name: "Alice", Value: 10},
		{ID: "r2", Name: "Bob", Value: 20},
		{ID: "r3", Name: "Carol", Value: 30},
	}

	for _, rec := range records {
		data, err := json.Marshal(rec)
		mustNoErr(t, err)
		mustNoErr(t, w.Append(wal.Entry{
			Op:        wal.OpPut,
			Entity:    "testentity",
			Key:       rec.ID,
			Data:      data,
			Timestamp: 1000,
		}))
	}

	// Simulate flush: replay WAL and write each record to DiskStore
	err = w.Replay(func(e wal.Entry) error {
		if e.Op == wal.OpPut {
			var rec TestRecord
			if err := json.Unmarshal(e.Data, &rec); err != nil {
				return err
			}
			return ds.Write(e.Key, rec)
		}
		return nil
	})
	mustNoErr(t, err)

	// Truncate WAL after successful flush
	mustNoErr(t, w.Truncate())

	// Verify WAL is empty
	count := 0
	err = w.Replay(func(e wal.Entry) error {
		count++
		return nil
	})
	mustNoErr(t, err)
	assert.Equal(t, 0, count)

	// Verify DiskStore has all records
	for _, rec := range records {
		got, err := ds.Read(rec.ID)
		mustNoErr(t, err)
		assert.Equal(t, rec, got)
	}
}

// TestCrashRecovery simulates:
// 1. Write records to WAL + in-memory map
// 2. Some records flushed to disk, some only in WAL
// 3. "Crash" (discard in-memory state)
// 4. Recover: load from disk + replay WAL
func TestCrashRecovery(t *testing.T) {
	dir := t.TempDir()
	walDir := filepath.Join(dir, "wal")
	recordsDir := filepath.Join(dir, "records")

	// Phase 1: Write records, partially flush
	w, err := wal.NewFileWAL(walDir, wal.SyncAlways)
	mustNoErr(t, err)

	ds, err := diskstore.NewDiskStore[TestRecord](recordsDir, "data")
	mustNoErr(t, err)

	// In-memory state (simulates the Store cache)
	memory := make(map[string]TestRecord)

	// Write 5 records to WAL + memory
	for i := 0; i < 5; i++ {
		rec := TestRecord{ID: fmt.Sprintf("r%d", i), Name: fmt.Sprintf("Name-%d", i), Value: i * 10}
		data, err := json.Marshal(rec)
		mustNoErr(t, err)
		mustNoErr(t, w.Append(wal.Entry{
			Op:     wal.OpPut,
			Entity: "data",
			Key:    rec.ID,
			Data:   data,
		}))
		memory[rec.ID] = rec
	}

	// Flush only first 3 records to disk (simulating partial flush)
	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("r%d", i)
		mustNoErr(t, ds.Write(key, memory[key]))
	}
	// DO NOT truncate WAL (crash before truncation)

	// Close WAL (simulates crash - we lose the in-memory state)
	mustNoErr(t, w.Close())
	memory = nil

	// Phase 2: Recovery - load from disk + replay WAL
	w2, err := wal.NewFileWAL(walDir, wal.SyncNone)
	mustNoErr(t, err)
	defer w2.Close()

	ds2, err := diskstore.NewDiskStore[TestRecord](recordsDir, "data")
	mustNoErr(t, err)

	// Load existing disk records into new memory
	recovered := make(map[string]TestRecord)
	err = ds2.Scan(func(key string, value TestRecord) error {
		recovered[key] = value
		return nil
	})
	mustNoErr(t, err)

	// Replay WAL to recover unflushed records
	err = w2.Replay(func(e wal.Entry) error {
		switch e.Op {
		case wal.OpPut:
			var rec TestRecord
			if err := json.Unmarshal(e.Data, &rec); err != nil {
				return err
			}
			recovered[e.Key] = rec
		case wal.OpDelete:
			delete(recovered, e.Key)
		}
		return nil
	})
	mustNoErr(t, err)

	// All 5 records should be recovered
	assert.Len(t, recovered, 5)
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("r%d", i)
		rec, exists := recovered[key]
		assert.True(t, exists, "missing record %s", key)
		assert.Equal(t, fmt.Sprintf("Name-%d", i), rec.Name)
		assert.Equal(t, i*10, rec.Value)
	}
}

// TestCrashRecoveryWithDeletes verifies that delete operations in the WAL
// are correctly applied during recovery.
func TestCrashRecoveryWithDeletes(t *testing.T) {
	dir := t.TempDir()
	walDir := filepath.Join(dir, "wal")
	recordsDir := filepath.Join(dir, "records")

	w, err := wal.NewFileWAL(walDir, wal.SyncAlways)
	mustNoErr(t, err)

	ds, err := diskstore.NewDiskStore[TestRecord](recordsDir, "data")
	mustNoErr(t, err)

	// Write 3 records to WAL + disk (fully flushed)
	for i := 0; i < 3; i++ {
		rec := TestRecord{ID: fmt.Sprintf("r%d", i), Name: fmt.Sprintf("Name-%d", i), Value: i}
		data, err := json.Marshal(rec)
		mustNoErr(t, err)
		mustNoErr(t, w.Append(wal.Entry{Op: wal.OpPut, Entity: "data", Key: rec.ID, Data: data}))
		mustNoErr(t, ds.Write(rec.ID, rec))
	}

	// Now truncate WAL (flush complete)
	mustNoErr(t, w.Truncate())

	// Delete r1 via WAL (but don't flush the delete to disk)
	mustNoErr(t, w.Append(wal.Entry{Op: wal.OpDelete, Entity: "data", Key: "r1"}))

	// Add a new record r3 via WAL (not flushed to disk)
	newRec := TestRecord{ID: "r3", Name: "New", Value: 99}
	data, err := json.Marshal(newRec)
	mustNoErr(t, err)
	mustNoErr(t, w.Append(wal.Entry{Op: wal.OpPut, Entity: "data", Key: "r3", Data: data}))

	// "Crash"
	mustNoErr(t, w.Close())

	// Recovery
	w2, err := wal.NewFileWAL(walDir, wal.SyncNone)
	mustNoErr(t, err)
	defer w2.Close()

	ds2, err := diskstore.NewDiskStore[TestRecord](recordsDir, "data")
	mustNoErr(t, err)

	recovered := make(map[string]TestRecord)
	err = ds2.Scan(func(key string, value TestRecord) error {
		recovered[key] = value
		return nil
	})
	mustNoErr(t, err)

	// Disk has r0, r1, r2 (the delete wasn't flushed)
	assert.Len(t, recovered, 3)

	// Replay WAL
	err = w2.Replay(func(e wal.Entry) error {
		switch e.Op {
		case wal.OpPut:
			var rec TestRecord
			if err := json.Unmarshal(e.Data, &rec); err != nil {
				return err
			}
			recovered[e.Key] = rec
		case wal.OpDelete:
			delete(recovered, e.Key)
		}
		return nil
	})
	mustNoErr(t, err)

	// After replay: r0, r2, r3 (r1 deleted, r3 added)
	assert.Len(t, recovered, 3)
	_, hasR0 := recovered["r0"]
	_, hasR1 := recovered["r1"]
	_, hasR2 := recovered["r2"]
	_, hasR3 := recovered["r3"]
	assert.True(t, hasR0)
	assert.False(t, hasR1) // deleted in WAL
	assert.True(t, hasR2)
	assert.True(t, hasR3)  // added in WAL
	assert.Equal(t, 99, recovered["r3"].Value)
}

// TestFullFlushCycle tests the complete lifecycle:
// write -> flush -> truncate -> write more -> flush -> truncate
func TestFullFlushCycle(t *testing.T) {
	dir := t.TempDir()
	walDir := filepath.Join(dir, "wal")
	recordsDir := filepath.Join(dir, "records")

	w, err := wal.NewFileWAL(walDir, wal.SyncNone)
	mustNoErr(t, err)
	defer w.Close()

	ds, err := diskstore.NewDiskStore[TestRecord](recordsDir, "data")
	mustNoErr(t, err)

	memory := make(map[string]TestRecord)

	// Helper: write a record
	writeRecord := func(rec TestRecord) {
		data, err := json.Marshal(rec)
		mustNoErr(t, err)
		mustNoErr(t, w.Append(wal.Entry{Op: wal.OpPut, Entity: "data", Key: rec.ID, Data: data}))
		memory[rec.ID] = rec
	}

	// Helper: flush all dirty records to disk and truncate WAL
	flush := func() {
		for key, rec := range memory {
			mustNoErr(t, ds.Write(key, rec))
		}
		mustNoErr(t, w.Truncate())
	}

	// Cycle 1
	writeRecord(TestRecord{ID: "a", Name: "A", Value: 1})
	writeRecord(TestRecord{ID: "b", Name: "B", Value: 2})
	flush()

	// Cycle 2
	writeRecord(TestRecord{ID: "c", Name: "C", Value: 3})
	writeRecord(TestRecord{ID: "a", Name: "A-updated", Value: 10}) // overwrite
	flush()

	// Verify final state on disk
	keys, err := ds.List()
	mustNoErr(t, err)
	assert.Len(t, keys, 3) // a, b, c

	got, err := ds.Read("a")
	mustNoErr(t, err)
	assert.Equal(t, "A-updated", got.Name)
	assert.Equal(t, 10, got.Value)

	got, err = ds.Read("b")
	mustNoErr(t, err)
	assert.Equal(t, "B", got.Name)

	got, err = ds.Read("c")
	mustNoErr(t, err)
	assert.Equal(t, "C", got.Name)

	// WAL should be empty
	count := 0
	err = w.Replay(func(e wal.Entry) error {
		count++
		return nil
	})
	mustNoErr(t, err)
	assert.Equal(t, 0, count)
}

// TestRecoveryAfterMultipleFlushCycles verifies that recovery works correctly
// when there have been multiple flush+truncate cycles before the crash.
func TestRecoveryAfterMultipleFlushCycles(t *testing.T) {
	dir := t.TempDir()
	walDir := filepath.Join(dir, "wal")
	recordsDir := filepath.Join(dir, "records")

	w, err := wal.NewFileWAL(walDir, wal.SyncAlways)
	mustNoErr(t, err)

	ds, err := diskstore.NewDiskStore[TestRecord](recordsDir, "data")
	mustNoErr(t, err)

	// Cycle 1: write + flush + truncate
	for i := 0; i < 3; i++ {
		rec := TestRecord{ID: fmt.Sprintf("r%d", i), Name: fmt.Sprintf("N%d", i), Value: i}
		data, err := json.Marshal(rec)
		mustNoErr(t, err)
		mustNoErr(t, w.Append(wal.Entry{Op: wal.OpPut, Entity: "data", Key: rec.ID, Data: data}))
		mustNoErr(t, ds.Write(rec.ID, rec))
	}
	mustNoErr(t, w.Truncate())

	// Cycle 2: write more (not flushed, simulating crash before flush)
	for i := 3; i < 6; i++ {
		rec := TestRecord{ID: fmt.Sprintf("r%d", i), Name: fmt.Sprintf("N%d", i), Value: i}
		data, err := json.Marshal(rec)
		mustNoErr(t, err)
		mustNoErr(t, w.Append(wal.Entry{Op: wal.OpPut, Entity: "data", Key: rec.ID, Data: data}))
	}

	// Also update an existing record in WAL
	updated := TestRecord{ID: "r0", Name: "Updated-R0", Value: 100}
	data, err := json.Marshal(updated)
	mustNoErr(t, err)
	mustNoErr(t, w.Append(wal.Entry{Op: wal.OpPut, Entity: "data", Key: "r0", Data: data}))

	// "Crash"
	mustNoErr(t, w.Close())

	// Recovery
	w2, err := wal.NewFileWAL(walDir, wal.SyncNone)
	mustNoErr(t, err)
	defer w2.Close()

	ds2, err := diskstore.NewDiskStore[TestRecord](recordsDir, "data")
	mustNoErr(t, err)

	recovered := make(map[string]TestRecord)
	err = ds2.Scan(func(key string, value TestRecord) error {
		recovered[key] = value
		return nil
	})
	mustNoErr(t, err)
	// Disk has r0, r1, r2 from cycle 1
	assert.Len(t, recovered, 3)

	err = w2.Replay(func(e wal.Entry) error {
		if e.Op == wal.OpPut {
			var rec TestRecord
			if err := json.Unmarshal(e.Data, &rec); err != nil {
				return err
			}
			recovered[e.Key] = rec
		}
		return nil
	})
	mustNoErr(t, err)

	// Should have r0 (updated), r1, r2, r3, r4, r5
	assert.Len(t, recovered, 6)
	assert.Equal(t, "Updated-R0", recovered["r0"].Name)
	assert.Equal(t, 100, recovered["r0"].Value)
	for i := 1; i < 6; i++ {
		key := fmt.Sprintf("r%d", i)
		assert.Equal(t, fmt.Sprintf("N%d", i), recovered[key].Name)
	}
}

// TestDiskStoreIndependence verifies that multiple DiskStores in the same
// base directory with different entity names don't interfere.
func TestDiskStoreIndependence(t *testing.T) {
	dir := t.TempDir()

	ds1, err := diskstore.NewDiskStore[TestRecord](dir, "processes")
	mustNoErr(t, err)

	ds2, err := diskstore.NewDiskStore[TestRecord](dir, "executors")
	mustNoErr(t, err)

	mustNoErr(t, ds1.Write("p1", TestRecord{ID: "p1", Name: "Process", Value: 1}))
	mustNoErr(t, ds2.Write("e1", TestRecord{ID: "e1", Name: "Executor", Value: 2}))

	// Each store should only see its own records
	keys1, err := ds1.List()
	mustNoErr(t, err)
	assert.Equal(t, []string{"p1"}, keys1)

	keys2, err := ds2.List()
	mustNoErr(t, err)
	assert.Equal(t, []string{"e1"}, keys2)

	// Cross-reads should fail
	_, err = ds1.Read("e1")
	assert.Error(t, err)
	_, err = ds2.Read("p1")
	assert.Error(t, err)
}

// TestWALMultiEntityRecovery verifies that a single WAL can handle entries
// for multiple entity types and they get routed correctly during recovery.
func TestWALMultiEntityRecovery(t *testing.T) {
	dir := t.TempDir()
	walDir := filepath.Join(dir, "wal")

	w, err := wal.NewFileWAL(walDir, wal.SyncAlways)
	mustNoErr(t, err)

	// Write entries for different entity types to the same WAL
	procData, _ := json.Marshal(TestRecord{ID: "p1", Name: "Process1", Value: 10})
	execData, _ := json.Marshal(TestRecord{ID: "e1", Name: "Executor1", Value: 20})
	colData, _ := json.Marshal(TestRecord{ID: "c1", Name: "Colony1", Value: 30})

	mustNoErr(t, w.Append(wal.Entry{Op: wal.OpPut, Entity: "processes", Key: "p1", Data: procData}))
	mustNoErr(t, w.Append(wal.Entry{Op: wal.OpPut, Entity: "executors", Key: "e1", Data: execData}))
	mustNoErr(t, w.Append(wal.Entry{Op: wal.OpPut, Entity: "colonies", Key: "c1", Data: colData}))
	mustNoErr(t, w.Append(wal.Entry{Op: wal.OpDelete, Entity: "executors", Key: "e1"}))

	mustNoErr(t, w.Close())

	// Recovery: route entries by entity type
	w2, err := wal.NewFileWAL(walDir, wal.SyncNone)
	mustNoErr(t, err)
	defer w2.Close()

	entities := make(map[string]map[string]TestRecord)
	err = w2.Replay(func(e wal.Entry) error {
		if entities[e.Entity] == nil {
			entities[e.Entity] = make(map[string]TestRecord)
		}
		switch e.Op {
		case wal.OpPut:
			var rec TestRecord
			if err := json.Unmarshal(e.Data, &rec); err != nil {
				return err
			}
			entities[e.Entity][e.Key] = rec
		case wal.OpDelete:
			delete(entities[e.Entity], e.Key)
		}
		return nil
	})
	mustNoErr(t, err)

	assert.Len(t, entities["processes"], 1)
	assert.Len(t, entities["executors"], 0) // deleted
	assert.Len(t, entities["colonies"], 1)
	assert.Equal(t, "Process1", entities["processes"]["p1"].Name)
	assert.Equal(t, "Colony1", entities["colonies"]["c1"].Name)
}

// TestEmptyDirectoryRecovery verifies that recovery from a completely empty
// data directory works (fresh start scenario).
func TestEmptyDirectoryRecovery(t *testing.T) {
	dir := t.TempDir()
	walDir := filepath.Join(dir, "wal")
	recordsDir := filepath.Join(dir, "records")

	// Create empty WAL and DiskStore
	w, err := wal.NewFileWAL(walDir, wal.SyncNone)
	mustNoErr(t, err)
	defer w.Close()

	ds, err := diskstore.NewDiskStore[TestRecord](recordsDir, "data")
	mustNoErr(t, err)

	// No entries anywhere
	count := 0
	err = w.Replay(func(e wal.Entry) error {
		count++
		return nil
	})
	mustNoErr(t, err)
	assert.Equal(t, 0, count)

	keys, err := ds.List()
	mustNoErr(t, err)
	assert.Empty(t, keys)

	// Now write something - should work fine
	rec := TestRecord{ID: "first", Name: "First", Value: 1}
	data, _ := json.Marshal(rec)
	mustNoErr(t, w.Append(wal.Entry{Op: wal.OpPut, Entity: "data", Key: "first", Data: data}))

	entries := 0
	err = w.Replay(func(e wal.Entry) error {
		entries++
		return nil
	})
	mustNoErr(t, err)
	assert.Equal(t, 1, entries)
}

// TestDiskStoreCleanup verifies that deleting records from DiskStore
// actually removes the files from disk.
func TestDiskStoreCleanup(t *testing.T) {
	dir := t.TempDir()

	ds, err := diskstore.NewDiskStore[TestRecord](dir, "data")
	mustNoErr(t, err)

	mustNoErr(t, ds.Write("k1", TestRecord{ID: "k1", Name: "One", Value: 1}))
	mustNoErr(t, ds.Write("k2", TestRecord{ID: "k2", Name: "Two", Value: 2}))

	// Verify files exist
	dataDir := filepath.Join(dir, "data")
	entries, err := os.ReadDir(dataDir)
	mustNoErr(t, err)
	jsonCount := 0
	for _, e := range entries {
		if !e.IsDir() {
			jsonCount++
		}
	}
	assert.Equal(t, 2, jsonCount)

	// Delete one
	mustNoErr(t, ds.Delete("k1"))

	entries, err = os.ReadDir(dataDir)
	mustNoErr(t, err)
	jsonCount = 0
	for _, e := range entries {
		if !e.IsDir() {
			jsonCount++
		}
	}
	assert.Equal(t, 1, jsonCount)
}
