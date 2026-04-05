# Embedded Database Concurrency Fix

## Problem

The embedded database has race conditions in multi-step write operations. Two bugs were
found and confirmed with tests:

### Bug 1: Duplicate executor registrations

`AddExecutor` performs a non-atomic read-modify-write across multiple data structures
(executor store + byColony index + byName index). When the HTTP handler calls
`RemoveExecutorByName` followed by `AddExecutor` concurrently for the same executor name,
the interleaving creates duplicate entries:

```
Thread 1: GetExecutorByName("llm") -> found (APPROVED)
Thread 2: GetExecutorByName("llm") -> found (APPROVED)
Thread 1: RemoveExecutorByName("llm") -> marks UNREGISTERED
Thread 2: RemoveExecutorByName("llm") -> no-op (already UNREGISTERED)
Thread 1: AddExecutor(id=A, name="llm") -> finds UNREGISTERED, replaces -> OK
Thread 2: AddExecutor(id=B, name="llm") -> finds A (PENDING), rejects...
          BUT without a lock, Thread 2 may interleave inside AddExecutor
          and both succeed -> 2 executors with name "llm"
```

The byColony MapIndex stores a set of executor IDs per colony. Both IDs get added,
so `GetExecutorsByColonyName` returns duplicates. Over many restarts, dozens accumulate.

PostgreSQL prevents this via PRIMARY KEY on `colonyname:executorname` -- only one row
per name can exist. The embedded DB's MapIndex has no such constraint.

**Test:** `TestConcurrentReregisterNoDuplicates` in `executor_bugs_test.go`
- Before fix: 6-10 duplicates from 10 concurrent re-registers
- After fix: exactly 1 executor

### Bug 2: Lost metric updates

`IncrementMetric` performs a read-modify-write (`Get` -> add delta -> `Put`) without
holding a lock across the operation. Concurrent increments lose updates:

```
Thread 1: Get("tokens") -> value=100
Thread 2: Get("tokens") -> value=100
Thread 1: Put("tokens", 100+50) -> value=150
Thread 2: Put("tokens", 100+50) -> value=150  (should be 200)
```

With multiple LLM executors reporting tokens on every request, this causes persistent
undercounting.

PostgreSQL avoids this with `UPDATE SET value = value + $1` which is atomic at the
row level.

**Test:** `TestConcurrentIncrementMetric` in `metrics_test.go`
- 50 goroutines x 100 increments = expected 5000
- Before fix: typically 200-1000 (massive loss)
- After fix: exactly 5000

## Current Fix (targeted)

Two targeted fixes shipped to unblock development:

1. **Executor mutex** (`executorMu sync.Mutex` on `EmbeddedDatabase`)
   - `AddExecutor` and `RemoveExecutorByName` acquire this mutex
   - Makes the multi-store check-modify-write atomic
   - Single lock, no nesting, no deadlock risk

