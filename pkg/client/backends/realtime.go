package backends

// RealtimeConnection represents a generic realtime connection interface
// Different backends can implement this using WebSockets, libp2p pubsub, etc.
type RealtimeConnection interface {
	// WriteMessage writes a message to the realtime connection
	WriteMessage(messageType int, data []byte) error
	
	// ReadMessage reads a message from the realtime connection
	ReadMessage() (messageType int, data []byte, err error)
	
	// Close closes the realtime connection
	Close() error
	
	// SetReadLimit sets the maximum size for incoming messages
	SetReadLimit(limit int64)
}

// RealtimeBackend defines the interface for establishing realtime connections
type RealtimeBackend interface {
	// EstablishRealtimeConn establishes a realtime connection for real-time operations
	EstablishRealtimeConn(jsonString string) (RealtimeConnection, error)
}

// MessageType constants for realtime connections (WebSocket compatible)
const (
	TextMessage   = 1
	BinaryMessage = 2
	CloseMessage  = 8
	PingMessage   = 9
	PongMessage   = 10
)