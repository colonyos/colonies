# Blueprint Implementation - Code Review Findings

This document contains findings from a code review of the blueprint implementation in the colonies repo and the docker-reconciler executor.

## Colonies Repo Issues

### CRITICAL

#### 1. ~~Race Condition in UpdateBlueprintStatus~~ [FIXED]

**Location:** `pkg/database/postgresql/blueprints.go:387-439`

**Issue:** Classic Read-Modify-Write race condition. Two concurrent requests updating blueprint status will lose the first update.

**Status:** FIXED - Replaced read-modify-write pattern with atomic PostgreSQL `jsonb_set` operation:

```go
// Atomic update using jsonb_set - no race condition
updateStatement := `UPDATE BLUEPRINTS SET DATA = jsonb_set(DATA::jsonb, '{status}', $1::jsonb) WHERE ID=$2`
```

**Tests added:**
- `TestUpdateBlueprintStatusConcurrent` - verifies 10 concurrent updates all succeed
- `TestUpdateBlueprintStatusAtomicUpdate` - verifies spec preserved during status update
- `TestUpdateBlueprintStatusSequentialUpdates` - verifies sequential updates work

---

#### 2. ~~Cron Naming Mismatch Between AddBlueprint and UpdateBlueprint~~ [FIXED]

**Location:**
- AddBlueprint: `pkg/server/handlers/blueprint/handlers.go:610`
- UpdateBlueprint: `pkg/server/handlers/blueprint/handlers.go:950`

**Issue:** Different naming schemes used:
- AddBlueprint: `reconcile-{Kind}-{locationName}`
- UpdateBlueprint: `reconcile-{Kind}-{executorType}`

**Status:** FIXED - Updated UpdateBlueprint to use the same naming pattern as AddBlueprint:
`reconcile-{Kind}-{locationName}`

**Test added:** `TestCronNamingConsistencyBetweenAddAndUpdate` - verifies both handlers use consistent naming and UpdateBlueprint triggers the cron correctly.

---

#### 3. ~~RemoveBlueprint Uses Wrong Cron Name~~ [FIXED]

**Location:** `pkg/server/handlers/blueprint/handlers.go:1034`

**Issue:** RemoveBlueprint looks for `reconcile-{Kind}` but cron was created as `reconcile-{Kind}-{location}`.

**Status:** FIXED - Updated RemoveBlueprint to use the same naming pattern as AddBlueprint:
`reconcile-{Kind}-{locationName}`. Also improved logic to only remove cron when no blueprints
of the same Kind remain at the same location.

**Tests added:**
- `TestRemoveBlueprintCronCleanup` - verifies cron is removed when last blueprint at location is deleted
- `TestRemoveBlueprintCronKeptWhenOthersExist` - verifies cron is kept when other blueprints at location exist

---

#### 4. ~~Transaction Safety Violation in HandleAddBlueprint~~ [MITIGATED]

**Location:** `pkg/server/handlers/blueprint/handlers.go:577-666`

**Issue:** Blueprint is added to database BEFORE cron is created. No database transaction wrapping both operations.

**Impact:** Crash between operations leaves blueprint without cron, preventing reconciliation.

**Status:** MITIGATED - Instead of adding complexity to the hot path, added `colonies blueprint doctor --fix` command that can detect and fix orphaned blueprints (blueprints without crons). This is a rare edge case and the doctor command provides an operational tool for when things go wrong.

**Doctor command can now:**
- Detect missing reconciliation crons for blueprints
- Create missing crons with `--fix` flag
- Trigger force reconciliation for replica mismatches with `--fix` flag

---

### HIGH PRIORITY

#### 5. ~~Unsafe Concurrent Access to CronController~~ [FIXED]

**Location:** Multiple locations in handlers.go

**Issue:** Get-then-create pattern without locking:
```go
crons, err := h.server.CronController().GetCrons(...)
// Race window here
addedCron, err := h.server.CronController().AddCron(cron)
```

**Impact:** Duplicate crons possible under concurrent blueprint creation.

**Status:** FIXED - Added composite UNIQUE constraint on (COLONY_NAME, NAME) in the CRONS table. The AddCron function now relies on the database constraint for atomicity instead of the check-then-act pattern:

