# Embedded Database - Known Issues

## Critical

### 1. Deadlock in ApplyRetentionPolicy [FIXED]
**File:** `pkg/database/embedded/database.go:743`

`db.logs.ForEach()` holds an RLock, but calls `db.logs.Delete()` inside the callback which tries to acquire a write Lock on the same `sync.RWMutex`. This will deadlock.

**Fix:** Collect IDs to delete during ForEach, then delete in a separate loop after ForEach returns.

**Status:** Fixed. Logs, attributes, and processes are now collected before deletion. Test: `TestSelectAndAssignFindCandidatesDeadlock`.

### 2. Deadlock between SelectAndAssign and FindCandidates [FIXED]
**File:** `pkg/database/embedded/processes.go`

Lock ordering conflict (ABBA pattern):
- `SelectAndAssign`: holds `processes` store Lock, then acquires CompoundIndex Lock (via `Remove`/`Add`)
- `FindCandidates`: holds CompoundIndex RLock (inside `AscendFirst` callback), then acquires `processes` store RLock (via `Get`)

Any concurrent call to both methods can deadlock.

**Fix:** FindCandidates and FindCandidatesByName now collect candidate IDs first from the CompoundIndex, then fetch processes outside the callback. Test: `TestSelectAndAssignFindCandidatesDeadlock`.

### 3. Close() truncates WAL without final flush [NOT A BUG]
**File:** `pkg/database/embedded/database.go:704-712`

**Status:** False positive. `flusher.Stop()` already performs a final flush before returning. The WAL truncation after stop is correct.

## High

### 4. Shallow copy in copyBlueprintDefinition [FIXED]
**File:** `pkg/database/embedded/blueprints.go`

Does not deep-copy:
- `Spec.Names.ShortNames []string`
- `Spec.Schema *ValidationSchema` (pointer field with nested maps/slices)

**Status:** Fixed. `copyBlueprintDefinition` now deep-copies ShortNames slice, Schema pointer (Properties map, Required slice, SchemaProperty fields). Test: `TestCopyBlueprintDefinitionDeepCopy`.

### 5. Shallow copy in copyBlueprintHistory [FIXED]
**File:** `pkg/database/embedded/blueprints.go`

`Spec` and `Status` (`map[string]interface{}`) are not deep-copied. Both the stored and returned copies share nested map/slice data.

**Status:** Fixed. `copyBlueprintHistory` now deep-copies Spec and Status maps via `copyInterfaceMap`. Test: `TestCopyBlueprintHistoryDeepCopy`.

### 6. copyInterfaceMap is only a shallow copy [FIXED]
**File:** `pkg/database/embedded/blueprints.go:179-188`

For `map[string]interface{}` values that contain nested maps or slices (common in blueprint spec/status), mutations through one copy affect the other.

**Status:** Fixed. Added `deepCopyInterface` helper that recursively copies nested `map[string]interface{}` and `[]interface{}` values. Test: `TestCopyBlueprintDeepCopyNestedSpec`.

## Medium

### 7. Unassign uses caller's Retries instead of stored value [FIXED]
**File:** `pkg/database/embedded/processes.go:401`

`cp.Retries = process.Retries + 1` uses the caller's `process.Retries`. If the caller holds a stale process object, the retry count could be wrong. Should use `p.Retries + 1` (from the stored record).

**Status:** Fixed. Changed to `cp.Retries = p.Retries + 1`. Test: `TestUnassignUsesStoredRetries`.

### 8. SetAllocations does not deep-copy Projects map [FIXED]
**File:** `pkg/database/embedded/executors.go:91-103`

`cp.Allocations = allocations` stores the caller's `Allocations` directly. The `Projects map[string]Project` is shared with the caller.

**Status:** Fixed. Uses `copyExecutor` + explicit Projects map deep copy. Test: `TestSetAllocationsDeepCopy`.

### 9. UpdateExecutorCapabilities does not deep-copy slices [FIXED]
**File:** `pkg/database/embedded/executors.go:290-302`

`cp.Capabilities = capabilities` shares `Hardware` and `Software` slices with the caller.

**Status:** Fixed. Uses `copyExecutor` + explicit Hardware/Software slice deep copy. Test: `TestUpdateExecutorCapabilitiesDeepCopy`.

### 10. copyFunction copies FunctionArg pointers, not structs [FIXED]
**File:** `pkg/database/embedded/functions.go:9-16`

Each `FunctionArg` contains `Enum []string`. The pointer slice is copied but the pointed-to structs (and their Enum slices) are shared between original and copy.

**Status:** Fixed. `copyFunction` now deep-copies each `*FunctionArg` struct and its Enum slice. Test: `TestCopyFunctionDeepCopyArgs`.

### 11. count*10 heuristic in FindCandidates and findProcessesByState
**File:** `pkg/database/embedded/processes.go:120,218,243`

With selective filters, scanning only `count*10` entries from the index can return fewer results than actually exist.

**Status:** Open. Acceptable trade-off for performance; unlikely to affect real workloads.

### 12. DiskStore creation errors silently ignored [FIXED]
**File:** `pkg/database/embedded/database.go:233-351`

Every `diskstore.NewDiskStore` call discards the error with `_, _ :=`. If a disk store fails to create (e.g., permissions), initialization proceeds without storage backing, and data loss will occur silently on restart.

**Status:** Fixed. All 18 `diskstore.NewDiskStore` calls now check and propagate errors.

## Low

