package consistency

import (
	"context"
	"fmt"
	"time"

	"github.com/colonyos/colonies/pkg/database/memdb/core"
	"github.com/colonyos/colonies/pkg/database/memdb/schema"
)

// ConsistencyLevel defines the consistency requirements for operations
type ConsistencyLevel int

const (
	Local ConsistencyLevel = iota
	Eventual
	Strong
	Quorum
)

func (c ConsistencyLevel) String() string {
	switch c {
	case Local:
		return "Local"
	case Eventual:
		return "Eventual" 
	case Strong:
		return "Strong"
	case Quorum:
		return "Quorum"
	default:
		return "Unknown"
	}
}

// Document represents a document with consistency metadata
type Document struct {
	*core.Document
	ConsistencyLevel ConsistencyLevel `json:"consistency_level"`
	ReplicationLevel int             `json:"replication_level"`
}

// DatabaseRouter routes operations to appropriate consistency handlers
type DatabaseRouter struct {
	localDB    LocalDatabase
	raftDB     RaftDatabase
	quorumDB   QuorumDatabase
	eventualDB EventualDatabase
	config     *RouterConfig
}

// RouterConfig holds configuration for the database router
type RouterConfig struct {
	DefaultConsistency  ConsistencyLevel
	QuorumSize         int
	ReplicationFactor  int
	EventualSyncDelay  time.Duration
}

// LocalDatabase interface for local operations
type LocalDatabase interface {
	Insert(ctx context.Context, collection string, doc *core.Document) error
	Update(ctx context.Context, collection string, id string, fields map[string]interface{}) (*core.Document, error)
	Delete(ctx context.Context, collection string, id string) error
	Get(ctx context.Context, collection string, id string) (*core.Document, error)
	Query(ctx context.Context, collection string, filter map[string]interface{}, limit, offset int) ([]*core.Document, error)
	Count(ctx context.Context, collection string) (int, error)
	CreateCollection(ctx context.Context, name string, sch *schema.Schema) error
	DropCollection(ctx context.Context, name string) error
	Close() error
}

// RaftDatabase interface for strong consistency operations
type RaftDatabase interface {
	LocalDatabase
	ProposeOperation(ctx context.Context, op *Operation) error
	IsLeader() bool
	WaitForApply(ctx context.Context, index uint64, timeout time.Duration) error
}

// QuorumDatabase interface for quorum-based operations
type QuorumDatabase interface {
	LocalDatabase
	QuorumRead(ctx context.Context, collection string, id string, quorumSize int) (*core.Document, error)
	QuorumWrite(ctx context.Context, collection string, doc *core.Document, quorumSize int) error
}

// EventualDatabase interface for eventually consistent operations
type EventualDatabase interface {
	LocalDatabase
	AsyncReplicate(ctx context.Context, operation *Operation) error
	GetReplicationStatus(ctx context.Context) (*ReplicationStatus, error)
}

