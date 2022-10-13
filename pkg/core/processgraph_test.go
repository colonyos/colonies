package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type processGraphStorageMock struct {
	processes map[string]*Process
}

func createProcessGraphStorageMock() *processGraphStorageMock {
	mock := &processGraphStorageMock{}
	mock.processes = make(map[string]*Process)
	return mock
}

func (mock *processGraphStorageMock) addProcess(process *Process) {
	mock.processes[process.ID] = process
}

func (mock *processGraphStorageMock) GetProcessByID(processID string) (*Process, error) {
	return mock.processes[processID], nil
}

func (mock *processGraphStorageMock) SetProcessState(processID string, state int) error {
	process := mock.processes[processID]
	process.State = state

	return nil
}

func (mock *processGraphStorageMock) SetWaitForParents(processID string, waitForParents bool) error {
	process := mock.processes[processID]
	process.WaitForParents = waitForParents

	return nil
}

func (mock *processGraphStorageMock) SetProcessGraphState(processGraphID string, state int) error {
	return nil
}

func createProcess() *Process {
	colonyID := GenerateRandomID()
	runtimeType := "test_runtime_type"
	maxWaitTime := -1
	maxExecTime := -1
	maxRetries := 3

	processSpec := CreateProcessSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{}, runtimeType, maxWaitTime, maxExecTime, maxRetries, make(map[string]string), []string{}, 1)
	process := CreateProcess(processSpec)

	return process
}

func TestProcessGraphGetRoot(t *testing.T) {
	process1 := createProcess()
	process2 := createProcess()
	process3 := createProcess()
	process4 := createProcess()

	//        process1
	//          / \
	//  process2   process3
	//          \ /
	//        process4

	process1.AddChild(process2.ID)
	process1.AddChild(process3.ID)
	process2.AddParent(process1.ID)
	process3.AddParent(process1.ID)
	process2.AddChild(process4.ID)
	process3.AddChild(process4.ID)
	process4.AddParent(process2.ID)
	process4.AddParent(process3.ID)

	mock := createProcessGraphStorageMock()
	mock.addProcess(process1)
	mock.addProcess(process2)
	mock.addProcess(process3)
	mock.addProcess(process4)

	colonyID := GenerateRandomID()

	graph, err := CreateProcessGraph(colonyID)
	assert.Nil(t, err)

	graph.storage = mock

	graph.AddRoot(process1.ID)

	root, err := graph.GetRoot(process1.ID)
	assert.Nil(t, err)
	assert.Equal(t, root.ID, process1.ID)

	root, err = graph.GetRoot(process2.ID)
	assert.Nil(t, err)
	assert.Equal(t, root.ID, process1.ID)

	root, err = graph.GetRoot(process3.ID)
	assert.Nil(t, err)
	assert.Equal(t, root.ID, process1.ID)

	root, err = graph.GetRoot(process4.ID)
	assert.Nil(t, err)
	assert.Equal(t, root.ID, process1.ID)
}

func TestProcessGraphGetRootLoop(t *testing.T) {
	process1 := createProcess()
	process2 := createProcess()
	process3 := createProcess()
	process4 := createProcess()

	//        process1------\
	//          / \         |
	//  process2   process3 |
	//          \ /         |
	//        process4 ----/

	process1.AddChild(process2.ID)
	process1.AddChild(process3.ID)
	process2.AddParent(process1.ID)
	process3.AddParent(process1.ID)
	process2.AddChild(process4.ID)
	process3.AddChild(process4.ID)
	process4.AddParent(process2.ID)
	process4.AddParent(process3.ID)

	// Create a loop
	process1.AddParent(process4.ID)

	mock := createProcessGraphStorageMock()
	mock.addProcess(process1)
	mock.addProcess(process2)
	mock.addProcess(process3)
	mock.addProcess(process4)

	colonyID := GenerateRandomID()

	graph, err := CreateProcessGraph(colonyID)
	assert.Nil(t, err)

	graph.storage = mock

	graph.AddRoot(process1.ID)

	_, err = graph.GetRoot(process1.ID)
	assert.NotNil(t, err)

	_, err = graph.GetRoot(process2.ID)
	assert.NotNil(t, err)

	_, err = graph.GetRoot(process3.ID)
	assert.NotNil(t, err)

	_, err = graph.GetRoot(process4.ID)
	assert.NotNil(t, err)
}

