package wal

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Stress tests for WAL concurrency.
// Run with: go test -race -count=1 -run TestStress ./pkg/database/embedded/wal/

func TestStressConcurrentAppends(t *testing.T) {
	// Many goroutines appending simultaneously. All entries must survive.
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	const numGoroutines = 20
	const entriesPerGoroutine = 500
	var wg sync.WaitGroup

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < entriesPerGoroutine; i++ {
				entry := makeEntry(OpPut, "test",
					fmt.Sprintf("g%d-k%d", gid, i),
					[]byte(fmt.Sprintf(`{"g":%d,"i":%d}`, gid, i)))
				if err := w.Append(entry); err != nil {
					t.Errorf("Append error from goroutine %d: %v", gid, err)
					return
				}
			}
		}(g)
	}

	wg.Wait()

	entries := collectEntries(t, w)
	assert.Equal(t, numGoroutines*entriesPerGoroutine, len(entries),
		"All appended entries must survive")
}

func TestStressConcurrentAppendsMixedOps(t *testing.T) {
	// Mixed Put and Delete appends from multiple goroutines.
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	const numGoroutines = 10
	const opsPerGoroutine = 300
	var totalAppends int64
	var wg sync.WaitGroup

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				var entry Entry
				if i%3 == 0 {
					entry = makeEntry(OpDelete, "test",
						fmt.Sprintf("g%d-k%d", gid, i), nil)
				} else {
					entry = makeEntry(OpPut, "test",
						fmt.Sprintf("g%d-k%d", gid, i),
						[]byte(fmt.Sprintf(`%d`, i)))
				}
				if err := w.Append(entry); err != nil {
					t.Errorf("Append error: %v", err)
					return
				}
				atomic.AddInt64(&totalAppends, 1)
			}
		}(g)
	}

	wg.Wait()

	entries := collectEntries(t, w)
	assert.Equal(t, int(atomic.LoadInt64(&totalAppends)), len(entries))
}

func TestStressConcurrentAppendsWithSyncAlways(t *testing.T) {
	// SyncAlways calls fsync on every append. Verify no corruption under concurrency.
	w := newTestWAL(t, SyncAlways)
	defer w.Close()

	const numGoroutines = 10
	const entriesPerGoroutine = 50 // fewer since fsync is slow
	var wg sync.WaitGroup

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < entriesPerGoroutine; i++ {
				entry := makeEntry(OpPut, "sync",
					fmt.Sprintf("g%d-k%d", gid, i),
					[]byte("data"))
				if err := w.Append(entry); err != nil {
					t.Errorf("Append error: %v", err)
					return
				}
			}
		}(g)
	}

	wg.Wait()

	entries := collectEntries(t, w)
	assert.Equal(t, numGoroutines*entriesPerGoroutine, len(entries))
}

func TestStressAppendAndReplayConcurrent(t *testing.T) {
	// Appends while replaying. Replay should see a consistent prefix.
	w := newTestWAL(t, SyncNone)
	defer w.Close()

	// Pre-seed some entries
	for i := 0; i < 100; i++ {
		w.Append(makeEntry(OpPut, "test", fmt.Sprintf("seed-%d", i), []byte("data")))
	}

	var wg sync.WaitGroup

	// Appenders
	for g := 0; g < 5; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				w.Append(makeEntry(OpPut, "test",
					fmt.Sprintf("new-g%d-k%d", gid, i), []byte("data")))
			}
		}(g)
	}

	// Readers (replaying)
	for g := 0; g < 3; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				var count int
				w.Replay(func(e Entry) error {
					count++
					return nil
				})
				// Should see at least the seeded entries
				if count < 100 {
					t.Errorf("Replay saw only %d entries, expected at least 100", count)
				}
			}
		}()
	}

	wg.Wait()
}
