package raft

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

// RaftNode wraps HashiCorp Raft for VelocityDB
type RaftNode struct {
	raft       *raft.Raft
	config     *RaftConfig
	fsm        *VelocityFSM
	transport  *raft.NetworkTransport
	logStore   raft.LogStore
	stableStore raft.StableStore
	snapshots  raft.SnapshotStore
	mu         sync.RWMutex
	shutdown   chan struct{}
}

// RaftConfig holds Raft configuration
type RaftConfig struct {
	NodeID       string
	RaftDir      string
	RaftBind     string
	LocalID      raft.ServerID
	LogLevel     string
	SnapshotRetain int
	SnapshotThreshold uint64
}

// LogEntry represents a Raft log entry for database operations
type LogEntry struct {
	Type       string                 `json:"type"`        // insert, update, delete, cas
	Collection string                 `json:"collection"`
	Key        string                 `json:"key"`
	Value      map[string]interface{} `json:"value,omitempty"`
	Expected   map[string]interface{} `json:"expected,omitempty"` // For CAS
	Fields     map[string]interface{} `json:"fields,omitempty"`   // For updates
	Timestamp  time.Time             `json:"timestamp"`
	RequestID  string                `json:"request_id"`
}

// LogResponse represents the response to a log entry
type LogResponse struct {
	Success   bool        `json:"success"`
	Error     string      `json:"error,omitempty"`
	Value     interface{} `json:"value,omitempty"`
	Version   uint64      `json:"version,omitempty"`
}

// StorageBackend interface for the underlying storage
type StorageBackend interface {
	Insert(ctx context.Context, collection string, key string, value map[string]interface{}) error
	Update(ctx context.Context, collection string, key string, fields map[string]interface{}) (map[string]interface{}, uint64, error)
	Delete(ctx context.Context, collection string, key string) error
	Get(ctx context.Context, collection string, key string) (map[string]interface{}, uint64, error)
	CompareAndSwap(ctx context.Context, collection string, key string, expected, value map[string]interface{}) (bool, map[string]interface{}, uint64, error)
}

// VelocityFSM implements the Raft finite state machine for VelocityDB
type VelocityFSM struct {
	storage StorageBackend
	mu      sync.RWMutex
}

// NewRaftNode creates a new Raft node
func NewRaftNode(config *RaftConfig, storage StorageBackend) (*RaftNode, error) {
	// Ensure raft directory exists
	if err := os.MkdirAll(config.RaftDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create raft directory: %w", err)
	}

	// Create FSM
	fsm := &VelocityFSM{
		storage: storage,
	}

	// Setup Raft configuration
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = config.LocalID
	raftConfig.LogLevel = config.LogLevel
	raftConfig.SnapshotThreshold = config.SnapshotThreshold
	raftConfig.TrailingLogs = 1024

	// Create transport
	addr, err := net.ResolveTCPAddr("tcp", config.RaftBind)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve raft bind address: %w", err)
	}

	transport, err := raft.NewTCPTransport(config.RaftBind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Create stores using BoltDB
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(config.RaftDir, "raft.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to create log store: %w", err)
	}

	// BoltDB can serve as both log and stable store
	stableStore := logStore

	// Create snapshot store
	snapshots, err := raft.NewFileSnapshotStore(config.RaftDir, config.SnapshotRetain, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot store: %w", err)
	}

	// Create Raft instance
	r, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, fmt.Errorf("failed to create raft: %w", err)
	}

	node := &RaftNode{
		raft:        r,
		config:      config,
		fsm:         fsm,
		transport:   transport,
		logStore:    logStore,
		stableStore: stableStore,
		snapshots:   snapshots,
		shutdown:    make(chan struct{}),
	}

	return node, nil
}

// Bootstrap creates a single-node cluster
func (r *RaftNode) Bootstrap() error {
	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      r.config.LocalID,
				Address: r.transport.LocalAddr(),
			},
		},
	}

	future := r.raft.BootstrapCluster(configuration)
	return future.Error()
}

// Join adds this node to an existing cluster
func (r *RaftNode) Join(leaderAddr string) error {
	return fmt.Errorf("join functionality not implemented yet")
}

// IsLeader returns true if this node is the Raft leader
func (r *RaftNode) IsLeader() bool {
	return r.raft.State() == raft.Leader
}

