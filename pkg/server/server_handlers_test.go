package server

import (
	"testing"

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
	env, client, server, serverPrvKey, done := setupTestEnv2(t)

	// Waiting
	numberOfWaitingProcesses := 5
	for i := 0; i < numberOfWaitingProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	// Running
	numberOfRunningProcesses := 6
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		_, err = client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	}

	// Successful
	numberOfSuccessfulProcesses := 7
	for i := 0; i < numberOfSuccessfulProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
		assert.Nil(t, err)
		err = client.CloseSuccessful(processFromServer.ID, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	// Failed
	numberOfFailedProcesses := 8
	for i := 0; i < numberOfFailedProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
		assert.Nil(t, err)
		err = client.CloseFailed(processFromServer.ID, "error", env.runtimePrvKey)
		assert.Nil(t, err)
	}

	stat, err := client.Statistics(serverPrvKey)
	assert.Nil(t, err)

	assert.Equal(t, stat.WaitingProcesses, numberOfWaitingProcesses)
	assert.Equal(t, stat.RunningProcesses, numberOfRunningProcesses)
	assert.Equal(t, stat.SuccessfulProcesses, numberOfSuccessfulProcesses)
	assert.Equal(t, stat.FailedProcesses, numberOfFailedProcesses)

	server.Shutdown()
	<-done
}

func TestGetClusterInfo(t *testing.T) {
	_, client, server, serverPrvKey, done := setupTestEnv2(t)

	cluster, err := client.GetClusterInfo(serverPrvKey)
	assert.Nil(t, err)
	assert.Len(t, cluster.Nodes, 1)
	assert.Equal(t, cluster.Nodes[0], cluster.Leader) // Since we only have one EtcdServer

	server.Shutdown()
	<-done
}

func TestCheckHealth(t *testing.T) {
	_, client, server, _, done := setupTestEnv2(t)

	assert.Nil(t, client.CheckHealth())

	server.Shutdown()
	<-done
}
