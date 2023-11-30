package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestColoniesControllerInvalidDB(t *testing.T) {
	controller, dbMock := createFakeColoniesController()

	dbMock.returnError = "GetProcessByID"
	err := controller.subscribeProcess("invalid_id", &subscription{})
	assert.NotNil(t, err)

	dbMock.returnError = "GetColonies"
	_, err = controller.getColonies()
	assert.NotNil(t, err)

	dbMock.returnError = "GetColonyByName"
	_, err = controller.getColony("invalid_id")
	assert.NotNil(t, err)

	dbMock.returnError = "AddColony"
	_, err = controller.addColony(nil)
	assert.NotNil(t, err)

	dbMock.returnError = ""
	_, err = controller.addColony(nil)
	assert.NotNil(t, err)

	dbMock.returnError = "GetColonyByID"
	_, err = controller.addColony(&core.Colony{})
	assert.NotNil(t, err)

	dbMock.returnError = "RemoveColonyByName"
	err = controller.removeColony("invalid_id")
	assert.NotNil(t, err)

	_, err = controller.addExecutor(nil, false)
	assert.NotNil(t, err)

	dbMock.returnError = "GetExecutorByName"
	_, err = controller.addExecutor(&core.Executor{}, false)
	assert.NotNil(t, err)

	dbMock.returnValue = "GetExecutorByName"
	_, err = controller.addExecutor(&core.Executor{}, false)
	assert.NotNil(t, err)
	dbMock.returnValue = ""

	dbMock.returnError = "AddExecutor"
	_, err = controller.addExecutor(&core.Executor{}, false)
	assert.NotNil(t, err)

	dbMock.returnError = "GetExecutorByID"
	_, err = controller.addExecutor(&core.Executor{}, false)
	assert.NotNil(t, err)

	dbMock.returnError = "GetExecutorByID"
	_, err = controller.getExecutor("invalid_id")
	assert.NotNil(t, err)

	dbMock.returnError = "GetExecutorByColonyName"
	_, err = controller.getExecutorByColonyName("invalid_id")
	assert.NotNil(t, err)

	_, err = controller.addProcessToDB(nil)
	assert.NotNil(t, err)

	dbMock.returnError = "AddProcess"
	_, err = controller.addProcessToDB(&core.Process{})
	assert.NotNil(t, err)

	dbMock.returnError = "GetProcessByID"
	_, err = controller.addProcessToDB(&core.Process{})
	assert.NotNil(t, err)

	dbMock.returnError = "AddProcess"
	_, err = controller.addProcess(&core.Process{})
	assert.NotNil(t, err)

	controller.stop()
}

func TestColoniesControllerAddColony(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	controller := createTestColoniesController(db)
	defer controller.stop()

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := controller.addColony(colony)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))
}

func TestColoniesControllerAddExecutor(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	controller := createTestColoniesController(db)
	defer controller.stop()

	executor, _, err := utils.CreateTestExecutorWithKey(core.GenerateRandomID())
	assert.Nil(t, err)

	addedExecutor, err := controller.addExecutor(executor, false)
	assert.Nil(t, err)
	assert.True(t, executor.Equals(addedExecutor))
}

func TestColoniesControllerAddProcess(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	controller := createTestColoniesController(db)
	defer controller.stop()

	colonyName := core.GenerateRandomID()

	executor, _, err := utils.CreateTestExecutorWithKey(colonyName)
	assert.Nil(t, err)

	_, err = controller.addExecutor(executor, false)
	assert.Nil(t, err)

	funcSpec := utils.CreateTestFunctionSpecWithEnv(colonyName, make(map[string]string))
	process := core.CreateProcess(funcSpec)

	addedProcess, err := controller.addProcess(process)
	assert.Nil(t, err)
	assert.True(t, process.ID == addedProcess.ID)
}

func TestColoniesControllerAssignExecutor(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	controller := createTestColoniesController(db)
	defer controller.stop()

	colonyName := core.GenerateRandomID()

	executor, _, err := utils.CreateTestExecutorWithKey(colonyName)
	assert.Nil(t, err)

	_, err = controller.addExecutor(executor, false)
	assert.Nil(t, err)

	funcSpec := utils.CreateTestFunctionSpecWithEnv(colonyName, make(map[string]string))
	process := core.CreateProcess(funcSpec)
	_, err = controller.addProcess(process)
	assert.Nil(t, err)

	assignedProcess, err := controller.assign(executor.ID, colonyName)
	assert.Nil(t, err)
	assert.True(t, process.ID == assignedProcess.ID)
}

// notest
func TestColoniesControllerAssignExecutorConcurrency(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	processCount := 100

	controller1 := createTestColoniesController(db)
	defer controller1.stop()
	controller2 := createTestColoniesController2(db)
	defer controller2.stop()

	colonyName := core.GenerateRandomID()

	executor1, _, err := utils.CreateTestExecutorWithKey(colonyName)
	assert.Nil(t, err)
	_, err = controller1.addExecutor(executor1, false)
	assert.Nil(t, err)

	executor2, _, err := utils.CreateTestExecutorWithKey(colonyName)
	assert.Nil(t, err)
	_, err = controller1.addExecutor(executor2, false)
	assert.Nil(t, err)

	for i := 0; i < processCount; i++ {
		funcSpec := utils.CreateTestFunctionSpecWithEnv(colonyName, make(map[string]string))
		process := core.CreateProcess(funcSpec)
		_, err = controller1.addProcess(process)
		assert.Nil(t, err)
	}

	countChan := make(chan int)

	go func() {
		for {
			_, err := controller1.assign(executor1.ID, colonyName)
			if err == nil {
				countChan <- 1
			}
		}
	}()

	// Since we are using two different controller there should be an error: "Process already assigned"
	// That can happen if two executor clients manage to be assigned the same process
	// A simple solution is just that the second clients gets an error

	go func() {
		for {
			_, err := controller2.assign(executor2.ID, colonyName)
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