```go
// Atomic insert - database constraint prevents duplicates
_, err := db.postgresql.Exec(sqlStatement, ...)
if err != nil {
    if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
        return errors.New("Cron with name <" + cron.Name + "> already exists")
    }
    return err
}
```

**Tests added:**
- `TestAddDuplicateCronRejected` - verifies duplicate crons (same colony, same name) are rejected
- `TestSameCronNameDifferentColoniesAllowed` - verifies same name in different colonies is allowed
- `TestAddCronConcurrentDuplicateRejected` - verifies concurrent insertions only allow one success

---

#### 6. ~~Inefficient GetBlueprintDefinitions Query~~ [FIXED]

**Location:** `pkg/server/handlers/blueprint/handlers.go:1109`

**Issue:** Fetches ALL blueprint definitions just to find one by Kind.

**Impact:** O(n) performance instead of O(1).

**Status:** FIXED - Added `GetBlueprintDefinitionByKind(kind string)` method to the database layer that uses direct SQL filtering on the KIND column. The handler now uses a single efficient query:

```go
// Before: O(n) - fetched all definitions and iterated
blueprintDefs, err := h.server.BlueprintDB().GetBlueprintDefinitions()
for _, def := range blueprintDefs {
    if def.Spec.Names.Kind == blueprint.Kind { ... }
}

// After: O(1) - direct database lookup
blueprintDef, err := h.server.BlueprintDB().GetBlueprintDefinitionByKind(blueprint.Kind)
```

**Tests added:**
- `TestGetBlueprintDefinitionByKind` - verifies lookup by kind, multiple kinds, and non-existent kind handling

---

#### 7. ~~Silent Failure in Immediate Reconciliation Process~~ [FIXED]

**Location:** `pkg/server/handlers/blueprint/handlers.go:687-711`

**Issue:** If immediate reconciliation process creation fails, error is logged but not reported to client.

**Impact:** Blueprints created without reconciliation running.

**Status:** FIXED - Changed from silent logging to returning errors to the client. If immediate reconciliation fails, the client receives an informative error:

```go
// Before: Silent failure - only logged, client got success
log.Warn("Failed to create immediate reconciliation process")

// After: Error returned to client
h.server.HandleHTTPError(c, fmt.Errorf("blueprint created but failed to create reconciliation process: %w", err), http.StatusInternalServerError)
return
```

**Tests added:**
- `TestImmediateReconciliationTriggeredOnBlueprintAdd` - verifies that immediate reconciliation processes are created for each blueprint added, testing the success path

---

#### 8. ~~Inconsistent Location Auto-creation~~ [FIXED]

**Location:** `pkg/server/handlers/blueprint/handlers.go:545-599`

**Issue:** Location is created as side effect before blueprint. If blueprint creation fails, location persists.

**Impact:** Orphaned locations accumulate.

**Status:** FIXED - Added cleanup logic that removes auto-created locations when blueprint creation fails. Pre-existing locations are never affected:

```go
var locationWasAutoCreated bool
// ... location creation sets flag to true ...

err = h.server.BlueprintDB().AddBlueprint(msg.Blueprint)
if err != nil {
    // Clean up auto-created location if blueprint creation fails
    if locationWasAutoCreated {
        h.server.LocationDB().RemoveLocationByName(colonyName, locationName)
    }
    // return error
}
```

**Tests added:**
- `TestLocationAutoCreatedWithBlueprint` - verifies location is auto-created and persists on success
- `TestPreExistingLocationNotAffectedByBlueprintFailure` - verifies pre-existing locations are not deleted on failure
- `TestLocationCleanupOnBlueprintCreationFailure` - verifies cleanup behavior and ordering

---

### MEDIUM PRIORITY

#### 9. ~~Missing Validation in UpdateBlueprint~~ [FIXED]

**Location:** `pkg/server/handlers/blueprint/handlers.go:875-879`

**Issue:** Updating a non-existent blueprint proceeds without error.

**Impact:** Update endpoint can create blueprints, causing API inconsistency.

**Status:** FIXED - Added explicit check for non-existent blueprint that returns 404:

