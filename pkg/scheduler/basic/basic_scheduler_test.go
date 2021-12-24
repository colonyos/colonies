package basic

import (
	"colonies/pkg/core"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSelectProcess(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := core.CreateProcess(colony.ID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	process1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))

	process2 := core.CreateProcess(colony.ID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	process3 := core.CreateProcess(colony.ID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Process{process1, process2, process3}

	scheduler := CreateScheduler()
	selectedProcess, err := scheduler.Select("computerid_1", candidates)
	assert.Nil(t, err)
	assert.NotNil(t, selectedProcess)
	assert.Equal(t, selectedProcess.ID, process2.ID)
}

func TestSelectProcess2(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := core.CreateProcess(colony.ID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	process1.SetSubmissionTime(startTime.Add(60 * time.Millisecond))

	process2 := core.CreateProcess(colony.ID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	process3 := core.CreateProcess(colony.ID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Process{process1, process2, process3}

	scheduler := CreateScheduler()
	selectedProcess, err := scheduler.Select("computerid_1", candidates)
	assert.Nil(t, err)
	assert.Equal(t, selectedProcess.ID, process1.ID)
}

func TestSelectProcessSameSubmissionTimes(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := core.CreateProcess(colony.ID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	process1.SetSubmissionTime(startTime)

	process2 := core.CreateProcess(colony.ID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	process2.SetSubmissionTime(startTime)

	process3 := core.CreateProcess(colony.ID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	process3.SetSubmissionTime(startTime)

	candidates := []*core.Process{process1, process2, process3}

	scheduler := CreateScheduler()
	selectedProcess, err := scheduler.Select("computerid_1", candidates)
	assert.Nil(t, err)
	assert.Equal(t, selectedProcess.ID, process1.ID)
}

func TestSelectProcessNoProcesss(t *testing.T) {
	candidates := []*core.Process{}

	scheduler := CreateScheduler()
	selectedProcess, err := scheduler.Select("computerid_1", candidates)
	assert.NotNil(t, err)
	assert.Nil(t, selectedProcess)
}

func TestSelectProccess5(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := core.CreateProcess(colony.ID, []string{"computerid_2"}, "dummy", -1, 3, 1000, 10, 1)
	process1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))

	process2 := core.CreateProcess(colony.ID, []string{"computerid_2"}, "dummy", -1, 3, 1000, 10, 1)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	process3 := core.CreateProcess(colony.ID, []string{"computerid_1"}, "dummy", -1, 3, 1000, 10, 1)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Process{process1, process2, process3}

	scheduler := CreateScheduler()
	selectedProcess, err := scheduler.Select("computerid_1", candidates)
	assert.Nil(t, err)
	assert.Equal(t, selectedProcess.ID, process3.ID)
}

func TestPrioritize(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := core.CreateProcess(colony.ID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	process1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))

	process2 := core.CreateProcess(colony.ID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	process3 := core.CreateProcess(colony.ID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Process{process1, process2, process3}

	scheduler := CreateScheduler()
	prioritizedProcesses := scheduler.Prioritize("computerid_1", candidates, 3)
	assert.Len(t, prioritizedProcesses, 3)

	assert.Equal(t, process2.ID, prioritizedProcesses[0].ID)
	assert.Equal(t, process3.ID, prioritizedProcesses[1].ID)
	assert.Equal(t, process1.ID, prioritizedProcesses[2].ID)

	prioritizedProcesses = scheduler.Prioritize("computerid_1", candidates, 2)
	assert.Len(t, prioritizedProcesses, 2)

	assert.Equal(t, process2.ID, prioritizedProcesses[0].ID)
	assert.Equal(t, process3.ID, prioritizedProcesses[1].ID)
}

func TestPrioritize2(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := core.CreateProcess(colony.ID, []string{"computerid_1"}, "dummy", -1, 3, 1000, 10, 1)
	process1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))

	process2 := core.CreateProcess(colony.ID, []string{"computerid_1"}, "dummy", -1, 3, 1000, 10, 1)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	process3 := core.CreateProcess(colony.ID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Process{process1, process2, process3}

	// In the scenario above, there is only possible proceess that computerid_2 can get, hence we should get 1 process
	// altought we are asking for 3 processes, this basically tests the min function in basic_scheduler.go
	scheduler := CreateScheduler()
	prioritizedProcesses := scheduler.Prioritize("computerid_2", candidates, 3)
	assert.Len(t, prioritizedProcesses, 1)
}
