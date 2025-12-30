# TODO

## Refactor: Remove hardcoded "reconcile" function name checks

**Problem:**

The server code has hardcoded checks for `FuncName == "reconcile"` in multiple places to detect reconciliation processes and update blueprint metadata. This is a code smell with several issues:

- Magic string `"reconcile"` - breaks silently if function is renamed
- Tight coupling between unrelated components and blueprint metadata
- Violates single responsibility principle
- Not extensible

### Affected Files

#### 1. Cron Controller
**File:** `pkg/server/controllers/cron_controller.go` (lines 236-274)

```go
if len(workflowSpec.FunctionSpecs) > 0 && workflowSpec.FunctionSpecs[0].FuncName == "reconcile" {
    // Updates blueprint.Metadata.LastReconciliationProcess
    // Updates blueprint.Metadata.LastReconciliationTime
}
```

**Purpose:** When a cron job fires, detect if it's a reconciliation cron and update the blueprint with the process ID and timestamp.

#### 2. Process Handlers (Close Successful)
**File:** `pkg/server/handlers/process/handlers.go` (line 831)

```go
if blueprintID == "" && process.FunctionSpec.FuncName == "reconcile" {
    if bpName, ok := process.FunctionSpec.KwArgs["blueprintName"].(string); ok {
        blueprintName = bpName
        // ... updates blueprint metadata when process closes successfully
    }
}
```

**Purpose:** When a process closes successfully, detect if it's a reconciliation process and update the blueprint status.

#### 3. Blueprint Handlers (Source of Truth)
**File:** `pkg/server/handlers/blueprint/handlers.go` (lines 125, 161, 165, 205, 209, 1316, 1320)

```go
funcSpec.FuncName = "reconcile"
funcSpec.NodeName = "reconcile"
```

**Purpose:** This is where the convention is established - when creating reconciliation processes, the FuncName is set to "reconcile". This is the source of truth.

### Current Behavior

Blueprint metadata tracking happens in two places:
1. **Cron trigger time:** `cron_controller.go` sets `LastReconciliationTime` when cron fires
2. **Process completion:** `process/handlers.go` updates status when reconciliation completes

### Proposed Solution

Move all blueprint metadata updates to the reconciler executor itself:

1. When the reconciler executor successfully completes reconciliation, it should call an API to update blueprint metadata
2. Remove the hardcoded checks from `cron_controller.go` and `process/handlers.go`
3. The server should provide a dedicated endpoint for reconcilers to report completion status

**Benefits:**
- Cron controller and process handlers don't need to know about blueprints
- Metadata is updated regardless of how reconciliation was triggered (manual, cron, or API)
- No magic strings in the server code
- Single source of truth for reconciliation completion
- Reconciler has full context about what happened and can provide richer metadata

### External Configuration Dependency

The function name is also defined in external blueprint definition files:

**File:** `executors/docker-reconciler/blueprint-definitions/docker-deployment-definition.json`

```json
"handler": {
  "executorType": "docker-reconciler",
  "functionName": "reconcile"
}
```

This creates **invisible coupling** between:
- External JSON configuration (`"functionName": "reconcile"`)
- Server Go code (`FuncName == "reconcile"`)

**No validation exists** - if they don't match, things silently break.

### Failure Modes

#### 1. Silent Breakage
If someone changes the blueprint definition to use a different function name:
```json
"functionName": "sync"  // or "deploy", "update", etc.
```

The server will:
- Still create processes with the new function name
- But hardcoded checks won't match
- Blueprint metadata never gets updated
- No errors, no warnings - just silent failure

#### 2. Unintended Side Effects
If someone creates an unrelated executor with a function named "reconcile":
```json
{
  "executorType": "inventory-manager",
  "functionName": "reconcile"  // reconcile inventory, not blueprints!
}
```

The server will:
- Incorrectly detect this as a blueprint reconciliation
- Try to update blueprint metadata using KwArgs["blueprintName"]
- Cause unexpected behavior or errors

### Alternative Solutions

