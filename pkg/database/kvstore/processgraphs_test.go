package kvstore

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func generateProcessGraph(t *testing.T, db *KVStoreDatabase, colonyName string) *core.ProcessGraph {
	process1 := utils.CreateTestProcess(colonyName)
	process2 := utils.CreateTestProcess(colonyName)
	process3 := utils.CreateTestProcess(colonyName)
	process4 := utils.CreateTestProcess(colonyName)

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

	err := db.AddProcess(process1)
	assert.Nil(t, err)
	err = db.AddProcess(process2)
	assert.Nil(t, err)
	err = db.AddProcess(process3)
	assert.Nil(t, err)
	err = db.AddProcess(process4)
	assert.Nil(t, err)

	graph, err := core.CreateProcessGraph(colonyName)
	assert.Nil(t, err)

	graph.AddRoot(process1.ID)

	return graph
}

func generateProcessGraph2(t *testing.T, db *KVStoreDatabase, colonyName string) (*core.Process, *core.ProcessGraph) {
	graph, err := core.CreateProcessGraph(colonyName)
	assert.Nil(t, err)

	process := utils.CreateTestProcess(colonyName)
	process.ProcessGraphID = graph.ID
	err = db.AddProcess(process)
	assert.Nil(t, err)

	graph.AddRoot(process.ID)

	return process, graph
}

func TestProcessGraphClosedDB(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)

	graph := generateProcessGraph(t, db, "invalid_id")

	db.Close()

	// KVStore operations work even after close (in-memory store)
	err = db.AddProcessGraph(graph)
	assert.Nil(t, err)

	_, err = db.GetProcessGraphByID("invalid_id")
	assert.Nil(t, err) // Returns nil, nil for non-existing

	err = db.SetProcessGraphState("invalid_id", 1)
	assert.NotNil(t, err) // Should error for non-existing

	_, err = db.FindWaitingProcessGraphs("invalid_id", 1)
	assert.Nil(t, err) // Returns empty slice

	_, err = db.FindRunningProcessGraphs("invalid_id", 1)
	assert.Nil(t, err) // Returns empty slice

	_, err = db.FindSuccessfulProcessGraphs("invalid_id", 1)
	assert.Nil(t, err) // Returns empty slice

	_, err = db.FindFailedProcessGraphs("invalid_id", 1)
	assert.Nil(t, err) // Returns empty slice

	err = db.RemoveProcessGraphByID("invalid_id")
	assert.NotNil(t, err) // Should error for non-existing

	err = db.RemoveAllProcessGraphsByColonyName("invalid_name")
	assert.Nil(t, err) // No error when nothing to remove

	err = db.RemoveAllWaitingProcessGraphsByColonyName("invalid_name")
	assert.Nil(t, err)

	err = db.RemoveAllRunningProcessGraphsByColonyName("invalid_name")
	assert.Nil(t, err)

	err = db.RemoveAllSuccessfulProcessGraphsByColonyName("invalid_name")
	assert.Nil(t, err)

	err = db.RemoveAllFailedProcessGraphsByColonyName("invalid_name")
	assert.Nil(t, err)

	_, err = db.CountWaitingProcessGraphs()
	assert.Nil(t, err)

	_, err = db.CountRunningProcessGraphs()
	assert.Nil(t, err)

	_, err = db.CountSuccessfulProcessGraphs()
	assert.Nil(t, err)

	_, err = db.CountFailedProcessGraphs()
	assert.Nil(t, err)

	_, err = db.CountWaitingProcessGraphsByColonyName("invalid_name")
	assert.Nil(t, err)

	_, err = db.CountRunningProcessGraphsByColonyName("invalid_name")
	assert.Nil(t, err)

	_, err = db.CountSuccessfulProcessGraphsByColonyName("invalid_name")
	assert.Nil(t, err)

	_, err = db.CountFailedProcessGraphsByColonyName("invalid_name")
	assert.Nil(t, err)
}

func TestAddProcessGraph(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Create colony
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Test adding nil process graph
	err = db.AddProcessGraph(nil)
	assert.NotNil(t, err)

	// Create and add process graph
	graph := generateProcessGraph(t, db, colony.Name)
	err = db.AddProcessGraph(graph)
	assert.Nil(t, err)

	// Test duplicate process graph
	err = db.AddProcessGraph(graph)
	assert.NotNil(t, err)

	// Verify process graph was added
	retrievedGraph, err := db.GetProcessGraphByID(graph.ID)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedGraph)
	assert.Equal(t, graph.ID, retrievedGraph.ID)
	assert.Equal(t, graph.ColonyName, retrievedGraph.ColonyName)
}

