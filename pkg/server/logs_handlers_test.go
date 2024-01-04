package server

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddGetLogByProcess(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyName)
	addedProcess, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(addedProcess.ID, "test_msg", env.executorPrvKey)
	assert.NotNil(t, err) // Failed to add log, not allowed to add log

	err = client.AddLog("invalid_process_id", "test_msg", env.executorPrvKey)
	assert.NotNil(t, err) // Failed to add log, process is nil

	assignedProcess, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(assignedProcess.ID, "test_msg", env.executorPrvKey)
	assert.Nil(t, err)

	_, err = client.GetLogsByProcessID(env.colonyName, assignedProcess.ID, MAX_LOG_COUNT+1, env.executorPrvKey)
	assert.NotNil(t, err) // Exceeds max log count

	logs, err := client.GetLogsByProcessID(env.colonyName, assignedProcess.ID, 100, env.executorPrvKey)
	assert.Len(t, logs, 1)
	assert.Equal(t, logs[0].Message, "test_msg")
	assert.Equal(t, logs[0].ProcessID, assignedProcess.ID)
	assert.Equal(t, logs[0].ColonyName, env.colonyName)

	server.Shutdown()
	<-done
}

func TestAddGetLogSinceByProcess(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyName)
	_, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(assignedProcess.ID, "test_msg1", env.executorPrvKey)
	assert.Nil(t, err)

	time.Sleep(2 * time.Second)

	err = client.AddLog(assignedProcess.ID, "test_msg2", env.executorPrvKey)
	assert.Nil(t, err)

	logs, err := client.GetLogsByProcessID(env.colonyName, assignedProcess.ID, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, logs, 2)

	since := logs[0].Timestamp
	logs, err = client.GetLogsByProcessIDSince(env.colonyName, assignedProcess.ID, 100, since, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, logs[0].Message, "test_msg2")

	server.Shutdown()
	<-done
}

func TestAddGetLogByExecutor(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	// Process 1
	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyName)
	_, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(assignedProcess.ID, "test_msg_process_1", env.executorPrvKey)
	assert.Nil(t, err)

	err = client.Close(assignedProcess.ID, env.executorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(assignedProcess.ID, "test_msg_process_1_2", env.executorPrvKey)
	assert.NotNil(t, err) // Not possible to add logs to closed processes

	// Process 2
	funcSpec1 = utils.CreateTestFunctionSpec(env.colonyName)
	_, err = client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(assignedProcess.ID, "test_msg_process_2", env.executorPrvKey)
	assert.Nil(t, err)

	logs, err := client.GetLogsByExecutor(env.colonyName, env.executorName, 10, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, logs, 2)

	err = client.AddLog(assignedProcess.ID, "test_msg_process_2_2", env.executorPrvKey)
	assert.Nil(t, err)

	logs, err = client.GetLogsByExecutorSince(env.colonyName, env.executorName, 10, logs[len(logs)-1].Timestamp, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, logs[0].Message, "test_msg_process_2_2")

	server.Shutdown()
	<-done
}
