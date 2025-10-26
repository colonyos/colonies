package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/client/grpc/proto"
	"github.com/stretchr/testify/assert"
)

// Mock RPC handler for testing
type mockRPCHandler struct {
	response string
	err      error
}

func (m *mockRPCHandler) HandleRPC(jsonPayload string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func TestNewBackend(t *testing.T) {
	backend := NewBackend()
	assert.NotNil(t, backend)

	// Check default mode
	assert.Equal(t, "release", backend.GetMode())

	// Verify it implements backends.Backend interface
	var _ backends.Backend = backend
}

func TestBackendSetGetMode(t *testing.T) {
	backend := NewBackend()

	// Test setting different modes
	backend.SetMode("debug")
	assert.Equal(t, "debug", backend.GetMode())

	backend.SetMode("release")
	assert.Equal(t, "release", backend.GetMode())

	backend.SetMode("test")
	assert.Equal(t, "test", backend.GetMode())
}

func TestBackendNewEnginePanics(t *testing.T) {
	backend := NewBackend()

	// NewEngine should panic for gRPC backend
	assert.Panics(t, func() {
		backend.NewEngine()
	}, "gRPC backend doesn't use the Engine pattern")
}

func TestBackendNewEngineWithDefaultsPanics(t *testing.T) {
	backend := NewBackend()

	// NewEngineWithDefaults should panic for gRPC backend
	assert.Panics(t, func() {
		backend.NewEngineWithDefaults()
	}, "gRPC backend doesn't use the Engine pattern")
}

func TestBackendNewServer(t *testing.T) {
	backend := NewBackend()

	// NewServer creates a server without handler (will warn)
	server := backend.NewServer(50051, nil)
	assert.NotNil(t, server)

	// Verify it implements backends.Server interface
	var _ backends.Server = server
}

func TestBackendNewServerWithAddr(t *testing.T) {
	backend := NewBackend()

	// NewServerWithAddr creates a server without handler
	server := backend.NewServerWithAddr(":50051", nil)
	assert.NotNil(t, server)

	// Check address
	assert.Equal(t, ":50051", server.GetAddr())
}

func TestBackendNewServerWithHandler(t *testing.T) {
	backend := NewBackend().(*Backend)

	handler := &mockRPCHandler{
		response: `{"status": "ok"}`,
	}

	server := backend.NewServerWithHandler(50051, handler)
	assert.NotNil(t, server)

	grpcServer := server.(*GRPCServer)
	assert.NotNil(t, grpcServer.handler)
}

func TestBackendNewServerWithAddrAndHandler(t *testing.T) {
	backend := NewBackend().(*Backend)

	handler := &mockRPCHandler{
		response: `{"status": "ok"}`,
	}

	server := backend.NewServerWithAddrAndHandler(":50051", handler)
	assert.NotNil(t, server)

	grpcServer := server.(*GRPCServer)
	assert.NotNil(t, grpcServer.handler)
	assert.Equal(t, ":50051", grpcServer.GetAddr())
}

func TestBackendLogger(t *testing.T) {
	backend := NewBackend()
	logger := backend.Logger()
	assert.NotNil(t, logger)

	// Logger is a no-op for gRPC
}

func TestBackendRecovery(t *testing.T) {
	backend := NewBackend()
	recovery := backend.Recovery()
	assert.NotNil(t, recovery)

	// Recovery is a no-op for gRPC
}

func TestGRPCServerSetGetAddr(t *testing.T) {
	server := NewGRPCServer(50051, nil)

	assert.Equal(t, ":50051", server.GetAddr())

	server.SetAddr(":50052")
	assert.Equal(t, ":50052", server.GetAddr())
}

func TestGRPCServerTimeouts(t *testing.T) {
	server := NewGRPCServer(50051, nil)

	// These are no-ops for gRPC but should not panic
	assert.NotPanics(t, func() {
		server.SetReadTimeout(10 * time.Second)
		server.SetWriteTimeout(10 * time.Second)
		server.SetIdleTimeout(60 * time.Second)
		server.SetReadHeaderTimeout(5 * time.Second)
	})
}

func TestGRPCServerEngine(t *testing.T) {
	server := NewGRPCServer(50051, nil)

	// gRPC doesn't use Engine pattern
	engine := server.Engine()
	assert.Nil(t, engine)
}

func TestGRPCServerHTTPServer(t *testing.T) {
	server := NewGRPCServer(50051, nil)

	// gRPC server is not an HTTP server
	httpServer := server.HTTPServer()
	assert.Nil(t, httpServer)
}

func TestGRPCServerShutdownBeforeStart(t *testing.T) {
	server := NewGRPCServer(50051, nil)

	// Shutdown before starting should not error
	err := server.ShutdownWithTimeout(1 * time.Second)
	assert.NoError(t, err)
}

func TestColoniesServiceImplSendMessage(t *testing.T) {
	handler := &mockRPCHandler{
		response: `{"result": "success"}`,
	}

	service := NewColoniesServiceImpl(handler)
	assert.NotNil(t, service)

	req := &proto.RPCRequest{
		JsonPayload: `{"method": "test"}`,
	}

	resp, err := service.SendMessage(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, `{"result": "success"}`, resp.JsonPayload)
	assert.Empty(t, resp.ErrorMessage)
}

func TestColoniesServiceImplSendMessageWithError(t *testing.T) {
	handler := &mockRPCHandler{
		err: errors.New("test error"),
	}

	service := NewColoniesServiceImpl(handler)

	req := &proto.RPCRequest{
		JsonPayload: `{"method": "test"}`,
	}

	resp, err := service.SendMessage(context.Background(), req)
	assert.NoError(t, err) // gRPC call doesn't error, error is in response
	assert.NotNil(t, resp)
	assert.Contains(t, resp.ErrorMessage, "test error")
}

func TestColoniesServiceImplSendMessageNoHandler(t *testing.T) {
	service := NewColoniesServiceImpl(nil)

	req := &proto.RPCRequest{
		JsonPayload: `{"method": "test"}`,
	}

	resp, err := service.SendMessage(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.ErrorMessage, "no RPC handler")
}

func TestColoniesServiceImplCheckHealth(t *testing.T) {
	service := NewColoniesServiceImpl(nil)

	req := &proto.HealthRequest{}

	resp, err := service.CheckHealth(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Healthy)
	assert.Equal(t, "1.0.0", resp.Version)
}

func TestGRPCServerImplementsInterface(t *testing.T) {
	// Verify GRPCServer implements backends.Server interface
	var _ backends.Server = (*GRPCServer)(nil)
}

func TestBackendImplementsInterface(t *testing.T) {
	// Verify Backend implements backends.Backend interface
	var _ backends.Backend = (*Backend)(nil)
}

func TestGRPCServerListenAndServe(t *testing.T) {
	// Skip integration test that requires actual network binding
	t.Skip("Skipping integration test - requires network binding")
}

func TestGRPCServerListenAndServeTLS(t *testing.T) {
	// Skip integration test that requires TLS certificates
	t.Skip("Skipping integration test - requires TLS certificates")
}

func TestGRPCServerShutdown(t *testing.T) {
	// Skip integration test
	t.Skip("Skipping integration test - requires running server")
}