### 13. Grammar: "does not exists" [FIXED]
**Files:** `pkg/database/embedded/colonies.go`, `pkg/database/embedded/attributes.go`, `pkg/database/embedded/snapshots.go`

Error messages say "does not exists" instead of "does not exist".

**Status:** Fixed across all affected files.

### 14. Dead code: unused newDisk helper [FIXED]
**File:** `pkg/database/embedded/database.go:221-225`

The `newDisk` closure is created and immediately discarded via `_ = newDisk`.

**Status:** Fixed. Removed the dead code.

### 15. Input mutations before copying
**Files:** `pkg/database/embedded/processes.go` (AddProcess), `pkg/database/embedded/files.go` (AddFile), `pkg/database/embedded/executors.go` (AddExecutor)

These methods mutate the caller's object (e.g., setting SubmissionTime, SequenceNumber, CommissionTime) before making the copy for storage. This matches PostgreSQL behavior but is a side effect on the caller's object.

**Status:** Open. Matches PostgreSQL behavior; not a bug.

### 16. GetExecutorsByColonyName returns unregistered executors [FIXED]
**File:** `pkg/database/embedded/executors.go`

`RemoveExecutorByName` marks executors as UNREGISTERED rather than deleting them, but `GetExecutorsByColonyName` does not filter by state. This means it can return unregistered executors, unlike `GetExecutors` which does filter.

**Status:** Fixed. Added `includeUnregistered bool` parameter to the `GetExecutorsByColonyName` interface method. Updated embedded DB, PostgreSQL, all mocks, handler, RPC message, client (`GetExecutorsWithOpts`), and CLI (`--all` flag). Test: `TestGetExecutorsByColonyNameExcludesUnregistered`.

## Test Coverage Gaps

### 17. Snapshot tests entirely missing [FIXED]
7 `SnapshotDatabase` interface methods have zero test coverage in `embedded_test.go`.

**Status:** Fixed. Added 10 tests: `TestCreateSnapshot`, `TestCreateSnapshotDuplicate`, `TestGetSnapshotByID`, `TestGetSnapshotByIDWrongColony`, `TestGetSnapshotsByColonyName`, `TestGetSnapshotByName`, `TestGetSnapshotByNameNotFound`, `TestRemoveSnapshotByID`, `TestRemoveSnapshotByName`, `TestRemoveSnapshotsByColonyName`.

### 18. BlueprintDefinition tests minimal [FIXED]
Only `AddBlueprintDefinition` and `GetBlueprintDefinitionByKind` are tested. Missing tests for: `GetBlueprintDefinitionByName`, `GetBlueprintDefinitions`, `GetBlueprintDefinitionsByNamespace`, `GetBlueprintDefinitionsByGroup`, `UpdateBlueprintDefinition`, `RemoveBlueprintDefinitionByID`, `RemoveBlueprintDefinitionByName`, `CountBlueprintDefinitions`.

**Status:** Fixed. Added 9 tests covering all missing methods.

### 19. Blueprint tests have gaps [FIXED]
Missing tests for: `GetBlueprints`, `GetBlueprintsByNamespace`, `GetBlueprintsByNamespaceAndKind`, `GetBlueprintsByNamespaceKindAndLocation`, `RemoveBlueprintByName`, `RemoveBlueprintsByNamespace`.

**Status:** Fixed. Added 6 tests covering all missing methods.

### 20. ApplyRetentionPolicy untested [FIXED]
Contains non-trivial deletion logic and the deadlock bug (issue 1). No test coverage.

**Status:** Fixed. Added `TestApplyRetentionPolicy` covering deletion of old SUCCESS processes, attributes, and logs based on retention period.

### 21. No persistence round-trip test [FIXED]
No test that writes data, closes the database, reopens it, and verifies all data survived through the full `EmbeddedDatabase` API.

**Status:** Fixed. Added `TestPersistenceRoundTrip` (full round-trip with colony, executor, process, function, log, generator, cron, file) and `TestPersistenceRoundTripWithWALReplay`.

### 22. No concurrency tests at EmbeddedDatabase level [FIXED]
All `embedded_test.go` tests are single-threaded. No tests exercise concurrent reads/writes through the database API.

**Status:** Fixed. Added 7 concurrency tests at the EmbeddedDatabase level:
- `TestConcurrentSelectAndAssignCompetition` - 10 executors competing for 100 processes
- `TestConcurrentAddProcessAndSelectAndAssign` - producers/consumers pattern
- `TestConcurrentStateTransitionsAndQueries` - concurrent MarkSuccessful/MarkFailed with count queries
- `TestConcurrentAttributeReadWrite` - concurrent attribute writers and readers
- `TestConcurrentRetentionPolicyAndLogWrites` - retention policy + log read/write
- `TestConcurrentAssignUnassignCycle` - assign/unassign cycles with FindCandidates queries
- `TestConcurrentMultiEntityOperations` - cross-entity concurrent operations

Additionally, added 16 stress tests for the lower-level storage components:
- `store/stress_test.go` - 5 tests (concurrent Put/Get/Delete, ForEach during mutations, FlushDirty during mutations, Lock/Unlock interleaving, concurrent All/Filter)
- `wal/stress_test.go` - 4 tests (concurrent appends, mixed ops, SyncAlways mode, append during replay)
- `flusher/stress_test.go` - 3 tests (many stores, dynamic AddStore, rapid Start/Stop)
- `diskstore/stress_test.go` - 4 tests (concurrent writes, read/write same key, write/delete, concurrent Scan)
