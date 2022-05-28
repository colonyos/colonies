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

func createProcess() *Process {
	colonyID := GenerateRandomID()
	runtimeType := "test_runtime_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	processSpec := CreateProcessSpec("test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string))
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

	graph, err := CreateProcessGraph(mock, process1.ID)
	assert.Nil(t, err)

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

	graph, err := CreateProcessGraph(mock, process1.ID)
	assert.Nil(t, err)

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

	graph, err := CreateProcessGraph(mock, process1.ID)
	assert.Nil(t, err)

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

	graph, err := CreateProcessGraph(mock, process1.ID)
	assert.Nil(t, err)

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

	graph, err := CreateProcessGraph(mock, process1.ID)
	assert.Nil(t, err)

	visited := make(map[string]bool)
	err = graph.Iterate(func(process *Process) error {
		visited[process.ID] = true
		return nil
	})
	assert.Nil(t, err)
	assert.Len(t, visited, 1)
	assert.True(t, visited[process1.ID])
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

	process2.WaitingForParents = true
	process3.WaitingForParents = true
	process4.WaitingForParents = true

	graph, err := CreateProcessGraph(mock, process1.ID)
	assert.Nil(t, err)

	waiting, err := graph.WaitingProcesses()
	assert.Nil(t, err)
	assert.True(t, waiting == 4)

	waitingForParents, err := graph.WaitingForParents()
	assert.Nil(t, err)
	assert.True(t, waitingForParents == 3)

	// Now, process1 finishes
	process1.State = SUCCESS
	err = graph.Resolve()
	assert.Nil(t, err)

	// Now only process4 should wait for parents
	waitingForParents, err = graph.WaitingForParents()
	assert.Nil(t, err)
	assert.True(t, waitingForParents == 1)

	// Now process 2 finishes
	process2.State = SUCCESS
	err = graph.Resolve()
	assert.Nil(t, err)

	// Process 4 still have to wait for process 3
	waitingForParents, err = graph.WaitingForParents()
	assert.Nil(t, err)
	assert.True(t, waitingForParents == 1)

	// Now process 3 finishes
	process3.State = SUCCESS
	err = graph.Resolve()
	assert.Nil(t, err)

	// Process 4 can now run
	waitingForParents, err = graph.WaitingForParents()
	assert.Nil(t, err)
	assert.True(t, waitingForParents == 0)

	// Restore states
	process1.State = WAITING
	process2.State = FAILED
	process3.State = WAITING
	process4.State = WAITING

	process2.WaitingForParents = true
	process3.WaitingForParents = true
	process4.WaitingForParents = true

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

	process2.WaitingForParents = true
	process3.WaitingForParents = true
	process4.WaitingForParents = true

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

	process2.WaitingForParents = true
	process3.WaitingForParents = true
	process4.WaitingForParents = true

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

	process2.WaitingForParents = true
	process3.WaitingForParents = true
	process4.WaitingForParents = true

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

	process2.WaitingForParents = false
	process3.WaitingForParents = false
	process4.WaitingForParents = true

	err = graph.Resolve()
	assert.Nil(t, err)

	// All process in graph failes as process 2 failed
	processes, err := graph.Processes()
	assert.Nil(t, err)
	assert.True(t, processes == 4)

	waitingProcesses, err := graph.WaitingProcesses()
	assert.Nil(t, err)
	assert.True(t, waitingProcesses == 1)

	waitingForParents, err = graph.WaitingForParents()
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
}
