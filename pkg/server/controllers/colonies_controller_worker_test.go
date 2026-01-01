package controllers

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCleanupStaleExecutors_SkipsZeroLastHeardFrom(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_CLEANUP_ZERO")
	defer db.Close()
	assert.Nil(t, err)

	controller := createTestColoniesController(db)
	defer controller.Stop()

	colonyName := core.GenerateRandomID()
	colony := core.CreateColony(core.GenerateRandomID(), colonyName)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Create an executor with zero LastHeardFromTime (default)
	executor, _, err := utils.CreateTestExecutorWithKey(colonyName)
	assert.Nil(t, err)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	// Verify executor exists
	executors, err := db.GetExecutorsByColonyName(colonyName)
	assert.Nil(t, err)
	assert.Len(t, executors, 1)

	// Run cleanup with a short stale duration - should NOT remove executor with zero LastHeardFromTime
	controller.staleExecutorDuration = 1 * time.Second
	controller.cleanupStaleExecutors()

	// Executor should still exist (not removed because LastHeardFromTime is zero)
	executors, err = db.GetExecutorsByColonyName(colonyName)
	assert.Nil(t, err)
	assert.Len(t, executors, 1, "Executor with zero LastHeardFromTime should not be removed")
}

func TestCleanupStaleExecutors_RemovesStaleExecutor(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_CLEANUP_STALE")
	defer db.Close()
	assert.Nil(t, err)

	controller := createTestColoniesController(db)
	defer controller.Stop()

	colonyName := core.GenerateRandomID()
	colony := core.CreateColony(core.GenerateRandomID(), colonyName)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Create an executor with a LastHeardFromTime in the past
	executor, _, err := utils.CreateTestExecutorWithKey(colonyName)
	assert.Nil(t, err)
	executor.LastHeardFromTime = time.Now().Add(-10 * time.Minute) // 10 minutes ago
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	// Verify executor exists
	executors, err := db.GetExecutorsByColonyName(colonyName)
	assert.Nil(t, err)
	assert.Len(t, executors, 1)

	// Run cleanup with 5 minute stale duration - should remove executor
	controller.staleExecutorDuration = 5 * time.Minute
	controller.cleanupStaleExecutors()

	// Executor should be removed (stale for 10 minutes, threshold is 5 minutes)
	// Note: RemoveExecutorByName marks executor as UNREGISTERED rather than deleting,
	// so we check that no ACTIVE executors remain
	executors, err = db.GetExecutorsByColonyName(colonyName)
	assert.Nil(t, err)
	activeExecutors := 0
	for _, e := range executors {
		if e.State != core.UNREGISTERED {
			activeExecutors++
		}
	}
	assert.Equal(t, 0, activeExecutors, "Stale executor should be marked as UNREGISTERED")
}

func TestCleanupStaleExecutors_KeepsRecentExecutor(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_CLEANUP_RECENT")
	defer db.Close()
	assert.Nil(t, err)

	controller := createTestColoniesController(db)
	defer controller.Stop()

	colonyName := core.GenerateRandomID()
	colony := core.CreateColony(core.GenerateRandomID(), colonyName)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Create an executor with a recent LastHeardFromTime
	executor, _, err := utils.CreateTestExecutorWithKey(colonyName)
	assert.Nil(t, err)
	executor.LastHeardFromTime = time.Now().Add(-1 * time.Minute) // 1 minute ago
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	// Verify executor exists
	executors, err := db.GetExecutorsByColonyName(colonyName)
	assert.Nil(t, err)
	assert.Len(t, executors, 1)

	// Run cleanup with 10 minute stale duration - should keep executor
	controller.staleExecutorDuration = 10 * time.Minute
	controller.cleanupStaleExecutors()

	// Executor should still exist (last heard 1 minute ago, threshold is 10 minutes)
	executors, err = db.GetExecutorsByColonyName(colonyName)
	assert.Nil(t, err)
	assert.Len(t, executors, 1, "Recent executor should not be removed")
}
