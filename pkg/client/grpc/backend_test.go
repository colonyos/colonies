package grpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/client/backends"
	"github.com/stretchr/testify/assert"
)

func TestNewGRPCClientBackend(t *testing.T) {
	// Skip this test as it requires a real gRPC server
	t.Skip("Skipping - requires running gRPC server")

	config := &backends.ClientConfig{
		BackendType: backends.GRPCClientBackendType,
		Host:        "localhost",
		Port:        50051,
		Insecure:    true,
	}

	backend, err := NewGRPCClientBackend(config)
	assert.NoError(t, err)
	assert.NotNil(t, backend)

	// Verify it implements ClientBackend interface
	var _ backends.ClientBackend = backend

	// Clean up
	backend.Close()
}

func TestNewGRPCClientBackendInvalidType(t *testing.T) {
	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType, // Wrong type
		Host:        "localhost",
		Port:        50051,
	}

	backend, err := NewGRPCClientBackend(config)
	assert.Error(t, err)
	assert.Nil(t, backend)
	assert.Contains(t, err.Error(), "invalid backend type")
}

func TestGRPCClientBackendSendRawMessage(t *testing.T) {
	t.Skip("Skipping - requires running gRPC server")
}

func TestGRPCClientBackendSendRawMessageWithError(t *testing.T) {
	t.Skip("Skipping - requires running gRPC server")
}

func TestGRPCClientBackendCheckHealth(t *testing.T) {
	t.Skip("Skipping - requires running gRPC server")
}

func TestGRPCClientBackendCheckHealthUnhealthy(t *testing.T) {
	t.Skip("Skipping - requires running gRPC server")
}

func TestGRPCClientBackendClose(t *testing.T) {
	// Test close on nil connection
	client := &GRPCClientBackend{}
	err := client.Close()
	assert.NoError(t, err)
}

func TestGRPCClientBackendFactory(t *testing.T) {
	factory := NewGRPCClientBackendFactory()
	assert.NotNil(t, factory)

	// Test GetBackendType
	backendType := factory.GetBackendType()
	assert.Equal(t, backends.GRPCClientBackendType, backendType)
}

func TestGRPCClientBackendFactoryCreateBackendInvalidType(t *testing.T) {
	factory := NewGRPCClientBackendFactory()

	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType, // Wrong type
		Host:        "localhost",
		Port:        50051,
	}

	backend, err := factory.CreateBackend(config)
	assert.Error(t, err)
	assert.Nil(t, backend)
}

func TestGRPCClientBackendImplementsInterface(t *testing.T) {
	// Verify GRPCClientBackend implements backends.ClientBackend interface
	var _ backends.ClientBackend = (*GRPCClientBackend)(nil)
}

func TestGRPCClientBackendWithContext(t *testing.T) {
	t.Skip("Skipping SendMessage test - requires full RPC message handling")
}

func TestGRPCClientBackendSecureConnection(t *testing.T) {
	// Test that secure configuration is accepted (but skip actual connection)
	t.Skip("Skipping TLS test - requires certificate setup")
}

func TestGRPCClientBackendSkipTLSVerify(t *testing.T) {
	// Test that skip TLS verify configuration is accepted
	t.Skip("Skipping TLS test - requires certificate setup")
}
