package gin

import (
	"github.com/gorilla/websocket"
	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/cluster"
)

// Factory implements the backends.RealtimeBackend interface for WebSocket
type Factory struct{}

// NewFactory creates a new WebSocket factory
func NewFactory() backends.RealtimeBackend {
	return &Factory{}
}

// CreateConnection creates a WebSocket connection from a raw websocket.Conn
func (f *Factory) CreateConnection(rawConn interface{}) (backends.RealtimeConnection, error) {
	wsConn, ok := rawConn.(*websocket.Conn)
	if !ok {
		return nil, ErrInvalidConnType
	}
	return NewWebSocketConnection(wsConn), nil
}

// CreateEventHandler creates an event handler
func (f *Factory) CreateEventHandler(relayServer interface{}) backends.RealtimeEventHandler {
	var relay *cluster.RelayServer
	if relayServer != nil {
		relay = relayServer.(*cluster.RelayServer)
	}
	return CreateEventHandler(relay)
}

// CreateTestableEventHandler creates a testable event handler
func (f *Factory) CreateTestableEventHandler(relayServer interface{}) backends.TestableRealtimeEventHandler {
	var relay *cluster.RelayServer
	if relayServer != nil {
		relay = relayServer.(*cluster.RelayServer)
	}
	return CreateTestableEventHandler(relay)
}

// CreateSubscriptionController creates a subscription controller
func (f *Factory) CreateSubscriptionController(eventHandler backends.RealtimeEventHandler) backends.RealtimeSubscriptionController {
	return NewSubscriptionController(eventHandler)
}