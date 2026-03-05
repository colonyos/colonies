# Plan: Embedded Database with Disk-Backed Persistence

## Goal

Build a custom embedded database implementation (`pkg/database/embedded/`) that implements the
full `Database` interface (16 sub-interfaces, ~200 methods) using in-memory data structures as
a cache with background flushing to disk for durability. No external dependencies (no
PostgreSQL, no SQLite).

The core infrastructure (Phases 1-2) is the foundation everything else depends on. It must be
tested extensively before building the entity implementations on top of it.

---

## Architecture

```
Write path:
  Method call -> WAL append -> memory mutation -> return
                                    |
                                    v  (background goroutine, periodic)
                                  disk files

Read path:
  Method call -> check memory -> if miss, load from disk -> return

Startup:
  Scan disk -> rebuild indexes (keys + sort fields only) -> replay WAL -> ready
```

All entity types go through the same unified path. No special cases per entity.

---

## Phase 1: Foundation - Disk Store and WAL

### 1.1 DiskStore

Location: `pkg/database/embedded/diskstore/`

A generic, entity-agnostic record store on disk.

**Layout:**
```
<datadir>/
  records/
    processes/<id>.json
    executors/<id>.json
    colonies/<id>.json
    users/<id>.json
    attributes/<id>.json
    processgraphs/<id>.json
    generators/<id>.json
    generatorargs/<id>.json
    crons/<id>.json
    logs/<processID>/<timestamp>-<seq>.json
    files/<id>.json
    snapshots/<id>.json
    blueprintdefs/<id>.json
    blueprints/<id>.json
    blueprinthistory/<id>.json
    locations/<id>.json
    server/server.json
```

**Interface:**
```go
type DiskStore[V any] struct { ... }

func NewDiskStore[V any](baseDir string, entityName string) *DiskStore[V]
func (d *DiskStore[V]) Write(key string, value V) error      // JSON marshal + write + fsync
func (d *DiskStore[V]) Read(key string) (V, error)           // read + JSON unmarshal
func (d *DiskStore[V]) Delete(key string) error              // remove file
func (d *DiskStore[V]) List() ([]string, error)              // list all keys (directory listing)
func (d *DiskStore[V]) Scan(fn func(key string, value V) error) error  // iterate all records
```

JSON encoding for debuggability. Records are individual files keyed by primary key.

**Estimated effort:** 1-2 days

### 1.2 Write-Ahead Log (WAL)

Location: `pkg/database/embedded/wal/`

Append-only log for crash safety. Ensures no data loss between memory writes and disk flushes.

**Operations logged:**
```go
type OpType int
const (
    OpPut OpType = iota
    OpDelete
)

type Entry struct {
    Op         OpType
    Entity     string  // e.g. "processes", "executors"
    Key        string  // primary key
    Data       []byte  // JSON-encoded record (nil for deletes)
    Timestamp  int64
}
```

**Interface (defined as an interface for future HA/Raft replacement):**
```go
type WAL interface {
    Append(entry Entry) error
    Replay(fn func(Entry) error) error
    Truncate() error
    Close() error
}
```

Initial implementation: `FileWAL` (local file). Future HA: `RaftWAL` (proposes entries through
Raft consensus). The rest of the codebase calls the interface and does not care which
implementation is behind it.

**Format:** Length-prefixed binary entries in a single file. New segment file after truncation.

**Durability modes (configurable):**
- `fsync` every append: no data loss, ~1-5ms per write (production)
- Batched `fsync` (every 100ms): lose up to 100ms on crash, microsecond writes
- No `fsync`: for testing, maximum speed

**Estimated effort:** 2 days

### 1.3 Tests for Phase 1

**The core infrastructure (DiskStore, WAL) must be tested extensively.** All higher layers
depend on their correctness. Bugs here propagate to every entity type and are hard to
diagnose later.

**DiskStore tests:**
- Write/read round-trip for various record types (strings, structs, nested objects)
- Write/read with special characters in keys (slashes, dots, unicode)
- Delete removes file and subsequent read returns not-found
- Delete of non-existent key returns appropriate error
- List returns all keys and only keys that exist
- List on empty store returns empty slice
- Scan iterates all records in deterministic order
- Scan with callback error stops iteration and propagates error
- Concurrent writes to different keys succeed
- Concurrent write and read of same key produce consistent results
- Write to read-only directory returns appropriate error
- Handling of corrupt/truncated JSON files on Read
- Large records (1MB+ JSON) write and read correctly
- Disk full handling (write returns error, existing data intact)

