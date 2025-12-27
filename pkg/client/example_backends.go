package client

import (
	"fmt"

	"github.com/colonyos/colonies/pkg/client/backends"
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

	// Example 3: Programmatic backend selection
	backendType := backends.GinClientBackendType // Could be determined at runtime
	config := &backends.ClientConfig{
		BackendType:   backendType,
		Host:          "localhost",
		Port:          8080,
		Insecure:      true,
		SkipTLSVerify: false,
	}

	client3 := CreateColoniesClientWithConfig(config)
	defer client3.CloseClient()

	fmt.Println("Backend abstraction examples created successfully")
	fmt.Printf("Client 1 backend: %s\n", client1.GetConfig().BackendType)
	fmt.Printf("Client 2 backend: %s\n", client2.GetConfig().BackendType)
	fmt.Printf("Client 3 backend: %s\n", client3.GetConfig().BackendType)
}
