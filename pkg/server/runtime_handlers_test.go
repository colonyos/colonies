package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddRuntime(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	runtime, runtimePrvKey, err := utils.CreateTestRuntimeWithKey(colony.ID)
	assert.Nil(t, err)
	addedRuntime, err := client.AddRuntime(runtime, colonyPrvKey)
	assert.Nil(t, err)
	assert.True(t, runtime.Equals(addedRuntime))
	err = client.ApproveRuntime(runtime.ID, colonyPrvKey)
	assert.Nil(t, err)

	// Just to make the comparison below work, the state will change after it has been approved
	addedRuntime.State = core.APPROVED

	runtimeFromServer, err := client.GetRuntime(runtime.ID, runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, runtimeFromServer)
	assert.True(t, addedRuntime.Equals(runtimeFromServer))

	server.Shutdown()
	<-done
}

func TestGetRuntimes(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	runtime1, runtime1PrvKey, err := utils.CreateTestRuntimeWithKey(colony.ID)
	_, err = client.AddRuntime(runtime1, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(runtime1.ID, colonyPrvKey)
	assert.Nil(t, err)

	runtime2, _, err := utils.CreateTestRuntimeWithKey(colony.ID)
	_, err = client.AddRuntime(runtime2, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(runtime2.ID, colonyPrvKey)
	assert.Nil(t, err)

	// Just to make the comparison below work, the state will change after it has been approved
	runtime1.State = core.APPROVED
	runtime2.State = core.APPROVED

	var runtimes []*core.Runtime
	runtimes = append(runtimes, runtime1)
	runtimes = append(runtimes, runtime2)

	runtimesFromServer, err := client.GetRuntimes(colony.ID, runtime1PrvKey)
	assert.Nil(t, err)
	assert.True(t, core.IsRuntimeArraysEqual(runtimes, runtimesFromServer))

	server.Shutdown()
	<-done
}

func TestApproveRejectRuntime(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	// Add an approved runtime to use for the test below
	approvedRuntime, approvedRuntimePrvKey, err := utils.CreateTestRuntimeWithKey(env.colonyID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(approvedRuntime, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(approvedRuntime.ID, env.colonyPrvKey)
	assert.Nil(t, err)

	testRuntime, _, err := utils.CreateTestRuntimeWithKey(env.colonyID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(testRuntime, env.colonyPrvKey)
	assert.Nil(t, err)

	runtimeFromServer, err := client.GetRuntime(testRuntime.ID, approvedRuntimePrvKey)
	assert.Nil(t, err)
	assert.False(t, runtimeFromServer.IsApproved())

	err = client.ApproveRuntime(testRuntime.ID, env.colonyPrvKey)
	assert.Nil(t, err)

	runtimeFromServer, err = client.GetRuntime(testRuntime.ID, approvedRuntimePrvKey)
	assert.Nil(t, err)
	assert.True(t, runtimeFromServer.IsApproved())

	err = client.RejectRuntime(testRuntime.ID, env.colonyPrvKey)
	assert.Nil(t, err)

	runtimeFromServer, err = client.GetRuntime(testRuntime.ID, approvedRuntimePrvKey)
	assert.Nil(t, err)
	assert.False(t, runtimeFromServer.IsApproved())

	server.Shutdown()
	<-done
}

func TestDeleteRuntime(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	runtime, runtimePrvKey, err := utils.CreateTestRuntimeWithKey(env.colonyID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(runtime, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(runtime.ID, env.colonyPrvKey)
	assert.Nil(t, err)

	// Try to get it
	runtimeFromServer, err := client.GetRuntime(runtime.ID, runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, runtimeFromServer)
	assert.True(t, runtime.ID == runtimeFromServer.ID)

	// Now delete it
	err = client.DeleteRuntime(runtime.ID, env.colonyPrvKey)
	assert.Nil(t, err)

	// Try to get it again, it should be gone
	runtimeFromServer, err = client.GetRuntime(runtime.ID, runtimePrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, runtimeFromServer)

	server.Shutdown()
	<-done
}