1. **Use FunctionSpec metadata:** Add a field like `FunctionSpec.Reconciliation.BlueprintName` instead of checking function names

2. **Use labels/annotations:** Mark reconciliation processes with a label that can be checked without hardcoding function names

3. **Event-driven:** Emit events when processes complete, let blueprint service subscribe to relevant events

4. **Explicit marker in handler config:** Add a boolean flag in blueprint definition:
   ```json
   "handler": {
     "executorType": "docker-reconciler",
     "functionName": "reconcile",
     "isReconciler": true  // Explicit marker, not inferred from name
   }
   ```

---

## Incremental Improvement Tasks

The following tasks can be done incrementally to improve code quality, test coverage, and architecture.

### Database Layer

- [ ] **Add database connection pool configuration**
  - File: `pkg/database/postgresql/database.go`
  - Add configurable settings: `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`, `ConnMaxIdleTime`
  - Log pool settings at startup

- [ ] **Remove unused distributed lock**
  - Remove `pkg/database/postgresql/lock.go` and `pkg/database/postgresql/lock_test.go`
  - Remove `Lock` and `Unlock` methods from `database.DatabaseCore` interface

- [ ] **Add process WaitDeadline optimization**
  - File: `pkg/database/postgresql/processes.go`
  - Set `process.WaitDeadline = deadline` in `AddProcess` for consistency

### Handler Refactoring (Direct DB Calls for Reads)

The goal is to remove controller adapter indirection for read operations. Handlers should call the database directly instead of going through `server_adapter.go` controller wrappers.

#### Functions to remove from `pkg/server/server_adapter.go`:

**Colony Controller Adapter:**
- `GetColonies() ([]*core.Colony, error)`
- `GetColony(colonyName string) (*core.Colony, error)`
- `GetColonyStatistics(colonyName string) (*core.Statistics, error)`
- `AddColony(colony *core.Colony) (*core.Colony, error)`
- `RemoveColony(colonyName string) error`
- `Controller() colony.Controller`

**Executor Controller Adapter:**
- `GetExecutor(executorID string) (*core.Executor, error)`
- `GetExecutorByColonyName(colonyName string) ([]*core.Executor, error)`
- `AddExecutor(executor *core.Executor, allowReregister bool) (*core.Executor, error)`
- `ExecutorController() executor.Controller`

**Process Controller Adapter:**
- `GetProcess(processID string) (*core.Process, error)`
- `GetExecutor(executorID string) (*core.Executor, error)`
- `FindProcessHistory(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error)`
- `SetOutput(processID string, output []interface{}) error`

**Function Controller Adapter:**
- `GetFunction(functionID string) (*core.Function, error)`
- `GetFunctions(colonyName string, executorName string, count int) ([]*core.Function, error)`
- `GetFunctionsByColonyName(colonyName string) ([]*core.Function, error)`
- `AddFunction(function *core.Function) (*core.Function, error)`
- `RemoveFunction(functionID string, initiatorID string) error`
- `FunctionController() functionhandlers.Controller`

**Generator Controller Adapter:**
- `GetGenerator(generatorID string) (*core.Generator, error)`
- `GetGenerators(colonyName string, count int) ([]*core.Generator, error)`
- `ResolveGenerator(colonyName string, generatorName string) (*core.Generator, error)`

**Cron Controller Adapter:**
- `GetCron(cronID string) (*core.Cron, error)`
- `GetCrons(colonyName string, count int) ([]*core.Cron, error)`
- `GetCronByName(colonyName string, cronName string) (*core.Cron, error)`

**Log Controller Adapter:**
- `GetProcess(processID string) (*core.Process, error)`
- `LogProcessController() loghandlers.Controller`

**Attribute Controller Adapter:**
- `GetProcess(processID string) (*core.Process, error)`
- `GetAttribute(attributeID string) (*core.Attribute, error)`
- `AddAttribute(attribute *core.Attribute) (*core.Attribute, error)`
- `AttributeController() attributehandlers.Controller`

**Server Controller Adapter:**
- `GetStatistics() (*core.Statistics, error)`