func TestProcessGraphIterate(t *testing.T) {
	process1 := createProcess()
	process2 := createProcess()
	process3 := createProcess()
	process4 := createProcess()

	//        process1
	//          / \
	//  process2   process3
	//          \ /
	//        process4

	process1.AddChild(process2.ID)
	process1.AddChild(process3.ID)
	process2.AddParent(process1.ID)
	process3.AddParent(process1.ID)
	process2.AddChild(process4.ID)
	process3.AddChild(process4.ID)
	process4.AddParent(process2.ID)
	process4.AddParent(process3.ID)

	mock := createProcessGraphStorageMock()
	mock.addProcess(process1)
	mock.addProcess(process2)
	mock.addProcess(process3)
	mock.addProcess(process4)

	colonyID := GenerateRandomID()

	graph, err := CreateProcessGraph(colonyID)
	assert.Nil(t, err)

	graph.storage = mock

	graph.AddRoot(process1.ID)

	visited := make(map[string]bool)
	err = graph.Iterate(func(process *Process) error {
		visited[process.ID] = true
		return nil
	})
	assert.Nil(t, err)
	assert.Len(t, visited, 4)
	assert.True(t, visited[process1.ID])
	assert.True(t, visited[process2.ID])
	assert.True(t, visited[process3.ID])
	assert.True(t, visited[process4.ID])
}

func TestProcessGraphIterate2(t *testing.T) {
	process1 := createProcess()
	process2 := createProcess()
	process3 := createProcess()
	process4 := createProcess()

	//        process1
	//          / \
	//  process2   process3
	//          \
	//        process4

	process1.AddChild(process2.ID)
	process1.AddChild(process3.ID)
	process2.AddParent(process1.ID)
	process3.AddParent(process1.ID)
	process2.AddChild(process4.ID)
	process4.AddParent(process2.ID)

	mock := createProcessGraphStorageMock()
	mock.addProcess(process1)
	mock.addProcess(process2)
	mock.addProcess(process3)
	mock.addProcess(process4)

	colonyID := GenerateRandomID()

	graph, err := CreateProcessGraph(colonyID)
	assert.Nil(t, err)

	graph.storage = mock

	graph.AddRoot(process1.ID)

	visited := make(map[string]bool)
	err = graph.Iterate(func(process *Process) error {
		visited[process.ID] = true
		return nil
	})
	assert.Nil(t, err)
	assert.Len(t, visited, 4)
	assert.True(t, visited[process1.ID])
	assert.True(t, visited[process2.ID])
	assert.True(t, visited[process3.ID])
	assert.True(t, visited[process4.ID])
}

func TestProcessGraphIterate3(t *testing.T) {
	process1 := createProcess()

	mock := createProcessGraphStorageMock()
	mock.addProcess(process1)

	colonyID := GenerateRandomID()

	graph, err := CreateProcessGraph(colonyID)
	assert.Nil(t, err)

	graph.storage = mock

	graph.AddRoot(process1.ID)

	visited := make(map[string]bool)
	err = graph.Iterate(func(process *Process) error {
		visited[process.ID] = true
		return nil
	})
	assert.Nil(t, err)
	assert.Len(t, visited, 1)
	assert.True(t, visited[process1.ID])
}

