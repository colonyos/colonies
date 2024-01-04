package postgresql

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestColonyClosedDB(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.NotNil(t, err)

	_, err = db.GetColonies()
	assert.NotNil(t, err)

	_, err = db.GetColonyByID("invalid_id")
	assert.NotNil(t, err)

	err = db.RenameColony("invalid_id", "invalid_name")
	assert.NotNil(t, err)

	err = db.RemoveColonyByName("invalid_id")
	assert.NotNil(t, err)

	_, err = db.CountColonies()
	assert.NotNil(t, err)
}

func TestAddColony(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(nil)
	assert.NotNil(t, err)

	err = db.AddColony(colony)
	assert.Nil(t, err)

	err = db.AddColony(colony) // Try to add the same colony again
	assert.NotNil(t, err)      // Error

	colonies, err := db.GetColonies()
	assert.Nil(t, err)

	colonyFromDB := colonies[0]
	assert.True(t, colony.Equals(colonyFromDB))

	colonyFromDB, err = db.GetColonyByID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(colonyFromDB))
}

func TestRenameColony(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	colonyFromDB, err := db.GetColonyByID(colony.ID)
	assert.Nil(t, err)
	assert.Equal(t, colonyFromDB.Name, "test_colony_name")

	err = db.RenameColony(colony.Name, "test_colony_new_name")
	assert.Nil(t, err)

	colonyFromDB, err = db.GetColonyByID(colony.ID)
	assert.Nil(t, err)
	assert.Equal(t, colonyFromDB.Name, "test_colony_new_name")
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

func TestGetColonyByName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")

	err = db.AddColony(colony2)
	assert.Nil(t, err)

	colonyFromDB, err := db.GetColonyByName("test_colony_name_1")
	assert.Nil(t, err)
	assert.Equal(t, colony1.ID, colonyFromDB.ID)
}

func TestRemoveColonies(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")

	err = db.AddColony(colony2)
	assert.Nil(t, err)

	user1 := utils.CreateTestUser(colony1.Name, "user1")
	err = db.AddUser(user1)
	assert.Nil(t, err)

	user2 := utils.CreateTestUser(colony2.Name, "user2")
	err = db.AddUser(user2)
	assert.Nil(t, err)

	generator1 := utils.FakeGenerator(t, colony1.Name, "test_initiator_id", "test_initiator_name")
	generator1.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator1)
	assert.Nil(t, err)

	generator2 := utils.FakeGenerator(t, colony2.Name, "test_initiator_id", "test_initiator_name")
	generator2.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator2)
	assert.Nil(t, err)

	cron1 := utils.FakeCron(t, colony1.Name, "test_initiator_id", "test_initiator_name")
	cron1.ID = core.GenerateRandomID()
	err = db.AddCron(cron1)
	assert.Nil(t, err)

	cron2 := utils.FakeCron(t, colony2.Name, "test_initiator_id", "test_initiator_name")
	cron2.ID = core.GenerateRandomID()
	err = db.AddCron(cron2)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony1.Name)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	function := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executor1.Name, ColonyName: colony1.Name, FuncName: "testfunc", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony1.Name)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	function = &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executor2.Name, ColonyName: colony1.Name, FuncName: "testfunc", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	executor3 := utils.CreateTestExecutor(colony2.Name)
	err = db.AddExecutor(executor3)
	assert.Nil(t, err)

	function = &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executor3.Name, ColonyName: colony2.Name, FuncName: "testfunc", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	err = db.AddLog("test_processid1", colony1.ID, "test_executor_name", time.Now().UTC().UnixNano(), "1")
	assert.Nil(t, err)

	err = db.AddLog("test_processid1", colony2.ID, "test_executor_name", time.Now().UTC().UnixNano(), "1")
	assert.Nil(t, err)

	file := utils.CreateTestFileWithID("test_id", colony1.Name, time.Now())
	file.ID = core.GenerateRandomID()
	file.Label = "/testdir"
	file.Name = "test_file2.txt"
	file.Size = 1
	err = db.AddFile(file)
	assert.Nil(t, err)

	file = utils.CreateTestFileWithID("test_id", colony2.Name, time.Now())
	file.ID = core.GenerateRandomID()
	file.Label = "/testdir"
	file.Name = "test_file2.txt"
	file.Size = 1
	err = db.AddFile(file)
	assert.Nil(t, err)

	_, err = db.CreateSnapshot(colony1.Name, "/testdir", "test_snapshot_name1")
	assert.Nil(t, err)
	_, err = db.CreateSnapshot(colony2.Name, "/testdir", "test_snapshot_name2")
	assert.Nil(t, err)

	err = db.RemoveColonyByName(core.GenerateRandomID())
	assert.NotNil(t, err)

	err = db.RemoveColonyByName(colony1.Name)
	assert.Nil(t, err)

	users, err := db.GetUsersByColonyName(colony1.Name)
	assert.Len(t, users, 0)

	users, err = db.GetUsersByColonyName(colony2.Name)
	assert.Len(t, users, 1)

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
	assert.NotNil(t, executorFromDB) // Belongs to Colony 2 and should therefore NOT be removed

	generatorFromDB, err := db.GetGeneratorByID(generator1.ID)
	assert.Nil(t, err)
	assert.Nil(t, generatorFromDB) // Should have been removed

	generatorFromDB, err = db.GetGeneratorByID(generator2.ID)
	assert.Nil(t, err)
	assert.NotNil(t, generatorFromDB) // Should NOT have been removed

	cronFromDB, err := db.GetCronByID(cron1.ID)
	assert.Nil(t, err)
	assert.Nil(t, cronFromDB) // Should have been removed

	cronFromDB, err = db.GetCronByID(cron2.ID)
	assert.Nil(t, err)
	assert.NotNil(t, cronFromDB) // Should NOT have been removed

	functions, err := db.GetFunctionsByColonyName(colony1.Name)
	assert.Nil(t, err)
	assert.Len(t, functions, 0)

	functions, err = db.GetFunctionsByColonyName(colony2.Name)
	assert.Nil(t, err)
	assert.Len(t, functions, 1)

	logsCount, err := db.CountLogs(colony1.Name)
	assert.Nil(t, err)
	assert.Equal(t, logsCount, 0)

	logsCount, err = db.CountFiles(colony2.Name)
	assert.Nil(t, err)
	assert.Equal(t, logsCount, 1)

	fileCount, err := db.CountFiles(colony1.Name)
	assert.Nil(t, err)
	assert.Equal(t, fileCount, 0)

	fileCount, err = db.CountFiles(colony2.Name)
	assert.Nil(t, err)
	assert.Equal(t, fileCount, 1)

	snapshots, err := db.GetSnapshotsByColonyName(colony1.Name)
	assert.Nil(t, err)
	assert.Len(t, snapshots, 0)

	snapshots, err = db.GetSnapshotsByColonyName(colony2.Name)
	assert.Nil(t, err)
	assert.Len(t, snapshots, 1)
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

func TestChangeColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	colonyFromDB, err := db.GetColonyByName(colony.Name)
	assert.Nil(t, err)

	err = db.ChangeColonyID(colony.Name, colony.ID, "new_id")
	assert.Nil(t, err)

	colonyFromDB, err = db.GetColonyByName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, "new_id", colonyFromDB.ID)
	assert.NotEqual(t, colony.ID, colonyFromDB.ID)
}
