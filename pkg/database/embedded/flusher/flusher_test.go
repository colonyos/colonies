package flusher

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockFlushable struct {
	mu         sync.Mutex
	flushCount int
	lastFlush  time.Time
}

func (m *mockFlushable) FlushDirty() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flushCount++
	m.lastFlush = time.Now()
	return nil
}

func (m *mockFlushable) getFlushCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.flushCount
}

func TestFlusherPeriodicFlush(t *testing.T) {
	f := NewFlusher(50 * time.Millisecond)
	m := &mockFlushable{}
	f.AddStore(m)

	f.Start()
	time.Sleep(200 * time.Millisecond)
	f.Stop()

	// Should have flushed multiple times (at least 2-3 times in 200ms at 50ms interval)
	count := m.getFlushCount()
	assert.GreaterOrEqual(t, count, 2, "expected at least 2 flushes, got %d", count)
}

func TestFlusherStopTriggersFlush(t *testing.T) {
	f := NewFlusher(10 * time.Second) // very long interval
	m := &mockFlushable{}
	f.AddStore(m)

	f.Start()
	// Don't wait for any tick
	f.Stop()

	// Stop should have triggered a final flush
	assert.GreaterOrEqual(t, m.getFlushCount(), 1)
}

func TestFlusherStopNoStores(t *testing.T) {
	f := NewFlusher(50 * time.Millisecond)
	f.Start()
	f.Stop() // should not panic
}

func TestFlusherMultipleStores(t *testing.T) {
	f := NewFlusher(50 * time.Millisecond)
	m1 := &mockFlushable{}
	m2 := &mockFlushable{}
	m3 := &mockFlushable{}
	f.AddStore(m1)
	f.AddStore(m2)
	f.AddStore(m3)

	f.Start()
	time.Sleep(150 * time.Millisecond)
	f.Stop()

	// All stores should have been flushed
	assert.GreaterOrEqual(t, m1.getFlushCount(), 1)
	assert.GreaterOrEqual(t, m2.getFlushCount(), 1)
	assert.GreaterOrEqual(t, m3.getFlushCount(), 1)
}

func TestFlusherConcurrentAddStore(t *testing.T) {
	f := NewFlusher(20 * time.Millisecond)
	f.Start()

	var wg sync.WaitGroup
	var totalFlushes int64

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m := &mockFlushable{}
			f.AddStore(m)
			time.Sleep(100 * time.Millisecond)
			atomic.AddInt64(&totalFlushes, int64(m.getFlushCount()))
		}()
	}

	wg.Wait()
	f.Stop()

	// At least some flushes should have happened
	assert.Greater(t, atomic.LoadInt64(&totalFlushes), int64(0))
}
