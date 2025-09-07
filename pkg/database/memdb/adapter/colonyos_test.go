package adapter

import (
	"fmt"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/database/memdb"
	"github.com/stretchr/testify/assert"
)

func TestColonyOSAdapter_BasicIntegration(t *testing.T) {
	// Create VelocityDB config
	config := &memdb.VelocityConfig{
		DataDir:   "/tmp/velocitydb_test",
		CacheSize: 100,
		InMemory:  true,
	}

	// Create the adapter
	adapter, err := NewColonyOSAdapter(config)
	assert.NoError(t, err)
	defer adapter.Close()

	// Verify it implements the Database interface
	var db database.Database = adapter
	assert.NotNil(t, db)

	// Test Colony operations
	colony := &core.Colony{
		ID:   "test-colony-id",
		Name: "test-colony",
	}

	// Add a colony
	err = adapter.AddColony(colony)
	assert.NoError(t, err)

	// Get colony by ID
	retrievedColony, err := adapter.GetColonyByID("test-colony-id")
	assert.NoError(t, err)
	assert.Equal(t, "test-colony-id", retrievedColony.ID)
	assert.Equal(t, "test-colony", retrievedColony.Name)

	// Get colony by name
	retrievedColony2, err := adapter.GetColonyByName("test-colony")
	assert.NoError(t, err)
	assert.Equal(t, "test-colony-id", retrievedColony2.ID)
	assert.Equal(t, "test-colony", retrievedColony2.Name)

	// Count colonies
	count, err := adapter.CountColonies()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Test Process operations
	process := &core.Process{
		ID:             "test-process-id",
		InitiatorID:    "test-initiator",
		InitiatorName:  "test-initiator-name",
		State:          core.WAITING,
		SubmissionTime: time.Now(),
		FunctionSpec: core.FunctionSpec{
			FuncName: "test-function",
			Conditions: core.Conditions{
				ColonyName: "test-colony",
			},
		},
	}

	// Add a process
	err = adapter.AddProcess(process)
	assert.NoError(t, err)

	// Get process by ID
	retrievedProcess, err := adapter.GetProcessByID("test-process-id")
	assert.NoError(t, err)
	assert.Equal(t, "test-process-id", retrievedProcess.ID)
	assert.Equal(t, core.WAITING, retrievedProcess.State)
	assert.Equal(t, "test-function", retrievedProcess.FunctionSpec.FuncName)

	// Test process assignment using CAS
	executor := &core.Executor{
		ID:         "test-executor-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
		Type:       "test-type",
		State:      core.APPROVED,
	}

	// Add executor
	err = adapter.AddExecutor(executor)
	assert.NoError(t, err)

	// Assign process to executor
	err = adapter.Assign("test-executor-id", process)
	assert.NoError(t, err)

	// Verify assignment
	assignedProcess, err := adapter.GetProcessByID("test-process-id")
	assert.NoError(t, err)
	assert.Equal(t, "test-executor-id", assignedProcess.AssignedExecutorID)
	assert.Equal(t, true, assignedProcess.IsAssigned)
	assert.Equal(t, core.RUNNING, assignedProcess.State)

	// Mark process as successful
	_, _, err = adapter.MarkSuccessful("test-process-id")
	assert.NoError(t, err)

	// Verify success
	successfulProcess, err := adapter.GetProcessByID("test-process-id")
	assert.NoError(t, err)
	assert.Equal(t, core.SUCCESS, successfulProcess.State)

	// Test executor operations
	retrievedExecutor, err := adapter.GetExecutorByID("test-executor-id")
	assert.NoError(t, err)
	assert.Equal(t, "test-executor-id", retrievedExecutor.ID)
	assert.Equal(t, "test-executor", retrievedExecutor.Name)
	assert.Equal(t, "test-colony", retrievedExecutor.ColonyName)

	// Count executors
	executorCount, err := adapter.CountExecutors()
	assert.NoError(t, err)
	assert.Equal(t, 1, executorCount)
}

