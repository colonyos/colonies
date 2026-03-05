package flusher

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Stress tests for Flusher concurrency.
// Run with: go test -race -count=1 -run TestStress ./pkg/database/embedded/flusher/

// slowFlushable simulates a store whose FlushDirty takes some time,
// increasing the window for concurrency issues.
type slowFlushable struct {
	mu         sync.Mutex
	flushCount int64
}

func (s *slowFlushable) FlushDirty() error {
	time.Sleep(time.Millisecond)
	atomic.AddInt64(&s.flushCount, 1)
	return nil
}

func (s *slowFlushable) getFlushCount() int64 {
	return atomic.LoadInt64(&s.flushCount)
}

func TestStressFlusherWithManyStores(t *testing.T) {
	// Many stores being flushed concurrently with short intervals.
	f := NewFlusher(5 * time.Millisecond)

	const numStores = 20
	stores := make([]*slowFlushable, numStores)
	for i := 0; i < numStores; i++ {
		stores[i] = &slowFlushable{}
		f.AddStore(stores[i])
	}

	f.Start()
	time.Sleep(200 * time.Millisecond)
	f.Stop()

	// All stores should have been flushed at least once
	for i, s := range stores {
		assert.Greater(t, s.getFlushCount(), int64(0),
			"Store %d should have been flushed at least once", i)
	}
}

func TestStressFlusherAddStoreWhileRunning(t *testing.T) {
	// Add stores while the flusher is actively flushing.
	f := NewFlusher(5 * time.Millisecond)
	f.Start()

	var wg sync.WaitGroup
	const numAdders = 10
	const storesPerAdder = 5

	allStores := make([]*slowFlushable, 0, numAdders*storesPerAdder)
	var mu sync.Mutex

	for g := 0; g < numAdders; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < storesPerAdder; i++ {
				s := &slowFlushable{}
				f.AddStore(s)
				mu.Lock()
				allStores = append(allStores, s)
				mu.Unlock()
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond) // let flusher run a bit more
	f.Stop()

	// At least some of the dynamically added stores should have been flushed
	flushedCount := 0
	for _, s := range allStores {
		if s.getFlushCount() > 0 {
			flushedCount++
		}
	}
	assert.Greater(t, flushedCount, 0, "At least some dynamically added stores should have been flushed")
}

func TestStressFlusherRapidStartStop(t *testing.T) {
	// Rapid Start/Stop cycles to test clean shutdown under pressure.
	for cycle := 0; cycle < 20; cycle++ {
		f := NewFlusher(time.Millisecond)
		s := &slowFlushable{}
		f.AddStore(s)
		f.Start()
		time.Sleep(5 * time.Millisecond)
		f.Stop() // must not panic or deadlock
	}
}
