package libp2p

import "github.com/colonyos/colonies/pkg/client/backends"

// GetLibP2PClientBackendFactory returns a libp2p client backend factory
// This function is used to avoid import cycles while allowing registration
func GetLibP2PClientBackendFactory() backends.ClientBackendFactory {
	return NewLibP2PClientBackendFactory()
}