package registry

import (
	"net/http/httptest"
	"testing"

	"github.com/colonyos/colonies/pkg/backends"
	ginbackends "github.com/colonyos/colonies/pkg/backends/gin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandlerRegistry(t *testing.T) {
	registry := NewHandlerRegistry()
	
	// Test registration
	testHandler := func(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
		c.String(200, "test_response")
	}
	
	err := registry.RegisterGin("test_payload", testHandler)
	assert.Nil(t, err)
	
	// Test duplicate registration
	err = registry.RegisterGin("test_payload", testHandler)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "already registered")
	
	// Test handler retrieval
	handler, exists := registry.GetHandler("test_payload")
	assert.True(t, exists)
	assert.NotNil(t, handler)
	
	// Test non-existent handler
	_, exists = registry.GetHandler("non_existent")
	assert.False(t, exists)
	
	// Test GetRegisteredTypes
	types := registry.GetRegisteredTypes()
	assert.Len(t, types, 1)
	assert.Contains(t, types, "test_payload")
}

func TestHandlerRegistryRequest(t *testing.T) {
	registry := NewHandlerRegistry()
	gin.SetMode(gin.TestMode)
	
	// Register test handler
	testHandler := func(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
		c.String(200, "handler_called")
	}
	
	err := registry.RegisterGin("test_payload", testHandler)
	assert.Nil(t, err)
	
	// Test request handling
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ginCtxWrapper := ginbackends.NewContext(ctx)
	ginCtxAdapter := backends.NewGinContextAdapter(ginCtxWrapper)
	
	handled := registry.HandleRequest(ginCtxAdapter, "test_id", "test_payload", "{}")
	assert.True(t, handled)
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "handler_called", recorder.Body.String())
	
	// Test unhandled request
	recorder2 := httptest.NewRecorder()
	ctx2, _ := gin.CreateTestContext(recorder2)
	ginCtxWrapper2 := ginbackends.NewContext(ctx2)
	ginCtxAdapter2 := backends.NewGinContextAdapter(ginCtxWrapper2)
	
	handled = registry.HandleRequest(ginCtxAdapter2, "test_id", "unknown_payload", "{}")
	assert.False(t, handled)
}