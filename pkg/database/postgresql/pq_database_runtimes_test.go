package postgresql

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddRuntime(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	runtime := utils.CreateTestRuntime(colony.ID)
	err = db.AddRuntime(runtime)
	assert.Nil(t, err)

	runtimes, err := db.GetRuntimes()
	assert.Nil(t, err)

	runtimeFromDB := runtimes[0]
	assert.True(t, runtime.Equals(runtimeFromDB))
	assert.True(t, runtimeFromDB.IsPending())
	assert.False(t, runtimeFromDB.IsApproved())
	assert.False(t, runtimeFromDB.IsRejected())
}

func TestAddTwoRuntime(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	runtime1 := utils.CreateTestRuntime(colony.ID)
	err = db.AddRuntime(runtime1)
	assert.Nil(t, err)

	runtime2 := utils.CreateTestRuntime(colony.ID)
	err = db.AddRuntime(runtime2)
	assert.Nil(t, err)

	var runtimes []*core.Runtime
	runtimes = append(runtimes, runtime1)
	runtimes = append(runtimes, runtime2)

	runtimesFromDB, err := db.GetRuntimes()
	assert.Nil(t, err)
	assert.True(t, core.IsRuntimeArraysEqual(runtimes, runtimesFromDB))
}

func TestGetRuntimeByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	runtime1 := utils.CreateTestRuntime(colony.ID)
	err = db.AddRuntime(runtime1)
	assert.Nil(t, err)

	runtime2 := utils.CreateTestRuntime(colony.ID)
	err = db.AddRuntime(runtime2)
	assert.Nil(t, err)

	runtimeFromDB, err := db.GetRuntimeByID(runtime1.ID)
	assert.Nil(t, err)
	assert.True(t, runtime1.Equals(runtimeFromDB))
}

func TestGetRuntimeByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	assert.Nil(t, err)
	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
	assert.Nil(t, err)

	err = db.AddColony(colony2)
	assert.Nil(t, err)

	runtime1 := utils.CreateTestRuntime(colony1.ID)
	err = db.AddRuntime(runtime1)
	assert.Nil(t, err)

	runtime2 := utils.CreateTestRuntime(colony1.ID)
	err = db.AddRuntime(runtime2)
	assert.Nil(t, err)

	runtime3 := utils.CreateTestRuntime(colony2.ID)
	err = db.AddRuntime(runtime3)
	assert.Nil(t, err)

	var runtimesColony1 []*core.Runtime
	runtimesColony1 = append(runtimesColony1, runtime1)
	runtimesColony1 = append(runtimesColony1, runtime2)

	runtimesColony1FromDB, err := db.GetRuntimesByColonyID(colony1.ID)
	assert.Nil(t, err)
	assert.True(t, core.IsRuntimeArraysEqual(runtimesColony1, runtimesColony1FromDB))
}

func TestMarkAlive(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	runtime := utils.CreateTestRuntime(colony.ID)
	err = db.AddRuntime(runtime)
	assert.Nil(t, err)

	time.Sleep(3000 * time.Millisecond)

	err = db.MarkAlive(runtime)
	assert.Nil(t, err)

	runtimeFromDB, err := db.GetRuntimeByID(runtime.ID)
	assert.Nil(t, err)

	assert.True(t, (runtimeFromDB.LastHeardFromTime.Unix()-runtime.LastHeardFromTime.Unix()) > 1)
}

func TestApproveRuntime(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	runtime := utils.CreateTestRuntime(colony.ID)
	err = db.AddRuntime(runtime)
	assert.Nil(t, err)

	assert.True(t, runtime.IsPending())

	err = db.ApproveRuntime(runtime)
	assert.Nil(t, err)

	assert.False(t, runtime.IsPending())
	assert.False(t, runtime.IsRejected())
	assert.True(t, runtime.IsApproved())

	runtimeFromDB, err := db.GetRuntimeByID(runtime.ID)
	assert.Nil(t, err)
	assert.True(t, runtimeFromDB.IsApproved())

	err = db.RejectRuntime(runtime)
	assert.Nil(t, err)
	assert.True(t, runtime.IsRejected())

	runtimeFromDB, err = db.GetRuntimeByID(runtime.ID)
	assert.Nil(t, err)
	assert.True(t, runtime.IsRejected())
}

func TestDeleteRuntimeMoveBackToQueue(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	runtime1 := utils.CreateTestRuntime(colony.ID)
	err = db.AddRuntime(runtime1)
	assert.Nil(t, err)

	runtime2 := utils.CreateTestRuntime(colony.ID)
	err = db.AddRuntime(runtime2)
	assert.Nil(t, err)

	env := make(map[string]string)

	process1 := utils.CreateTestProcessWithEnv(colony.ID, env)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcessWithEnv(colony.ID, env)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcessWithEnv(colony.ID, env)
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	process4 := utils.CreateTestProcessWithEnv(colony.ID, env)
	err = db.AddProcess(process4)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == "")

	processFromDB, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == "")

	processFromDB, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == "")

	processFromDB, err = db.GetProcessByID(process4.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == "")

	err = db.AssignRuntime(runtime1.ID, process1)
	assert.Nil(t, err)
	err = db.AssignRuntime(runtime1.ID, process2)
	assert.Nil(t, err)
	err = db.AssignRuntime(runtime2.ID, process3)
	assert.Nil(t, err)
	err = db.AssignRuntime(runtime1.ID, process4)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == runtime1.ID)

	processFromDB, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == runtime1.ID)

	processFromDB, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == runtime2.ID)

	count, err := db.CountWaitingProcessesByColonyID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, count == 0)

	err = db.MarkSuccessful(process4.ID)
	assert.Nil(t, err)

	err = db.DeleteRuntimeByID(runtime1.ID)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == "")

	processFromDB, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == "")

	processFromDB, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == runtime2.ID)

	count, err = db.CountWaitingProcessesByColonyID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, count == 2)

	count, err = db.CountSuccessfulProcessesByColonyID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, count == 1)
}

