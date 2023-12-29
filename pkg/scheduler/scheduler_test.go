package scheduler

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestSelectProcess(t *testing.T) {
	startTime := time.Now()

	mock := createProcessLookupMock()
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	executor := utils.CreateTestExecutor(colony.Name)
	executor.Name = "executor1"

	process1 := utils.CreateTestProcess(colony.Name)
	process1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))
	mock.addProcess(process1)

	process2 := utils.CreateTestProcess(colony.Name)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))
	mock.addProcess(process2)

	process3 := utils.CreateTestProcess(colony.Name)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))
	mock.addProcess(process3)

	s := CreateScheduler(mock)
	selectedProcess, err := s.Select(colony.Name, executor)
	assert.Nil(t, err)
	assert.NotNil(t, selectedProcess)
	assert.Equal(t, selectedProcess.ID, process2.ID)
}

func TestSelectProcess2(t *testing.T) {
	startTime := time.Now()

	mock := createProcessLookupMock()
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	executor := utils.CreateTestExecutor(colony.Name)
	executor.Name = "executor1"

	process1 := utils.CreateTestProcess(colony.Name)
	process1.SetSubmissionTime(startTime.Add(60 * time.Millisecond))
	mock.addProcess(process1)

	process2 := utils.CreateTestProcess(colony.Name)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))
	mock.addProcess(process2)

	process3 := utils.CreateTestProcess(colony.Name)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))
	mock.addProcess(process3)

	s := CreateScheduler(mock)
	selectedProcess, err := s.Select(colony.Name, executor)
	assert.Nil(t, err)
	assert.Equal(t, selectedProcess.ID, process1.ID)
}

func TestSelectProcessNoProcesss(t *testing.T) {
	mock := createProcessLookupMock()
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	executor := utils.CreateTestExecutor(colony.Name)
	executor.Name = "executor1"

	s := CreateScheduler(mock)
	selectedProcess, err := s.Select(colony.Name, executor)
	assert.NotNil(t, err)
	assert.Nil(t, selectedProcess)
}

func TestPrioritize(t *testing.T) {
	startTime := time.Now()

	mock := createProcessLookupMock()
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	executor := utils.CreateTestExecutor(colony.Name)
	executor.Name = "executor1"

	process1 := utils.CreateTestProcess(colony.Name)
	process1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))
	mock.addProcess(process1)

	process2 := utils.CreateTestProcess(colony.Name)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))
	mock.addProcess(process2)

	process3 := utils.CreateTestProcess(colony.Name)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))
	mock.addProcess(process3)

	s := CreateScheduler(mock)
	prioritizedProcesses, err := s.Prioritize(colony.Name, executor, 3)
	assert.Nil(t, err)
	assert.Len(t, prioritizedProcesses, 3)

	assert.Equal(t, process2.ID, prioritizedProcesses[0].ID)
	assert.Equal(t, process3.ID, prioritizedProcesses[1].ID)
	assert.Equal(t, process1.ID, prioritizedProcesses[2].ID)

	prioritizedProcesses, err = s.Prioritize(colony.Name, executor, 2)
	assert.Nil(t, err)
	assert.Len(t, prioritizedProcesses, 2)

	assert.Equal(t, process2.ID, prioritizedProcesses[0].ID)
	assert.Equal(t, process3.ID, prioritizedProcesses[1].ID)
}

func TestPrioritizeByName(t *testing.T) {
	startTime := time.Now()

	mock := createProcessLookupMock()
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	executor1 := utils.CreateTestExecutor(colony.Name)
	executor1.Name = "executor1"

	executor2 := utils.CreateTestExecutor(colony.Name)
	executor2.Name = "executor2"

	process1 := utils.CreateTestProcess(colony.Name)
	process1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))
	process1.FunctionSpec.Conditions.ExecutorNames = []string{"executor1"}
	mock.addProcess(process1)

	process2 := utils.CreateTestProcess(colony.Name)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))
	process2.FunctionSpec.Conditions.ExecutorNames = []string{"executor2"}
	mock.addProcess(process2)

	process3 := utils.CreateTestProcess(colony.Name)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))
	mock.addProcess(process3)

	s := CreateScheduler(mock)
	prioritizedProcesses, err := s.Prioritize(colony.Name, executor1, 3)
	assert.Nil(t, err)
	assert.Len(t, prioritizedProcesses, 2)
	assert.Equal(t, prioritizedProcesses[0].ID, process3.ID)
	assert.Equal(t, prioritizedProcesses[1].ID, process1.ID)

	prioritizedProcesses, err = s.Prioritize(colony.Name, executor2, 3)
	assert.Nil(t, err)
	assert.Len(t, prioritizedProcesses, 2)
	assert.Equal(t, prioritizedProcesses[0].ID, process2.ID)
	assert.Equal(t, prioritizedProcesses[1].ID, process3.ID)
}