func TestGetProcessGraphByID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Create colony
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Add process graph
	graph := generateProcessGraph(t, db, colony.Name)
	err = db.AddProcessGraph(graph)
	assert.Nil(t, err)

	// Get existing process graph
	retrievedGraph, err := db.GetProcessGraphByID(graph.ID)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedGraph)
	assert.Equal(t, graph.ID, retrievedGraph.ID)

	// Test non-existing process graph
	nonExistingGraph, err := db.GetProcessGraphByID("non_existing_id")
	assert.Nil(t, err)
	assert.Nil(t, nonExistingGraph)

	// Test empty ID
	emptyGraph, err := db.GetProcessGraphByID("")
	assert.Nil(t, err)
	assert.Nil(t, emptyGraph)
}

func TestSetProcessGraphState(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Create colony
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Add process graph
	graph := generateProcessGraph(t, db, colony.Name)
	err = db.AddProcessGraph(graph)
	assert.Nil(t, err)

	// Test state changes
	err = db.SetProcessGraphState(graph.ID, core.RUNNING)
	assert.Nil(t, err)

	retrievedGraph, err := db.GetProcessGraphByID(graph.ID)
	assert.Nil(t, err)
	assert.Equal(t, retrievedGraph.State, core.RUNNING)

	err = db.SetProcessGraphState(graph.ID, core.SUCCESS)
	assert.Nil(t, err)

	retrievedGraph, err = db.GetProcessGraphByID(graph.ID)
	assert.Nil(t, err)
	assert.Equal(t, retrievedGraph.State, core.SUCCESS)

	// Test invalid process graph ID
	err = db.SetProcessGraphState("invalid_id", core.RUNNING)
	assert.NotNil(t, err)
}

func TestFindProcessGraphs(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Create colony
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Add process graphs in different states
	graph1 := generateProcessGraph(t, db, colony.Name)
	graph1.State = core.WAITING
	err = db.AddProcessGraph(graph1)
	assert.Nil(t, err)

	graph2 := generateProcessGraph(t, db, colony.Name)
	graph2.State = core.RUNNING
	err = db.AddProcessGraph(graph2)
	assert.Nil(t, err)

	graph3 := generateProcessGraph(t, db, colony.Name)
	graph3.State = core.SUCCESS
	err = db.AddProcessGraph(graph3)
	assert.Nil(t, err)

	graph4 := generateProcessGraph(t, db, colony.Name)
	graph4.State = core.FAILED
	err = db.AddProcessGraph(graph4)
	assert.Nil(t, err)

	// Test finding by state
	waitingGraphs, err := db.FindWaitingProcessGraphs(colony.Name, 10)
	assert.Nil(t, err)
	assert.Len(t, waitingGraphs, 1)
	assert.Equal(t, waitingGraphs[0].State, core.WAITING)

	runningGraphs, err := db.FindRunningProcessGraphs(colony.Name, 10)
	assert.Nil(t, err)
	assert.Len(t, runningGraphs, 1)
	assert.Equal(t, runningGraphs[0].State, core.RUNNING)

	successfulGraphs, err := db.FindSuccessfulProcessGraphs(colony.Name, 10)
	assert.Nil(t, err)
	assert.Len(t, successfulGraphs, 1)
	assert.Equal(t, successfulGraphs[0].State, core.SUCCESS)

	failedGraphs, err := db.FindFailedProcessGraphs(colony.Name, 10)
	assert.Nil(t, err)
	assert.Len(t, failedGraphs, 1)
	assert.Equal(t, failedGraphs[0].State, core.FAILED)

	// Test finding from invalid colony
	invalidGraphs, err := db.FindWaitingProcessGraphs("invalid_colony", 10)
	assert.Nil(t, err)
	assert.Empty(t, invalidGraphs)
}

