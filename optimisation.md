# Optimization Notes

## Signal Mechanism Location-Aware Routing

**Status:** Not implemented (works correctly, optimization only)

**Current Behavior:**
When a process is added, the event handler's `Signal()` function wakes up ONE waiting executor based on `executorType` only. The target key is constructed from `executorType + state`.

**Issue:**
With multiple executors of the same type at different locations (e.g., multiple `docker-reconciler` instances), the signal may wake the wrong executor:

1. Process for `home_linux_server` is submitted
2. Signal wakes ONE `docker-reconciler` (round-robin selection)
3. If the `edge` reconciler gets the signal, it calls `Assign()`
4. Scheduler correctly filters by location - no match found
5. Edge reconciler goes back to waiting
6. The `home_linux_server` reconciler was never signaled
7. Process is picked up when the correct reconciler's poll timeout expires or cron triggers

**Impact:**
- Wasted round-trip when wrong executor is signaled
- Latency for location-specific processes (up to poll timeout, typically 100 seconds)
- Not a correctness issue - scheduler ensures correct assignment

**Potential Fix:**
Include location in the signal target key:

1. Modify `WaitForProcess(executorType, state, processID, location, ctx)`
2. Modify `target()` function to include location: `executorType + ":" + location + ":" + state`
3. Update `sendSignal()` to use process's `FunctionSpec.Conditions.LocationName`
4. Update all callers of `WaitForProcess` to pass executor's location

**Files to modify:**
- `pkg/backends/gin/eventhandler.go` - target key generation, WaitForProcess signature
- `pkg/backends/libp2p/eventhandler.go` - same changes for libp2p backend
- `pkg/backends/interfaces.go` - EventHandler interface
- `pkg/server/handlers/process/handlers.go` - HandleAssignProcess caller

**Considerations:**
- Backwards compatibility with processes that have no location specified
- Empty location should match all executors (broadcast to all locations)
- Need to handle case where executor registers without location
