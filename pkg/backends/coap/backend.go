package coap

import (
	"github.com/colonyos/colonies/pkg/backends"
	log "github.com/sirupsen/logrus"
)

// Backend implements the backends.Backend interface for CoAP
type Backend struct {
	mode string
}

// NewBackend creates a new CoAP backend
func NewBackend() backends.Backend {
	return &Backend{
		mode: "release",
	}
}

// NewEngine panics as CoAP doesn't use the Engine pattern
func (b *Backend) NewEngine() backends.Engine {
	panic("CoAP backend doesn't use the Engine pattern - use NewServer directly")
}

// NewEngineWithDefaults panics as CoAP doesn't use the Engine pattern
func (b *Backend) NewEngineWithDefaults() backends.Engine {
	panic("CoAP backend doesn't use the Engine pattern - use NewServer directly")
}

// NewServer creates a new CoAP server with the given port and handler
// The engine parameter is ignored for CoAP as it doesn't use the Engine pattern
func (b *Backend) NewServer(port int, engine backends.Engine) backends.Server {
	// For CoAP, we need an RPC handler, not an Engine
	// This will need to be provided through a different mechanism
	log.Warn("NewServer called without RPC handler - server will not process requests")
	return NewCoAPServer(port, nil)
}

// NewServerWithAddr creates a new CoAP server with the given address
func (b *Backend) NewServerWithAddr(addr string, engine backends.Engine) backends.Server {
	log.Warn("NewServerWithAddr called without RPC handler - server will not process requests")
	return NewCoAPServerWithAddr(addr, nil)
}

// NewServerWithHandler creates a new CoAP server with an RPC handler
func (b *Backend) NewServerWithHandler(port int, handler RPCHandler) backends.Server {
	return NewCoAPServer(port, handler)
}

// NewServerWithAddrAndHandler creates a new CoAP server with address and handler
func (b *Backend) NewServerWithAddrAndHandler(addr string, handler RPCHandler) backends.Server {
	return NewCoAPServerWithAddr(addr, handler)
}

// SetMode sets the backend mode (release, debug, test)
func (b *Backend) SetMode(mode string) {
	b.mode = mode
	// CoAP doesn't have built-in mode settings like Gin
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

// Logger returns a no-op middleware as CoAP handles logging differently
func (b *Backend) Logger() backends.MiddlewareFunc {
	return func(c backends.Context) {
		// CoAP doesn't use middleware pattern
		// This is a no-op to satisfy the interface
		c.Next()
	}
}

// Recovery returns a no-op middleware as CoAP handles panics differently
func (b *Backend) Recovery() backends.MiddlewareFunc {
	return func(c backends.Context) {
		// CoAP has its own error handling
		// This is a no-op to satisfy the interface
		c.Next()
	}
}

// Compile-time check that Backend implements backends.Backend
var _ backends.Backend = (*Backend)(nil)
