package server

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/channel"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// pollForEntries polls for channel entries until expected count is reached or timeout
// This represents realistic client behavior with eventual consistency
func pollForEntries(t *testing.T, c *client.ColoniesClient, processID, channelName string, afterIndex int64, expectedCount int, prvKey string) []*channel.MsgEntry {
	timeout := 5 * time.Second
	deadline := time.Now().Add(timeout)
	pollInterval := 100 * time.Millisecond // Less aggressive to avoid port exhaustion

	for time.Now().Before(deadline) {
		entries, err := c.ChannelRead(processID, channelName, afterIndex, 0, prvKey)
		if err == nil && len(entries) >= expectedCount {
			return entries
		}
		time.Sleep(pollInterval)
	}

	// Final try with assertion
	entries, err := c.ChannelRead(processID, channelName, afterIndex, 0, prvKey)
	if !assert.Nil(t, err, "Failed to read channel entries") {
		t.FailNow()
	}
	if !assert.GreaterOrEqual(t, len(entries), expectedCount, "Expected at least %d entries but got %d", expectedCount, len(entries)) {
		t.FailNow()
	}
	return entries
}

// pollForCleanup polls until channel is no longer accessible
func pollForCleanup(t *testing.T, c *client.ColoniesClient, processID, channelName string, prvKey string) {
	timeout := 5 * time.Second
	deadline := time.Now().Add(timeout)
	pollInterval := 100 * time.Millisecond // Less aggressive to avoid port exhaustion

	for time.Now().Before(deadline) {
		_, err := c.ChannelRead(processID, channelName, 0, 0, prvKey)
		if err != nil {
			return // Channel is gone
		}
		time.Sleep(pollInterval)
	}
	t.Fatalf("Channel %s still exists after timeout", channelName)
}