```go
if oldBlueprint == nil {
    h.server.HandleHTTPError(c, fmt.Errorf("blueprint '%s' not found in namespace '%s'",
        msg.Blueprint.Metadata.Name, msg.Blueprint.Metadata.ColonyName), http.StatusNotFound)
    return
}
```

**Tests added:**
- `TestUpdateBlueprintReturns404ForNonExistent` - verifies 404 status code is returned with proper error message
- `TestUpdateBlueprintSucceedsForExistingBlueprint` - verifies update works correctly for existing blueprints

---

#### 10. ~~Generation Bumping Logic Edge Case~~ [FIXED]

**Location:** `pkg/server/handlers/blueprint/handlers.go:891-915`

**Issue:** If oldBlueprint is nil, generation is never set properly.

**Impact:** Incorrect generation tracking for blueprints created via update endpoint.

**Status:** FIXED - This is now a non-issue because the fix for issue #9 ensures `oldBlueprint` can never be nil at the generation logic point. The 404 check at line 875-879 returns early if the blueprint doesn't exist.

---

#### 11. ~~deepEqual Memory Safety~~ [FIXED]

**Location:** `pkg/core/blueprint.go:601-605`

**Issue:** json.Marshal errors are ignored using `_`.

**Impact:** Incorrect diff computation if marshaling fails.

**Status:** FIXED - Added error handling that falls back to `reflect.DeepEqual` when JSON marshaling fails:

```go
func deepEqual(a, b interface{}) bool {
    aJSON, errA := json.Marshal(a)
    bJSON, errB := json.Marshal(b)
    if errA != nil || errB != nil {
        // Fall back to reflect.DeepEqual if JSON serialization fails
        return reflect.DeepEqual(a, b)
    }
    return string(aJSON) == string(bJSON)
}
```

**Tests added:**
- `TestDeepEqualWithJSONSerializableValues` - verifies normal JSON-serializable values work
- `TestDeepEqualWithUnmarshalableValues` - verifies fallback to reflect.DeepEqual for channels
- `TestDeepEqualDoesNotReturnTrueForDifferentUnmarshalableValues` - verifies bug fix for different unmarshable values
- `TestDeepEqualWithBlueprintSpec` - verifies real-world BlueprintSpec comparison works

---

#### 12. ~~Silent Audit Trail Failures~~ [FIXED]

**Location:** Multiple locations (lines 603-607, 952-957)

**Issue:** Blueprint history save failures are logged but ignored.

**Impact:** Audit trail can be lost without client awareness.

**Status:** FIXED - Changed to return error to client instead of silent logging:

```go
// AddBlueprint (line 603-607)
if err := h.server.BlueprintDB().AddBlueprintHistory(history); err != nil {
    log.WithFields(log.Fields{"Error": err, "BlueprintID": msg.Blueprint.ID}).Error("Failed to save blueprint history")
    h.server.HandleHTTPError(c, fmt.Errorf("blueprint created but failed to save audit history: %w", err), http.StatusInternalServerError)
    return
}

// UpdateBlueprint (line 952-957)
if err := h.server.BlueprintDB().AddBlueprintHistory(history); err != nil {
    log.WithFields(log.Fields{"Error": err, "BlueprintID": msg.Blueprint.ID}).Error("Failed to save blueprint history")
    h.server.HandleHTTPError(c, fmt.Errorf("blueprint updated but failed to save audit history: %w", err), http.StatusInternalServerError)
    return
}
```

This ensures audit trail integrity - if history cannot be saved, the client is notified and can retry the operation.

---

### LOW PRIORITY

#### 13. ~~Inefficient Cron Lookup Pattern~~ [FIXED]

**Location:** Multiple handlers (AddBlueprint, UpdateBlueprint, RemoveBlueprint)

**Issue:** Gets ALL crons (up to 1000) just to find one by name.

**Status:** FIXED - Exposed `GetCronByName(colonyName, cronName)` method through CronController for O(1) lookups:

