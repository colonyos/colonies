package embedded

// Bug report: Duplicate executor registrations in embedded database
//
// Symptom:
//   `colonies executor ls` shows multiple executors with the same name after
//   restarting the exec binary. Over time, dozens of duplicate registrations
//   accumulate and cannot be cleaned up — RemoveExecutor always returns success
//   even on already-removed executors, causing cleanup loops to run forever.
//
// Root cause:
//   The embedded database had two bugs that together caused unbounded duplication:
//
//   1. Race condition in AddExecutor (concurrent re-registration)
//      The HandleAddExecutor handler performs a non-atomic read-remove-add sequence:
//        a) GetExecutorByName → finds existing executor (state=APPROVED)
//        b) RemoveExecutorByName → marks it UNREGISTERED
//        c) AddExecutor → adds new executor with same name
//      Without a mutex, concurrent calls interleave: both threads read the same
//      APPROVED executor, both mark it UNREGISTERED, then both add a new one.
//      The byColony index (a MapIndex/set) happily stores multiple executor IDs
//      for the same colony, resulting in duplicates visible in the listing.
//
//      In PostgreSQL this cannot happen because the NAME column is a PRIMARY KEY
//      (`colonyname:executorname`), so only one row per name can ever exist.
//      The embedded DB had no equivalent constraint.
//
//   2. No uniqueness enforcement in MapIndex
//      The byName index maps `"colony:name"` → set of executor IDs. The MapIndex.Add
//      method appends to the set without checking cardinality. Multiple executor IDs
//      could be associated with the same name, violating the one-executor-per-name
//      invariant that PostgreSQL enforces via PRIMARY KEY.
//
// Fix:
//   Added `executorMu sync.Mutex` to EmbeddedDatabase. Both AddExecutor and
//   RemoveExecutorByName acquire this mutex, making the check-and-modify sequence
//   atomic. This matches the transactional semantics that PostgreSQL provides
//   implicitly.
//
// Note on RemoveExecutorByName behavior:
//   Both PostgreSQL and embedded DB mark executors as UNREGISTERED rather than
//   deleting them (for traceability). Calling RemoveExecutorByName on an already
//   UNREGISTERED executor is a no-op that returns nil — this matches PostgreSQL
//   where the UPDATE sets state=UNREGISTERED on a row that's already UNREGISTERED.
//   Client-side cleanup loops must check GetExecutorsByColonyName (which filters
//   out UNREGISTERED) rather than relying on RemoveExecutor returning an error.

import (
	"sync"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

// RemoveExecutorByName on an already-UNREGISTERED executor is a no-op (matches PostgreSQL).
// The executor stays UNREGISTERED and can be re-registered via AddExecutor.
func TestRemoveUnregisteredExecutorIsNoOp(t *testing.T) {
	db := setupTestDB(t)

	colony := core.CreateColony(core.GenerateRandomID(), "test-colony")
	err := db.AddColony(colony)
	assert.NoError(t, err)

	executor := core.CreateExecutor(core.GenerateRandomID(), "test-type", "my-executor", colony.Name, time.Now(), time.Now())
	err = db.AddExecutor(executor)
	assert.NoError(t, err)

	// First remove should succeed
	err = db.RemoveExecutorByName(colony.Name, "my-executor")
	assert.NoError(t, err)

	// Verify it's gone from visible listing
	executors, err := db.GetExecutorsByColonyName(colony.Name, false)
	assert.NoError(t, err)
	assert.Len(t, executors, 0)

	// Second remove succeeds (no-op, matches PostgreSQL behavior)
	err = db.RemoveExecutorByName(colony.Name, "my-executor")
	assert.NoError(t, err)

	// Can still re-register after double remove
	newExec := core.CreateExecutor(core.GenerateRandomID(), "test-type", "my-executor", colony.Name, time.Now(), time.Now())
	err = db.AddExecutor(newExec)
	assert.NoError(t, err)

	executors, err = db.GetExecutorsByColonyName(colony.Name, false)
	assert.NoError(t, err)
	assert.Len(t, executors, 1)
}

// Bug 2: Concurrent AddExecutor with AllowReregister pattern creates duplicates.
// The handler does: GetExecutorByName → RemoveExecutorByName → AddExecutor
// Without atomicity, two concurrent calls can both succeed and create duplicates.
func TestConcurrentReregisterNoDuplicates(t *testing.T) {
	db := setupTestDB(t)

	colony := core.CreateColony(core.GenerateRandomID(), "test-colony")
	err := db.AddColony(colony)
	assert.NoError(t, err)

	name := "contested-executor"

	// Add initial executor
	executor := core.CreateExecutor(core.GenerateRandomID(), "test-type", name, colony.Name, time.Now(), time.Now())
	err = db.AddExecutor(executor)
	assert.NoError(t, err)
	err = db.ApproveExecutor(executor)
	assert.NoError(t, err)

	// Simulate 10 concurrent re-registrations
	// Each goroutine does what HandleAddExecutor does:
	// 1. GetExecutorByName → find existing
	// 2. RemoveExecutorByName → mark as UNREGISTERED
	// 3. AddExecutor → add new one (embedded DB handles UNREGISTERED → replace)
	const numGoroutines = 10
	var wg sync.WaitGroup
	successes := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Step 1: Check if exists
			existing, err := db.GetExecutorByName(colony.Name, name)
			if err != nil {
				successes <- false
				return
			}

			// Step 2: If exists, remove it
			if existing != nil {
				db.RemoveExecutorByName(colony.Name, name)
			}

			// Step 3: Add new executor with same name
			newExec := core.CreateExecutor(core.GenerateRandomID(), "test-type", name, colony.Name, time.Now(), time.Now())
			err = db.AddExecutor(newExec)
			successes <- (err == nil)
		}()
	}
	wg.Wait()
	close(successes)

	successCount := 0
	for s := range successes {
		if s {
			successCount++
		}
	}
	t.Logf("Concurrent re-register: %d/%d succeeded", successCount, numGoroutines)

	// CRITICAL: There should be exactly 1 executor with this name (excluding UNREGISTERED)
	executors, err := db.GetExecutorsByColonyName(colony.Name, false)
	assert.NoError(t, err)

	count := 0
	for _, e := range executors {
		if e.Name == name {
			count++
		}
	}
	assert.Equal(t, 1, count, "Expected exactly 1 executor named %s, got %d (race condition!)", name, count)
}

