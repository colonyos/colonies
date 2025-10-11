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
	
	// Close the backend and clean up resources
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
	GinClientBackendType    ClientBackendType = "gin"
	LibP2PClientBackendType ClientBackendType = "libp2p"
)

// ClientBackendFactory creates backend-specific clients
type ClientBackendFactory interface {
	CreateBackend(config *ClientConfig) (ClientBackend, error)
	GetBackendType() ClientBackendType
}

// ClientConfig holds configuration for client backends
type ClientConfig struct {
	BackendType    ClientBackendType
	Host           string
	Port           int
	Insecure       bool
	SkipTLSVerify  bool
	BootstrapPeers string // Comma-separated multiaddresses for LibP2P bootstrap peers
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

// CreateLibP2PClientConfig creates a client config for LibP2P backend
// host parameter should be a libp2p multiaddr (e.g., "/ip4/127.0.0.1/tcp/5000/p2p/12D3KooW...")
func CreateLibP2PClientConfig(host string) *ClientConfig {
	return &ClientConfig{
		BackendType:   LibP2PClientBackendType,
		Host:          host,
		Port:          0, // Not used for LibP2P
		Insecure:      false,
		SkipTLSVerify: false,
	}
}