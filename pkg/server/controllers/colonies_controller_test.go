package controllers

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	websockethandlers "github.com/colonyos/colonies/pkg/server/handlers/websocket"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestColoniesControllerInvalidDB(t *testing.T) {
	controller, dbMock := createFakeColoniesController()

	dbMock.ReturnError = "GetProcessByID"
	err := controller.SubscribeProcess("invalid_id", &websockethandlers.Subscription{})
	assert.NotNil(t, err)

	dbMock.ReturnError = "GetColonies"
	_, err = controller.GetColonies()
	assert.NotNil(t, err)

	dbMock.ReturnError = "GetColonyByName"
	_, err = controller.GetColony("invalid_id")
	assert.NotNil(t, err)

	dbMock.ReturnError = "AddColony"
	_, err = controller.AddColony(nil)
	assert.NotNil(t, err)

	dbMock.ReturnError = ""
	_, err = controller.AddColony(nil)
	assert.NotNil(t, err)

	dbMock.ReturnError = "GetColonyByID"
	_, err = controller.AddColony(&core.Colony{})
	assert.NotNil(t, err)

	dbMock.ReturnError = "RemoveColonyByName"
	err = controller.RemoveColony("invalid_id")
	assert.NotNil(t, err)

	_, err = controller.AddExecutor(nil, false)
	assert.NotNil(t, err)

	dbMock.ReturnError = "GetExecutorByName"
	_, err = controller.AddExecutor(&core.Executor{}, false)
	assert.NotNil(t, err)

	dbMock.ReturnValue = "GetExecutorByName"
	_, err = controller.AddExecutor(&core.Executor{}, false)
	assert.NotNil(t, err)
	dbMock.ReturnValue = ""

	dbMock.ReturnError = "AddExecutor"
	_, err = controller.AddExecutor(&core.Executor{}, false)
	assert.NotNil(t, err)

	dbMock.ReturnError = "GetExecutorByID"
	_, err = controller.AddExecutor(&core.Executor{}, false)
	assert.NotNil(t, err)

	dbMock.ReturnError = "GetExecutorByID"
	_, err = controller.GetExecutor("invalid_id")
	assert.NotNil(t, err)

	dbMock.ReturnError = "GetExecutorByColonyName"
	_, err = controller.GetExecutorByColonyName("invalid_id")
	assert.NotNil(t, err)

	_, err = controller.AddProcessToDB(nil)
	assert.NotNil(t, err)

	dbMock.ReturnError = "AddProcess"
	_, err = controller.AddProcessToDB(&core.Process{})
	assert.NotNil(t, err)

	dbMock.ReturnError = "GetProcessByID"
	_, err = controller.AddProcessToDB(&core.Process{})
	assert.NotNil(t, err)

	dbMock.ReturnError = "AddProcess"
	_, err = controller.AddProcess(&core.Process{})
	assert.NotNil(t, err)

	controller.Stop()
}

func TestColoniesControllerAddColony(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	controller := createTestColoniesController(db)
	defer controller.Stop()

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := controller.AddColony(colony)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))
}

func TestColoniesControllerAddExecutor(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	controller := createTestColoniesController(db)
	defer controller.Stop()

	executor, _, err := utils.CreateTestExecutorWithKey(core.GenerateRandomID())
	assert.Nil(t, err)

	addedExecutor, err := controller.AddExecutor(executor, false)
	assert.Nil(t, err)
	assert.True(t, executor.Equals(addedExecutor))
}

func TestColoniesControllerAddProcess(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	controller := createTestColoniesController(db)
	defer controller.Stop()

	colonyName := core.GenerateRandomID()

	executor, _, err := utils.CreateTestExecutorWithKey(colonyName)
	assert.Nil(t, err)

	_, err = controller.AddExecutor(executor, false)
	assert.Nil(t, err)

	funcSpec := utils.CreateTestFunctionSpecWithEnv(colonyName, make(map[string]string))
	process := core.CreateProcess(funcSpec)

	addedProcess, err := controller.AddProcess(process)
	assert.Nil(t, err)
	assert.True(t, process.ID == addedProcess.ID)
}

