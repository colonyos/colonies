package store

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Stress tests for Store concurrency.
// Run with: go test -race -count=1 -run TestStress ./pkg/database/embedded/store/

func TestStressConcurrentPutGetDelete(t *testing.T) {
	// Heavy concurrent Put/Get/Delete on overlapping keys.
	s := newTestStoreNoDisk(t)

	const numGoroutines = 20
	const opsPerGoroutine = 500
	var wg sync.WaitGroup

	// Writers
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := fmt.Sprintf("k%d", i%50) // overlap across goroutines
				s.Put(key, rec(key, gid*1000+i))
			}
		}(g)
	}

	// Readers
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := fmt.Sprintf("k%d", i%50)
				if v, ok := s.Get(key); ok {
					_ = v.Value // access to detect races
				}
			}
		}()
	}

	// Deleters
	for g := 0; g < 5; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := fmt.Sprintf("k%d", i%50)
				s.Delete(key)
			}
		}()
	}

	wg.Wait()
}

func TestStressForEachDuringMutations(t *testing.T) {
	// ForEach holds RLock. Ensure concurrent Puts (which need write Lock)
	// don't deadlock or corrupt.
	s := newTestStoreNoDisk(t)

	// Seed data
	for i := 0; i < 100; i++ {
		s.Put(fmt.Sprintf("k%d", i), rec(fmt.Sprintf("k%d", i), i))
	}

	var wg sync.WaitGroup

	// ForEach readers
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for iter := 0; iter < 100; iter++ {
				count := 0
				s.ForEach(func(k string, v *TestRecord) {
					count++
					_ = v.Name // access
				})
			}
		}()
	}

	// Concurrent writers
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				key := fmt.Sprintf("new-%d-%d", gid, i)
				s.Put(key, rec(key, i))
			}
		}(g)
	}

	// Concurrent deleters
	for g := 0; g < 5; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				s.Delete(fmt.Sprintf("k%d", i))
			}
		}()
	}

	wg.Wait()
}

func TestStressFlushDuringMutations(t *testing.T) {
	// FlushDirty snapshots dirty set under Lock then does disk I/O outside lock.
	// Ensure concurrent Put/Delete + FlushDirty don't race.
	s, _ := newTestStore(t)

	var wg sync.WaitGroup

	// Writers
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				key := fmt.Sprintf("w%d-%d", gid, i)
				s.Put(key, rec(key, i))
			}
		}(g)
	}

	// Flushers
	for g := 0; g < 3; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				s.FlushDirty()
				time.Sleep(time.Millisecond)
			}
		}()
	}

	// Readers
	for g := 0; g < 5; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				s.Len()
				s.DirtyCount()
				s.DeletedCount()
			}
		}()
	}

	wg.Wait()

	// Final flush to ensure no data loss
	err := s.FlushDirty()
	assert.NoError(t, err)
	assert.Equal(t, 0, s.DirtyCount())
}

func TestStressLockUnlockWithRegularOps(t *testing.T) {
	// Lock/Unlock (exclusive) interleaved with regular Put/Get (which also lock).
	// Tests that explicit Lock doesn't deadlock with the internal locking.
	s := newTestStoreNoDisk(t)

	// Seed
	for i := 0; i < 20; i++ {
		s.Put(fmt.Sprintf("k%d", i), rec(fmt.Sprintf("k%d", i), i))
	}

	var wg sync.WaitGroup

	// Exclusive lock operations (simulating SelectAndAssign pattern)
	for g := 0; g < 5; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				s.Lock()
				// Read + modify + write under exclusive lock
				key := fmt.Sprintf("k%d", i%20)
				if v, ok := s.GetUnlocked(key); ok {
					updated := *v
					updated.Value = gid*1000 + i
					s.PutUnlocked(key, &updated)
				}
				s.Unlock()
			}
		}(g)
	}

	// Regular operations (which internally lock)
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				s.Get(fmt.Sprintf("k%d", i%20))
				s.Filter(func(r *TestRecord) bool { return r.Value > 50 })
				s.Count(func(r *TestRecord) bool { return true })
				s.All()
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, 20, s.Len())
}

func TestStressConcurrentAllAndFilter(t *testing.T) {
	// All() and Filter() hold RLock. Ensure they don't starve writers
	// and return consistent snapshots.
	s := newTestStoreNoDisk(t)

	for i := 0; i < 50; i++ {
		s.Put(fmt.Sprintf("k%d", i), rec(fmt.Sprintf("k%d", i), i))
	}

	var wg sync.WaitGroup

	// Readers doing All
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				all := s.All()
				// Snapshot should be internally consistent
				for _, r := range all {
					if r == nil {
						t.Error("Got nil record from All()")
					}
				}
			}
		}()
	}

	// Readers doing Filter
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				result := s.Filter(func(r *TestRecord) bool {
					return r.Value >= 25
				})
				_ = result
			}
		}()
	}

	// Writers
	for g := 0; g < 5; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				key := fmt.Sprintf("k%d", i%50)
				s.Put(key, rec(key, gid*1000+i))
			}
		}(g)
	}

	wg.Wait()
}