**WAL tests:**
- Append single entry and replay returns it
- Append multiple entries and replay returns all in order
- Replay of empty WAL calls callback zero times
- Truncate clears all entries, subsequent replay returns nothing
- Append after truncate works correctly (new segment)
- Close and reopen preserves all entries
- Replay after crash with partial write (truncated last entry) skips corrupt entry
  and recovers all complete entries
- Concurrent appends from multiple goroutines produce valid log (no interleaving)
- Replay callback error stops replay and propagates error
- Large entries (1MB+) write and replay correctly
- High volume: 100k entries append and replay correctly
- Truncate while concurrent append does not corrupt
- Fsync modes: verify data survives simulated crash (kill process, reopen, replay)

**Integration tests (DiskStore + WAL together):**
- Write through WAL, flush to DiskStore, truncate WAL, verify DiskStore has data
- Simulate crash: WAL has entries not yet flushed to DiskStore. Replay WAL, verify
  all records recovered.
- Mixed puts and deletes through WAL, flush, verify final state on disk

**Estimated effort:** 2-3 days

---

## Phase 2: In-Memory Cache Layer with Indexes

### 2.1 Generic Cache Store

Location: `pkg/database/embedded/store/`

The core abstraction. Wraps a `map` for primary key access with dirty tracking and
integration with DiskStore and WAL.

```go
type Store[K comparable, V any] struct {
    mu       sync.RWMutex
    records  map[K]*V           // primary key -> record
    dirty    map[K]bool         // needs flush to disk
    deleted  map[K]bool         // tombstones pending flush
    disk     *diskstore.DiskStore[V]
    wal      *wal.WAL
    keyFunc  func(V) K          // extract primary key from record
}

func (s *Store[K, V]) Put(key K, value *V) error
func (s *Store[K, V]) Get(key K) (*V, bool)
func (s *Store[K, V]) Delete(key K) error
func (s *Store[K, V]) All() []*V
func (s *Store[K, V]) Filter(fn func(*V) bool) []*V
func (s *Store[K, V]) Count(fn func(*V) bool) int
func (s *Store[K, V]) FlushDirty() error
func (s *Store[K, V]) LoadAll() error             // startup: load all from disk into memory
```

Read path: always from memory (all records loaded at startup).
Write path: WAL append -> memory mutation -> mark dirty.
Background flusher calls `FlushDirty()` periodically.

**Estimated effort:** 2-3 days

### 2.2 Secondary Indexes

Location: `pkg/database/embedded/index/`

Two index types needed:

**MapIndex** - for equality lookups (most entities):
```go
// Maps a secondary key to a set of primary keys
type MapIndex[K comparable, SK comparable] struct {
    index map[SK]map[K]bool
}

func (i *MapIndex[K, SK]) Add(key K, secondaryKey SK)
func (i *MapIndex[K, SK]) Remove(key K, secondaryKey SK)
func (i *MapIndex[K, SK]) Lookup(secondaryKey SK) []K
```

Example uses:
- Processes by colony name: `MapIndex[string, string]`
- Processes by state: `MapIndex[string, int]`
- Attributes by target ID: `MapIndex[string, string]`
- Users by colony name: `MapIndex[string, string]`

**OrderedIndex** - for sorted access + range queries (processes, logs):

Use `github.com/google/btree` (BTreeG generic variant).

```go
type OrderedIndex[K comparable] struct {
    tree *btree.BTreeG[IndexEntry[K]]
}

type IndexEntry[K comparable] struct {
    SortKey   int64  // e.g. priorityTime, timestamp
    PrimaryKey K
}

func (i *OrderedIndex[K]) Add(entry IndexEntry[K])
func (i *OrderedIndex[K]) Remove(entry IndexEntry[K])
func (i *OrderedIndex[K]) AscendRange(min, max int64, fn func(IndexEntry[K]) bool)
func (i *OrderedIndex[K]) AscendFirst(n int, fn func(IndexEntry[K]) bool)
func (i *OrderedIndex[K]) DescendFirst(n int, fn func(IndexEntry[K]) bool)
```