// Operation represents a database operation for replication
type Operation struct {
	Type       string      `json:"type"`        // insert, update, delete
	Collection string      `json:"collection"`
	Document   *core.Document `json:"document,omitempty"`
	ID         string      `json:"id,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
	NodeID     string      `json:"node_id"`
}

// ReplicationStatus provides replication status information
type ReplicationStatus struct {
	PendingOps    int       `json:"pending_ops"`
	LastSync      time.Time `json:"last_sync"`
	ReplicationLag time.Duration `json:"replication_lag"`
}

// NewDatabaseRouter creates a new database router
func NewDatabaseRouter(config *RouterConfig, localDB LocalDatabase, raftDB RaftDatabase, quorumDB QuorumDatabase, eventualDB EventualDatabase) *DatabaseRouter {
	if config == nil {
		config = &RouterConfig{
			DefaultConsistency: Local,
			QuorumSize:        3,
			ReplicationFactor: 3,
			EventualSyncDelay: 100 * time.Millisecond,
		}
	}

	return &DatabaseRouter{
		localDB:    localDB,
		raftDB:     raftDB,
		quorumDB:   quorumDB,
		eventualDB: eventualDB,
		config:     config,
	}
}

// Insert inserts a document with the specified consistency level
func (r *DatabaseRouter) Insert(ctx context.Context, collection string, doc *core.Document, consistency ConsistencyLevel) error {
	switch consistency {
	case Local:
		return r.localDB.Insert(ctx, collection, doc)
	case Strong:
		if r.raftDB == nil {
			return fmt.Errorf("raft database not configured for strong consistency")
		}
		op := &Operation{
			Type:       "insert",
			Collection: collection,
			Document:   doc,
			Timestamp:  time.Now(),
		}
		return r.raftDB.ProposeOperation(ctx, op)
	case Quorum:
		if r.quorumDB == nil {
			return fmt.Errorf("quorum database not configured")
		}
		return r.quorumDB.QuorumWrite(ctx, collection, doc, r.config.QuorumSize)
	case Eventual:
		if r.eventualDB == nil {
			return fmt.Errorf("eventual database not configured")
		}
		// Write locally first
		err := r.localDB.Insert(ctx, collection, doc)
		if err != nil {
			return err
		}
		// Async replication
		op := &Operation{
			Type:       "insert",
			Collection: collection,
			Document:   doc,
			Timestamp:  time.Now(),
		}
		return r.eventualDB.AsyncReplicate(ctx, op)
	default:
		return fmt.Errorf("unsupported consistency level: %v", consistency)
	}
}

// Update updates a document with the specified consistency level
func (r *DatabaseRouter) Update(ctx context.Context, collection string, id string, fields map[string]interface{}, consistency ConsistencyLevel) (*core.Document, error) {
	switch consistency {
	case Local:
		return r.localDB.Update(ctx, collection, id, fields)
	case Strong:
		if r.raftDB == nil {
			return nil, fmt.Errorf("raft database not configured for strong consistency")
		}
		op := &Operation{
			Type:       "update",
			Collection: collection,
			ID:         id,
			Fields:     fields,
			Timestamp:  time.Now(),
		}
		err := r.raftDB.ProposeOperation(ctx, op)
		if err != nil {
			return nil, err
		}
		// Return updated document
		return r.raftDB.Get(ctx, collection, id)
	case Quorum:
		if r.quorumDB == nil {
			return nil, fmt.Errorf("quorum database not configured")
		}
		// For quorum updates, we need to implement a more complex protocol
		return r.quorumUpdate(ctx, collection, id, fields)
	case Eventual:
		if r.eventualDB == nil {
			return nil, fmt.Errorf("eventual database not configured")
		}
		// Update locally first
		doc, err := r.localDB.Update(ctx, collection, id, fields)
		if err != nil {
			return nil, err
		}
		// Async replication
		op := &Operation{
			Type:       "update",
			Collection: collection,
			ID:         id,
			Fields:     fields,
			Timestamp:  time.Now(),
		}
		r.eventualDB.AsyncReplicate(ctx, op)
		return doc, nil
	default:
		return nil, fmt.Errorf("unsupported consistency level: %v", consistency)
	}
}

// Delete deletes a document with the specified consistency level
func (r *DatabaseRouter) Delete(ctx context.Context, collection string, id string, consistency ConsistencyLevel) error {
	switch consistency {
	case Local:
		return r.localDB.Delete(ctx, collection, id)
	case Strong:
		if r.raftDB == nil {
			return fmt.Errorf("raft database not configured for strong consistency")
		}
		op := &Operation{
			Type:       "delete",
			Collection: collection,
			ID:         id,
			Timestamp:  time.Now(),
		}
		return r.raftDB.ProposeOperation(ctx, op)
	case Quorum:
		if r.quorumDB == nil {
			return fmt.Errorf("quorum database not configured")
		}
		return r.quorumDelete(ctx, collection, id)
	case Eventual:
		if r.eventualDB == nil {
			return fmt.Errorf("eventual database not configured")
		}
		// Delete locally first
		err := r.localDB.Delete(ctx, collection, id)
		if err != nil {
			return err
		}
		// Async replication
		op := &Operation{
			Type:       "delete",
			Collection: collection,
			ID:         id,
			Timestamp:  time.Now(),
		}
		return r.eventualDB.AsyncReplicate(ctx, op)
	default:
		return fmt.Errorf("unsupported consistency level: %v", consistency)
	}
}

// Get retrieves a document with the specified consistency level
func (r *DatabaseRouter) Get(ctx context.Context, collection string, id string, consistency ConsistencyLevel) (*core.Document, error) {
	switch consistency {
	case Local:
		return r.localDB.Get(ctx, collection, id)
	case Strong:
		if r.raftDB == nil {
			return nil, fmt.Errorf("raft database not configured for strong consistency")
		}
		return r.raftDB.Get(ctx, collection, id)
	case Quorum:
		if r.quorumDB == nil {
			return nil, fmt.Errorf("quorum database not configured")
		}
		return r.quorumDB.QuorumRead(ctx, collection, id, r.config.QuorumSize)
	case Eventual:
		// For eventual consistency, local read is acceptable
		return r.localDB.Get(ctx, collection, id)
	default:
		return nil, fmt.Errorf("unsupported consistency level: %v", consistency)
	}
}

// Query queries documents with the specified consistency level
func (r *DatabaseRouter) Query(ctx context.Context, collection string, filter map[string]interface{}, limit, offset int, consistency ConsistencyLevel) ([]*core.Document, error) {
	switch consistency {
	case Local, Eventual:
		// For queries, local is usually acceptable even for eventual consistency
		return r.localDB.Query(ctx, collection, filter, limit, offset)
	case Strong:
		if r.raftDB == nil {
			return nil, fmt.Errorf("raft database not configured for strong consistency")
		}
		return r.raftDB.Query(ctx, collection, filter, limit, offset)
	case Quorum:
		// For quorum queries, this would require a more complex implementation
		// For now, fall back to strong consistency
		if r.raftDB != nil {
			return r.raftDB.Query(ctx, collection, filter, limit, offset)
		}
		return r.localDB.Query(ctx, collection, filter, limit, offset)
	default:
		return nil, fmt.Errorf("unsupported consistency level: %v", consistency)
	}
}

// CreateCollection creates a collection with the specified consistency level
func (r *DatabaseRouter) CreateCollection(ctx context.Context, name string, sch *schema.Schema, consistency ConsistencyLevel) error {
	switch consistency {
	case Local:
		return r.localDB.CreateCollection(ctx, name, sch)
	case Strong:
		if r.raftDB == nil {
			return fmt.Errorf("raft database not configured for strong consistency")
		}
		// Collection creation should be replicated
		return r.raftDB.CreateCollection(ctx, name, sch)
	case Quorum, Eventual:
		// For collection operations, use strong consistency to ensure schema consistency
		if r.raftDB != nil {
			return r.raftDB.CreateCollection(ctx, name, sch)
		}
		return r.localDB.CreateCollection(ctx, name, sch)
	default:
		return fmt.Errorf("unsupported consistency level: %v", consistency)
	}
}

// Helper methods for quorum operations

func (r *DatabaseRouter) quorumUpdate(ctx context.Context, collection string, id string, fields map[string]interface{}) (*core.Document, error) {
	// Simplified quorum update - in a real implementation this would involve
	// reading from quorum, checking versions, and writing to quorum
	return r.localDB.Update(ctx, collection, id, fields)
}

func (r *DatabaseRouter) quorumDelete(ctx context.Context, collection string, id string) error {
	// Simplified quorum delete - in a real implementation this would involve
	// coordinating with quorum of nodes
	return r.localDB.Delete(ctx, collection, id)
}

// Close closes all database connections
func (r *DatabaseRouter) Close() error {
	var errs []error
	
	if r.localDB != nil {
		if err := r.localDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("local db close: %w", err))
		}
	}
	
	if r.raftDB != nil {
		if err := r.raftDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("raft db close: %w", err))
		}
	}
	
	if r.quorumDB != nil {
		if err := r.quorumDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("quorum db close: %w", err))
		}
	}
	
	if r.eventualDB != nil {
		if err := r.eventualDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("eventual db close: %w", err))
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}
	
	return nil
}