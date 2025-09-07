package gin

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// Server wraps http.Server with gin integration
type Server struct {
	httpServer *http.Server
	engine     *Engine
}

// NewServer creates a new HTTP server with the given engine
func NewServer(port int, engine *Engine) *Server {
	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: engine.Handler(),
	}

	return &Server{
		httpServer: httpServer,
		engine:     engine,
	}
}

// NewServerWithAddr creates a new HTTP server with the given address and engine
func NewServerWithAddr(addr string, engine *Engine) *Server {
	httpServer := &http.Server{
		Addr:    addr,
		Handler: engine.Handler(),
	}

	return &Server{
		httpServer: httpServer,
		engine:     engine,
	}
}

// ListenAndServe starts the server and blocks until an error occurs
func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

// ListenAndServeTLS starts the HTTPS server and blocks until an error occurs
func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	return s.httpServer.ListenAndServeTLS(certFile, keyFile)
}

// Shutdown gracefully shuts down the server without interrupting active connections
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// ShutdownWithTimeout gracefully shuts down the server with a timeout
func (s *Server) ShutdownWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

// Engine returns the underlying Engine
func (s *Server) Engine() *Engine {
	return s.engine
}

// HTTPServer returns the underlying http.Server
func (s *Server) HTTPServer() *http.Server {
	return s.httpServer
}

// SetReadTimeout sets the read timeout for the HTTP server
func (s *Server) SetReadTimeout(timeout time.Duration) {
	s.httpServer.ReadTimeout = timeout
}

// SetWriteTimeout sets the write timeout for the HTTP server
func (s *Server) SetWriteTimeout(timeout time.Duration) {
	s.httpServer.WriteTimeout = timeout
}

// SetIdleTimeout sets the idle timeout for the HTTP server
func (s *Server) SetIdleTimeout(timeout time.Duration) {
	s.httpServer.IdleTimeout = timeout
}

// SetReadHeaderTimeout sets the read header timeout for the HTTP server
func (s *Server) SetReadHeaderTimeout(timeout time.Duration) {
	s.httpServer.ReadHeaderTimeout = timeout
}

// SetAddr sets the server address
func (s *Server) SetAddr(addr string) {
	s.httpServer.Addr = addr
}

// GetAddr returns the server address
func (s *Server) GetAddr() string {
	return s.httpServer.Addr
}