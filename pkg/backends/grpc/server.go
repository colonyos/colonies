package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/client/grpc/proto"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GRPCServer wraps a gRPC server to implement backends.Server interface
type GRPCServer struct {
	grpcServer *grpclib.Server
	listener   net.Listener
	addr       string
	port       int
	handler    RPCHandler

	// TLS configuration
	certFile string
	keyFile  string
}

// NewGRPCServer creates a new gRPC server wrapper
func NewGRPCServer(port int, handler RPCHandler) *GRPCServer {
	return &GRPCServer{
		addr:    fmt.Sprintf(":%d", port),
		port:    port,
		handler: handler,
	}
}

// NewGRPCServerWithAddr creates a new gRPC server with a specific address
func NewGRPCServerWithAddr(addr string, handler RPCHandler) *GRPCServer {
	return &GRPCServer{
		addr:    addr,
		handler: handler,
	}
}

// ListenAndServe starts the gRPC server
func (s *GRPCServer) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.listener = listener

	// Create gRPC server
	s.grpcServer = grpclib.NewServer()

	// Register service
	service := NewColoniesServiceImpl(s.handler)
	proto.RegisterColoniesServiceServer(s.grpcServer, service)

	// Start serving
	return s.grpcServer.Serve(listener)
}

// ListenAndServeTLS starts the gRPC server with TLS
func (s *GRPCServer) ListenAndServeTLS(certFile, keyFile string) error {
	s.certFile = certFile
	s.keyFile = keyFile

	// Load TLS credentials
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS credentials: %w", err)
	}

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.listener = listener

	// Create gRPC server with TLS
	s.grpcServer = grpclib.NewServer(grpclib.Creds(creds))

	// Register service
	service := NewColoniesServiceImpl(s.handler)
	proto.RegisterColoniesServiceServer(s.grpcServer, service)

	// Start serving
	return s.grpcServer.Serve(listener)
}

// Shutdown gracefully shuts down the gRPC server
func (s *GRPCServer) Shutdown(ctx context.Context) error {
	if s.grpcServer == nil {
		return nil
	}

	// Create a channel to signal when shutdown is complete
	done := make(chan struct{})

	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()

	// Wait for shutdown or context cancellation
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		// Force stop if context expires
		s.grpcServer.Stop()
		return ctx.Err()
	}
}

// ShutdownWithTimeout shuts down with a timeout
func (s *GRPCServer) ShutdownWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.Shutdown(ctx)
}

// SetAddr sets the server address
func (s *GRPCServer) SetAddr(addr string) {
	s.addr = addr
}

// SetHandler sets the RPC handler (must be called before ListenAndServe)
func (s *GRPCServer) SetHandler(handler RPCHandler) {
	s.handler = handler
}

// GetAddr returns the server address
func (s *GRPCServer) GetAddr() string {
	return s.addr
}

// SetReadTimeout is a no-op for gRPC (gRPC handles timeouts differently)
func (s *GRPCServer) SetReadTimeout(timeout time.Duration) {
	// gRPC doesn't use HTTP timeouts - it uses context deadlines
}

// SetWriteTimeout is a no-op for gRPC
func (s *GRPCServer) SetWriteTimeout(timeout time.Duration) {
	// gRPC doesn't use HTTP timeouts
}

// SetIdleTimeout is a no-op for gRPC
func (s *GRPCServer) SetIdleTimeout(timeout time.Duration) {
	// gRPC handles connection management internally
}

// SetReadHeaderTimeout is a no-op for gRPC
func (s *GRPCServer) SetReadHeaderTimeout(timeout time.Duration) {
	// gRPC doesn't use HTTP headers in the same way
}

// Engine returns nil as gRPC doesn't use the Engine pattern
func (s *GRPCServer) Engine() backends.Engine {
	return nil
}

// HTTPServer returns nil as gRPC server is not an HTTP server
func (s *GRPCServer) HTTPServer() *http.Server {
	return nil
}

// GRPCServer returns the underlying gRPC server
func (s *GRPCServer) GRPCServer() *grpclib.Server {
	return s.grpcServer
}

// Compile-time check that GRPCServer implements backends.Server
var _ backends.Server = (*GRPCServer)(nil)
