package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubmitWorkflowSpec(t *testing.T) {
	// This is workflow we are going to test. Task2 and Task3 cannot be assigned before Task1 is closed as successful.
	// Task4 cannot be assigned until both Task2 and Task3 is closed as successful.
	//
	//         task1
	//          / \
	//     task2   task3
	//          \ /
	//         task4

	env, client, server, _, done := setupTestEnv2(t)

	diamond := generateDiamondtWorkflowSpec(env.colonyID)
	submittedGraph, err := client.SubmitWorkflowSpec(diamond, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	graphs, err := client.GetWaitingProcessGraphs(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	processes, err := client.GetWaitingProcesses(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processes, 4)

	assignedProcess1, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess1.ProcessSpec.Name == "task1")

	// We cannot be assigned more tasks until task1 is closed
	_, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.NotNil(t, err) // Note error

	graphs, err = client.GetRunningProcessGraphs(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	// Close task1
	err = client.CloseSuccessful(assignedProcess1.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcess2, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess2.ProcessSpec.Name == "task2" || assignedProcess2.ProcessSpec.Name == "task3")

	assignedProcess3, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess3.ProcessSpec.Name == "task2" || assignedProcess3.ProcessSpec.Name == "task3")

	// We cannot be assigned more tasks (task4 is left) until task2 and task3 finish
	_, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.NotNil(t, err) // Note error

	// Close task2
	err = client.CloseSuccessful(assignedProcess2.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	// Close task3
	err = client.CloseSuccessful(assignedProcess3.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	// Now it should be possible to assign task4 to a worker
	assignedProcess4, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess4.ProcessSpec.Name == "task4")

	// Close task4
	err = client.CloseSuccessful(assignedProcess4.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	graphs, err = client.GetWaitingProcessGraphs(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 0)

	graphs, err = client.GetRunningProcessGraphs(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 0)

	graphs, err = client.GetSuccessfulProcessGraphs(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	graphs, err = client.GetFailedProcessGraphs(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 0)

	server.Shutdown()
	<-done
}

func TestSubmitWorkflowSpecFailed(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	diamond := generateDiamondtWorkflowSpec(env.colonyID)
	submittedGraph, err := client.SubmitWorkflowSpec(diamond, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	assignedProcess1, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	err = client.CloseFailed(assignedProcess1.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	graphs, err := client.GetFailedProcessGraphs(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	server.Shutdown()
	<-done
}

func TestGetProcessGraph(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	diamond := generateDiamondtWorkflowSpec(env.colonyID)
	submittedGraph, err := client.SubmitWorkflowSpec(diamond, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	graphFromServer, err := client.GetProcessGraph(submittedGraph.ID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, submittedGraph.ID == graphFromServer.ID)

	server.Shutdown()
	<-done
}

func TestDeleteProcessGraph(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	diamond := generateDiamondtWorkflowSpec(env.colonyID)
	submittedGraph1, err := client.SubmitWorkflowSpec(diamond, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph1)

	diamond = generateDiamondtWorkflowSpec(env.colonyID)
	submittedGraph2, err := client.SubmitWorkflowSpec(diamond, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph2)

	graphFromServer, err := client.GetProcessGraph(submittedGraph1.ID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, submittedGraph1.ID == graphFromServer.ID)

	graphFromServer, err = client.GetProcessGraph(submittedGraph2.ID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, submittedGraph2.ID == graphFromServer.ID)

	err = client.DeleteProcessGraph(submittedGraph1.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	graphFromServer, err = client.GetProcessGraph(submittedGraph1.ID, env.runtimePrvKey)
	assert.NotNil(t, err)

	graphFromServer, err = client.GetProcessGraph(submittedGraph2.ID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, submittedGraph2.ID == graphFromServer.ID)

	server.Shutdown()
	<-done
}

func TestDeleteAllProcessGraphs(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	diamond := generateDiamondtWorkflowSpec(env.colonyID)
	submittedGraph1, err := client.SubmitWorkflowSpec(diamond, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph1)

	diamond = generateDiamondtWorkflowSpec(env.colonyID)
	submittedGraph2, err := client.SubmitWorkflowSpec(diamond, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph2)

	graphFromServer, err := client.GetProcessGraph(submittedGraph1.ID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, submittedGraph1.ID == graphFromServer.ID)

	graphFromServer, err = client.GetProcessGraph(submittedGraph2.ID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, submittedGraph2.ID == graphFromServer.ID)

	err = client.DeleteAllProcessGraphs(env.colonyID, env.colonyPrvKey)
	assert.Nil(t, err)

	graphFromServer, err = client.GetProcessGraph(submittedGraph1.ID, env.runtimePrvKey)
	assert.NotNil(t, err)

	graphFromServer, err = client.GetProcessGraph(submittedGraph2.ID, env.runtimePrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}
