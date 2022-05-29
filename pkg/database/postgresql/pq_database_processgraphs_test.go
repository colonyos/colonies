package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func generateProcessGraph(t *testing.T, db *PQDatabase) *core.ProcessGraph {
	colonyID := core.GenerateRandomID()
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

	graph, err := core.CreateProcessGraph(db, process1.ID)
	assert.Nil(t, err)

	return graph
}

func TestAddProcessGraph(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	graph := generateProcessGraph(t, db)

	err = db.AddProcessGraph(graph)
	assert.Nil(t, err)

	graph2, err := db.GetProcessGraphByID(db, graph.ID)
	assert.Nil(t, err)
	assert.True(t, graph.Equals(graph2))
}

func TestSetProcessGraphState(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	graph := generateProcessGraph(t, db)

	err = db.AddProcessGraph(graph)
	assert.Nil(t, err)

	err = db.SetProcessGraphState(graph.ID, core.WAITING)
	assert.Nil(t, err)
	graph2, err := db.GetProcessGraphByID(db, graph.ID)
	assert.Nil(t, err)
	assert.True(t, graph2.State == core.WAITING)

	err = db.SetProcessGraphState(graph.ID, core.FAILED)
	assert.Nil(t, err)
	graph2, err = db.GetProcessGraphByID(db, graph.ID)
	assert.Nil(t, err)
	assert.True(t, graph2.State == core.FAILED)
}