**CompoundIndex** - for multi-field lookups:
```go
// Nesting: e.g. colony -> state -> OrderedIndex (by priorityTime)
type CompoundIndex[K comparable] struct {
    index map[string]map[int]*OrderedIndex[K]
    //     colony     state    sorted by priorityTime
}
```

This covers the critical scheduler path:
`WHERE colony=$1 AND state=WAITING AND ... ORDER BY PRIORITYTIME ASC LIMIT 1`

**Estimated effort:** 3-4 days

### 2.3 Background Flusher and Eviction

Location: `pkg/database/embedded/flusher/`

```go
type Flusher struct {
    stores   []Flushable
    interval time.Duration
    stopCh   chan struct{}
}

func (f *Flusher) Start()                         // background goroutine
func (f *Flusher) Stop()                          // graceful shutdown, final flush
```

Eviction policy (optional, for memory-constrained deployments):
- Track `lastAccess` time per record
- Evict records not accessed within a configurable window
- Never evict dirty records (must flush first)
- On cache miss after eviction, reload from disk

For initial implementation: **skip eviction**, keep all records in memory. Add eviction
later if memory becomes a concern. The disk store provides durability regardless.

**Estimated effort:** 1 day

### 2.4 Tests for Phase 2

**The cache layer and indexes are the second critical foundation. They must be tested
extensively before building entity implementations on top.**

**Store tests:**
- Put/Get round-trip returns correct record
- Put overwrites existing record with same key
- Get non-existent key returns not-found
- Delete removes record, subsequent Get returns not-found
- Delete non-existent key is a no-op (no error)
- All() returns all records
- Filter() returns only matching records
- Filter() with no matches returns empty slice
- Count() with predicate returns correct count
- Put marks record as dirty, FlushDirty writes to disk
- FlushDirty clears dirty set
- Records not modified are not flushed (dirty tracking is accurate)
- Delete marks tombstone, FlushDirty removes from disk
- LoadAll() populates memory from disk
- LoadAll() on empty disk results in empty store
- Put + FlushDirty + clear memory + LoadAll() recovers all records
- WAL integration: Put appends to WAL before memory mutation
- WAL integration: crash after WAL append but before flush, replay recovers record
- Concurrent Put from 100 goroutines, all records present
- Concurrent Put and Get, no panics or data races (run with -race)
- Concurrent Put and Delete of overlapping keys, final state consistent
- Concurrent Filter during Put operations returns consistent snapshots

**MapIndex tests:**
- Add and Lookup returns correct primary keys
- Add same primary key to multiple secondary keys
- Remove primary key from secondary key
- Remove non-existent key is a no-op
- Lookup non-existent secondary key returns empty slice
- Lookup after all primary keys removed returns empty slice
- Add duplicate (same primary + secondary) is idempotent
- Large index: 10k entries, lookups return correct results

**OrderedIndex tests:**
- Add entries and AscendFirst returns in ascending order
- Add entries and DescendFirst returns in descending order
- AscendRange returns only entries within range (inclusive bounds)
- AscendRange with empty range returns nothing
- AscendFirst(n) returns exactly n entries (or all if fewer)
- DescendFirst(n) returns exactly n entries
- Remove entry, no longer appears in traversal
- Remove non-existent entry is a no-op
- Entries with same SortKey: stable ordering by PrimaryKey
- Large index: 100k entries, AscendFirst(10) returns correct top 10
- Add/Remove interleaved, final traversal is correct

**CompoundIndex tests:**
- Nested lookup: colony -> state -> ordered results
- Query for non-existent colony returns empty
- Query for non-existent state within colony returns empty
- Add entries across multiple colonies and states, queries are isolated
- Remove entry from compound index, verify removal at all levels
- Combined with Store: index stays consistent after Put/Delete sequences

**Flusher tests:**
- Start flusher, put dirty records, verify they appear on disk within interval
- Stop flusher triggers final flush
- Stop flusher with no dirty records completes cleanly
- Flusher does not flush non-dirty records
- Flusher handles concurrent puts during flush cycle

**Full integration tests (Store + Indexes + WAL + DiskStore + Flusher):**
- Complete lifecycle: Initialize -> Put records -> indexes updated -> flusher writes
  to disk -> Stop -> Reinitialize from disk -> indexes rebuilt -> queries return
  same results
