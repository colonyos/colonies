package libp2p

import (
	"fmt"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/sirupsen/logrus"
)

// Connection implements the backends.RealtimeConnection interface for libp2p streams
type Connection struct {
	stream network.Stream
}

// NewConnection creates a new libp2p connection wrapper
func NewConnection(rawConn interface{}) backends.RealtimeConnection {
	stream, ok := rawConn.(network.Stream)
	if !ok {
		logrus.Error("Invalid connection type for libp2p - expected network.Stream")
		return nil
	}
	
	return &Connection{
		stream: stream,
	}
}

// WriteMessage sends a message through the libp2p stream
func (c *Connection) WriteMessage(msgType int, data []byte) error {
	if c.stream == nil {
		return fmt.Errorf("stream is nil")
	}
	
	_, err := c.stream.Write(data)
	return err
}

// Close closes the libp2p stream
func (c *Connection) Close() error {
	if c.stream == nil {
		return nil
	}
	return c.stream.Close()
}

// IsOpen returns true if the stream is still open
func (c *Connection) IsOpen() bool {
	if c.stream == nil {
		return false
	}
	
	// In libp2p, we check if the connection is still active
	return c.stream.Conn().IsClosed() == false
}

// Compile-time check that Connection implements backends.RealtimeConnection
var _ backends.RealtimeConnection = (*Connection)(nil)