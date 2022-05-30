package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func generateProcessGraph(t *testing.T, db *PQDatabase, colonyID string) *core.ProcessGraph {
	process1 := utils.CreateTestProcess(colonyID)
	process2 := utils.CreateTestProcess(colonyID)
	process3 := utils.CreateTestProcess(colonyID)
	process4 := utils.CreateTestProcess(colonyID)

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

	graph, err := core.CreateProcessGraph(colonyID)
	assert.Nil(t, err)

	graph.AddRoot(process1.ID)

	graph.RuntimeGroup = "local"

	return graph
}

func TestAddProcessGraph(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colonyID := core.GenerateRandomID()

	graph := generateProcessGraph(t, db, colonyID)

	err = db.AddProcessGraph(graph)
	assert.Nil(t, err)

	graph2, err := db.GetProcessGraphByID(graph.ID)
	assert.Nil(t, err)
	assert.True(t, graph.Equals(graph2))
}

func TestSetProcessGraphState(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colonyID := core.GenerateRandomID()

	graph := generateProcessGraph(t, db, colonyID)

	err = db.AddProcessGraph(graph)
	assert.Nil(t, err)

	err = db.SetProcessGraphState(graph.ID, core.WAITING)
	assert.Nil(t, err)
	graph2, err := db.GetProcessGraphByID(graph.ID)
	assert.Nil(t, err)
	assert.True(t, graph2.State == core.WAITING)

	err = db.SetProcessGraphState(graph.ID, core.FAILED)
	assert.Nil(t, err)
	graph2, err = db.GetProcessGraphByID(graph.ID)
	assert.Nil(t, err)
	assert.True(t, graph2.State == core.FAILED)
}

func TestFindProcessGraphs(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	var colonyID string
	for j := 0; j < 2; j++ {
		colonyID = core.GenerateRandomID()
		for i := 0; i < 10; i++ {
			graph := generateProcessGraph(t, db, colonyID)
			err = db.AddProcessGraph(graph)
			assert.Nil(t, err)
			err = db.SetProcessGraphState(graph.ID, core.WAITING)
			assert.Nil(t, err)
		}

		for i := 0; i < 9; i++ {
			graph := generateProcessGraph(t, db, colonyID)
			err = db.AddProcessGraph(graph)
			assert.Nil(t, err)
			err = db.SetProcessGraphState(graph.ID, core.RUNNING)
			assert.Nil(t, err)
		}

		for i := 0; i < 8; i++ {
			graph := generateProcessGraph(t, db, colonyID)
			err = db.AddProcessGraph(graph)
			assert.Nil(t, err)
			err = db.SetProcessGraphState(graph.ID, core.FAILED)
			assert.Nil(t, err)
		}

		for i := 0; i < 7; i++ {
			graph := generateProcessGraph(t, db, colonyID)
			err = db.AddProcessGraph(graph)
			assert.Nil(t, err)
			err = db.SetProcessGraphState(graph.ID, core.SUCCESS)
			assert.Nil(t, err)
		}
	}

	graphs, err := db.FindWaitingProcessGraphs(colonyID, 100)
	assert.Nil(t, err)
	assert.Len(t, graphs, 10)

	graphs, err = db.FindRunningProcessGraphs(colonyID, 100)
	assert.Nil(t, err)
	assert.Len(t, graphs, 9)

	graphs, err = db.FindFailedProcessGraphs(colonyID, 100)
	assert.Nil(t, err)
	assert.Len(t, graphs, 8)

	graphs, err = db.FindSuccessfulProcessGraphs(colonyID, 100)
	assert.Nil(t, err)
	assert.Len(t, graphs, 7)

	count, err := db.NrOfWaitingProcessGraphsForColony(colonyID)
	assert.Nil(t, err)
	assert.True(t, count == 10)

	count, err = db.NrOfRunningProcessGraphsForColony(colonyID)
	assert.Nil(t, err)
	assert.True(t, count == 9)

	count, err = db.NrOfFailedProcessGraphsForColony(colonyID)
	assert.Nil(t, err)
	assert.True(t, count == 8)

	count, err = db.NrOfSuccessfulProcessGraphsForColony(colonyID)
	assert.Nil(t, err)
	assert.True(t, count == 7)

	count, err = db.NrOfWaitingProcessGraphs()
	assert.Nil(t, err)
	assert.True(t, count == 10*2)

	count, err = db.NrOfRunningProcessGraphs()
	assert.Nil(t, err)
	assert.True(t, count == 9*2)

	count, err = db.NrOfFailedProcessGraphs()
	assert.Nil(t, err)
	assert.True(t, count == 8*2)

	count, err = db.NrOfSuccessfulProcessGraphs()
	assert.Nil(t, err)
	assert.True(t, count == 7*2)
}
