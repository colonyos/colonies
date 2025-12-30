# Changes in coverage_and_cleanup Branch

## Summary

This branch focuses on code cleanup, test coverage improvements, and architectural refactoring of the Colonies server. The main themes are:

1. **Handler Refactoring**: Handlers now use direct database calls instead of going through the controller for read operations
2. **Database Connection Pool Configuration**: Added configurable connection pool settings
3. **Test Coverage Improvements**: Added comprehensive unit tests for handlers
4. **Stale Executor Cleanup**: Made stale executor duration configurable via CLI
5. **Code Cleanup**: Removed unused code and simplified the architecture

---

## Database Layer

### Connection Pool Configuration
- Added configurable database connection pool settings in `pkg/database/postgresql/database.go`:
  - `MaxOpenConns`: 100 (default)
  - `MaxIdleConns`: 100 (default)
  - `ConnMaxLifetime`: 5 minutes
  - `ConnMaxIdleTime`: 1 minute
- Pool settings are logged at startup for visibility

### Removed Distributed Lock
- Removed `pkg/database/postgresql/lock.go` and `pkg/database/postgresql/lock_test.go`
- Distributed locking was unused and added unnecessary complexity
- Removed `Lock` and `Unlock` methods from `database.DatabaseCore` interface

### Process Storage Optimization
- Added `process.WaitDeadline = deadline` in `AddProcess` to ensure consistency between in-memory object and database

---

## Handler Refactoring

Handlers have been refactored to use direct database calls for read operations, reducing controller overhead:

### Colony Handlers (`pkg/server/handlers/colony/`)
- `GetColonies`, `GetColony`, `GetColonyStatistics` now call database directly

### Executor Handlers (`pkg/server/handlers/executor/`)
- `GetExecutors`, `GetExecutor`, `GetExecutorByID` now call database directly
- Removed reconciliation trigger code (`TriggerReconciliationForReconciler`)

### Process Handlers (`pkg/server/handlers/process/`)
- `GetProcess`, `GetProcesses`, `FindProcessHistory` now call database directly

### Function Handlers (`pkg/server/handlers/function/`)
- `GetFunctions` now calls database directly

### Log Handlers (`pkg/server/handlers/log/`)
- `GetLogs`, `SearchLogs` now call database directly

### Generator Handlers (`pkg/server/handlers/generator/`)
- `GetGenerators`, `GetGenerator` now call database directly

### Cron Handlers (`pkg/server/handlers/cron/`)
- `GetCrons`, `GetCron` now call database directly

### Attribute Handlers (`pkg/server/handlers/attribute/`)
- `GetAttribute` now calls database directly

### Server Handlers (`pkg/server/handlers/server/`)
- Simplified cluster information retrieval

---

## Controller Changes

### ColoniesController (`pkg/server/controllers/`)
- Added `staleExecutorDuration` field to make cleanup duration configurable
- Added `staleExecutorDuration` parameter to `CreateColoniesController`
- Updated `CleanupWorker` to use configurable stale executor duration
- Added skip logic for executors that have never been heard from (LastHeardFromTime not set)

### Removed Code
- Removed `pkg/server/controllers/controller.go` (unused controller abstraction)
- Simplified `colonies_controller_test.go` by removing redundant tests

---

## Server Adapter Simplification

### Removed Functions from `pkg/server/server_adapter.go`
- Removed `TriggerReconciliationForReconciler` (reconciliation trigger moved/removed)
- Removed unused adapter methods that were no longer needed after handler refactoring
- Significantly reduced file size (297+ lines removed)

---

## CLI Changes

### Server Command (`internal/cli/server.go`)
- Added `--stale-executor-duration` flag to configure how long before an executor is considered stale
- Default value: 600 seconds (10 minutes)

---

## Test Coverage Improvements

### New Unit Test Files
- `pkg/server/handlers/executor/handlers_unit_test.go` - 94.3% coverage for executor handlers
- `pkg/server/handlers/blueprint/handlers_unit_test.go` - 90%+ coverage for blueprint handlers
- `pkg/server/handlers/log/handlers_unit_test.go` - 82.9% coverage for log handlers
- `pkg/server/controllers/colonies_controller_worker_test.go` - Worker tests

### Test Infrastructure
- Added atomic port counter in `pkg/server/controllers/mock_test.go` to prevent port conflicts
- Fixed test flakiness issues related to subscription race conditions
- Improved test isolation

---

## Performance Testing

### K8s Benchmark Improvements (`tests/performance/k8s/`)
- Enhanced `plot_results.py` with better visualization
- Added `GUIDE.md` for running performance tests
- Simplified benchmark configuration
- Removed old experiment data files

---

## Bug Fixes

- Fixed double HTTP error response bug in process handlers
- Fixed subscription race condition causing test flakiness
- Fixed error comparison using `errors.Is()` instead of `==` for sentinel errors

---

## Files Changed Summary

- **65 files changed**
- **1,183 insertions**
- **1,334 deletions**

Net reduction in code complexity while improving test coverage and maintainability.
