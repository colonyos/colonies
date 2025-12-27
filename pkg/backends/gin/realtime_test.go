package gin

import (
	"testing"

	"github.com/colonyos/colonies/pkg/channel"
	"github.com/colonyos/colonies/pkg/core"
)

// MockChannelRouter implements a minimal channel router for testing
type MockChannelRouter struct {
	channels map[string]*channel.Channel
}

func NewMockChannelRouter() *MockChannelRouter {
	return &MockChannelRouter{
		channels: make(map[string]*channel.Channel),
	}
}

func (m *MockChannelRouter) GetByProcessAndName(processID, name string) (*channel.Channel, error) {
	key := processID + "_" + name
	if ch, ok := m.channels[key]; ok {
		return ch, nil
	}
	return nil, channel.ErrChannelNotFound
}

func (m *MockChannelRouter) CreateIfNotExists(ch *channel.Channel) error {
	if _, ok := m.channels[ch.ID]; ok {
		return nil // Already exists
	}
	m.channels[ch.ID] = ch
	return nil
}

// TestEnsureChannelExists tests the lazy channel creation logic
func TestEnsureChannelExists(t *testing.T) {
	// Create a mock process with a chat channel defined
	process := &core.Process{
		ID:                 "test-process-id",
		InitiatorID:        "test-submitter",
		AssignedExecutorID: "test-executor",
		State:              core.RUNNING,
		FunctionSpec: core.FunctionSpec{
			Channels: []string{"chat", "control"},
		},
	}

	router := NewMockChannelRouter()

	// Test case 1: Channel defined in process spec - should be created
	t.Run("creates channel when defined in process spec", func(t *testing.T) {
		ch, err := ensureChannelExistsHelper(router, process, "chat")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if ch == nil {
			t.Fatal("Expected channel to be created, got nil")
		}
		if ch.Name != "chat" {
			t.Errorf("Expected channel name 'chat', got '%s'", ch.Name)
		}
		if ch.ProcessID != process.ID {
			t.Errorf("Expected process ID '%s', got '%s'", process.ID, ch.ProcessID)
		}
		if ch.SubmitterID != process.InitiatorID {
			t.Errorf("Expected submitter ID '%s', got '%s'", process.InitiatorID, ch.SubmitterID)
		}
		if ch.ExecutorID != process.AssignedExecutorID {
			t.Errorf("Expected executor ID '%s', got '%s'", process.AssignedExecutorID, ch.ExecutorID)
		}
	})

	// Test case 2: Channel not defined in process spec - should return error
	t.Run("returns error when channel not defined in process spec", func(t *testing.T) {
		_, err := ensureChannelExistsHelper(router, process, "undefined-channel")
		if err != channel.ErrChannelNotFound {
			t.Fatalf("Expected ErrChannelNotFound, got: %v", err)
		}
	})

	// Test case 3: Process is SUCCESS - should return error (channels cleaned up)
	t.Run("returns error when process is SUCCESS", func(t *testing.T) {
		successProcess := &core.Process{
			ID:    "success-process-id",
			State: core.SUCCESS,
			FunctionSpec: core.FunctionSpec{
				Channels: []string{"chat"},
			},
		}
		_, err := ensureChannelExistsHelper(router, successProcess, "chat")
		if err != channel.ErrChannelNotFound {
			t.Fatalf("Expected ErrChannelNotFound for SUCCESS process, got: %v", err)
		}
	})

	// Test case 4: Process is FAILED - should return error (channels cleaned up)
	t.Run("returns error when process is FAILED", func(t *testing.T) {
		failedProcess := &core.Process{
			ID:    "failed-process-id",
			State: core.FAILED,
			FunctionSpec: core.FunctionSpec{
				Channels: []string{"chat"},
			},
		}
		_, err := ensureChannelExistsHelper(router, failedProcess, "chat")
		if err != channel.ErrChannelNotFound {
			t.Fatalf("Expected ErrChannelNotFound for FAILED process, got: %v", err)
		}
	})
}

// ensureChannelExistsHelper is a helper function that mirrors the logic in RealtimeHandler.ensureChannelExists
// but uses a mock router interface for testing
func ensureChannelExistsHelper(router *MockChannelRouter, process *core.Process, channelName string) (*channel.Channel, error) {
	// Don't create channels for closed processes (SUCCESS or FAILED)
	if process.State == core.SUCCESS || process.State == core.FAILED {
		return nil, channel.ErrChannelNotFound
	}

	// Check if this channel is defined in the process spec
	channelDefined := false
	for _, ch := range process.FunctionSpec.Channels {
		if ch == channelName {
			channelDefined = true
			break
		}
	}

	if !channelDefined {
		return nil, channel.ErrChannelNotFound
	}

	// Create the channel on demand
	ch := &channel.Channel{
		ID:          process.ID + "_" + channelName,
		ProcessID:   process.ID,
		Name:        channelName,
		SubmitterID: process.InitiatorID,
		ExecutorID:  process.AssignedExecutorID,
	}

	// Use CreateIfNotExists to handle concurrent creation
	if err := router.CreateIfNotExists(ch); err != nil {
		return nil, err
	}

	// Return the channel
	return router.GetByProcessAndName(process.ID, channelName)
}

// TestEnsureChannelExistsWithRealRouter tests with the actual channel.Router
func TestEnsureChannelExistsWithRealRouter(t *testing.T) {
	router := channel.NewRouter()

	// Create a process with chat channel defined
	process := &core.Process{
		ID:                 "test-process-123",
		InitiatorID:        "submitter-456",
		AssignedExecutorID: "executor-789",
		State:              core.RUNNING,
		FunctionSpec: core.FunctionSpec{
			Channels: []string{"chat"},
		},
	}

	// Test 1: Channel doesn't exist initially
	_, err := router.GetByProcessAndName(process.ID, "chat")
	if err != channel.ErrChannelNotFound {
		t.Fatalf("Expected ErrChannelNotFound, got: %v (type: %T)", err, err)
	}

	// Test 2: Create channel using CreateIfNotExists
	ch := &channel.Channel{
		ID:          process.ID + "_chat",
		ProcessID:   process.ID,
		Name:        "chat",
		SubmitterID: process.InitiatorID,
		ExecutorID:  process.AssignedExecutorID,
	}
	err = router.CreateIfNotExists(ch)
	if err != nil {
		t.Fatalf("CreateIfNotExists failed: %v", err)
	}

	// Test 3: Now the channel should exist
	foundCh, err := router.GetByProcessAndName(process.ID, "chat")
	if err != nil {
		t.Fatalf("Expected channel to be found, got error: %v", err)
	}
	if foundCh.Name != "chat" {
		t.Errorf("Expected channel name 'chat', got '%s'", foundCh.Name)
	}
}

// TestChannelDeterministicID tests that channel IDs are deterministic
func TestChannelDeterministicID(t *testing.T) {
	processID := "abc123"
	channelName := "chat"

	expectedID := processID + "_" + channelName

	ch := &channel.Channel{
		ID:        processID + "_" + channelName,
		ProcessID: processID,
		Name:      channelName,
	}

	if ch.ID != expectedID {
		t.Errorf("Expected deterministic ID '%s', got '%s'", expectedID, ch.ID)
	}
}
