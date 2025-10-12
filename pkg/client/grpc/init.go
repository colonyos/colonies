package grpc

import (
	"github.com/colonyos/colonies/pkg/client/backends"
	_ "google.golang.org/grpc"
)

// GetGRPCClientBackendFactory returns a gRPC client backend factory
// This function is used to avoid import cycles while allowing registration
func GetGRPCClientBackendFactory() backends.ClientBackendFactory {
	return NewGRPCClientBackendFactory()
}