func TestColonyOSAdapter_ProcessSearchAndFiltering(t *testing.T) {
	config := &memdb.VelocityConfig{
		DataDir:   "/tmp/velocitydb_search_test",
		CacheSize: 100,
		InMemory:  true,
	}

	adapter, err := NewColonyOSAdapter(config)
	assert.NoError(t, err)
	defer adapter.Close()

	// Add test colony
	colony := &core.Colony{
		ID:   "search-colony-id",
		Name: "search-colony",
	}
	err = adapter.AddColony(colony)
	assert.NoError(t, err)

	// Add multiple processes with different states
	processes := []*core.Process{
		{
			ID:    "waiting-1",
			State: core.WAITING,
			FunctionSpec: core.FunctionSpec{
				Conditions: core.Conditions{ColonyName: "search-colony"},
			},
		},
		{
			ID:    "waiting-2", 
			State: core.WAITING,
			FunctionSpec: core.FunctionSpec{
				Conditions: core.Conditions{ColonyName: "search-colony"},
			},
		},
		{
			ID:    "running-1",
			State: core.RUNNING,
			FunctionSpec: core.FunctionSpec{
				Conditions: core.Conditions{ColonyName: "search-colony"},
			},
		},
		{
			ID:    "success-1",
			State: core.SUCCESS,
			FunctionSpec: core.FunctionSpec{
				Conditions: core.Conditions{ColonyName: "search-colony"},
			},
		},
	}

	// Add all processes
	for _, process := range processes {
		err = adapter.AddProcess(process)
		assert.NoError(t, err)
	}

	// Test filtering by state
	waitingProcesses, err := adapter.FindWaitingProcesses("search-colony", "", "", "", 10)
	assert.NoError(t, err)
	assert.Len(t, waitingProcesses, 2)

	runningProcesses, err := adapter.FindRunningProcesses("search-colony", "", "", "", 10)
	assert.NoError(t, err) 
	assert.Len(t, runningProcesses, 1)

	successProcesses, err := adapter.FindSuccessfulProcesses("search-colony", "", "", "", 10)
	assert.NoError(t, err)
	assert.Len(t, successProcesses, 1)

	// Test counting by state
	waitingCount, err := adapter.CountWaitingProcessesByColonyName("search-colony")
	assert.NoError(t, err)
	assert.Equal(t, 2, waitingCount)

	runningCount, err := adapter.CountRunningProcessesByColonyName("search-colony")
	assert.NoError(t, err)
	assert.Equal(t, 1, runningCount)

	successCount, err := adapter.CountSuccessfulProcessesByColonyName("search-colony")
	assert.NoError(t, err)
	assert.Equal(t, 1, successCount)
}

func TestColonyOSAdapter_ConcurrentProcessAssignment(t *testing.T) {
	config := &memdb.VelocityConfig{
		DataDir:   "/tmp/velocitydb_cas_test",
		CacheSize: 100,
		InMemory:  true,
	}

	adapter, err := NewColonyOSAdapter(config)
	assert.NoError(t, err)
	defer adapter.Close()

	// Create a process for assignment
	process := &core.Process{
		ID:                 "cas-test-process",
		State:              core.WAITING,
		AssignedExecutorID: "",
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{ColonyName: "cas-colony"},
		},
	}

	err = adapter.AddProcess(process)
	assert.NoError(t, err)

	// First assignment should succeed
	err = adapter.Assign("executor-1", process)
	assert.NoError(t, err)

	// Verify first assignment
	assignedProcess, err := adapter.GetProcessByID("cas-test-process")
	assert.NoError(t, err)
	assert.Equal(t, "executor-1", assignedProcess.AssignedExecutorID)
	assert.Equal(t, true, assignedProcess.IsAssigned)
	assert.Equal(t, core.RUNNING, assignedProcess.State)

	// Second assignment to same process should fail due to CAS
	err = adapter.Assign("executor-2", process)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already assigned")

	// Verify process is still assigned to executor-1
	stillAssignedProcess, err := adapter.GetProcessByID("cas-test-process")
	assert.NoError(t, err)
	assert.Equal(t, "executor-1", stillAssignedProcess.AssignedExecutorID)
}

func BenchmarkColonyOSAdapter_ProcessOperations(b *testing.B) {
	config := &memdb.VelocityConfig{
		DataDir:   "/tmp/velocitydb_bench",
		CacheSize: 100,
		InMemory:  true,
	}

	adapter, err := NewColonyOSAdapter(config)
	if err != nil {
		b.Fatal(err)
	}
	defer adapter.Close()

	// Add test colony
	colony := &core.Colony{ID: "bench-colony", Name: "bench"}
	adapter.AddColony(colony)

	b.ResetTimer()

	b.Run("AddProcess", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			process := &core.Process{
				ID:    fmt.Sprintf("process-%d", i),
				State: core.WAITING,
				FunctionSpec: core.FunctionSpec{
					Conditions: core.Conditions{ColonyName: "bench-colony"},
				},
			}
			adapter.AddProcess(process)
		}
	})

	b.Run("GetProcessByID", func(b *testing.B) {
		// Pre-add a process
		process := &core.Process{
			ID:    "get-test-process",
			State: core.WAITING,
			FunctionSpec: core.FunctionSpec{
				Conditions: core.Conditions{ColonyName: "bench-colony"},
			},
		}
		adapter.AddProcess(process)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			adapter.GetProcessByID("get-test-process")
		}
	})
}