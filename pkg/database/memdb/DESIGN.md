# MemDB Database Design Document

## Overview

MemDB is a generic, high-availability, in-memory database engine that provides persistence, replication, and configurable storage modes. It uses a pluggable schema system that can adapt to any domain model, making it suitable for ColonyOS, microservices, IoT systems, and other real-time applications.

## Architecture

### Core Components

```
pkg/database/memdb/
├── engine/
│   ├── database.go      # Main Database engine & coordination
│   ├── schema.go        # Generic schema system
│   ├── collection.go    # Generic collection management
│   └── query.go         # Generic query engine
├── storage/
│   ├── memory.go        # In-memory storage engine
│   ├── persistence.go   # Disk persistence layer
│   └── compaction.go    # Data compaction and cleanup
├── replication/
│   ├── raft.go         # Raft consensus implementation
│   ├── log.go          # Replication log management
│   └── snapshot.go     # Snapshot creation and restoration
├── indexing/
│   ├── btree.go        # B-tree indexes for range queries
│   ├── hash.go         # Hash indexes for exact matches
│   ├── timeseries.go   # Time-series optimized indexes
│   └── composite.go    # Composite/multi-field indexes
├── adapters/
│   ├── colonies.go     # ColonyOS-specific adapter
│   ├── generic.go      # Generic key-value adapter
│   └── sql.go          # SQL-like interface adapter
├── config/
│   └── config.go       # Configuration management
└── metrics/
    └── metrics.go      # Performance metrics and monitoring
```

## Storage Architecture

### 1. Multi-Mode Storage

```go
type StorageMode string

const (
    MemoryOnly     StorageMode = "memory-only"     // Pure in-memory, no persistence
    MemoryFirst    StorageMode = "memory-first"    // Memory with background persistence
    HybridMode     StorageMode = "hybrid"          // Hot data in memory, cold on disk
    PersistentMode StorageMode = "persistent"      // All data persisted immediately
)
```

### 2. Generic Schema System

```go
// Generic schema definition
type Schema struct {
    Name        string                    `json:"name"`
    Version     string                    `json:"version"`
    Collections map[string]*Collection    `json:"collections"`
}

type Collection struct {
    Name        string                    `json:"name"`
    PrimaryKey  string                    `json:"primary_key"`
    Fields      map[string]*Field         `json:"fields"`
    Indexes     map[string]*Index         `json:"indexes"`
    Relationships map[string]*Relationship `json:"relationships"`
}

type Field struct {
    Name        string      `json:"name"`
    Type        FieldType   `json:"type"`
    Required    bool        `json:"required"`
    Unique      bool        `json:"unique"`
    Default     interface{} `json:"default,omitempty"`
    Constraints []string    `json:"constraints,omitempty"`
}

type FieldType string
const (
    StringType    FieldType = "string"
    IntType       FieldType = "int"
    FloatType     FieldType = "float"
    BoolType      FieldType = "bool"
    TimeType      FieldType = "time"
    JSONType      FieldType = "json"
    BytesType     FieldType = "bytes"
    ArrayType     FieldType = "array"
    MapType       FieldType = "map"
)
```

### 3. Memory Management

```go
type MemoryManager struct {
    // Generic data storage - collection-based
    collections map[string]*CollectionStore
    
    // Schema management
    schema      *Schema
    
    // Indexing structures
    indexes     *IndexManager
    
    // Memory limits and thresholds
    maxMemoryMB    int64
    currentMemoryMB int64
    flushThreshold float64  // e.g., 0.8 for 80% full
    
    // Synchronization
    rwMutex sync.RWMutex
}

type CollectionStore struct {
    name        string
    primaryKey  string
    data        map[string]interface{}  // key -> document
    indexes     map[string]Index        // index_name -> index
    rwMutex     sync.RWMutex
}
```

### 3. Persistence Layer

```go
type PersistenceEngine struct {
    // Write-ahead log for durability
    wal         *WriteAheadLog
    
    // Snapshot management
    snapshots   *SnapshotManager
    
    // Background persistence
    flushQueue  chan FlushRequest
    compactor   *DataCompactor
    
    // Storage backends
    diskPath    string
    compression bool
}
```

