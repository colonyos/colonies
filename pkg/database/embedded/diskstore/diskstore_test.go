package diskstore

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test types

type SimpleRecord struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type NestedRecord struct {
	ID       string            `json:"id"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
	Inner    *SimpleRecord     `json:"inner"`
}

func tempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func newStore[V any](t *testing.T) *DiskStore[V] {
	t.Helper()
	ds, err := NewDiskStore[V](tempDir(t), "test")
	if err != nil {
		t.Fatalf("failed to create DiskStore: %v", err)
	}
	return ds
}

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

// --- Write/Read round-trip tests ---

func TestWriteReadSimple(t *testing.T) {
	ds := newStore[SimpleRecord](t)
	rec := SimpleRecord{ID: "1", Name: "Alice", Age: 30}

	mustNoErr(t, ds.Write("1", rec))

	got, err := ds.Read("1")
	mustNoErr(t, err)
	assert.Equal(t, rec, got)
}

func TestWriteReadString(t *testing.T) {
	ds := newStore[string](t)

	mustNoErr(t, ds.Write("key1", "hello world"))

	got, err := ds.Read("key1")
	mustNoErr(t, err)
	assert.Equal(t, "hello world", got)
}

func TestWriteReadNested(t *testing.T) {
	ds := newStore[NestedRecord](t)
	rec := NestedRecord{
		ID:       "n1",
		Tags:     []string{"tag1", "tag2", "tag3"},
		Metadata: map[string]string{"env": "prod", "region": "us-east"},
		Inner:    &SimpleRecord{ID: "inner1", Name: "Bob", Age: 25},
	}

	mustNoErr(t, ds.Write("n1", rec))

	got, err := ds.Read("n1")
	mustNoErr(t, err)
	assert.Equal(t, rec, got)
}

func TestWriteOverwrite(t *testing.T) {
	ds := newStore[SimpleRecord](t)
	rec1 := SimpleRecord{ID: "1", Name: "Alice", Age: 30}
	rec2 := SimpleRecord{ID: "1", Name: "Alice", Age: 31}

	mustNoErr(t, ds.Write("1", rec1))
	mustNoErr(t, ds.Write("1", rec2))

	got, err := ds.Read("1")
	mustNoErr(t, err)
	assert.Equal(t, 31, got.Age)
}

// --- Special characters in keys ---

func TestSpecialCharacterKeys(t *testing.T) {
	ds := newStore[string](t)

	keys := []string{
		"simple",
		"with spaces",
		"with/slash",
		"with:colon",
		"with.dots.many",
		"colony:user:name",
		"path/to/resource",
		"special*chars?here",
	}

	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			mustNoErr(t, ds.Write(key, "value-for-"+key))

			got, err := ds.Read(key)
			mustNoErr(t, err)
			assert.Equal(t, "value-for-"+key, got)
		})
	}
}

func TestUnicodeKeys(t *testing.T) {
	ds := newStore[string](t)

	mustNoErr(t, ds.Write("key-unicode", "value"))

	got, err := ds.Read("key-unicode")
	mustNoErr(t, err)
	assert.Equal(t, "value", got)
}

// --- Delete tests ---

func TestDeleteExisting(t *testing.T) {
	ds := newStore[string](t)

	mustNoErr(t, ds.Write("k1", "v1"))

	mustNoErr(t, ds.Delete("k1"))

	_, err := ds.Read("k1")
	mustErr(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteNonExistent(t *testing.T) {
	ds := newStore[string](t)

	err := ds.Delete("does-not-exist")
	mustErr(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteThenWrite(t *testing.T) {
	ds := newStore[string](t)

	mustNoErr(t, ds.Write("k1", "v1"))
	mustNoErr(t, ds.Delete("k1"))
	mustNoErr(t, ds.Write("k1", "v2"))

	got, err := ds.Read("k1")
	mustNoErr(t, err)
	assert.Equal(t, "v2", got)
}

// --- List tests ---

func TestListEmpty(t *testing.T) {
	ds := newStore[string](t)

	keys, err := ds.List()
	mustNoErr(t, err)
	assert.Empty(t, keys)
}

func TestListReturnsAllKeys(t *testing.T) {
	ds := newStore[string](t)

	mustNoErr(t, ds.Write("c", "v"))
	mustNoErr(t, ds.Write("a", "v"))
	mustNoErr(t, ds.Write("b", "v"))

	keys, err := ds.List()
	mustNoErr(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, keys)
}

func TestListAfterDelete(t *testing.T) {
	ds := newStore[string](t)

	mustNoErr(t, ds.Write("a", "v"))
	mustNoErr(t, ds.Write("b", "v"))
	mustNoErr(t, ds.Write("c", "v"))
	mustNoErr(t, ds.Delete("b"))

	keys, err := ds.List()
	mustNoErr(t, err)
	assert.Equal(t, []string{"a", "c"}, keys)
}

func TestListIgnoresTempFiles(t *testing.T) {
	ds := newStore[string](t)

	mustNoErr(t, ds.Write("real", "v"))

	// Manually create a temp file that should be ignored
	tmpPath := filepath.Join(ds.baseDir, ".tmp-leftover.json")
	mustNoErr(t, os.WriteFile(tmpPath, []byte("junk"), 0644))

	keys, err := ds.List()
	mustNoErr(t, err)
	assert.Equal(t, []string{"real"}, keys)
}

func TestListWithSpecialCharKeys(t *testing.T) {
	ds := newStore[string](t)

	mustNoErr(t, ds.Write("colony:user", "v1"))
	mustNoErr(t, ds.Write("path/to/thing", "v2"))

	keys, err := ds.List()
	mustNoErr(t, err)
	assert.Contains(t, keys, "colony:user")
	assert.Contains(t, keys, "path/to/thing")
}

// --- Scan tests ---

func TestScanEmpty(t *testing.T) {
	ds := newStore[string](t)

	count := 0
	err := ds.Scan(func(key string, value string) error {
		count++
		return nil
	})
	mustNoErr(t, err)
	assert.Equal(t, 0, count)
}

func TestScanAllRecords(t *testing.T) {
	ds := newStore[SimpleRecord](t)

	records := map[string]SimpleRecord{
		"1": {ID: "1", Name: "Alice", Age: 30},
		"2": {ID: "2", Name: "Bob", Age: 25},
		"3": {ID: "3", Name: "Carol", Age: 35},
	}

	for k, v := range records {
		mustNoErr(t, ds.Write(k, v))
	}

	got := make(map[string]SimpleRecord)
	err := ds.Scan(func(key string, value SimpleRecord) error {
		got[key] = value
		return nil
	})
	mustNoErr(t, err)
	assert.Equal(t, records, got)
}

func TestScanSortedOrder(t *testing.T) {
	ds := newStore[string](t)

	mustNoErr(t, ds.Write("c", "3"))
	mustNoErr(t, ds.Write("a", "1"))
	mustNoErr(t, ds.Write("b", "2"))

	var order []string
	err := ds.Scan(func(key string, value string) error {
		order = append(order, key)
		return nil
	})
	mustNoErr(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, order)
}

func TestScanCallbackError(t *testing.T) {
	ds := newStore[string](t)

	mustNoErr(t, ds.Write("a", "1"))
	mustNoErr(t, ds.Write("b", "2"))
	mustNoErr(t, ds.Write("c", "3"))

	expectedErr := errors.New("stop here")
	count := 0
	err := ds.Scan(func(key string, value string) error {
		count++
		if count == 2 {
			return expectedErr
		}
		return nil
	})
	assert.ErrorIs(t, err, expectedErr)
	assert.Equal(t, 2, count)
}

// --- Read non-existent ---

func TestReadNonExistent(t *testing.T) {
	ds := newStore[string](t)

	_, err := ds.Read("nope")
	mustErr(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// --- Corrupt file handling ---

func TestReadCorruptJSON(t *testing.T) {
	ds := newStore[SimpleRecord](t)

	path := ds.filePath("corrupt")
	mustNoErr(t, os.WriteFile(path, []byte("{invalid json"), 0644))

	_, err := ds.Read("corrupt")
	mustErr(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestReadTruncatedJSON(t *testing.T) {
	ds := newStore[SimpleRecord](t)

	path := ds.filePath("truncated")
	mustNoErr(t, os.WriteFile(path, []byte(`{"id":"1","name":"Al`), 0644))

	_, err := ds.Read("truncated")
	mustErr(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestReadEmptyFile(t *testing.T) {
	ds := newStore[SimpleRecord](t)

	path := ds.filePath("empty")
	mustNoErr(t, os.WriteFile(path, []byte{}, 0644))

	_, err := ds.Read("empty")
	mustErr(t, err)
}

// --- Large records ---

func TestLargeRecord(t *testing.T) {
	ds := newStore[SimpleRecord](t)

	bigName := strings.Repeat("x", 1024*1024) // ~1MB
	rec := SimpleRecord{ID: "big", Name: bigName, Age: 1}

	mustNoErr(t, ds.Write("big", rec))

	got, err := ds.Read("big")
	mustNoErr(t, err)
	assert.Equal(t, rec.ID, got.ID)
	assert.Equal(t, len(bigName), len(got.Name))
	assert.Equal(t, rec.Age, got.Age)
}

// --- Concurrent access ---

func TestConcurrentWritesDifferentKeys(t *testing.T) {
	ds := newStore[SimpleRecord](t)

	var wg sync.WaitGroup
	n := 100
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			rec := SimpleRecord{ID: key, Name: fmt.Sprintf("name-%d", i), Age: i}
			err := ds.Write(key, rec)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	keys, err := ds.List()
	mustNoErr(t, err)
	assert.Equal(t, n, len(keys))

	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key-%d", i)
		got, err := ds.Read(key)
		mustNoErr(t, err)
		assert.Equal(t, key, got.ID)
	}
}

func TestConcurrentWriteAndReadSameKey(t *testing.T) {
	ds := newStore[SimpleRecord](t)

	mustNoErr(t, ds.Write("shared", SimpleRecord{ID: "shared", Name: "initial", Age: 0}))

	var wg sync.WaitGroup
	n := 50

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			rec := SimpleRecord{ID: "shared", Name: fmt.Sprintf("writer-%d", i), Age: i}
			err := ds.Write("shared", rec)
			assert.NoError(t, err)
		}(i)
	}

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			got, err := ds.Read("shared")
			if err == nil {
				assert.Equal(t, "shared", got.ID)
			}
		}()
	}

	wg.Wait()

	got, err := ds.Read("shared")
	mustNoErr(t, err)
	assert.Equal(t, "shared", got.ID)
}

