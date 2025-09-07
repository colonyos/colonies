package memdb

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVelocityDB_BasicOperations(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	ctx := context.Background()

	// Test Insert
	doc := &VelocityDocument{
		ID: "test1",
		Fields: map[string]interface{}{
			"name": "John Doe",
			"age":  30,
		},
	}

	err := db.Insert(ctx, "users", doc)
	assert.NoError(t, err)

	// Test Get
	retrieved, err := db.Get(ctx, "users", "test1")
	assert.NoError(t, err)
	assert.Equal(t, "test1", retrieved.ID)
	assert.Equal(t, "John Doe", retrieved.Fields["name"])
	assert.Equal(t, float64(30), retrieved.Fields["age"]) // JSON unmarshaling makes numbers float64
	assert.Equal(t, uint64(1), retrieved.Version)
	assert.False(t, retrieved.Created.IsZero())

	// Test duplicate insert should fail
	err = db.Insert(ctx, "users", doc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestVelocityDB_Update(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	ctx := context.Background()

	// Insert initial document
	doc := &VelocityDocument{
		ID: "test1",
		Fields: map[string]interface{}{
			"name": "John",
			"age":  25,
		},
	}
	err := db.Insert(ctx, "users", doc)
	require.NoError(t, err)

	// Update document
	updatedFields := map[string]interface{}{
		"age":  26,
		"city": "New York",
	}

	updated, err := db.Update(ctx, "users", "test1", updatedFields)
	assert.NoError(t, err)
	// The update operation preserves the original type
	assert.Equal(t, 26, updated.Fields["age"])
	assert.Equal(t, "New York", updated.Fields["city"])
	assert.Equal(t, "John", updated.Fields["name"]) // Existing field preserved
	assert.Equal(t, uint64(2), updated.Version)     // Version incremented

	// Test update non-existent document
	_, err = db.Update(ctx, "users", "nonexistent", updatedFields)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestVelocityDB_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	ctx := context.Background()

	// Insert and then delete
	doc := &VelocityDocument{
		ID:     "test1",
		Fields: map[string]interface{}{"name": "John"},
	}
	err := db.Insert(ctx, "users", doc)
	require.NoError(t, err)

	err = db.Delete(ctx, "users", "test1")
	assert.NoError(t, err)

	// Verify deletion
	_, err = db.Get(ctx, "users", "test1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test delete non-existent document
	err = db.Delete(ctx, "users", "nonexistent")
	assert.Error(t, err)
}

func TestVelocityDB_List(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	ctx := context.Background()

	// Insert multiple documents
	docs := []*VelocityDocument{
		{ID: "user1", Fields: map[string]interface{}{"name": "Alice", "age": 25}},
		{ID: "user2", Fields: map[string]interface{}{"name": "Bob", "age": 30}},
		{ID: "user3", Fields: map[string]interface{}{"name": "Charlie", "age": 35}},
	}

	for _, doc := range docs {
		err := db.Insert(ctx, "users", doc)
		require.NoError(t, err)
	}

	// Test list all
	allUsers, err := db.List(ctx, "users", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, allUsers, 3)

	// Test pagination
	firstTwo, err := db.List(ctx, "users", 2, 0)
	assert.NoError(t, err)
	assert.Len(t, firstTwo, 2)

	nextOne, err := db.List(ctx, "users", 2, 2)
	assert.NoError(t, err)
	assert.Len(t, nextOne, 1)

	// Test count
	count, err := db.Count(ctx, "users")
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestVelocityDB_CompareAndSwap(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	ctx := context.Background()

	// Test CAS on non-existent document (creation)
	cas := &VelocityCASRequest{
		Key:      "process1",
		Expected: nil,
		Value: map[string]interface{}{
			"state":       "running",
			"executor_id": "executor1",
		},
	}

	result, err := db.CompareAndSwap(ctx, "processes", cas)
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, uint64(1), result.Version)

	// Verify document was created
	doc, err := db.Get(ctx, "processes", "process1")
	assert.NoError(t, err)
	assert.Equal(t, "running", doc.Fields["state"])
	assert.Equal(t, "executor1", doc.Fields["executor_id"])

	// Test successful CAS update
	cas2 := &VelocityCASRequest{
		Key: "process1",
		Expected: map[string]interface{}{
			"state":       "running",
			"executor_id": "executor1",
		},
		Value: map[string]interface{}{
			"state":       "completed",
			"executor_id": "executor1",
			"end_time":    time.Now().Unix(),
		},
	}

	result2, err := db.CompareAndSwap(ctx, "processes", cas2)
	assert.NoError(t, err)
	assert.True(t, result2.Success)
	assert.Equal(t, uint64(2), result2.Version)

	// Test failed CAS (wrong expected value)
	cas3 := &VelocityCASRequest{
		Key: "process1",
		Expected: map[string]interface{}{
			"state": "running", // This is wrong, should be "completed"
		},
		Value: map[string]interface{}{
			"state": "failed",
		},
	}

	result3, err := db.CompareAndSwap(ctx, "processes", cas3)
	assert.NoError(t, err)
	assert.False(t, result3.Success)
	assert.NotNil(t, result3.CurrentValue)
}

func TestVelocityDB_ConcurrentAccess(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	ctx := context.Background()
	const numWorkers = 10
	const operationsPerWorker = 100

	var wg sync.WaitGroup
	errors := make([]error, numWorkers*operationsPerWorker)
	errorIndex := 0
	var errorMutex sync.Mutex

	recordError := func(err error) {
		errorMutex.Lock()
		defer errorMutex.Unlock()
		if err != nil {
			errors[errorIndex] = err
			errorIndex++
		}
	}

	// Concurrent inserts
	t.Run("ConcurrentInserts", func(t *testing.T) {
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < operationsPerWorker; j++ {
					doc := &VelocityDocument{
						ID: fmt.Sprintf("worker%d_doc%d", workerID, j),
						Fields: map[string]interface{}{
							"worker_id": workerID,
							"doc_id":    j,
							"timestamp": time.Now().Unix(),
						},
					}
					err := db.Insert(ctx, "concurrent_test", doc)
					recordError(err)
				}
			}(i)
		}
		wg.Wait()

		// Check for errors
		for i := 0; i < errorIndex; i++ {
			assert.NoError(t, errors[i])
		}

		// Verify all documents were inserted
		count, err := db.Count(ctx, "concurrent_test")
		assert.NoError(t, err)
		assert.Equal(t, numWorkers*operationsPerWorker, count)
	})

	// Reset for next test
	errorIndex = 0

	// Concurrent CAS operations (simulate process assignment)
	t.Run("ConcurrentCAS", func(t *testing.T) {
		// Create initial processes
		for i := 0; i < operationsPerWorker; i++ {
			doc := &VelocityDocument{
				ID: fmt.Sprintf("process_%d", i),
				Fields: map[string]interface{}{
					"state": "waiting",
				},
			}
			err := db.Insert(ctx, "cas_test", doc)
			require.NoError(t, err)
		}

		successCount := int64(0)
		var successMutex sync.Mutex

		// Multiple workers try to assign processes
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < operationsPerWorker; j++ {
					cas := &VelocityCASRequest{
						Key:      fmt.Sprintf("process_%d", j),
						Expected: map[string]interface{}{"state": "waiting"},
						Value: map[string]interface{}{
							"state":       "running",
							"executor_id": fmt.Sprintf("executor_%d", workerID),
							"assigned_at": time.Now().Unix(),
						},
					}

					result, err := db.CompareAndSwap(ctx, "cas_test", cas)
					recordError(err)
					if err == nil && result.Success {
						successMutex.Lock()
						successCount++
						successMutex.Unlock()
					}
				}
			}(i)
		}
		wg.Wait()

		// Each process should be assigned exactly once
		assert.Equal(t, int64(operationsPerWorker), successCount)

		// Verify no process is assigned to multiple executors
		for i := 0; i < operationsPerWorker; i++ {
			doc, err := db.Get(ctx, "cas_test", fmt.Sprintf("process_%d", i))
			assert.NoError(t, err)
			assert.Equal(t, "running", doc.Fields["state"])
			assert.NotNil(t, doc.Fields["executor_id"])
		}
	})
}

