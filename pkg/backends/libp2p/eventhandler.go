package libp2p

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/sirupsen/logrus"
)

var ErrEventHandlerStopped = errors.New("event handler has been stopped")

// EventHandler implements backends.RealtimeEventHandler for libp2p.
//
// Signal Distribution Strategy (Thundering Herd Prevention):
//
// This event handler uses a two-pass signal distribution to prevent the "thundering herd"
// problem that can cause massive database load amplification.
//
// The Problem:
// When a new process is submitted, executors waiting for work need to be notified.
// If ALL waiting executors wake up simultaneously (broadcast), they all call Assign()
// at once, creating a stampede of database operations. With 8 executors and buffered
// channels allowing rapid re-wake, this caused 1,840x amplification in production
// (19.7M database updates for 1,341 processes instead of ~10K expected).
//
// The Solution - Two-Pass Distribution:
//
// Pass 1 (Broadcast): Notify ALL listeners waiting for a SPECIFIC processID.
// This is used for state change notifications (e.g., process completed) where
// the caller needs to know about their specific process.
//
// Pass 2 (Single Wake-up with Round-Robin): Wake exactly ONE general listener.
// General listeners are executors waiting for ANY process of their type.
// Only one executor should wake up per signal to attempt assignment.
// Round-robin ensures fair distribution across executors.
//
// Round-Robin Load Balancing:
// The nextExecutor map tracks which executor should receive the next signal
// for each target (executorType+state combination). This ensures all executors
// get equal opportunity to receive work, preventing any single executor from
// being starved or overloaded.
type EventHandler struct {
	relayServer  interface{}
	listeners    map[string][]chan *core.Process // key -> list of listener channels
	nextExecutor map[string]int                  // key -> round-robin index for fair distribution
	listenersMu  sync.RWMutex
	stopped      bool
	stopChan     chan struct{}
}

// NewEventHandler creates a new libp2p event handler
func NewEventHandler(relayServer interface{}) backends.RealtimeEventHandler {
	return &EventHandler{
		relayServer:  relayServer,
		listeners:    make(map[string][]chan *core.Process),
		nextExecutor: make(map[string]int),
		stopChan:     make(chan struct{}),
	}
}

// Signal distributes a process event to registered listeners using two-pass distribution.
//
// This is the core of the thundering herd prevention mechanism. See EventHandler
// documentation for the full explanation of the problem and solution.
//
// Algorithm:
//  1. Pass 1: Broadcast to all listeners waiting for this specific processID
//  2. Pass 2: Wake exactly ONE general listener using round-robin selection
//
// The round-robin selection ensures fair distribution:
//   - Listeners are stored in slice order (stable for libp2p unlike gin's map)
//   - nextExecutor[key] tracks which listener is next in rotation
//   - If the selected listener's channel is full, try the next one
//   - After successful send, advance the index for next time
func (e *EventHandler) Signal(process *core.Process) {
	e.listenersMu.RLock()
	defer e.listenersMu.RUnlock()

	if e.stopped {
		return
	}

	// Pass 1: Broadcast to all listeners waiting for this specific processID.
	// These are callers waiting for state changes on a process they submitted/own.
	key := e.getListenerKey(process.FunctionSpec.Conditions.ExecutorType, process.State, process.ID)
	if listeners, exists := e.listeners[key]; exists {
		for _, ch := range listeners {
			select {
			case ch <- process:
			default:
				// Channel is full, skip to avoid blocking
				logrus.Warn("Listener channel full, dropping process event")
			}
		}
	}

	// Pass 2: Wake ONE general listener using round-robin.
	// General listeners are executors waiting for ANY process of their type.
	generalKey := e.getListenerKey(process.FunctionSpec.Conditions.ExecutorType, process.State, "")
	if listeners, exists := e.listeners[generalKey]; exists {
		if len(listeners) == 0 {
			return
		}

		// Get current round-robin index and ensure it's valid
		idx := e.nextExecutor[generalKey] % len(listeners)

		// Try each listener starting from idx, wrapping around if channel is full
		for i := 0; i < len(listeners); i++ {
			ch := listeners[(idx+i)%len(listeners)]
			select {
			case ch <- process:
				// Success - advance round-robin index for next signal
				e.nextExecutor[generalKey] = (idx + i + 1) % len(listeners)
				return
			default:
				continue // Channel full, try next listener
			}
		}
	}
}