**Other:**
- `TriggerReconciliationForReconciler(colonyName, executorType, locationName string) error`
- `GetController() interface{}`

#### Specific Call Changes (Controller -> Direct DB):

**Colony Handlers** (`pkg/server/handlers/colony/handlers.go`):
| Before (Controller) | After (Direct DB) |
|---------------------|-------------------|
| `Controller().AddColony()` | `ColonyDB().AddColony()` + `ColonyDB().GetColonyByID()` |
| `Controller().RemoveColony()` | `ColonyDB().RemoveColonyByName()` |
| `Controller().GetColonies()` | `ColonyDB().GetColonies()` |
| `Controller().GetColony()` | `ColonyDB().GetColonyByName()` |
| `Controller().GetColonyStatistics()` | Multiple direct DB calls for counts |

**Executor Handlers** (`pkg/server/handlers/executor/handlers.go`):
| Before (Controller) | After (Direct DB) |
|---------------------|-------------------|
| `ExecutorController().GetExecutor()` | `ExecutorDB().GetExecutorByID()` |
| `ExecutorController().GetExecutorByColonyName()` | `ExecutorDB().GetExecutorsByColonyName()` |

**Process Handlers** (`pkg/server/handlers/process/handlers.go`):
| Before (Controller) | After (Direct DB) |
|---------------------|-------------------|
| `ProcessController().GetProcess()` | `ProcessDB().GetProcessByID()` |
| `ProcessController().GetExecutor()` | `ExecutorDB().GetExecutorByID()` |
| `ProcessController().FindProcessHistory()` | `ProcessDB().FindProcessesByColonyName()` or `FindProcessesByExecutorID()` |
| `ProcessController().SetOutput()` | `ProcessDB().SetOutput()` |

**Function Handlers** (`pkg/server/handlers/function/handlers.go`):
| Before (Controller) | After (Direct DB) |
|---------------------|-------------------|
| `FunctionController().GetFunctions()` | `FunctionDB().GetFunctions()` |
| `FunctionController().GetFunctionsByColonyName()` | `FunctionDB().GetFunctionsByColonyName()` |

**Generator Handlers** (`pkg/server/handlers/generator/handlers.go`):
| Before (Controller) | After (Direct DB) |
|---------------------|-------------------|
| `GeneratorController().GetGenerator()` | `GeneratorDB().GetGeneratorByID()` |
| `GeneratorController().GetGenerators()` | `GeneratorDB().FindGeneratorsByColonyName()` |
| `GeneratorController().ResolveGenerator()` | `GeneratorDB().GetGeneratorByName()` |

**Cron Handlers** (`pkg/server/handlers/cron/handlers.go`):
| Before (Controller) | After (Direct DB) |
|---------------------|-------------------|
| `CronController().GetCron()` | `CronDB().GetCronByID()` |
| `CronController().GetCrons()` | `CronDB().FindCronsByColonyName()` |
| `CronController().GetCronByName()` | `CronDB().GetCronByName()` |

**Log Handlers** (`pkg/server/handlers/log/handlers.go`):
| Before (Controller) | After (Direct DB) |
|---------------------|-------------------|
| `LogProcessController().GetProcess()` | `ProcessDB().GetProcessByID()` |

**Attribute Handlers** (`pkg/server/handlers/attribute/handlers.go`):
| Before (Controller) | After (Direct DB) |
|---------------------|-------------------|
| `AttributeController().GetProcess()` | `ProcessDB().GetProcessByID()` |
| `AttributeController().AddAttribute()` | `AttributeDB().AddAttribute()` + `GetAttributeByID()` |
| `AttributeController().GetAttribute()` | `AttributeDB().GetAttributeByID()` |

**Server Handlers** (`pkg/server/handlers/server/handlers.go`):
| Before (Controller) | After (Direct DB) |
|---------------------|-------------------|
| `Controller().GetStatistics()` | Multiple direct DB Count calls |

#### Tasks:

