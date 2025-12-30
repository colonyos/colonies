package colony_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddColony(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	server.Shutdown()
	<-done
}

func TestRemoveColony(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	coloniesFromServer, err := client.GetColonies(serverPrvKey)
	assert.Nil(t, err)
	assert.Len(t, coloniesFromServer, 1)

	err = client.RemoveColony(addedColony.Name, serverPrvKey)
	assert.Nil(t, err)

	coloniesFromServer, err = client.GetColonies(serverPrvKey)
	assert.Nil(t, err)
	assert.Len(t, coloniesFromServer, 0)

	server.Shutdown()
	<-done
}

func TestGetColony(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	colonyFromServer, err := client.GetColonyByName(colony.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(colonyFromServer))

	server.Shutdown()
	<-done
}

func TestGetColonies(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

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
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Waiting
	numberOfWaitingProcesses := 2
	for i := 0; i < numberOfWaitingProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
		_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
	}

	// Running
	numberOfRunningProcesses := 3
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
		_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
		_, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	}

	// Successful
	numberOfSuccessfulProcesses := 1
	for i := 0; i < numberOfSuccessfulProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
		_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
		assert.Nil(t, err)
		err = client.Close(processFromServer.ID, env.ExecutorPrvKey)
		assert.Nil(t, err)
	}

	// Failed
	numberOfFailedProcesses := 2
	for i := 0; i < numberOfFailedProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
		_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
		assert.Nil(t, err)
		err = client.Fail(processFromServer.ID, []string{"error"}, env.ExecutorPrvKey)
		assert.Nil(t, err)
	}

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assert.Equal(t, stat.WaitingProcesses, numberOfWaitingProcesses)
	assert.Equal(t, stat.RunningProcesses, numberOfRunningProcesses)
	assert.Equal(t, stat.SuccessfulProcesses, numberOfSuccessfulProcesses)
	assert.Equal(t, stat.FailedProcesses, numberOfFailedProcesses)

	server.Shutdown()
	<-done
}

// TestAddColonyDuplicate tests adding a colony with duplicate name
func TestAddColonyDuplicate(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Try to add colony with same name
	colony2, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	colony2.Name = colony.Name
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestAddColonyUnauthorized tests adding colony without server key
func TestAddColonyUnauthorized(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Try to add another colony with colony key instead of server key
	colony2, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, colonyPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestRemoveColonyUnauthorized tests removing colony without server key
func TestRemoveColonyUnauthorized(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Try to remove colony with colony key instead of server key
	err = client.RemoveColony(colony.Name, colonyPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestRemoveColonyNotFound tests removing a non-existent colony
func TestRemoveColonyNotFound(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	err := client.RemoveColony("non_existent_colony", serverPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestGetColoniesUnauthorized tests getting colonies without server key
func TestGetColoniesUnauthorized(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Try to get colonies with colony key instead of server key
	_, err = client.GetColonies(colonyPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestGetColonyUnauthorized tests getting a colony without membership
func TestGetColonyUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, _, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Try to get colony1 with executor2's key
	_, err = client.GetColonyByName(colony1.Name, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetColonyStatisticsNotFound tests getting statistics for non-existent colony
func TestGetColonyStatisticsNotFound(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	_, err := client.ColonyStatistics("non_existent_colony", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}
