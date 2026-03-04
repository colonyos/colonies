package flusher

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Flushable is implemented by any store that can flush dirty records to disk.
type Flushable interface {
	FlushDirty() error
}

// Flusher periodically flushes dirty records from all registered stores to disk.
type Flusher struct {
	stores   []Flushable
	interval time.Duration
	stopCh   chan struct{}
	doneCh   chan struct{}
	mu       sync.Mutex
}

// NewFlusher creates a new Flusher that flushes at the given interval.
func NewFlusher(interval time.Duration) *Flusher {
	return &Flusher{
		interval: interval,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// AddStore registers a store to be flushed periodically.
func (f *Flusher) AddStore(s Flushable) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stores = append(f.stores, s)
}

// Start begins the background flush loop.
func (f *Flusher) Start() {
	go f.loop()
}

// Stop signals the flusher to perform a final flush and stop.
// Blocks until the flush loop has exited.
func (f *Flusher) Stop() {
	close(f.stopCh)
	<-f.doneCh
}

func (f *Flusher) loop() {
	defer close(f.doneCh)

	ticker := time.NewTicker(f.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			f.flushAll()
		case <-f.stopCh:
			f.flushAll()
			return
		}
	}
}

func (f *Flusher) flushAll() {
	f.mu.Lock()
	stores := make([]Flushable, len(f.stores))
	copy(stores, f.stores)
	f.mu.Unlock()

	for _, s := range stores {
		if err := s.FlushDirty(); err != nil {
			log.WithError(err).Error("failed to flush dirty records")
		}
	}
}
