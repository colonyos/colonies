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

// EventHandler implements backends.RealtimeEventHandler for libp2p
type EventHandler struct {
	relayServer interface{}
	listeners   map[string][]chan *core.Process
	listenersMu sync.RWMutex
	stopped     bool
	stopChan    chan struct{}
}

// NewEventHandler creates a new libp2p event handler
func NewEventHandler(relayServer interface{}) backends.RealtimeEventHandler {
	return &EventHandler{
		relayServer: relayServer,
		listeners:   make(map[string][]chan *core.Process),
		stopChan:    make(chan struct{}),
	}
}

// Signal sends a process event to all registered listeners
func (e *EventHandler) Signal(process *core.Process) {
	e.listenersMu.RLock()
	defer e.listenersMu.RUnlock()
	
	if e.stopped {
		return
	}
	
	// Create key for this process type and state
	key := e.getListenerKey(process.FunctionSpec.Conditions.ExecutorType, process.State, process.ID)
	
	// Send to specific listeners
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
	
	// Send to general listeners (without specific process ID)
	generalKey := e.getListenerKey(process.FunctionSpec.Conditions.ExecutorType, process.State, "")
	if listeners, exists := e.listeners[generalKey]; exists {
		for _, ch := range listeners {
			select {
			case ch <- process:
			default:
				logrus.Warn("Listener channel full, dropping process event")
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
			relayServer: relayServer,
			listeners:   make(map[string][]chan *core.Process),
			stopChan:    make(chan struct{}),
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