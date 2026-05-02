package backends

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

// ErrSubscriberOverflowed is sent on a subscriber's error channel when its
// buffered event channel filled up and at least one event was dropped.
// Subscribers can use this as a signal to resync via GetFileData rather
// than relying on real-time delivery alone.
var ErrSubscriberOverflowed = errors.New("file subscriber overflowed; events dropped")

// FileEventBus is the in-process publish/subscribe primitive for ColonyFS
// file events. It is intentionally small: Publish fans out matching events
// to every active subscription, Subscribe registers a new subscription
// scoped to a colony / label-prefix / kind filter.
//
// Lifecycle: subscriptions live until the caller's ctx is canceled or
// Stop is called on the bus. The bus does NOT own the websocket — the
// websocket handler owns it and translates events to the wire format.
//
// Backpressure: if a subscriber's buffered event channel fills up, the
// oldest events are dropped and ErrSubscriberOverflowed is delivered on
// the error channel exactly once per overflow streak. Subscribers reset
// the streak by draining the event channel.
type FileEventBus interface {
	// Publish synchronously fans out an event to every matching active
	// subscriber. Non-blocking on a per-subscriber basis: a slow
	// subscriber drops events instead of stalling the publisher.
	Publish(ev *core.FileEvent)

	// Subscribe registers a new subscriber. The returned channels are
	// closed when ctx is canceled or the bus is stopped. Callers must
	// drain both channels to avoid memory leaks until they close.
	Subscribe(colonyName, labelPrefix string, kinds []int, ctx context.Context) (<-chan *core.FileEvent, <-chan error)

	// NumberOfSubscribers returns the current count of active
	// subscribers for a colony. Useful for tests and monitoring.
	NumberOfSubscribers(colonyName string) int

	// Stop drains every active subscriber. Safe to call multiple times.
	// After Stop, Publish becomes a no-op and Subscribe immediately
	// returns closed channels.
	Stop()
}

// fileSubscriber holds the per-subscription state.
type fileSubscriber struct {
	colonyName  string
	labelPrefix string
	kinds       []int
	eventChan   chan *core.FileEvent
	errChan     chan error
	cancel      context.CancelFunc
	overflowed  bool // true while we owe an ErrSubscriberOverflowed delivery
}

// inMemoryFileEventBus is the default FileEventBus implementation. It is
// safe for concurrent Publish/Subscribe calls. Single-server scope: events
// published on one server instance are NOT replicated to other servers in
// a cluster (matches the existing process-subscription model).
type inMemoryFileEventBus struct {
	mu          sync.RWMutex
	subscribers map[*fileSubscriber]struct{}
	bufferSize  int
	stopped     bool
}

// DefaultEventBufferSize is the per-subscription event buffer. 256 was
// picked to match the order of magnitude used by the channel-router code
// elsewhere in this repo. Adjustable via NewInMemoryFileEventBusWithBuffer
// in tests that need to exercise overflow behaviour deterministically.
const DefaultEventBufferSize = 256

// NewInMemoryFileEventBus returns a bus with the default buffer size.
func NewInMemoryFileEventBus() FileEventBus {
	return NewInMemoryFileEventBusWithBuffer(DefaultEventBufferSize)
}

// NewInMemoryFileEventBusWithBuffer returns a bus with a configurable
// per-subscription buffer. Tests use a small buffer to exercise overflow.
func NewInMemoryFileEventBusWithBuffer(bufferSize int) FileEventBus {
	if bufferSize < 1 {
		bufferSize = 1
	}
	return &inMemoryFileEventBus{
		subscribers: make(map[*fileSubscriber]struct{}),
		bufferSize:  bufferSize,
	}
}

func (b *inMemoryFileEventBus) Publish(ev *core.FileEvent) {
	if ev == nil {
		return
	}
	// Hold the bus read lock for the whole fan-out so removeSubscriber
	// (which takes the write lock before closing the channels) blocks
	// until in-flight sends finish. Sends are non-blocking — slow
	// subscribers drop events instead of stalling the publisher — so
	// the window where this lock is held is bounded by the number of
	// active subscribers, not by their drain rate.
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.stopped {
		return
	}

	for s := range b.subscribers {
		// Filter: same colony, prefix-matched label, kind-matched.
		if s.colonyName != ev.ColonyName {
			continue
		}
		if !ev.MatchesPrefix(s.labelPrefix) {
			continue
		}
		if !ev.MatchesKinds(s.kinds) {
			continue
		}

		// Non-blocking send. On overflow we drop the event and arm the
		// overflow signal so the next time the subscriber drains, it
		// receives one ErrSubscriberOverflowed before normal events
		// resume. We drain one event from the head of the channel to
		// keep the buffer's age bounded — drop-oldest semantics.
		select {
		case s.eventChan <- ev:
		default:
			b.markOverflowLocked(s)
			// Drop the oldest event to make room and try once more.
			select {
			case <-s.eventChan:
			default:
			}
			select {
			case s.eventChan <- ev:
			default:
				// Buffer is being thrashed; give up on this event.
			}
		}
	}
}

