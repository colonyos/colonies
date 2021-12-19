package basic

import (
	"colonies/pkg/core"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateProcess(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	process1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))

	process2 := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	process3 := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Process{process1, process2, process3}

	scheduler := CreateScheduler()
	selectedProcess := scheduler.Select("computerid_1", candidates)
	assert.Equal(t, selectedProcess.ID(), process1.ID())
}

func TestCreateProcess2(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	process1.SetSubmissionTime(startTime.Add(60 * time.Millisecond))

	process2 := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	process3 := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Process{process1, process2, process3}

	scheduler := CreateScheduler()
	selectedProcess := scheduler.Select("computerid_1", candidates)
	assert.Equal(t, selectedProcess.ID(), process3.ID())
}

func TestCreateProcessSameSubmissionTimes(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	process1.SetSubmissionTime(startTime)

	process2 := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	process2.SetSubmissionTime(startTime)

	process3 := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	process3.SetSubmissionTime(startTime)

	candidates := []*core.Process{process1, process2, process3}

	scheduler := CreateScheduler()
	selectedProcess := scheduler.Select("computerid_1", candidates)
	assert.Equal(t, selectedProcess.ID(), process1.ID())
}

func TestCreateProcessNoProcesss(t *testing.T) {
	candidates := []*core.Process{}

	scheduler := CreateScheduler()
	selectedProcess := scheduler.Select("computerid_1", candidates)
	assert.Nil(t, selectedProcess)
}

func TestCreateProcess5(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	process1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))

	process2 := core.CreateProcess(colony.ID(), []string{"computerid_1"}, "dummy", -1, 3, 1000, 10, 1)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	process3 := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Process{process1, process2, process3}

	scheduler := CreateScheduler()
	selectedProcess := scheduler.Select("computerid_1", candidates)
	assert.Equal(t, selectedProcess.ID(), process2.ID())
}