func TestProcessGraphIterateMultipleRoots(t *testing.T) {
	process1 := createProcess()
	process2 := createProcess()
	process3 := createProcess()
	process4 := createProcess()
	process5 := createProcess()
	process6 := createProcess()

	//        process1          process5
	//          / \                 |
	//  process2   process3     process6
	//          \ /
	//        process4

	process1.AddChild(process2.ID)
	process1.AddChild(process3.ID)
	process2.AddParent(process1.ID)
	process3.AddParent(process1.ID)
	process2.AddChild(process4.ID)
	process3.AddChild(process4.ID)
	process4.AddParent(process2.ID)
	process4.AddParent(process3.ID)
	process6.AddParent(process5.ID)
	process5.AddChild(process6.ID)

	mock := createProcessGraphStorageMock()
	mock.addProcess(process1)
	mock.addProcess(process2)
	mock.addProcess(process3)
	mock.addProcess(process4)
	mock.addProcess(process5)
	mock.addProcess(process6)

	colonyID := GenerateRandomID()

	graph, err := CreateProcessGraph(colonyID)
	assert.Nil(t, err)

	graph.storage = mock

	graph.AddRoot(process1.ID)
	graph.AddRoot(process5.ID)

	visited := make(map[string]bool)
	err = graph.Iterate(func(process *Process) error {
		visited[process.ID] = true
		return nil
	})
	assert.Nil(t, err)
	assert.Len(t, visited, 6)
	assert.True(t, visited[process1.ID])
	assert.True(t, visited[process2.ID])
	assert.True(t, visited[process3.ID])
	assert.True(t, visited[process4.ID])
	assert.True(t, visited[process5.ID])
	assert.True(t, visited[process6.ID])
}