// LeaderAddr returns the address of the current leader
func (r *RaftNode) LeaderAddr() raft.ServerAddress {
	return r.raft.Leader()
}

// ApplyLog applies a log entry to the Raft cluster
func (r *RaftNode) ApplyLog(entry *LogEntry) (*LogResponse, error) {
	if !r.IsLeader() {
		return nil, fmt.Errorf("not the leader")
	}

	// Serialize the log entry
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Apply to Raft
	future := r.raft.Apply(data, 5*time.Second)
	if err := future.Error(); err != nil {
		return nil, fmt.Errorf("failed to apply log: %w", err)
	}

	// Extract response
	response, ok := future.Response().(*LogResponse)
	if !ok {
		return nil, fmt.Errorf("invalid response type")
	}

	return response, nil
}

// WaitForLeader waits for a leader to be elected
func (r *RaftNode) WaitForLeader(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for leader")
		case <-ticker.C:
			if r.raft.Leader() != "" {
				return nil
			}
		}
	}
}

// Stats returns Raft statistics
func (r *RaftNode) Stats() map[string]string {
	return r.raft.Stats()
}

// Shutdown gracefully shuts down the Raft node
func (r *RaftNode) Shutdown() error {
	close(r.shutdown)

	// Shutdown Raft
	future := r.raft.Shutdown()
	if err := future.Error(); err != nil {
		return fmt.Errorf("failed to shutdown raft: %w", err)
	}

	// Close stores
	if closer, ok := r.logStore.(io.Closer); ok {
		closer.Close()
	}
	if closer, ok := r.stableStore.(io.Closer); ok {
		closer.Close()
	}

	return nil
}

// FSM implementation

// Apply applies a Raft log entry to the FSM
func (fsm *VelocityFSM) Apply(log *raft.Log) interface{} {
	var entry LogEntry
	if err := json.Unmarshal(log.Data, &entry); err != nil {
		return &LogResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to unmarshal log entry: %v", err),
		}
	}

	ctx := context.Background()
	
	switch entry.Type {
	case "insert":
		err := fsm.storage.Insert(ctx, entry.Collection, entry.Key, entry.Value)
		if err != nil {
			return &LogResponse{
				Success: false,
				Error:   err.Error(),
			}
		}
		return &LogResponse{Success: true}

	case "update":
		value, version, err := fsm.storage.Update(ctx, entry.Collection, entry.Key, entry.Fields)
		if err != nil {
			return &LogResponse{
				Success: false,
				Error:   err.Error(),
			}
		}
		return &LogResponse{
			Success: true,
			Value:   value,
			Version: version,
		}

	case "delete":
		err := fsm.storage.Delete(ctx, entry.Collection, entry.Key)
		if err != nil {
			return &LogResponse{
				Success: false,
				Error:   err.Error(),
			}
		}
		return &LogResponse{Success: true}

	case "cas":
		success, value, version, err := fsm.storage.CompareAndSwap(ctx, entry.Collection, entry.Key, entry.Expected, entry.Value)
		if err != nil {
			return &LogResponse{
				Success: false,
				Error:   err.Error(),
			}
		}
		return &LogResponse{
			Success: success,
			Value:   value,
			Version: version,
		}

	default:
		return &LogResponse{
			Success: false,
			Error:   fmt.Sprintf("unknown log entry type: %s", entry.Type),
		}
	}
}

// Snapshot creates a snapshot of the FSM state
func (fsm *VelocityFSM) Snapshot() (raft.FSMSnapshot, error) {
	// For now, return a simple snapshot
	// In a full implementation, this would serialize the entire state
	return &VelocitySnapshot{}, nil
}

// Restore restores the FSM state from a snapshot
func (fsm *VelocityFSM) Restore(rc io.ReadCloser) error {
	// For now, just close the reader
	// In a full implementation, this would restore the entire state
	return rc.Close()
}

// VelocitySnapshot implements raft.FSMSnapshot
type VelocitySnapshot struct{}

func (s *VelocitySnapshot) Persist(sink raft.SnapshotSink) error {
	// Write empty snapshot for now
	_, err := sink.Write([]byte("{}"))
	if err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *VelocitySnapshot) Release() {}