func TestColoniesControllerAssignExecutor(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	controller := createTestColoniesController(db)
	defer controller.Stop()

	colonyName := core.GenerateRandomID()

	executor, _, err := utils.CreateTestExecutorWithKey(colonyName)
	assert.Nil(t, err)

	_, err = controller.AddExecutor(executor, false)
	assert.Nil(t, err)

	funcSpec := utils.CreateTestFunctionSpecWithEnv(colonyName, make(map[string]string))
	process := core.CreateProcess(funcSpec)
	_, err = controller.AddProcess(process)
	assert.Nil(t, err)

	result, err := controller.Assign(executor.ID, colonyName, 0, 0)
	assert.Nil(t, err)
	assert.False(t, result.IsPaused)
	assert.NotNil(t, result.Process)
	assert.True(t, process.ID == result.Process.ID)
}

// notest
func TestColoniesControllerAssignExecutorConcurrency(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_2")
	defer db.Close()
	assert.Nil(t, err)

	processCount := 100

	controller1 := createTestColoniesController(db)
	defer controller1.Stop()
	controller2 := createTestColoniesController2(db)
	defer controller2.Stop()

	colonyName := core.GenerateRandomID()

	executor1, _, err := utils.CreateTestExecutorWithKey(colonyName)
	assert.Nil(t, err)
	_, err = controller1.AddExecutor(executor1, false)
	assert.Nil(t, err)

	executor2, _, err := utils.CreateTestExecutorWithKey(colonyName)
	assert.Nil(t, err)
	_, err = controller1.AddExecutor(executor2, false)
	assert.Nil(t, err)

	for i := 0; i < processCount; i++ {
		funcSpec := utils.CreateTestFunctionSpecWithEnv(colonyName, make(map[string]string))
		process := core.CreateProcess(funcSpec)
		_, err = controller1.AddProcess(process)
		assert.Nil(t, err)
	}

	countChan := make(chan int)

	go func() {
		for {
			result, err := controller1.Assign(executor1.ID, colonyName, 0, 0)
			if err == nil && !result.IsPaused && result.Process != nil {
				countChan <- 1
			}
		}
	}()

	// Since we are using two different controller there should be an error: "Process already assigned"
	// That can happen if two executor clients manage to be assigned the same process
	// A simple solution is just that the second clients gets an error

	go func() {
		for {
			result, err := controller2.Assign(executor2.ID, colonyName, 0, 0)
			if err == nil && !result.IsPaused && result.Process != nil {
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

func TestColoniesControllerPauseResumeAssignments(t *testing.T) {
	controller, _ := createFakeColoniesController()
	defer controller.Stop()

	colonyName := "test_colony"

	// Test pause assignments
	err := controller.PauseColonyAssignments(colonyName)
	assert.NoError(t, err)

	// Test resume assignments
	err = controller.ResumeColonyAssignments(colonyName)
	assert.NoError(t, err)

	// Test check assignments paused status
	paused, err := controller.AreColonyAssignmentsPaused(colonyName)
	assert.NoError(t, err)
	assert.False(t, paused)
}

func TestColoniesControllerPauseResumeAssignmentsWithEtcdServer(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_PAUSE_RESUME")
	defer db.Close()
	assert.Nil(t, err)

	controller := createTestColoniesController(db)
	defer controller.Stop()

	colonyName := "test_colony"

	// Test initial state - should not be paused
	paused, err := controller.AreColonyAssignmentsPaused(colonyName)
	assert.NoError(t, err)
	assert.False(t, paused)

	// Test pause assignments
	err = controller.PauseColonyAssignments(colonyName)
	assert.NoError(t, err)

	// Verify assignments are paused
	paused, err = controller.AreColonyAssignmentsPaused(colonyName)
	assert.NoError(t, err)
	assert.True(t, paused)

	// Test resume assignments
	err = controller.ResumeColonyAssignments(colonyName)
	assert.NoError(t, err)

	// Verify assignments are not paused
	paused, err = controller.AreColonyAssignmentsPaused(colonyName)
	assert.NoError(t, err)
	assert.False(t, paused)
}
