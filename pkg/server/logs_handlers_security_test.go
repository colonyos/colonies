package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddLogSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	output := make([]interface{}, 2)
	output[0] = "result1"

	funcSpec := utils.CreateTestFunctionSpec(env.colony1Name)
	_, err := client.Submit(funcSpec, env.executor1PrvKey)
	assert.Nil(t, err)
	processFromServer, err := client.Assign(env.colony1Name, -1, env.executor1PrvKey)
	assert.Nil(t, err)

	err = client.AddLog(processFromServer.ID, "test_msg", env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.AddLog(processFromServer.ID, "test_msg", env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.AddLog(processFromServer.ID, "test_msg", env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.AddLog(processFromServer.ID, "test_msg", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetLogsSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	funcSpec := utils.CreateTestFunctionSpec(env.colony1Name)
	_, err := client.Submit(funcSpec, env.executor1PrvKey)
	assert.Nil(t, err)
	processFromServer, err := client.Assign(env.colony1Name, -1, env.executor1PrvKey)
	assert.Nil(t, err)

	err = client.AddLog(processFromServer.ID, "test_msg", env.executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetLogsByProcessID(processFromServer.ID, 100, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetLogsByProcessID(processFromServer.ID, 100, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetLogsByProcessID(processFromServer.ID, 100, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetLogsByProcessID(processFromServer.ID, 100, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}