```go
// Before: O(n) - fetched all crons and iterated
crons, err := h.server.CronController().GetCrons(colonyName, 1000)
var existingCron *core.Cron
for _, cron := range crons {
    if cron.Name == cronName {
        existingCron = cron
        break
    }
}

// After: O(1) - direct database lookup
existingCron, err := h.server.CronController().GetCronByName(colonyName, cronName)
```

**Changes:**
- Added `GetCronByName` to CronController (`pkg/server/controllers/cron_controller.go`)
- Updated Controller interface (`pkg/server/controllers/controller.go`)
- Updated server interfaces and adapters (`pkg/server/interfaces.go`, `pkg/server/server_adapter.go`)
- Updated blueprint and cron handler interfaces
- Updated 3 locations in `pkg/server/handlers/blueprint/handlers.go`

**Tests added:**
- `TestGetCronByName` - verifies lookup by name, different colonies, and non-existent cases

---

#### 14. ~~SQL Injection in LIMIT Clause~~ [FIXED]

**Location:** `pkg/database/postgresql/blueprints.go:540-559`

**Issue:** Using fmt.Sprintf for LIMIT instead of parameterization.

**Status:** FIXED - Changed to use proper SQL parameterization:

```go
// Before: String interpolation (bad practice)
if limit > 0 {
    sqlStatement += fmt.Sprintf(" LIMIT %d", limit)
}
rows, err := db.postgresql.Query(sqlStatement, blueprintID)

// After: Parameterized query (safe)
if limit > 0 {
    sqlStatement := `SELECT ... FROM BLUEPRINT_HISTORY
        WHERE BLUEPRINT_ID=$1
        ORDER BY TIMESTAMP DESC
        LIMIT $2`
    rows, err = db.postgresql.Query(sqlStatement, blueprintID, limit)
} else {
    sqlStatement := `SELECT ... FROM BLUEPRINT_HISTORY
        WHERE BLUEPRINT_ID=$1
        ORDER BY TIMESTAMP DESC`
    rows, err = db.postgresql.Query(sqlStatement, blueprintID)
}
```

**Tests added:**
- `TestGetBlueprintHistoryParameterizedLimit` - verifies limit=0, limit=1, limit=3, limit>total, negative limit, and non-existent blueprint cases

---

## Docker-Reconciler Issues

### CRITICAL

#### 1. ~~Container Lifecycle Ordering~~ [FIXED]

**Location:** `pkg/reconciler/executor_deployment.go:238-269`

**Issue:** Executor is deregistered BEFORE container is stopped. If stop fails, container is orphaned.

```go
// Current unsafe sequence
r.client.RemoveExecutor(...)  // Executor gone
r.stopAndRemoveContainer(...) // If this fails, container orphaned
```

**Impact:** Container leaks, resource exhaustion.

**Status:** FIXED - Reversed the lifecycle order to stop container FIRST, then deregister executor:

```go
// Stop container FIRST, then deregister executor
// This prevents orphaned containers if stop fails - executor stays registered for retry
if err := r.stopAndRemoveContainer(containerID); err != nil {
    log.WithFields(log.Fields{
        "Error":       err,
        "ContainerID": truncateID(containerID, 12),
    }).Warn("Failed to stop container, keeping executor registered for retry")
    r.addLog(process, fmt.Sprintf("Warning: Failed to stop container %s: %v (will retry)", truncateID(containerID, 12), err))
    continue // Skip deregistration, retry on next reconciliation
}
// Now deregister executor (only after container is successfully stopped)
```

**Tests added:**
- Updated `TestScaleDownDeregistration` tests to verify new order (container stop before executor deregistration)
- Added "Container stop failure prevents executor deregistration" test case

---

#### 2. ~~No Timeout on Image Pull~~ [FIXED]

**Location:** `pkg/reconciler/container.go:17-65`

**Issue:** Image pulls have no timeout and no cancellation support.

**Impact:** Goroutine leaks, stuck reconciliation processes.

**Status:** FIXED - Added 10-minute timeout for image pull operations:

```go
// pkg/constants/constants.go
const ImagePullTimeout = 10 * time.Minute

// container.go - doPullImageWithTimeout
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()
// ...
select {
case <-ctx.Done():
    return fmt.Errorf("image pull timed out after %s: %s", timeout, image)
// ...
}
```

