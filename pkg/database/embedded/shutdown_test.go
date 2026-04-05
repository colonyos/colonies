package embedded

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

// TestGracefulShutdownPersistsState verifies that all pending state changes
// survive a Close() + reopen cycle. This is the scenario where Ctrl+C triggers
// db.Close() before the process exits.
func TestGracefulShutdownPersistsState(t *testing.T) {
	dir := t.TempDir()

	// Phase 1: create DB, add executor, mark it UNREGISTERED, close gracefully
	db := CreateEmbeddedDatabase(dir)
	assert.NoError(t, db.Initialize())

	colony := core.CreateColony(core.GenerateRandomID(), "test-colony")
	assert.NoError(t, db.AddColony(colony))

	executor := &core.Executor{
		ID:         core.GenerateRandomID(),
		Name:       "test-executor",
		ColonyName: "test-colony",
		Type:       "test",
		State:      core.PENDING,
	}
	assert.NoError(t, db.AddExecutor(executor))
	assert.NoError(t, db.ApproveExecutor(executor))

	// Mark as unregistered (this is what happens on executor shutdown)
	assert.NoError(t, db.RemoveExecutorByName("test-colony", "test-executor"))

	// Also add a metric to verify it persists
	m := core.CreateMetric("test-colony", "test-executor", "tokens", core.COUNTER, 0)
	assert.NoError(t, db.SetMetric(m))
	assert.NoError(t, db.IncrementMetric("test-colony", "test-executor", "tokens", core.PERIOD_NONE, time.Time{}, 500))

	// Graceful shutdown
	db.Close()

	// Phase 2: reopen and verify state persisted
	db2 := CreateEmbeddedDatabase(dir)
	assert.NoError(t, db2.Initialize())
	defer db2.Close()

	// Executor should be UNREGISTERED, not APPROVED
	e, err := db2.GetExecutorByName("test-colony", "test-executor")
	assert.NoError(t, err)
	assert.NotNil(t, e)
	assert.Equal(t, core.UNREGISTERED, e.State)

	// Metric should have persisted
	metric, err := db2.GetMetric("test-colony", "test-executor", "tokens", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, 500.0, metric.Value)
}

// TestGracefulShutdownProcessState verifies that process state transitions
// persist across a shutdown/restart cycle.
func TestGracefulShutdownProcessState(t *testing.T) {
	dir := t.TempDir()

	db := CreateEmbeddedDatabase(dir)
	assert.NoError(t, db.Initialize())

	colony := core.CreateColony(core.GenerateRandomID(), "test-colony")
	assert.NoError(t, db.AddColony(colony))

	executor := &core.Executor{
		ID:         core.GenerateRandomID(),
		Name:       "worker",
		ColonyName: "test-colony",
		Type:       "test",
		State:      core.PENDING,
	}
	assert.NoError(t, db.AddExecutor(executor))
	assert.NoError(t, db.ApproveExecutor(executor))

	// Submit a process, assign it, mark successful
	process := &core.Process{
		ID: core.GenerateRandomID(),
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{
				ColonyName:   "test-colony",
				ExecutorType: "test",
			},
		},
	}
	assert.NoError(t, db.AddProcess(process))
	assert.NoError(t, db.Assign(executor.ID, process))
	_, _, err := db.MarkSuccessful(process.ID)
	assert.NoError(t, err)

	db.Close()

	// Reopen
	db2 := CreateEmbeddedDatabase(dir)
	assert.NoError(t, db2.Initialize())
	defer db2.Close()

	p, err := db2.GetProcessByID(process.ID)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, core.SUCCESS, p.State)
	assert.Equal(t, executor.ID, p.AssignedExecutorID)
}

// TestReregisterAfterRestart verifies the exact bug scenario:
// executor registers, unregisters, server restarts, executor re-registers.
// Should end up with exactly 1 executor, not duplicates.
func TestReregisterAfterRestart(t *testing.T) {
	dir := t.TempDir()

	// Phase 1: executor registers and unregisters
	db := CreateEmbeddedDatabase(dir)
	assert.NoError(t, db.Initialize())

	colony := core.CreateColony(core.GenerateRandomID(), "test-colony")
	assert.NoError(t, db.AddColony(colony))

	executor := &core.Executor{
		ID:         core.GenerateRandomID(),
		Name:       "llm-executor",
		ColonyName: "test-colony",
		Type:       "llm",
		State:      core.PENDING,
	}
	assert.NoError(t, db.AddExecutor(executor))
	assert.NoError(t, db.ApproveExecutor(executor))
	assert.NoError(t, db.RemoveExecutorByName("test-colony", "llm-executor"))

	db.Close()

	// Phase 2: server restarts, executor re-registers with new ID
	db2 := CreateEmbeddedDatabase(dir)
	assert.NoError(t, db2.Initialize())
	defer db2.Close()

	newExecutor := &core.Executor{
		ID:         core.GenerateRandomID(),
		Name:       "llm-executor",
		ColonyName: "test-colony",
		Type:       "llm",
		State:      core.PENDING,
	}
	assert.NoError(t, db2.AddExecutor(newExecutor))
	assert.NoError(t, db2.ApproveExecutor(newExecutor))

	// Should be exactly 1 executor, not 2
	executors, err := db2.GetExecutorsByColonyName("test-colony", false)
	assert.NoError(t, err)
	assert.Len(t, executors, 1)
	assert.Equal(t, newExecutor.ID, executors[0].ID)
	assert.Equal(t, core.APPROVED, executors[0].State)
}
