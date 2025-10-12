package coap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
	coapNet "github.com/plgd-dev/go-coap/v3/net"
	"github.com/plgd-dev/go-coap/v3/options"
	"github.com/plgd-dev/go-coap/v3/udp"
	"github.com/plgd-dev/go-coap/v3/udp/server"
	log "github.com/sirupsen/logrus"
)

// CoAPServer wraps a CoAP server to implement backends.Server interface
type CoAPServer struct {
	server  *server.Server
	listener *coapNet.UDPConn
	addr    string
	port    int
	handler RPCHandler
	router  *mux.Router
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewCoAPServer creates a new CoAP server wrapper
func NewCoAPServer(port int, handler RPCHandler) *CoAPServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &CoAPServer{
		addr:    fmt.Sprintf(":%d", port),
		port:    port,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// NewCoAPServerWithAddr creates a new CoAP server with a specific address
func NewCoAPServerWithAddr(addr string, handler RPCHandler) *CoAPServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &CoAPServer{
		addr:    addr,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// handleCoAPRequest handles incoming CoAP requests
func (s *CoAPServer) handleCoAPRequest(w mux.ResponseWriter, r *mux.Message) {
	if s.handler == nil {
		log.Error("No RPC handler configured for CoAP server")
		if err := w.SetResponse(codes.InternalServerError, message.TextPlain, bytes.NewReader([]byte("no RPC handler configured"))); err != nil {
			log.WithError(err).Error("Failed to set error response")
		}
		return
	}

	// Extract payload from CoAP request
	payload, err := io.ReadAll(r.Body())
	if err != nil {
		log.WithError(err).Error("Failed to read CoAP request body")
		if err := w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte("failed to read request body"))); err != nil {
			log.WithError(err).Error("Failed to set error response")
		}
		return
	}

	jsonPayload := string(payload)
	log.WithFields(log.Fields{
		"PayloadSize": len(jsonPayload),
	}).Debug("Received CoAP RPC request")

	// Call the shared RPC handler
	response, err := s.handler.HandleRPC(jsonPayload)
	if err != nil {
		log.WithError(err).Error("CoAP RPC handler error")
		if err := w.SetResponse(codes.InternalServerError, message.TextPlain, bytes.NewReader([]byte(err.Error()))); err != nil {
			log.WithError(err).Error("Failed to set error response")
		}
		return
	}

	// Send response back via CoAP
	if err := w.SetResponse(codes.Content, message.AppJSON, bytes.NewReader([]byte(response))); err != nil {
		log.WithError(err).Error("Failed to set success response")
	}
}

// handleHealthCheck handles CoAP health check requests
func (s *CoAPServer) handleHealthCheck(w mux.ResponseWriter, r *mux.Message) {
	log.Debug("CoAP health check request received")
	if err := w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("healthy"))); err != nil {
		log.WithError(err).Error("Failed to set health check response")
	}
}

// ListenAndServe starts the CoAP server
func (s *CoAPServer) ListenAndServe() error {
	log.WithField("Addr", s.addr).Info("Starting CoAP server")

	// Create router for CoAP endpoints
	s.router = mux.NewRouter()
	s.router.Handle("/api", mux.HandlerFunc(s.handleCoAPRequest))
	s.router.Handle("/health", mux.HandlerFunc(s.handleHealthCheck))

	// Create UDP listener
	listener, err := coapNet.NewListenUDP("udp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to create UDP listener: %w", err)
	}
	s.listener = listener

	// Create CoAP server with router
	s.server = udp.NewServer(
		options.WithMux(s.router),
		options.WithContext(s.ctx),
	)

	// Start serving (blocking call)
	err = s.server.Serve(listener)
	if err != nil {
		return fmt.Errorf("failed to start CoAP server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the CoAP server
func (s *CoAPServer) Shutdown(ctx context.Context) error {
	log.Info("Shutting down CoAP server")

	// Stop the server
	if s.server != nil {
		s.server.Stop()
	}

	// Cancel the server context
	if s.cancel != nil {
		s.cancel()
	}

	// Close the UDP listener
	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}

// ShutdownWithTimeout shuts down with a timeout
func (s *CoAPServer) ShutdownWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.Shutdown(ctx)
}

// SetAddr sets the server address
func (s *CoAPServer) SetAddr(addr string) {
	s.addr = addr
}

// SetHandler sets the RPC handler (must be called before ListenAndServe)
func (s *CoAPServer) SetHandler(handler RPCHandler) {
	s.handler = handler
}

// GetAddr returns the server address
func (s *CoAPServer) GetAddr() string {
	return s.addr
}

// SetReadTimeout is a no-op for CoAP (CoAP handles timeouts differently)
func (s *CoAPServer) SetReadTimeout(timeout time.Duration) {
	// CoAP has its own timeout mechanism
}

// SetWriteTimeout is a no-op for CoAP
func (s *CoAPServer) SetWriteTimeout(timeout time.Duration) {
	// CoAP has its own timeout mechanism
}

// SetIdleTimeout is a no-op for CoAP
func (s *CoAPServer) SetIdleTimeout(timeout time.Duration) {
	// CoAP handles connection management internally
}

// SetReadHeaderTimeout is a no-op for CoAP
func (s *CoAPServer) SetReadHeaderTimeout(timeout time.Duration) {
	// CoAP doesn't have headers like HTTP
}

// Engine returns nil as CoAP doesn't use the Engine pattern
func (s *CoAPServer) Engine() backends.Engine {
	return nil
}

// HTTPServer returns nil as CoAP server is not an HTTP server
func (s *CoAPServer) HTTPServer() *http.Server {
	return nil
}

// ListenAndServeTLS starts the CoAP server with DTLS (not implemented yet)
func (s *CoAPServer) ListenAndServeTLS(certFile, keyFile string) error {
	return fmt.Errorf("CoAP with DTLS not yet implemented")
}

// Compile-time check that CoAPServer implements backends.Server
var _ backends.Server = (*CoAPServer)(nil)
