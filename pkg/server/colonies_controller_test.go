package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestColoniesControllerAddColony(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	controller := createColoniesController(db, nil)

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := controller.addColony(colony)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))
}

func TestColoniesControllerAddRuntime(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	controller := createColoniesController(db, nil)

	runtime, _, err := utils.CreateTestRuntimeWithKey(core.GenerateRandomID())
	assert.Nil(t, err)

	addedRuntime, err := controller.addRuntime(runtime)
	assert.Nil(t, err)
	assert.True(t, runtime.Equals(addedRuntime))
}

func TestColoniesControllerApproveRuntime(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	controller := createColoniesController(db, nil)

	runtime, _, err := utils.CreateTestRuntimeWithKey(core.GenerateRandomID())
	assert.Nil(t, err)

	addedRuntime, err := controller.addRuntime(runtime)
	assert.Nil(t, err)
	assert.True(t, runtime.Equals(addedRuntime))

	err = controller.approveRuntime(runtime.ID)
	assert.Nil(t, err)

	runtimeFromController, err := controller.getRuntimeByID(runtime.ID)
	assert.Nil(t, err)
	assert.True(t, runtimeFromController.IsApproved())
	assert.False(t, runtime.IsApproved())
}

func TestColoniesControllerAddProcess(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	controller := createColoniesController(db, nil)

	colonyID := core.GenerateRandomID()

	runtime, _, err := utils.CreateTestRuntimeWithKey(colonyID)
	assert.Nil(t, err)

	_, err = controller.addRuntime(runtime)
	assert.Nil(t, err)

	processSpec := utils.CreateTestProcessSpecWithEnv(colonyID, make(map[string]string))
	process := core.CreateProcess(processSpec)

	addedProcess, err := controller.addProcess(process)
	assert.Nil(t, err)
	assert.True(t, process.ID == addedProcess.ID)
}

func TestColoniesControllerAssignRuntime(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	controller := createColoniesController(db, nil)

	colonyID := core.GenerateRandomID()

	runtime, _, err := utils.CreateTestRuntimeWithKey(colonyID)
	assert.Nil(t, err)

	_, err = controller.addRuntime(runtime)
	assert.Nil(t, err)

	processSpec := utils.CreateTestProcessSpecWithEnv(colonyID, make(map[string]string))
	process := core.CreateProcess(processSpec)
	_, err = controller.addProcess(process)
	assert.Nil(t, err)

	assignedProcess, err := controller.assignRuntime(runtime.ID, colonyID)
	assert.Nil(t, err)
	assert.True(t, process.ID == assignedProcess.ID)
}

// notest
func TestColoniesControllerAssignRuntimeConcurrency(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	processCount := 100

	controller1 := createColoniesController(db, nil)
	controller2 := createColoniesController(db, nil)

	colonyID := core.GenerateRandomID()

	runtime1, _, err := utils.CreateTestRuntimeWithKey(colonyID)
	assert.Nil(t, err)
	_, err = controller1.addRuntime(runtime1)
	assert.Nil(t, err)

	runtime2, _, err := utils.CreateTestRuntimeWithKey(colonyID)
	assert.Nil(t, err)
	_, err = controller1.addRuntime(runtime2)
	assert.Nil(t, err)

	for i := 0; i < processCount; i++ {
		processSpec := utils.CreateTestProcessSpecWithEnv(colonyID, make(map[string]string))
		process := core.CreateProcess(processSpec)
		_, err = controller1.addProcess(process)
		assert.Nil(t, err)
	}

	countChan := make(chan int)

	go func() {
		for {
			_, err := controller1.assignRuntime(runtime1.ID, colonyID)
			if err == nil {
				countChan <- 1
			}
		}
	}()

	// Since we are using two different controller there should be an error: "Process already assigned"
	// That can happen if two runtime clients manage to be assigned the same process
	// A simple solution is just that the second clients gets an error

	go func() {
		for {
			_, err := controller2.assignRuntime(runtime2.ID, colonyID)
			if err == nil {
				countChan <- 1
			}
		}
	}()

	count := 0
	for {
		count += <-countChan
		if count == processCount {
			break
		}
	}
}
