package gin

import "github.com/colonyos/colonies/pkg/client/backends"

// GetGinClientBackendFactory returns a gin client backend factory
// This function is used to avoid import cycles while allowing registration
func GetGinClientBackendFactory() backends.ClientBackendFactory {
	return NewGinClientBackendFactory()
}