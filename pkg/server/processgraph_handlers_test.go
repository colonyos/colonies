package server

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
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

	wf := generateDiamondtWorkflowSpec(env.colonyID)
	submittedGraph, err := client.SubmitWorkflowSpec(wf, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	graphs, err := client.GetWaitingProcessGraphs(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	processes, err := client.GetWaitingProcesses(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processes, 4)

	assignedProcess1, err := client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess1.ProcessSpec.Name == "task1")

	// We cannot be assigned more tasks until task1 is closed
	_, err = client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.NotNil(t, err) // Note error

	graphs, err = client.GetRunningProcessGraphs(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	// Close task1
	err = client.Close(assignedProcess1.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcess2, err := client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess2.ProcessSpec.Name == "task2" || assignedProcess2.ProcessSpec.Name == "task3")

	assignedProcess3, err := client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess3.ProcessSpec.Name == "task2" || assignedProcess3.ProcessSpec.Name == "task3")

	// We cannot be assigned more tasks (task4 is left) until task2 and task3 finish
	_, err = client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.NotNil(t, err) // Note error

	// Close task2
	err = client.Close(assignedProcess2.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	// Close task3
	err = client.Close(assignedProcess3.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	// Now it should be possible to assign task4 to a worker
	assignedProcess4, err := client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess4.ProcessSpec.Name == "task4")

	// Close task4
	err = client.Close(assignedProcess4.ID, env.runtimePrvKey)
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

func TestProcessGraphFailed(t *testing.T) {
	// This is workflow we are going to test. Task2 and Task3 cannot be assigned before Task1 is closed as successful.
	// Task4 cannot be assigned until both Task2 and Task3 is closed as successful.
	//
	//         task1
	//          / \
	//     task2   task3
	//          \ /
	//         task4

	env, client, server, _, done := setupTestEnv2(t)

	wf := generateDiamondtWorkflowSpec(env.colonyID)
	submittedGraph, err := client.SubmitWorkflowSpec(wf, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.Nil(t, err)

	// Close task1
	err = client.Fail(assignedProcess.ID, []string{}, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.NotNil(t, err) // Error, all processes in the entire graph will fail, i.e no processes can be selected for runtime with Id

	processGraph, err := client.GetProcessGraph(submittedGraph.ID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, processGraph.State, core.FAILED)

	server.Shutdown()
	<-done
}

func TestAddChild(t *testing.T) {
	//         task1
	//          / \
	//     task2   task3

	env, client, server, _, done := setupTestEnv2(t)

	wf := generateTreeWorkflowSpec(env.colonyID)
	submittedGraph, err := client.SubmitWorkflowSpec(wf, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	assignedProcess, err := client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess.ProcessSpec.Name == "task1")

	// Add task5 to task1
	childProcessSpec := utils.CreateTestProcessSpec(env.colonyID)
	childProcessSpec.Name = "task5"
	_, err = client.AddChild(assignedProcess.ProcessGraphID, assignedProcess.ID, childProcessSpec, env.runtimePrvKey)
	assert.Nil(t, err)
	err = client.Close(assignedProcess.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	var names []string
	assignedProcess, err = client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.Nil(t, err)
	names = append(names, assignedProcess.ProcessSpec.Name)
	err = client.Close(assignedProcess.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.Nil(t, err)
	names = append(names, assignedProcess.ProcessSpec.Name)
	err = client.Close(assignedProcess.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.Nil(t, err)
	names = append(names, assignedProcess.ProcessSpec.Name)
	err = client.Close(assignedProcess.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	counter := 0
	for _, name := range names {
		if name == "task2" {
			counter++
		}
		if name == "task3" {
			counter++
		}
		if name == "task5" {
			counter++
		}
	}

	assert.Len(t, names, 3)
	assert.Equal(t, counter, 3)

	server.Shutdown()
	<-done
}

func TestAddChildMaxWaitBug(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	wf := generateSingleWorkflowSpec(env.colonyID)
	submittedGraph, err := client.SubmitWorkflowSpec(wf, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	assignedProcess, err := client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.Nil(t, err)

	processGraph, err := client.GetProcessGraph(submittedGraph.ID, env.runtimePrvKey)

	// Add task2 to task1
	childProcessSpec := utils.CreateTestProcessSpec(env.colonyID)
	childProcessSpec.MaxWaitTime = 1
	childProcessSpec.Name = "task2"
	_, err = client.AddChild(assignedProcess.ProcessGraphID, assignedProcess.ID, childProcessSpec, env.runtimePrvKey)
	assert.Nil(t, err)

	// Add task3 to task1
	childProcessSpec = utils.CreateTestProcessSpec(env.colonyID)
	childProcessSpec.MaxWaitTime = 1
	childProcessSpec.Name = "task3"
	_, err = client.AddChild(assignedProcess.ProcessGraphID, assignedProcess.ID, childProcessSpec, env.runtimePrvKey)
	assert.Nil(t, err)

	processGraph, err = client.GetProcessGraph(submittedGraph.ID, env.runtimePrvKey)

	err = client.Close(assignedProcess.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	// Wait the task2 and task3 to timeout
	time.Sleep(5 * time.Second)

	processGraph, err = client.GetProcessGraph(submittedGraph.ID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, processGraph.State, core.FAILED)

	server.Shutdown()
	<-done
}

func TestSubmitWorkflowSpecWithInputOutput(t *testing.T) {
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

	assignedProcess1, err := client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess1.ProcessSpec.Name == "task1")

	// Close task1
	err = client.CloseWithOutput(assignedProcess1.ID, []string{"output_task1"}, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcess2, err := client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, assignedProcess2.Input, 1)
	assert.Equal(t, assignedProcess2.Input[0], "output_task1")

	assignedProcess3, err := client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, assignedProcess3.Input, 1)
	assert.Equal(t, assignedProcess3.Input[0], "output_task1")

	// Close task2
	err = client.CloseWithOutput(assignedProcess2.ID, []string{"output_task2"}, env.runtimePrvKey)
	assert.Nil(t, err)

	// Close task3
	err = client.CloseWithOutput(assignedProcess3.ID, []string{"output_task3"}, env.runtimePrvKey)
	assert.Nil(t, err)

	// Now it should be possible to assign task4 to a worker
	assignedProcess4, err := client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, assignedProcess4.Input, 2)

	ok := false
	if assignedProcess4.Input[0] == "output_task2" && assignedProcess4.Input[1] == "output_task3" {
		ok = true
	} else if assignedProcess4.Input[0] == "output_task3" && assignedProcess4.Input[1] == "output_task2" {
		ok = true
	}
	assert.True(t, ok)

	// Close task4
	err = client.Close(assignedProcess4.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestSubmitWorkflowSpecFailed(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	diamond := generateDiamondtWorkflowSpec(env.colonyID)
	submittedGraph, err := client.SubmitWorkflowSpec(diamond, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	assignedProcess1, err := client.AssignProcess(env.colonyID, -1, env.runtimePrvKey)
	assert.Nil(t, err)
	err = client.Fail(assignedProcess1.ID, []string{"error"}, env.runtimePrvKey)
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