**Changes:**
- Created `pkg/constants/constants.go` with `ImagePullTimeout = 10 * time.Minute`
- Added `ImagePuller` interface to `types.go` for testability
- Refactored `doPullImage` to use `doPullImageWithTimeout` for testability

**Tests added:**
- `TestImagePullTimeout/Image_pull_times_out_after_specified_duration` - verifies timeout works
- `TestImagePullTimeout/Image_pull_succeeds_before_timeout` - verifies success path
- `TestImagePullTimeout/Image_pull_error_is_returned` - verifies errors are propagated

---

### HIGH PRIORITY

#### 3. ~~Executor Name Generation Race~~ [FIXED]

**Location:** `pkg/reconciler/executor_deployment.go:186-200`

**Issue:** Name availability check and registration are not atomic. Under high concurrency, duplicate names possible.

**Impact:** Duplicate executor registrations.

**Status:** FIXED - Replaced check-then-act pattern with atomic registration:

```go
// OLD: Race condition between check and registration
func (r *Reconciler) generateUniqueExecutorName(...) (string, error) {
    for i := 0; i < maxRetries; i++ {
        name := generateName()
        taken, _ := r.isExecutorNameTaken(name)  // CHECK
        if !taken {
            return name, nil  // RACE WINDOW before registration
        }
    }
}

// NEW: Atomic registration with retry on duplicate
for attempt := 0; attempt < maxRetries; attempt++ {
    containerName = generateExecutorName(blueprint.Metadata.Name)
    lastErr = r.startContainer(...)  // Tries to register
    if lastErr == nil {
        break  // Success
    }
    if isDuplicateExecutorError(lastErr) {
        continue  // Retry with new name
    }
    return lastErr  // Non-duplicate error
}
```

**Changes:**
- Removed `isExecutorNameTaken` check-then-act pattern from `executor_manager.go`
- Added `isDuplicateExecutorError` helper to detect duplicate errors
- Updated scale-up loop to retry with new name on duplicate error
- Server-side PRIMARY KEY constraint ensures atomicity

**Tests added (Colonies):**
- `TestAddDuplicateExecutorRejected` - verifies duplicate rejection
- `TestAddDuplicateExecutorConcurrentRejected` - verifies concurrent duplicates rejected
- `TestSameExecutorNameDifferentColoniesAllowed` - verifies cross-colony names allowed

**Tests added (Docker-Reconciler):**
- `TestAtomicExecutorRegistration/isDuplicateExecutorError_detects_duplicate_errors`
- `TestAtomicExecutorRegistration/generateExecutorName_generates_unique_names`
- `TestAtomicExecutorRegistration/generateUniqueExecutorName_returns_name_without_pre-check`

---

#### 4. ~~No Cleanup on Failed Container Start~~ [FIXED]

**Location:** `pkg/reconciler/executor_deployment.go:551-580`

**Issue:** If ContainerStart fails, created container is left in "created" state.

**Impact:** Incorrect replica counts, cascading recreations.

**Status:** FIXED - Added cleanup logic for both ContainerStart and waitForContainerRunning failures:

```go
// If ContainerStart fails, remove the created container
if err := r.dockerClient.ContainerStart(...); err != nil {
    // Cleanup: remove the created container to prevent orphaned containers
    if removeErr := r.dockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true}); removeErr != nil {
        log.Warn("Failed to cleanup container after start failure")
    }
    return fmt.Errorf("failed to start container: %w", err)
}

// If waitForContainerRunning fails, stop and remove the container
if err := r.waitForContainerRunning(...); err != nil {
    // Cleanup: stop and remove the container to prevent orphaned containers
    if removeErr := r.stopAndRemoveContainer(resp.ID); removeErr != nil {
        log.Warn("Failed to cleanup container after wait failure")
    }
    return fmt.Errorf("container failed to start: %w", err)
}
```

**Tests added:**
- `TestContainerCleanupOnStartFailure/ContainerStart_failure_triggers_cleanup`
- `TestContainerCleanupOnStartFailure/waitForContainerRunning_failure_triggers_cleanup`
- `TestContainerCleanupOnStartFailure/Cleanup_failure_is_logged_but_original_error_returned`

