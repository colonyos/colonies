package localstore

import "io"

// ObjectStore defines the interface for file data storage backends.
type ObjectStore interface {
	// Put stores an object from reader. Overwrites if exists.
	Put(colonyName, objectName string, reader io.Reader, size int64) error

	// PutChunk stores a single chunk of a chunked upload.
	PutChunk(colonyName, objectName string, chunkIndex int, reader io.Reader, size int64) error

	// AssembleChunks combines all chunks into the final object and removes staging data.
	AssembleChunks(colonyName, objectName string, totalChunks int) error

	// GetChunkStatus returns the list of chunk indices that have been uploaded.
	GetChunkStatus(colonyName, objectName string) ([]int, error)

	// Get returns a reader for the entire object and its size.
	Get(colonyName, objectName string) (io.ReadCloser, int64, error)

	// GetRange returns a reader for a byte range of the object.
	GetRange(colonyName, objectName string, offset, length int64) (io.ReadCloser, error)

	// Remove deletes an object.
	Remove(colonyName, objectName string) error

	// RemoveStaging deletes staging data for a chunked upload.
	RemoveStaging(colonyName, objectName string) error

	// Exists returns true if the object exists.
	Exists(colonyName, objectName string) bool

	// Stat returns metadata about an object.
	Stat(colonyName, objectName string) (ObjectInfo, error)

	// List returns all object names in a colony.
	List(colonyName string) ([]string, error)

	// RemoveAll deletes all objects for a colony.
	RemoveAll(colonyName string) error

	// DiskUsage returns the total bytes used by a colony.
	DiskUsage(colonyName string) (int64, error)
}

// ObjectInfo contains metadata about a stored object.
type ObjectInfo struct {
	Size int64
}
