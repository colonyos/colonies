package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateTask(t *testing.T) {
	colonyID := GenerateRandomID()
	worker1ID := GenerateRandomID()
	worker2ID := GenerateRandomID()
	workerType := "test_worker_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	task := CreateTask(colonyID, []string{worker1ID, worker2ID}, workerType, timeout, maxRetries, mem, cores, gpus)

	assert.Equal(t, colonyID, task.TargetColonyID())
	assert.Contains(t, task.TargetWorkerIDs(), worker1ID)
	assert.Contains(t, task.TargetWorkerIDs(), worker2ID)
	assert.Equal(t, workerType, task.WorkerType())
	assert.Equal(t, timeout, task.Timeout())
	assert.Equal(t, maxRetries, task.MaxRetries())
	assert.Equal(t, mem, task.Mem())
	assert.Equal(t, cores, task.Cores())
	assert.Equal(t, gpus, task.GPUs())
	assert.False(t, task.Assigned())
	task.Assign()
	assert.True(t, task.Assigned())
	task.Unassign()
	assert.False(t, task.Assigned())
}

func TestTimeCalc(t *testing.T) {
	colonyID := GenerateRandomID()
	workerType := "test_worker_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	task := CreateTask(colonyID, []string{}, workerType, timeout, maxRetries, mem, cores, gpus)

	startTime := time.Now()

	task.SetSubmissionTime(startTime)
	task.SetStartTime(startTime.Add(1 * time.Second))
	task.SetEndTime(startTime.Add(4 * time.Second))

	assert.False(t, task.WaitingTime() < 900000000 && task.WaitingTime() > 1200000000)
	assert.False(t, task.WaitingTime() < 3000000000 && task.WaitingTime() > 4000000000)
}
