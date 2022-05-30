package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddRuntimeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)
	runtime3, _, err := utils.CreateTestRuntimeWithKey(env.colony1ID)
	assert.Nil(t, err)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2
	//   runtime3 is bound to colony1, but not yet a member

	// Now, try to add runtime 3 to colony1 using colony 2 credentials
	_, err = client.AddRuntime(runtime3, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now, try to add runtime 3 to colony1 using colony 1 credentials
	_, err = client.AddRuntime(runtime3, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetRuntimesByColonySecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Now try to access runtime1 using credential of runtime2
	_, err := client.GetRuntimes(env.colony1ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to access runtime1 using runtime1 credential
	_, err = client.GetRuntimes(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access runtime1 using colony1 credential
	_, err = client.GetRuntimes(env.colony1ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work, colony owner can also get runtimes

	// Now try to access runtime1 using colony1 credential
	_, err = client.GetRuntimes(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work, cannot use colony2 credential

	server.Shutdown()
	<-done
}

func TestGetRuntimeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Now try to access runtime1 using credentials of runtime2
	_, err := client.GetRuntime(env.runtime1ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to access runtime1 using runtime1 credential
	_, err = client.GetRuntime(env.runtime1ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access runtime1 using colony1 credential
	_, err = client.GetRuntime(env.runtime1ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should work, cannot use colony1 crendential

	// Now try to access runtime1 using colony1 credential
	_, err = client.GetRuntime(env.runtime1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	server.Shutdown()
	<-done
}

func TestApproveRuntimeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)
	runtime3, _, err := utils.CreateTestRuntimeWithKey(env.colony1ID)
	assert.Nil(t, err)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2
	//   runtime3 is a not yet approved member of colony 1

	_, err = client.AddRuntime(runtime3, env.colony1PrvKey)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime3.ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.ApproveRuntime(runtime3.ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestRejectRuntimeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)
	runtime3, _, err := utils.CreateTestRuntimeWithKey(env.colony1ID)
	assert.Nil(t, err)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2
	//   runtime3 is a not yet approved member of colony 1

	_, err = client.AddRuntime(runtime3, env.colony1PrvKey)
	assert.Nil(t, err)

	err = client.RejectRuntime(runtime3.ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RejectRuntime(runtime3.ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestNonApprovedRuntimeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Add another runtime to colony1 and list all runtimes, it should be possible
	runtime3, runtime3PrvKey, err := utils.CreateTestRuntimeWithKey(env.colony1ID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(runtime3, env.colony1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetRuntimes(env.colony1ID, runtime3PrvKey)
	assert.Nil(t, err) // Should work, runtime should be able to list all runtimes even if it is not approved

	err = client.ApproveRuntime(runtime3.ID, env.colony1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetRuntimes(env.colony1ID, runtime3PrvKey)
	assert.Nil(t, err) // Should also work

	server.Shutdown()
	<-done
}

func TestDeleteRuntimeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Add another runtime to colony1 and list all runtimes, it should be possible
	runtime3, runtime3PrvKey, err := utils.CreateTestRuntimeWithKey(env.colony1ID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(runtime3, env.colony1PrvKey)
	assert.Nil(t, err)

	err = client.DeleteRuntime(runtime3.ID, runtime3PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteRuntime(runtime3.ID, env.runtime1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteRuntime(runtime3.ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteRuntime(runtime3.ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteRuntime(runtime3.ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}
