package server_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// TODO: sometimes we get these errors:
//
// server_handlers_test.go:57:
//     	Error Trace:	server_handlers_test.go:57
//     	Error:      	Not equal:
//     	            	expected: 6
//     	            	actual  : 5
//     	Test:       	TestGetStatistics
// server_handlers_test.go:59:
//     	Error Trace:	server_handlers_test.go:59
//     	Error:      	Not equal:
//     	            	expected: 6
//     	            	actual  : 7
//     	Test:       	TestGetStatistics

func TestGetStatistics(t *testing.T) {
	env, client, coloniesServer, serverPrvKey, done := server.SetupTestEnv2(t)

	// Waiting
	numberOfWaitingProcesses := 5
	for i := 0; i < numberOfWaitingProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
		_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
	}

	// Running
	numberOfRunningProcesses := 6
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
		_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
		_, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
		assert.Nil(t, err)
	}

	// Successful
	numberOfSuccessfulProcesses := 7
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
	numberOfFailedProcesses := 8
	for i := 0; i < numberOfFailedProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
		_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
		assert.Nil(t, err)
		err = client.Fail(processFromServer.ID, []string{"error"}, env.ExecutorPrvKey)
		assert.Nil(t, err)
	}

	stat, err := client.Statistics(serverPrvKey)
	assert.Nil(t, err)

	assert.Equal(t, stat.WaitingProcesses, numberOfWaitingProcesses)
	assert.Equal(t, stat.RunningProcesses, numberOfRunningProcesses)
	assert.Equal(t, stat.SuccessfulProcesses, numberOfSuccessfulProcesses)
	assert.Equal(t, stat.FailedProcesses, numberOfFailedProcesses)

	coloniesServer.Shutdown()
	<-done
}

func TestGetClusterInfo(t *testing.T) {
	_, client, coloniesServer, serverPrvKey, done := server.SetupTestEnv2(t)

	cluster, err := client.GetClusterInfo(serverPrvKey)
	assert.Nil(t, err)
	assert.Len(t, cluster.Nodes, 1)
	assert.Equal(t, cluster.Nodes[0], cluster.Leader) // Since we only have one EtcdServer

	coloniesServer.Shutdown()
	<-done
}

func TestCheckHealth(t *testing.T) {
	_, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	assert.Nil(t, client.CheckHealth())

	coloniesServer.Shutdown()
	<-done
}

// func TestResetDatabase(t *testing.T) {
// 	_, client, coloniesServer, serverPrvKey, done := server.SetupTestEnv2(t)
//
// 	colonies, err := client.GetColonies(serverPrvKey)
// 	assert.Nil(t, err)
// 	assert.Len(t, colonies, 1)
//
// 	err = client.ResetDatabase(serverPrvKey)
// 	assert.Nil(t, err)
//
// 	colonies, err = client.GetColonies(serverPrvKey)
// 	assert.Nil(t, err)
// 	assert.Len(t, colonies, 0)
//
// 	coloniesServer.Shutdown()
// 	<-done
// }
