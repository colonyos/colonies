package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetStatistics(t *testing.T) {
	env, client, server, serverPrvKey, done := setupTestEnv2(t)

	// Waiting
	numberOfWaitingProcesses := 5
	for i := 0; i < numberOfWaitingProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	// Running
	numberOfRunningProcesses := 6
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		_, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
	}

	// Successful
	numberOfSuccessfulProcesses := 7
	for i := 0; i < numberOfSuccessfulProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
		err = client.CloseSuccessful(processFromServer.ID, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	// Failed
	numberOfFailedProcesses := 8
	for i := 0; i < numberOfFailedProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
		err = client.CloseFailed(processFromServer.ID, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	stat, err := client.Statistics(serverPrvKey)
	assert.Nil(t, err)

	assert.Equal(t, stat.Waiting, numberOfWaitingProcesses)
	assert.Equal(t, stat.Running, numberOfRunningProcesses)
	assert.Equal(t, stat.Success, numberOfSuccessfulProcesses)
	assert.Equal(t, stat.Failed, numberOfFailedProcesses)

	server.Shutdown()
	<-done
}
