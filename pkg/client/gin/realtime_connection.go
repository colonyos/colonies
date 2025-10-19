package gin

import (
	"github.com/colonyos/colonies/pkg/client/backends"
	"github.com/gorilla/websocket"
)

// WebSocketRealtimeConnection wraps a WebSocket connection to implement RealtimeConnection
type WebSocketRealtimeConnection struct {
	conn *websocket.Conn
}

// NewWebSocketRealtimeConnection creates a new WebSocket-based realtime connection
func NewWebSocketRealtimeConnection(conn *websocket.Conn) *WebSocketRealtimeConnection {
	return &WebSocketRealtimeConnection{
		conn: conn,
	}
}

// WriteMessage writes a message to the WebSocket connection
func (w *WebSocketRealtimeConnection) WriteMessage(messageType int, data []byte) error {
	return w.conn.WriteMessage(messageType, data)
}

// ReadMessage reads a message from the WebSocket connection
func (w *WebSocketRealtimeConnection) ReadMessage() (messageType int, data []byte, err error) {
	return w.conn.ReadMessage()
}

// Close closes the WebSocket connection
func (w *WebSocketRealtimeConnection) Close() error {
	return w.conn.Close()
}

// SetReadLimit sets the maximum size for incoming messages
func (w *WebSocketRealtimeConnection) SetReadLimit(limit int64) {
	w.conn.SetReadLimit(limit)
}

// Compile-time check that WebSocketRealtimeConnection implements RealtimeConnection
var _ backends.RealtimeConnection = (*WebSocketRealtimeConnection)(nil)