package libp2p

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/backends"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/sirupsen/logrus"
)

// StreamContext adapts a libp2p stream to the backends.Context interface
type StreamContext struct {
	stream network.Stream
	pubsub *pubsub.PubSub
	
	// Request data
	requestData []byte
	peerID      string
	
	// Context storage
	store map[string]interface{}
	
	// Flow control
	aborted bool
}

// NewStreamContext creates a new stream context
func NewStreamContext(stream network.Stream, pubsub *pubsub.PubSub) *StreamContext {
	return &StreamContext{
		stream: stream,
		pubsub: pubsub,
		peerID: stream.Conn().RemotePeer().String(),
		store:  make(map[string]interface{}),
	}
}

// GetHeader returns a header value (libp2p doesn't have headers, return empty)
func (c *StreamContext) GetHeader(key string) string {
	// LibP2P streams don't have headers like HTTP
	return ""
}

// Query returns a query parameter (libp2p doesn't have query params, return empty)
func (c *StreamContext) Query(key string) string {
	// LibP2P streams don't have query parameters like HTTP
	return ""
}

// DefaultQuery returns a query parameter with default value
func (c *StreamContext) DefaultQuery(key, defaultValue string) string {
	// LibP2P streams don't have query parameters like HTTP
	return defaultValue
}

// Param returns a URL parameter (libp2p doesn't have URL params, return empty)
func (c *StreamContext) Param(key string) string {
	// LibP2P streams don't have URL parameters like HTTP
	return ""
}

// PostForm returns a POST form value (libp2p doesn't have forms, return empty)
func (c *StreamContext) PostForm(key string) string {
	// LibP2P streams don't have POST forms like HTTP
	return ""
}

// DefaultPostForm returns a POST form value with default
func (c *StreamContext) DefaultPostForm(key, defaultValue string) string {
	// LibP2P streams don't have POST forms like HTTP
	return defaultValue
}

// Bind reads and unmarshals the request data
func (c *StreamContext) Bind(obj interface{}) error {
	if len(c.requestData) == 0 {
		// Read from stream if not already read
		buf := make([]byte, 4096)
		n, err := c.stream.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read from stream: %w", err)
		}
		c.requestData = buf[:n]
	}
	
	return json.Unmarshal(c.requestData, obj)
}

// ShouldBind is like Bind but returns error instead of aborting
func (c *StreamContext) ShouldBind(obj interface{}) error {
	return c.Bind(obj)
}

// BindJSON reads and unmarshals JSON data
func (c *StreamContext) BindJSON(obj interface{}) error {
	return c.Bind(obj)
}

// ShouldBindJSON is like BindJSON but returns error instead of aborting
func (c *StreamContext) ShouldBindJSON(obj interface{}) error {
	return c.Bind(obj)
}

// JSON writes a JSON response to the stream
func (c *StreamContext) JSON(code int, obj interface{}) {
	data, err := json.Marshal(obj)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal JSON response")
		c.sendError(fmt.Sprintf("JSON marshal error: %v", err))
		return
	}
	
	_, err = c.stream.Write(data)
	if err != nil {
		logrus.WithError(err).WithField("peer_id", c.peerID).Error("Failed to write response to stream")
	}
}

// String writes a string response to the stream
func (c *StreamContext) String(code int, format string, values ...interface{}) {
	data := fmt.Sprintf(format, values...)
	_, err := c.stream.Write([]byte(data))
	if err != nil {
		logrus.WithError(err).WithField("peer_id", c.peerID).Error("Failed to write string response to stream")
	}
}

// XML writes an XML response to the stream
func (c *StreamContext) XML(code int, obj interface{}) {
	// For simplicity, convert to JSON for libp2p
	c.JSON(code, obj)
}

// Data writes raw data to the stream
func (c *StreamContext) Data(code int, contentType string, data []byte) {
	_, err := c.stream.Write(data)
	if err != nil {
		logrus.WithError(err).WithField("peer_id", c.peerID).Error("Failed to write data response to stream")
	}
}

// Status sets the response status (no-op for libp2p)
func (c *StreamContext) Status(code int) {
	// LibP2P streams don't have HTTP status codes
}

// Request returns a dummy HTTP request (libp2p doesn't have HTTP requests)
func (c *StreamContext) Request() *http.Request {
	// LibP2P streams don't have HTTP requests, return nil
	return nil
}

// ReadBody reads the request body
func (c *StreamContext) ReadBody() ([]byte, error) {
	if len(c.requestData) == 0 {
		buf := make([]byte, 4096)
		n, err := c.stream.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to read from stream: %w", err)
		}
		c.requestData = buf[:n]
	}
	return c.requestData, nil
}

// Header sets a response header (no-op for libp2p)
func (c *StreamContext) Header(key, value string) {
	// LibP2P streams don't have headers like HTTP
}

// SendError sends an error response
func (c *StreamContext) SendError(message string) {
	c.sendError(message)
}

// sendError is a helper to send error responses
func (c *StreamContext) sendError(message string) {
	failure := core.CreateFailure(400, message)
	failureJSON, _ := failure.ToJSON()
	rpcReply, err := rpc.CreateRPCReplyMsg("error", failureJSON)
	if err != nil {
		logrus.WithError(err).Error("Failed to create RPC reply message")
		c.stream.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", message)))
		return
	}
	
	replyJSON, err := rpcReply.ToJSON()
	if err != nil {
		logrus.WithError(err).Error("Failed to create error response")
		c.stream.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", message)))
		return
	}
	
	_, err = c.stream.Write([]byte(replyJSON))
	if err != nil {
		logrus.WithError(err).WithField("peer_id", c.peerID).Error("Failed to write error response to stream")
	}
}

// GetPeerID returns the peer ID of the remote peer
func (c *StreamContext) GetPeerID() string {
	return c.peerID
}

// GetStream returns the underlying stream
func (c *StreamContext) GetStream() network.Stream {
	return c.stream
}

// GetPubSub returns the pubsub instance
func (c *StreamContext) GetPubSub() *pubsub.PubSub {
	return c.pubsub
}

// Context storage methods
func (c *StreamContext) Set(key string, value interface{}) {
	c.store[key] = value
}

func (c *StreamContext) Get(key string) (value interface{}, exists bool) {
	value, exists = c.store[key]
	return
}

func (c *StreamContext) GetString(key string) string {
	if value, exists := c.store[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func (c *StreamContext) GetBool(key string) bool {
	if value, exists := c.store[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

func (c *StreamContext) GetInt(key string) int {
	if value, exists := c.store[key]; exists {
		if i, ok := value.(int); ok {
			return i
		}
	}
	return 0
}

func (c *StreamContext) GetInt64(key string) int64 {
	if value, exists := c.store[key]; exists {
		if i, ok := value.(int64); ok {
			return i
		}
	}
	return 0
}

func (c *StreamContext) GetFloat64(key string) float64 {
	if value, exists := c.store[key]; exists {
		if f, ok := value.(float64); ok {
			return f
		}
	}
	return 0.0
}

// Flow control methods
func (c *StreamContext) Abort() {
	c.aborted = true
}

func (c *StreamContext) AbortWithStatus(code int) {
	c.aborted = true
}

func (c *StreamContext) AbortWithStatusJSON(code int, jsonObj interface{}) {
	c.JSON(code, jsonObj)
	c.aborted = true
}

func (c *StreamContext) IsAborted() bool {
	return c.aborted
}

func (c *StreamContext) Next() {
	// No-op for libp2p, middleware pattern doesn't apply the same way
}

// Compile-time check that StreamContext implements backends.Context
var _ backends.Context = (*StreamContext)(nil)