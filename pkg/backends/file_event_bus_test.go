package backends

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

// drainOne reads one event from a channel with a generous test timeout.
// Returns the event and true on success; nil/false on timeout.
func drainOne(t *testing.T, ch <-chan *core.FileEvent, d time.Duration) (*core.FileEvent, bool) {
	t.Helper()
	select {
	case ev, ok := <-ch:
		if !ok {
			return nil, false
		}
		return ev, true
	case <-time.After(d):
		return nil, false
	}
}

// expectNoEvent asserts no event arrives within the duration. Used when
// the matcher should reject — confirms filtering rather than just slowness.
func expectNoEvent(t *testing.T, ch <-chan *core.FileEvent, d time.Duration) {
	t.Helper()
	select {
	case ev := <-ch:
		t.Fatalf("expected no event, got %+v", ev)
	case <-time.After(d):
	}
}

// TestPublishDeliversToMatchingSubscriber is the happy path: subscribe to
// a colony + label prefix, publish a matching event, receive it intact.
func TestPublishDeliversToMatchingSubscriber(t *testing.T) {
	bus := NewInMemoryFileEventBus()
	defer bus.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	evCh, _ := bus.Subscribe("c1", "/home/root/inbox", nil, ctx)

	ev := core.CreateFileAddedEvent(&core.File{
		ID:         "id1",
		ColonyName: "c1",
		Label:      "/home/root/inbox",
		Name:       "x.md",
		Size:       16,
	}, time.Now())
	bus.Publish(ev)

	got, ok := drainOne(t, evCh, time.Second)
	assert.True(t, ok)
	assert.True(t, got.Equals(ev))
}

func TestPublishFiltersByColony(t *testing.T) {
	bus := NewInMemoryFileEventBus()
	defer bus.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	evCh, _ := bus.Subscribe("c1", "", nil, ctx)

	bus.Publish(core.CreateFileAddedEvent(&core.File{ID: "id", ColonyName: "c2", Label: "/l", Name: "n"}, time.Now()))
	expectNoEvent(t, evCh, 50*time.Millisecond)

	bus.Publish(core.CreateFileAddedEvent(&core.File{ID: "id", ColonyName: "c1", Label: "/l", Name: "n"}, time.Now()))
	_, ok := drainOne(t, evCh, time.Second)
	assert.True(t, ok)
}

func TestPublishFiltersByLabelPrefix(t *testing.T) {
	bus := NewInMemoryFileEventBus()
	defer bus.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	evCh, _ := bus.Subscribe("c1", "/home/root/inbox", nil, ctx)

	cases := []struct {
		label string
		match bool
	}{
		{"/home/root/inbox", true},
		{"/home/root/inbox/2026", true},
		{"/home/root/outbox", false},
		{"/home/root/in", false},      // not a separator-aligned prefix
		{"/home/root/inboxx", false},  // not separator-aligned
	}
	for _, c := range cases {
		bus.Publish(core.CreateFileAddedEvent(&core.File{
			ID: "id", ColonyName: "c1", Label: c.label, Name: "n",
		}, time.Now()))
		if c.match {
			ev, ok := drainOne(t, evCh, time.Second)
			assert.True(t, ok, "expected match for %s", c.label)
			assert.Equal(t, c.label, ev.Label)
		} else {
			expectNoEvent(t, evCh, 30*time.Millisecond)
		}
	}
}

// Empty prefix is the documented "all labels in this colony" wildcard.
// Verify it doesn't accidentally leak across colonies.
func TestPublishEmptyPrefixMatchesAllLabelsInColony(t *testing.T) {
	bus := NewInMemoryFileEventBus()
	defer bus.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	evCh, _ := bus.Subscribe("c1", "", nil, ctx)

	for _, label := range []string{"/a", "/b/c", "/x/y/z"} {
		bus.Publish(core.CreateFileAddedEvent(&core.File{
			ID: "id", ColonyName: "c1", Label: label, Name: "n",
		}, time.Now()))
		ev, ok := drainOne(t, evCh, time.Second)
		assert.True(t, ok)
		assert.Equal(t, label, ev.Label)
	}

	// Foreign colony is still rejected.
	bus.Publish(core.CreateFileAddedEvent(&core.File{
		ID: "id", ColonyName: "c2", Label: "/anything", Name: "n",
	}, time.Now()))
	expectNoEvent(t, evCh, 30*time.Millisecond)
}

