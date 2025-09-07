package memdb

import (
	"context"
	"time"
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

// StorageMode defines how data is stored and persisted
type StorageMode int

const (
	MemoryOnly StorageMode = iota
	MemoryFirst
	Hybrid
	Persistent
)

func (s StorageMode) String() string {
	switch s {
	case MemoryOnly:
		return "MemoryOnly"
	case MemoryFirst:
		return "MemoryFirst"
	case Hybrid:
		return "Hybrid"
	case Persistent:
		return "Persistent"
	default:
		return "Unknown"
	}
}

// CASRequest represents a Compare-And-Swap operation
type CASRequest struct {
	Key      string
	Expected interface{}
	Value    interface{}
	TTL      time.Duration
}

// CASResponse represents the result of a CAS operation
type CASResponse struct {
	Success      bool
	CurrentValue interface{}
	Version      uint64
}

// Document represents a generic document in the database
type Document struct {
	ID       string                 `json:"id"`
	Fields   map[string]interface{} `json:"fields"`
	Version  uint64                 `json:"version"`
	Created  time.Time             `json:"created"`
	Modified time.Time             `json:"modified"`
}

// Query represents a database query
type Query struct {
	Collection string
	Filter     map[string]interface{}
	Sort       []string
	Limit      int
	Offset     int
}

// QueryResult represents query results
type QueryResult struct {
	Documents []*Document
	Total     int
	HasMore   bool
}

// Database is the main interface for MemDB operations
type Database interface {
	// Collection operations
	CreateCollection(ctx context.Context, name string, sch interface{}, consistency ConsistencyLevel) error
	DropCollection(ctx context.Context, name string, consistency ConsistencyLevel) error
	ListCollections(ctx context.Context) ([]string, error)

	// Document operations
	Insert(ctx context.Context, collection string, doc *Document, consistency ConsistencyLevel) error
	Update(ctx context.Context, collection string, id string, fields map[string]interface{}, consistency ConsistencyLevel) (*Document, error)
	Delete(ctx context.Context, collection string, id string, consistency ConsistencyLevel) error
	Get(ctx context.Context, collection string, id string, consistency ConsistencyLevel) (*Document, error)
	
	// Query operations
	Query(ctx context.Context, query *Query, consistency ConsistencyLevel) (*QueryResult, error)
	Count(ctx context.Context, collection string, filter map[string]interface{}, consistency ConsistencyLevel) (int, error)

	// CAS operations
	CompareAndSwap(ctx context.Context, collection string, cas *CASRequest, consistency ConsistencyLevel) (*CASResponse, error)

	// Batch operations
	Batch(ctx context.Context, ops []BatchOperation, consistency ConsistencyLevel) error

	// Health and status
	Health(ctx context.Context) error
	Stats(ctx context.Context) (*DatabaseStats, error)

	// Lifecycle
	Close() error
}

// BatchOperation represents a single operation in a batch
type BatchOperation struct {
	Type       string      // "insert", "update", "delete"
	Collection string
	Document   *Document
	ID         string
	Fields     map[string]interface{}
}

// DatabaseStats provides database statistics
type DatabaseStats struct {
	Collections int64
	Documents   int64
	MemoryUsage int64
	DiskUsage   int64
	Uptime      time.Duration
}