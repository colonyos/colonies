package memdb

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/colonyos/colonies/pkg/database/memdb/cluster"
	velocityRaft "github.com/colonyos/colonies/pkg/database/memdb/raft"
	"github.com/google/uuid"
	"github.com/hashicorp/raft"
)

// DistributedVelocityDB combines VelocityDB with Raft consensus and cluster membership
type DistributedVelocityDB struct {
	localDB     *VelocityDB
	raftNode    *velocityRaft.RaftNode
	cluster     *cluster.ClusterManager
	config      *DistributedConfig
	nodeID      string
	mu          sync.RWMutex
	readyC      chan struct{}
	isReady     bool
}

// DistributedConfig holds configuration for distributed VelocityDB
type DistributedConfig struct {
	// Local storage config
	DataDir     string
	CacheSize   int
	InMemory    bool

	// Node identity
	NodeName    string
	NodeID      string

	// Raft configuration
	RaftDir     string
	RaftBind    string
	RaftPort    int

	// Memberlist configuration  
	ClusterBind string
	ClusterPort int
	SeedNodes   []string

	// Cluster behavior
	Bootstrap           bool
	WaitForLeaderTimeout time.Duration
	MinClusterSize       int
}

// StorageAdapter adapts VelocityDB to the Raft storage interface
type StorageAdapter struct {
	db *VelocityDB
}

// NewDistributedVelocityDB creates a new distributed VelocityDB instance
func NewDistributedVelocityDB(config *DistributedConfig) (*DistributedVelocityDB, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Set defaults
	if config.NodeID == "" {
		config.NodeID = uuid.New().String()
	}
	if config.NodeName == "" {
		config.NodeName = config.NodeID
	}
	if config.WaitForLeaderTimeout == 0 {
		config.WaitForLeaderTimeout = 30 * time.Second
	}
	if config.MinClusterSize == 0 {
		config.MinClusterSize = 1
	}

	// Create local VelocityDB
	velocityConfig := &VelocityConfig{
		DataDir:   config.DataDir,
		CacheSize: config.CacheSize,
		InMemory:  config.InMemory,
	}

	localDB, err := NewVelocityDB(velocityConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create local VelocityDB: %w", err)
	}

	// Create storage adapter for Raft
	storageAdapter := &StorageAdapter{db: localDB}

	// Create Raft node
	raftConfig := &velocityRaft.RaftConfig{
		NodeID:            config.NodeID,
		RaftDir:           config.RaftDir,
		RaftBind:          fmt.Sprintf("%s:%d", config.RaftBind, config.RaftPort),
		LocalID:           raft.ServerID(config.NodeName),
		LogLevel:          "WARN",
		SnapshotRetain:    2,
		SnapshotThreshold: 1000,
	}

	raftNode, err := velocityRaft.NewRaftNode(raftConfig, storageAdapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Raft node: %w", err)
	}

	// Create cluster manager
	clusterConfig := &cluster.ClusterConfig{
		NodeName:  config.NodeName,
		BindAddr:  config.ClusterBind,
		BindPort:  config.ClusterPort,
		RaftAddr:  raftConfig.RaftBind,
		SeedNodes: config.SeedNodes,
		LogOutput: false,
	}

	clusterManager, err := cluster.NewClusterManager(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster manager: %w", err)
	}

	ddb := &DistributedVelocityDB{
		localDB:  localDB,
		raftNode: raftNode,
		cluster:  clusterManager,
		config:   config,
		nodeID:   config.NodeID,
		readyC:   make(chan struct{}),
	}

	// Start the distributed database
	if err := ddb.start(); err != nil {
		return nil, fmt.Errorf("failed to start distributed database: %w", err)
	}

	return ddb, nil
}

// start initializes and starts the distributed database
func (d *DistributedVelocityDB) start() error {
	// Join cluster
	if err := d.cluster.Join(d.config.SeedNodes); err != nil {
		return fmt.Errorf("failed to join cluster: %w", err)
	}

	// Wait for minimum cluster size
	if d.config.MinClusterSize > 1 {
		discovery := cluster.NewDiscoveryHelper(d.cluster)
		err := discovery.WaitForMinimumMembers(d.config.MinClusterSize, d.config.WaitForLeaderTimeout)
		if err != nil {
			return fmt.Errorf("failed to reach minimum cluster size: %w", err)
		}
	}

	// Bootstrap or join Raft cluster
	if d.config.Bootstrap {
		if err := d.raftNode.Bootstrap(); err != nil {
			return fmt.Errorf("failed to bootstrap Raft: %w", err)
		}
	}

	// Wait for Raft leader
	if err := d.raftNode.WaitForLeader(d.config.WaitForLeaderTimeout); err != nil {
		return fmt.Errorf("failed to elect leader: %w", err)
	}

	// Mark as ready
	d.mu.Lock()
	d.isReady = true
	close(d.readyC)
	d.mu.Unlock()

	return nil
}

