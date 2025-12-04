package channel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMultiServerReplication(t *testing.T) {
	// Create 3 servers
	server1 := NewRouter()
	server2 := NewRouter()
	server3 := NewRouter()

	// Set up replication: server1 replicates to server2 and server3
	replicator := NewInMemoryReplicator([]*Router{server2, server3})
	server1.SetReplicator(replicator)

	// Create channel on server1
	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "input",
		SubmitterID: "user-789",
		ExecutorID:  "exec-abc",
	}
	err := server1.Create(channel)
	assert.Nil(t, err)

	// Wait for async replication
	time.Sleep(50 * time.Millisecond)

	// Channel should exist on all servers
	_, err = server1.Get("ch-123")
	assert.Nil(t, err)
	_, err = server2.Get("ch-123")
	assert.Nil(t, err)
	_, err = server3.Get("ch-123")
	assert.Nil(t, err)
}

func TestMultiServerMessageReplication(t *testing.T) {
	// Create 3 servers
	server1 := NewRouter()
	server2 := NewRouter()
	server3 := NewRouter()

	// Set up replication
	replicator := NewInMemoryReplicator([]*Router{server2, server3})
	server1.SetReplicator(replicator)

	// Create channel on server1
	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "input",
		SubmitterID: "user-789",
		ExecutorID:  "exec-abc",
	}
	err := server1.Create(channel)

	// Wait for channel replication
	time.Sleep(50 * time.Millisecond)

	// Append message on server1
	err = server1.Append("ch-123", "user-789", 1, 0, []byte("hello"))
	assert.Nil(t, err)

	// Wait for message replication
	time.Sleep(50 * time.Millisecond)

	// Message should be readable from all servers
	entries1, _ := server1.ReadAfter("ch-123", "user-789", 0, 0)
	entries2, _ := server2.ReadAfter("ch-123", "user-789", 0, 0)
	entries3, _ := server3.ReadAfter("ch-123", "user-789", 0, 0)

	assert.Len(t, entries1, 1)
	assert.Len(t, entries2, 1)
	assert.Len(t, entries3, 1)

	assert.Equal(t, []byte("hello"), entries1[0].Payload)
	assert.Equal(t, []byte("hello"), entries2[0].Payload)
	assert.Equal(t, []byte("hello"), entries3[0].Payload)
}

func TestMultiServerWriteFromDifferentServers(t *testing.T) {
	// Simulate: client writes to server1, executor reads from server2
	server1 := NewRouter()
	server2 := NewRouter()

	// Bidirectional replication
	rep1to2 := NewInMemoryReplicator([]*Router{server2})
	rep2to1 := NewInMemoryReplicator([]*Router{server1})
	server1.SetReplicator(rep1to2)
	server2.SetReplicator(rep2to1)

	// Create channel on server1 (simulating process submission)
	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "input",
		SubmitterID: "user-789",
		ExecutorID:  "exec-abc",
	}
	server1.Create(channel)
	time.Sleep(50 * time.Millisecond)

	// Client writes to server1
	server1.Append("ch-123", "user-789", 1, 0, []byte("message 1"))
	time.Sleep(50 * time.Millisecond)

	// Executor reads from server2
	entries, err := server2.ReadAfter("ch-123", "exec-abc", 0, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, []byte("message 1"), entries[0].Payload)

	// Executor writes response on server2
	server2.Append("ch-123", "exec-abc", 1, 1, []byte("response 1"))
	time.Sleep(50 * time.Millisecond)

	// Client reads from server1
	entries, err = server1.ReadAfter("ch-123", "user-789", 0, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, []byte("message 1"), entries[0].Payload)
	assert.Equal(t, []byte("response 1"), entries[1].Payload)
}

func TestMultiServerExecutorAssignment(t *testing.T) {
	server1 := NewRouter()
	server2 := NewRouter()

	replicator := NewInMemoryReplicator([]*Router{server2})
	server1.SetReplicator(replicator)

	// Create channel without executor
	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "input",
		SubmitterID: "user-789",
		ExecutorID:  "", // Not assigned yet
	}
	server1.Create(channel)
	time.Sleep(50 * time.Millisecond)

	// Executor cannot access on either server
	var err error
	err = server1.Append("ch-123", "exec-abc", 1, 0, []byte("denied"))
	assert.Equal(t, ErrUnauthorized, err)
	err = server2.Append("ch-123", "exec-abc", 1, 0, []byte("denied"))
	assert.Equal(t, ErrUnauthorized, err)

	// Assign executor on server1
	server1.SetExecutorIDForProcess("proc-456", "exec-abc")
	time.Sleep(50 * time.Millisecond)

	// Now executor can access on both servers
	err = server1.Append("ch-123", "exec-abc", 1, 0, []byte("allowed"))
	assert.Nil(t, err)
	err = server2.Append("ch-123", "exec-abc", 2, 0, []byte("allowed"))
	assert.Nil(t, err)
}