- [ ] **Refactor colony handlers**
  - File: `pkg/server/handlers/colony/handlers.go`
  - Make `GetColonies`, `GetColony`, `GetColonyStatistics` call database directly
  - Remove `colony.Controller` interface dependency

- [ ] **Refactor executor handlers**
  - File: `pkg/server/handlers/executor/handlers.go`
  - Make `GetExecutors`, `GetExecutor`, `GetExecutorByID` call database directly
  - Remove `TriggerReconciliationForReconciler` code
  - Remove `executor.Controller` interface dependency

- [ ] **Refactor process handlers**
  - File: `pkg/server/handlers/process/handlers.go`
  - Make `GetProcess`, `GetProcesses`, `FindProcessHistory` call database directly

- [ ] **Refactor function handlers**
  - File: `pkg/server/handlers/function/handlers.go`
  - Make `GetFunctions`, `GetFunctionsByColonyName` call database directly
  - Remove `functionhandlers.Controller` interface dependency

- [ ] **Refactor log handlers**
  - File: `pkg/server/handlers/log/handlers.go`
  - Make `GetLogs`, `SearchLogs` call database directly
  - Remove `loghandlers.Controller` interface dependency

- [ ] **Refactor generator handlers**
  - File: `pkg/server/handlers/generator/handlers.go`
  - Make `GetGenerators`, `GetGenerator`, `ResolveGenerator` call database directly

- [ ] **Refactor cron handlers**
  - File: `pkg/server/handlers/cron/handlers.go`
  - Make `GetCrons`, `GetCron`, `GetCronByName` call database directly

- [ ] **Refactor attribute handlers**
  - File: `pkg/server/handlers/attribute/handlers.go`
  - Make `GetAttribute` call database directly
  - Remove `attributehandlers.Controller` interface dependency

- [ ] **Simplify server handlers**
  - File: `pkg/server/handlers/server/handlers.go`
  - Simplify cluster information retrieval

### Controller Changes

- [ ] **Add configurable stale executor duration**
  - File: `pkg/server/controllers/colonies_controller.go`
  - Add `staleExecutorDuration` field to `ColoniesController` struct
  - Add parameter to `CreateColoniesController`
  - Update `CleanupWorker` to use configurable duration

- [ ] **Add skip logic for new executors in cleanup**
  - File: `pkg/server/controllers/colonies_controller_worker.go`
  - Skip executors where `LastHeardFromTime` is not set

- [ ] **Remove unused controller abstraction**
  - Remove `pkg/server/controllers/controller.go`
  - Simplify `colonies_controller_test.go`

### Server Adapter Cleanup

- [ ] **Remove unused adapter methods**
  - File: `pkg/server/server_adapter.go`
  - Remove `TriggerReconciliationForReconciler`
  - Remove adapter methods no longer needed after handler refactoring

### CLI Changes

- [ ] **Add stale executor duration CLI flag**
  - File: `internal/cli/server.go`
  - Add `--stale-executor-duration` flag
  - Default: 600 seconds (10 minutes)

### Test Coverage

- [ ] **Add executor handler unit tests**
  - File: `pkg/server/handlers/executor/handlers_unit_test.go`
  - Target: 90%+ coverage

- [ ] **Add blueprint handler unit tests**
  - File: `pkg/server/handlers/blueprint/handlers_unit_test.go`
  - Target: 90%+ coverage

- [ ] **Add log handler unit tests**
  - File: `pkg/server/handlers/log/handlers_unit_test.go`
  - Target: 80%+ coverage

- [ ] **Add controller worker tests**
  - File: `pkg/server/controllers/colonies_controller_worker_test.go`

### Test Infrastructure

- [ ] **Fix port conflicts in tests**
  - File: `pkg/server/controllers/mock_test.go`
  - Add atomic port counter to prevent conflicts

- [ ] **Fix subscription race conditions**
  - Fix test flakiness related to subscription timing

### Bug Fixes

These are already fixed in main:
- [x] **Fix double HTTP error response** (commit `525b1af3`)
- [x] **Fix subscription race condition** (commit `3de00b38`)