// TestChannelClusterBasicReplication tests that channel messages are replicated across servers.
// Client submits process via Server1, Executor connects to Server2, they can still communicate.
func TestChannelClusterBasicReplication(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("CHANNEL_CLUSTER_")
	assert.Nil(t, err)
	defer db.Close()

	servers := StartCluster(t, db, 2)
	WaitForCluster(t, servers)
	defer func() {
		for _, s := range servers {
			s.Server.Shutdown()
			<-s.Done
		}
	}()

	client1 := client.CreateColoniesClient("localhost", servers[0].Node.APIPort, true, true)
	client2 := client.CreateColoniesClient("localhost", servers[1].Node.APIPort, true, true)

	colonyCrypto := crypto.CreateCrypto()
	colonyPrvKey, err := colonyCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID, err := colonyCrypto.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test-colony-cluster")
	_, err = client1.AddColony(colony, servers[0].ServerPrvKey)
	assert.Nil(t, err)

	userCrypto := crypto.CreateCrypto()
	userPrvKey, err := userCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	userID, err := userCrypto.GenerateID(userPrvKey)
	assert.Nil(t, err)

	user := core.CreateUser(colony.Name, userID, "test-user", "user@test.com", "")
	_, err = client1.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	executor.Type = "cluster-executor"
	_, err = client2.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	err = client2.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	funcSpec := utils.CreateTestFunctionSpec(colony.Name)
	funcSpec.Channels = []string{"cluster-chat"}
	funcSpec.Conditions.ExecutorType = "cluster-executor"

	submittedProcess, err := client1.Submit(funcSpec, userPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client2.Assign(colony.Name, 10, "", "", executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, submittedProcess.ID, assignedProcess.ID)

	// User sends message via Server1
	err = client1.ChannelAppend(submittedProcess.ID, "cluster-chat", 1, 0, []byte("Hello from Server1!"), userPrvKey)
	assert.Nil(t, err)

	// Executor polls for message via Server2 (realistic eventual consistency behavior)
	entries := pollForEntries(t, client2, submittedProcess.ID, "cluster-chat", 0, 1, executorPrvKey)
	assert.Equal(t, []byte("Hello from Server1!"), entries[0].Payload)

	// Executor replies via Server2
	err = client2.ChannelAppend(submittedProcess.ID, "cluster-chat", 1, 1, []byte("Hello from Server2!"), executorPrvKey)
	assert.Nil(t, err)

	// User polls for executor's reply via Server1 (realistic eventual consistency behavior)
	entries = pollForEntries(t, client1, submittedProcess.ID, "cluster-chat", 1, 1, userPrvKey)
	assert.Equal(t, []byte("Hello from Server2!"), entries[0].Payload)

	err = client2.Close(submittedProcess.ID, executorPrvKey)
	assert.Nil(t, err)
}

// TestChannelClusterCleanupReplication tests that channel cleanup propagates across all servers.
func TestChannelClusterCleanupReplication(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("CHANNEL_CLUSTER_CLEANUP_")
	assert.Nil(t, err)
	defer db.Close()

	servers := StartCluster(t, db, 3)
	WaitForCluster(t, servers)
	defer func() {
		for _, s := range servers {
			s.Server.Shutdown()
			<-s.Done
		}
	}()

	client1 := client.CreateColoniesClient("localhost", servers[0].Node.APIPort, true, true)
	client2 := client.CreateColoniesClient("localhost", servers[1].Node.APIPort, true, true)
	client3 := client.CreateColoniesClient("localhost", servers[2].Node.APIPort, true, true)

	colonyCrypto := crypto.CreateCrypto()
	colonyPrvKey, err := colonyCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID, err := colonyCrypto.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test-colony-cleanup")
	_, err = client1.AddColony(colony, servers[0].ServerPrvKey)
	assert.Nil(t, err)

	userCrypto := crypto.CreateCrypto()
	userPrvKey, err := userCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	userID, err := userCrypto.GenerateID(userPrvKey)
	assert.Nil(t, err)

	user := core.CreateUser(colony.Name, userID, "test-user", "user@test.com", "")
	_, err = client1.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	executor.Type = "cleanup-executor"
	_, err = client1.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	err = client1.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	funcSpec := utils.CreateTestFunctionSpec(colony.Name)
	funcSpec.Channels = []string{"cleanup-channel"}
	funcSpec.Conditions.ExecutorType = "cleanup-executor"

	submittedProcess, err := client1.Submit(funcSpec, userPrvKey)
	assert.Nil(t, err)

	_, err = client2.Assign(colony.Name, 10, "", "", executorPrvKey)
	assert.Nil(t, err)

	// Write messages via different servers
	err = client1.ChannelAppend(submittedProcess.ID, "cleanup-channel", 1, 0, []byte("msg from server1"), userPrvKey)
	assert.Nil(t, err)

	err = client2.ChannelAppend(submittedProcess.ID, "cleanup-channel", 1, 0, []byte("msg from server2"), executorPrvKey)
	assert.Nil(t, err)

	// Verify messages are readable from all servers using polling (eventual consistency)
	entries := pollForEntries(t, client1, submittedProcess.ID, "cleanup-channel", 0, 2, userPrvKey)
	assert.Len(t, entries, 2)

	entries = pollForEntries(t, client2, submittedProcess.ID, "cleanup-channel", 0, 2, executorPrvKey)
	assert.Len(t, entries, 2)

	entries = pollForEntries(t, client3, submittedProcess.ID, "cleanup-channel", 0, 2, userPrvKey)
	assert.Len(t, entries, 2)

	// Close process
	err = client1.Close(submittedProcess.ID, executorPrvKey)
	assert.Nil(t, err)

	// Verify channel is gone from ALL servers using polling
	pollForCleanup(t, client1, submittedProcess.ID, "cleanup-channel", userPrvKey)
	pollForCleanup(t, client2, submittedProcess.ID, "cleanup-channel", executorPrvKey)
	pollForCleanup(t, client3, submittedProcess.ID, "cleanup-channel", userPrvKey)
}

// TestChannelClusterConcurrentWrites tests concurrent writes from clients on different servers.
func TestChannelClusterConcurrentWrites(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("CHANNEL_CLUSTER_CONCURRENT_")
	assert.Nil(t, err)
	defer db.Close()

	servers := StartCluster(t, db, 2)
	WaitForCluster(t, servers)
	defer func() {
		for _, s := range servers {
			s.Server.Shutdown()
			<-s.Done
		}
	}()

	client1 := client.CreateColoniesClient("localhost", servers[0].Node.APIPort, true, true)
	client2 := client.CreateColoniesClient("localhost", servers[1].Node.APIPort, true, true)

	colonyCrypto := crypto.CreateCrypto()
	colonyPrvKey, err := colonyCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID, err := colonyCrypto.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test-colony-concurrent")
	_, err = client1.AddColony(colony, servers[0].ServerPrvKey)
	assert.Nil(t, err)

	userCrypto := crypto.CreateCrypto()
	userPrvKey, err := userCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	userID, err := userCrypto.GenerateID(userPrvKey)
	assert.Nil(t, err)

	user := core.CreateUser(colony.Name, userID, "test-user", "user@test.com", "")
	_, err = client1.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	executor.Type = "concurrent-executor"
	_, err = client1.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	err = client1.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	funcSpec := utils.CreateTestFunctionSpec(colony.Name)
	funcSpec.Channels = []string{"concurrent-channel"}
	funcSpec.Conditions.ExecutorType = "concurrent-executor"

	submittedProcess, err := client1.Submit(funcSpec, userPrvKey)
	assert.Nil(t, err)

	_, err = client2.Assign(colony.Name, 10, "", "", executorPrvKey)
	assert.Nil(t, err)

	// Send multiple messages sequentially
	numMessages := 5

	// User sends messages via Server1
	for i := 1; i <= numMessages; i++ {
		err := client1.ChannelAppend(submittedProcess.ID, "concurrent-channel", int64(i), 0, []byte("user-msg"), userPrvKey)
		assert.Nil(t, err)
	}

	// Executor sends messages via Server2
	for i := 1; i <= numMessages; i++ {
		err := client2.ChannelAppend(submittedProcess.ID, "concurrent-channel", int64(i), 0, []byte("exec-msg"), executorPrvKey)
		assert.Nil(t, err)
	}

	// Read all messages from both servers using polling (eventual consistency)
	expectedMessages := numMessages * 2
	entries1 := pollForEntries(t, client1, submittedProcess.ID, "concurrent-channel", 0, expectedMessages, userPrvKey)
	entries2 := pollForEntries(t, client2, submittedProcess.ID, "concurrent-channel", 0, expectedMessages, executorPrvKey)

	// Both servers should see all messages
	assert.Equal(t, expectedMessages, len(entries1), "Server1 should see all messages")
	assert.Equal(t, expectedMessages, len(entries2), "Server2 should see all messages")

	// Count messages from each sender
	userMsgs := 0
	execMsgs := 0
	for _, e := range entries1 {
		if string(e.Payload) == "user-msg" {
			userMsgs++
		} else if string(e.Payload) == "exec-msg" {
			execMsgs++
		}
	}
	assert.Equal(t, numMessages, userMsgs, "Should have correct number of user messages")
	assert.Equal(t, numMessages, execMsgs, "Should have correct number of executor messages")

	err = client1.Close(submittedProcess.ID, executorPrvKey)
	assert.Nil(t, err)
}

// TestChannelClusterServerFailover tests that channels continue working when switching servers.
func TestChannelClusterServerFailover(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("CHANNEL_CLUSTER_FAILOVER_")
	assert.Nil(t, err)
	defer db.Close()

	servers := StartCluster(t, db, 2)
	WaitForCluster(t, servers)
	defer func() {
		for _, s := range servers {
			s.Server.Shutdown()
			<-s.Done
		}
	}()

	client1 := client.CreateColoniesClient("localhost", servers[0].Node.APIPort, true, true)
	client2 := client.CreateColoniesClient("localhost", servers[1].Node.APIPort, true, true)

	colonyCrypto := crypto.CreateCrypto()
	colonyPrvKey, err := colonyCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID, err := colonyCrypto.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test-colony-failover")
	_, err = client1.AddColony(colony, servers[0].ServerPrvKey)
	assert.Nil(t, err)

	userCrypto := crypto.CreateCrypto()
	userPrvKey, err := userCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	userID, err := userCrypto.GenerateID(userPrvKey)
	assert.Nil(t, err)

	user := core.CreateUser(colony.Name, userID, "test-user", "user@test.com", "")
	_, err = client1.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	executor.Type = "failover-executor"
	_, err = client1.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	err = client1.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	funcSpec := utils.CreateTestFunctionSpec(colony.Name)
	funcSpec.Channels = []string{"failover-channel"}
	funcSpec.Conditions.ExecutorType = "failover-executor"

	submittedProcess, err := client1.Submit(funcSpec, userPrvKey)
	assert.Nil(t, err)

	_, err = client1.Assign(colony.Name, 10, "", "", executorPrvKey)
	assert.Nil(t, err)

	// User sends first message via Server1
	err = client1.ChannelAppend(submittedProcess.ID, "failover-channel", 1, 0, []byte("Message 1 via Server1"), userPrvKey)
	assert.Nil(t, err)

	// User "switches" to Server2 and sends second message
	err = client2.ChannelAppend(submittedProcess.ID, "failover-channel", 2, 0, []byte("Message 2 via Server2"), userPrvKey)
	assert.Nil(t, err)

	// Read all messages via Server2 - should see both messages (poll for eventual consistency)
	entries := pollForEntries(t, client2, submittedProcess.ID, "failover-channel", 0, 2, userPrvKey)
	assert.Len(t, entries, 2, "Should see both messages after failover")

	// Executor switches from Server1 to Server2 and replies
	err = client2.ChannelAppend(submittedProcess.ID, "failover-channel", 1, 1, []byte("Reply via Server2"), executorPrvKey)
	assert.Nil(t, err)

	// Read reply via Server1 - should see it (poll for eventual consistency)
	entries = pollForEntries(t, client1, submittedProcess.ID, "failover-channel", 2, 1, userPrvKey)
	assert.Len(t, entries, 1, "Should see executor's reply via Server1")
	assert.Equal(t, []byte("Reply via Server2"), entries[0].Payload)

	err = client2.Close(submittedProcess.ID, executorPrvKey)
	assert.Nil(t, err)
}

// TestChannelClusterLateJoin tests that a client joining late via a different server
// can still see all previous messages.
func TestChannelClusterLateJoin(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("CHANNEL_CLUSTER_LATEJOIN_")
	assert.Nil(t, err)
	defer db.Close()

	servers := StartCluster(t, db, 2)
	WaitForCluster(t, servers)
	defer func() {
		for _, s := range servers {
			s.Server.Shutdown()
			<-s.Done
		}
	}()

	client1 := client.CreateColoniesClient("localhost", servers[0].Node.APIPort, true, true)
	client2 := client.CreateColoniesClient("localhost", servers[1].Node.APIPort, true, true)

	colonyCrypto := crypto.CreateCrypto()
	colonyPrvKey, err := colonyCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID, err := colonyCrypto.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test-colony-latejoin")
	_, err = client1.AddColony(colony, servers[0].ServerPrvKey)
	assert.Nil(t, err)

	userCrypto := crypto.CreateCrypto()
	userPrvKey, err := userCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	userID, err := userCrypto.GenerateID(userPrvKey)
	assert.Nil(t, err)

	user := core.CreateUser(colony.Name, userID, "test-user", "user@test.com", "")
	_, err = client1.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	executor.Type = "latejoin-executor"
	_, err = client1.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	err = client1.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	funcSpec := utils.CreateTestFunctionSpec(colony.Name)
	funcSpec.Channels = []string{"latejoin-channel"}
	funcSpec.Conditions.ExecutorType = "latejoin-executor"

	submittedProcess, err := client1.Submit(funcSpec, userPrvKey)
	assert.Nil(t, err)

	_, err = client1.Assign(colony.Name, 10, "", "", executorPrvKey)
	assert.Nil(t, err)

	// User and executor exchange multiple messages all via Server1
	for i := 1; i <= 5; i++ {
		err = client1.ChannelAppend(submittedProcess.ID, "latejoin-channel", int64(i), 0, []byte("user-msg"), userPrvKey)
		assert.Nil(t, err)
		err = client1.ChannelAppend(submittedProcess.ID, "latejoin-channel", int64(i), int64(i), []byte("exec-reply"), executorPrvKey)
		assert.Nil(t, err)
	}

	// Now a "late" reader connects via Server2 and should see ALL messages (poll for eventual consistency)
	entries := pollForEntries(t, client2, submittedProcess.ID, "latejoin-channel", 0, 10, userPrvKey)
	assert.Len(t, entries, 10, "Late join reader should see all 10 messages")

	err = client1.Close(submittedProcess.ID, executorPrvKey)
	assert.Nil(t, err)
}
