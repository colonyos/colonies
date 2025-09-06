package executor_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/service"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddExecutorSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)
	executor3, _, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2
	//   executor3 is bound to colony1, but not yet a member

	// Now, try to add executor 3 to colony1 using colony 2 credentials
	_, err = client.AddExecutor(executor3, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now, try to add executor 3 to colony1 using colony 1 credentials
	_, err = client.AddExecutor(executor3, env.Colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestReportAllocationSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2
	//   executor3 is bound to colony1, but not yet a member

	project := core.Project{AllocatedCPU: 1, UsedCPU: 2, AllocatedGPU: 3, UsedGPU: 4, AllocatedStorage: 5, UsedStorage: 6}
	projects := make(map[string]core.Project)
	projects["test_project"] = project
	alloc := core.Allocations{Projects: projects}

	err := client.ReportAllocation(env.Colony1Name, env.Executor1Name, alloc, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.ReportAllocation(env.Colony1Name, env.Executor1Name, alloc, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.ReportAllocation(env.Colony1Name, env.Executor1Name, alloc, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.ReportAllocation(env.Colony1Name, env.Executor1Name, alloc, env.Executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetExecutorsByColonySecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Now try to access executor1 using credential of executor2
	_, err := client.GetExecutors(env.Colony1Name, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to access executor1 using executor1 credential
	_, err = client.GetExecutors(env.Colony1Name, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access executor1 using colony1 credential
	_, err = client.GetExecutors(env.Colony1Name, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work, colony owner can't get executors

	// Now try to access executor1 using colony1 credential
	_, err = client.GetExecutors(env.Colony1Name, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work, cannot use colony2 credential

	server.Shutdown()
	<-done
}

func TestGetExecutorSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Now try to access executor1 using credentials of executor2
	_, err := client.GetExecutor(env.Colony1Name, env.Executor1Name, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to access executor1 using executor1 credential
	_, err = client.GetExecutor(env.Colony1Name, env.Executor1Name, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access executor1 using colony1 credential
	_, err = client.GetExecutor(env.Colony1Name, env.Executor1Name, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should work, cannot use colony1 crendential

	// Now try to access executor1 using colony1 credential
	_, err = client.GetExecutor(env.Colony1Name, env.Executor1ID, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	server.Shutdown()
	<-done
}

func TestApproveExecutorSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)
	executor3, _, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2
	//   executor3 is a not yet approved member of colony 1

	_, err = client.AddExecutor(executor3, env.Colony1PrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(env.Colony2Name, executor3.Name, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.ApproveExecutor(env.Colony1Name, executor3.Name, env.Colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestRejectExecutorSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)
	executor3, _, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2
	//   executor3 is a not yet approved member of colony 1

	_, err = client.AddExecutor(executor3, env.Colony1PrvKey)
	assert.Nil(t, err)

	err = client.RejectExecutor(env.Colony2Name, executor3.Name, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RejectExecutor(env.Colony1Name, executor3.Name, env.Colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestNonApprovedExecutorSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Add another executor to colony1 and list all executors, it should be possible
	executor3, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor3, env.Colony1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetExecutors(env.Colony1Name, executor3PrvKey)
	assert.Nil(t, err) // Should work, executor should be able to list all executors even if it is not approved

	err = client.ApproveExecutor(env.Colony1Name, executor3.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetExecutors(env.Colony1Name, executor3PrvKey)
	assert.Nil(t, err) // Should also work

	server.Shutdown()
	<-done
}

func TestRemoveExecutorSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Add another executor to colony1 and list all executors, it should be possible
	executor3, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor3, env.Colony1PrvKey)
	assert.Nil(t, err)

	err = client.RemoveExecutor(env.Colony1Name, executor3.Name, executor3PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveExecutor(env.Colony1Name, executor3.Name, env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveExecutor(env.Colony1Name, executor3.Name, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveExecutor(env.Colony1Name, executor3.Name, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveExecutor(env.Colony1Name, executor3.Name, env.Colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}