func TestProcessGraphResolve(t *testing.T) {
	process1 := createProcess()
	process2 := createProcess()
	process3 := createProcess()
	process4 := createProcess()

	//        process1
	//          / \
	//  process2   process3
	//          \ /
	//        process4

	process1.AddChild(process2.ID)
	process1.AddChild(process3.ID)
	process2.AddParent(process1.ID)
	process3.AddParent(process1.ID)
	process2.AddChild(process4.ID)
	process3.AddChild(process4.ID)
	process4.AddParent(process2.ID)
	process4.AddParent(process3.ID)

	mock := createProcessGraphStorageMock()
	mock.addProcess(process1)
	mock.addProcess(process2)
	mock.addProcess(process3)
	mock.addProcess(process4)

	process1.State = WAITING
	process2.State = WAITING
	process3.State = WAITING
	process4.State = WAITING

	process2.WaitForParents = true
	process3.WaitForParents = true
	process4.WaitForParents = true

	colonyID := GenerateRandomID()

	graph, err := CreateProcessGraph(colonyID)
	assert.Nil(t, err)

	graph.storage = mock

	graph.AddRoot(process1.ID)

	waiting, err := graph.WaitProcesses()
	assert.Nil(t, err)
	assert.True(t, waiting == 4)

	waitingForParents, err := graph.WaitForParents()
	assert.Nil(t, err)
	assert.True(t, waitingForParents == 3)

	// Now, process1 finishes
	process1.State = SUCCESS
	err = graph.Resolve()
	assert.Nil(t, err)

	// Now only process4 should wait for parents
	waitingForParents, err = graph.WaitForParents()
	assert.Nil(t, err)
	assert.True(t, waitingForParents == 1)

	// Now process 2 finishes
	process2.State = SUCCESS
	err = graph.Resolve()
	assert.Nil(t, err)

	// Process 4 still have to wait for process 3
	waitingForParents, err = graph.WaitForParents()
	assert.Nil(t, err)
	assert.True(t, waitingForParents == 1)

	// Now process 3 finishes
	process3.State = SUCCESS
	err = graph.Resolve()
	assert.Nil(t, err)

	// Process 4 can now run
	waitingForParents, err = graph.WaitForParents()
	assert.Nil(t, err)
	assert.True(t, waitingForParents == 0)

	// Restore states
	process1.State = WAITING
	process2.State = FAILED
	process3.State = WAITING
	process4.State = WAITING

	process2.WaitForParents = true
	process3.WaitForParents = true
	process4.WaitForParents = true

	err = graph.Resolve()
	assert.Nil(t, err)

	// All process in graph failes as process 2 failed
	failedProcesses, err := graph.FailedProcesses()
	assert.Nil(t, err)
	assert.True(t, failedProcesses == 4)

	// Restore states
	process1.State = FAILED
	process2.State = WAITING
	process3.State = WAITING
	process4.State = WAITING

	process2.WaitForParents = true
	process3.WaitForParents = true
	process4.WaitForParents = true

	err = graph.Resolve()
	assert.Nil(t, err)

	// All process in graph failes as process 2 failed
	failedProcesses, err = graph.FailedProcesses()
	assert.Nil(t, err)
	assert.True(t, failedProcesses == 4)

	// Restore states
	process1.State = WAITING
	process2.State = WAITING
	process3.State = WAITING
	process4.State = FAILED

	process2.WaitForParents = true
	process3.WaitForParents = true
	process4.WaitForParents = true

	err = graph.Resolve()
	assert.Nil(t, err)
	assert.True(t, failedProcesses == 4)

	// All process in graph failes as process 2 failed
	failedProcesses, err = graph.FailedProcesses()
	assert.Nil(t, err)

	// Restore states
	process1.State = WAITING
	process2.State = WAITING
	process3.State = WAITING
	process4.State = WAITING

	process2.WaitForParents = true
	process3.WaitForParents = true
	process4.WaitForParents = true

	err = graph.Resolve()
	assert.Nil(t, err)

	// All process in graph failes as process 2 failed
	failedProcesses, err = graph.FailedProcesses()
	assert.Nil(t, err)
	assert.True(t, failedProcesses == 0)

	// Restore states
	process1.State = SUCCESS
	process2.State = RUNNING
	process3.State = RUNNING
	process4.State = WAITING

	process2.WaitForParents = false
	process3.WaitForParents = false
	process4.WaitForParents = true

	err = graph.Resolve()
	assert.Nil(t, err)

	// All process in graph failes as process 2 failed
	processes, err := graph.Processes()
	assert.Nil(t, err)
	assert.True(t, processes == 4)

	waitingProcesses, err := graph.WaitProcesses()
	assert.Nil(t, err)
	assert.True(t, waitingProcesses == 1)

	waitingForParents, err = graph.WaitForParents()
	assert.Nil(t, err)
	assert.True(t, waitingForParents == 1)

	runningProcesses, err := graph.RunningProcesses()
	assert.Nil(t, err)
	assert.True(t, runningProcesses == 2)

	successfulProcesses, err := graph.SuccessfulProcesses()
	assert.Nil(t, err)
	assert.True(t, successfulProcesses == 1)

	err = graph.Resolve()
	assert.Nil(t, err)
	assert.Nil(t, err)
	assert.True(t, graph.State == RUNNING)

	process1.State = WAITING
	process2.State = WAITING
	process3.State = WAITING
	process4.State = WAITING

	err = graph.Resolve()
	assert.Nil(t, err)
	assert.Nil(t, err)
	assert.True(t, graph.State == WAITING)

	process1.State = WAITING
	process2.State = FAILED
	process3.State = WAITING
	process4.State = WAITING

	err = graph.Resolve()
	assert.Nil(t, err)
	assert.Nil(t, err)
	assert.True(t, graph.State == FAILED)

	process1.State = SUCCESS
	process2.State = SUCCESS
	process3.State = SUCCESS
	process4.State = SUCCESS

	err = graph.Resolve()
	assert.Nil(t, err)

	assert.Nil(t, err)
	assert.True(t, graph.State == SUCCESS)
}

