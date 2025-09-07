package libp2p

import (
	"fmt"

	"github.com/colonyos/colonies/pkg/backends"
)

// Backend implements the backends.Backend interface for libp2p
type Backend struct {
	mode string
}

// NewBackend creates a new libp2p backend implementation
func NewBackend() backends.Backend {
	return &Backend{
		mode: "release", // default mode
	}
}

// NewEngine creates a new libp2p engine (not applicable for libp2p)
func (l *Backend) NewEngine() backends.Engine {
	panic("libp2p backend doesn't use engines - it uses direct peer-to-peer connections")
}

// NewEngineWithDefaults creates a new libp2p engine with defaults (not applicable for libp2p)
func (l *Backend) NewEngineWithDefaults() backends.Engine {
	panic("libp2p backend doesn't use engines - it uses direct peer-to-peer connections")
}

// NewServer creates a new libp2p server (not applicable in this pattern)
func (l *Backend) NewServer(port int, engine backends.Engine) backends.Server {
	panic("libp2p backend doesn't use the Engine/Server pattern - use libp2p-specific constructors")
}

// NewServerWithAddr creates a new libp2p server with address (not applicable in this pattern)
func (l *Backend) NewServerWithAddr(addr string, engine backends.Engine) backends.Server {
	panic("libp2p backend doesn't use the Engine/Server pattern - use libp2p-specific constructors")
}

// SetMode sets the backend mode
func (l *Backend) SetMode(mode string) {
	l.mode = mode
}

// GetMode returns the backend mode
func (l *Backend) GetMode() string {
	return l.mode
}

// Logger returns a logger middleware (no-op for libp2p)
func (l *Backend) Logger() backends.MiddlewareFunc {
	return func(c backends.Context) {
		// No-op for libp2p - logging is handled at the protocol level
		c.Next()
	}
}

// Recovery returns a recovery middleware (no-op for libp2p)
func (l *Backend) Recovery() backends.MiddlewareFunc {
	return func(c backends.Context) {
		// No-op for libp2p - error handling is done at the stream level
		c.Next()
	}
}

// Compile-time check that Backend implements backends.Backend
var _ backends.Backend = (*Backend)(nil)

// RealtimeBackend implements the backends.RealtimeBackend interface for libp2p
type RealtimeBackend struct {
	*Backend
}

// NewRealtimeBackend creates a new libp2p realtime backend implementation
func NewRealtimeBackend() backends.RealtimeBackend {
	return &RealtimeBackend{
		Backend: &Backend{mode: "release"},
	}
}

// CreateConnection creates a libp2p connection from a raw connection
func (r *RealtimeBackend) CreateConnection(rawConn interface{}) (backends.RealtimeConnection, error) {
	// For libp2p, we expect a network.Stream
	stream, ok := rawConn.(interface{})
	if !ok {
		return nil, fmt.Errorf("invalid connection type for libp2p backend")
	}
	
	return NewConnection(stream), nil
}

// CreateEventHandler creates a libp2p event handler
func (r *RealtimeBackend) CreateEventHandler(relayServer interface{}) backends.RealtimeEventHandler {
	return NewEventHandler(relayServer)
}

// CreateTestableEventHandler creates a testable libp2p event handler
func (r *RealtimeBackend) CreateTestableEventHandler(relayServer interface{}) backends.TestableRealtimeEventHandler {
	return NewTestableEventHandler(relayServer)
}

// CreateSubscriptionController creates a libp2p subscription controller
func (r *RealtimeBackend) CreateSubscriptionController(eventHandler backends.RealtimeEventHandler) backends.RealtimeSubscriptionController {
	return NewSubscriptionController(eventHandler)
}

// Compile-time check that RealtimeBackend implements backends.RealtimeBackend
var _ backends.RealtimeBackend = (*RealtimeBackend)(nil)

// FullBackend implements both Backend and RealtimeBackend
type FullBackend struct {
	*Backend
	*RealtimeBackend
}

// NewFullBackend creates a new complete libp2p backend
func NewFullBackend() backends.FullBackend {
	backend := &Backend{mode: "release"}
	return &FullBackend{
		Backend: backend,
		RealtimeBackend: &RealtimeBackend{Backend: backend},
	}
}

// Compile-time check that FullBackend implements backends.FullBackend
var _ backends.FullBackend = (*FullBackend)(nil)