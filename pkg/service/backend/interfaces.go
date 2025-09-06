package backend

import (
	"context"
)

// BackendType defines the type of backend
type BackendType string

const (
	HTTPBackend   BackendType = "http"
	GRPCBackend   BackendType = "grpc"
	LibP2PBackend BackendType = "libp2p"
)

// Request represents an incoming request regardless of backend
type Request interface {
	GetBody() ([]byte, error)
	GetHeader(key string) string
	GetMethod() string
	GetPath() string
	GetRemoteAddr() string
	GetQuery(key string) string
	GetParam(key string) string
}

// Response represents an outgoing response
type Response interface {
	SetStatus(code int)
	SetHeader(key, value string)
	Write(data []byte) error
	WriteJSON(data interface{}) error
	WriteString(data string) error
	GetStatus() int
}

// Context provides request/response context abstraction
type Context interface {
	Request() Request
	Response() Response
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
	Abort()
	IsAborted() bool
	Context() context.Context
}

// HandlerFunc represents a generic request handler
type HandlerFunc func(ctx Context)

// MiddlewareFunc represents middleware function
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Backend defines the server backend interface
type Backend interface {
	Start(addr string) error
	Stop() error
	Handle(method, path string, handler HandlerFunc)
	Use(middleware MiddlewareFunc)
	SetCORS(config CORSConfig)
	Name() string
	Type() BackendType
}

// CORSConfig defines CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
}

// BackendConfig defines backend configuration
type BackendConfig struct {
	Type     BackendType            `yaml:"type"`
	Name     string                 `yaml:"name"`
	Address  string                 `yaml:"address"`
	Settings map[string]interface{} `yaml:"settings"`
}