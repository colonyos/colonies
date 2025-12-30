package server

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/constants"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// TestChannelEndToEndIntegration tests the complete channel workflow:
// 1. Start Colonies server
// 2. Create colony and executor
// 3. Submit process with channels
// 4. Executor assigns process
// 5. Client and executor communicate via channels using HTTP
//
// This test verifies the complete channel workflow from process submission
// through HTTP channel communication to process completion.
func TestChannelEndToEndIntegration(t *testing.T) {
	// Setup test database
	db, err := postgresql.PrepareTestsWithPrefix("CHANNEL_TEST_")
	assert.Nil(t, err)
	defer db.Close()

	// Create server
	port := 8081
	thisNode := cluster.Node{
		Name:           "test-node",
		Host:           "localhost",
		APIPort:        port,
		EtcdClientPort: 2379,
		EtcdPeerPort:   2380,
		RelayPort:      25100,
	}
	clusterConfig := cluster.Config{
		Nodes: []cluster.Node{thisNode},
	}

	server := CreateServer(
		db,
		port,
		false, // no TLS
		"",
		"",
		thisNode,
		clusterConfig,
		"/tmp/test-etcd-"+time.Now().Format("20060102150405"), // etcd path in /tmp
		10,    // generator period
		10,    // cron period
		false, // exclusive assign
		true,  // allow executor reregister
		false, // retention
		0,     // retention policy
		0,     // retention period
		time.Duration(constants.DEFAULT_STALE_EXECUTOR_DURATION)*time.Second, // stale executor duration
	)

	// Start server in background
	go server.ServeForever()
	time.Sleep(500 * time.Millisecond) // Wait for server to start

	// Create client
	colonies := client.CreateColoniesClient("localhost", port, true, true)

	// Create server private key and register it
	serverCrypto := crypto.CreateCrypto()
	serverPrvKey, err := serverCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	serverID, err := serverCrypto.GenerateID(serverPrvKey)
	assert.Nil(t, err)
	err = db.SetServerID("", serverID) // empty oldServerID for initial setup
	assert.Nil(t, err)

	// Create colony
	colonyCrypto := crypto.CreateCrypto()
	colonyPrvKey, err := colonyCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID, err := colonyCrypto.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test-colony")
	addedColony, err := colonies.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedColony)

	// Create user (process submitter)
	userCrypto := crypto.CreateCrypto()
	userPrvKey, err := userCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	userID, err := userCrypto.GenerateID(userPrvKey)
	assert.Nil(t, err)

	user := core.CreateUser(colony.Name, userID, "test-user", "user@test.com", "")
	addedUser, err := colonies.AddUser(user, colonyPrvKey) // Use colony key, not server key
	assert.Nil(t, err)
	assert.NotNil(t, addedUser)

	// Create executor
	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	executor.Type = "ollama-executor"
	executorID := executor.ID
	addedExecutor, err := colonies.AddExecutor(executor, colonyPrvKey) // Use colony key
	assert.Nil(t, err)
	assert.NotNil(t, addedExecutor)
	err = colonies.ApproveExecutor(colony.Name, addedExecutor.Name, colonyPrvKey) // Use colony key
	assert.Nil(t, err)

	// Submit process with channels
	funcSpec := utils.CreateTestFunctionSpec(colony.Name)
	funcSpec.Channels = []string{"chat"}
	funcSpec.Conditions.ExecutorType = "ollama-executor"

	submittedProcess, err := colonies.Submit(funcSpec, userPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedProcess)
	assert.Equal(t, core.WAITING, submittedProcess.State)
	t.Logf("Submitted process ID: %s, Channels in FuncSpec: %v", submittedProcess.ID, submittedProcess.FunctionSpec.Channels)

	// Executor assigns the process
	assignedProcess, err := colonies.Assign(colony.Name, 10, "", "", executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, assignedProcess)
	assert.Equal(t, submittedProcess.ID, assignedProcess.ID)
	assert.Equal(t, core.RUNNING, assignedProcess.State)

	// User sends first message to channel
	err = colonies.ChannelAppend(submittedProcess.ID, "chat", 1, 0, []byte("What is 2+2?"), userPrvKey)
	assert.Nil(t, err)

	// Executor reads from channel
	entries, err := colonies.ChannelRead(submittedProcess.ID, "chat", 0, 0, executorPrvKey)
	assert.Nil(t, err)
	if len(entries) == 0 {
		t.Fatalf("Expected 1 entry, got 0. Channel might not have user's message yet")
	}
	assert.Len(t, entries, 1)
	assert.Equal(t, int64(1), entries[0].Sequence)
	assert.Equal(t, userID, entries[0].SenderID)
	assert.Equal(t, []byte("What is 2+2?"), entries[0].Payload)
	assert.Equal(t, int64(0), entries[0].InReplyTo)

	// Executor streams response tokens, all referencing client sequence 1
	err = colonies.ChannelAppend(submittedProcess.ID, "chat", 1, 1, []byte("The"), executorPrvKey)
	assert.Nil(t, err)
	err = colonies.ChannelAppend(submittedProcess.ID, "chat", 2, 1, []byte(" answer"), executorPrvKey)
	assert.Nil(t, err)
	err = colonies.ChannelAppend(submittedProcess.ID, "chat", 3, 1, []byte(" is"), executorPrvKey)
	assert.Nil(t, err)
	err = colonies.ChannelAppend(submittedProcess.ID, "chat", 4, 1, []byte(" 4"), executorPrvKey)
	assert.Nil(t, err)

	// User reads responses
	entries, err = colonies.ChannelRead(submittedProcess.ID, "chat", 1, 0, userPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 4) // 4 tokens from executor

	// Verify all tokens reference the client's question
	for _, entry := range entries {
		assert.Equal(t, executorID, entry.SenderID)
		assert.Equal(t, int64(1), entry.InReplyTo) // All reference client seq 1
	}

	// Reconstruct streamed response
	response := ""
	for _, entry := range entries {
		response += string(entry.Payload)
	}
	assert.Equal(t, "The answer is 4", response)

	// User sends second question
	err = colonies.ChannelAppend(submittedProcess.ID, "chat", 2, 0, []byte("What is 3+3?"), userPrvKey)
	assert.Nil(t, err)

	// Executor reads new messages from index 5 (after previous 4 tokens + 1 question)
	entries, err = colonies.ChannelRead(submittedProcess.ID, "chat", 5, 0, executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, int64(2), entries[0].Sequence)
	assert.Equal(t, []byte("What is 3+3?"), entries[0].Payload)

	// Executor replies
	err = colonies.ChannelAppend(submittedProcess.ID, "chat", 5, 2, []byte("6"), executorPrvKey)
	assert.Nil(t, err)

	// User reads final message
	entries, err = colonies.ChannelRead(submittedProcess.ID, "chat", 6, 0, userPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, int64(5), entries[0].Sequence)
	assert.Equal(t, int64(2), entries[0].InReplyTo) // References client seq 2
	assert.Equal(t, []byte("6"), entries[0].Payload)

	// Close process
	err = colonies.Close(submittedProcess.ID, executorPrvKey)
	assert.Nil(t, err)

	// Verify channels are cleaned up (try to read - should fail)
	_, err = colonies.ChannelRead(submittedProcess.ID, "chat", 0, 0, userPrvKey)
	assert.NotNil(t, err) // Channel should be gone

	// Stop server
	server.Shutdown()
}

