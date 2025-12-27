# Performance Analysis: Exclusive vs Distributed Assign

## Summary

This analysis compares two assign modes in ColonyOS:
- **Exclusive Assign**: All assign requests forwarded to cluster leader
- **Distributed Assign**: Each replica handles assigns directly using `SELECT FOR UPDATE SKIP LOCKED`

**Key Finding**: Distributed assign is **10-13x faster** than exclusive assign.

## Test Configuration

- **Total assigns per test**: 10,000
- **Replica configurations tested**: 1, 3, 5, 7, 9
- **Database**: Single PostgreSQL instance with TimescaleDB
- **Connection pool**: 100 max open, 100 max idle connections

## Results

### Exclusive Assign

| Replicas | Successful | Failed | Avg Latency (ms) | P95 (ms) | P99 (ms) |
|----------|------------|--------|------------------|----------|----------|
| 1        | 10,000     | 0      | 340.757          | 444.380  | 507.223  |
| 3        | 10,000     | 0      | 351.360          | 448.294  | 477.196  |
| 5        | 10,000     | 0      | 345.926          | 432.772  | 474.493  |
| 7        | 10,000     | 0      | 347.947          | 454.058  | 487.043  |
| 9        | 10,000     | 0      | 354.448          | 447.201  | 480.380  |

### Distributed Assign

| Replicas | Successful | Failed | Avg Latency (ms) | P95 (ms) | P99 (ms) |
|----------|------------|--------|------------------|----------|----------|
| 1        | 10,000     | 0      | 25.420           | 34.532   | 45.050   |
| 3        | 10,000     | 0      | 26.193           | 36.994   | 47.217   |
| 5        | 10,000     | 0      | 31.065           | 44.965   | 61.711   |
| 7        | 10,000     | 0      | 32.281           | 45.896   | 59.662   |
| 9        | 10,000     | 0      | 35.112           | 49.836   | 64.777   |

## Analysis

### Why Distributed Assign is Faster

#### 1. No Leader Forwarding

**Exclusive Assign**:
```
Client -> Any Replica -> Forward to Leader -> Database -> Leader -> Replica -> Client
```

**Distributed Assign**:
```
Client -> Any Replica -> Database -> Replica -> Client
```

Savings: ~10-15ms per request

#### 2. No Single-Threaded Command Queue

The ColoniesController uses a command queue pattern:

```go
func (controller *ColoniesController) CmdQueueWorker() {
    for {
        select {
        case cmd := <-controller.cmdQueue:  // Single goroutine
            cmd.handler(cmd)
        }
    }
}
```

With exclusive assign, all 10,000 requests serialize through this single-threaded queue on the leader. Each request waits for all previous ones.

**Distributed assign bypasses this entirely** - it makes direct database calls:

```go
func (controller *ColoniesController) DistributedAssign(...) {
    controller.executorDB.MarkAlive(executor)           // Direct
    controller.AreColonyAssignmentsPaused(colonyName)   // Direct
    controller.processDB.SelectAndAssign(...)           // Direct
}
```

Savings: ~300ms per request (queue wait time)

#### 3. Concurrent Database Access

Distributed assign uses PostgreSQL's `SELECT FOR UPDATE SKIP LOCKED`:

```sql
SELECT PROCESS_ID FROM PROCESSES
WHERE STATE = 0 AND IS_ASSIGNED = FALSE ...
ORDER BY PRIORITYTIME ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
```

This allows multiple replicas to concurrently select different processes without blocking each other. If a row is locked, it's skipped and the next available row is selected.

### Why Latency Increases with More Replicas (Distributed Mode)

With distributed assign, latency increases slightly with more replicas (25ms -> 35ms for 1 -> 9 replicas). This is expected:

1. **Same total work, more concurrency**: All replicas hit the same database
2. **Database contention**: More concurrent transactions scanning the same table
3. **Lock overhead**: More `SKIP LOCKED` operations as replicas compete

However, this is **sublinear degradation**: 9x replicas only causes 1.4x latency increase.

### Why Latency is Constant in Exclusive Mode

With exclusive assign, latency stays constant (~345ms) regardless of replica count because:
- All requests funnel through the single leader
- Additional replicas just proxy to the same bottleneck
- The queue wait time dominates

## Optimizations Applied

### 1. Connection Pooling

**Problem**: Without connection pooling, high concurrency caused port exhaustion:
```
dial tcp: connect: cannot assign requested address
```

**Solution**: Added configurable connection pool in `database.go`:
```go
db.postgresql.SetMaxOpenConns(100)
db.postgresql.SetMaxIdleConns(100)
db.postgresql.SetConnMaxLifetime(5 * time.Minute)
db.postgresql.SetConnMaxIdleTime(1 * time.Minute)
```

**Result**: 0% failure rate (was 73% before)

### 2. Database Indexes

Added optimized indexes for the assign query:

```sql
-- GIN index for array membership search
CREATE INDEX PROCESSES_ASSIGN_GIN ON PROCESSES
USING GIN (TARGET_EXECUTOR_NAMES)
WHERE STATE = 0 AND IS_ASSIGNED = FALSE;

-- Partial B-tree index for assign hot path
CREATE INDEX PROCESSES_ASSIGN_BTREE ON PROCESSES
(TARGET_COLONY_NAME, EXECUTOR_TYPE, PRIORITYTIME)
WHERE STATE = 0 AND IS_ASSIGNED = FALSE AND WAIT_FOR_PARENTS = FALSE;
```

### 3. Eliminated Duplicate Executor Lookup

**Before**: Executor looked up twice per assign (handler + controller)
**After**: Executor passed from handler to controller

Reduced database operations from 5 to 4 per assign.

## Conclusions

1. **Distributed assign is the clear winner** for performance: 10-13x faster than exclusive assign

2. **Single replica is optimal for latency** when using distributed assign with a single database

3. **Multiple replicas provide**:
   - Fault tolerance (survives node failures)
   - Higher concurrent executor capacity
   - Geographic distribution capability

4. **The database is the scaling bottleneck**: To scale beyond current limits, consider:
   - Database read replicas
   - Database sharding by colony
   - PgBouncer for connection pooling at the database level

5. **Exclusive assign should only be used** when strict ordering guarantees are required

## Recommendations

| Scenario | Recommended Configuration |
|----------|---------------------------|
| Low-medium load, single region | 1 replica, distributed assign |
| High availability required | 3 replicas, distributed assign |
| Multiple colonies, high load | Consider database sharding |
| Strict ordering required | Exclusive assign (accept latency cost) |
