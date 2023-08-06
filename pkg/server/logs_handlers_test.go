package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddGetLog(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyID)
	addedProcess, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(addedProcess.ID, "test_msg", env.executorPrvKey)
	assert.NotNil(t, err) // Failed to add log, not allowed to add log

	err = client.AddLog("invalid_process_id", "test_msg", env.executorPrvKey)
	assert.NotNil(t, err) // Failed to add log, process is nil

	assignedProcess, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)

	err = client.AddLog(assignedProcess.ID, "test_msg", env.executorPrvKey)
	assert.Nil(t, err)

	_, err = client.GetLogsByProcessID(assignedProcess.ID, MAX_LOG_COUNT+1, env.executorPrvKey)
	assert.NotNil(t, err) // Exceeds mac log count

	logs, err := client.GetLogsByProcessID(assignedProcess.ID, 100, env.executorPrvKey)
	assert.Len(t, logs, 1)
	assert.Equal(t, logs[0].Message, "test_msg")
	assert.Equal(t, logs[0].ProcessID, assignedProcess.ID)
	assert.Equal(t, logs[0].ColonyID, env.colonyID)

	server.Shutdown()
	<-done
}
