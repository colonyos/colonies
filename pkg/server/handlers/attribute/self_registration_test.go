package attribute

import (
	"testing"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/registry"
	"github.com/stretchr/testify/assert"
)

// SimpleTestServer implements the minimal ColoniesServer interface for testing
type SimpleTestServer struct{}

func (s *SimpleTestServer) HandleHTTPError(c backends.Context, err error, errorCode int) bool {
	return false
}

func (s *SimpleTestServer) SendHTTPReply(c backends.Context, payloadType string, jsonString string) {
}

func (s *SimpleTestServer) Validator() security.Validator {
	return nil
}

func (s *SimpleTestServer) ProcessDB() database.ProcessDatabase {
	return nil
}

func (s *SimpleTestServer) AttributeDB() database.AttributeDatabase {
	return nil
}

func TestAttributeHandlerSelfRegistration(t *testing.T) {
	// Create test server
	testServer := &SimpleTestServer{}
	
	// Create handlers
	handlers := NewHandlers(testServer)
	
	// Create registry
	handlerRegistry := registry.NewHandlerRegistry()
	
	// Test registration
	err := handlers.RegisterHandlers(handlerRegistry)
	assert.Nil(t, err)
	
	// Verify handlers are registered
	addHandler, exists := handlerRegistry.GetHandler(rpc.AddAttributePayloadType)
	assert.True(t, exists)
	assert.NotNil(t, addHandler)
	
	getHandler, exists := handlerRegistry.GetHandler(rpc.GetAttributePayloadType)
	assert.True(t, exists)
	assert.NotNil(t, getHandler)
	
	// Verify correct number of handlers registered
	registeredTypes := handlerRegistry.GetRegisteredTypes()
	assert.Len(t, registeredTypes, 2)
	assert.Contains(t, registeredTypes, rpc.AddAttributePayloadType)
	assert.Contains(t, registeredTypes, rpc.GetAttributePayloadType)
}

func TestAttributeHandlerRegistrationError(t *testing.T) {
	testServer := &SimpleTestServer{}
	handlers := NewHandlers(testServer)
	handlerRegistry := registry.NewHandlerRegistry()
	
	// Register once
	err := handlers.RegisterHandlers(handlerRegistry)
	assert.Nil(t, err)
	
	// Try to register again - should fail
	err = handlers.RegisterHandlers(handlerRegistry)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "already registered")
}