package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/colonyos/colonies/pkg/client/backends"
)

// ColoniesClient is the main client for interacting with Colonies server
type ColoniesClient struct {
	backend backends.ClientBackend
	config  *backends.ClientConfig
}

// CreateColoniesClient creates a new ColoniesClient with default HTTP/Gin backend
func CreateColoniesClient(host string, port int, insecure bool, skipTLSVerify bool) *ColoniesClient {
	config := backends.CreateDefaultClientConfig(host, port, insecure, skipTLSVerify)
	return CreateColoniesClientWithConfig(config)
}

// CreateColoniesClientWithConfig creates a new ColoniesClient with specified configuration
func CreateColoniesClientWithConfig(config *backends.ClientConfig) *ColoniesClient {
	client := &ColoniesClient{
		config: config,
	}
	
	// Initialize with the appropriate backend
	factory := getBackendFactory(config.BackendType)
	if factory == nil {
		// Default to gin backend if unknown type
		factory = &ginBackendFactory{}
	}
	
	backend, err := factory.CreateBackend(config)
	if err != nil {
		// For backward compatibility, panic on initialization error
		panic(fmt.Sprintf("Failed to create client backend: %v", err))
	}
	
	client.backend = backend
	return client
}

// CreateColoniesClientWithMultipleBackends creates a client that tries multiple backends with fallback
func CreateColoniesClientWithMultipleBackends(configs []*backends.ClientConfig) *ColoniesClient {
	if len(configs) == 0 {
		panic("At least one backend configuration required")
	}

	// If only one config, use single backend
	if len(configs) == 1 {
		return CreateColoniesClientWithConfig(configs[0])
	}

	// Create multi-backend client
	multiBackend, err := backends.NewMultiBackendClient(configs, backendFactories)
	if err != nil {
		panic(fmt.Sprintf("Failed to create multi-backend client: %v", err))
	}

	return &ColoniesClient{
		backend: multiBackend,
		config:  configs[0], // Use first config for compatibility
	}
}

// SendRawMessage sends a raw JSON message using the underlying backend
func (client *ColoniesClient) SendRawMessage(jsonString string, insecure bool) (string, error) {
	return client.backend.SendRawMessage(jsonString, insecure)
}

// sendMessage sends an RPC message using the underlying backend
func (client *ColoniesClient) sendMessage(method string, jsonString string, prvKey string, insecure bool, ctx context.Context) (string, error) {
	return client.backend.SendMessage(method, jsonString, prvKey, insecure, ctx)
}

// establishRealtimeConn establishes a realtime connection using the underlying backend
func (client *ColoniesClient) establishRealtimeConn(jsonString string) (backends.RealtimeConnection, error) {
	// Check if backend supports realtime connections
	if realtimeBackend, ok := client.backend.(backends.RealtimeBackend); ok {
		return realtimeBackend.EstablishRealtimeConn(jsonString)
	}
	return nil, errors.New("backend does not support realtime connections")
}

// CloseClient closes the client and cleans up resources
func (client *ColoniesClient) CloseClient() error {
	if client.backend != nil {
		return client.backend.Close()
	}
	return nil
}

// GetConfig returns the client configuration
func (client *ColoniesClient) GetConfig() *backends.ClientConfig {
	return client.config
}

// Backend registry for different client backends
var backendFactories = make(map[backends.ClientBackendType]backends.ClientBackendFactory)

// RegisterBackendFactory registers a backend factory
func RegisterBackendFactory(factory backends.ClientBackendFactory) {
	backendFactories[factory.GetBackendType()] = factory
}

// getBackendFactory returns a backend factory for the given type
func getBackendFactory(backendType backends.ClientBackendType) backends.ClientBackendFactory {
	return backendFactories[backendType]
}

// Default gin backend factory for backward compatibility
type ginBackendFactory struct{}

func (f *ginBackendFactory) CreateBackend(config *backends.ClientConfig) (backends.ClientBackend, error) {
	// This will be replaced by importing the gin package
	return nil, fmt.Errorf("gin backend not registered - import github.com/colonyos/colonies/pkg/client/gin")
}

func (f *ginBackendFactory) GetBackendType() backends.ClientBackendType {
	return backends.GinClientBackendType
}

func init() {
	// Register default backend factory
	RegisterBackendFactory(&ginBackendFactory{})
}