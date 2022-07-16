package basic

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestSelectProcess(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := utils.CreateTestProcess(colony.ID)
	process1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))

	process2 := utils.CreateTestProcess(colony.ID)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	process3 := utils.CreateTestProcess(colony.ID)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Process{process1, process2, process3}

	planner := CreatePlanner()
	selectedProcess, err := planner.Select("runtimeid_1", candidates, false)
	assert.Nil(t, err)
	assert.NotNil(t, selectedProcess)
	assert.Equal(t, selectedProcess.ID, process2.ID)

	selectedProcess, err = planner.Select("runtimeid_1", candidates, true)
	assert.Nil(t, err)
	assert.NotNil(t, selectedProcess)
	assert.Equal(t, selectedProcess.ID, process1.ID)
}

func TestSelectProcess2(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := utils.CreateTestProcess(colony.ID)
	process1.SetSubmissionTime(startTime.Add(60 * time.Millisecond))

	process2 := utils.CreateTestProcess(colony.ID)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	process3 := utils.CreateTestProcess(colony.ID)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Process{process1, process2, process3}

	planner := CreatePlanner()
	selectedProcess, err := planner.Select("runtimeid_1", candidates, false)
	assert.Nil(t, err)
	assert.Equal(t, selectedProcess.ID, process1.ID)
}

func TestSelectProcessSameSubmissionTimes(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := utils.CreateTestProcess(colony.ID)
	process1.SetSubmissionTime(startTime)

	process2 := utils.CreateTestProcess(colony.ID)
	process2.SetSubmissionTime(startTime)

	process3 := utils.CreateTestProcess(colony.ID)
	process3.SetSubmissionTime(startTime)

	candidates := []*core.Process{process1, process2, process3}

	planner := CreatePlanner()
	selectedProcess, err := planner.Select("runtimeid_1", candidates, false)
	assert.Nil(t, err)
	assert.Equal(t, selectedProcess.ID, process1.ID)
}

func TestSelectProcessNoProcesss(t *testing.T) {
	candidates := []*core.Process{}

	planner := CreatePlanner()
	selectedProcess, err := planner.Select("runtimeid_1", candidates, false)
	assert.NotNil(t, err)
	assert.Nil(t, selectedProcess)
}

func TestSelectProccess5(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := utils.CreateTestProcessWithTargets(colony.ID, []string{"runtimeid_2"})
	process1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))

	process2 := utils.CreateTestProcessWithTargets(colony.ID, []string{"runtimeid_2"})
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	process3 := utils.CreateTestProcessWithTargets(colony.ID, []string{"runtimeid_1"})
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Process{process1, process2, process3}

	planner := CreatePlanner()
	selectedProcess, err := planner.Select("runtimeid_1", candidates, false)
	assert.Nil(t, err)
	assert.Equal(t, selectedProcess.ID, process3.ID)
}

func TestPrioritize(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := utils.CreateTestProcess(colony.ID)
	process1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))

	process2 := utils.CreateTestProcess(colony.ID)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	process3 := utils.CreateTestProcess(colony.ID)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Process{process1, process2, process3}

	planner := CreatePlanner()
	prioritizedProcesses := planner.Prioritize("runtimeid_1", candidates, 3, false)
	assert.Len(t, prioritizedProcesses, 3)

	assert.Equal(t, process2.ID, prioritizedProcesses[0].ID)
	assert.Equal(t, process3.ID, prioritizedProcesses[1].ID)
	assert.Equal(t, process1.ID, prioritizedProcesses[2].ID)

	prioritizedProcesses = planner.Prioritize("runtimeid_1", candidates, 2, false)
	assert.Len(t, prioritizedProcesses, 2)

	assert.Equal(t, process2.ID, prioritizedProcesses[0].ID)
	assert.Equal(t, process3.ID, prioritizedProcesses[1].ID)

	prioritizedProcesses = planner.Prioritize("runtimeid_1", candidates, 2, true)
	assert.Len(t, prioritizedProcesses, 2)

	assert.Equal(t, process1.ID, prioritizedProcesses[0].ID)
	assert.Equal(t, process3.ID, prioritizedProcesses[1].ID)
}

func TestPrioritize2(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := utils.CreateTestProcessWithTargets(colony.ID, []string{"runtimeid_1"})
	process1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))

	process2 := utils.CreateTestProcessWithTargets(colony.ID, []string{"runtimeid_1"})
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	process3 := utils.CreateTestProcess(colony.ID)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Process{process1, process2, process3}

	// In the scenario above, there is only possible proceess that runtimeid_2 can get, hence we should get 1 process
	// altought we are asking for 3 processes, this basically tests the min function in basic_planner.go
	planner := CreatePlanner()
	prioritizedProcesses := planner.Prioritize("runtimeid_2", candidates, 3, false)
	assert.Len(t, prioritizedProcesses, 1)
}
