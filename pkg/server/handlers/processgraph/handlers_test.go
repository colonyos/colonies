package processgraph_test

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
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

	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	wf := server.GenerateDiamondtWorkflowSpec(env.ColonyName)
	submittedGraph, err := client.SubmitWorkflowSpec(wf, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	graphs, err := client.GetWaitingProcessGraphs(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	processes, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processes, 4)

	// It should not be possible to remove processes part of a workflow
	err = client.RemoveProcess(processes[0].ID, env.ExecutorPrvKey)
	assert.NotNil(t, err) // Error

	assignedProcess1, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess1.FunctionSpec.NodeName == "task1")

	// We cannot be assigned more tasks until task1 is closed
	_, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.NotNil(t, err) // Note error

	graphs, err = client.GetRunningProcessGraphs(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	// Close task1
	err = client.Close(assignedProcess1.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess2, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess2.FunctionSpec.NodeName == "task2" || assignedProcess2.FunctionSpec.NodeName == "task3")

	assignedProcess3, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess3.FunctionSpec.NodeName == "task2" || assignedProcess3.FunctionSpec.NodeName == "task3")

	// We cannot be assigned more tasks (task4 is left) until task2 and task3 finish
	_, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.NotNil(t, err) // Note error

	// Close task2
	err = client.Close(assignedProcess2.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Close task3
	err = client.Close(assignedProcess3.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Now it should be possible to assign task4 to an executor
	assignedProcess4, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess4.FunctionSpec.NodeName == "task4")

	// Close task4
	err = client.Close(assignedProcess4.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	graphs, err = client.GetWaitingProcessGraphs(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 0)

	graphs, err = client.GetRunningProcessGraphs(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 0)

	graphs, err = client.GetSuccessfulProcessGraphs(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	graphs, err = client.GetFailedProcessGraphs(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 0)

	coloniesServer.Shutdown()
	<-done
}

func TestSubmitInvalidWorkflowSpec(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	wf := core.CreateWorkflowSpec(env.ColonyName)

	funcSpec1 := core.CreateEmptyFunctionSpec()
	//funcSpec1.NodeName = "task1"
	funcSpec1.Conditions.ColonyName = env.ColonyName
	funcSpec1.Conditions.ExecutorType = "test_executor_type"

	funcSpec2 := core.CreateEmptyFunctionSpec()
	//funcSpec2.NodeName = "task2"
	funcSpec2.Conditions.ColonyName = env.ColonyName
	funcSpec2.Conditions.ExecutorType = "test_executor_type"

	funcSpec3 := core.CreateEmptyFunctionSpec()
	//funcSpec3.NodeName = "task3"
	funcSpec3.Conditions.ColonyName = env.ColonyName
	funcSpec3.Conditions.ExecutorType = "test_executor_type"

	funcSpec4 := core.CreateEmptyFunctionSpec()
	//funcSpec4.NodeName = "task4"
	funcSpec4.Conditions.ColonyName = env.ColonyName
	funcSpec4.Conditions.ExecutorType = "test_executor_type"

	funcSpec2.AddDependency("task1")
	funcSpec3.AddDependency("task1")
	funcSpec4.AddDependency("task2")
	funcSpec4.AddDependency("task3")

	wf.AddFunctionSpec(funcSpec1)
	wf.AddFunctionSpec(funcSpec2)
	wf.AddFunctionSpec(funcSpec3)
	wf.AddFunctionSpec(funcSpec4)

	_, err := client.SubmitWorkflowSpec(wf, env.ExecutorPrvKey)
	assert.NotNil(t, err) // Error: nodename must be specified

	graphs, err := client.GetWaitingProcessGraphs(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 0)

	coloniesServer.Shutdown()
	<-done
}

func TestSubmitInvalidWorkflowSpec2(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	wf := core.CreateWorkflowSpec(env.ColonyName)

	funcSpec1 := core.CreateEmptyFunctionSpec()
	funcSpec1.NodeName = "task1"
	funcSpec1.Conditions.ColonyName = env.ColonyName
	funcSpec1.Conditions.ExecutorType = "test_executor_type"

	funcSpec2 := core.CreateEmptyFunctionSpec()
	funcSpec2.NodeName = "task2"
	funcSpec2.Conditions.ColonyName = env.ColonyName
	funcSpec2.Conditions.ExecutorType = "test_executor_type"

	funcSpec3 := core.CreateEmptyFunctionSpec()
	funcSpec3.NodeName = "task3"
	funcSpec3.Conditions.ColonyName = env.ColonyName
	funcSpec3.Conditions.ExecutorType = "test_executor_type"

	funcSpec4 := core.CreateEmptyFunctionSpec()
	funcSpec4.NodeName = "task3" // Duplicate nodename
	funcSpec4.Conditions.ColonyName = env.ColonyName
	funcSpec4.Conditions.ExecutorType = "test_executor_type"

	funcSpec2.AddDependency("task1")
	funcSpec3.AddDependency("task1")
	funcSpec4.AddDependency("task2")

	wf.AddFunctionSpec(funcSpec1)
	wf.AddFunctionSpec(funcSpec2)
	wf.AddFunctionSpec(funcSpec3)
	wf.AddFunctionSpec(funcSpec4)

	_, err := client.SubmitWorkflowSpec(wf, env.ExecutorPrvKey)
	assert.NotNil(t, err) // Error: nodename must be specified

	graphs, err := client.GetWaitingProcessGraphs(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 0)

	coloniesServer.Shutdown()
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

	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	wf := server.GenerateDiamondtWorkflowSpec(env.ColonyName)
	submittedGraph, err := client.SubmitWorkflowSpec(wf, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Close task1
	err = client.Fail(assignedProcess.ID, []string{}, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.NotNil(t, err) // Error, all processes in the entire graph will fail, i.e no processes can be selected for executor with Id

	processGraph, err := client.GetProcessGraph(submittedGraph.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, processGraph.State, core.FAILED)

	coloniesServer.Shutdown()
	<-done
}

func TestAddChild(t *testing.T) {
	//         task1
	//          / \
	//     task2   task3

	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	wf := server.GenerateTreeWorkflowSpec(env.ColonyName)
	submittedGraph, err := client.SubmitWorkflowSpec(wf, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess.FunctionSpec.NodeName == "task1")

	// Add task5 to task1
	childFunctionSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	childFunctionSpec.NodeName = "task5"
	_, err = client.AddChild(assignedProcess.ProcessGraphID, assignedProcess.ID, "", childFunctionSpec, false, env.ExecutorPrvKey)
	assert.Nil(t, err)
	err = client.Close(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	var names []string
	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	names = append(names, assignedProcess.FunctionSpec.NodeName)
	err = client.Close(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	names = append(names, assignedProcess.FunctionSpec.NodeName)
	err = client.Close(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	names = append(names, assignedProcess.FunctionSpec.NodeName)
	err = client.Close(assignedProcess.ID, env.ExecutorPrvKey)
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

	coloniesServer.Shutdown()
	<-done
}

func TestInsertChild(t *testing.T) {
	//         task1
	//          / \
	//     task2   task3
	//
	// Will become:
	//
	//         task1
	//           |
	//         task4
	//          / \
	//     task2   task3

	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	wf := server.GenerateTreeWorkflowSpec(env.ColonyName)
	submittedGraph, err := client.SubmitWorkflowSpec(wf, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	// We must be assigned to a process in order to insert a child in processgraph
	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess.FunctionSpec.NodeName == "task1")

	childFunctionSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	childFunctionSpec.NodeName = "task4"
	process, err := client.AddChild(assignedProcess.ProcessGraphID, assignedProcess.ID, "", childFunctionSpec, true, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, process.Parents, 1)

	parentProcess, err := client.GetProcess(process.Parents[0], env.ExecutorPrvKey)
	assert.Equal(t, parentProcess.FunctionSpec.NodeName, "task1")

	task2Found := false
	task3Found := false
	for _, childID := range process.Children {
		childProcess, err := client.GetProcess(childID, env.ExecutorPrvKey)
		assert.Nil(t, err)
		if childProcess.FunctionSpec.NodeName == "task2" {
			task2Found = true
		}
		if childProcess.FunctionSpec.NodeName == "task3" {
			task3Found = true
		}
	}

	assert.True(t, task2Found)
	assert.True(t, task3Found)

	coloniesServer.Shutdown()
	<-done
}

func TestInsertChild2(t *testing.T) {
	//         task1
	//          / \
	//     task2   task3
	//
	// Will become:
	//
	//             task1
	//          /    |    \
	//     task2   task3  task4

	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	wf := server.GenerateTreeWorkflowSpec(env.ColonyName)
	submittedGraph, err := client.SubmitWorkflowSpec(wf, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	// We must be assigned to a process in order to insert a child in processgraph
	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess.FunctionSpec.NodeName == "task1")

	childFunctionSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	childFunctionSpec.NodeName = "task4"
	process, err := client.AddChild(assignedProcess.ProcessGraphID, assignedProcess.ID, "", childFunctionSpec, false, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, process.Parents, 1)

	parentProcess, err := client.GetProcess(process.Parents[0], env.ExecutorPrvKey)
	assert.Equal(t, parentProcess.FunctionSpec.NodeName, "task1")
	assert.Len(t, process.Children, 0)

	coloniesServer.Shutdown()
	<-done
}

func TestInsertChild3(t *testing.T) {
	//         task1
	//          / \
	//     task2   task3
	//
	// Will become:
	//
	//         task1
	//          / \
	//     task2   task4
	//               |
	//             task3

	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	wf := server.GenerateTreeWorkflowSpec(env.ColonyName)
	submittedGraph, err := client.SubmitWorkflowSpec(wf, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	var task3Process *core.Process
	for _, processID := range submittedGraph.ProcessIDs {
		process, err := client.GetProcess(processID, env.ExecutorPrvKey)
		assert.Nil(t, err)
		if process.FunctionSpec.NodeName == "task3" {
			task3Process = process
		}
	}
	assert.NotNil(t, task3Process)

	// We must be assigned to a process in order to insert a child in processgraph
	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess.FunctionSpec.NodeName == "task1")

	childFunctionSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	childFunctionSpec.NodeName = "task4"
	process, err := client.AddChild(assignedProcess.ProcessGraphID, assignedProcess.ID, task3Process.ID, childFunctionSpec, false, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, process.Parents, 1)

	parentProcess, err := client.GetProcess(process.Parents[0], env.ExecutorPrvKey)
	assert.Equal(t, parentProcess.FunctionSpec.NodeName, "task1")
	assert.Len(t, process.Children, 1)
	assert.Equal(t, process.Children[0], task3Process.ID)

	coloniesServer.Shutdown()
	<-done
}

func TestAddChildMaxWaitBug(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	wf := server.GenerateSingleWorkflowSpec(env.ColonyName)
	submittedGraph, err := client.SubmitWorkflowSpec(wf, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	processGraph, err := client.GetProcessGraph(submittedGraph.ID, env.ExecutorPrvKey)

	// Add task2 to task1
	childFunctionSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	childFunctionSpec.MaxWaitTime = 1
	childFunctionSpec.NodeName = "task2"
	_, err = client.AddChild(assignedProcess.ProcessGraphID, assignedProcess.ID, "", childFunctionSpec, false, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Add task3 to task1
	childFunctionSpec = utils.CreateTestFunctionSpec(env.ColonyName)
	childFunctionSpec.MaxWaitTime = 1
	childFunctionSpec.NodeName = "task3"
	_, err = client.AddChild(assignedProcess.ProcessGraphID, assignedProcess.ID, "", childFunctionSpec, false, env.ExecutorPrvKey)
	assert.Nil(t, err)

	processGraph, err = client.GetProcessGraph(submittedGraph.ID, env.ExecutorPrvKey)

	err = client.Close(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Wait the task2 and task3 to timeout
	time.Sleep(5 * time.Second)

	processGraph, err = client.GetProcessGraph(submittedGraph.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, processGraph.State, core.FAILED)

	coloniesServer.Shutdown()
	<-done
}

func TestSubmitWorkflowSpecWithInputOutput(t *testing.T) {
	//         task1
	//          / \
	//     task2   task3
	//          \ /
	//         task4

	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	diamond := server.GenerateDiamondtWorkflowSpec(env.ColonyName)
	submittedGraph, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	graphs, err := client.GetWaitingProcessGraphs(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	processes, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processes, 4)

	assignedProcess1, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, assignedProcess1.FunctionSpec.NodeName == "task1")

	// Close task1
	output := make([]interface{}, 1)
	output[0] = "output_task1"
	err = client.CloseWithOutput(assignedProcess1.ID, output, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess2, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, assignedProcess2.Input, 1)
	assert.Equal(t, assignedProcess2.Input[0], "output_task1")

	assignedProcess3, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, assignedProcess3.Input, 1)
	assert.Equal(t, assignedProcess3.Input[0], "output_task1")

	// Close task2
	output = make([]interface{}, 1)
	output[0] = "output_task2"
	err = client.CloseWithOutput(assignedProcess2.ID, output, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Close task3
	output = make([]interface{}, 1)
	output[0] = "output_task3"
	err = client.CloseWithOutput(assignedProcess3.ID, output, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Now it should be possible to assign task4 to an executor
	assignedProcess4, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
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
	err = client.Close(assignedProcess4.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestSubmitWorkflowSpecFailed(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	diamond := server.GenerateDiamondtWorkflowSpec(env.ColonyName)
	submittedGraph, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	assignedProcess1, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	err = client.Fail(assignedProcess1.ID, []string{"error"}, env.ExecutorPrvKey)
	assert.Nil(t, err)

	graphs, err := client.GetFailedProcessGraphs(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	coloniesServer.Shutdown()
	<-done
}

func TestGetProcessGraph(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	diamond := server.GenerateDiamondtWorkflowSpec(env.ColonyName)
	submittedGraph, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	graphFromServer, err := client.GetProcessGraph(submittedGraph.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, submittedGraph.ID == graphFromServer.ID)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveProcessGraph(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	diamond := server.GenerateDiamondtWorkflowSpec(env.ColonyName)
	submittedGraph1, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph1)

	diamond = server.GenerateDiamondtWorkflowSpec(env.ColonyName)
	submittedGraph2, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph2)

	graphFromServer, err := client.GetProcessGraph(submittedGraph1.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, submittedGraph1.ID == graphFromServer.ID)

	graphFromServer, err = client.GetProcessGraph(submittedGraph2.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, submittedGraph2.ID == graphFromServer.ID)

	err = client.RemoveProcessGraph(submittedGraph1.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	graphFromServer, err = client.GetProcessGraph(submittedGraph1.ID, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	graphFromServer, err = client.GetProcessGraph(submittedGraph2.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, submittedGraph2.ID == graphFromServer.ID)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveAllProcessGraphs(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	diamond := server.GenerateDiamondtWorkflowSpec(env.ColonyName)
	submittedGraph1, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph1)

	diamond = server.GenerateDiamondtWorkflowSpec(env.ColonyName)
	submittedGraph2, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph2)

	graphFromServer, err := client.GetProcessGraph(submittedGraph1.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, submittedGraph1.ID == graphFromServer.ID)

	graphFromServer, err = client.GetProcessGraph(submittedGraph2.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, submittedGraph2.ID == graphFromServer.ID)

	err = client.RemoveAllProcessGraphs(env.ColonyName, env.ColonyPrvKey)
	assert.Nil(t, err)

	graphFromServer, err = client.GetProcessGraph(submittedGraph1.ID, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	graphFromServer, err = client.GetProcessGraph(submittedGraph2.ID, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveAllProcessGraphsWithStateWaiting(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	diamond := server.GenerateSingleWorkflowSpec(env.ColonyName)
	submittedGraph1, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph1)

	diamond = server.GenerateSingleWorkflowSpec(env.ColonyName)
	submittedGraph2, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph2)

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingWorkflows, 2)
	assert.Equal(t, stat.RunningWorkflows, 0)
	assert.Equal(t, stat.SuccessfulWorkflows, 0)
	assert.Equal(t, stat.FailedWorkflows, 0)

	err = client.RemoveAllProcessGraphsWithState(env.ColonyName, core.PENDING, env.ColonyPrvKey)
	assert.Nil(t, err)

	stat, err = client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingWorkflows, 0)
	assert.Equal(t, stat.RunningWorkflows, 0)
	assert.Equal(t, stat.SuccessfulWorkflows, 0)
	assert.Equal(t, stat.FailedWorkflows, 0)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveAllProcessGraphsWithStateRunning(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	diamond := server.GenerateSingleWorkflowSpec(env.ColonyName)
	submittedGraph1, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph1)

	diamond = server.GenerateSingleWorkflowSpec(env.ColonyName)
	submittedGraph2, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph2)

	_, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingWorkflows, 1)
	assert.Equal(t, stat.RunningWorkflows, 1)
	assert.Equal(t, stat.SuccessfulWorkflows, 0)
	assert.Equal(t, stat.FailedWorkflows, 0)

	err = client.RemoveAllProcessGraphsWithState(env.ColonyName, core.RUNNING, env.ColonyPrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveAllProcessGraphsWithStateSuccessful(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	diamond := server.GenerateSingleWorkflowSpec(env.ColonyName)
	submittedGraph1, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph1)

	diamond = server.GenerateSingleWorkflowSpec(env.ColonyName)
	submittedGraph2, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph2)

	process, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	err = client.Close(process.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingWorkflows, 1)
	assert.Equal(t, stat.RunningWorkflows, 0)
	assert.Equal(t, stat.SuccessfulWorkflows, 1)
	assert.Equal(t, stat.FailedWorkflows, 0)

	err = client.RemoveAllProcessGraphsWithState(env.ColonyName, core.SUCCESS, env.ColonyPrvKey)
	assert.Nil(t, err)

	stat, err = client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingWorkflows, 1)
	assert.Equal(t, stat.RunningWorkflows, 0)
	assert.Equal(t, stat.SuccessfulWorkflows, 0)
	assert.Equal(t, stat.FailedWorkflows, 0)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveAllProcessGraphsWithStateFailed(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	diamond := server.GenerateSingleWorkflowSpec(env.ColonyName)
	submittedGraph1, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph1)

	diamond = server.GenerateSingleWorkflowSpec(env.ColonyName)
	submittedGraph2, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph2)

	process, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	err = client.Fail(process.ID, []string{"error"}, env.ExecutorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingWorkflows, 1)
	assert.Equal(t, stat.RunningWorkflows, 0)
	assert.Equal(t, stat.SuccessfulWorkflows, 0)
	assert.Equal(t, stat.FailedWorkflows, 1)

	err = client.RemoveAllProcessGraphsWithState(env.ColonyName, core.FAILED, env.ColonyPrvKey)
	assert.Nil(t, err)

	stat, err = client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingWorkflows, 1)
	assert.Equal(t, stat.RunningWorkflows, 0)
	assert.Equal(t, stat.SuccessfulWorkflows, 0)
	assert.Equal(t, stat.FailedWorkflows, 0)

	coloniesServer.Shutdown()
	<-done
}

// TestGetProcessGraphNotFound tests getting a process graph that doesn't exist
func TestGetProcessGraphNotFound(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	_, err := client.GetProcessGraph("non_existent_id", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}

// TestGetProcessGraphUnauthorized tests that non-members cannot get process graphs
func TestGetProcessGraphUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Submit workflow to colony1
	wf := core.CreateWorkflowSpec(colony1.Name)
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.NodeName = "task1"
	funcSpec.Conditions.ColonyName = colony1.Name
	funcSpec.Conditions.ExecutorType = executor1.Type
	wf.AddFunctionSpec(funcSpec)

	graph, err := client.SubmitWorkflowSpec(wf, executor1PrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, graph)

	// Try to get process graph from colony2 executor
	_, err = client.GetProcessGraph(graph.ID, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestRemoveProcessGraphNotFound tests removing a process graph that doesn't exist
func TestRemoveProcessGraphNotFound(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	err := client.RemoveProcessGraph("non_existent_id", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}

// TestRemoveProcessGraphUnauthorized tests that non-members cannot remove process graphs
func TestRemoveProcessGraphUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Submit workflow to colony1
	wf := core.CreateWorkflowSpec(colony1.Name)
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.NodeName = "task1"
	funcSpec.Conditions.ColonyName = colony1.Name
	funcSpec.Conditions.ExecutorType = executor1.Type
	wf.AddFunctionSpec(funcSpec)

	graph, err := client.SubmitWorkflowSpec(wf, executor1PrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, graph)

	// Try to remove process graph from colony2 executor
	err = client.RemoveProcessGraph(graph.ID, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestRemoveAllProcessGraphsNotColonyOwner tests that non-owners cannot remove all process graphs
func TestRemoveAllProcessGraphsNotColonyOwner(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	diamond := server.GenerateSingleWorkflowSpec(env.ColonyName)
	_, err := client.SubmitWorkflowSpec(diamond, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to remove all process graphs with executor key (not colony owner)
	err = client.RemoveAllProcessGraphs(env.ColonyName, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}

// TestGetProcessGraphsUnauthorized tests that non-members cannot get process graphs list
func TestGetProcessGraphsUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Submit workflow to colony1
	wf := core.CreateWorkflowSpec(colony1.Name)
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.NodeName = "task1"
	funcSpec.Conditions.ColonyName = colony1.Name
	funcSpec.Conditions.ExecutorType = executor1.Type
	wf.AddFunctionSpec(funcSpec)

	_, err = client.SubmitWorkflowSpec(wf, executor1PrvKey)
	assert.Nil(t, err)

	// Try to get process graphs from colony1 using colony2 executor's key
	_, err = client.GetWaitingProcessGraphs(colony1.Name, 100, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestSubmitWorkflowSpecUnauthorized tests that non-members cannot submit workflows
func TestSubmitWorkflowSpecUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, _, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Try to submit workflow to colony1 using colony2 executor's key
	wf := core.CreateWorkflowSpec(colony1.Name)
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.NodeName = "task1"
	funcSpec.Conditions.ColonyName = colony1.Name
	funcSpec.Conditions.ExecutorType = executor1.Type
	wf.AddFunctionSpec(funcSpec)

	_, err = client.SubmitWorkflowSpec(wf, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestAddChildUnauthorized tests that non-members cannot add children to process graphs
func TestAddChildUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Submit workflow to colony1
	wf := core.CreateWorkflowSpec(colony1.Name)
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.NodeName = "task1"
	funcSpec.Conditions.ColonyName = colony1.Name
	funcSpec.Conditions.ExecutorType = executor1.Type
	wf.AddFunctionSpec(funcSpec)

	graph, err := client.SubmitWorkflowSpec(wf, executor1PrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, graph)

	// Assign the process
	process, err := client.Assign(colony1.Name, -1, "", "", executor1PrvKey)
	assert.Nil(t, err)

	// Try to add child from colony2 executor
	childFunctionSpec := utils.CreateTestFunctionSpec(colony2.Name) // Note: using colony2
	childFunctionSpec.NodeName = "task2"
	_, err = client.AddChild(graph.ID, process.ID, "", childFunctionSpec, false, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestRemoveAllProcessGraphsWithStateUnauthorized tests unauthorized access to RemoveAllProcessGraphsWithState
func TestRemoveAllProcessGraphsWithStateUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	// Submit workflow to colony1
	wf := core.CreateWorkflowSpec(colony1.Name)
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.NodeName = "task1"
	funcSpec.Conditions.ColonyName = colony1.Name
	funcSpec.Conditions.ExecutorType = executor1.Type
	wf.AddFunctionSpec(funcSpec)

	_, err = client.SubmitWorkflowSpec(wf, executor1PrvKey)
	assert.Nil(t, err)

	// Try to remove all process graphs from colony1 using colony2's key
	err = client.RemoveAllProcessGraphsWithState(colony1.Name, core.PENDING, colonyPrvKey2)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestRemoveRunningProcessGraph tests that a running processgraph cannot be removed
func TestRemoveRunningProcessGraph(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	// Create a simple workflow with a single task
	wf := server.GenerateSingleWorkflowSpec(env.ColonyName)
	submittedGraph, err := client.SubmitWorkflowSpec(wf, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, submittedGraph)

	// Verify processgraph is in WAITING state initially
	graphFromServer, err := client.GetProcessGraph(submittedGraph.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, graphFromServer.State)

	// Assign the process (making the processgraph RUNNING)
	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, assignedProcess)

	// Verify processgraph is now RUNNING
	graphFromServer, err = client.GetProcessGraph(submittedGraph.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.RUNNING, graphFromServer.State)

	// Try to remove the running processgraph - should fail
	err = client.RemoveProcessGraph(submittedGraph.ID, env.ExecutorPrvKey)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cannot remove a running processgraph")

	// Close the process (making the processgraph SUCCESSFUL)
	err = client.Close(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify processgraph is now SUCCESSFUL
	graphFromServer, err = client.GetProcessGraph(submittedGraph.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.SUCCESS, graphFromServer.State)

	// Now removing should succeed
	err = client.RemoveProcessGraph(submittedGraph.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}
