package grpc

import (
	"github.com/colonyos/colonies/pkg/backends"
	log "github.com/sirupsen/logrus"
)

// Backend implements the backends.Backend interface for gRPC
type Backend struct {
	mode string
}

// NewBackend creates a new gRPC backend
func NewBackend() backends.Backend {
	return &Backend{
		mode: "release",
	}
}

// NewEngine panics as gRPC doesn't use the Engine pattern
func (b *Backend) NewEngine() backends.Engine {
	panic("gRPC backend doesn't use the Engine pattern - use NewServer directly")
}

// NewEngineWithDefaults panics as gRPC doesn't use the Engine pattern
func (b *Backend) NewEngineWithDefaults() backends.Engine {
	panic("gRPC backend doesn't use the Engine pattern - use NewServer directly")
}

// NewServer creates a new gRPC server with the given port and handler
// The engine parameter is ignored for gRPC as it doesn't use the Engine pattern
func (b *Backend) NewServer(port int, engine backends.Engine) backends.Server {
	// For gRPC, we need an RPC handler, not an Engine
	// This will need to be provided through a different mechanism
	log.Warn("NewServer called without RPC handler - server will not process requests")
	return NewGRPCServer(port, nil)
}

// NewServerWithAddr creates a new gRPC server with the given address
func (b *Backend) NewServerWithAddr(addr string, engine backends.Engine) backends.Server {
	log.Warn("NewServerWithAddr called without RPC handler - server will not process requests")
	return NewGRPCServerWithAddr(addr, nil)
}

// NewServerWithHandler creates a new gRPC server with an RPC handler
func (b *Backend) NewServerWithHandler(port int, handler RPCHandler) backends.Server {
	return NewGRPCServer(port, handler)
}

// NewServerWithAddrAndHandler creates a new gRPC server with address and handler
func (b *Backend) NewServerWithAddrAndHandler(addr string, handler RPCHandler) backends.Server {
	return NewGRPCServerWithAddr(addr, handler)
}

// SetMode sets the backend mode (release, debug, test)
func (b *Backend) SetMode(mode string) {
	b.mode = mode
	// gRPC doesn't have built-in mode settings like Gin
	if mode == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

// GetMode returns the current mode
func (b *Backend) GetMode() string {
	return b.mode
}

// Logger returns a no-op middleware as gRPC handles logging differently
func (b *Backend) Logger() backends.MiddlewareFunc {
	return func(c backends.Context) {
		// gRPC uses interceptors for logging, not middleware
		// This is a no-op to satisfy the interface
		c.Next()
	}
}

// Recovery returns a no-op middleware as gRPC handles panics differently
func (b *Backend) Recovery() backends.MiddlewareFunc {
	return func(c backends.Context) {
		// gRPC has built-in panic recovery
		// This is a no-op to satisfy the interface
		c.Next()
	}
}

// Compile-time check that Backend implements backends.Backend
var _ backends.Backend = (*Backend)(nil)
