package memdb

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func BenchmarkVelocityDB_Insert(b *testing.B) {
	db := setupBenchDB(b)
	defer cleanupTestDB(db)

	ctx := context.Background()
	docs := make([]*VelocityDocument, b.N)

	// Pre-generate documents
	for i := 0; i < b.N; i++ {
		docs[i] = &VelocityDocument{
			ID: fmt.Sprintf("bench_insert_%d", i),
			Fields: map[string]interface{}{
				"name":      fmt.Sprintf("User %d", i),
				"email":     fmt.Sprintf("user%d@example.com", i),
				"age":       20 + (i % 50),
				"active":    i%2 == 0,
				"timestamp": time.Now().Unix(),
			},
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := db.Insert(ctx, "bench_users", docs[i])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVelocityDB_Get(b *testing.B) {
	db := setupBenchDB(b)
	defer cleanupTestDB(db)

	ctx := context.Background()

	// Pre-insert documents
	numDocs := 10000
	for i := 0; i < numDocs; i++ {
		doc := &VelocityDocument{
			ID: fmt.Sprintf("bench_get_%d", i),
			Fields: map[string]interface{}{
				"name":  fmt.Sprintf("User %d", i),
				"value": i * 100,
			},
		}
		err := db.Insert(ctx, "bench_users", doc)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("bench_get_%d", i%numDocs)
		_, err := db.Get(ctx, "bench_users", id)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVelocityDB_Update(b *testing.B) {
	db := setupBenchDB(b)
	defer cleanupTestDB(db)

	ctx := context.Background()

	// Pre-insert documents
	numDocs := 1000
	for i := 0; i < numDocs; i++ {
		doc := &VelocityDocument{
			ID: fmt.Sprintf("bench_update_%d", i),
			Fields: map[string]interface{}{
				"name":    fmt.Sprintf("User %d", i),
				"counter": 0,
			},
		}
		err := db.Insert(ctx, "bench_users", doc)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("bench_update_%d", i%numDocs)
		fields := map[string]interface{}{
			"counter":      i,
			"last_updated": time.Now().Unix(),
		}
		_, err := db.Update(ctx, "bench_users", id, fields)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVelocityDB_CAS(b *testing.B) {
	db := setupBenchDB(b)
	defer cleanupTestDB(db)

	ctx := context.Background()

	// Pre-insert documents in "waiting" state
	numDocs := 1000
	for i := 0; i < numDocs; i++ {
		doc := &VelocityDocument{
			ID: fmt.Sprintf("bench_cas_%d", i),
			Fields: map[string]interface{}{
				"state": "waiting",
			},
		}
		err := db.Insert(ctx, "processes", doc)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	successCount := 0
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("bench_cas_%d", i%numDocs)
		
		cas := &VelocityCASRequest{
			Key:      id,
			Expected: map[string]interface{}{"state": "waiting"},
			Value: map[string]interface{}{
				"state":       "running",
				"executor_id": fmt.Sprintf("executor_%d", i),
				"assigned_at": time.Now().Unix(),
			},
		}

		result, err := db.CompareAndSwap(ctx, "processes", cas)
		if err != nil {
			b.Fatal(err)
		}
		if result.Success {
			successCount++
		}

		// Reset document for next iteration
		if i%numDocs == numDocs-1 {
			// Reset all documents back to waiting
			for j := 0; j < numDocs; j++ {
				resetID := fmt.Sprintf("bench_cas_%d", j)
				db.Update(ctx, "processes", resetID, map[string]interface{}{"state": "waiting"})
			}
		}
	}

	b.Logf("CAS success rate: %.2f%%", float64(successCount)/float64(b.N)*100)
}

func BenchmarkVelocityDB_List(b *testing.B) {
	db := setupBenchDB(b)
	defer cleanupTestDB(db)

	ctx := context.Background()

	// Pre-insert documents
	numDocs := 1000
	for i := 0; i < numDocs; i++ {
		doc := &VelocityDocument{
			ID: fmt.Sprintf("bench_list_%d", i),
			Fields: map[string]interface{}{
				"name":  fmt.Sprintf("User %d", i),
				"index": i,
			},
		}
		err := db.Insert(ctx, "bench_users", doc)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := db.List(ctx, "bench_users", 100, i%10*100)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVelocityDB_ConcurrentReads(b *testing.B) {
	db := setupBenchDB(b)
	defer cleanupTestDB(db)

	ctx := context.Background()

	// Pre-insert documents
	numDocs := 1000
	for i := 0; i < numDocs; i++ {
		doc := &VelocityDocument{
			ID: fmt.Sprintf("concurrent_read_%d", i),
			Fields: map[string]interface{}{
				"name":  fmt.Sprintf("User %d", i),
				"value": i * 100,
			},
		}
		err := db.Insert(ctx, "bench_users", doc)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			id := fmt.Sprintf("concurrent_read_%d", counter%numDocs)
			_, err := db.Get(ctx, "bench_users", id)
			if err != nil {
				b.Error(err)
			}
			counter++
		}
	})
}

func BenchmarkVelocityDB_ConcurrentWrites(b *testing.B) {
	db := setupBenchDB(b)
	defer cleanupTestDB(db)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	var counter int64
	var mutex sync.Mutex

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mutex.Lock()
			counter++
			localCounter := counter
			mutex.Unlock()

			doc := &VelocityDocument{
				ID: fmt.Sprintf("concurrent_write_%d", localCounter),
				Fields: map[string]interface{}{
					"name":      fmt.Sprintf("User %d", localCounter),
					"timestamp": time.Now().Unix(),
				},
			}

			err := db.Insert(ctx, "concurrent_writes", doc)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

func BenchmarkVelocityDB_MixedWorkload(b *testing.B) {
	db := setupBenchDB(b)
	defer cleanupTestDB(db)

	ctx := context.Background()

	// Pre-insert some documents
	numDocs := 1000
	for i := 0; i < numDocs; i++ {
		doc := &VelocityDocument{
			ID: fmt.Sprintf("mixed_%d", i),
			Fields: map[string]interface{}{
				"name":    fmt.Sprintf("User %d", i),
				"counter": 0,
			},
		}
		err := db.Insert(ctx, "mixed_workload", doc)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		switch i % 10 {
		case 0, 1, 2, 3, 4, 5, 6: // 70% reads
			id := fmt.Sprintf("mixed_%d", i%numDocs)
			_, err := db.Get(ctx, "mixed_workload", id)
			if err != nil {
				b.Fatal(err)
			}

		case 7, 8: // 20% updates
			id := fmt.Sprintf("mixed_%d", i%numDocs)
			fields := map[string]interface{}{
				"counter": i,
				"updated": time.Now().Unix(),
			}
			_, err := db.Update(ctx, "mixed_workload", id, fields)
			if err != nil {
				b.Fatal(err)
			}

		case 9: // 10% inserts
			doc := &VelocityDocument{
				ID: fmt.Sprintf("mixed_new_%d", i),
				Fields: map[string]interface{}{
					"name": fmt.Sprintf("New User %d", i),
				},
			}
			err := db.Insert(ctx, "mixed_workload", doc)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// Memory usage benchmark
func BenchmarkVelocityDB_MemoryUsage(b *testing.B) {
	db := setupBenchDB(b)
	defer cleanupTestDB(db)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	// Insert documents and measure memory
	for i := 0; i < b.N; i++ {
		doc := &VelocityDocument{
			ID: fmt.Sprintf("memory_test_%d", i),
			Fields: map[string]interface{}{
				"name":        fmt.Sprintf("User %d", i),
				"email":       fmt.Sprintf("user%d@example.com", i),
				"age":         20 + (i % 50),
				"active":      i%2 == 0,
				"metadata":    map[string]interface{}{"key1": "value1", "key2": "value2"},
				"tags":        []string{"tag1", "tag2", "tag3"},
				"timestamp":   time.Now().Unix(),
				"description": "This is a test document with some content to measure memory usage",
			},
		}

		err := db.Insert(ctx, "memory_test", doc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Cache performance benchmark
func BenchmarkVelocityDB_CachePerformance(b *testing.B) {
	db := setupBenchDB(b)
	defer cleanupTestDB(db)

	ctx := context.Background()

	// Insert documents
	numDocs := 100
	for i := 0; i < numDocs; i++ {
		doc := &VelocityDocument{
			ID: fmt.Sprintf("cache_test_%d", i),
			Fields: map[string]interface{}{
				"name":  fmt.Sprintf("User %d", i),
				"value": i * 100,
			},
		}
		err := db.Insert(ctx, "cache_test", doc)
		if err != nil {
			b.Fatal(err)
		}
	}

	// First pass - populate cache
	for i := 0; i < numDocs; i++ {
		id := fmt.Sprintf("cache_test_%d", i)
		_, err := db.Get(ctx, "cache_test", id)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	// Second pass - should hit cache
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("cache_test_%d", i%numDocs)
		_, err := db.Get(ctx, "cache_test", id)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func setupBenchDB(b *testing.B) *VelocityDB {
	config := &VelocityConfig{
		DataDir:   b.TempDir(),
		CacheSize: 100, // Larger cache for benchmarks
		InMemory:  true,
	}

	db, err := NewVelocityDB(config)
	if err != nil {
		b.Fatal(err)
	}
	return db
}