// Subscribe registers a subscription and returns channels for process events and errors
func (e *EventHandler) Subscribe(executorType string, state int, processID string, ctx context.Context) (chan *core.Process, chan error) {
	e.listenersMu.Lock()
	defer e.listenersMu.Unlock()
	
	if e.stopped {
		errCh := make(chan error, 1)
		errCh <- ErrEventHandlerStopped
		return nil, errCh
	}
	
	processCh := make(chan *core.Process, 100) // Buffered channel
	errCh := make(chan error, 1)
	
	key := e.getListenerKey(executorType, state, processID)
	e.listeners[key] = append(e.listeners[key], processCh)
	
	// Handle context cancellation
	go func() {
		<-ctx.Done()
		e.unsubscribe(key, processCh)
		close(processCh)
	}()
	
	return processCh, errCh
}

// WaitForProcess waits for a specific process state change
func (e *EventHandler) WaitForProcess(executorType string, state int, processID string, ctx context.Context) (*core.Process, error) {
	processCh, errCh := e.Subscribe(executorType, state, processID, ctx)
	
	select {
	case process := <-processCh:
		return process, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Stop stops the event handler
func (e *EventHandler) Stop() {
	e.listenersMu.Lock()
	defer e.listenersMu.Unlock()
	
	if e.stopped {
		return
	}
	
	e.stopped = true
	close(e.stopChan)
	
	// Close all listener channels
	for _, listeners := range e.listeners {
		for _, ch := range listeners {
			close(ch)
		}
	}
	
	// Clear listeners
	e.listeners = make(map[string][]chan *core.Process)
}

// getListenerKey creates a key for the listeners map
func (e *EventHandler) getListenerKey(executorType string, state int, processID string) string {
	if processID == "" {
		return fmt.Sprintf("%s_%d", executorType, state)
	}
	return fmt.Sprintf("%s_%d_%s", executorType, state, processID)
}

// unsubscribe removes a channel from the listeners
func (e *EventHandler) unsubscribe(key string, ch chan *core.Process) {
	e.listenersMu.Lock()
	defer e.listenersMu.Unlock()
	
	if listeners, exists := e.listeners[key]; exists {
		for i, listener := range listeners {
			if listener == ch {
				// Remove this listener
				e.listeners[key] = append(listeners[:i], listeners[i+1:]...)
				break
			}
		}
		
		// Clean up empty listener lists
		if len(e.listeners[key]) == 0 {
			delete(e.listeners, key)
		}
	}
}

// Compile-time check that EventHandler implements backends.RealtimeEventHandler
var _ backends.RealtimeEventHandler = (*EventHandler)(nil)

// TestableEventHandler extends EventHandler for testing
type TestableEventHandler struct {
	*EventHandler
}

// NewTestableEventHandler creates a new testable libp2p event handler
func NewTestableEventHandler(relayServer interface{}) backends.TestableRealtimeEventHandler {
	return &TestableEventHandler{
		EventHandler: &EventHandler{
			relayServer:  relayServer,
			listeners:    make(map[string][]chan *core.Process),
			nextExecutor: make(map[string]int),
			stopChan:     make(chan struct{}),
		},
	}
}

// NumberOfListeners returns listener counts for testing
func (t *TestableEventHandler) NumberOfListeners(executorType string, state int) (int, int, int) {
	t.listenersMu.RLock()
	defer t.listenersMu.RUnlock()
	
	key := t.getListenerKey(executorType, state, "")
	count := len(t.listeners[key])
	
	// For libp2p, return same count for all three values (simplified)
	return count, count, count
}

// HasStopped returns whether the handler has stopped for testing
func (t *TestableEventHandler) HasStopped() bool {
	t.listenersMu.RLock()
	defer t.listenersMu.RUnlock()
	return t.stopped
}

// Compile-time check that TestableEventHandler implements backends.TestableRealtimeEventHandler
var _ backends.TestableRealtimeEventHandler = (*TestableEventHandler)(nil)