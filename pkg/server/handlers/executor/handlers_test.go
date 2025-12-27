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

// TestReconcilerWorkflow simulates the docker-reconciler workflow:
// 1. Reconciler (with colony owner key) adds executor
// 2. Reconciler approves executor
// 3. Child executor (inside container) calls GetExecutor
// 4. Child executor calls UpdateExecutorCapabilities
// 5. Verify executor is still APPROVED after all operations
func TestReconcilerWorkflow(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	// Create colony
	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Step 1: Reconciler adds executor (using colony owner key)
	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	addedExecutor, err := client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedExecutor)
	t.Logf("Added executor: Name=%s, ID=%s, State=%d", addedExecutor.Name, addedExecutor.ID, addedExecutor.State)

	// Verify initial state is PENDING
	assert.Equal(t, core.PENDING, addedExecutor.State)

	// Step 2: Reconciler approves executor (using colony owner key)
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)
	t.Log("ApproveExecutor returned success")

	// Verify state is APPROVED immediately after approval (using executor key since it's now approved)
	executorAfterApprove, err := client.GetExecutor(colony.Name, executor.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, executorAfterApprove)
	t.Logf("After approve: Name=%s, ID=%s, State=%d", executorAfterApprove.Name, executorAfterApprove.ID, executorAfterApprove.State)
	assert.Equal(t, core.APPROVED, executorAfterApprove.State, "Executor should be APPROVED after ApproveExecutor")

	// Step 3: Child executor calls GetExecutor again (using its own key)
	executorFromChild, err := client.GetExecutor(colony.Name, executor.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, executorFromChild)
	t.Logf("Child GetExecutor: Name=%s, ID=%s, State=%d", executorFromChild.Name, executorFromChild.ID, executorFromChild.State)
	assert.Equal(t, core.APPROVED, executorFromChild.State, "Executor should still be APPROVED after child GetExecutor")

	// Step 4: Child executor calls UpdateExecutorCapabilities (using its own key)
	newCapabilities := core.Capabilities{
		Hardware: []core.Hardware{
			{
				Model: "Test Model",
				CPU:   "Test CPU",
				Cores: 8,
			},
		},
		Software: []core.Software{
			{
				Name:    "ollama",
				Type:    "llm",
				Version: "0.1.0",
			},
		},
	}
	err = client.UpdateExecutorCapabilities(colony.Name, executor.Name, newCapabilities, executorPrvKey)
	assert.Nil(t, err)
	t.Log("UpdateExecutorCapabilities returned success")

	// Step 5: Verify executor is still APPROVED after all operations
	executorFinal, err := client.GetExecutor(colony.Name, executor.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, executorFinal)
	t.Logf("Final state: Name=%s, ID=%s, State=%d", executorFinal.Name, executorFinal.ID, executorFinal.State)
	assert.Equal(t, core.APPROVED, executorFinal.State, "Executor should still be APPROVED after UpdateExecutorCapabilities")

	// Verify capabilities were updated
	assert.Len(t, executorFinal.Capabilities.Hardware, 1)
	assert.Equal(t, "Test Model", executorFinal.Capabilities.Hardware[0].Model)
	assert.Len(t, executorFinal.Capabilities.Software, 1)
	assert.Equal(t, "ollama", executorFinal.Capabilities.Software[0].Name)

	server.Shutdown()
	<-done
}

// TestApproveExecutorWithUnregisteredExecutor tests that re-registering an UNREGISTERED
// executor and then approving it works correctly
func TestApproveExecutorWithUnregisteredExecutor(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	// Create colony
	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Create and add executor
	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	// Approve executor
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	// Verify APPROVED
	executorFromServer, err := client.GetExecutor(colony.Name, executor.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.APPROVED, executorFromServer.State)
	t.Logf("Initial executor ID: %s, State: %d", executorFromServer.ID, executorFromServer.State)

	// Remove executor (marks as UNREGISTERED)
	err = client.RemoveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	// Enable re-registration
	server.SetAllowExecutorReregister(true)

	// Re-add the same executor (should reactivate the UNREGISTERED one)
	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	executor2.Name = executor.Name // Same name
	assert.Nil(t, err)
	addedExecutor2, err := client.AddExecutor(executor2, colonyPrvKey)
	assert.Nil(t, err)
	t.Logf("Re-added executor ID: %s, State: %d", addedExecutor2.ID, addedExecutor2.State)

	// The re-added executor should have a NEW ID and be PENDING
	assert.NotEqual(t, executorFromServer.ID, addedExecutor2.ID, "Re-added executor should have new ID")
	assert.Equal(t, core.PENDING, addedExecutor2.State, "Re-added executor should be PENDING")

	// Approve the re-added executor
	err = client.ApproveExecutor(colony.Name, executor2.Name, colonyPrvKey)
	assert.Nil(t, err)

	// Verify APPROVED with the new key
	executorFinal, err := client.GetExecutor(colony.Name, executor2.Name, executor2PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.APPROVED, executorFinal.State, "Re-added executor should be APPROVED")
	assert.Equal(t, addedExecutor2.ID, executorFinal.ID, "Executor ID should match the re-added one")

	server.Shutdown()
	<-done
}

// TestApproveExecutorVerifyDatabaseUpdate verifies that ApproveExecutor actually
// updates the database and the change persists
func TestApproveExecutorVerifyDatabaseUpdate(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	// Create colony
	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Create executor
	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	addedExecutor, err := client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	// Store the executor ID
	executorID := addedExecutor.ID
	t.Logf("Executor ID: %s", executorID)

	// Approve
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	// Get executor multiple times to ensure consistency
	for i := 0; i < 5; i++ {
		executorFromServer, err := client.GetExecutor(colony.Name, executor.Name, executorPrvKey)
		assert.Nil(t, err)
		assert.Equal(t, core.APPROVED, executorFromServer.State, "Iteration %d: Executor should be APPROVED", i)
		assert.Equal(t, executorID, executorFromServer.ID, "Iteration %d: Executor ID should not change", i)
	}

	server.Shutdown()
	<-done
}