func TestProcessGraphResolveMultipleRoots(t *testing.T) {
	process1 := createProcess()
	process2 := createProcess()
	process3 := createProcess()
	process4 := createProcess()
	process5 := createProcess()
	process6 := createProcess()

	//        process1          process5
	//          / \                 |
	//  process2   process3     process6
	//          \ /
	//        process4

	process1.AddChild(process2.ID)
	process1.AddChild(process3.ID)
	process2.AddParent(process1.ID)
	process3.AddParent(process1.ID)
	process2.AddChild(process4.ID)
	process3.AddChild(process4.ID)
	process4.AddParent(process2.ID)
	process4.AddParent(process3.ID)
	process6.AddParent(process5.ID)
	process5.AddChild(process6.ID)

	mock := createProcessGraphStorageMock()
	mock.addProcess(process1)
	mock.addProcess(process2)
	mock.addProcess(process3)
	mock.addProcess(process4)
	mock.addProcess(process5)
	mock.addProcess(process6)

	process1.State = WAITING
	process2.State = WAITING
	process3.State = WAITING
	process4.State = WAITING
	process5.State = WAITING
	process6.State = WAITING

	process2.WaitForParents = true
	process3.WaitForParents = true
	process4.WaitForParents = true
	process6.WaitForParents = true

	colonyID := GenerateRandomID()

	graph, err := CreateProcessGraph(colonyID)
	assert.Nil(t, err)

	graph.storage = mock

	graph.AddRoot(process1.ID)
	graph.AddRoot(process5.ID)

	processes, err := graph.Processes()
	assert.Nil(t, err)
	assert.True(t, processes == 6)

	waiting, err := graph.WaitProcesses()
	assert.Nil(t, err)
	assert.True(t, waiting == 6)

	waitingForParents, err := graph.WaitForParents()
	assert.Nil(t, err)
	assert.True(t, waitingForParents == 4)

	// Now, process1 finishes
	process1.State = SUCCESS
	err = graph.Resolve()
	assert.Nil(t, err)

	// Still process4 and process6 should wait for parents
	waitingForParents, err = graph.WaitForParents()
	assert.Nil(t, err)
	assert.True(t, waitingForParents == 2)

	// Now process 2 finishes
	process2.State = SUCCESS
	err = graph.Resolve()
	assert.Nil(t, err)

	// Still process4 and process6 should wait for parents
	waitingForParents, err = graph.WaitForParents()
	assert.Nil(t, err)
	assert.True(t, waitingForParents == 2)

	// Now process 3 finishes
	process3.State = SUCCESS
	err = graph.Resolve()
	assert.Nil(t, err)

	// process 4 can now run, but process 6 is still waiting for process 5
	waitingForParents, err = graph.WaitForParents()
	assert.Nil(t, err)
	assert.True(t, waitingForParents == 1)

	process4.State = SUCCESS
	err = graph.Resolve()
	assert.Nil(t, err)

	// Now process 5 finishes
	process5.State = SUCCESS
	err = graph.Resolve()
	assert.Nil(t, err)

	// process 6 can now run
	waitingForParents, err = graph.WaitForParents()
	assert.Nil(t, err)
	assert.True(t, waitingForParents == 0)

	process6.State = SUCCESS
	err = graph.Resolve()
	assert.Nil(t, err)
	assert.True(t, graph.State == SUCCESS)
}

func TestProcessGraphJSON(t *testing.T) {
	process1 := createProcess()
	process2 := createProcess()
	process3 := createProcess()
	process4 := createProcess()

	//        process1
	//          / \
	//  process2   process3
	//          \ /
	//        process4

	process1.AddChild(process2.ID)
	process1.AddChild(process3.ID)
	process2.AddParent(process1.ID)
	process3.AddParent(process1.ID)
	process2.AddChild(process4.ID)
	process3.AddChild(process4.ID)
	process4.AddParent(process2.ID)
	process4.AddParent(process3.ID)

	mock := createProcessGraphStorageMock()
	mock.addProcess(process1)
	mock.addProcess(process2)
	mock.addProcess(process3)
	mock.addProcess(process4)

	colonyID := GenerateRandomID()

	graph, err := CreateProcessGraph(colonyID)
	assert.Nil(t, err)

	graph.storage = mock

	graph.AddRoot(process1.ID)

	jsonStr, err := graph.ToJSON()
	assert.Nil(t, err)

	graph2, err := ConvertJSONToProcessGraphWithStorage(jsonStr)
	assert.Nil(t, err)
	assert.True(t, graph.Equals(graph2))
	assert.True(t, graph2.ColonyID == colonyID)

	graph2, err = ConvertJSONToProcessGraph(jsonStr)
	assert.Nil(t, err)
	assert.True(t, graph.Equals(graph2))
	assert.True(t, graph2.ColonyID == colonyID)
}