2. **Metrics store lock** (using Store's built-in `Lock()`/`*Unlocked()` methods)
   - `IncrementMetric` acquires `db.metrics.Lock()` and uses `GetUnlocked`/`PutUnlocked`
   - Atomic read-modify-write without an extra mutex
   - Works because the operation only touches one store

## Generic Fix: Database-level RWMutex

The targeted fixes are ad-hoc. Other entity types (processes, process graphs, generators)
may have similar patterns. A generic solution prevents future bugs.

### Design

Add a single `sync.RWMutex` to `EmbeddedDatabase`:

```go
type EmbeddedDatabase struct {
    mu sync.RWMutex  // database-level transaction lock
    // ... existing fields
}
```

**Write methods** acquire `db.mu.Lock()` and use normal store methods (which still
acquire their own internal locks):

```go
func (db *EmbeddedDatabase) AddExecutor(executor *core.Executor) error {
    db.mu.Lock()
    defer db.mu.Unlock()
    existing, ok := db.executors.Get(...)  // store lock held internally
    ...
    db.executors.Put(...)                  // store lock held internally
}
```

**Read methods** -- most do NOT need `db.mu.RLock()`. A single-store read (e.g.,
`GetExecutorByID` doing one `Get()`) is already atomic via the store's internal lock.
Only reads that span multiple stores and need cross-store consistency would need
`db.mu.RLock()`. In practice, very few reads fall into this category.

### Why keep store-level locks

The plan originally proposed removing store-level locks once `db.mu` is in place
(since `db.mu` makes them redundant). This is **wrong** -- the background flusher
runs on its own goroutine and calls `Store.FlushDirty()`, which uses the store's
internal lock to iterate dirty records. Without store locks, flusher iteration
would race with `Put()` modifications to the dirty set.

Keeping both lock layers:
- `db.mu` provides transaction-level atomicity for multi-step writes
- Store locks protect internal data structures (maps, dirty tracking) for the flusher
- Lock ordering is naturally safe: `db.mu` always acquired first, store locks inside
- Double-locking in-memory maps has negligible overhead

### Why not MVCC

MVCC (Multi-Version Concurrency Control) would require versioned records, snapshot
isolation, conflict detection, garbage collection, and a transaction manager. This is
a substantial rewrite of the Store layer -- essentially building a mini database engine.

The RWMutex approach trades some write concurrency for simplicity:
- Reads are unaffected (no RLock needed for single-store reads)
- Writes serialize (Lock exclusive), but are fast (in-memory maps, microseconds)
- Zero deadlock risk (single lock, no ordering constraints)

The colonies server's workload (moderate write rate, point-lookup reads, no long-running
read transactions) doesn't benefit from MVCC.

### Retention policy: batched deletes

`ApplyRetentionPolicy` iterates all processes, attributes, logs, and process graphs,
deleting expired entries. Under an exclusive `db.mu.Lock()`, this would block all
writes for the entire duration -- unacceptable for a system that must not hang.

**Solution: batch deletes with yield.**

The retention method must NOT hold `db.mu` for the entire sweep. Instead:

1. **Collect phase** (read lock): scan for expired IDs under `db.mu.RLock()`,
   collect up to N candidate IDs per batch
2. **Delete phase** (write lock): acquire `db.mu.Lock()`, delete the batch,
   release the lock
3. **Repeat** until no more expired entries

```go
func (db *EmbeddedDatabase) ApplyRetentionPolicy(retentionPeriod int64) error {
    const batchSize = 100
    cutoff := time.Now().Add(-time.Duration(retentionPeriod) * time.Second)

    for {
        // Collect a batch of expired process IDs (read lock)
        db.mu.RLock()
        var batch []string
        for _, p := range db.processes.All() {
            if p.State == core.SUCCESS && p.SubmissionTime.Before(cutoff) {
                batch = append(batch, p.ID)
                if len(batch) >= batchSize {
                    break
                }
            }
        }
        db.mu.RUnlock()

        if len(batch) == 0 {
            break
        }

        // Delete the batch (write lock)
        db.mu.Lock()
        for _, id := range batch {
            // re-check under write lock (may have been deleted by another goroutine)
            if p, ok := db.processes.Get(id); ok {
                if p.State == core.SUCCESS && p.SubmissionTime.Before(cutoff) {
                    db.removeProcessFromIndexes(p)
                    db.RemoveAllAttributesByTargetID(p.ID)
                    db.processes.Delete(p.ID)
                }
            }
        }
        db.mu.Unlock()
    }

    // Same pattern for logs, process graphs, attributes...
    return nil
}
```

This ensures:
- Each write lock is held for at most `batchSize` deletes (microseconds)
- Other writes (process assignment, metric updates) can proceed between batches
- The system never hangs, even with millions of expired entries
- Re-check under write lock handles races where entries are deleted between phases

## Implementation Plan

### Phase 1: Audit all write methods

Scan all `*.go` files in the embedded package for multi-step write patterns:

```go
// Pattern 1: read-modify-write (dangerous)
existing, ok := db.store.Get(key)
// ... modify ...
db.store.Put(key, modified)

// Pattern 2: check-then-act (dangerous)
existing := db.store.Get(key)
if existing != nil { return error }
db.store.Put(key, new)

// Pattern 3: iterate-delete (safe if single store, but holds lock a while)
for _, item := range db.store.All() {
    db.store.Delete(item.ID)
}
```

Expected locations:
- `executors.go` -- AddExecutor (reactivate UNREGISTERED), RemoveExecutorByName
- `metrics.go` -- IncrementMetric
- `processes.go` -- Assign, Close, Fail, Cancel (state transitions)
- `processgraphs.go` -- state transitions
- `generators.go` -- increment counters
- `crons.go` -- update last run time
- `colonies.go` -- AddColony (duplicate name check)
- `users.go` -- AddUser (duplicate name check)

### Phase 2: Add db.mu and convert write methods

1. Add `mu sync.RWMutex` to `EmbeddedDatabase` struct
2. For each write method:
   - Add `db.mu.Lock()` / `defer db.mu.Unlock()` at the top
   - Keep using normal `db.store.Get()` / `db.store.Put()` (store locks stay)
3. Remove the targeted fixes:
   - Remove `executorMu` from EmbeddedDatabase
   - Revert `IncrementMetric` from `Lock()`/`*Unlocked()` pattern to normal
     `Get()`/`Put()` under `db.mu`

### Phase 3: Add db.mu.RLock to multi-store reads (if any)

Review read methods that query multiple stores in a single call. If any exist and
need consistency, add `db.mu.RLock()`. Single-store reads (the vast majority) need
no changes -- the store's internal lock is sufficient.

### Phase 4: Concurrency test suite

The existing tests mostly run single-threaded. After adding `db.mu`, we need a
comprehensive concurrent test suite that exercises real contention patterns. All
tests must pass with `go test -race`.

#### 4a: Entity-level concurrent CRUD

For each major entity type, test concurrent create/read/update/delete. The pattern
is: N goroutines doing operations on the same entity type simultaneously. Verify
no panics, no data corruption, no lost writes.

| Test | What it does | Verifies |
|------|-------------|----------|
| `TestConcurrentAddRemoveExecutors` | 20 goroutines each adding and removing unique executors | No duplicate index entries, correct count after all complete |
| `TestConcurrentReregisterNoDuplicates` | (existing) 10 goroutines re-registering same executor name | Exactly 1 executor at the end |
| `TestConcurrentIncrementMetric` | (existing) 50 goroutines incrementing same counter | Exact sum, no lost updates |
| `TestConcurrentSetGaugeMetrics` | 20 goroutines setting different gauge keys on same executor | All keys present, no cross-contamination |
| `TestConcurrentAddRemoveProcesses` | 20 goroutines adding processes, 10 goroutines removing them | No panics, no orphaned index entries |
| `TestConcurrentProcessStateTransitions` | Add processes, then concurrently assign/close/fail them | Each process ends in exactly one terminal state |
| `TestConcurrentAddRemoveAttributes` | Concurrent attribute add/remove on different target IDs | Correct attribute counts per target |
| `TestConcurrentAddRemoveColonies` | 10 goroutines creating colonies with unique names | All colonies exist, no duplicates |
| `TestConcurrentAddRemoveUsers` | Concurrent user creation across different colonies | Correct user counts per colony |
| `TestConcurrentAddRemoveFunctions` | Register/remove functions concurrently | No orphaned index entries |
| `TestConcurrentAddRemoveCrons` | Add/update/remove crons concurrently | Consistent state |
| `TestConcurrentAddRemoveGenerators` | Add/remove generators concurrently | Consistent state |
| `TestConcurrentLogWrites` | 50 goroutines writing logs concurrently | All logs present, correct count |
| `TestConcurrentFileOperations` | Add/remove files concurrently | Sequence numbers never collide |

#### 4b: Cross-entity concurrent operations