// --- Write to read-only directory ---

func TestWriteReadOnlyDirectory(t *testing.T) {
	dir := t.TempDir()
	entityDir := filepath.Join(dir, "readonly")
	mustNoErr(t, os.MkdirAll(entityDir, 0755))

	ds, err := NewDiskStore[string](dir, "readonly")
	mustNoErr(t, err)

	mustNoErr(t, os.Chmod(entityDir, 0444))
	defer os.Chmod(entityDir, 0755)

	err = ds.Write("test", "value")
	mustErr(t, err)
}

// --- NewDiskStore creates directory ---

func TestNewDiskStoreCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	entityPath := filepath.Join(dir, "newentity")

	_, err := NewDiskStore[string](dir, "newentity")
	mustNoErr(t, err)

	info, err := os.Stat(entityPath)
	mustNoErr(t, err)
	assert.True(t, info.IsDir())
}

// --- Multiple records lifecycle ---

func TestFullLifecycle(t *testing.T) {
	ds := newStore[SimpleRecord](t)

	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("rec-%d", i)
		rec := SimpleRecord{ID: key, Name: fmt.Sprintf("Name-%d", i), Age: 20 + i}
		mustNoErr(t, ds.Write(key, rec))
	}

	keys, err := ds.List()
	mustNoErr(t, err)
	assert.Len(t, keys, 10)

	mustNoErr(t, ds.Delete("rec-3"))
	mustNoErr(t, ds.Delete("rec-7"))

	keys, err = ds.List()
	mustNoErr(t, err)
	assert.Len(t, keys, 8)

	mustNoErr(t, ds.Write("rec-5", SimpleRecord{ID: "rec-5", Name: "Updated", Age: 99}))

	found := make(map[string]SimpleRecord)
	err = ds.Scan(func(key string, value SimpleRecord) error {
		found[key] = value
		return nil
	})
	mustNoErr(t, err)
	assert.Len(t, found, 8)
	assert.Equal(t, "Updated", found["rec-5"].Name)
	assert.Equal(t, 99, found["rec-5"].Age)

	_, exists := found["rec-3"]
	assert.False(t, exists)
	_, exists = found["rec-7"]
	assert.False(t, exists)
}
