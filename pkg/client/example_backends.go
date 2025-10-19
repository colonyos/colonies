package client

import (
	"fmt"
	
	"github.com/colonyos/colonies/pkg/client/backends"
	"github.com/colonyos/colonies/pkg/client/libp2p"
)

// ExampleBackendUsage demonstrates how to create clients with different backends
func ExampleBackendUsage() {
	// Example 1: Default HTTP/Gin backend (backward compatible)
	client1 := CreateColoniesClient("localhost", 8080, true, false)
	defer client1.CloseClient()
	
	// Example 2: Explicit HTTP/Gin backend configuration
	ginConfig := &backends.ClientConfig{
		BackendType:   backends.GinClientBackendType,
		Host:          "localhost",
		Port:          8080,
		Insecure:      true,
		SkipTLSVerify: false,
	}
	client2 := CreateColoniesClientWithConfig(ginConfig)
	defer client2.CloseClient()
	
	// Example 3: LibP2P backend configuration (placeholder)
	libp2pConfig := &backends.ClientConfig{
		BackendType:   backends.LibP2PClientBackendType,
		Host:          "", // Not used for libp2p
		Port:          0,  // Not used for libp2p
		Insecure:      false,
		SkipTLSVerify: false,
	}
	
	// Register the libp2p backend factory
	RegisterBackendFactory(libp2p.GetLibP2PClientBackendFactory())
	
	// Create client with libp2p backend (would fail with "not implemented" errors)
	client3 := CreateColoniesClientWithConfig(libp2pConfig)
	defer client3.CloseClient()
	
	// Example 4: Programmatic backend selection
	backendType := backends.GinClientBackendType // Could be determined at runtime
	config := &backends.ClientConfig{
		BackendType:   backendType,
		Host:          "localhost",
		Port:          8080,
		Insecure:      true,
		SkipTLSVerify: false,
	}
	
	client4 := CreateColoniesClientWithConfig(config)
	defer client4.CloseClient()
	
	// Example 5: Custom backend registration (for third-party backends)
	// RegisterBackendFactory(myCustomBackendFactory)
	
	fmt.Println("Backend abstraction examples created successfully")
	fmt.Printf("Client 1 backend: %s\n", client1.GetConfig().BackendType)
	fmt.Printf("Client 2 backend: %s\n", client2.GetConfig().BackendType)
	fmt.Printf("Client 3 backend: %s\n", client3.GetConfig().BackendType)
	fmt.Printf("Client 4 backend: %s\n", client4.GetConfig().BackendType)
}