These test operations that touch multiple entity types simultaneously, which is
the pattern most likely to deadlock or produce inconsistent state.

| Test | What it does | Verifies |
|------|-------------|----------|
| `TestConcurrentProcessWithAttributes` | Add processes and their attributes concurrently | Attributes correctly linked to processes |
| `TestConcurrentProcessGraphWithProcesses` | Create process graphs while processes are being modified | Graph state consistent with child process states |
| `TestConcurrentExecutorWithFunctions` | Add/remove executors while registering functions | No orphaned functions for removed executors |
| `TestConcurrentRetentionDuringWrites` | Run ApplyRetentionPolicy while adding new processes | New processes not accidentally deleted, retention completes |
| `TestConcurrentMetricsDuringExecutorRemove` | Set metrics while removing the executor | No panics, metrics for removed executor are cleanable |

#### 4c: Server simulation stress tests

These are the most important tests. They simulate realistic multi-executor server
workloads with dozens of goroutines hitting the database simultaneously for an
extended duration. The goal is to prove: no deadlocks, no panics, no data corruption,
no race detector violations.

Each simulation runs for a fixed duration (e.g., 3 seconds) with a `context.Context`
controlling shutdown. Goroutines run in a loop until the context is cancelled, then
a verification phase checks invariants.

**Simulation 1: LLM executor fleet**

Models a production deployment with multiple LLM executors processing requests,
reporting token metrics, and periodically re-registering.

```
TestSimulationLLMFleet (duration: 3s):

  Setup:
    - 1 colony
    - 5 executor names ("llm-1" through "llm-5")

  Concurrent actors:
    - 5 executor heartbeat goroutines:
        loop: MarkAlive(executor)
    - 10 process submitter goroutines:
        loop: AddProcess(random funcspec targeting one of the 5 executors)
    - 5 process worker goroutines (one per executor):
        loop: find a WAITING process, Assign it, sleep briefly, Close it
    - 5 metric reporter goroutines (one per executor):
        loop: IncrementMetric("tokens_used", PERIOD_DAY, random delta 1-100)
              IncrementMetric("tokens_used", PERIOD_NONE, same delta)
              SetMetric("gpu_temp", GAUGE, random 60-90)
    - 3 metric reader goroutines:
        loop: GetMetrics(random executor)
              GetMetricHistory("tokens_used", PERIOD_DAY, last 7 days)
    - 1 executor re-register goroutine:
        loop: pick random executor, RemoveExecutor, AddExecutor (re-register)

  Verification:
    - No panics during execution
    - No -race violations
    - Each executor name exists exactly once
    - All PERIOD_NONE token counters > 0
    - No WAITING processes left assigned to a non-existent executor
    - Total tokens across PERIOD_DAY buckets == PERIOD_NONE total (per executor)
```

**Simulation 2: Multi-tenant colony**

Models multiple colonies with executors, processes, and cross-cutting operations
like retention and colony deletion.

```
TestSimulationMultiTenant (duration: 3s):

  Setup:
    - 3 colonies ("tenant-1" through "tenant-3")
    - 2 executors per colony
    - 2 users per colony

  Concurrent actors:
    - 6 process submitter goroutines (2 per colony):
        loop: AddProcess, add input attributes
    - 6 process worker goroutines (1 per executor):
        loop: Assign process, add output attributes, Close process
    - 3 function registrar goroutines (1 per colony):
        loop: AddFunction, RemoveFunction for random function names
    - 3 cron manager goroutines (1 per colony):
        loop: AddCron, GetCrons, RemoveCron
    - 2 log writer goroutines:
        loop: AddLog for random processes
    - 1 retention goroutine:
        loop: ApplyRetentionPolicy(short period to force deletes)
    - 1 colony destroyer goroutine:
        loop: after 1s, remove colony "tenant-3" and all its entities,
              then re-create it

  Verification:
    - No panics, no -race violations
    - Colonies "tenant-1" and "tenant-2" have consistent state
    - Colony "tenant-3" exists (was re-created)
    - No orphaned attributes (attributes whose target process doesn't exist)
    - No orphaned functions (functions whose executor doesn't exist)
```

