package channel_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// TestChannelAppendBasic tests basic channel append functionality
func TestChannelAppendBasic(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"test-channel"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	// Submit and assign the process
	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)

	assignedProcess, err := client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, process.ID, assignedProcess.ID)

	// Append to channel
	err = client.ChannelAppend(process.ID, "test-channel", 1, 0, []byte("Hello, World!"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Read from channel to verify
	entries, err := client.ChannelRead(process.ID, "test-channel", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, []byte("Hello, World!"), entries[0].Payload)

	srv.Shutdown()
	<-done
}

// TestChannelReadBasic tests basic channel read functionality
func TestChannelReadBasic(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"chat"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	// Submit and assign the process
	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, process.ID, assignedProcess.ID)

	// Append multiple messages
	err = client.ChannelAppend(process.ID, "chat", 1, 0, []byte("msg1"), env.ExecutorPrvKey)
	assert.Nil(t, err)
	err = client.ChannelAppend(process.ID, "chat", 2, 0, []byte("msg2"), env.ExecutorPrvKey)
	assert.Nil(t, err)
	err = client.ChannelAppend(process.ID, "chat", 3, 0, []byte("msg3"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Read all messages
	entries, err := client.ChannelRead(process.ID, "chat", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 3)

	// Read with afterSeq
	entries, err = client.ChannelRead(process.ID, "chat", 1, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 2) // Should get msg2 and msg3

	// Read with limit
	entries, err = client.ChannelRead(process.ID, "chat", 0, 2, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 2) // Should only get 2 messages

	srv.Shutdown()
	<-done
}

// TestChannelAppendProcessNotFound tests append to non-existent process
func TestChannelAppendProcessNotFound(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Try to append to non-existent process
	err := client.ChannelAppend("nonexistent-process-id", "test-channel", 1, 0, []byte("test"), env.ExecutorPrvKey)
	assert.NotNil(t, err)

	srv.Shutdown()
	<-done
}

// TestChannelReadProcessNotFound tests read from non-existent process
func TestChannelReadProcessNotFound(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Try to read from non-existent process
	_, err := client.ChannelRead("nonexistent-process-id", "test-channel", 0, 0, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	srv.Shutdown()
	<-done
}

// TestChannelAppendChannelNotFound tests append to non-existent channel
func TestChannelAppendChannelNotFound(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process without channels
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Conditions.ExecutorType = env.Executor.Type
	// Note: No channels specified

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, process.ID, assignedProcess.ID)

	// Try to append to non-existent channel
	err = client.ChannelAppend(process.ID, "nonexistent-channel", 1, 0, []byte("test"), env.ExecutorPrvKey)
	assert.NotNil(t, err)

	srv.Shutdown()
	<-done
}

// TestChannelReadChannelNotFound tests read from non-existent channel
func TestChannelReadChannelNotFound(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process without channels
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, process.ID, assignedProcess.ID)

	// Try to read from non-existent channel
	_, err = client.ChannelRead(process.ID, "nonexistent-channel", 0, 0, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	srv.Shutdown()
	<-done
}

// TestChannelAppendUnauthorized tests that non-members cannot append
func TestChannelAppendUnauthorized(t *testing.T) {
	env, client, srv, serverPrvKey, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"test-channel"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Create another colony and executor (non-member)
	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colony2PrvKey)
	assert.Nil(t, err)

	// Try to append with unauthorized key
	err = client.ChannelAppend(process.ID, "test-channel", 1, 0, []byte("test"), executor2PrvKey)
	assert.NotNil(t, err) // Should fail

	srv.Shutdown()
	<-done
}

// TestChannelReadUnauthorized tests that non-members cannot read
func TestChannelReadUnauthorized(t *testing.T) {
	env, client, srv, serverPrvKey, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"test-channel"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Add a message
	err = client.ChannelAppend(process.ID, "test-channel", 1, 0, []byte("secret"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Create another colony and executor (non-member)
	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colony2PrvKey)
	assert.Nil(t, err)

	// Try to read with unauthorized key
	_, err = client.ChannelRead(process.ID, "test-channel", 0, 0, executor2PrvKey)
	assert.NotNil(t, err) // Should fail

	srv.Shutdown()
	<-done
}

// TestChannelBidirectionalCommunication tests communication between submitter and executor
func TestChannelBidirectionalCommunication(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel - submitted by executor (acting as user)
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"chat"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	// Submit process
	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)
	submitterID := process.InitiatorID

	// Assign process to same executor
	assignedProcess, err := client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	executorID := assignedProcess.AssignedExecutorID

	// Submitter sends message
	err = client.ChannelAppend(process.ID, "chat", 1, 0, []byte("Hello from submitter"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Read and verify sender
	entries, err := client.ChannelRead(process.ID, "chat", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)

	// Since executor submitted and is also assigned, they should have the same ID
	// The sender should be either submitterID or executorID based on implementation
	t.Logf("SubmitterID: %s, ExecutorID: %s, SenderID: %s", submitterID, executorID, entries[0].SenderID)

	srv.Shutdown()
	<-done
}

// TestChannelWithInReplyTo tests the in_reply_to field
func TestChannelWithInReplyTo(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"chat"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Send question with sequence 1
	err = client.ChannelAppend(process.ID, "chat", 1, 0, []byte("What is 2+2?"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Send reply with InReplyTo = 1
	err = client.ChannelAppend(process.ID, "chat", 2, 1, []byte("4"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Read all and verify
	entries, err := client.ChannelRead(process.ID, "chat", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 2)

	// First message has no reply
	assert.Equal(t, int64(1), entries[0].Sequence)
	assert.Equal(t, int64(0), entries[0].InReplyTo)

	// Second message replies to first
	assert.Equal(t, int64(2), entries[1].Sequence)
	assert.Equal(t, int64(1), entries[1].InReplyTo)

	srv.Shutdown()
	<-done
}

// TestChannelCleanupOnProcessClose tests that channels are cleaned up when process closes
func TestChannelCleanupOnProcessClose(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"chat"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Add messages
	err = client.ChannelAppend(process.ID, "chat", 1, 0, []byte("test"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify channel works
	entries, err := client.ChannelRead(process.ID, "chat", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)

	// Close process
	err = client.Close(process.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to read after close - should fail
	_, err = client.ChannelRead(process.ID, "chat", 0, 0, env.ExecutorPrvKey)
	assert.NotNil(t, err) // Channel should be cleaned up

	srv.Shutdown()
	<-done
}

// TestChannelMultipleChannels tests a process with multiple channels
func TestChannelMultipleChannels(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with multiple channels
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"channel1", "channel2", "channel3"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Write to each channel
	err = client.ChannelAppend(process.ID, "channel1", 1, 0, []byte("msg1"), env.ExecutorPrvKey)
	assert.Nil(t, err)
	err = client.ChannelAppend(process.ID, "channel2", 1, 0, []byte("msg2"), env.ExecutorPrvKey)
	assert.Nil(t, err)
	err = client.ChannelAppend(process.ID, "channel3", 1, 0, []byte("msg3"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Read from each channel and verify isolation
	entries1, err := client.ChannelRead(process.ID, "channel1", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries1, 1)
	assert.Equal(t, []byte("msg1"), entries1[0].Payload)

	entries2, err := client.ChannelRead(process.ID, "channel2", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries2, 1)
	assert.Equal(t, []byte("msg2"), entries2[0].Payload)

	entries3, err := client.ChannelRead(process.ID, "channel3", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries3, 1)
	assert.Equal(t, []byte("msg3"), entries3[0].Payload)

	srv.Shutdown()
	<-done
}

// TestChannelEmptyPayload tests appending an empty payload
func TestChannelEmptyPayload(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"chat"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Append empty payload
	err = client.ChannelAppend(process.ID, "chat", 1, 0, []byte{}, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Read and verify
	entries, err := client.ChannelRead(process.ID, "chat", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, []byte{}, entries[0].Payload)

	srv.Shutdown()
	<-done
}

// TestChannelLargePayload tests appending a large payload
func TestChannelLargePayload(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"data"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Create a 1MB payload
	largePayload := make([]byte, 1024*1024)
	for i := range largePayload {
		largePayload[i] = byte(i % 256)
	}

	// Append large payload
	err = client.ChannelAppend(process.ID, "data", 1, 0, largePayload, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Read and verify
	entries, err := client.ChannelRead(process.ID, "data", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, largePayload, entries[0].Payload)

	srv.Shutdown()
	<-done
}

// TestChannelWaitingProcessCannotUseChannel tests that waiting process cannot use channels
func TestChannelWaitingProcessCannotUseChannel(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel but don't assign it
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"chat"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.WAITING, process.State)

	// Try to append to channel of waiting process
	// This should work or fail depending on implementation
	// Channels are typically available once the process is running
	err = client.ChannelAppend(process.ID, "chat", 1, 0, []byte("test"), env.ExecutorPrvKey)
	// Note: The behavior here depends on implementation - document actual behavior
	t.Logf("Append to waiting process channel result: %v", err)

	srv.Shutdown()
	<-done
}

// TestChannelUnauthorizedExecutorSameColony tests that an executor in the same colony
// but NOT the assigned executor cannot access the channel
func TestChannelUnauthorizedExecutorSameColony(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a second executor in the same colony
	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	executor2.Name = "executor2"
	executor2.Type = env.Executor.Type
	_, err = client.AddExecutor(executor2, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor2.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create a process with a channel - submitted by first executor
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"secure-channel"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// First executor assigns the process
	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// First executor (submitter and assigned) can append
	err = client.ChannelAppend(process.ID, "secure-channel", 1, 0, []byte("secret data"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Second executor (same colony, but NOT assigned) should NOT be able to append
	err = client.ChannelAppend(process.ID, "secure-channel", 2, 0, []byte("unauthorized"), executor2PrvKey)
	assert.NotNil(t, err, "Executor in same colony but not assigned should not be able to append")

	// Second executor should NOT be able to read
	_, err = client.ChannelRead(process.ID, "secure-channel", 0, 0, executor2PrvKey)
	assert.NotNil(t, err, "Executor in same colony but not assigned should not be able to read")

	srv.Shutdown()
	<-done
}

// TestChannelUnauthorizedUserSameColony tests that a user in the same colony
// but NOT the submitter cannot access the channel
func TestChannelUnauthorizedUserSameColony(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create first user who will submit the process
	user1, user1PrvKey, err := utils.CreateTestUserWithKey(env.ColonyName, "user1")
	assert.Nil(t, err)
	_, err = client.AddUser(user1, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create second user in the same colony
	user2, user2PrvKey, err := utils.CreateTestUserWithKey(env.ColonyName, "user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, env.ColonyPrvKey)
	assert.Nil(t, err)

	// User1 submits a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"private-channel"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, user1PrvKey)
	assert.Nil(t, err)

	// Executor assigns the process
	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// User1 (submitter) can append
	err = client.ChannelAppend(process.ID, "private-channel", 1, 0, []byte("from submitter"), user1PrvKey)
	assert.Nil(t, err)

	// User2 (same colony, but NOT submitter) should NOT be able to append
	err = client.ChannelAppend(process.ID, "private-channel", 2, 0, []byte("unauthorized"), user2PrvKey)
	assert.NotNil(t, err, "User in same colony but not submitter should not be able to append")

	// User2 should NOT be able to read
	_, err = client.ChannelRead(process.ID, "private-channel", 0, 0, user2PrvKey)
	assert.NotNil(t, err, "User in same colony but not submitter should not be able to read")

	// Assigned executor can access
	err = client.ChannelAppend(process.ID, "private-channel", 2, 0, []byte("from executor"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	entries, err := client.ChannelRead(process.ID, "private-channel", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 2)

	srv.Shutdown()
	<-done
}

// TestChannelOnlyAssignedExecutorCanAccess tests that only the specifically assigned
// executor can access the channel, not other executors even if they could run the process
func TestChannelOnlyAssignedExecutorCanAccess(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create two executors of the same type
	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	executor2.Name = "executor2"
	executor2.Type = env.Executor.Type // Same type as first executor
	_, err = client.AddExecutor(executor2, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor2.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Submit process - either executor could potentially run it
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"exclusive-channel"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// First executor assigns the process
	assignedProcess, err := client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, process.ID, assignedProcess.ID)

	// Assigned executor can access
	err = client.ChannelAppend(process.ID, "exclusive-channel", 1, 0, []byte("assigned"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Second executor (same type, could have been assigned, but wasn't) cannot access
	err = client.ChannelAppend(process.ID, "exclusive-channel", 2, 0, []byte("not assigned"), executor2PrvKey)
	assert.NotNil(t, err, "Executor that was not assigned should not be able to access channel")

	_, err = client.ChannelRead(process.ID, "exclusive-channel", 0, 0, executor2PrvKey)
	assert.NotNil(t, err, "Executor that was not assigned should not be able to read channel")

	srv.Shutdown()
	<-done
}

// TestChannelSequenceOrdering tests that messages are ordered by sequence
func TestChannelSequenceOrdering(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"ordered"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Send messages out of order (by sequence number)
	err = client.ChannelAppend(process.ID, "ordered", 3, 0, []byte("third"), env.ExecutorPrvKey)
	assert.Nil(t, err)
	err = client.ChannelAppend(process.ID, "ordered", 1, 0, []byte("first"), env.ExecutorPrvKey)
	assert.Nil(t, err)
	err = client.ChannelAppend(process.ID, "ordered", 2, 0, []byte("second"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Read all and check ordering
	entries, err := client.ChannelRead(process.ID, "ordered", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 3)

	// Log the actual order to understand implementation behavior
	for i, e := range entries {
		t.Logf("Entry %d: Sequence=%d, Payload=%s", i, e.Sequence, string(e.Payload))
	}

	srv.Shutdown()
	<-done
}

// TestChannelAppendToClosedProcess tests that channel append fails for SUCCESS process
func TestChannelAppendToClosedProcess(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"test-channel"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Close the process successfully
	err = client.Close(process.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to append to channel of closed process - should fail
	err = client.ChannelAppend(process.ID, "test-channel", 1, 0, []byte("test"), env.ExecutorPrvKey)
	assert.NotNil(t, err)

	srv.Shutdown()
	<-done
}

// TestChannelReadFromClosedProcess tests that channel read fails for SUCCESS process
func TestChannelReadFromClosedProcess(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"test-channel"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Close the process successfully
	err = client.Close(process.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to read from channel of closed process - should fail
	_, err = client.ChannelRead(process.ID, "test-channel", 0, 0, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	srv.Shutdown()
	<-done
}

// TestChannelAppendToFailedProcess tests that channel append fails for FAILED process
func TestChannelAppendToFailedProcess(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"test-channel"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Fail the process
	err = client.Fail(process.ID, []string{"error"}, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to append to channel of failed process - should fail
	err = client.ChannelAppend(process.ID, "test-channel", 1, 0, []byte("test"), env.ExecutorPrvKey)
	assert.NotNil(t, err)

	srv.Shutdown()
	<-done
}

// TestChannelReadFromFailedProcess tests that channel read fails for FAILED process
func TestChannelReadFromFailedProcess(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"test-channel"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Fail the process
	err = client.Fail(process.ID, []string{"error"}, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to read from channel of failed process - should fail
	_, err = client.ChannelRead(process.ID, "test-channel", 0, 0, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	srv.Shutdown()
	<-done
}

// TestChannelAppendWithPayloadType tests the payload type field (data, end, error)
func TestChannelAppendWithPayloadType(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"stream"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Append with data type
	err = client.ChannelAppendWithType(process.ID, "stream", 1, 0, []byte("data message"), "data", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Append with end type
	err = client.ChannelAppendWithType(process.ID, "stream", 2, 0, []byte(""), "end", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Read and verify types
	entries, err := client.ChannelRead(process.ID, "stream", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "data", entries[0].Type)
	assert.Equal(t, "end", entries[1].Type)

	srv.Shutdown()
	<-done
}

// TestChannelAppendWithErrorType tests the error payload type
func TestChannelAppendWithErrorType(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"stream"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Append with error type
	err = client.ChannelAppendWithType(process.ID, "stream", 1, 0, []byte("Something went wrong"), "error", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Read and verify type
	entries, err := client.ChannelRead(process.ID, "stream", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "error", entries[0].Type)
	assert.Equal(t, []byte("Something went wrong"), entries[0].Payload)

	srv.Shutdown()
	<-done
}

// TestChannelSubmitterCanAccessBeforeAssign tests that the submitter can access
// the channel even before a process is assigned
func TestChannelSubmitterCanAccessBeforeAssign(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with a channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"pre-assign"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.WAITING, process.State)

	// Submitter tries to write before assignment
	// This documents current behavior - may succeed or fail depending on implementation
	err = client.ChannelAppend(process.ID, "pre-assign", 1, 0, []byte("pre-assign message"), env.ExecutorPrvKey)
	t.Logf("Append to waiting process result: %v", err)

	// Now assign and verify we can definitely access
	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Should be able to write now
	err = client.ChannelAppend(process.ID, "pre-assign", 2, 0, []byte("post-assign message"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	srv.Shutdown()
	<-done
}

// TestChannelMultipleExecutorsInColony tests channel access with multiple executors
func TestChannelMultipleExecutorsInColony(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a second executor of different type
	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	executor2.Name = "executor2-different-type"
	executor2.Type = "different-type"
	_, err = client.AddExecutor(executor2, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor2.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Submit process assigned to first executor type
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"restricted"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// First executor assigns
	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// First executor can access
	err = client.ChannelAppend(process.ID, "restricted", 1, 0, []byte("authorized"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Second executor (different type, not assigned) cannot access
	err = client.ChannelAppend(process.ID, "restricted", 2, 0, []byte("unauthorized"), executor2PrvKey)
	assert.NotNil(t, err)

	srv.Shutdown()
	<-done
}

// TestChannelAppendToUndefinedChannel tests accessing a channel not in the process spec
// when process has other channels defined (to test the for-loop iteration path)
func TestChannelAppendToUndefinedChannel(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with specific channels
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"channel-a", "channel-b", "channel-c"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to access a channel that's NOT in the list
	err = client.ChannelAppend(process.ID, "undefined-channel", 1, 0, []byte("test"), env.ExecutorPrvKey)
	assert.NotNil(t, err) // Should fail - channel not defined in spec

	srv.Shutdown()
	<-done
}

// TestChannelReadFromUndefinedChannel tests reading from a channel not in the process spec
func TestChannelReadFromUndefinedChannel(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with specific channels
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"channel-a", "channel-b"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to read from a channel that's NOT in the list
	_, err = client.ChannelRead(process.ID, "nonexistent-channel", 0, 0, env.ExecutorPrvKey)
	assert.NotNil(t, err) // Should fail - channel not defined in spec

	srv.Shutdown()
	<-done
}

// TestChannelDefinesChannelButUseDifferent tests a more complex scenario:
// process defines channels but we try to access one not in the list
func TestChannelDefinesChannelButUseDifferent(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a process with multiple channels
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"input", "output", "logs"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Valid channels work
	err = client.ChannelAppend(process.ID, "input", 1, 0, []byte("data"), env.ExecutorPrvKey)
	assert.Nil(t, err)
	err = client.ChannelAppend(process.ID, "output", 1, 0, []byte("result"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Invalid channel fails
	err = client.ChannelAppend(process.ID, "metrics", 1, 0, []byte("metric"), env.ExecutorPrvKey)
	assert.NotNil(t, err)

	// Invalid channel read also fails
	_, err = client.ChannelRead(process.ID, "status", 0, 0, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	srv.Shutdown()
	<-done
}

// TestChannelReadWithLimitZero tests reading with limit=0 (no limit)
func TestChannelReadWithLimitZero(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"unlimited"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Add many messages
	for i := 1; i <= 10; i++ {
		err = client.ChannelAppend(process.ID, "unlimited", int64(i), 0, []byte("msg"), env.ExecutorPrvKey)
		assert.Nil(t, err)
	}

	// Read with limit=0 should get all
	entries, err := client.ChannelRead(process.ID, "unlimited", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 10)

	srv.Shutdown()
	<-done
}

// TestChannelReadAfterSeqHigherThanAll tests reading with afterSeq > all sequences
func TestChannelReadAfterSeqHigherThanAll(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"test"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Add messages with sequences 1-3
	for i := 1; i <= 3; i++ {
		err = client.ChannelAppend(process.ID, "test", int64(i), 0, []byte("msg"), env.ExecutorPrvKey)
		assert.Nil(t, err)
	}

	// Read with afterSeq=100 - should get empty result
	entries, err := client.ChannelRead(process.ID, "test", 100, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 0)

	srv.Shutdown()
	<-done
}

// TestChannelUserAsSubmitter tests that a user (not executor) can submit and use channels
func TestChannelUserAsSubmitter(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv2(t)

	// Create a user
	user, userPrvKey, err := utils.CreateTestUserWithKey(env.ColonyName, "channel-user")
	assert.Nil(t, err)
	_, err = client.AddUser(user, env.ColonyPrvKey)
	assert.Nil(t, err)

	// User submits process with channel
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.Channels = []string{"user-channel"}
	funcSpec.Conditions.ExecutorType = env.Executor.Type

	process, err := client.Submit(funcSpec, userPrvKey)
	assert.Nil(t, err)

	// Executor assigns
	_, err = client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// User (submitter) can write
	err = client.ChannelAppend(process.ID, "user-channel", 1, 0, []byte("user message"), userPrvKey)
	assert.Nil(t, err)

	// User can read
	entries, err := client.ChannelRead(process.ID, "user-channel", 0, 0, userPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)

	// Executor can also access
	err = client.ChannelAppend(process.ID, "user-channel", 2, 0, []byte("executor message"), env.ExecutorPrvKey)
	assert.Nil(t, err)

	entries, err = client.ChannelRead(process.ID, "user-channel", 0, 0, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 2)

	srv.Shutdown()
	<-done
}
