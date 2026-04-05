# Embedded Database Architecture

## Overview and Motivation

The embedded database provides an alternative to PostgreSQL for ColonyOS deployments
that do not require a separate database server. Both implementations satisfy the same
`database.Database` interface (defined in `pkg/database/database.go`), so the rest of
the codebase is agnostic to which backend is in use.

**Use cases:**

- **Single-server deployments** where running PostgreSQL adds unwanted operational
  complexity.
- **Edge devices** with constrained resources where a full RDBMS is impractical.
- **Development and testing** where zero-dependency startup is preferred.

**Design goals:**

- Zero external dependencies -- no database server, no C libraries.
- Fast reads -- all data served from memory, no disk I/O on the read path.
- Crash recovery -- a write-ahead log guarantees that committed writes survive
  process crashes.
- Interface compatibility -- every method on `database.Database` is implemented,
  allowing transparent substitution for the PostgreSQL backend.


## Storage Architecture

The embedded database uses a three-layer architecture:

```
+-------------------+
|  In-Memory Store  |  <-- all reads served here
+-------------------+
         |
    WAL append (sequential writes)
         |
+-------------------+
|  Write-Ahead Log  |  <-- durability guarantee
+-------------------+
         |
    async flush (background goroutine, every 5 s)
         |
+-------------------+
|    Disk Store     |  <-- one JSON file per record
+-------------------+
```

**Write flow:**

1. The WAL entry is appended (sequential I/O).
2. The in-memory store is updated.
3. The record is marked dirty.
4. The background flusher periodically writes dirty records to the disk store and
   truncates the WAL.

**Read flow:**

All reads are served directly from the in-memory store. The disk store is only
read during startup to populate memory.


## Write-Ahead Log (WAL)

The WAL ensures that every committed write can be recovered after a crash, even
if the background flusher has not yet persisted the data to the disk store.

### Interface

The WAL is defined as an interface (`wal.WAL` in `wal/wal.go`) with four methods:

```go
type WAL interface {
    Append(entry Entry) error
    Replay(fn func(Entry) error) error
    Truncate() error
    Close() error
}
```

The concrete implementation is `FileWAL` (`wal/filewal.go`), backed by a single
append-only file (`wal.log`).

### Binary Format

Each entry is encoded as:

```
[total_len:4][crc32:4][op:1][entity_len:2][entity][key_len:2][key][ts:8][data_len:4][data]
```

- **total_len** (4 bytes, little-endian uint32): length of everything after this
  field (CRC through data).
- **crc32** (4 bytes, IEEE CRC32): covers all bytes after itself (op through
  data). Used to detect corruption.
- **op** (1 byte): operation type -- `OpPut` (1) or `OpDelete` (2).
- **entity_len** + **entity**: variable-length entity name (e.g. `"processes"`).
- **key_len** + **key**: variable-length primary key.
- **ts** (8 bytes): Unix nanosecond timestamp.
- **data_len** + **data**: JSON-encoded record payload (empty for deletes).

### Sync Modes

```go
const (
    SyncAlways SyncMode = iota  // fsync after every append
    SyncNone                     // no fsync, OS decides when to flush
)
```

The default is `SyncNone` for performance. `SyncAlways` provides stronger
durability at the cost of write latency.

### Replay and Truncation

On startup, `Replay` reads entries sequentially from the WAL file and calls a
callback for each valid entry. Corrupt or truncated trailing entries (from a
mid-write crash) are silently skipped -- replay stops at the first unreadable
entry.

After the background flusher has written all dirty records to disk, the WAL is
truncated (replaced with an empty file) so it does not grow unboundedly.

### Buffered I/O

`FileWAL` uses a 64 KB `bufio.Writer` to batch small writes. The buffer is
flushed before replay and before truncation to avoid data loss. A reusable encode
buffer (`encBuf`) avoids allocations on the hot append path.

### References

- C. Mohan, D. Haderle, B. Lindsay, H. Pirahesh, P. Schwarz. "ARIES: A
  Transaction Recovery Method Supporting Fine-Granularity Locking and Partial
  Rollbacks Using Write-Ahead Logging." ACM Transactions on Database Systems,
  1992.
