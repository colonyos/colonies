package backends

import (
	"context"
)

// ClientBackend defines the interface for different client transport implementations
type ClientBackend interface {
	// Send a raw JSON message (used for health checks and other simple operations)
	SendRawMessage(jsonString string, insecure bool) (string, error)
	
	// Send an RPC message with authentication
	SendMessage(method string, jsonString string, prvKey string, insecure bool, ctx context.Context) (string, error)
	
	// Check the health of the backend connection
	CheckHealth() error
	
	// Close the backend and clean up blueprints
	Close() error
}

// ClientBackendWithRealtime extends ClientBackend with realtime capabilities
type ClientBackendWithRealtime interface {
	ClientBackend
	RealtimeBackend
}

// BackendType represents different client backend types
type ClientBackendType string

const (
	GinClientBackendType ClientBackendType = "gin"
)

// ClientBackendFactory creates backend-specific clients
type ClientBackendFactory interface {
	CreateBackend(config *ClientConfig) (ClientBackend, error)
	GetBackendType() ClientBackendType
}

// ClientConfig holds configuration for client backends
type ClientConfig struct {
	BackendType   ClientBackendType
	Host          string
	Port          int
	Insecure      bool
	SkipTLSVerify bool
}

// CreateDefaultClientConfig creates a default client config for HTTP/Gin backend
func CreateDefaultClientConfig(host string, port int, insecure bool, skipTLSVerify bool) *ClientConfig {
	return &ClientConfig{
		BackendType:   GinClientBackendType,
		Host:          host,
		Port:          port,
		Insecure:      insecure,
		SkipTLSVerify: skipTLSVerify,
	}
}