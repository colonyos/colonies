package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateProcess(t *testing.T) {
	colonyID := GenerateRandomID()
	computer1ID := GenerateRandomID()
	computer2ID := GenerateRandomID()
	computerType := "test_computer_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	process := CreateProcess(colonyID, []string{computer1ID, computer2ID}, computerType, timeout, maxRetries, mem, cores, gpus)

	assert.Equal(t, colonyID, process.TargetColonyID())
	assert.Contains(t, process.TargetComputerIDs(), computer1ID)
	assert.Contains(t, process.TargetComputerIDs(), computer2ID)
	assert.Equal(t, computerType, process.ComputerType())
	assert.Equal(t, timeout, process.Timeout())
	assert.Equal(t, maxRetries, process.MaxRetries())
	assert.Equal(t, mem, process.Mem())
	assert.Equal(t, cores, process.Cores())
	assert.Equal(t, gpus, process.GPUs())
	assert.False(t, process.Assigned())
	process.Assign()
	assert.True(t, process.Assigned())
	process.Unassign()
	assert.False(t, process.Assigned())
}

func TestTimeCalc(t *testing.T) {
	colonyID := GenerateRandomID()
	computerType := "test_computer_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	process := CreateProcess(colonyID, []string{}, computerType, timeout, maxRetries, mem, cores, gpus)

	startTime := time.Now()

	process.SetSubmissionTime(startTime)
	process.SetStartTime(startTime.Add(1 * time.Second))
	process.SetEndTime(startTime.Add(4 * time.Second))

	assert.False(t, process.WaitingTime() < 900000000 && process.WaitingTime() > 1200000000)
	assert.False(t, process.WaitingTime() < 3000000000 && process.WaitingTime() > 4000000000)
}
