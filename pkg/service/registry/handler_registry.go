package registry

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
)

// HandlerFunc defines the signature for request handlers
type HandlerFunc func(c *gin.Context, recoveredID string, payloadType string, jsonString string)

// HandlerFuncWithRawRequest defines the signature for request handlers that need access to the raw request
type HandlerFuncWithRawRequest func(c *gin.Context, recoveredID string, payloadType string, jsonString string, rawRequest string)

// HandlerRegistry manages the registration of all RPC handlers
type HandlerRegistry struct {
	handlers             map[string]HandlerFunc
	handlersWithRawReq   map[string]HandlerFuncWithRawRequest
	mutex                sync.RWMutex
}

// NewHandlerRegistry creates a new handler registry
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers:           make(map[string]HandlerFunc),
		handlersWithRawReq: make(map[string]HandlerFuncWithRawRequest),
	}
}

// Register registers a handler function for a specific payload type
func (r *HandlerRegistry) Register(payloadType string, handler HandlerFunc) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.handlers[payloadType]; exists {
		return fmt.Errorf("handler for payload type %s is already registered", payloadType)
	}
	if _, exists := r.handlersWithRawReq[payloadType]; exists {
		return fmt.Errorf("handler for payload type %s is already registered", payloadType)
	}
	
	r.handlers[payloadType] = handler
	return nil
}

// RegisterWithRawRequest registers a handler function that needs access to the raw request
func (r *HandlerRegistry) RegisterWithRawRequest(payloadType string, handler HandlerFuncWithRawRequest) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.handlers[payloadType]; exists {
		return fmt.Errorf("handler for payload type %s is already registered", payloadType)
	}
	if _, exists := r.handlersWithRawReq[payloadType]; exists {
		return fmt.Errorf("handler for payload type %s is already registered", payloadType)
	}
	
	r.handlersWithRawReq[payloadType] = handler
	return nil
}

// GetHandler retrieves a handler for a specific payload type
func (r *HandlerRegistry) GetHandler(payloadType string) (HandlerFunc, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	handler, exists := r.handlers[payloadType]
	return handler, exists
}

// GetRegisteredTypes returns all registered payload types
func (r *HandlerRegistry) GetRegisteredTypes() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	types := make([]string, 0, len(r.handlers)+len(r.handlersWithRawReq))
	for payloadType := range r.handlers {
		types = append(types, payloadType)
	}
	for payloadType := range r.handlersWithRawReq {
		types = append(types, payloadType)
	}
	return types
}

// HandleRequest handles an RPC request by looking up the appropriate handler
func (r *HandlerRegistry) HandleRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) bool {
	handler, exists := r.GetHandler(payloadType)
	if exists {
		handler(c, recoveredID, payloadType, jsonString)
		return true
	}
	return false
}

// HandleRequestWithRaw handles an RPC request that may need access to raw request data
func (r *HandlerRegistry) HandleRequestWithRaw(c *gin.Context, recoveredID string, payloadType string, jsonString string, rawRequest string) bool {
	// First try handlers that need raw request access
	r.mutex.RLock()
	handlerWithRaw, exists := r.handlersWithRawReq[payloadType]
	r.mutex.RUnlock()
	
	if exists {
		handlerWithRaw(c, recoveredID, payloadType, jsonString, rawRequest)
		return true
	}
	
	// Fall back to regular handlers
	return r.HandleRequest(c, recoveredID, payloadType, jsonString)
}