// WaitForReady waits until the distributed database is ready
func (d *DistributedVelocityDB) WaitForReady(timeout time.Duration) error {
	select {
	case <-d.readyC:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for database to become ready")
	}
}

// IsReady returns true if the distributed database is ready
func (d *DistributedVelocityDB) IsReady() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.isReady
}

// IsLeader returns true if this node is the Raft leader
func (d *DistributedVelocityDB) IsLeader() bool {
	return d.raftNode.IsLeader()
}

// LeaderAddr returns the address of the current Raft leader
func (d *DistributedVelocityDB) LeaderAddr() string {
	return string(d.raftNode.LeaderAddr())
}

// GetClusterMembers returns all cluster members
func (d *DistributedVelocityDB) GetClusterMembers() []*cluster.ClusterMember {
	return d.cluster.GetMembers()
}

// Insert adds a new document with strong consistency (Raft)
func (d *DistributedVelocityDB) Insert(ctx context.Context, collection string, doc *VelocityDocument) error {
	if !d.IsReady() {
		return fmt.Errorf("database not ready")
	}

	entry := &velocityRaft.LogEntry{
		Type:       "insert",
		Collection: collection,
		Key:        doc.ID,
		Value:      doc.Fields,
		Timestamp:  time.Now(),
		RequestID:  uuid.New().String(),
	}

	response, err := d.raftNode.ApplyLog(entry)
	if err != nil {
		return fmt.Errorf("failed to apply insert: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("insert failed: %s", response.Error)
	}

	return nil
}

// Get retrieves a document with local consistency (fast read)
func (d *DistributedVelocityDB) Get(ctx context.Context, collection string, id string) (*VelocityDocument, error) {
	// Local reads for performance
	return d.localDB.Get(ctx, collection, id)
}

// GetStrong retrieves a document with strong consistency (leader read)
func (d *DistributedVelocityDB) GetStrong(ctx context.Context, collection string, id string) (*VelocityDocument, error) {
	if !d.IsLeader() {
		return nil, fmt.Errorf("strong reads require leader")
	}
	
	return d.localDB.Get(ctx, collection, id)
}

// Update modifies a document with strong consistency
func (d *DistributedVelocityDB) Update(ctx context.Context, collection string, id string, fields map[string]interface{}) (*VelocityDocument, error) {
	if !d.IsReady() {
		return nil, fmt.Errorf("database not ready")
	}

	entry := &velocityRaft.LogEntry{
		Type:       "update",
		Collection: collection,
		Key:        id,
		Fields:     fields,
		Timestamp:  time.Now(),
		RequestID:  uuid.New().String(),
	}

	response, err := d.raftNode.ApplyLog(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to apply update: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("update failed: %s", response.Error)
	}

	// Return updated document
	return d.localDB.Get(ctx, collection, id)
}

// Delete removes a document with strong consistency
func (d *DistributedVelocityDB) Delete(ctx context.Context, collection string, id string) error {
	if !d.IsReady() {
		return fmt.Errorf("database not ready")
	}

	entry := &velocityRaft.LogEntry{
		Type:       "delete",
		Collection: collection,
		Key:        id,
		Timestamp:  time.Now(),
		RequestID:  uuid.New().String(),
	}

	response, err := d.raftNode.ApplyLog(entry)
	if err != nil {
		return fmt.Errorf("failed to apply delete: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("delete failed: %s", response.Error)
	}

	return nil
}

// CompareAndSwap performs atomic compare-and-swap with strong consistency
func (d *DistributedVelocityDB) CompareAndSwap(ctx context.Context, collection string, cas *VelocityCASRequest) (*VelocityCASResult, error) {
	if !d.IsReady() {
		return nil, fmt.Errorf("database not ready")
	}

	expectedMap, ok := cas.Expected.(map[string]interface{})
	if !ok && cas.Expected != nil {
		return nil, fmt.Errorf("expected value must be map[string]interface{} or nil")
	}

	valueMap, ok := cas.Value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("value must be map[string]interface{}")
	}

	entry := &velocityRaft.LogEntry{
		Type:       "cas",
		Collection: collection,
		Key:        cas.Key,
		Expected:   expectedMap,
		Value:      valueMap,
		Timestamp:  time.Now(),
		RequestID:  uuid.New().String(),
	}

	response, err := d.raftNode.ApplyLog(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to apply CAS: %w", err)
	}

	return &VelocityCASResult{
		Success:      response.Success,
		CurrentValue: response.Value,
		Version:      response.Version,
	}, nil
}

// List returns documents with local consistency (fast read)
func (d *DistributedVelocityDB) List(ctx context.Context, collection string, limit, offset int) ([]*VelocityDocument, error) {
	return d.localDB.List(ctx, collection, limit, offset)
}

// Count returns document count with local consistency
func (d *DistributedVelocityDB) Count(ctx context.Context, collection string) (int, error) {
	return d.localDB.Count(ctx, collection)
}

// Health checks the health of the distributed database
func (d *DistributedVelocityDB) Health(ctx context.Context) error {
	if !d.IsReady() {
		return fmt.Errorf("database not ready")
	}

	// Check local database
	if err := d.localDB.Health(ctx); err != nil {
		return fmt.Errorf("local database unhealthy: %w", err)
	}

	// Check Raft status
	stats := d.raftNode.Stats()
	state, ok := stats["state"]
	if !ok || (state != "Leader" && state != "Follower") {
		return fmt.Errorf("raft node in invalid state: %s", state)
	}

	return nil
}

// GetClusterInfo returns information about the cluster
func (d *DistributedVelocityDB) GetClusterInfo() map[string]interface{} {
	info := make(map[string]interface{})
	
	info["node_id"] = d.nodeID
	info["is_leader"] = d.IsLeader()
	info["leader_addr"] = d.LeaderAddr()
	info["cluster_size"] = d.cluster.NumMembers()
	info["alive_members"] = d.cluster.NumAliveMembers()
	info["raft_stats"] = d.raftNode.Stats()
	
	members := d.cluster.GetMembers()
	memberInfo := make([]map[string]interface{}, len(members))
	for i, member := range members {
		memberInfo[i] = map[string]interface{}{
			"id":        member.ID,
			"name":      member.Name,
			"addr":      member.Addr,
			"port":      member.Port,
			"raft_addr": member.RaftAddr,
			"state":     member.State,
		}
	}
	info["members"] = memberInfo
	
	return info
}

// Shutdown gracefully shuts down the distributed database
func (d *DistributedVelocityDB) Shutdown() error {
	var errs []error

	// Shutdown cluster first
	if err := d.cluster.Leave(5 * time.Second); err != nil {
		errs = append(errs, fmt.Errorf("cluster leave error: %w", err))
	}

	// Shutdown Raft
	if err := d.raftNode.Shutdown(); err != nil {
		errs = append(errs, fmt.Errorf("raft shutdown error: %w", err))
	}

	// Close cluster
	if err := d.cluster.Shutdown(); err != nil {
		errs = append(errs, fmt.Errorf("cluster shutdown error: %w", err))
	}

	// Close local database
	if err := d.localDB.Close(); err != nil {
		errs = append(errs, fmt.Errorf("local db close error: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}

// StorageAdapter implementation

func (s *StorageAdapter) Insert(ctx context.Context, collection string, key string, value map[string]interface{}) error {
	doc := &VelocityDocument{
		ID:     key,
		Fields: value,
	}
	return s.db.Insert(ctx, collection, doc)
}

func (s *StorageAdapter) Update(ctx context.Context, collection string, key string, fields map[string]interface{}) (map[string]interface{}, uint64, error) {
	updated, err := s.db.Update(ctx, collection, key, fields)
	if err != nil {
		return nil, 0, err
	}
	return updated.Fields, updated.Version, nil
}

func (s *StorageAdapter) Delete(ctx context.Context, collection string, key string) error {
	return s.db.Delete(ctx, collection, key)
}

func (s *StorageAdapter) Get(ctx context.Context, collection string, key string) (map[string]interface{}, uint64, error) {
	doc, err := s.db.Get(ctx, collection, key)
	if err != nil {
		return nil, 0, err
	}
	return doc.Fields, doc.Version, nil
}

func (s *StorageAdapter) CompareAndSwap(ctx context.Context, collection string, key string, expected, value map[string]interface{}) (bool, map[string]interface{}, uint64, error) {
	cas := &VelocityCASRequest{
		Key:      key,
		Expected: expected,
		Value:    value,
	}

	result, err := s.db.CompareAndSwap(ctx, collection, cas)
	if err != nil {
		return false, nil, 0, err
	}

	currentMap, ok := result.CurrentValue.(map[string]interface{})
	if !ok {
		currentMap = make(map[string]interface{})
	}

	return result.Success, currentMap, result.Version, nil
}

// Helper functions

// parseAddr parses "host:port" and returns host, port
func parseAddr(addr string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, err
	}
	
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}
	
	return host, port, nil
}