- Crash simulation: Put records -> kill before flush -> restart -> WAL replay ->
  indexes rebuilt -> all records recovered
- Retention: Put records with timestamps -> apply retention -> old records gone from
  memory, disk, and indexes

**Estimated effort:** 3-4 days

---

## Phase 3: Implement Database Interface - Simple Entities

Location: `pkg/database/embedded/`

Main struct:
```go
type EmbeddedDatabase struct {
    dataDir       string
    wal           *wal.WAL
    flusher       *flusher.Flusher

    colonies      *store.Store[string, core.Colony]
    users         *store.Store[string, core.User]
    executors     *store.Store[string, core.Executor]
    functions     *store.Store[string, core.Function]
    generators    *store.Store[string, core.Generator]
    generatorArgs *store.Store[string, core.GeneratorArg]
    crons         *store.Store[string, core.Cron]
    snapshots     *store.Store[string, core.Snapshot]
    locations     *store.Store[string, core.Location]
    blueprintDefs *store.Store[string, core.BlueprintDefinition]
    blueprints    *store.Store[string, core.Blueprint]
    bpHistory     *store.Store[string, core.BlueprintHistory]
    server        *store.Store[string, string]

    // Complex entities (Phase 4)
    processes     *store.Store[string, core.Process]
    attributes    *store.Store[string, core.Attribute]
    processGraphs *store.Store[string, core.ProcessGraph]
    logs          *store.Store[string, core.Log]
    files         *store.Store[string, core.File]
}
```

### 3.1 DatabaseCore (4 methods)
- `Initialize()`: create data directories, open WAL, load all stores from disk, start flusher
- `Drop()`: stop flusher, clear all in-memory stores, delete data directory
- `Close()`: stop flusher (final flush), close WAL
- `ApplyRetentionPolicy(retentionPeriod int64)`: see Retention section below

### 3.2 ColonyDatabase (7 methods)
Simple map operations. Secondary index: name -> colony.

### 3.3 UserDatabase (7 methods)
Compound key stored as "colonyName:userName". Secondary index: colonyName -> []users.

### 3.4 ExecutorDatabase (16 methods)
Secondary indexes: executorID, (colonyName, name), blueprintID.

### 3.5 FunctionDatabase (11 methods)
Secondary indexes: (colonyName, executorName), (colonyName, executorName, funcName).

### 3.6 GeneratorDatabase (15 methods including GeneratorArgs)
Secondary indexes: colonyName, (colonyName, name), generatorID (for args).

### 3.7 CronDatabase (8 methods)
Secondary indexes: (colonyName, name). UNIQUE constraint on (colonyName, name).

### 3.8 SnapshotDatabase (7 methods)
Secondary indexes: colonyName, (colonyName, name).

### 3.9 LocationDatabase (7 methods)
Secondary indexes: colonyName, (colonyName, name).

### 3.10 BlueprintDatabase (29 methods)
Most complex of the "simple" entities due to three sub-types (definitions, blueprints, history).
Secondary indexes: (colonyName, name), kind, apiGroup, (blueprintID, generation).
JSONB location filtering implemented as: unmarshal DATA field, extract metadata.locationname,
case-insensitive comparison.

### 3.11 SecurityDatabase (5 methods)
Trivial. Single server record, update colony/user/executor IDs in their respective stores.

### 3.12 Tests for Phase 3

Run existing PostgreSQL tests against the new implementation. The test files in
`pkg/database/postgresql/*_test.go` contain the test logic. Strategy:
- Create a shared test suite in `pkg/database/testing/` that accepts a `database.Database`
- Both PostgreSQL and Embedded implementations run the same tests
- Alternatively, duplicate the test setup function to use Embedded instead

**Estimated effort for Phase 3:** 1.5-2 weeks

---

## Phase 4: Implement Database Interface - Complex Entities

### 4.1 ProcessDatabase (~45 methods)

The most complex entity. Requires:

**Indexes needed:**
- Primary: processID -> Process
- By colony + state + priorityTime: CompoundIndex for scheduler queries
- By colony + state + submissionTime: for time-ordered history queries
- By colony + state + startTime/endTime: for running/completed queries
- By assignedExecutorID: for executor's processes
- By processGraphID: for workflow membership
- By targetExecutorNames (array): for named executor matching

**Critical methods:**

