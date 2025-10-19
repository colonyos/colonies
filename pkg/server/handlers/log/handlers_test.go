package log_test

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

const MAX_LOG_COUNT = 500
const MAX_COUNT = 100
const MAX_DAYS = 30

func TestAddGetLogByProcess(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	addedProcess, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(addedProcess.ID, "test_msg", env.ExecutorPrvKey)
	assert.NotNil(t, err) // Failed to add log, not allowed to add log

	err = client.AddLog("invalid_process_id", "test_msg", env.ExecutorPrvKey)
	assert.NotNil(t, err) // Failed to add log, process is nil

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(assignedProcess.ID, "test_msg", env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.GetLogsByProcess(env.ColonyName, assignedProcess.ID, MAX_LOG_COUNT+1, env.ExecutorPrvKey)
	assert.NotNil(t, err) // Exceeds max log count

	logs, err := client.GetLogsByProcess(env.ColonyName, assignedProcess.ID, 100, env.ExecutorPrvKey)
	assert.Len(t, logs, 1)
	assert.Equal(t, logs[0].Message, "test_msg")
	assert.Equal(t, logs[0].ProcessID, assignedProcess.ID)
	assert.Equal(t, logs[0].ColonyName, env.ColonyName)

	server.Shutdown()
	<-done
}

func TestAddGetLogSinceByProcess(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(assignedProcess.ID, "test_msg1", env.ExecutorPrvKey)
	assert.Nil(t, err)

	time.Sleep(2 * time.Second)

	err = client.AddLog(assignedProcess.ID, "test_msg2", env.ExecutorPrvKey)
	assert.Nil(t, err)

	logs, err := client.GetLogsByProcess(env.ColonyName, assignedProcess.ID, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, logs, 2)

	since := logs[0].Timestamp
	logs, err = client.GetLogsByProcessSince(env.ColonyName, assignedProcess.ID, 100, since, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, logs[0].Message, "test_msg2")

	server.Shutdown()
	<-done
}

func TestAddGetLogByExecutor(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Process 1
	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(assignedProcess.ID, "test_msg_process_1", env.ExecutorPrvKey)
	assert.Nil(t, err)

	err = client.Close(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(assignedProcess.ID, "test_msg_process_1_2", env.ExecutorPrvKey)
	assert.NotNil(t, err) // Not possible to add logs to closed processes

	// Process 2
	funcSpec1 = utils.CreateTestFunctionSpec(env.ColonyName)
	_, err = client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(assignedProcess.ID, "test_msg_process_2", env.ExecutorPrvKey)
	assert.Nil(t, err)

	logs, err := client.GetLogsByExecutor(env.ColonyName, env.ExecutorName, 10, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, logs, 2)

	err = client.AddLog(assignedProcess.ID, "test_msg_process_2_2", env.ExecutorPrvKey)
	assert.Nil(t, err)

	logs, err = client.GetLogsByExecutorSince(env.ColonyName, env.ExecutorName, 10, logs[len(logs)-1].Timestamp, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, logs[0].Message, "test_msg_process_2_2")

	server.Shutdown()
	<-done
}

func TestSearchLogsByExecutor(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Process 1
	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(assignedProcess.ID, "test_msg_process_1", env.ExecutorPrvKey)
	assert.Nil(t, err)

	err = client.Close(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Process 2
	funcSpec1 = utils.CreateTestFunctionSpec(env.ColonyName)
	_, err = client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(assignedProcess.ID, "ERROR", env.ExecutorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(assignedProcess.ID, "ERROR", env.ExecutorPrvKey)
	assert.Nil(t, err)

	logs, err := client.SearchLogs(env.ColonyName, "ERROR", 1, MAX_COUNT+1, env.ExecutorPrvKey)
	assert.NotNil(t, err) // Exceeds max count

	logs, err = client.SearchLogs(env.ColonyName, "ERROR", MAX_DAYS+1, 1, env.ExecutorPrvKey)
	assert.NotNil(t, err) // Exceeds max count

	logs, err = client.SearchLogs(env.ColonyName, "ERROR", 1, 10, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, logs, 2)
	assert.Equal(t, logs[0].Message, "ERROR")
	assert.Equal(t, logs[1].Message, "ERROR")

	server.Shutdown()
	<-done
}