## High Availability & Replication

### 1. Raft Consensus

```go
type ReplicationManager struct {
    // Raft consensus
    raftNode    *raft.Node
    nodeID      string
    peers       []string
    
    // State machine
    stateMachine *MemDBStateMachine
    
    // Log replication
    logStore    *raft.LogStore
    
    // Leader election
    leaderCh    chan bool
    isLeader    bool
}
```

### 2. Data Replication Strategy

- **Synchronous replication** for critical operations (colony creation, executor registration)
- **Asynchronous replication** for high-volume operations (process updates, logs)
- **Quorum-based consistency** requiring majority agreement
- **Automatic failover** with leader election

### 3. Network Partitions

```go
type PartitionHandling struct {
    // Split-brain detection
    quorumSize      int
    connectedNodes  int
    
    // Partition tolerance
    readOnlyMode    bool
    conflictResolver *ConflictResolver
}
```

## Performance Optimizations

### 1. Generic Indexing System

```go
type IndexManager struct {
    // Generic index storage
    indexes         map[string]Index            // index_name -> index implementation
    collections     map[string]*CollectionStore // collection_name -> store
    schema          *Schema
}

// Generic index interface
type Index interface {
    Name() string
    Type() IndexType
    Fields() []string
    Insert(key interface{}, docID string) error
    Delete(key interface{}, docID string) error
    Find(query *IndexQuery) ([]string, error) // returns document IDs
    Range(start, end interface{}) ([]string, error)
    Stats() *IndexStats
}

type IndexType string
const (
    HashIndex       IndexType = "hash"        // Fast exact matches
    BTreeIndex      IndexType = "btree"       // Range queries, sorting
    TimeSeriesIndex IndexType = "timeseries"  // Time-based queries
    FullTextIndex   IndexType = "fulltext"    // Text search
    CompositeIndex  IndexType = "composite"   // Multi-field indexes
    GeoIndex        IndexType = "geo"         // Geospatial queries
)

type IndexQuery struct {
    Type        QueryType    `json:"type"`
    Field       string       `json:"field"`
    Value       interface{}  `json:"value"`
    StartValue  interface{}  `json:"start_value,omitempty"`
    EndValue    interface{}  `json:"end_value,omitempty"`
    Conditions  []*Condition `json:"conditions,omitempty"`
}
```

### 2. Query Optimization

- **Index-aware query planning** for complex filters
- **Batch operations** for bulk inserts/updates
- **Connection pooling** for concurrent access
- **Read replicas** for scaling read workloads

### 3. Memory Efficiency

```go
type MemoryOptimization struct {
    // Object pooling
    processPool     sync.Pool
    stringPool      sync.Pool
    
    // Compression for cold data
    compressor      *Compressor
    
    // Garbage collection optimization
    gcTuning        *GCTuner
}
```

## Configuration

### 1. Database Configuration

```go
type Config struct {
    // Storage configuration
    StorageMode         StorageMode     `yaml:"storage_mode"`
    MaxMemoryMB        int64           `yaml:"max_memory_mb"`
    FlushThresholdPct  float64         `yaml:"flush_threshold_pct"`
    PersistencePath    string          `yaml:"persistence_path"`
    CompressionEnabled bool            `yaml:"compression_enabled"`
    
    // Replication configuration  
    ReplicationEnabled bool            `yaml:"replication_enabled"`
    NodeID            string          `yaml:"node_id"`
    PeerNodes         []string        `yaml:"peer_nodes"`
    QuorumSize        int             `yaml:"quorum_size"`
    
    // Performance tuning
    IndexingEnabled   bool            `yaml:"indexing_enabled"`
    BatchSize        int             `yaml:"batch_size"`
    FlushIntervalMS  int             `yaml:"flush_interval_ms"`
    
    // Monitoring
    MetricsEnabled   bool            `yaml:"metrics_enabled"`
    HealthCheckMS    int             `yaml:"health_check_ms"`
}
```

