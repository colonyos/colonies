// Package backends provides generic interfaces for HTTP server backends
// This allows the application to work with different web frameworks
package backends

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

// Context represents a generic HTTP request/response context
type Context interface {
	// Request/Response handling
	String(code int, format string, values ...interface{})
	JSON(code int, obj interface{})
	XML(code int, obj interface{})
	Data(code int, contentType string, data []byte)
	Status(code int)
	
	// Request access
	Request() *http.Request
	ReadBody() ([]byte, error)
	
	// Headers
	GetHeader(key string) string
	Header(key, value string)
	
	// URL parameters and query strings
	Param(key string) string
	Query(key string) string
	DefaultQuery(key, defaultValue string) string
	
	// Form data
	PostForm(key string) string
	DefaultPostForm(key, defaultValue string) string
	
	// Data binding
	Bind(obj interface{}) error
	ShouldBind(obj interface{}) error
	BindJSON(obj interface{}) error
	ShouldBindJSON(obj interface{}) error
	
	// Context storage
	Set(key string, value interface{})
	Get(key string) (value interface{}, exists bool)
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetInt64(key string) int64
	GetFloat64(key string) float64
	
	// Flow control
	Abort()
	AbortWithStatus(code int)
	AbortWithStatusJSON(code int, jsonObj interface{})
	IsAborted() bool
	Next()
}

// HandlerFunc represents a generic HTTP handler function
type HandlerFunc func(Context)

// MiddlewareFunc is an alias for HandlerFunc to represent middleware
type MiddlewareFunc = HandlerFunc

// Engine represents a generic HTTP engine/router
type Engine interface {
	// HTTP methods
	GET(relativePath string, handlers ...HandlerFunc)
	POST(relativePath string, handlers ...HandlerFunc)
	PUT(relativePath string, handlers ...HandlerFunc)
	DELETE(relativePath string, handlers ...HandlerFunc)
	PATCH(relativePath string, handlers ...HandlerFunc)
	
	// Middleware
	Use(middleware ...HandlerFunc)
	
	// HTTP Handler for integration with net/http
	Handler() http.Handler
}

// Server represents a generic HTTP server
type Server interface {
	// Server lifecycle
	ListenAndServe() error
	ListenAndServeTLS(certFile, keyFile string) error
	Shutdown(ctx context.Context) error
	ShutdownWithTimeout(timeout time.Duration) error
	
	// Configuration
	SetAddr(addr string)
	GetAddr() string
	SetReadTimeout(timeout time.Duration)
	SetWriteTimeout(timeout time.Duration)
	SetIdleTimeout(timeout time.Duration)
	SetReadHeaderTimeout(timeout time.Duration)
	
	// Access to underlying components
	Engine() Engine
	HTTPServer() *http.Server
}

// Backend represents a complete HTTP backend implementation
type Backend interface {
	// Factory methods
	NewEngine() Engine
	NewEngineWithDefaults() Engine
	NewServer(port int, engine Engine) Server
	NewServerWithAddr(addr string, engine Engine) Server
	
	// Backend-specific configuration
	SetMode(mode string)
	GetMode() string
	
	// Common middleware
	Logger() MiddlewareFunc
	Recovery() MiddlewareFunc
}

// CORSConfig represents CORS configuration options
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// CORSBackend represents a backend that supports CORS middleware
type CORSBackend interface {
	Backend
	
	// CORS middleware
	CORS() MiddlewareFunc
	CORSWithConfig(config CORSConfig) MiddlewareFunc
}

// ResponseWriter represents a generic response writer interface
type ResponseWriter interface {
	http.ResponseWriter
	
	// Additional methods that some frameworks provide
	Size() int
	Status() int
	Written() bool
	WriteHeaderNow()
}

// LogFormatter represents a log formatter function
type LogFormatter func(params LogFormatterParams) string

// LogFormatterParams contains the parameters for log formatting
type LogFormatterParams struct {
	Request      *http.Request
	TimeStamp    time.Time
	StatusCode   int
	Latency      time.Duration
	ClientIP     string
	Method       string
	Path         string
	ErrorMessage string
	BodySize     int
	Keys         map[string]interface{}
}

// LoggingBackend represents a backend that supports custom logging
type LoggingBackend interface {
	Backend
	
	// Logging middleware with custom configuration
	LoggerWithFormatter(formatter LogFormatter) MiddlewareFunc
	LoggerWithWriter(out io.Writer, notlogged ...string) MiddlewareFunc
}

// AuthBackend represents a backend that supports authentication middleware
type AuthBackend interface {
	Backend
	
	// Authentication middleware
	BasicAuth(accounts map[string]string) MiddlewareFunc
}

// =============================================================================
// Realtime Communication Interfaces
// =============================================================================

// RealtimeConnection represents a generic realtime connection interface
// This abstracts away specific transport implementations (WebSocket, gRPC, libp2p, etc.)
type RealtimeConnection interface {
	// WriteMessage sends a message through the connection
	WriteMessage(msgType int, data []byte) error
	// Close closes the connection
	Close() error
	// IsOpen returns true if the connection is still open
	IsOpen() bool
}

// RealtimeSubscription represents a subscription to process events
type RealtimeSubscription struct {
	Connection   RealtimeConnection
	MsgType      int
	Timeout      int
	ExecutorType string
	State        int
	ProcessID    string
}

// RealtimeEventHandler handles process events and manages subscriptions
type RealtimeEventHandler interface {
	// Signal sends a process event to all registered listeners
	Signal(process *core.Process)
	// Subscribe registers a subscription and returns channels for process events and errors
	Subscribe(executorType string, state int, processID string, ctx context.Context) (chan *core.Process, chan error)
	// WaitForProcess waits for a specific process state change
	WaitForProcess(executorType string, state int, processID string, ctx context.Context) (*core.Process, error)
	// Stop stops the event handler
	Stop()
}

// TestableRealtimeEventHandler extends RealtimeEventHandler with methods for testing
type TestableRealtimeEventHandler interface {
	RealtimeEventHandler
	// NumberOfListeners returns listener counts for testing
	NumberOfListeners(executorType string, state int) (int, int, int)
	// HasStopped returns whether the handler has stopped for testing
	HasStopped() bool
}

// RealtimeSubscriptionController manages subscriptions using the abstract Connection interface
type RealtimeSubscriptionController interface {
	// AddProcessesSubscriber adds a subscription for all processes of a certain type
	AddProcessesSubscriber(executorID string, subscription *RealtimeSubscription) error
	// AddProcessSubscriber adds a subscription for a specific process
	AddProcessSubscriber(executorID string, process *core.Process, subscription *RealtimeSubscription) error
}

// RealtimeBackend represents a backend that supports realtime communication
type RealtimeBackend interface {
	// CreateConnection creates a connection from a raw connection (e.g., *websocket.Conn)
	CreateConnection(rawConn interface{}) (RealtimeConnection, error)
	// CreateEventHandler creates an event handler
	CreateEventHandler(relayServer interface{}) RealtimeEventHandler
	// CreateTestableEventHandler creates a testable event handler
	CreateTestableEventHandler(relayServer interface{}) TestableRealtimeEventHandler
	// CreateSubscriptionController creates a subscription controller
	CreateSubscriptionController(eventHandler RealtimeEventHandler) RealtimeSubscriptionController
}

// FullBackend represents a complete backend implementation with both HTTP and realtime support
type FullBackend interface {
	Backend
	RealtimeBackend
}