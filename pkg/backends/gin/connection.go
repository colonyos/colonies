package gin

import (
	"github.com/gorilla/websocket"
	"github.com/colonyos/colonies/pkg/backends"
)

// WebSocketConnection implements the backends.RealtimeConnection interface
type WebSocketConnection struct {
	conn *websocket.Conn
}

// NewWebSocketConnection creates a new WebSocketConnection
func NewWebSocketConnection(conn *websocket.Conn) backends.RealtimeConnection {
	return &WebSocketConnection{conn: conn}
}

// WriteMessage implements backends.RealtimeConnection
func (w *WebSocketConnection) WriteMessage(msgType int, data []byte) error {
	if w.conn == nil {
		return ErrConnectionClosed
	}
	return w.conn.WriteMessage(msgType, data)
}

// Close implements backends.RealtimeConnection
func (w *WebSocketConnection) Close() error {
	if w.conn == nil {
		return nil
	}
	err := w.conn.Close()
	w.conn = nil
	return err
}

// IsOpen implements backends.RealtimeConnection
func (w *WebSocketConnection) IsOpen() bool {
	return w.conn != nil
}