### 2. Example Configurations

```yaml
# High-performance in-memory only
memory_only:
  storage_mode: "memory-only"
  max_memory_mb: 8192
  indexing_enabled: true
  replication_enabled: true
  quorum_size: 3

# Persistent with HA
ha_persistent:
  storage_mode: "memory-first"
  max_memory_mb: 4096
  flush_threshold_pct: 0.8
  persistence_path: "/var/lib/colonies/memdb"
  compression_enabled: true
  replication_enabled: true
  peer_nodes: ["node1:8080", "node2:8080", "node3:8080"]

# Single node persistent
single_persistent:
  storage_mode: "persistent"
  persistence_path: "/var/lib/colonies/memdb"
  replication_enabled: false
```

## Implementation Phases

### Phase 1: Core Storage Engine
- [ ] Basic in-memory storage
- [ ] Interface implementations
- [ ] Memory management
- [ ] Simple persistence

### Phase 2: Indexing & Performance  
- [ ] B-tree and hash indexes
- [ ] Query optimization
- [ ] Batch operations
- [ ] Memory efficiency

### Phase 3: High Availability
- [ ] Raft consensus implementation
- [ ] Data replication
- [ ] Leader election
- [ ] Failover handling

### Phase 4: Advanced Features
- [ ] Conflict resolution
- [ ] Metrics and monitoring  
- [ ] Performance tuning
- [ ] Backup/restore

## Consistency-Aware Database Interface

MemDB provides **API-level consistency control**, allowing developers to choose the appropriate consistency model for each operation at runtime.

### 1. Consistency Levels

```go
type ConsistencyLevel string

const (
    Local      ConsistencyLevel = "local"      // Local node only - fastest
    Eventual   ConsistencyLevel = "eventual"   // Async replication - high throughput  
    Strong     ConsistencyLevel = "strong"     // Raft consensus - strong consistency
    Quorum     ConsistencyLevel = "quorum"     // Majority agreement - balanced
)
```

### 2. Core Generic Interface

```go
// Generic database interface with consistency control
type Database interface {
    // Schema management
    LoadSchema(schema *Schema) error
    GetSchema() *Schema
    
    // Collection operations
    CreateCollection(name string, config *CollectionConfig) error
    DropCollection(name string) error
    ListCollections() []string
    
    // Document operations with consistency parameter
    Insert(collection string, doc interface{}, level ConsistencyLevel) (string, error)
    Update(collection, id string, updates map[string]interface{}, level ConsistencyLevel) error
    Delete(collection, id string, level ConsistencyLevel) error
    Get(collection, id string, level ConsistencyLevel) (interface{}, error)
    
    // Query operations with consistency control
    Find(collection string, query *Query, level ConsistencyLevel) (*ResultSet, error)
    Count(collection string, query *Query, level ConsistencyLevel) (int64, error)
    Aggregate(collection string, pipeline []AggregateStage, level ConsistencyLevel) (*ResultSet, error)
    
    // Atomic operations with consistency guarantees
    CompareAndSwap(collection, id string, expected, new map[string]interface{}, level ConsistencyLevel) (bool, error)
    
    // Batch operations
    BatchOperations(ops []Operation, level ConsistencyLevel) ([]Result, error)
    
    // Transaction operations
    BeginTransaction(level ConsistencyLevel) (Transaction, error)
    
    // Index operations
    CreateIndex(collection string, index *IndexDefinition) error
    DropIndex(collection string, indexName string) error
    
    // Lifecycle
    Close() error
}
```

### 3. ColonyOS Adapter with Smart Consistency

