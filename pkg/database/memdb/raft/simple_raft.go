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
	"go.etcd.io/bbolt"
)

// SimpleRaftNode is a simplified Raft implementation using built-in stores
type SimpleRaftNode struct {
	raft       *raft.Raft
	config     *SimpleRaftConfig
	fsm        *SimpleFSM
	transport  *raft.NetworkTransport
	logStore   raft.LogStore
	stableStore raft.StableStore
	snapshots  raft.SnapshotStore
	mu         sync.RWMutex
}

// SimpleRaftConfig holds simplified Raft configuration
type SimpleRaftConfig struct {
	NodeID       string
	RaftDir      string
	RaftBind     string
	LocalID      raft.ServerID
}

// SimpleFSM implements a basic finite state machine
type SimpleFSM struct {
	storage map[string]map[string]interface{} // collection -> key -> value
	mu      sync.RWMutex
}

// SimpleLogStore implements raft.LogStore using BoltDB
type SimpleLogStore struct {
	db *bbolt.DB
}

// NewSimpleRaftNode creates a simplified Raft node
func NewSimpleRaftNode(config *SimpleRaftConfig) (*SimpleRaftNode, error) {
	// Ensure raft directory exists
	if err := os.MkdirAll(config.RaftDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create raft directory: %w", err)
	}

	// Create FSM
	fsm := &SimpleFSM{
		storage: make(map[string]map[string]interface{}),
	}

	// Setup Raft configuration
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = config.LocalID
	raftConfig.LogLevel = "WARN"

	// Create transport
	addr, err := net.ResolveTCPAddr("tcp", config.RaftBind)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve raft bind address: %w", err)
	}

	transport, err := raft.NewTCPTransport(config.RaftBind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Create stores
	logStore, err := NewSimpleLogStore(filepath.Join(config.RaftDir, "logs.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to create log store: %w", err)
	}

	stableStore := logStore // Use same store for both

	// Create snapshot store
	snapshots, err := raft.NewFileSnapshotStore(config.RaftDir, 2, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot store: %w", err)
	}

	// Create Raft instance
	r, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, fmt.Errorf("failed to create raft: %w", err)
	}

	node := &SimpleRaftNode{
		raft:        r,
		config:      config,
		fsm:         fsm,
		transport:   transport,
		logStore:    logStore,
		stableStore: stableStore,
		snapshots:   snapshots,
	}

	return node, nil
}

// Bootstrap creates a single-node cluster
func (r *SimpleRaftNode) Bootstrap() error {
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

// IsLeader returns true if this node is the Raft leader
func (r *SimpleRaftNode) IsLeader() bool {
	return r.raft.State() == raft.Leader
}

// Apply applies a log entry to the Raft cluster
func (r *SimpleRaftNode) Apply(data []byte) error {
	if !r.IsLeader() {
		return fmt.Errorf("not the leader")
	}

	future := r.raft.Apply(data, 5*time.Second)
	return future.Error()
}

// WaitForLeader waits for a leader to be elected
func (r *SimpleRaftNode) WaitForLeader(timeout time.Duration) error {
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

// Shutdown shuts down the Raft node
func (r *SimpleRaftNode) Shutdown() error {
	future := r.raft.Shutdown()
	if err := future.Error(); err != nil {
		return err
	}

	// Close stores
	if r.logStore != nil {
		if closer, ok := r.logStore.(io.Closer); ok {
			closer.Close()
		}
	}

	return nil
}

// SimpleFSM implementation

func (fsm *SimpleFSM) Apply(log *raft.Log) interface{} {
	var entry map[string]interface{}
	if err := json.Unmarshal(log.Data, &entry); err != nil {
		return fmt.Errorf("failed to unmarshal: %v", err)
	}

	fsm.mu.Lock()
	defer fsm.mu.Unlock()

	op, ok := entry["op"].(string)
	if !ok {
		return fmt.Errorf("missing operation")
	}

	collection, ok := entry["collection"].(string)
	if !ok {
		return fmt.Errorf("missing collection")
	}

	if fsm.storage[collection] == nil {
		fsm.storage[collection] = make(map[string]interface{})
	}

	switch op {
	case "set":
		key := entry["key"].(string)
		value := entry["value"]
		fsm.storage[collection][key] = value
		return "OK"
	case "delete":
		key := entry["key"].(string)
		delete(fsm.storage[collection], key)
		return "OK"
	default:
		return fmt.Errorf("unknown operation: %s", op)
	}
}

func (fsm *SimpleFSM) Snapshot() (raft.FSMSnapshot, error) {
	fsm.mu.RLock()
	defer fsm.mu.RUnlock()

	// Make a copy of the data
	snapshot := make(map[string]map[string]interface{})
	for collection, data := range fsm.storage {
		snapshot[collection] = make(map[string]interface{})
		for key, value := range data {
			snapshot[collection][key] = value
		}
	}

	return &SimpleSnapshot{data: snapshot}, nil
}

func (fsm *SimpleFSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	var data map[string]map[string]interface{}
	if err := json.NewDecoder(rc).Decode(&data); err != nil {
		return err
	}

	fsm.mu.Lock()
	fsm.storage = data
	fsm.mu.Unlock()

	return nil
}

// Get retrieves a value
func (fsm *SimpleFSM) Get(collection, key string) (interface{}, bool) {
	fsm.mu.RLock()
	defer fsm.mu.RUnlock()

	if collectionData, exists := fsm.storage[collection]; exists {
		value, ok := collectionData[key]
		return value, ok
	}
	return nil, false
}

// SimpleSnapshot implements raft.FSMSnapshot
type SimpleSnapshot struct {
	data map[string]map[string]interface{}
}

func (s *SimpleSnapshot) Persist(sink raft.SnapshotSink) error {
	encoder := json.NewEncoder(sink)
	if err := encoder.Encode(s.data); err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *SimpleSnapshot) Release() {}

// SimpleLogStore implementation using BoltDB

func NewSimpleLogStore(path string) (*SimpleLogStore, error) {
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	// Create buckets
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("logs"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("conf"))
		return err
	})

	if err != nil {
		db.Close()
		return nil, err
	}

	return &SimpleLogStore{db: db}, nil
}