`SelectAndAssign()` - atomic select + update:
```go
func (db *EmbeddedDatabase) SelectAndAssign(...) (*core.Process, error) {
    db.processes.mu.Lock()
    defer db.processes.mu.Unlock()

    // Query compound index: colony + WAITING state, ordered by priorityTime
    // Filter by: executor type, resource constraints, location, executor name/wildcard
    // Take first match
    // Atomically update: state=RUNNING, assignedExecutorID, startTime, execDeadline
    // Return updated process
}
```

The `sync.RWMutex` on the store provides the equivalent of `FOR UPDATE SKIP LOCKED`
since all data is in memory. No need for row-level locks.

`FindCandidates()` / `FindCandidatesByName()`:
- Use compound index to get candidates ordered by priorityTime
- Filter by resource constraints (CPU, memory, storage, nodes, processes, processesPerNode)
- Filter by location (case-insensitive, NULL/empty means any)
- FindCandidatesByName additionally checks `executorName in targetExecutorNames`

**Process state queries (FindWaiting/Running/Successful/Failed/Cancelled):**
- Use compound index: colony + state
- Optional filters: executorType, label, initiatorName
- ORDER BY submissionTime/startTime/endTime DESC, LIMIT count

**Count methods:** Maintain counters per (colony, state) updated on every state transition.
Avoids full scans.

**Estimated effort:** 5-6 days

### 4.2 AttributeDatabase (18 methods)

**Indexes needed:**
- Primary: attributeID -> Attribute
- By targetID: for getting all attributes of a process
- By (targetID, key, attributeType): for specific attribute lookup
- By colonyName: for colony-wide operations
- By processGraphID: for workflow attribute cleanup

**Batch insert:** `AddAttributes()` - iterate and add each, all under one lock.

**Estimated effort:** 2-3 days

### 4.3 ProcessGraphDatabase (25 methods)

Similar to ProcessDatabase but simpler (fewer query patterns).

**Indexes needed:**
- Primary: processGraphID -> ProcessGraph
- By (colonyName, state): for state-based queries
- Ordered by submissionTime within each (colony, state) group

**Count methods:** Maintain counters per (colony, state).

**Estimated effort:** 2-3 days

### 4.4 LogDatabase (10 methods)

Disk-backed with memory cache for recent entries.

**Design:**
- Logs are always written to disk immediately (append to per-process file)
- Recent logs cached in memory with bounded size per process/executor
- On read, check memory first, fall back to disk scan

**Disk layout for logs:**
```
<datadir>/logs/
  by-process/<processID>.log     # newline-delimited JSON, sorted by TS
  by-executor/<executorName>.log # newline-delimited JSON, sorted by TS
  by-colony/<colonyName>.log     # newline-delimited JSON, for search/count/delete
```

Each AddLog writes to all three files (triple-write, but logs are append-only so this is fast).

**Query implementations:**
- `GetLogsByProcessID`: read from by-process file, take first N
- `GetLogsByProcessIDSince`: scan by-process file, skip until TS > since, take N
- `GetLogsByProcessIDLatest`: read last N entries from by-process file
- `SearchLogs`: scan by-colony file, filter by time window + string contains
- `CountLogs`: count lines in by-colony file (or maintain counter)
- `RemoveLogsByColonyName`: delete all files associated with that colony

**Indexes needed (in-memory):**
- colonyName -> set of processIDs (for delete/search)
- colonyName -> set of executorNames (for delete)
- colonyName -> count (maintained on add/delete)

**Estimated effort:** 3-4 days

### 4.5 FileDatabase (12 methods)

**Indexes needed:**
- Primary: fileID -> File
- By (colonyName, label, name): for file lookups
- By (colonyName, label): for label listing
- Sequence counter per (colonyName, label, name) for `nextval` equivalent
- LIKE prefix matching for `GetFileLabelsByName`

**Estimated effort:** 2-3 days

### 4.6 Tests for Phase 4

Same strategy as Phase 3 - run existing test suites against Embedded implementation.
Focus extra testing on:
- `SelectAndAssign` under concurrent access (multiple goroutines)
- Process state transitions and counter consistency
- Log disk persistence and recovery
- Retention policy across all entity types

**Estimated effort:** 3-4 days

---

## Phase 5: Integration and Testing

### 5.1 Factory Function

In `pkg/database/database.go`, add:
```go
func CreateEmbeddedDatabase(dataDir string) (Database, error)
```

