package libp2p

import (
	"testing"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/stretchr/testify/assert"
)

func TestNewBackend(t *testing.T) {
	backend := NewBackend()
	assert.NotNil(t, backend)

	// Check default mode
	assert.Equal(t, "release", backend.GetMode())
}

func TestBackendSetGetMode(t *testing.T) {
	backend := NewBackend()

	// Test setting different modes
	backend.SetMode("debug")
	assert.Equal(t, "debug", backend.GetMode())

	backend.SetMode("release")
	assert.Equal(t, "release", backend.GetMode())

	backend.SetMode("test")
	assert.Equal(t, "test", backend.GetMode())
}

func TestBackendImplementsInterface(t *testing.T) {
	backend := NewBackend()

	// Verify it implements backends.Backend interface
	var _ backends.Backend = backend
}

func TestBackendLogger(t *testing.T) {
	backend := NewBackend()
	logger := backend.Logger()
	assert.NotNil(t, logger)

	// Logger should be a no-op middleware that just calls Next()
	// We can't easily test the middleware without a full context setup
}

func TestBackendRecovery(t *testing.T) {
	backend := NewBackend()
	recovery := backend.Recovery()
	assert.NotNil(t, recovery)

	// Recovery should be a no-op middleware
}

func TestBackendNewEnginePanics(t *testing.T) {
	backend := NewBackend()

	// NewEngine should panic for libp2p backend
	assert.Panics(t, func() {
		backend.NewEngine()
	}, "libp2p backend doesn't use engines")
}

func TestBackendNewEngineWithDefaultsPanics(t *testing.T) {
	backend := NewBackend()

	// NewEngineWithDefaults should panic for libp2p backend
	assert.Panics(t, func() {
		backend.NewEngineWithDefaults()
	}, "libp2p backend doesn't use engines")
}

func TestBackendNewServerPanics(t *testing.T) {
	backend := NewBackend()

	// NewServer should panic for libp2p backend
	assert.Panics(t, func() {
		backend.NewServer(8080, nil)
	}, "libp2p backend doesn't use the Engine/Server pattern")
}

func TestBackendNewServerWithAddrPanics(t *testing.T) {
	backend := NewBackend()

	// NewServerWithAddr should panic for libp2p backend
	assert.Panics(t, func() {
		backend.NewServerWithAddr("localhost:8080", nil)
	}, "libp2p backend doesn't use the Engine/Server pattern")
}

func TestNewRealtimeBackend(t *testing.T) {
	rtBackend := NewRealtimeBackend()
	assert.NotNil(t, rtBackend)

	// Verify it implements backends.RealtimeBackend interface
	var _ backends.RealtimeBackend = rtBackend
}

func TestRealtimeBackendCreateConnection(t *testing.T) {
	rtBackend := NewRealtimeBackend()

	// Test with invalid connection type - should return nil without error
	// (the error is logged but not returned)
	conn, err := rtBackend.CreateConnection("not a stream")
	assert.NoError(t, err) // No error is returned, just logs
	assert.Nil(t, conn)    // Connection is nil for invalid type
}

func TestRealtimeBackendCreateEventHandler(t *testing.T) {
	rtBackend := NewRealtimeBackend()

	// Create event handler with nil relay server
	handler := rtBackend.CreateEventHandler(nil)
	assert.NotNil(t, handler)
}

func TestRealtimeBackendCreateTestableEventHandler(t *testing.T) {
	rtBackend := NewRealtimeBackend()

	// Create testable event handler
	handler := rtBackend.CreateTestableEventHandler(nil)
	assert.NotNil(t, handler)
}

func TestRealtimeBackendCreateSubscriptionController(t *testing.T) {
	rtBackend := NewRealtimeBackend()

	// Create event handler first
	eventHandler := rtBackend.CreateEventHandler(nil)

	// Create subscription controller
	controller := rtBackend.CreateSubscriptionController(eventHandler)
	assert.NotNil(t, controller)
}

func TestNewFullBackend(t *testing.T) {
	fullBackend := NewFullBackend()
	assert.NotNil(t, fullBackend)

	// Check default mode
	assert.Equal(t, "release", fullBackend.GetMode())

	// Verify it implements backends.FullBackend interface
	var _ backends.FullBackend = fullBackend
}

func TestFullBackendHasBothInterfaces(t *testing.T) {
	fullBackend := NewFullBackend()

	// Should implement both Backend and RealtimeBackend
	var _ backends.Backend = fullBackend
	var _ backends.RealtimeBackend = fullBackend
}

func TestFullBackendModeManagement(t *testing.T) {
	fullBackend := NewFullBackend()

	// Test mode changes
	fullBackend.SetMode("debug")
	assert.Equal(t, "debug", fullBackend.GetMode())

	// Mode should be shared between Backend and RealtimeBackend
	fullBackend.SetMode("production")
	assert.Equal(t, "production", fullBackend.GetMode())
}

func TestFullBackendMiddleware(t *testing.T) {
	fullBackend := NewFullBackend()

	// Test logger middleware
	logger := fullBackend.Logger()
	assert.NotNil(t, logger)

	// Test recovery middleware
	recovery := fullBackend.Recovery()
	assert.NotNil(t, recovery)
}

func TestFullBackendPanicMethods(t *testing.T) {
	fullBackend := NewFullBackend()

	// All engine/server creation methods should panic
	assert.Panics(t, func() {
		fullBackend.NewEngine()
	})

	assert.Panics(t, func() {
		fullBackend.NewEngineWithDefaults()
	})

	assert.Panics(t, func() {
		fullBackend.NewServer(8080, nil)
	})

	assert.Panics(t, func() {
		fullBackend.NewServerWithAddr(":8080", nil)
	})
}