func TestPublishFiltersByKind(t *testing.T) {
	bus := NewInMemoryFileEventBus()
	defer bus.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Only FileAdded events
	evCh, _ := bus.Subscribe("c1", "", []int{int(core.FileAdded)}, ctx)

	bus.Publish(core.CreateFileAddedEvent(&core.File{ID: "id", ColonyName: "c1", Label: "/l", Name: "n"}, time.Now()))
	bus.Publish(core.CreateFileUpdatedEvent(&core.File{ID: "id", ColonyName: "c1", Label: "/l", Name: "n"}, time.Now()))
	bus.Publish(core.CreateFileRemovedEvent("c1", "/l", "n", time.Now()))

	got, ok := drainOne(t, evCh, time.Second)
	assert.True(t, ok)
	assert.Equal(t, core.FileAdded, got.Kind)
	expectNoEvent(t, evCh, 50*time.Millisecond)
}

// Multi-subscriber fan-out: every matching subscriber receives the event,
// and they don't see each other's traffic when filters differ.
func TestPublishFansOutToMultipleSubscribers(t *testing.T) {
	bus := NewInMemoryFileEventBus()
	defer bus.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wide, _ := bus.Subscribe("c1", "", nil, ctx)
	narrow, _ := bus.Subscribe("c1", "/home/root/inbox", nil, ctx)
	otherColony, _ := bus.Subscribe("c2", "", nil, ctx)

	bus.Publish(core.CreateFileAddedEvent(&core.File{
		ID: "id", ColonyName: "c1", Label: "/home/root/inbox", Name: "n",
	}, time.Now()))

	_, ok := drainOne(t, wide, time.Second)
	assert.True(t, ok, "wide subscriber should match")
	_, ok = drainOne(t, narrow, time.Second)
	assert.True(t, ok, "narrow subscriber should match")
	expectNoEvent(t, otherColony, 30*time.Millisecond)
}

// Cancelling the context closes both channels and removes the subscriber
// from the bus. Without this the bus would leak per-subscription
// goroutines and channels indefinitely.
func TestSubscribeUnsubscribesOnContextCancel(t *testing.T) {
	bus := NewInMemoryFileEventBus()
	defer bus.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	evCh, errCh := bus.Subscribe("c1", "", nil, ctx)

	assert.Equal(t, 1, bus.NumberOfSubscribers("c1"))
	cancel()

	// Channels close.
	for {
		_, ok := <-evCh
		if !ok {
			break
		}
	}
	for {
		_, ok := <-errCh
		if !ok {
			break
		}
	}

	// Subscriber count drops back to zero.
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if bus.NumberOfSubscribers("c1") == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("subscriber count did not drop to zero after cancel")
}

// Stop closes every active subscription and makes Publish a no-op.
// Subsequent Subscribe calls return already-closed channels.
func TestStopClosesAllSubscribers(t *testing.T) {
	bus := NewInMemoryFileEventBus()
	ctx := context.Background()
	evCh1, _ := bus.Subscribe("c1", "", nil, ctx)
	evCh2, _ := bus.Subscribe("c2", "", nil, ctx)

	bus.Stop()

	// Both channels close.
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		_, ok1 := drainOne(t, evCh1, 10*time.Millisecond)
		_, ok2 := drainOne(t, evCh2, 10*time.Millisecond)
		if !ok1 && !ok2 {
			break
		}
	}

	// Publish after Stop is a silent no-op.
	bus.Publish(core.CreateFileAddedEvent(&core.File{ColonyName: "c1", Label: "/l", Name: "n"}, time.Now()))

	// Subscribe after Stop returns immediately-closed channels.
	evCh3, errCh3 := bus.Subscribe("c1", "", nil, ctx)
	_, ok := <-evCh3
	assert.False(t, ok)
	_, ok = <-errCh3
	assert.False(t, ok)

	// Stop is idempotent.
	bus.Stop()
}

