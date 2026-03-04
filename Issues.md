# Embedded Database - Known Issues

## Critical

### 1. Deadlock in ApplyRetentionPolicy
**File:** `pkg/database/embedded/database.go:743`

`db.logs.ForEach()` holds an RLock, but calls `db.logs.Delete()` inside the callback which tries to acquire a write Lock on the same `sync.RWMutex`. This will deadlock.

**Fix:** Collect IDs to delete during ForEach, then delete in a separate loop after ForEach returns.

### 2. Deadlock between SelectAndAssign and FindCandidates
**File:** `pkg/database/embedded/processes.go`

Lock ordering conflict (ABBA pattern):
- `SelectAndAssign`: holds `processes` store Lock, then acquires CompoundIndex Lock (via `Remove`/`Add`)
- `FindCandidates`: holds CompoundIndex RLock (inside `AscendFirst` callback), then acquires `processes` store RLock (via `Get`)

Any concurrent call to both methods can deadlock.

**Fix:** Either use a consistent lock ordering across all methods, or avoid calling `db.processes.Get()` inside the `AscendFirst` callback (collect IDs first, then fetch outside the callback).

### 3. Close() truncates WAL without final flush
**File:** `pkg/database/embedded/database.go:704-712`

Dirty records in memory that have not been flushed to disk are lost when the WAL is truncated during `Close()`. The flusher is stopped but there is no explicit final flush of all stores before WAL truncation.

**Fix:** Add a final flush of all stores before truncating the WAL in `Close()`.

## High

### 4. Shallow copy in copyBlueprintDefinition
**File:** `pkg/database/embedded/blueprints.go`

Does not deep-copy:
- `Spec.Names.ShortNames []string`
- `Spec.Schema *ValidationSchema` (pointer field with nested maps/slices)

### 5. Shallow copy in copyBlueprintHistory
**File:** `pkg/database/embedded/blueprints.go`

`Spec` and `Status` (`map[string]interface{}`) are not deep-copied. Both the stored and returned copies share nested map/slice data.

### 6. copyInterfaceMap is only a shallow copy
**File:** `pkg/database/embedded/blueprints.go:179-188`

For `map[string]interface{}` values that contain nested maps or slices (common in blueprint spec/status), mutations through one copy affect the other.

## Medium

### 7. Unassign uses caller's Retries instead of stored value
**File:** `pkg/database/embedded/processes.go:401`

`cp.Retries = process.Retries + 1` uses the caller's `process.Retries`. If the caller holds a stale process object, the retry count could be wrong. Should use `p.Retries + 1` (from the stored record).

### 8. SetAllocations does not deep-copy Projects map
**File:** `pkg/database/embedded/executors.go:91-103`

`cp.Allocations = allocations` stores the caller's `Allocations` directly. The `Projects map[string]Project` is shared with the caller.

### 9. UpdateExecutorCapabilities does not deep-copy slices
**File:** `pkg/database/embedded/executors.go:290-302`

`cp.Capabilities = capabilities` shares `Hardware` and `Software` slices with the caller.

### 10. copyFunction copies FunctionArg pointers, not structs
**File:** `pkg/database/embedded/functions.go:9-16`

Each `FunctionArg` contains `Enum []string`. The pointer slice is copied but the pointed-to structs (and their Enum slices) are shared between original and copy.

### 11. count*10 heuristic in FindCandidates and findProcessesByState
**File:** `pkg/database/embedded/processes.go:120,218,243`

With selective filters, scanning only `count*10` entries from the index can return fewer results than actually exist.

### 12. DiskStore creation errors silently ignored
**File:** `pkg/database/embedded/database.go:233-351`

Every `diskstore.NewDiskStore` call discards the error with `_, _ :=`. If a disk store fails to create (e.g., permissions), initialization proceeds without storage backing, and data loss will occur silently on restart.

## Low

### 13. Grammar: "does not exists"
**Files:** `pkg/database/embedded/colonies.go`, `pkg/database/embedded/attributes.go`

Error messages say "does not exists" instead of "does not exist".

### 14. Dead code: unused newDisk helper
**File:** `pkg/database/embedded/database.go:221-225`

The `newDisk` closure is created and immediately discarded via `_ = newDisk`.

### 15. Input mutations before copying
**Files:** `pkg/database/embedded/processes.go` (AddProcess), `pkg/database/embedded/files.go` (AddFile), `pkg/database/embedded/executors.go` (AddExecutor)

These methods mutate the caller's object (e.g., setting SubmissionTime, SequenceNumber, CommissionTime) before making the copy for storage. This matches PostgreSQL behavior but is a side effect on the caller's object.

### 16. GetExecutorsByColonyName returns unregistered executors
**File:** `pkg/database/embedded/executors.go`

`RemoveExecutorByName` marks executors as UNREGISTERED rather than deleting them, but `GetExecutorsByColonyName` does not filter by state. This means it can return unregistered executors, unlike `GetExecutors` which does filter.

## Test Coverage Gaps

### 17. Snapshot tests entirely missing
7 `SnapshotDatabase` interface methods have zero test coverage in `embedded_test.go`.

### 18. BlueprintDefinition tests minimal
Only `AddBlueprintDefinition` and `GetBlueprintDefinitionByKind` are tested. Missing tests for: `GetBlueprintDefinitionByName`, `GetBlueprintDefinitions`, `GetBlueprintDefinitionsByNamespace`, `GetBlueprintDefinitionsByGroup`, `UpdateBlueprintDefinition`, `RemoveBlueprintDefinitionByID`, `RemoveBlueprintDefinitionByName`, `CountBlueprintDefinitions`.

### 19. Blueprint tests have gaps
Missing tests for: `GetBlueprints`, `GetBlueprintsByNamespace`, `GetBlueprintsByNamespaceAndKind`, `GetBlueprintsByNamespaceKindAndLocation`, `RemoveBlueprintByName`, `RemoveBlueprintsByNamespace`.

### 20. ApplyRetentionPolicy untested
Contains non-trivial deletion logic and the deadlock bug (issue 1). No test coverage.

### 21. No persistence round-trip test
No test that writes data, closes the database, reopens it, and verifies all data survived through the full `EmbeddedDatabase` API.

### 22. No concurrency tests at EmbeddedDatabase level
All `embedded_test.go` tests are single-threaded. No tests exercise concurrent reads/writes through the database API.
