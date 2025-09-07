package registry

import (
	"github.com/colonyos/colonies/pkg/backends"
	"github.com/gin-gonic/gin"
)

// GinHandlerFunc defines the legacy signature for gin handlers
type GinHandlerFunc func(c *gin.Context, recoveredID string, payloadType string, jsonString string)

// GinHandlerFuncWithRawRequest defines the legacy signature for gin handlers with raw request
type GinHandlerFuncWithRawRequest func(c *gin.Context, recoveredID string, payloadType string, jsonString string, rawRequest string)

// RegisterGin registers a legacy Gin handler by wrapping it in the new interface
func (r *HandlerRegistry) RegisterGin(payloadType string, handler GinHandlerFunc) error {
	wrappedHandler := func(c backends.Context, recoveredID string, payloadType string, jsonString string) {
		// Extract the underlying gin.Context from our generic context
		if ginAdapter, ok := c.(*backends.GinContextAdapter); ok {
			handler(ginAdapter.GinContext(), recoveredID, payloadType, jsonString)
		} else {
			// This shouldn't happen if we're using the correct backend
			panic("Expected GinContextAdapter but got different context type")
		}
	}
	
	return r.Register(payloadType, wrappedHandler)
}

// RegisterGinWithRawRequest registers a legacy Gin handler that needs raw request access
func (r *HandlerRegistry) RegisterGinWithRawRequest(payloadType string, handler GinHandlerFuncWithRawRequest) error {
	wrappedHandler := func(c backends.Context, recoveredID string, payloadType string, jsonString string, rawRequest string) {
		// Extract the underlying gin.Context from our generic context
		if ginAdapter, ok := c.(*backends.GinContextAdapter); ok {
			handler(ginAdapter.GinContext(), recoveredID, payloadType, jsonString, rawRequest)
		} else {
			// This shouldn't happen if we're using the correct backend
			panic("Expected GinContextAdapter but got different context type")
		}
	}
	
	return r.RegisterWithRawRequest(payloadType, wrappedHandler)
}