**Simulation 3: Burst traffic with contention**

Stress-tests the hot path: process assignment under heavy contention. Many executors
compete for a limited number of WAITING processes.

```
TestSimulationBurstAssignment (duration: 3s):

  Setup:
    - 1 colony
    - 20 executors (all competing for same processes)

  Concurrent actors:
    - 5 submitter goroutines:
        loop: submit a process with generic funcspec (any executor can take it)
    - 20 assigner goroutines (one per executor):
        loop: try to assign a WAITING process, if successful close it after brief work
    - 5 reader goroutines:
        loop: GetProcesses(WAITING), GetProcesses(RUNNING), count them
    - 2 canceller goroutines:
        loop: find a WAITING process, cancel it (competing with assigners)

  Verification:
    - No process assigned to two different executors
    - No process in both RUNNING and SUCCESS/FAILED/CANCELLED
    - Total (SUCCESS + FAILED + CANCELLED + WAITING + RUNNING) == total submitted
    - No -race violations
```

**Simulation 4: Deadlock detector**

Specifically designed to trigger deadlock if lock ordering is wrong. Runs operations
that would require multiple locks in different orders if the implementation were
using per-entity locks instead of a single db.mu.

```
TestSimulationDeadlockDetector (duration: 5s, timeout: 10s):

  Concurrent actors:
    - 10 goroutines: AddProcess (touches processes store + attributes + indexes)
    - 10 goroutines: Close process (touches processes + attributes + process graphs)
    - 5 goroutines: AddExecutor/RemoveExecutor (touches executors + functions)
    - 5 goroutines: ApplyRetentionPolicy (touches processes + attributes + logs + graphs)
    - 5 goroutines: RemoveAllProcessesByColonyName (touches processes + attributes)
    - 5 goroutines: AddProcessGraph + modify child process states

  Verification:
    - Test completes within timeout (deadlock = test hangs = timeout failure)
    - No panics, no -race violations

  The test timeout is the deadlock detector: if db.mu ordering is wrong or
  there's a nested lock acquisition, the test will hang and fail via timeout.
```

#### Implementation pattern for simulations

All simulations follow the same structure:

```go
func TestSimulationLLMFleet(t *testing.T) {
    db := setupTestDB(t)
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    // Setup: create colony, executors, etc.
    colony := core.CreateColony(core.GenerateRandomID(), "test-colony")
    db.AddColony(colony)
    // ...

    var wg sync.WaitGroup
    var errors atomic.Int64  // count non-fatal errors (e.g., "process not found" races)

    // Launch actors
    wg.Add(1)
    go func() {
        defer wg.Done()
        for {
            select {
            case <-ctx.Done():
                return
            default:
                // do work, count errors
            }
        }
    }()

    // ... more actors ...

    wg.Wait()

    // Verification phase
    executors, _ := db.GetExecutorsByColonyName("test-colony", false)
    assert.Equal(t, 5, len(executors), "expected exactly 5 executors")
    // ... more invariant checks ...
}
```

#### 4d: Persistence under concurrency

Verify that concurrent writes are correctly persisted and recoverable.

| Test | What it does | Verifies |
|------|-------------|----------|
| `TestConcurrentWritesThenRestart` | 20 goroutines writing different entities, then Close and reopen DB | All written data present after restart |
| `TestConcurrentWritesDuringFlush` | Write rapidly while flusher is running | No corruption, WAL replay consistent |

#### Running the suite

```bash
# Full suite with race detector
go test -race -count=1 -timeout 120s ./pkg/database/embedded/

# Just concurrency tests
go test -race -v -run "TestConcurrent" -timeout 60s ./pkg/database/embedded/

# Stress test with multiple iterations to catch rare races
go test -race -count=10 -run "TestConcurrentMixedWorkload" ./pkg/database/embedded/
```