func (s *SimpleLogStore) Close() error {
	return s.db.Close()
}

// Implement raft.LogStore interface
func (s *SimpleLogStore) FirstIndex() (uint64, error) {
	var first uint64
	err := s.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("logs")).Cursor()
		if k, _ := c.First(); k != nil {
			first = bytesToUint64(k)
		}
		return nil
	})
	return first, err
}

func (s *SimpleLogStore) LastIndex() (uint64, error) {
	var last uint64
	err := s.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("logs")).Cursor()
		if k, _ := c.Last(); k != nil {
			last = bytesToUint64(k)
		}
		return nil
	})
	return last, err
}

func (s *SimpleLogStore) GetLog(index uint64, log *raft.Log) error {
	return s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("logs"))
		val := bucket.Get(uint64ToBytes(index))
		if val == nil {
			return raft.ErrLogNotFound
		}
		return json.Unmarshal(val, log)
	})
}

func (s *SimpleLogStore) StoreLog(log *raft.Log) error {
	return s.StoreLogs([]*raft.Log{log})
}

func (s *SimpleLogStore) StoreLogs(logs []*raft.Log) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("logs"))
		for _, log := range logs {
			key := uint64ToBytes(log.Index)
			val, err := json.Marshal(log)
			if err != nil {
				return err
			}
			if err := bucket.Put(key, val); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SimpleLogStore) DeleteRange(min, max uint64) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("logs")).Cursor()
		for k, _ := c.Seek(uint64ToBytes(min)); k != nil && bytesToUint64(k) <= max; k, _ = c.Next() {
			if err := c.Delete(); err != nil {
				return err
			}
		}
		return nil
	})
}

// Implement raft.StableStore interface
func (s *SimpleLogStore) Set(key []byte, val []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("conf"))
		return bucket.Put(key, val)
	})
}

func (s *SimpleLogStore) Get(key []byte) ([]byte, error) {
	var val []byte
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("conf"))
		val = bucket.Get(key)
		if val != nil {
			// Make a copy since BoltDB data is only valid during transaction
			val = append([]byte(nil), val...)
		}
		return nil
	})
	return val, err
}

func (s *SimpleLogStore) SetUint64(key []byte, val uint64) error {
	return s.Set(key, uint64ToBytes(val))
}

func (s *SimpleLogStore) GetUint64(key []byte) (uint64, error) {
	val, err := s.Get(key)
	if err != nil || val == nil {
		return 0, err
	}
	return bytesToUint64(val), nil
}

// Helper functions
func uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	for i := 0; i < 8; i++ {
		buf[i] = byte(u >> (8 * (7 - i)))
	}
	return buf
}

func bytesToUint64(b []byte) uint64 {
	var u uint64
	for i := 0; i < 8; i++ {
		u |= uint64(b[i]) << (8 * (7 - i))
	}
	return u
}