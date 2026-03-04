package store

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/colonyos/colonies/pkg/database/embedded/diskstore"
	"github.com/colonyos/colonies/pkg/database/embedded/wal"
)

type BenchRecord struct {
	ID       string `json:"id"`
	Colony   string `json:"colony"`
	State    int    `json:"state"`
	Priority int64  `json:"priority"`
	Value    int    `json:"value"`
}

func benchIdentity(s string) string { return s }

func newBenchStore(b *testing.B, withWAL bool, withDisk bool) *Store[string, BenchRecord] {
	cfg := Config[string, BenchRecord]{
		EntityName: "bench",
		KeyToStr:   benchIdentity,
		StrToKey:   benchIdentity,
	}

	if withWAL {
		dir := b.TempDir()
		w, err := wal.NewFileWAL(filepath.Join(dir, "wal"), wal.SyncNone)
		if err != nil {
			b.Fatal(err)
		}
		cfg.WAL = w
		b.Cleanup(func() { w.Close() })
	}

	if withDisk {
		dir := b.TempDir()
		ds, err := diskstore.NewDiskStore[BenchRecord](filepath.Join(dir, "data"), "bench")
		if err != nil {
			b.Fatal(err)
		}
		cfg.Disk = ds
	}

	return NewStore[string, BenchRecord](cfg)
}

func benchRecord(id string, i int) *BenchRecord {
	return &BenchRecord{
		ID:       id,
		Colony:   fmt.Sprintf("colony-%d", i%10),
		State:    i % 3,
		Priority: int64(i),
		Value:    i,
	}
}

// --- Put benchmarks ---

func BenchmarkStorePutInMemory(b *testing.B) {
	s := newBenchStore(b, false, false)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("r-%d", i)
		s.Put(id, benchRecord(id, i))
	}
}

func BenchmarkStorePutWithWAL(b *testing.B) {
	s := newBenchStore(b, true, false)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("r-%d", i)
		s.Put(id, benchRecord(id, i))
	}
}

// --- Get benchmarks ---

func BenchmarkStoreGet(b *testing.B) {
	s := newBenchStore(b, false, false)
	for i := 0; i < 10000; i++ {
		id := fmt.Sprintf("r-%d", i)
		s.Put(id, benchRecord(id, i))
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Get(fmt.Sprintf("r-%d", i%10000))
	}
}

func BenchmarkStoreGetMiss(b *testing.B) {
	s := newBenchStore(b, false, false)
	for i := 0; i < 10000; i++ {
		id := fmt.Sprintf("r-%d", i)
		s.Put(id, benchRecord(id, i))
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Get(fmt.Sprintf("miss-%d", i))
	}
}

// --- Filter benchmarks ---

func BenchmarkStoreFilter(b *testing.B) {
	for _, size := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			s := newBenchStore(b, false, false)
			for i := 0; i < size; i++ {
				id := fmt.Sprintf("r-%d", i)
				s.Put(id, benchRecord(id, i))
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				s.Filter(func(r *BenchRecord) bool {
					return r.Colony == "colony-0" && r.State == 0
				})
			}
		})
	}
}

// --- Count benchmarks ---

func BenchmarkStoreCount(b *testing.B) {
	s := newBenchStore(b, false, false)
	for i := 0; i < 10000; i++ {
		id := fmt.Sprintf("r-%d", i)
		s.Put(id, benchRecord(id, i))
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Count(func(r *BenchRecord) bool {
			return r.Colony == "colony-0"
		})
	}
}

// --- FlushDirty benchmarks ---

func BenchmarkStoreFlushDirty(b *testing.B) {
	for _, dirty := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("dirty_%d", dirty), func(b *testing.B) {
			s := newBenchStore(b, false, true)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				for j := 0; j < dirty; j++ {
					id := fmt.Sprintf("r-%d", j)
					s.Put(id, benchRecord(id, j))
				}
				b.StartTimer()
				s.FlushDirty()
			}
		})
	}
}

// --- Delete benchmarks ---

func BenchmarkStoreDelete(b *testing.B) {
	s := newBenchStore(b, false, false)
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("r-%d", i)
		s.Put(id, benchRecord(id, i))
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Delete(fmt.Sprintf("r-%d", i))
	}
}

// --- All benchmarks ---

func BenchmarkStoreAll(b *testing.B) {
	for _, size := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			s := newBenchStore(b, false, false)
			for i := 0; i < size; i++ {
				id := fmt.Sprintf("r-%d", i)
				s.Put(id, benchRecord(id, i))
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = s.All()
			}
		})
	}
}

// --- Realistic scheduler hot path benchmark ---

func BenchmarkStoreSchedulerHotPath(b *testing.B) {
	s := newBenchStore(b, false, false)
	// Populate with 10k records across 10 colonies, 3 states
	for i := 0; i < 10000; i++ {
		id := fmt.Sprintf("r-%d", i)
		s.Put(id, benchRecord(id, i))
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Simulate: find all waiting processes for colony-0
		s.Filter(func(r *BenchRecord) bool {
			return r.Colony == "colony-0" && r.State == 0
		})
	}
}
