package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddExecutorSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)
	executor3, _, err := utils.CreateTestExecutorWithKey(env.colony1Name)
	assert.Nil(t, err)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2
	//   executor3 is bound to colony1, but not yet a member

	// Now, try to add executor 3 to colony1 using colony 2 credentials
	_, err = client.AddExecutor(executor3, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now, try to add executor 3 to colony1 using colony 1 credentials
	_, err = client.AddExecutor(executor3, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetExecutorsByColonySecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Now try to access executor1 using credential of executor2
	_, err := client.GetExecutors(env.colony1Name, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to access executor1 using executor1 credential
	_, err = client.GetExecutors(env.colony1Name, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access executor1 using colony1 credential
	_, err = client.GetExecutors(env.colony1Name, env.colony1PrvKey)
	assert.Nil(t, err) // Should work, colony owner can also get executors

	// Now try to access executor1 using colony1 credential
	_, err = client.GetExecutors(env.colony1Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work, cannot use colony2 credential

	server.Shutdown()
	<-done
}

func TestGetExecutorSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Now try to access executor1 using credentials of executor2
	_, err := client.GetExecutor(env.colony1Name, env.executor1Name, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to access executor1 using executor1 credential
	_, err = client.GetExecutor(env.colony1Name, env.executor1Name, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access executor1 using colony1 credential
	_, err = client.GetExecutor(env.colony1Name, env.executor1Name, env.colony1PrvKey)
	assert.NotNil(t, err) // Should work, cannot use colony1 crendential

	// Now try to access executor1 using colony1 credential
	_, err = client.GetExecutor(env.colony1Name, env.executor1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	server.Shutdown()
	<-done
}

func TestApproveExecutorSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)
	executor3, _, err := utils.CreateTestExecutorWithKey(env.colony1Name)
	assert.Nil(t, err)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2
	//   executor3 is a not yet approved member of colony 1

	_, err = client.AddExecutor(executor3, env.colony1PrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(env.colony2Name, executor3.Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.ApproveExecutor(env.colony1Name, executor3.Name, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestRejectExecutorSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)
	executor3, _, err := utils.CreateTestExecutorWithKey(env.colony1Name)
	assert.Nil(t, err)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2
	//   executor3 is a not yet approved member of colony 1

	_, err = client.AddExecutor(executor3, env.colony1PrvKey)
	assert.Nil(t, err)

	err = client.RejectExecutor(env.colony2Name, executor3.Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RejectExecutor(env.colony1Name, executor3.Name, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestNonApprovedExecutorSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Add another executor to colony1 and list all executors, it should be possible
	executor3, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.colony1Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor3, env.colony1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetExecutors(env.colony1Name, executor3PrvKey)
	assert.Nil(t, err) // Should work, executor should be able to list all executors even if it is not approved

	err = client.ApproveExecutor(env.colony1Name, executor3.Name, env.colony1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetExecutors(env.colony1Name, executor3PrvKey)
	assert.Nil(t, err) // Should also work

	server.Shutdown()
	<-done
}

func TestRemoveExecutorSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Add another executor to colony1 and list all executors, it should be possible
	executor3, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.colony1Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor3, env.colony1PrvKey)
	assert.Nil(t, err)

	err = client.RemoveExecutor(env.colony1Name, executor3.Name, executor3PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveExecutor(env.colony1Name, executor3.Name, env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveExecutor(env.colony1Name, executor3.Name, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveExecutor(env.colony1Name, executor3.Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveExecutor(env.colony1Name, executor3.Name, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}
