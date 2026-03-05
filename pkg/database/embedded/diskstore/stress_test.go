package diskstore

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Stress tests for DiskStore concurrency.
// Run with: go test -race -count=1 -run TestStress ./pkg/database/embedded/diskstore/

func TestStressConcurrentWritesDifferentKeys(t *testing.T) {
	// Many goroutines writing to different keys simultaneously.
	ds := newStore[SimpleRecord](t)

	const numGoroutines = 20
	const writesPerGoroutine = 100
	var wg sync.WaitGroup

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < writesPerGoroutine; i++ {
				key := fmt.Sprintf("g%d-k%d", gid, i)
				rec := SimpleRecord{ID: key, Name: fmt.Sprintf("name-%d-%d", gid, i), Age: gid*1000 + i}
				if err := ds.Write(key, rec); err != nil {
					t.Errorf("Write error: %v", err)
					return
				}
			}
		}(g)
	}

	wg.Wait()

	keys, err := ds.List()
	mustNoErr(t, err)
	assert.Equal(t, numGoroutines*writesPerGoroutine, len(keys))
}

func TestStressConcurrentWriteReadSameKey(t *testing.T) {
	// Multiple goroutines writing and reading the same key.
	// Tests atomic file writes (write-to-temp + rename).
	ds := newStore[SimpleRecord](t)

	mustNoErr(t, ds.Write("shared", SimpleRecord{ID: "shared", Name: "initial", Age: 0}))

	const numWriters = 10
	const numReaders = 10
	const iterations = 100
	var wg sync.WaitGroup

	// Writers
	for g := 0; g < numWriters; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				rec := SimpleRecord{ID: "shared", Name: fmt.Sprintf("w%d-%d", gid, i), Age: gid*1000 + i}
				ds.Write("shared", rec)
			}
		}(g)
	}

	// Readers
	for g := 0; g < numReaders; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				got, err := ds.Read("shared")
				if err == nil {
					// If read succeeds, the record must be valid JSON with correct ID
					if got.ID != "shared" {
						t.Errorf("Read got corrupted record: ID=%s", got.ID)
					}
				}
			}
		}()
	}

	wg.Wait()

	// Final read should succeed
	got, err := ds.Read("shared")
	mustNoErr(t, err)
	assert.Equal(t, "shared", got.ID)
}

func TestStressConcurrentWriteAndDelete(t *testing.T) {
	// Concurrent writes and deletes on overlapping key ranges.
	ds := newStore[SimpleRecord](t)

	const numGoroutines = 10
	const opsPerGoroutine = 100
	var wg sync.WaitGroup

	// Writers
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := fmt.Sprintf("k%d", i%30)
				ds.Write(key, SimpleRecord{ID: key, Name: "w", Age: gid})
			}
		}(g)
	}

	// Deleters
	for g := 0; g < 5; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := fmt.Sprintf("k%d", i%30)
				ds.Delete(key)
			}
		}()
	}

	wg.Wait()
	// State is non-deterministic but should not panic or corrupt
}

func TestStressConcurrentScan(t *testing.T) {
	// Scan while writes are happening. Scan must not crash or return corrupt data.
	ds := newStore[SimpleRecord](t)

	// Seed
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("k%03d", i)
		ds.Write(key, SimpleRecord{ID: key, Name: "initial", Age: i})
	}

	var wg sync.WaitGroup

	// Writers
	for g := 0; g < 5; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				key := fmt.Sprintf("k%03d", i%50)
				ds.Write(key, SimpleRecord{ID: key, Name: fmt.Sprintf("w%d", gid), Age: gid*100 + i})
			}
		}(g)
	}

	// Scanners
	for g := 0; g < 5; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				ds.Scan(func(key string, value SimpleRecord) error {
					if value.ID == "" {
						t.Errorf("Scan returned record with empty ID for key %s", key)
					}
					return nil
				})
			}
		}()
	}

	wg.Wait()
}