func TestMultiServerCleanup(t *testing.T) {
	server1 := NewRouter()
	server2 := NewRouter()
	server3 := NewRouter()

	replicator := NewInMemoryReplicator([]*Router{server2, server3})
	server1.SetReplicator(replicator)

	// Create channels
	channel1 := &Channel{
		ID:          "ch-1",
		ProcessID:   "proc-1",
		Name:        "input",
		SubmitterID: "user-1",
	}
	channel2 := &Channel{
		ID:          "ch-2",
		ProcessID:   "proc-1",
		Name:        "output",
		SubmitterID: "user-1",
	}
	server1.Create(channel1)
	server1.Create(channel2)
	time.Sleep(50 * time.Millisecond)

	// Verify channels exist on all servers
	_, err := server2.Get("ch-1")
	assert.Nil(t, err)
	_, err = server3.Get("ch-2")
	assert.Nil(t, err)

	// Cleanup on server1
	server1.CleanupProcess("proc-1")
	time.Sleep(50 * time.Millisecond)

	// Channels should be gone from all servers
	_, err = server1.Get("ch-1")
	assert.Equal(t, ErrChannelNotFound, err)
	_, err = server2.Get("ch-1")
	assert.Equal(t, ErrChannelNotFound, err)
	_, err = server3.Get("ch-1")
	assert.Equal(t, ErrChannelNotFound, err)
}

func TestMultiServerSequenceConsistency(t *testing.T) {
	// Test that sequence numbers from leader are consistent across replicas
	server1 := NewRouter() // Leader
	server2 := NewRouter()

	replicator := NewInMemoryReplicator([]*Router{server2})
	server1.SetReplicator(replicator)

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "input",
		SubmitterID: "user-789",
		ExecutorID:  "exec-abc",
	}
	server1.Create(channel)
	time.Sleep(50 * time.Millisecond)

	// Multiple appends on server1 (leader)
	server1.Append("ch-123", "user-789", 1, 0, []byte("msg1"))
	server1.Append("ch-123", "user-789", 2, 0, []byte("msg2"))
	server1.Append("ch-123", "user-789", 3, 0, []byte("msg3"))
	time.Sleep(50 * time.Millisecond)

	// Read from replica - should have same sequence numbers
	entries, _ := server2.ReadAfter("ch-123", "user-789", 0, 0)
	assert.Len(t, entries, 3)
	assert.Equal(t, int64(1), entries[0].Sequence)
	assert.Equal(t, int64(2), entries[1].Sequence)
	assert.Equal(t, int64(3), entries[2].Sequence)
}

func TestChatScenario(t *testing.T) {
	// Simulate complete chat scenario with multiple servers
	server1 := NewRouter() // Client connects here
	server2 := NewRouter() // Executor connects here

	// Bidirectional replication
	rep1to2 := NewInMemoryReplicator([]*Router{server2})
	rep2to1 := NewInMemoryReplicator([]*Router{server1})
	server1.SetReplicator(rep1to2)
	server2.SetReplicator(rep2to1)

	// 1. Process submitted - channels created on server1
	inputChannel := &Channel{
		ID:          "input-ch",
		ProcessID:   "proc-1",
		Name:        "input",
		SubmitterID: "user-1",
	}
	outputChannel := &Channel{
		ID:          "output-ch",
		ProcessID:   "proc-1",
		Name:        "output",
		SubmitterID: "user-1",
	}
	server1.Create(inputChannel)
	server1.Create(outputChannel)
	time.Sleep(50 * time.Millisecond)

	// 2. Process assigned to executor
	server1.SetExecutorIDForProcess("proc-1", "exec-1")
	time.Sleep(50 * time.Millisecond)

	// 3. User sends message via server1
	server1.Append("input-ch", "user-1", 1, 0, []byte(`{"text": "Hello AI"}`))
	time.Sleep(50 * time.Millisecond)

	// 4. Executor reads from server2
	entries, _ := server2.ReadAfter("input-ch", "exec-1", 0, 0)
	assert.Len(t, entries, 1)

	// 5. Executor sends response via server2
	server2.Append("output-ch", "exec-1", 1, 1, []byte(`{"text": "Hello human!"}`))
	time.Sleep(50 * time.Millisecond)

	// 6. User reads response from server1
	entries, _ = server1.ReadAfter("output-ch", "user-1", 0, 0)
	assert.Len(t, entries, 1)
	assert.Equal(t, []byte(`{"text": "Hello human!"}`), entries[0].Payload)

	// 7. Process completes - cleanup
	server1.CleanupProcess("proc-1")
	time.Sleep(50 * time.Millisecond)

	// Both servers should be clean
	_, err := server1.Get("input-ch")
	assert.Equal(t, ErrChannelNotFound, err)
	_, err = server2.Get("output-ch")
	assert.Equal(t, ErrChannelNotFound, err)
}