### Phase 5: Architecture documentation

Generate `pkg/database/embedded/README.md` -- a comprehensive document describing
the embedded database architecture. This serves as onboarding material for new
contributors and as a design reference for future changes.

#### Contents

1. **Overview and motivation**
   - Why an embedded DB exists alongside PostgreSQL
   - Use cases: single-server deployments, edge nodes, development/testing
   - Design goals: zero external dependencies, fast reads, crash recovery

2. **Storage architecture**
   - Three-layer design: in-memory store -> WAL -> disk store
   - All reads served from memory (O(1) map lookups)
   - Writes: WAL append (fsync) -> memory update -> async flush to disk
   - Why this hybrid approach vs pure WAL (like SQLite) or pure in-memory (like Redis)

3. **Write-Ahead Log (WAL)**
   - Purpose: crash recovery without fsync on every write to individual files
   - Binary format: CRC32 checksum, operation type (Put/Delete), entity name, key, JSON data
   - Sync modes: SyncAlways (safe, slower) vs SyncNone (fast, risk of last few writes)
   - Replay on startup: WAL entries applied to in-memory stores before disk load
   - Truncation: WAL truncated after successful flush of all dirty records
   - References:
     - "ARIES: A Transaction Recovery Method" (Mohan et al., 1992) -- foundational WAL paper
     - PostgreSQL WAL documentation: https://www.postgresql.org/docs/current/wal-intro.html
     - SQLite WAL mode: https://www.sqlite.org/wal.html
     - "Designing Data-Intensive Applications" (Kleppmann, 2017), Chapter 3: Storage and Retrieval

4. **In-memory store (Store)**
   - Generic `Store[K, V]` with type parameters
   - Internal `sync.RWMutex` for per-store thread safety
   - Dirty tracking: records modified since last flush marked dirty
   - `Put()`, `Get()`, `Delete()`, `All()` -- locked variants
   - `PutUnlocked()`, `GetUnlocked()`, `DeleteUnlocked()` -- for callers holding external locks
   - References:
     - Go generics: https://go.dev/doc/tutorial/generics