- PostgreSQL documentation: "Reliability and the Write-Ahead Log"
  (https://www.postgresql.org/docs/current/wal.html)
- SQLite documentation: "Write-Ahead Logging"
  (https://www.sqlite.org/wal.html)


## In-Memory Store

The `Store[K, V]` generic type (`store/store.go`) is the central data structure.
Each entity type (processes, executors, colonies, etc.) gets its own store
instance.

### Type Signature

```go
type Store[K comparable, V any] struct {
    mu         sync.RWMutex
    records    map[K]*V
    dirty      map[K]struct{}
    deleted    map[K]struct{}
    disk       *diskstore.DiskStore[V]
    wal        wal.WAL
    entityName string
    keyToStr   func(K) string
    strToKey   func(string) K
}
```

### Dirty Tracking

Each store maintains two sets:

- `dirty` -- keys that have been written since the last flush. The flusher writes
  these to disk.
- `deleted` -- keys that have been deleted since the last flush. The flusher
  removes the corresponding files from disk.

Both sets are cleared after a successful flush.

### Locked and Unlocked Method Variants

Every mutating operation has two variants:

| Locked (acquires `mu`)   | Unlocked (caller holds `mu`) |
|--------------------------|------------------------------|
| `Put(key, value)`        | `PutUnlocked(key, value)`    |
| `Get(key)`               | `GetUnlocked(key)`           |
| `Delete(key)`            | `DeleteUnlocked(key)`        |
| `Filter(fn)`             | `FilterUnlocked(fn)`         |
| `Count(fn)`              | `CountUnlocked(fn)`          |

The unlocked variants exist for atomic multi-step operations. For example,
`SelectAndAssign` (process scheduling) must read and update a process atomically.
The caller acquires `Lock()`, performs multiple unlocked operations, then calls
`Unlock()`.

### Replay Methods

During WAL replay, `ReplayPut` and `ReplayDelete` modify memory without writing
to the WAL (avoiding double-logging). They also require the caller to hold the
write lock.

```go
func (s *Store[K, V]) ReplayPut(key K, value *V)
func (s *Store[K, V]) ReplayDelete(key K)
```

### WAL Lifecycle

Stores are created with `WAL: nil` during initialization. After WAL replay
completes, `SetWAL(w)` is called to enable WAL logging for subsequent writes.
This prevents replayed entries from being re-logged.


## Disk Store

The `DiskStore[V]` (`diskstore/diskstore.go`) persists each record as a separate
JSON file on disk.

### Type Signature

```go
type DiskStore[V any] struct {
    mu      sync.RWMutex
    baseDir string
}
```

### Key Sanitization

Keys may contain characters that are unsafe for filenames. A `strings.NewReplacer`
maps these to safe tokens:

| Character | Replacement    |
|-----------|----------------|
| `/`       | `__SLASH__`    |
| `\`       | `__BSLASH__`   |
| `:`       | `__COLON__`    |
| `*`       | `__STAR__`     |
| `?`       | `__QMARK__`    |
| `"`       | `__QUOTE__`    |
| `<`       | `__LT__`       |
| `>`       | `__GT__`       |
| `\|`      | `__PIPE__`     |

The inverse mapping (`desanitizeKey`) restores original keys when listing files.

### Atomic Writes

Writes use the temp-file-then-rename pattern for atomicity:

1. Create a temporary file in the same directory (`.tmp-*` prefix).
2. Write the JSON payload.
3. `fsync` the temporary file.
4. Rename it to the target path.

If the process crashes at any point before the rename, the old file remains
intact. Temporary files left behind by crashes are filtered out during `Scan` and
`List`.

### Scanning

`Scan` iterates all `.json` files in sorted key order, deserializing each and
calling the provided callback. Used during startup (`LoadAll`) to populate the
in-memory store.


## Background Flusher

The `Flusher` (`flusher/flusher.go`) is a background goroutine that periodically
writes dirty records from all stores to disk.

### Type Signature

```go
type Flusher struct {
    stores   []Flushable
    interval time.Duration
    stopCh   chan struct{}
    doneCh   chan struct{}
    mu       sync.Mutex
}

type Flushable interface {
    FlushDirty() error
}
```

### Behavior

- Default interval: **5 seconds** (configured in `Initialize()`).
- On each tick, the flusher calls `FlushDirty()` on every registered store.
- `FlushDirty()` snapshots the dirty and deleted sets under the store lock, then
  performs disk I/O outside the lock. This minimizes lock contention with
  concurrent readers and writers.
- On `Stop()`, a final flush is performed before the goroutine exits.
- After all stores are flushed, the WAL can be truncated (done during `Close()`).

### FlushDirty Implementation

```go
func (s *Store[K, V]) FlushDirty() error {
    s.mu.Lock()
    // snapshot dirty entries and deleted keys
    // clear dirty/deleted sets
    s.mu.Unlock()

    // write dirty entries to disk (outside lock)
    // delete tombstoned records from disk (outside lock)
}
```

The snapshot-then-release pattern ensures that the store lock is not held during
potentially slow disk I/O.


## Indexing

Indexes accelerate lookups that would otherwise require a full scan of the
in-memory store. They are **not persisted** -- they are rebuilt from data on every
startup.

### MapIndex

`MapIndex[K, SK]` (`index/mapindex.go`) maps a secondary key to a set of primary
keys. Used for equality lookups such as "all executors in colony X."

```go
type MapIndex[K comparable, SK comparable] struct {
    mu    sync.RWMutex
    index map[SK]map[K]struct{}
}
```

Methods: `Add(key, secondaryKey)`, `Remove(key, secondaryKey)`,
`Lookup(secondaryKey) []K`, `Count(secondaryKey) int`, `Clear()`.

### CompoundIndex

`CompoundIndex` (`index/compoundindex.go`) provides two-level nesting:
`group -> subgroup -> OrderedIndex`. Designed for queries like
`WHERE colony = $1 AND state = $2 ORDER BY priorityTime`.

```go
type CompoundIndex struct {
    mu    sync.RWMutex
    index map[string]map[int]*OrderedIndex[string]
}
```

The group key is a string (e.g. colony name), the subgroup key is an int
(e.g. process state), and each leaf is an `OrderedIndex` sorted by a numeric
sort key (e.g. priority time or submission time).

Methods: `Add(group, subgroup, entry)`, `Remove(group, subgroup, entry)`,
`AscendFirst(group, subgroup, n, fn)`, `DescendFirst(group, subgroup, n, fn)`,
`Count(group, subgroup)`, `CountAll(group)`, `CountBySubgroup(subgroup)`.

### OrderedIndex

`OrderedIndex[K]` (`index/orderedindex.go`) wraps a B-tree
(`github.com/google/btree`) for sorted access.

```go
type IndexEntry[K comparable] struct {
    SortKey    int64
    PrimaryKey K
}

type OrderedIndex[K comparable] struct {
    mu   sync.RWMutex
    tree *btree.BTreeG[IndexEntry[K]]
    less func(a, b IndexEntry[K]) bool
}
```

Entries are sorted by `SortKey` ascending, with ties broken by primary key.
Supports ascending/descending iteration, range queries (`AscendRange`), and
threshold queries (`AscendGreaterThan`).

### Index Rebuild

`rebuildIndexes()` iterates every record in every store and populates all index
structures. This runs once during `Initialize()`, after both WAL replay and disk
load are complete.


## Concurrency Model

The embedded database uses a two-level locking scheme.

### Database-Level Lock (`db.mu`)

```go
type EmbeddedDatabase struct {
    mu sync.RWMutex  // database-level transaction lock for write atomicity
    // ...
}
```

Every write method acquires `db.mu` as a write lock to ensure atomicity across
multiple store and index updates. For example, `AddProcess` must update the
process store, the attribute store, and several indexes as a single atomic
operation.

Read-only methods that touch multiple stores (e.g. `enrichProcess`, which reads
from both the process store and the attribute store) acquire `db.mu` as a read
lock for consistency.

### Store-Level Lock (`store.mu`)

Each `Store[K, V]` has its own `sync.RWMutex` that protects the `records`,
`dirty`, and `deleted` maps. This lock is also used by the background flusher:
`FlushDirty()` briefly holds the store lock to snapshot dirty entries, then
releases it before performing disk I/O.

### Lock Ordering

The invariant is: **`db.mu` must be acquired before any `store.mu`**. This
prevents deadlocks between database-level operations and the background flusher.

### Avoiding Deadlocks from Nested Calls

When a public method (e.g. `AddProcess`) already holds `db.mu`, it must not call
another public method that also acquires `db.mu`. Instead, it calls an internal
(unexported) helper. For example:

- `AddProcess` holds `db.mu`, then calls `addAttributes` (not `AddAttribute`)
  to add process attributes without re-acquiring the database lock.
- `DeleteColony` holds `db.mu`, then calls `removeUsersByColonyName` and similar
  internal helpers for cascade deletes.

### Retention Policy and Batched Deletes

`ApplyRetentionPolicy` avoids holding locks for extended periods by processing
records in batches of 100:

1. Acquire `db.mu.RLock()`, scan for up to 100 records matching the retention
   criteria, collect their IDs.
2. Release the read lock.
3. Acquire `db.mu.Lock()`, delete the batch, release the write lock.
4. Repeat until no more records match.

This allows concurrent readers to proceed between batches.


## Startup and Recovery Sequence

`Initialize()` orchestrates the full startup sequence:

```
1. Create data directory          os.MkdirAll(db.dataDir)
2. Open WAL                       wal.NewFileWAL(walPath, wal.SyncNone)
3. Create all stores (WAL=nil)    db.createStores(nil)
4. Create all indexes             db.createIndexes()
5. Replay WAL into stores         db.replayWAL(w)
6. Load from disk                 db.loadAll()
7. Set WAL on all stores          db.setWALOnStores(w)
8. Rebuild indexes                db.rebuildIndexes()
9. Start background flusher       db.flusher.Start()
```

Key details:

- **Step 3**: Stores are created without a WAL reference. This ensures that
  replayed entries (step 5) are not re-appended to the WAL.
- **Step 5**: `replayWAL` locks all stores, then calls `ReplayPut` /
  `ReplayDelete` which modify memory directly without WAL writes.
- **Step 6**: `LoadAll` reads every JSON file from disk into memory. Records
  already present from WAL replay are **not overwritten** -- the WAL version is
  newer.
- **Step 7**: After replay and load are complete, the WAL is connected to all
  stores so that future writes are logged.
- **Step 8**: Indexes are rebuilt from the final in-memory state, which
  represents the union of disk data and WAL replay.

### Shutdown

`Close()` stops the flusher (which performs a final flush), truncates the WAL
(since all data is now on disk), and closes the WAL file.

`Drop()` stops the flusher, closes the WAL, and removes the entire data
directory.


## Comparison with PostgreSQL Implementation

Both backends implement `database.Database`:

```go
type Database interface {
    DatabaseCore
    UserDatabase
    ColonyDatabase
    ExecutorDatabase
    FunctionDatabase
    ProcessDatabase
    AttributeDatabase
    ProcessGraphDatabase
    GeneratorDatabase
    CronDatabase
    LogDatabase
    FileDatabase
    SnapshotDatabase
    BlueprintDatabase
    SecurityDatabase
    LocationDatabase
    MetricDatabase
}
```

| Aspect                | Embedded                        | PostgreSQL                         |
|-----------------------|---------------------------------|------------------------------------|
| Dependencies          | None                            | PostgreSQL + TimescaleDB           |
| Read path             | In-memory map lookup            | SQL query over network             |
| Write durability      | WAL + async disk flush          | PostgreSQL WAL + fsync             |
| Concurrency           | Go mutexes, single-process      | MVCC, multi-process                |
| Indexing              | In-memory B-tree and hash maps  | B-tree on disk (PostgreSQL)        |
| Query capability      | Programmatic filters            | Full SQL                           |
| Scalability           | Single node, memory-bound       | Horizontal read replicas, sharding |
| Backup                | Copy data directory             | pg_dump, pg_basebackup, PITR      |
| Time-series           | Not supported                   | TimescaleDB hypertables            |


## Limitations and Future Work

- **No range queries on arbitrary fields.** Only the `OrderedIndex` (used for
  process priority and graph submission time) supports ordered iteration. Other
  lookups require `MapIndex` equality or full-scan `Filter`.
- **No joins.** Cross-entity queries (e.g. enriching a process with its
  attributes) are performed as multiple map lookups in application code.
- **Write serialization.** All writes acquire `db.mu` exclusively. Under heavy
  write load, this becomes a bottleneck compared to PostgreSQL's row-level MVCC.
- **WAL growth between flushes.** The WAL is only truncated on shutdown (via
  `Close`). Between flusher cycles, the WAL file grows proportionally to write
  volume.
- **No built-in backup.** There is no snapshot or incremental backup mechanism.
  The data directory can be copied while the server is stopped, but online backup
  requires stopping writes.
- **Memory-bound.** All data must fit in memory. There is no eviction or
  memory-mapped fallback.


## References

- C. Mohan et al. "ARIES: A Transaction Recovery Method Supporting
  Fine-Granularity Locking and Partial Rollbacks Using Write-Ahead Logging."
  *ACM TODS*, 17(1), 1992.
- Martin Kleppmann. *Designing Data-Intensive Applications*. O'Reilly, 2017.
  Chapters 3 (Storage and Retrieval) and 7 (Transactions).
- Alex Petrov. *Database Internals*. O'Reilly, 2019. Chapters on B-trees,
  LSM-trees, and write-ahead logging.
- PostgreSQL documentation: "Reliability and the Write-Ahead Log."
  https://www.postgresql.org/docs/current/wal.html
- SQLite documentation: "Write-Ahead Logging."
  https://www.sqlite.org/wal.html
- BoltDB (etcd-io/bbolt): Single-file B+ tree database for Go.
  https://github.com/etcd-io/bbolt
- BadgerDB (dgraph-io/badger): LSM-tree key-value store for Go.
  https://github.com/dgraph-io/badger