// Verify that GetExecutorByName DOES return UNREGISTERED executors.
// This matches PostgreSQL behavior — the handler uses this to detect existing executors.
func TestGetExecutorByNameReturnsUnregistered(t *testing.T) {
	db := setupTestDB(t)

	colony := core.CreateColony(core.GenerateRandomID(), "test-colony")
	err := db.AddColony(colony)
	assert.NoError(t, err)

	executor := core.CreateExecutor(core.GenerateRandomID(), "test-type", "my-executor", colony.Name, time.Now(), time.Now())
	err = db.AddExecutor(executor)
	assert.NoError(t, err)

	// Remove it (marks as UNREGISTERED)
	err = db.RemoveExecutorByName(colony.Name, "my-executor")
	assert.NoError(t, err)

	// GetExecutorByName SHOULD return the UNREGISTERED executor (like PostgreSQL)
	found, err := db.GetExecutorByName(colony.Name, "my-executor")
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, core.UNREGISTERED, found.State)

	// But GetExecutorsByColonyName with includeUnregistered=false should NOT return it
	executors, err := db.GetExecutorsByColonyName(colony.Name, false)
	assert.NoError(t, err)
	assert.Len(t, executors, 0)
}

// Verify that re-registering after remove works and updates the executor ID.
func TestReregisterUpdatesExecutorID(t *testing.T) {
	db := setupTestDB(t)

	colony := core.CreateColony(core.GenerateRandomID(), "test-colony")
	err := db.AddColony(colony)
	assert.NoError(t, err)

	id1 := core.GenerateRandomID()
	executor1 := core.CreateExecutor(id1, "test-type", "my-executor", colony.Name, time.Now(), time.Now())
	err = db.AddExecutor(executor1)
	assert.NoError(t, err)

	// Remove it
	err = db.RemoveExecutorByName(colony.Name, "my-executor")
	assert.NoError(t, err)

	// Re-register with new ID
	id2 := core.GenerateRandomID()
	executor2 := core.CreateExecutor(id2, "test-type", "my-executor", colony.Name, time.Now(), time.Now())
	err = db.AddExecutor(executor2)
	assert.NoError(t, err)

	// Should have new ID
	found, err := db.GetExecutorByName(colony.Name, "my-executor")
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, id2, found.ID)
	assert.Equal(t, core.PENDING, found.State)

	// Old ID should be gone
	old, err := db.GetExecutorByID(id1)
	assert.NoError(t, err)
	assert.Nil(t, old)
}
