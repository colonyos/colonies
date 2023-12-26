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

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process1 := utils.CreateTestProcess(colony.ID)
	process1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))

	process2 := utils.CreateTestProcess(colony.ID)
	process2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	process3 := utils.CreateTestProcess(colony.ID)
	process3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Process{process1, process2, process3}

	p := CreatePrioritizer()
	selectedProcess, err := p.Select("executorid_1", candidates)
	assert.Nil(t, err)
	assert.NotNil(t, selectedProcess)
	assert.Equal(t, selectedProcess.ID, process2.ID)
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

	p := CreatePrioritizer()
	selectedProcess, err := p.Select("executorid_1", candidates)
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

	p := CreatePrioritizer()
	selectedProcess, err := p.Select("executorid_1", candidates)
	assert.Nil(t, err)
	assert.Equal(t, selectedProcess.ID, process1.ID)
}

func TestSelectProcessNoProcesss(t *testing.T) {
	candidates := []*core.Process{}

	p := CreatePrioritizer()
	selectedProcess, err := p.Select("executorid_1", candidates)
	assert.NotNil(t, err)
	assert.Nil(t, selectedProcess)
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

	p := CreatePrioritizer()
	prioritizedProcesses := p.Prioritize("executorid_1", candidates, 3)
	assert.Len(t, prioritizedProcesses, 3)

	assert.Equal(t, process2.ID, prioritizedProcesses[0].ID)
	assert.Equal(t, process3.ID, prioritizedProcesses[1].ID)
	assert.Equal(t, process1.ID, prioritizedProcesses[2].ID)

	prioritizedProcesses = p.Prioritize("executorid_1", candidates, 2)
	assert.Len(t, prioritizedProcesses, 2)

	assert.Equal(t, process2.ID, prioritizedProcesses[0].ID)
	assert.Equal(t, process3.ID, prioritizedProcesses[1].ID)
}