---

#### 5. ~~Executor Name Parsing Fragility~~ [FIXED]

**Location:** `pkg/reconciler/cleanup.go:356-487`

**Issue:** Parsing executor names assumes deployment names don't contain hyphens.

**Impact:** Orphaned executors if deployment names contain hyphens.

**Status:** FIXED - Added structured labels for direct executor-to-container matching:

**Changes to `executor_deployment.go`:**
```go
// Set BlueprintID and BlueprintGen on executor for tracking
newExecutor.BlueprintID = blueprint.ID
newExecutor.BlueprintGen = blueprint.Metadata.Generation

// Add colonies.executor label to container for direct matching
labels["colonies.executor"] = executorName
```

**Changes to `cleanup.go`:**
```go
// Build executor set from container labels for direct matching
containerExecutors := make(map[string]bool)
for _, cont := range containers {
    if execName, ok := cont.Labels["colonies.executor"]; ok {
        containerExecutors[execName] = true
    }
}

// First try direct label match (preferred)
if containerExecutors[executor.Name] {
    continue // Container exists
}
// Fallback to legacy name parsing for backward compatibility
```

**Tests added:**
- `TestExecutorLabelMatching/Executor_with_colonies.executor_label_is_matched_correctly`
- `TestExecutorLabelMatching/Hyphenated_deployment_names_work_correctly` (tests simple, my-app, my-app-test, my-app-test-prod)
- `TestExecutorLabelMatching/Stale_executor_without_container_is_detected`

---

#### 6. ~~Concurrent Deregistration Handling~~ [NOT AN ISSUE]

**Location:** `pkg/reconciler/executor_deployment.go:238-269`

**Issue:** Multiple reconcilers might attempt same deregistration. "Not found" errors treated same as real errors.

**Status:** NOT AN ISSUE - Reconciliation is cron-driven and each executor can only get one process at a time. Multiple reconcilers cannot attempt the same deregistration concurrently because:
1. Process assignment is atomic (only one reconciler gets a given reconcile process)
2. Each reconciler works on its own assigned process independently
3. "Not found" errors during deregistration are already logged as warnings and don't cause failures

The current behavior is correct - it logs a warning and continues, which is acceptable for edge cases like executors that timed out.

---

### MEDIUM PRIORITY

#### 7. ~~Force Reconcile Timing~~ [FIXED]

**Location:** `pkg/reconciler/reconciler.go:209-250`

**Issue:** Force reconcile pulls images before checking if containers can start. If pull hangs, entire operation blocked.

**Impact:** Extended service downtime.

**Status:** FIXED - Force reconcile now uses fail-fast approach:

1. Uses shorter timeout (5 minutes vs 10 minutes for normal pulls)
2. Aborts immediately if ANY image pull fails
3. Only proceeds to remove containers AFTER all images are successfully pulled
4. Preserves running containers if image pull fails (maintains service availability)

**Changes to `pkg/constants/constants.go`:**
```go
// Force reconcile image pull timeout - 5 minutes
// Shorter than regular pull timeout to fail fast during force reconcile.
const ForceReconcileImagePullTimeout = 5 * time.Minute
```

**Changes to `pkg/reconciler/reconciler.go`:**
```go
// Pull all images BEFORE removing any containers
for _, image := range images {
    if err := r.forcePullImageWithTimeout(process, image, constants.ForceReconcileImagePullTimeout); err != nil {
        r.addLog(process, fmt.Sprintf("ERROR: Failed to pull image %s - aborting to preserve service availability", image))
        return fmt.Errorf("failed to pull image %s (aborting to preserve running containers): %w", image, err)
    }
}
// Only remove containers after all images are pulled successfully
```

**Tests added:**
- `TestForceReconcileTiming/Force_reconcile_aborts_if_image_pull_fails` - verifies abort on pull failure
- `TestForceReconcileTiming/Force_reconcile_proceeds_to_container_removal_when_image_pull_succeeds` - verifies normal flow
- `TestForceReconcileTiming/Force_reconcile_preserves_containers_on_pull_timeout` - verifies containers preserved on timeout

---

#### 8. ~~Missing Image Validation~~ [FIXED]

