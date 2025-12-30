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