// markOverflowLocked delivers ErrSubscriberOverflowed on s.errChan exactly
// once per overflow streak. Caller must hold b.mu (read lock is fine) so
// the subscriber's channels stay alive for the duration of the send. Uses
// a non-blocking send so a wedged error reader can't block the publisher.
func (b *inMemoryFileEventBus) markOverflowLocked(s *fileSubscriber) {
	if s.overflowed {
		return
	}
	s.overflowed = true
	select {
	case s.errChan <- ErrSubscriberOverflowed:
	default:
		// Error buffer full; the next overflow attempt will retry the
		// signal once this one drains.
		s.overflowed = false
	}
}

func (b *inMemoryFileEventBus) Subscribe(colonyName, labelPrefix string, kinds []int, ctx context.Context) (<-chan *core.FileEvent, <-chan error) {
	b.mu.Lock()
	if b.stopped {
		b.mu.Unlock()
		eventChan := make(chan *core.FileEvent)
		errChan := make(chan error)
		close(eventChan)
		close(errChan)
		return eventChan, errChan
	}

	// Wrap the caller's context so we can cancel from Stop without
	// disturbing the parent.
	subCtx, cancel := context.WithCancel(ctx)

	s := &fileSubscriber{
		colonyName:  colonyName,
		labelPrefix: labelPrefix,
		kinds:       append([]int{}, kinds...), // defensive copy
		eventChan:   make(chan *core.FileEvent, b.bufferSize),
		// Error channel is buffered enough for one overflow signal
		// outstanding plus a final cancel-error.
		errChan: make(chan error, 2),
		cancel:  cancel,
	}
	b.subscribers[s] = struct{}{}
	b.mu.Unlock()

	// On ctx cancel, remove the subscription and close its channels so
	// the consumer goroutine exits cleanly.
	go func() {
		<-subCtx.Done()
		b.removeSubscriber(s)
	}()

	// Reset overflow watermark when the subscriber drains. We can't tell
	// from inside Publish that the channel has drained without polling,
	// so we run a tiny watcher goroutine that re-arms the flag whenever
	// the channel goes from full to non-full. Cheap because it only
	// wakes on actual sends/receives and exits on cancel.
	go b.overflowWatchdog(s, subCtx)

	return s.eventChan, s.errChan
}

func (b *inMemoryFileEventBus) overflowWatchdog(s *fileSubscriber, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
		}
		// Re-arm if the channel currently has headroom.
		b.mu.Lock()
		if s.overflowed && len(s.eventChan) < b.bufferSize {
			s.overflowed = false
		}
		b.mu.Unlock()
	}
}

// removeSubscriber takes the write lock to wait for any in-flight Publish
// to finish before closing the subscriber's channels. Without this, a
// concurrent Publish could panic with "send on closed channel".
func (b *inMemoryFileEventBus) removeSubscriber(s *fileSubscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.subscribers[s]; !ok {
		return
	}
	delete(b.subscribers, s)
	close(s.eventChan)
	close(s.errChan)
}

func (b *inMemoryFileEventBus) NumberOfSubscribers(colonyName string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	n := 0
	for s := range b.subscribers {
		if s.colonyName == colonyName {
			n++
		}
	}
	return n
}

func (b *inMemoryFileEventBus) Stop() {
	b.mu.Lock()
	if b.stopped {
		b.mu.Unlock()
		return
	}
	b.stopped = true
	subs := make([]*fileSubscriber, 0, len(b.subscribers))
	for s := range b.subscribers {
		subs = append(subs, s)
	}
	b.mu.Unlock()

	for _, s := range subs {
		s.cancel() // triggers removeSubscriber via the goroutine above
	}
}