**Location:** `pkg/reconciler/executor_deployment.go:305-582`

**Issue:** startContainer doesn't verify image exists before creating container.

**Impact:** Service downtime if image unavailable.

**Status:** FIXED - Added image validation before container creation in `startContainer`:

```go
// Validate image exists locally before creating container
if _, _, err := r.dockerClient.ImageInspectWithRaw(ctx, spec.Image); err != nil {
    r.addErrorLog(process, fmt.Sprintf("Docker: Image %s not found locally - ensure image is pulled before creating container", spec.Image))
    return fmt.Errorf("image not found: %s (pull the image first)", spec.Image)
}
```

This provides a clear error message if the image is not available locally, preventing confusing Docker errors and making it obvious what needs to be fixed.

**Tests added:**
- `TestImageValidation/startContainer_fails_if_image_not_found_locally` - verifies clear error when image missing
- `TestImageValidation/startContainer_proceeds_when_image_exists_locally` - verifies normal flow with image present

---

#### 9. ~~Dirty Container Recreation Race~~ [NOT AN ISSUE]

**Location:** `pkg/reconciler/executor_deployment.go:64-122`

**Issue:** No locking between dirty container detection and recreation.

**Status:** NOT AN ISSUE - ColonyOS guarantees that only one reconciler gets assigned a reconciliation process at a time. Reconcilers are stateless and process dirty containers sequentially within a single reconciliation run. The atomic process assignment model provides the coordination needed - two reconcilers cannot process the same blueprint simultaneously.

---

### LOW PRIORITY

#### 10. ~~Logging Thread Safety~~ [NOT AN ISSUE]

**Location:** `pkg/reconciler/utils.go:20-51`

**Issue:** Lock only protects formatting, not client call. Slow log operations can block reconciliation.

**Status:** NOT AN ISSUE - Reconcilers are single-threaded by design. Each reconciler handles one reconciliation process at a time, and all operations within a reconciliation run sequentially. There are no concurrent goroutines competing to log, so the lock provides ordering guarantees without actual contention.

---

#### 11. ~~Full Stopped Container Scan~~ [NOT AN ISSUE]

**Location:** `pkg/reconciler/cleanup.go:309-354`

**Issue:** Lists ALL stopped containers on Docker daemon without pagination.

**Status:** NOT AN ISSUE - The query is already filtered by `colonies.managed=true` label, limiting scope to reconciler-managed containers only. Stopped containers are cleaned up regularly preventing accumulation. Docker's label-based filtering is efficient (done server-side).

---

## Architectural Concerns

### ~~Lack of Distributed Locking~~ [NOT AN ISSUE]

Multiple reconcilers might try to reconcile the same blueprint simultaneously, causing resource conflicts.

**Status:** NOT AN ISSUE - The Colonies process assignment model already provides distributed coordination. Only one executor can be assigned a reconciliation process at a time. When a blueprint triggers reconciliation, a process is created and assigned to exactly one reconciler - other reconcilers cannot claim the same process.

### ~~No Transactional Semantics~~ [MITIGATED]

Blueprint + cron operations are separate without transaction boundaries. System crashes leave inconsistent state.

**Status:** MITIGATED - Rather than adding transaction complexity, the `colonies blueprint doctor --fix` command can detect and repair inconsistent state. This is a pragmatic solution for a rare edge case.

---

## Doctor Command

The `colonies blueprint doctor` command diagnoses blueprint configuration issues. Use `--fix` to automatically fix issues where possible.

### Usage

```bash
# Diagnose all blueprints
colonies blueprint doctor

# Diagnose specific blueprint
colonies blueprint doctor --name my-deployment

# Diagnose and fix issues
colonies blueprint doctor --fix
```

### Checks Performed

| Check | Can Auto-Fix |
|-------|--------------|
| Reconciler exists at location | No |
| Handler type matches reconciler | No |
| Reconciler is online (heartbeat) | No |
| Replica count matches desired | Yes - triggers force reconciliation |
| Reconciliation cron exists | Yes - creates missing cron |

### Example Output