5. **Disk store**
   - One JSON file per record: `data/<entity>/<sanitized_key>.json`
   - Key sanitization: escapes `/`, `\`, `:` etc. for safe filenames
   - Atomic writes: write to temp file, then rename (POSIX atomic rename guarantee)
   - Loaded on startup after WAL replay (WAL entries take precedence over disk)
   - References:
     - POSIX rename atomicity: https://pubs.opengroup.org/onlinepubs/9699919799/functions/rename.html
     - "Don't fear the fsync" -- best practices for durable file writes

6. **Background flusher**
   - Periodic goroutine (default: every 5 seconds)
   - Iterates dirty records in each store, writes to disk store
   - Holds store-level read lock during iteration (does not hold db.mu)
   - After flush: WAL can be truncated (all dirty data now on disk)
   - Trade-off: longer flush interval = more data at risk on crash, less I/O

7. **Indexing**
   - `MapIndex[K, V]`: maps a lookup key to a set of primary keys
   - Example: `byColony` maps colony name -> set of executor IDs
   - `CompoundIndex`: two-level index for state-based queries (colony -> state -> ordered set)
   - Indexes are not persisted -- rebuilt from store data on startup (`rebuildIndexes`)
   - Why not persist indexes: they're derived data, rebuilding is fast, avoids consistency bugs
   - References:
     - "Database Internals" (Petrov, 2019), Chapter 6: B-Tree Variants -- general index theory

8. **Concurrency model**
   - Database-level `sync.RWMutex` (`db.mu`) for transaction atomicity on writes
   - Store-level `sync.RWMutex` for internal data structure protection (maps, dirty tracking)
   - Lock ordering: `db.mu` always acquired before store locks (no deadlock)
   - Single-store reads do not need `db.mu` (store lock sufficient)
   - Retention policy uses batched deletes to avoid long lock holds
   - References:
     - "The Art of Multiprocessor Programming" (Herlihy & Shavit), Chapter 8: Monitors and Blocking
     - Go sync.RWMutex: https://pkg.go.dev/sync#RWMutex
     - Why not MVCC: see PLAN_CONCURRENCY.md for detailed rationale

9. **Startup and recovery sequence**
   - `Initialize()` flow:
     1. Create data directories
     2. Open WAL file
     3. Create all stores (without WAL -- no logging during replay)
     4. Replay WAL entries into in-memory stores
     5. Load remaining records from disk (WAL entries take precedence)
     6. Set WAL on all stores (future writes are logged)
     7. Rebuild all indexes from loaded data
     8. Start background flusher
   - Crash recovery: WAL replay reconstructs any writes not yet flushed to disk
   - Clean shutdown: `Close()` flushes all dirty records, truncates WAL

10. **Comparison with PostgreSQL implementation**
    - Same `Database` interface, different trade-offs
    - PostgreSQL: ACID transactions, SQL queries, connection pooling, TimescaleDB for time-series
    - Embedded: zero dependencies, in-process, microsecond reads, limited query capabilities
    - Feature parity: both implement the full `Database` interface (compile-time checked)
    - When to use which: PostgreSQL for production multi-server, embedded for edge/dev/single-node

11. **Limitations and future work**
    - No range queries (all filtering is linear scan over index results)
    - No join operations (cross-entity queries done at application level)
    - Write serialization under `db.mu` (acceptable for current workload, MVCC if needed later)
    - WAL grows unbounded between flushes (mitigated by periodic truncation)
    - No built-in backup mechanism (copy data directory while stopped, or snapshot WAL)

#### References (consolidated)

- Mohan et al., "ARIES: A Transaction Recovery Method Supporting Fine-Granularity Locking
  and Partial Rollbacks Using Write-Ahead Logging", ACM TODS, 1992
- Kleppmann, "Designing Data-Intensive Applications", O'Reilly, 2017
- Petrov, "Database Internals", O'Reilly, 2019
- Herlihy & Shavit, "The Art of Multiprocessor Programming", Morgan Kaufmann, 2012
- PostgreSQL WAL: https://www.postgresql.org/docs/current/wal-intro.html
- SQLite WAL: https://www.sqlite.org/wal.html
- Go sync package: https://pkg.go.dev/sync
- BoltDB design (similar embedded approach): https://github.com/etcd-io/bbolt
- BadgerDB (LSM-tree alternative): https://dgraph.io/docs/badger/

## Files to modify

| File | Write methods needing db.mu.Lock() |
|------|-----------------------------------|
| `database.go` | Add `mu sync.RWMutex` field, ApplyRetentionPolicy |
| `executors.go` | AddExecutor, ApproveExecutor, RejectExecutor, MarkAlive, RemoveExecutorByName, RemoveExecutorsByColonyName, UpdateExecutorCapabilities, SetAllocations |
| `metrics.go` | SetMetric, IncrementMetric, RemoveMetric, RemoveAllMetrics* |
| `processes.go` | AddProcess, SetProcessState, Assign, Close, Fail, Cancel, Remove* |
| `processgraphs.go` | AddProcessGraph, SetGraphState, Remove* |
| `colonies.go` | AddColony, RemoveColony* |
| `users.go` | AddUser, RemoveUser* |
| `functions.go` | AddFunction, UpdateFunctionStats, RemoveFunction* |
| `generators.go` | AddGenerator, Remove*, update counters |
| `crons.go` | AddCron, UpdateCron, RemoveCron* |
| `logs.go` | AddLog, RemoveLogs* |
| `files.go` | AddFile, RemoveFile* |
| `snapshots.go` | AddSnapshot, RemoveSnapshot* |
| `blueprints.go` | Add/Update/Remove operations |
| `locations.go` | Add/Update/Remove operations |
| `security.go` | SetServerID |
| `attributes.go` | AddAttribute, UpdateAttribute, Remove*, setAttributeState |