func TestCountProcessGraphs(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Create colony
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Add process graphs in different states
	graph1 := generateProcessGraph(t, db, colony.Name)
	graph1.State = core.WAITING
	err = db.AddProcessGraph(graph1)
	assert.Nil(t, err)

	graph2 := generateProcessGraph(t, db, colony.Name)
	graph2.State = core.RUNNING
	err = db.AddProcessGraph(graph2)
	assert.Nil(t, err)

	graph3 := generateProcessGraph(t, db, colony.Name)
	graph3.State = core.SUCCESS
	err = db.AddProcessGraph(graph3)
	assert.Nil(t, err)

	graph4 := generateProcessGraph(t, db, colony.Name)
	graph4.State = core.FAILED
	err = db.AddProcessGraph(graph4)
	assert.Nil(t, err)

	// Test global counts
	waitingCount, err := db.CountWaitingProcessGraphs()
	assert.Nil(t, err)
	assert.Equal(t, waitingCount, 1)

	runningCount, err := db.CountRunningProcessGraphs()
	assert.Nil(t, err)
	assert.Equal(t, runningCount, 1)

	successfulCount, err := db.CountSuccessfulProcessGraphs()
	assert.Nil(t, err)
	assert.Equal(t, successfulCount, 1)

	failedCount, err := db.CountFailedProcessGraphs()
	assert.Nil(t, err)
	assert.Equal(t, failedCount, 1)

	// Test colony-specific counts
	waitingByColony, err := db.CountWaitingProcessGraphsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, waitingByColony, 1)

	runningByColony, err := db.CountRunningProcessGraphsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, runningByColony, 1)

	successfulByColony, err := db.CountSuccessfulProcessGraphsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, successfulByColony, 1)

	failedByColony, err := db.CountFailedProcessGraphsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, failedByColony, 1)

	// Test invalid colony counts
	invalidCount, err := db.CountWaitingProcessGraphsByColonyName("invalid_colony")
	assert.Nil(t, err)
	assert.Equal(t, invalidCount, 0)
}

func TestRemoveProcessGraphs(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Create colony
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Add process graph
	graph := generateProcessGraph(t, db, colony.Name)
	err = db.AddProcessGraph(graph)
	assert.Nil(t, err)

	// Verify process graph exists
	retrievedGraph, err := db.GetProcessGraphByID(graph.ID)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedGraph)

	// Remove process graph by ID
	err = db.RemoveProcessGraphByID(graph.ID)
	assert.Nil(t, err)

	// Verify process graph is gone
	removedGraph, err := db.GetProcessGraphByID(graph.ID)
	assert.Nil(t, err)
	assert.Nil(t, removedGraph)

	// Test removing non-existing process graph
	err = db.RemoveProcessGraphByID("non_existing_id")
	assert.NotNil(t, err)
}

func TestRemoveAllProcessGraphs(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Create colony
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Add multiple process graphs in different states
	graph1 := generateProcessGraph(t, db, colony.Name)
	graph1.State = core.WAITING
	err = db.AddProcessGraph(graph1)
	assert.Nil(t, err)

	graph2 := generateProcessGraph(t, db, colony.Name)
	graph2.State = core.RUNNING
	err = db.AddProcessGraph(graph2)
	assert.Nil(t, err)

	graph3 := generateProcessGraph(t, db, colony.Name)
	graph3.State = core.SUCCESS
	err = db.AddProcessGraph(graph3)
	assert.Nil(t, err)

	graph4 := generateProcessGraph(t, db, colony.Name)
	graph4.State = core.FAILED
	err = db.AddProcessGraph(graph4)
	assert.Nil(t, err)

	// Verify counts before removal
	totalCount, err := db.CountWaitingProcessGraphsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, totalCount, 1)

	// Test removing by state
	err = db.RemoveAllWaitingProcessGraphsByColonyName(colony.Name)
	assert.Nil(t, err)

	waitingCount, err := db.CountWaitingProcessGraphsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, waitingCount, 0)

	err = db.RemoveAllRunningProcessGraphsByColonyName(colony.Name)
	assert.Nil(t, err)

	err = db.RemoveAllSuccessfulProcessGraphsByColonyName(colony.Name)
	assert.Nil(t, err)

	err = db.RemoveAllFailedProcessGraphsByColonyName(colony.Name)
	assert.Nil(t, err)

	// Verify all are gone
	allCount, err := db.CountWaitingProcessGraphsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, allCount, 0)

	runningCount, err := db.CountRunningProcessGraphsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, runningCount, 0)

	successfulCount, err := db.CountSuccessfulProcessGraphsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, successfulCount, 0)

	failedCount, err := db.CountFailedProcessGraphsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, failedCount, 0)
}

func TestProcessGraphWithProcesses(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Create colony
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Create process graph with associated processes
	process, graph := generateProcessGraph2(t, db, colony.Name)
	err = db.AddProcessGraph(graph)
	assert.Nil(t, err)

	// Verify the process graph was created
	retrievedGraph, err := db.GetProcessGraphByID(graph.ID)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedGraph)
	assert.Equal(t, graph.ID, retrievedGraph.ID)

	// Verify the process has the correct process graph ID
	retrievedProcess, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, graph.ID, retrievedProcess.ProcessGraphID)
}