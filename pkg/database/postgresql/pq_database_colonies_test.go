package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddColony(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	colonies, err := db.GetColonies()
	assert.Nil(t, err)

	colonyFromDB := colonies[0]
	assert.True(t, colony.Equals(colonyFromDB))

	colonyFromDB, err = db.GetColonyByID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(colonyFromDB))
}

func TestAddTwoColonies(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	var colonies []*core.Colony
	colonies = append(colonies, colony1)
	colonies = append(colonies, colony2)

	coloniesFromDB, err := db.GetColonies()
	assert.Nil(t, err)
	assert.True(t, core.IsColonyArraysEqual(colonies, coloniesFromDB))
}

func TestGetColonyByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")

	err = db.AddColony(colony2)
	assert.Nil(t, err)

	colonyFromDB, err := db.GetColonyByID(colony1.ID)
	assert.Nil(t, err)
	assert.Equal(t, colony1.ID, colonyFromDB.ID)

	colonyFromDB, err = db.GetColonyByID(core.GenerateRandomID())
	assert.Nil(t, err)
}

func TestDeleteColonies(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")

	err = db.AddColony(colony2)
	assert.Nil(t, err)

	generator1 := utils.FakeGenerator(t, colony1.ID)
	generator1.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator1)
	assert.Nil(t, err)

	generator2 := utils.FakeGenerator(t, colony2.ID)
	generator2.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator2)
	assert.Nil(t, err)

	cron1 := utils.FakeCron(t, colony1.ID)
	cron1.ID = core.GenerateRandomID()
	err = db.AddCron(cron1)
	assert.Nil(t, err)

	cron2 := utils.FakeCron(t, colony2.ID)
	cron2.ID = core.GenerateRandomID()
	err = db.AddCron(cron2)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony1.ID)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	function := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: executor1.ID, ColonyID: colony1.ID, FuncName: "testfunc", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony1.ID)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	function = &core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: executor2.ID, ColonyID: colony1.ID, FuncName: "testfunc", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	executor3 := utils.CreateTestExecutor(colony2.ID)
	err = db.AddExecutor(executor3)
	assert.Nil(t, err)

	function = &core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: executor3.ID, ColonyID: colony2.ID, FuncName: "testfunc", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	err = db.DeleteColonyByID(colony1.ID)
	assert.Nil(t, err)

	colonyFromDB, err := db.GetColonyByID(colony1.ID)
	assert.Nil(t, err)
	assert.Nil(t, colonyFromDB)

	executorFromDB, err := db.GetExecutorByID(executor1.ID)
	assert.Nil(t, err)
	assert.Nil(t, executorFromDB)

	executorFromDB, err = db.GetExecutorByID(executor2.ID)
	assert.Nil(t, err)
	assert.Nil(t, executorFromDB)

	executorFromDB, err = db.GetExecutorByID(executor3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, executorFromDB) // Belongs to Colony 2 and should therefore NOT be deleted

	generatorFromDB, err := db.GetGeneratorByID(generator1.ID)
	assert.Nil(t, err)
	assert.Nil(t, generatorFromDB) // Should have been deleted

	generatorFromDB, err = db.GetGeneratorByID(generator2.ID)
	assert.Nil(t, err)
	assert.NotNil(t, generatorFromDB) // Should NOT have been deleted

	cronFromDB, err := db.GetCronByID(cron1.ID)
	assert.Nil(t, err)
	assert.Nil(t, cronFromDB) // Should have been deleted

	cronFromDB, err = db.GetCronByID(cron2.ID)
	assert.Nil(t, err)
	assert.NotNil(t, cronFromDB) // Should NOT have been deleted

	functions, err := db.GetFunctionsByColonyID(colony1.ID)
	assert.Len(t, functions, 0)

	functions, err = db.GetFunctionsByColonyID(colony2.ID)
	assert.Len(t, functions, 1)
}

func TestCountColonies(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	coloniesCount, err := db.CountColonies()
	assert.Nil(t, err)
	assert.True(t, coloniesCount == 0)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	coloniesCount, err = db.CountColonies()
	assert.Nil(t, err)
	assert.True(t, coloniesCount == 1)

	colony = core.CreateColony(core.GenerateRandomID(), "test_colony_name2")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	coloniesCount, err = db.CountColonies()
	assert.Nil(t, err)
	assert.True(t, coloniesCount == 2)
}