```go
// ColonyOS-specific adapter with consistency-aware operations
type ColoniesDatabase struct {
    core Database
    schemas map[string]*Schema
}

// Process assignment - requires strong consistency to prevent race conditions
func (cd *ColoniesDatabase) AssignProcess(executorID string) (*core.Process, error) {
    // Fast local read to find candidates
    processes, err := cd.core.Find("processes", &Query{
        Filter: map[string]interface{}{
            "state": "waiting",
            "executor_type": executorID.Type,
        },
        Limit: 1,
    }, Local)
    
    if err != nil || len(processes) == 0 {
        return nil, ErrNoAvailableProcesses
    }
    
    process := processes[0].(*core.Process)
    
    // Critical assignment - use strong consistency with CAS
    success, err := cd.core.CompareAndSwap("processes", process.ID,
        map[string]interface{}{"state": "waiting"}, // Expected
        map[string]interface{}{                     // New
            "state": "assigned",
            "assigned_to": executorID,
            "assigned_at": time.Now(),
        },
        Strong, // Raft consensus ensures exactly-once assignment
    )
    
    if !success || err != nil {
        return nil, ErrAssignmentFailed
    }
    
    return process, nil
}

// Colony creation - business critical, strong consistency
func (cd *ColoniesDatabase) AddColony(colony *core.Colony) (*core.Colony, error) {
    id, err := cd.core.Insert("colonies", colony, Strong)
    if err != nil {
        return nil, err
    }
    colony.ID = id
    return colony, nil
}

// Read operations - local for speed
func (cd *ColoniesDatabase) GetColony(id string) (*core.Colony, error) {
    result, err := cd.core.Get("colonies", id, Local)
    if err != nil {
        return nil, err
    }
    return result.(*core.Colony), nil
}

// Process state updates - eventual consistency for performance
func (cd *ColoniesDatabase) UpdateProcessState(processID string, state core.ProcessState) error {
    return cd.core.Update("processes", processID, 
        map[string]interface{}{
            "state": state,
            "updated_at": time.Now(),
        },
        Eventual, // High throughput, async replication
    )
}

// Log operations - eventual consistency, high volume
func (cd *ColoniesDatabase) AddProcessLog(processID, message string) error {
    _, err := cd.core.Insert("process_logs", map[string]interface{}{
        "process_id": processID,
        "message": message,
        "timestamp": time.Now(),
    }, Eventual)
    
    return err
}

// Search operations - local reads for dashboard performance
func (cd *ColoniesDatabase) SearchProcesses(query *ProcessQuery) ([]*core.Process, error) {
    results, err := cd.core.Find("processes", &Query{
        Filter: query.ToFilter(),
        Sort: map[string]SortOrder{"created_at": Descending},
        Limit: query.Limit,
    }, Local) // Fast local search
    
    if err != nil {
        return nil, err
    }
    
    processes := make([]*core.Process, len(results))
    for i, result := range results {
        processes[i] = result.(*core.Process)
    }
    
    return processes, nil
}
```

### 4. Smart Database Implementation

```go
// Smart database router that handles different consistency levels
type SmartMemDB struct {
    // Storage layers
    localStore    *LocalMemDB     // Local in-memory storage
    raftStore     *RaftMemDB      // Distributed consensus storage
    
    // Cluster state
    isLeader      bool
    clusterNodes  []*Node
    
    // Replication
    replicator    *AsyncReplicator
}

// Main operation dispatcher based on consistency level
func (db *SmartMemDB) Execute(operation string, params map[string]interface{}, level ConsistencyLevel) (interface{}, error) {
    switch level {
    case Local:
        return db.executeLocal(operation, params)
    case Eventual:
        return db.executeEventual(operation, params)
    case Strong:
        return db.executeStrong(operation, params)
    case Quorum:
        return db.executeQuorum(operation, params)
    default:
        return nil, ErrUnknownConsistencyLevel
    }
}

// Local execution - fastest, read from local node only
func (db *SmartMemDB) executeLocal(operation string, params map[string]interface{}) (interface{}, error) {
    return db.localStore.Execute(operation, params)
}

// Eventual consistency - local write + async replication
func (db *SmartMemDB) executeEventual(operation string, params map[string]interface{}) (interface{}, error) {
    // Execute locally first for fast response
    result, err := db.localStore.Execute(operation, params)
    if err != nil {
        return nil, err
    }
    
    // Async replication to other nodes (fire and forget for performance)
    go db.replicator.Replicate(operation, params, result)
    
    return result, nil
}

// Strong consistency - Raft consensus required
func (db *SmartMemDB) executeStrong(operation string, params map[string]interface{}) (interface{}, error) {
    if !db.isLeader {
        return nil, ErrNotLeader
    }
    
    cmd := &Command{
        Operation: operation,
        Params:    params,
        Timestamp: time.Now(),
    }
    
    // All nodes must agree before returning
    future := db.raftStore.Apply(cmd, 5*time.Second)
    return future.Response(), future.Error()
}

// Quorum consistency - majority agreement required
func (db *SmartMemDB) executeQuorum(operation string, params map[string]interface{}) (interface{}, error) {
    return db.executeWithQuorum(operation, params, len(db.clusterNodes)/2+1)
}
```

