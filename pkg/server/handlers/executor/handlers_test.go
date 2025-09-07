package executor_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddExecutor(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	addedExecutor, err := client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	assert.True(t, executor.Equals(addedExecutor))
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	// Just to make the comparison below work, the state will change after it has been approved
	addedExecutor.State = core.APPROVED

	executorFromServer, err := client.GetExecutor(colony.Name, executor.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, executorFromServer)
	assert.True(t, addedExecutor.Equals(executorFromServer))

	server.Shutdown()
	<-done
}

func TestReportAllocations(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	addedExecutor, err := client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	assert.True(t, executor.Equals(addedExecutor))
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	project := core.Project{AllocatedCPU: 1, UsedCPU: 2, AllocatedGPU: 3, UsedGPU: 4, AllocatedStorage: 5, UsedStorage: 6}
	projects := make(map[string]core.Project)
	projects["test_project"] = project
	alloc := core.Allocations{Projects: projects}

	err = client.ReportAllocation(colony.Name, executor.Name, alloc, executorPrvKey)
	assert.Nil(t, err)

	executorFromServer, err := client.GetExecutor(colony.Name, executor.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, executorFromServer)

	testProj := executorFromServer.Allocations.Projects["test_project"]
	assert.Equal(t, testProj.AllocatedCPU, int64(1))
	assert.Equal(t, testProj.UsedCPU, int64(2))
	assert.Equal(t, testProj.AllocatedGPU, int64(3))
	assert.Equal(t, testProj.UsedGPU, int64(4))
	assert.Equal(t, testProj.AllocatedStorage, int64(5))
	assert.Equal(t, testProj.UsedStorage, int64(6))

	server.Shutdown()
	<-done
}

func TestAddExecutorReRegister(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, _, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.NotNil(t, err)

	server.SetAllowExecutorReregister(true)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetExecutors(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	_, err = client.AddExecutor(executor1, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony.Name, executor1.Name, colonyPrvKey)
	assert.Nil(t, err)

	executor2, _, err := utils.CreateTestExecutorWithKey(colony.Name)
	_, err = client.AddExecutor(executor2, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony.Name, executor2.Name, colonyPrvKey)
	assert.Nil(t, err)

	// Just to make the comparison below work, the state will change after it has been approved
	executor1.State = core.APPROVED
	executor2.State = core.APPROVED

	var executors []*core.Executor
	executors = append(executors, executor1)
	executors = append(executors, executor2)

	executorsFromServer, err := client.GetExecutors(colony.Name, executor1PrvKey)
	assert.Nil(t, err)
	assert.True(t, core.IsExecutorArraysEqual(executors, executorsFromServer))

	server.Shutdown()
	<-done
}

func TestApproveRejectExecutor(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add an approved eecutor to use for the test below
	approvedExecutor, approvedExecutorPrvKey, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	_, err = client.AddExecutor(approvedExecutor, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, approvedExecutor.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	testExecutor, _, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	_, err = client.AddExecutor(testExecutor, env.ColonyPrvKey)
	assert.Nil(t, err)

	eecutorFromServer, err := client.GetExecutor(env.ColonyName, testExecutor.Name, approvedExecutorPrvKey)
	assert.Nil(t, err)
	assert.False(t, eecutorFromServer.IsApproved())

	err = client.ApproveExecutor(env.ColonyName, testExecutor.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	eecutorFromServer, err = client.GetExecutor(env.ColonyName, testExecutor.Name, approvedExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, eecutorFromServer.IsApproved())

	err = client.RejectExecutor(env.ColonyName, testExecutor.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	eecutorFromServer, err = client.GetExecutor(env.ColonyName, testExecutor.Name, approvedExecutorPrvKey)
	assert.Nil(t, err)
	assert.False(t, eecutorFromServer.IsApproved())

	server.Shutdown()
	<-done
}

func TestRemoveExecutor(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Try to get it
	executorFromServer, err := client.GetExecutor(env.ColonyName, executor.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, executorFromServer)
	assert.True(t, executor.ID == executorFromServer.ID)

	// Now remove it
	err = client.RemoveExecutor(env.ColonyName, executor.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Try to get it again, it should be gone
	executorFromServer, err = client.GetExecutor(env.ColonyName, executor.Name, executorPrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, executorFromServer)

	server.Shutdown()
	<-done
}