func TestProcessGraphArrayJSON(t *testing.T) {
	var graphs []*ProcessGraph
	colonyID := GenerateRandomID()

	for i := 0; i < 10; i++ {
		process1 := createProcess()
		graph, err := CreateProcessGraph(colonyID)
		assert.Nil(t, err)
		graph.AddRoot(process1.ID)
		graphs = append(graphs, graph)
	}

	jsonStr, err := ConvertProcessGraphArrayToJSON(graphs)
	assert.Nil(t, err)

	graphs2, err := ConvertJSONToProcessGraphArray(jsonStr)
	assert.Nil(t, err)
	assert.True(t, IsProcessGraphArraysEqual(graphs, graphs2))
}

func TestUpdateProcessIDs(t *testing.T) {
	process1 := createProcess()
	process2 := createProcess()
	process3 := createProcess()

	//  process1
	//     |
	//  process2
	//     |
	//  process3

	process1.AddChild(process2.ID)
	process2.AddParent(process1.ID)
	process2.AddChild(process3.ID)
	process3.AddParent(process2.ID)

	mock := createProcessGraphStorageMock()
	mock.addProcess(process1)
	mock.addProcess(process2)
	mock.addProcess(process3)

	colonyID := GenerateRandomID()
	graph, err := CreateProcessGraph(colonyID)
	assert.Nil(t, err)

	graph.storage = mock
	graph.AddRoot(process1.ID)

	assert.Len(t, graph.ProcessIDs, 0)

	err = graph.UpdateProcessIDs()
	assert.Nil(t, err)
	assert.Len(t, graph.ProcessIDs, 3)
}

func TestProcessGraphGetLeaves(t *testing.T) {
	process1 := createProcess()
	process2 := createProcess()
	process3 := createProcess()
	process4 := createProcess()

	//        process1
	//          / \
	//  process2   process3
	//          \ /
	//        process4

	process1.AddChild(process2.ID)
	process1.AddChild(process3.ID)
	process2.AddParent(process1.ID)
	process3.AddParent(process1.ID)
	process2.AddChild(process4.ID)
	process3.AddChild(process4.ID)
	process4.AddParent(process2.ID)
	process4.AddParent(process3.ID)

	mock := createProcessGraphStorageMock()
	mock.addProcess(process1)
	mock.addProcess(process2)
	mock.addProcess(process3)
	mock.addProcess(process4)

	colonyID := GenerateRandomID()

	graph, err := CreateProcessGraph(colonyID)
	assert.Nil(t, err)

	graph.storage = mock

	graph.AddRoot(process1.ID)

	leaves, err := graph.Leaves()
	assert.Nil(t, err)
	assert.Len(t, leaves, 1)
	assert.Equal(t, leaves[0], process4.ID)
}

func TestProcessGraphGetLeaves2(t *testing.T) {
	process1 := createProcess()
	process2 := createProcess()
	process3 := createProcess()
	process4 := createProcess()
	process5 := createProcess()

	//        process1
	//          / \
	//  process2   process3
	//     |           |
	//  process4   process5

	process1.AddChild(process2.ID)
	process1.AddChild(process3.ID)
	process2.AddParent(process1.ID)
	process3.AddParent(process1.ID)
	process2.AddChild(process4.ID)
	process3.AddChild(process5.ID)
	process4.AddParent(process2.ID)
	process5.AddParent(process3.ID)

	mock := createProcessGraphStorageMock()
	mock.addProcess(process1)
	mock.addProcess(process2)
	mock.addProcess(process3)
	mock.addProcess(process4)
	mock.addProcess(process5)

	colonyID := GenerateRandomID()

	graph, err := CreateProcessGraph(colonyID)
	assert.Nil(t, err)

	graph.storage = mock

	graph.AddRoot(process1.ID)

	leaves, err := graph.Leaves()
	assert.Nil(t, err)
	assert.Len(t, leaves, 2)
	assert.Equal(t, leaves[0], process4.ID)
	assert.Equal(t, leaves[1], process5.ID)
}