```
=== my-deployment ===
  [OK] Reconciler found at location 'dc1'
  [OK] Handler type 'docker-reconciler' matches reconciler at location
  [OK] Reconciler 'local-docker-reconciler' is active (last heard 5s ago)
  [WARN] Only 1/3 replicas running
  [FIXED] Triggered force reconciliation
  [OK] Reconciliation cron 'reconcile-ExecutorDeployment-dc1' exists

Found 1 issue(s), fixed 1
```

---

## Priority Summary

| Priority | Colonies Issues | Docker-Reconciler Issues |
|----------|-----------------|--------------------------|
| CRITICAL | 0 issues (4 fixed/mitigated) | 0 issues (2 fixed) |
| HIGH | 0 issues (4 fixed) | 0 issues (3 fixed, 1 not an issue) |
| MEDIUM | 0 issues (4 fixed) | 0 issues (2 fixed, 1 not an issue) |
| LOW | 0 issues (2 fixed) | 0 issues (2 not an issue) |

**All issues resolved.**

## Fixed Issues

1. **Race Condition in UpdateBlueprintStatus** (CRITICAL) - Fixed using atomic `jsonb_set` operation
2. **Cron Naming Mismatch Between AddBlueprint and UpdateBlueprint** (CRITICAL) - Fixed UpdateBlueprint to use `reconcile-{Kind}-{locationName}` pattern
3. **RemoveBlueprint Uses Wrong Cron Name** (CRITICAL) - Fixed to use `reconcile-{Kind}-{locationName}` pattern with per-location cleanup logic
4. **Transaction Safety Violation in HandleAddBlueprint** (CRITICAL) - Mitigated with `colonies blueprint doctor --fix` command
5. **Unsafe Concurrent Access to CronController** (HIGH) - Fixed using composite UNIQUE constraint on (COLONY_NAME, NAME) in CRONS table
6. **Inefficient GetBlueprintDefinitions Query** (HIGH) - Fixed by adding `GetBlueprintDefinitionByKind` method for O(1) lookups
7. **Silent Failure in Immediate Reconciliation Process** (HIGH) - Fixed by returning errors to client instead of silent logging
8. **Inconsistent Location Auto-creation** (HIGH) - Fixed by cleaning up auto-created locations when blueprint creation fails
9. **Missing Validation in UpdateBlueprint** (MEDIUM) - Fixed by returning 404 for non-existent blueprints
10. **Generation Bumping Logic Edge Case** (MEDIUM) - Fixed as side-effect of issue #9 fix (oldBlueprint can never be nil)
11. **deepEqual Memory Safety** (MEDIUM) - Fixed by adding error handling with fallback to reflect.DeepEqual
12. **Silent Audit Trail Failures** (MEDIUM) - Fixed by returning error to client instead of silent logging
13. **Inefficient Cron Lookup Pattern** (LOW) - Fixed by exposing GetCronByName for O(1) lookups
14. **SQL Injection in LIMIT Clause** (LOW) - Fixed by using parameterized query for LIMIT

### Docker-Reconciler Fixed Issues

15. **Container Lifecycle Ordering** (CRITICAL) - Fixed by reversing order: stop container first, then deregister executor
16. **No Timeout on Image Pull** (CRITICAL) - Fixed by adding 10-minute timeout with context cancellation
17. **Executor Name Generation Race** (HIGH) - Fixed by using atomic registration with retry on duplicate
18. **No Cleanup on Failed Container Start** (HIGH) - Fixed by adding cleanup on ContainerStart and waitForContainerRunning failures
19. **Executor Name Parsing Fragility** (HIGH) - Fixed by using colonies.executor label for direct matching
20. **Concurrent Deregistration Handling** (HIGH) - Not an issue due to cron-driven process assignment
21. **Force Reconcile Timing** (MEDIUM) - Fixed with fail-fast approach and shorter timeout
22. **Missing Image Validation** (MEDIUM) - Fixed by validating image exists before container creation
23. **Dirty Container Recreation Race** (MEDIUM) - Not an issue due to atomic process assignment
24. **Logging Thread Safety** (LOW) - Not an issue due to single-threaded reconciler design
25. **Full Stopped Container Scan** (LOW) - Not an issue due to label filtering and regular cleanup