func TestDeleteRuntimesMoveBackToQueue(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	runtime1 := utils.CreateTestRuntime(colony.ID)
	err = db.AddRuntime(runtime1)
	assert.Nil(t, err)

	runtime2 := utils.CreateTestRuntime(colony.ID)
	err = db.AddRuntime(runtime2)
	assert.Nil(t, err)

	env := make(map[string]string)

	process1 := utils.CreateTestProcessWithEnv(colony.ID, env)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcessWithEnv(colony.ID, env)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcessWithEnv(colony.ID, env)
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	process4 := utils.CreateTestProcessWithEnv(colony.ID, env)
	err = db.AddProcess(process4)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == "")

	processFromDB, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == "")

	processFromDB, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == "")

	processFromDB, err = db.GetProcessByID(process4.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == "")

	err = db.AssignRuntime(runtime1.ID, process1)
	assert.Nil(t, err)
	err = db.AssignRuntime(runtime1.ID, process2)
	assert.Nil(t, err)
	err = db.AssignRuntime(runtime2.ID, process3)
	assert.Nil(t, err)
	err = db.AssignRuntime(runtime1.ID, process4)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == runtime1.ID)

	processFromDB, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == runtime1.ID)

	processFromDB, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == runtime2.ID)

	count, err := db.CountWaitingProcessesByColonyID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, count == 0)

	err = db.MarkSuccessful(process4.ID)
	assert.Nil(t, err)

	err = db.DeleteRuntimesByColonyID(colony.ID)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == "")

	processFromDB, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == "")

	processFromDB, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedRuntimeID == "")

	count, err = db.CountWaitingProcessesByColonyID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, count == 3)

	count, err = db.CountSuccessfulProcessesByColonyID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, count == 1)
}

func TestDeleteRuntimes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")

	err = db.AddColony(colony2)
	assert.Nil(t, err)

	runtime1 := utils.CreateTestRuntime(colony1.ID)
	err = db.AddRuntime(runtime1)
	assert.Nil(t, err)

	runtime2 := utils.CreateTestRuntime(colony1.ID)
	err = db.AddRuntime(runtime2)
	assert.Nil(t, err)

	runtime3 := utils.CreateTestRuntime(colony2.ID)
	err = db.AddRuntime(runtime3)
	assert.Nil(t, err)

	err = db.DeleteRuntimeByID(runtime2.ID)
	assert.Nil(t, err)

	runtimeFromDB, err := db.GetRuntimeByID(runtime2.ID)
	assert.Nil(t, err)
	assert.Nil(t, runtimeFromDB)

	err = db.AddRuntime(runtime2)
	assert.Nil(t, err)

	runtimeFromDB, err = db.GetRuntimeByID(runtime2.ID)
	assert.Nil(t, err)
	assert.NotNil(t, runtimeFromDB)

	err = db.DeleteRuntimesByColonyID(colony1.ID)
	assert.Nil(t, err)

	runtimeFromDB, err = db.GetRuntimeByID(runtime1.ID)
	assert.Nil(t, err)
	assert.Nil(t, runtimeFromDB)

	runtimeFromDB, err = db.GetRuntimeByID(runtime2.ID)
	assert.Nil(t, err)
	assert.Nil(t, runtimeFromDB)

	runtimeFromDB, err = db.GetRuntimeByID(runtime3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, runtimeFromDB)
}

func TestCountRuntimes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	runtimeCount, err := db.CountRuntimes()
	assert.Nil(t, err)
	assert.True(t, runtimeCount == 0)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	runtime := utils.CreateTestRuntime(colony.ID)
	err = db.AddRuntime(runtime)
	assert.Nil(t, err)

	runtimeCount, err = db.CountRuntimes()
	assert.Nil(t, err)
	assert.True(t, runtimeCount == 1)
}

func TestCountRuntimesByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony1)
	assert.Nil(t, err)

	runtime := utils.CreateTestRuntime(colony1.ID)
	err = db.AddRuntime(runtime)
	assert.Nil(t, err)

	runtime = utils.CreateTestRuntime(colony1.ID)
	err = db.AddRuntime(runtime)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	runtime = utils.CreateTestRuntime(colony2.ID)
	err = db.AddRuntime(runtime)
	assert.Nil(t, err)

	runtimeCount, err := db.CountRuntimes()
	assert.Nil(t, err)
	assert.True(t, runtimeCount == 3)

	runtimeCount, err = db.CountRuntimesByColonyID(colony1.ID)
	assert.Nil(t, err)
	assert.True(t, runtimeCount == 2)

	runtimeCount, err = db.CountRuntimesByColonyID(colony2.ID)
	assert.Nil(t, err)
	assert.True(t, runtimeCount == 1)
}
