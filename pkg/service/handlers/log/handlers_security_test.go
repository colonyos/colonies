package log_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/service"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddLogSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	output := make([]interface{}, 2)
	output[0] = "result1"

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	_, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)
	processFromServer, err := client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.AddLog(processFromServer.ID, "test_msg", env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.AddLog(processFromServer.ID, "test_msg", env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.AddLog(processFromServer.ID, "test_msg", env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.AddLog(processFromServer.ID, "test_msg", env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetLogsSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	_, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)
	processFromServer, err := client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.AddLog(processFromServer.ID, "test_msg", env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetLogsByProcess(env.Colony2Name, processFromServer.ID, 100, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetLogsByProcess(env.Colony1Name, processFromServer.ID, 100, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetLogsByProcess(env.Colony2Name, processFromServer.ID, 100, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetLogsByProcess(env.Colony1Name, processFromServer.ID, 100, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestSearchLogsSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	_, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)
	processFromServer, err := client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.AddLog(processFromServer.ID, "test_msg", env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.SearchLogs(env.Colony1Name, "test_msg", 3, 10, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.SearchLogs(env.Colony1Name, "test_msg", 3, 10, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.SearchLogs(env.Colony1Name, "test_msg", 10, 10, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.SearchLogs(env.Colony1Name, "test_msg", 10, 10, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}
