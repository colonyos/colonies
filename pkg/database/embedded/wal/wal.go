package wal

// OpType represents the type of WAL operation.
type OpType uint8

const (
	OpPut    OpType = 1
	OpDelete OpType = 2
)

// Entry represents a single WAL entry.
type Entry struct {
	Op        OpType
	Entity    string // e.g. "processes", "executors"
	Key       string // primary key
	Data      []byte // JSON-encoded record (nil for deletes)
	Timestamp int64  // unix nanoseconds
}

// SyncMode controls when the WAL calls fsync.
type SyncMode int

const (
	// SyncAlways calls fsync after every append (safest, slowest).
	SyncAlways SyncMode = iota
	// SyncNone never calls fsync (fastest, data loss on crash).
	SyncNone
)

// WAL is the interface for write-ahead logging. Implementations must be safe
// for concurrent use by multiple goroutines.
type WAL interface {
	// Append writes an entry to the log. Depending on the SyncMode, the
	// entry may or may not be durable when Append returns.
	Append(entry Entry) error

	// Replay reads all entries from the log in order, calling fn for each.
	// If fn returns an error, replay stops and the error is returned.
	Replay(fn func(Entry) error) error

	// Truncate removes all entries from the log. Called after a successful
	// flush to disk.
	Truncate() error

	// Close flushes any buffered data and closes the log.
	Close() error
}