// TestChannelCleanupOnProcessFail verifies channels are properly cleaned up
// when a process fails (not just when it succeeds).
func TestChannelCleanupOnProcessFail(t *testing.T) {
	// Setup test database
	db, err := postgresql.PrepareTestsWithPrefix("CHANNEL_FAIL_TEST_")
	assert.Nil(t, err)
	defer db.Close()

	// Create server
	port := 8082
	thisNode := cluster.Node{
		Name:           "test-node",
		Host:           "localhost",
		APIPort:        port,
		EtcdClientPort: 2379,
		EtcdPeerPort:   2380,
		RelayPort:      25101,
	}
	clusterConfig := cluster.Config{
		Nodes: []cluster.Node{thisNode},
	}

	server := CreateServer(
		db,
		port,
		false, // no TLS
		"",
		"",
		thisNode,
		clusterConfig,
		"/tmp/test-etcd-fail-"+time.Now().Format("20060102150405"),
		10,    // generator period
		10,    // cron period
		false, // exclusive assign
		true,  // allow executor reregister
		false, // retention
		0,     // retention policy
		0,     // retention period
		time.Duration(constants.DEFAULT_STALE_EXECUTOR_DURATION)*time.Second, // stale executor duration
	)

	// Start server in background
	go server.ServeForever()
	time.Sleep(500 * time.Millisecond)

	// Create client
	colonies := client.CreateColoniesClient("localhost", port, true, true)

	// Create server private key and register it
	serverCrypto := crypto.CreateCrypto()
	serverPrvKey, err := serverCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	serverID, err := serverCrypto.GenerateID(serverPrvKey)
	assert.Nil(t, err)
	err = db.SetServerID("", serverID)
	assert.Nil(t, err)

	// Create colony
	colonyCrypto := crypto.CreateCrypto()
	colonyPrvKey, err := colonyCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID, err := colonyCrypto.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test-colony-fail")
	_, err = colonies.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Create user
	userCrypto := crypto.CreateCrypto()
	userPrvKey, err := userCrypto.GeneratePrivateKey()
	assert.Nil(t, err)
	userID, err := userCrypto.GenerateID(userPrvKey)
	assert.Nil(t, err)

	user := core.CreateUser(colony.Name, userID, "test-user", "user@test.com", "")
	_, err = colonies.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)

	// Create executor
	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	executor.Type = "test-executor"
	_, err = colonies.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	err = colonies.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	// Submit process with channels
	funcSpec := utils.CreateTestFunctionSpec(colony.Name)
	funcSpec.Channels = []string{"data", "control"}
	funcSpec.Conditions.ExecutorType = "test-executor"

	submittedProcess, err := colonies.Submit(funcSpec, userPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedProcess)

	// Executor assigns the process
	assignedProcess, err := colonies.Assign(colony.Name, 10, "", "", executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, submittedProcess.ID, assignedProcess.ID)

	// User sends message to channel
	err = colonies.ChannelAppend(submittedProcess.ID, "data", 1, 0, []byte("test data"), userPrvKey)
	assert.Nil(t, err)

	// Verify channel is accessible
	entries, err := colonies.ChannelRead(submittedProcess.ID, "data", 0, 0, executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)

	// Fail the process
	err = colonies.Fail(submittedProcess.ID, []string{"simulated error"}, executorPrvKey)
	assert.Nil(t, err)

	// Verify process is in FAILED state
	failedProcess, err := colonies.GetProcess(submittedProcess.ID, userPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.FAILED, failedProcess.State)

	// Verify channels are cleaned up (try to read - should fail)
	_, err = colonies.ChannelRead(submittedProcess.ID, "data", 0, 0, userPrvKey)
	assert.NotNil(t, err, "Channel should be cleaned up after process failure")

	_, err = colonies.ChannelRead(submittedProcess.ID, "control", 0, 0, userPrvKey)
	assert.NotNil(t, err, "All channels should be cleaned up after process failure")

	// Stop server
	server.Shutdown()
}