func TestVelocityDB_Persistence(t *testing.T) {
	tempDir := t.TempDir()

	// Create database with persistence
	config := &VelocityConfig{
		DataDir:   tempDir,
		CacheSize: 10,
		InMemory:  false, // Enable persistence
	}

	db1, err := NewVelocityDB(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Insert some data
	doc := &VelocityDocument{
		ID: "persistent_test",
		Fields: map[string]interface{}{
			"name":    "Persistent Data",
			"value":   12345,
			"created": time.Now().Unix(),
		},
	}

	err = db1.Insert(ctx, "persistence_test", doc)
	require.NoError(t, err)

	// Close database
	err = db1.Close()
	require.NoError(t, err)

	// Reopen database
	db2, err := NewVelocityDB(config)
	require.NoError(t, err)
	defer db2.Close()

	// Verify data persisted
	retrieved, err := db2.Get(ctx, "persistence_test", "persistent_test")
	assert.NoError(t, err)
	assert.Equal(t, "persistent_test", retrieved.ID)
	assert.Equal(t, "Persistent Data", retrieved.Fields["name"])
	assert.Equal(t, float64(12345), retrieved.Fields["value"])
}

func TestVelocityDB_Health(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	ctx := context.Background()

	// Test health check on healthy database
	err := db.Health(ctx)
	assert.NoError(t, err)

	// Close database and test health
	db.Close()
	err = db.Health(ctx)
	assert.Error(t, err)
}

func TestVelocityDB_ErrorHandling(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	ctx := context.Background()

	// Test operations on non-existent documents
	_, err := db.Get(ctx, "nonexistent", "missing")
	assert.Error(t, err)

	_, err = db.Update(ctx, "nonexistent", "missing", map[string]interface{}{"x": 1})
	assert.Error(t, err)

	err = db.Delete(ctx, "nonexistent", "missing")
	assert.Error(t, err)

	// Test invalid CAS operations
	cas := &VelocityCASRequest{
		Key:      "test",
		Expected: "invalid_type", // Should be map or nil
		Value:    map[string]interface{}{"x": 1},
	}
	result, err := db.CompareAndSwap(ctx, "test", cas)
	// Should handle gracefully
	if err != nil {
		assert.Contains(t, err.Error(), "must be a map")
	} else {
		assert.False(t, result.Success)
	}
}

// Helper functions
func setupTestDB(t *testing.T) *VelocityDB {
	config := &VelocityConfig{
		DataDir:   t.TempDir(),
		CacheSize: 10, // Small cache for testing
		InMemory:  true,
	}

	db, err := NewVelocityDB(config)
	require.NoError(t, err)
	return db
}

func cleanupTestDB(db *VelocityDB) {
	if db != nil {
		db.Close()
	}
}