### 5.2 Test Infrastructure Integration

Modify or add test setup parallel to `pkg/database/postgresql/test_utils.go`:
```go
// pkg/database/embedded/test_utils.go
func PrepareTests() *EmbeddedDatabase {
    dir, _ := os.MkdirTemp("", "colonies-test-*")
    db := NewEmbeddedDatabase(dir)
    db.Initialize()
    return db
}
```

### 5.3 Server Integration

The server creates its database in startup code. Add a configuration option to select
between `postgresql` and `embedded` backends:
```go
switch config.DBType {
case "postgresql":
    db = postgresql.CreatePQDatabase(...)
case "embedded":
    db = embedded.CreateEmbeddedDatabase(config.DataDir)
}
```

### 5.4 Run Full Test Suite

Run all existing tests (`pkg/database/postgresql/*_test.go` patterns) against the embedded
implementation to verify behavioral parity.

Run server-level tests (`pkg/server/*_test.go`) with embedded backend.

### 5.5 CLI Flag

Add `--db-type embedded --data-dir /path/to/data` flags to the CLI for selecting the backend.

**Estimated effort for Phase 5:** 3-4 days

---

## Dependency: External Libraries

- `github.com/google/btree` - BTreeG for ordered indexes (well-maintained, zero dependencies)
- No other new dependencies

---

## File Structure

```
pkg/database/embedded/
    database.go          # EmbeddedDatabase struct, Initialize, Drop, Close
    colonies.go          # ColonyDatabase implementation
    users.go             # UserDatabase implementation
    executors.go         # ExecutorDatabase implementation
    functions.go         # FunctionDatabase implementation
    processes.go         # ProcessDatabase implementation (largest)
    attributes.go        # AttributeDatabase implementation
    processgraphs.go     # ProcessGraphDatabase implementation
    generators.go        # GeneratorDatabase + GeneratorArgs implementation
    crons.go             # CronDatabase implementation
    logs.go              # LogDatabase implementation (disk-backed)
    files.go             # FileDatabase implementation
    snapshots.go         # SnapshotDatabase implementation
    blueprints.go        # BlueprintDatabase implementation
    locations.go         # LocationDatabase implementation
    security.go          # SecurityDatabase implementation
    retention.go         # ApplyRetentionPolicy implementation
    test_utils.go        # Test setup helpers

    store/
        store.go         # Generic Store[K, V] with dirty tracking
        store_test.go

    index/
        mapindex.go      # MapIndex for equality lookups
        orderedindex.go  # OrderedIndex using btree for sorted access
        compoundindex.go # CompoundIndex for multi-field lookups
        index_test.go

    diskstore/
        diskstore.go     # Generic file-per-record disk storage
        diskstore_test.go

    wal/
        wal.go           # WAL interface definition
        filewal.go       # FileWAL implementation (local file)
        wal_test.go

    flusher/
        flusher.go       # Background flush goroutine
        flusher_test.go
```

---

## Estimated Total Effort

| Phase | Description | Effort |
|-------|-------------|--------|
| 1 | DiskStore + WAL + extensive tests | 5-7 days |
| 2 | Cache Store + Indexes + Flusher + extensive tests | 8-11 days |
| 3 | Simple entities (11 interfaces, ~100 methods) | 8-10 days |
| 4 | Complex entities (5 interfaces, ~100 methods) | 12-16 days |
| 5 | Integration, testing, CLI | 3-4 days |
| **Total** | | **6-8 weeks** |

**Note:** Phases 1 and 2 have proportionally more testing time than implementation time.
This is intentional. The core infrastructure must be rock-solid before building on it.
A bug in the Store or WAL layer can manifest as subtle data corruption in any of the
200+ entity methods, making it extremely hard to diagnose. Invest the testing time upfront.

---

## Build Order and Milestones

1. **Milestone: Storage works** (end of Phase 1)
   - DiskStore and WAL pass unit tests
   - Records survive write/read/delete round-trips
   - WAL replay recovers unflushed writes

2. **Milestone: Cache works** (end of Phase 2)
   - Generic Store with indexes passes concurrent tests
   - Background flusher persists dirty records
   - Startup loads from disk and rebuilds indexes

3. **Milestone: Simple entities work** (end of Phase 3)
   - Colony, User, Executor, Function, Generator, Cron, Snapshot, Location,
     Blueprint, Security interfaces pass existing test suites

