package coap

import (
	"github.com/colonyos/colonies/pkg/client/backends"
	_ "github.com/plgd-dev/go-coap/v3/udp"
)

// GetCoAPClientBackendFactory returns a CoAP client backend factory
// This function is used to avoid import cycles while allowing registration
func GetCoAPClientBackendFactory() backends.ClientBackendFactory {
	return NewCoAPClientBackendFactory()
}