### 5. Advanced Usage Patterns

```go
// Builder pattern for complex queries with consistency control
type QueryBuilder struct {
    collection string
    query      *Query
    level      ConsistencyLevel
    timeout    time.Duration
    db         Database
}

func (db Database) Query(collection string) *QueryBuilder {
    return &QueryBuilder{
        collection: collection,
        query:      &Query{},
        level:      Local, // default
        db:         db,
    }
}

func (qb *QueryBuilder) Filter(filter map[string]interface{}) *QueryBuilder {
    qb.query.Filter = filter
    return qb
}

func (qb *QueryBuilder) Consistency(level ConsistencyLevel) *QueryBuilder {
    qb.level = level
    return qb
}

func (qb *QueryBuilder) Execute() ([]interface{}, error) {
    return qb.db.Find(qb.collection, qb.query, qb.level)
}

// Usage - fluent API for complex operations
processes, err := db.Query("processes").
    Filter(map[string]interface{}{"state": "waiting"}).
    Consistency(Strong).
    Execute()

// Context-based consistency hints
func WithConsistency(ctx context.Context, level ConsistencyLevel) context.Context {
    return context.WithValue(ctx, "consistency", level)
}

// Database operations respect context consistency hints
func (db *SmartMemDB) GetWithContext(ctx context.Context, collection, id string) (interface{}, error) {
    level := Local // default
    if consistency := ctx.Value("consistency"); consistency != nil {
        level = consistency.(ConsistencyLevel)
    }
    return db.Get(collection, id, level)
}
```

## Deployment Scenarios

### 1. Generic Usage - Any Application

```go
// Create database with custom schema
schema := &memdb.Schema{
    Name: "my-app",
    Version: "1.0.0",
    Collections: map[string]*memdb.Collection{
        "users": {
            Name: "users",
            PrimaryKey: "id",
            Fields: map[string]*memdb.Field{
                "id":       {Name: "id", Type: memdb.StringType, Required: true, Unique: true},
                "email":    {Name: "email", Type: memdb.StringType, Required: true, Unique: true},
                "name":     {Name: "name", Type: memdb.StringType, Required: true},
                "created":  {Name: "created", Type: memdb.TimeType, Required: true},
            },
        },
    },
}

config := &memdb.Config{
    StorageMode: memdb.MemoryOnly,
    MaxMemoryMB: 1024,
}
db, err := memdb.NewDatabase(config)
db.LoadSchema(schema)

// Generic operations
user := map[string]interface{}{
    "id": "user-123",
    "email": "john@example.com",
    "name": "John Doe",
    "created": time.Now(),
}
db.Insert("users", user)
```

### 2. ColonyOS Integration with Consistency Control

```go
// Load ColonyOS schema
coloniesSchema := memdb.LoadColoniesSchema()
db.LoadSchema(coloniesSchema)

// Create ColonyOS adapter
coloniesDB := memdb.NewColoniesAdapter(db)

// Use exactly like PostgreSQL - adapter handles consistency automatically
server := server.CreateServer(coloniesDB, ...)

// Specific consistency examples:
// Fast dashboard reads
processes := coloniesDB.GetProcesses() // Uses Local consistency internally

// Critical process assignment  
assigned := coloniesDB.AssignProcess(executorID) // Uses Strong consistency internally

// High-volume logging
coloniesDB.AddLogs(logs) // Uses Eventual consistency internally
```

