package wal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mustNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

func newTestWAL(t *testing.T, syncMode SyncMode) *FileWAL {
	t.Helper()
	dir := t.TempDir()
	w, err := NewFileWAL(dir, syncMode)
	if err != nil {
		t.Fatalf("failed to create WAL: %v", err)
	}
	return w
}

func newTestWALInDir(t *testing.T, dir string, syncMode SyncMode) *FileWAL {
	t.Helper()
	w, err := NewFileWAL(dir, syncMode)
	if err != nil {
		t.Fatalf("failed to create WAL in dir %s: %v", dir, err)
	}
	return w
}

func makeEntry(op OpType, entity, key string, data []byte) Entry {
	return Entry{
		Op:        op,
		Entity:    entity,
		Key:       key,
		Data:      data,
		Timestamp: 1234567890,
	}
}

func collectEntries(t *testing.T, w WAL) []Entry {
	t.Helper()
	var entries []Entry
	err := w.Replay(func(e Entry) error {
		entries = append(entries, e)
		return nil
	})
	mustNoErr(t, err)
	return entries
}

// --- Basic append/replay ---

func TestAppendAndReplaySingle(t *testing.T) {
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	entry := makeEntry(OpPut, "processes", "abc123", []byte(`{"id":"abc123"}`))
	mustNoErr(t, w.Append(entry))

	entries := collectEntries(t, w)
	assert.Len(t, entries, 1)
	assert.Equal(t, entry.Op, entries[0].Op)
	assert.Equal(t, entry.Entity, entries[0].Entity)
	assert.Equal(t, entry.Key, entries[0].Key)
	assert.Equal(t, entry.Data, entries[0].Data)
	assert.Equal(t, entry.Timestamp, entries[0].Timestamp)
}

func TestAppendAndReplayMultiple(t *testing.T) {
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	n := 100
	for i := 0; i < n; i++ {
		entry := makeEntry(OpPut, "executors", fmt.Sprintf("key-%d", i), []byte(fmt.Sprintf(`{"i":%d}`, i)))
		entry.Timestamp = int64(i)
		mustNoErr(t, w.Append(entry))
	}

	entries := collectEntries(t, w)
	assert.Len(t, entries, n)

	for i := 0; i < n; i++ {
		assert.Equal(t, fmt.Sprintf("key-%d", i), entries[i].Key)
		assert.Equal(t, int64(i), entries[i].Timestamp)
	}
}

// --- Replay empty WAL ---

func TestReplayEmpty(t *testing.T) {
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	entries := collectEntries(t, w)
	assert.Len(t, entries, 0)
}

// --- Replay non-existent WAL file ---

func TestReplayNonExistentFile(t *testing.T) {
	dir := t.TempDir()
	w, err := NewFileWAL(dir, SyncNone)
	mustNoErr(t, err)
	w.Close()

	os.Remove(filepath.Join(dir, "wal.log"))

	w2 := newTestWALInDir(t, dir, SyncNone)
	defer w2.Close()

	entries := collectEntries(t, w2)
	assert.Len(t, entries, 0)
}

// --- Truncate ---

func TestTruncate(t *testing.T) {
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	for i := 0; i < 10; i++ {
		mustNoErr(t, w.Append(makeEntry(OpPut, "test", fmt.Sprintf("k%d", i), []byte("data"))))
	}

	mustNoErr(t, w.Truncate())

	entries := collectEntries(t, w)
	assert.Len(t, entries, 0)
}

func TestAppendAfterTruncate(t *testing.T) {
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	mustNoErr(t, w.Append(makeEntry(OpPut, "test", "before", []byte("old"))))
	mustNoErr(t, w.Truncate())
	mustNoErr(t, w.Append(makeEntry(OpPut, "test", "after", []byte("new"))))

	entries := collectEntries(t, w)
	assert.Len(t, entries, 1)
	assert.Equal(t, "after", entries[0].Key)
}

// --- Close and reopen ---

func TestCloseAndReopen(t *testing.T) {
	dir := t.TempDir()

	w1 := newTestWALInDir(t, dir, SyncAlways)
	for i := 0; i < 5; i++ {
		mustNoErr(t, w1.Append(makeEntry(OpPut, "test", fmt.Sprintf("k%d", i), []byte("data"))))
	}
	mustNoErr(t, w1.Close())

	w2 := newTestWALInDir(t, dir, SyncNone)
	defer w2.Close()

	entries := collectEntries(t, w2)
	assert.Len(t, entries, 5)
}

