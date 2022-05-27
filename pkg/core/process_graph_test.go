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

func TestGraphGetRoot(t *testing.T) {
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

	graph := CreateProcessGraph(mock)
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

func TestGraphGetRootLoop(t *testing.T) {
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

	graph := CreateProcessGraph(mock)
	_, err := graph.GetRoot(process1.ID)
	assert.NotNil(t, err)

	_, err = graph.GetRoot(process2.ID)
	assert.NotNil(t, err)

	_, err = graph.GetRoot(process3.ID)
	assert.NotNil(t, err)

	_, err = graph.GetRoot(process4.ID)
	assert.NotNil(t, err)
}