// Backpressure: when a subscriber doesn't drain, the bus drops events
// rather than blocking the publisher, and reports overflow on the error
// channel.
func TestOverflowDropsAndSignals(t *testing.T) {
	bus := NewInMemoryFileEventBusWithBuffer(2)
	defer bus.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	evCh, errCh := bus.Subscribe("c1", "", nil, ctx)

	for i := 0; i < 50; i++ {
		bus.Publish(core.CreateFileAddedEvent(&core.File{
			ID: "id", ColonyName: "c1", Label: "/l", Name: "n",
		}, time.Now()))
	}

	select {
	case err := <-errCh:
		assert.Equal(t, ErrSubscriberOverflowed, err)
	case <-time.After(time.Second):
		t.Fatal("expected overflow error on errChan")
	}

	// Drain a bunch — confirm we only get up to (buffer + 1) events,
	// not all 50, because the publisher dropped on overflow.
	count := 0
	for {
		_, ok := drainOne(t, evCh, 30*time.Millisecond)
		if !ok {
			break
		}
		count++
	}
	assert.LessOrEqual(t, count, 4, "should have dropped most events on overflow")
}

// Negative subscribe arg: zero or negative buffer must clamp to >= 1
// rather than panicking on channel construction.
func TestBufferSizeClampedToMinimumOne(t *testing.T) {
	bus := NewInMemoryFileEventBusWithBuffer(0)
	defer bus.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	evCh, _ := bus.Subscribe("c1", "", nil, ctx)

	bus.Publish(core.CreateFileAddedEvent(&core.File{ColonyName: "c1", Label: "/l", Name: "n"}, time.Now()))
	_, ok := drainOne(t, evCh, time.Second)
	assert.True(t, ok)
}

// Publish on a nil event must be a no-op (not a panic). Internal callers
// occasionally short-circuit and pass nil; we tolerate it.
func TestPublishNilIsNoop(t *testing.T) {
	bus := NewInMemoryFileEventBus()
	defer bus.Stop()
	bus.Publish(nil) // should not panic
}

// NumberOfSubscribers reports per-colony counts independently.
func TestNumberOfSubscribersIsPerColony(t *testing.T) {
	bus := NewInMemoryFileEventBus()
	defer bus.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus.Subscribe("c1", "", nil, ctx)
	bus.Subscribe("c1", "/x", nil, ctx)
	bus.Subscribe("c2", "", nil, ctx)

	assert.Equal(t, 2, bus.NumberOfSubscribers("c1"))
	assert.Equal(t, 1, bus.NumberOfSubscribers("c2"))
	assert.Equal(t, 0, bus.NumberOfSubscribers("c3"))
}

// Concurrent Publish + Subscribe + cancel: the bus must not panic, leak,
// or deadlock under multi-goroutine pressure. This is the key concurrency
// safety test.
func TestConcurrentPublishAndSubscribe(t *testing.T) {
	bus := NewInMemoryFileEventBus()
	defer bus.Stop()

	var wg sync.WaitGroup

	// Publishers
	for p := 0; p < 4; p++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				bus.Publish(core.CreateFileAddedEvent(&core.File{
					ID: "id", ColonyName: "c1", Label: "/x", Name: "n",
				}, time.Now()))
			}
		}()
	}

	// Subscribers churning subscribe/cancel
	for s := 0; s < 8; s++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				ctx, cancel := context.WithCancel(context.Background())
				evCh, _ := bus.Subscribe("c1", "", nil, ctx)
				// Drain a few events then bail
				go func() {
					for range evCh {
					}
				}()
				time.Sleep(time.Millisecond)
				cancel()
			}
		}()
	}

	wg.Wait()
}

// Defensive copy of the kinds slice: a caller mutating the slice after
// subscribing must not alter the active subscription's filter.
func TestSubscribeKindsAreCopied(t *testing.T) {
	bus := NewInMemoryFileEventBus()
	defer bus.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kinds := []int{int(core.FileAdded)}
	evCh, _ := bus.Subscribe("c1", "", kinds, ctx)

	// Caller mutates after subscribing.
	kinds[0] = int(core.FileRemoved)

	// FileAdded should still match per the captured filter.
	bus.Publish(core.CreateFileAddedEvent(&core.File{ColonyName: "c1", Label: "/l", Name: "n"}, time.Now()))
	_, ok := drainOne(t, evCh, time.Second)
	assert.True(t, ok)

	// FileRemoved should NOT match — the post-subscribe mutation must not
	// have changed the filter.
	bus.Publish(core.CreateFileRemovedEvent("c1", "/l", "n", time.Now()))
	expectNoEvent(t, evCh, 30*time.Millisecond)
}