### 3. Microservice with HA

```go
config := &memdb.Config{
    StorageMode:        memdb.MemoryFirst,
    ReplicationEnabled: true,
    PeerNodes:         []string{"node1:8080", "node2:8080", "node3:8080"},
    QuorumSize:        2,
    PersistencePath:   "/var/lib/myservice/memdb",
}
db, err := memdb.NewDatabase(config)

// Load custom service schema
db.LoadSchema(myServiceSchema)
```

## Monitoring & Metrics

```go
type Metrics struct {
    // Storage metrics
    MemoryUsageMB      int64
    DiskUsageMB       int64
    RecordCount       int64
    IndexSizeMB       int64
    
    // Performance metrics  
    QueriesPerSecond   float64
    AverageLatencyMS   float64
    FlushRate         float64
    
    // Replication metrics
    ReplicationLag    time.Duration  
    LeaderElections   int64
    NodeStatus        map[string]string
}
```

## Migration Strategy

### For ColonyOS
1. **Development**: Start with MemoryOnly mode
2. **Testing**: Use MemoryFirst for persistence testing  
3. **Staging**: Deploy HA configuration
4. **Production**: Full HA with monitoring

### For Other Applications
1. **Define Schema**: Create schema definition for your domain model
2. **Implement Adapter**: Create adapter interface for your existing code
3. **Gradual Migration**: Replace existing database calls incrementally
4. **Scale Out**: Add replication and HA as needed

## Consistency Model Benefits

### **🎯 Per-Operation Control**

| Operation Type | Consistency Level | Benefits |
|----------------|-------------------|----------|
| **Process Assignment** | Strong | Prevents race conditions, exactly-once semantics |
| **Colony Creation** | Strong | Business critical, prevents duplicates |
| **Process State Updates** | Eventual | High throughput, acceptable delays |
| **Log Writes** | Eventual | Maximum performance for high-volume data |
| **Dashboard Reads** | Local | Fastest response times |
| **Search Queries** | Local | Low latency for user interfaces |

### **⚡ Performance Characteristics**

```go
// Latency comparison (typical values)
Local:    < 1ms      // Memory access only
Eventual: 1-5ms      // Local write + async replication  
Quorum:   5-20ms     // Majority node agreement
Strong:   10-50ms    // Full Raft consensus
```

### **🔧 Runtime Flexibility**

```go
// Adapt consistency based on context
func (api *ColoniesAPI) AssignProcess(ctx context.Context, executorID string) (*Process, error) {
    // Normal case: strong consistency
    level := Strong
    
    // During network partition: prefer availability
    if api.isPartitioned() {
        level = Quorum
    }
    
    // During maintenance: local only
    if api.isMaintenanceMode() {
        level = Local
    }
    
    return api.db.AssignProcessWithLevel(executorID, level)
}
```

## Benefits of API-Level Consistency Control

### ✅ **Granular Control**
- Choose consistency per operation, not globally
- Runtime decisions based on business requirements
- No configuration files or restarts needed

### ✅ **Performance Optimization**
- Fast local reads for dashboards and searches  
- Strong consistency only where race conditions matter
- Eventual consistency for high-throughput operations

### ✅ **Developer Experience**
- Explicit consistency semantics in code
- Easy to understand performance implications
- Familiar API patterns with consistency parameter

### ✅ **Operational Flexibility**
- Adjust consistency during network partitions
- Degrade gracefully under load
- Test different consistency models easily

### ✅ **Generic Reusability**
- Not tied to ColonyOS domain model
- Pluggable schema system for any application
- Adapter pattern for existing interfaces

The **API-level consistency control** makes MemDB both a **high-performance specialized database** for ColonyOS and a **generic, reusable database engine** for any Go application requiring flexible consistency guarantees! 🎯