4. **Milestone: All entities work** (end of Phase 4)
   - Process, Attribute, ProcessGraph, Log, File interfaces pass existing test suites
   - SelectAndAssign works correctly under concurrent access

5. **Milestone: Production-ready** (end of Phase 5)
   - Server boots with `--db-type embedded`
   - Full test suite passes
   - Data persists across server restarts

---

## Risks and Mitigations

**Risk: Behavioral differences from PostgreSQL**
- Mitigation: Run the same test suite against both implementations. Add edge case tests
  for NULL handling, empty arrays, case sensitivity, and ordering ties.

**Risk: Concurrent access bugs in SelectAndAssign**
- Mitigation: Use a single mutex for the processes store. This serializes assignment but
  is correct. PostgreSQL does the same with row-level locks. Optimize later if throughput
  becomes an issue.

**Risk: Memory growth from logs**
- Mitigation: Logs are disk-backed (Phase 4.4). Only recent entries cached in memory.
  Retention policy cleans up old log files.

**Risk: Startup time with large datasets**
- Mitigation: For initial implementation, load everything at startup. If this becomes slow,
  add lazy loading (load on first access per entity type) as an optimization.

**Risk: Disk I/O performance on flush**
- Mitigation: Background flusher batches writes. Only dirty records are flushed. JSON
  encoding is simple and fast for the record sizes in this system.

---

## Retention Policy

The `ApplyRetentionPolicy(retentionPeriod int64)` method is called periodically by the
server to clean up old data. The retention period is in seconds. It must delete records
from memory, disk, and all indexes.

**What gets deleted (matching existing PostgreSQL behavior):**
- Attributes where `Added < now - retentionPeriod` AND `State == SUCCESS`
- Logs where `Added < now - retentionPeriod`
- Processes where `SubmissionTime < now - retentionPeriod` AND `State == SUCCESS`
- ProcessGraphs where `SubmissionTime < now - retentionPeriod` AND `State == SUCCESS`

**Implementation:**
Each entity store needs an efficient way to find records older than a threshold. Two options:

1. **OrderedIndex by timestamp** - already needed for processes and logs. Use
   `AscendRange(0, cutoff)` to find all expired records, then delete them.

2. **Linear scan with filter** - for attributes, which don't have a time-ordered
   index. Filter by `Added < cutoff AND State == SUCCESS`, collect keys, delete.

**Retention for disk-backed logs:**
The log files on disk (by-process, by-executor, by-colony) grow unbounded without retention.
The retention policy must:
- Scan log files and remove entries older than the retention period
- For append-only files, rewrite the file excluding old entries (or maintain a
  "start offset" marker per file)
- Clean up empty files after all entries are removed
- Update the in-memory colony log count

**Retention must update all layers:**
1. Remove from memory (cache)
2. Remove from all secondary indexes
3. Remove from disk (DiskStore files or log files)
4. Update any maintained counters

**Testing retention:**
- Add records with old timestamps, apply retention, verify they are gone from memory
  and disk
- Add records with recent timestamps, apply retention, verify they survive
- Mixed old and new records, verify only old ones removed
- Retention on empty database is a no-op
- Retention for logs: verify log files are cleaned up and count is updated
- Retention does not delete WAITING, RUNNING, or FAILED processes (only SUCCESS)
- Concurrent retention + writes: new records added during retention are not affected

---

## Future: High Availability

The WAL is defined as an interface from the start to enable HA in the future without
changing the rest of the codebase.

**Single-node (current plan):**
```
Store -> FileWAL -> local disk
```

**Future HA:**
```
Store -> RaftWAL -> Raft consensus -> all nodes apply to memory
```

The `RaftWAL` would implement the same `WAL` interface but propose entries through
Raft consensus (e.g. `hashicorp/raft`) instead of writing to a local file. The Store,
indexes, flusher, and all 200+ database methods stay exactly the same.

The components map directly to what `hashicorp/raft` needs:
- `WAL` entries = Raft LogStore
- `Store.Apply()` = Raft FSM (finite state machine)
- `DiskStore` = Raft SnapshotStore (point-in-time dump)

No changes to the current plan are needed to support this path. The WAL interface
abstraction is the only prerequisite, and it is included in Phase 1.
