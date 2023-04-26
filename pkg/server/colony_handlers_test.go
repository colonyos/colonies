package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddColony(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	server.Shutdown()
	<-done
}

func TestDeleteColony(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	coloniesFromServer, err := client.GetColonies(serverPrvKey)
	assert.Nil(t, err)
	assert.Len(t, coloniesFromServer, 1)

	err = client.DeleteColony(addedColony.ID, serverPrvKey)
	assert.Nil(t, err)

	coloniesFromServer, err = client.GetColonies(serverPrvKey)
	assert.Nil(t, err)
	assert.Len(t, coloniesFromServer, 0)

	server.Shutdown()
	<-done
}

func TestRenameColony(t *testing.T) {
	_, client, server, serverPrvKey, done := setupTestEnv2(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	colony.Name = "test_name_1"
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(addedColony.ID)
	assert.Nil(t, err)

	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(executor.ID, colonyPrvKey)
	assert.Nil(t, err)

	colonyFromServer, err := client.GetColonyByID(addedColony.ID, executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, colonyFromServer.Name, "test_name_1")

	err = client.RenameColony(addedColony.ID, "test_name_2", colonyPrvKey)
	assert.Nil(t, err)

	colonyFromServer, err = client.GetColonyByID(addedColony.ID, executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, colonyFromServer.Name, "test_name_2")

	server.Shutdown()
	<-done
}

func TestGetColony(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.ID)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(executor.ID, colonyPrvKey)
	assert.Nil(t, err)

	colonyFromServer, err := client.GetColonyByID(colony.ID, executorPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(colonyFromServer))

	server.Shutdown()
	<-done
}

func TestGetColonies(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony1, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	colony2, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	var colonies []*core.Colony
	colonies = append(colonies, colony1)
	colonies = append(colonies, colony2)

	coloniesFromServer, err := client.GetColonies(serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, core.IsColonyArraysEqual(colonies, coloniesFromServer))

	server.Shutdown()
	<-done
}

func TestGetColonyStatistics(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	// Waiting
	numberOfWaitingProcesses := 2
	for i := 0; i < numberOfWaitingProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
	}

	// Running
	numberOfRunningProcesses := 3
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
		_, err = client.Assign(env.colonyID, -1, env.executorPrvKey)
	}

	// Successful
	numberOfSuccessfulProcesses := 1
	for i := 0; i < numberOfSuccessfulProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
		assert.Nil(t, err)
		err = client.Close(processFromServer.ID, env.executorPrvKey)
		assert.Nil(t, err)
	}

	// Failed
	numberOfFailedProcesses := 2
	for i := 0; i < numberOfFailedProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
		assert.Nil(t, err)
		err = client.Fail(processFromServer.ID, []string{"error"}, env.executorPrvKey)
		assert.Nil(t, err)
	}

	stat, err := client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)

	assert.Equal(t, stat.WaitingProcesses, numberOfWaitingProcesses)
	assert.Equal(t, stat.RunningProcesses, numberOfRunningProcesses)
	assert.Equal(t, stat.SuccessfulProcesses, numberOfSuccessfulProcesses)
	assert.Equal(t, stat.FailedProcesses, numberOfFailedProcesses)

	server.Shutdown()
	<-done
}