// --- Crash recovery: partial/truncated entry ---

func TestPartialWriteRecovery(t *testing.T) {
	dir := t.TempDir()

	w := newTestWALInDir(t, dir, SyncAlways)
	mustNoErr(t, w.Append(makeEntry(OpPut, "test", "valid1", []byte("data1"))))
	mustNoErr(t, w.Append(makeEntry(OpPut, "test", "valid2", []byte("data2"))))
	mustNoErr(t, w.Close())

	// Append garbage to simulate a partial write (torn entry)
	path := filepath.Join(dir, "wal.log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	mustNoErr(t, err)
	f.Write([]byte{0x20, 0x00, 0x00, 0x00}) // total_len = 32
	f.Write([]byte{0xFF, 0xFF})              // only 2 bytes of body (truncated)
	f.Close()

	w2 := newTestWALInDir(t, dir, SyncNone)
	defer w2.Close()

	entries := collectEntries(t, w2)
	assert.Len(t, entries, 2)
	assert.Equal(t, "valid1", entries[0].Key)
	assert.Equal(t, "valid2", entries[1].Key)
}

// --- Corrupt CRC ---

func TestCorruptCRCRecovery(t *testing.T) {
	dir := t.TempDir()

	w := newTestWALInDir(t, dir, SyncAlways)
	mustNoErr(t, w.Append(makeEntry(OpPut, "test", "good1", []byte("data1"))))
	mustNoErr(t, w.Append(makeEntry(OpPut, "test", "good2", []byte("data2"))))
	mustNoErr(t, w.Close())

	// Corrupt the CRC of the second entry
	path := filepath.Join(dir, "wal.log")
	data, err := os.ReadFile(path)
	mustNoErr(t, err)

	firstTotalLen := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	secondEntryOffset := 4 + firstTotalLen

	// Corrupt CRC byte (offset + 4 bytes into total_len header of second entry)
	data[secondEntryOffset+4] ^= 0xFF
	mustNoErr(t, os.WriteFile(path, data, 0644))

	w2 := newTestWALInDir(t, dir, SyncNone)
	defer w2.Close()

	entries := collectEntries(t, w2)
	assert.Len(t, entries, 1)
	assert.Equal(t, "good1", entries[0].Key)
}

// --- Replay callback error ---

func TestReplayCallbackError(t *testing.T) {
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	for i := 0; i < 5; i++ {
		mustNoErr(t, w.Append(makeEntry(OpPut, "test", fmt.Sprintf("k%d", i), []byte("data"))))
	}

	count := 0
	expectedErr := fmt.Errorf("stop at 3")
	err := w.Replay(func(e Entry) error {
		count++
		if count == 3 {
			return expectedErr
		}
		return nil
	})
	assert.ErrorIs(t, err, expectedErr)
	assert.Equal(t, 3, count)
}

// --- Delete entry (nil data) ---

func TestDeleteEntry(t *testing.T) {
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	mustNoErr(t, w.Append(makeEntry(OpPut, "test", "k1", []byte("data"))))
	mustNoErr(t, w.Append(makeEntry(OpDelete, "test", "k1", nil)))

	entries := collectEntries(t, w)
	assert.Len(t, entries, 2)
	assert.Equal(t, OpPut, entries[0].Op)
	assert.Equal(t, OpDelete, entries[1].Op)
	assert.Nil(t, entries[1].Data)
}

// --- Large entries ---

func TestLargeEntry(t *testing.T) {
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	bigData := []byte(strings.Repeat("x", 1024*1024)) // 1MB
	entry := makeEntry(OpPut, "test", "bigkey", bigData)
	mustNoErr(t, w.Append(entry))

	entries := collectEntries(t, w)
	assert.Len(t, entries, 1)
	assert.Equal(t, len(bigData), len(entries[0].Data))
	assert.Equal(t, bigData, entries[0].Data)
}

// --- High volume ---

func TestHighVolume(t *testing.T) {
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	n := 100000
	for i := 0; i < n; i++ {
		entry := makeEntry(OpPut, "proc", fmt.Sprintf("p%d", i), []byte(fmt.Sprintf(`%d`, i)))
		mustNoErr(t, w.Append(entry))
	}

	count := 0
	err := w.Replay(func(e Entry) error {
		count++
		return nil
	})
	mustNoErr(t, err)
	assert.Equal(t, n, count)
}

// --- Concurrent appends ---

func TestConcurrentAppends(t *testing.T) {
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	var wg sync.WaitGroup
	n := 100
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			entry := makeEntry(OpPut, "test", fmt.Sprintf("k%d", i), []byte(fmt.Sprintf(`{"i":%d}`, i)))
			err := w.Append(entry)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	entries := collectEntries(t, w)
	assert.Equal(t, n, len(entries))
}

// --- SyncAlways mode ---

func TestSyncAlwaysMode(t *testing.T) {
	dir := t.TempDir()

	w := newTestWALInDir(t, dir, SyncAlways)
	for i := 0; i < 10; i++ {
		mustNoErr(t, w.Append(makeEntry(OpPut, "test", fmt.Sprintf("k%d", i), []byte("data"))))
	}
	mustNoErr(t, w.Close())

	w2 := newTestWALInDir(t, dir, SyncNone)
	defer w2.Close()

	entries := collectEntries(t, w2)
	assert.Equal(t, 10, len(entries))
}

// --- Entry encoding edge cases ---

func TestEmptyEntityAndKey(t *testing.T) {
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	entry := makeEntry(OpPut, "", "", []byte("data"))
	mustNoErr(t, w.Append(entry))

	entries := collectEntries(t, w)
	assert.Len(t, entries, 1)
	assert.Equal(t, "", entries[0].Entity)
	assert.Equal(t, "", entries[0].Key)
}

func TestEmptyData(t *testing.T) {
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	entry := makeEntry(OpPut, "test", "k1", []byte{})
	mustNoErr(t, w.Append(entry))

	entries := collectEntries(t, w)
	assert.Len(t, entries, 1)
	assert.Empty(t, entries[0].Data)
}

func TestLongEntityAndKey(t *testing.T) {
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	longEntity := strings.Repeat("e", 1000)
	longKey := strings.Repeat("k", 10000)
	entry := makeEntry(OpPut, longEntity, longKey, []byte("data"))
	mustNoErr(t, w.Append(entry))

	entries := collectEntries(t, w)
	assert.Len(t, entries, 1)
	assert.Equal(t, longEntity, entries[0].Entity)
	assert.Equal(t, longKey, entries[0].Key)
}

// --- Mixed operations ---

func TestMixedPutAndDelete(t *testing.T) {
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	mustNoErr(t, w.Append(makeEntry(OpPut, "proc", "p1", []byte(`{"id":"p1"}`))))
	mustNoErr(t, w.Append(makeEntry(OpPut, "proc", "p2", []byte(`{"id":"p2"}`))))
	mustNoErr(t, w.Append(makeEntry(OpDelete, "proc", "p1", nil)))
	mustNoErr(t, w.Append(makeEntry(OpPut, "proc", "p3", []byte(`{"id":"p3"}`))))
	mustNoErr(t, w.Append(makeEntry(OpPut, "proc", "p1", []byte(`{"id":"p1-v2"}`))))

	entries := collectEntries(t, w)
	assert.Len(t, entries, 5)

	// Simulate applying the WAL to a map
	state := make(map[string]string)
	for _, e := range entries {
		switch e.Op {
		case OpPut:
			state[e.Key] = string(e.Data)
		case OpDelete:
			delete(state, e.Key)
		}
	}
	assert.Len(t, state, 3)
	assert.Equal(t, `{"id":"p1-v2"}`, state["p1"])
	assert.Equal(t, `{"id":"p2"}`, state["p2"])
	assert.Equal(t, `{"id":"p3"}`, state["p3"])
}

// --- Multiple truncate cycles ---

func TestMultipleTruncateCycles(t *testing.T) {
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	for cycle := 0; cycle < 5; cycle++ {
		for i := 0; i < 10; i++ {
			mustNoErr(t, w.Append(makeEntry(OpPut, "test", fmt.Sprintf("c%d-k%d", cycle, i), []byte("data"))))
		}

		entries := collectEntries(t, w)
		assert.Equal(t, 10, len(entries), "cycle %d", cycle)

		mustNoErr(t, w.Truncate())

		entries = collectEntries(t, w)
		assert.Equal(t, 0, len(entries), "cycle %d after truncate", cycle)
	}
}
