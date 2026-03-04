package embedded

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/colonyos/colonies/pkg/database/embedded/index"
	"github.com/colonyos/colonies/pkg/database/embedded/store"
	"github.com/colonyos/colonies/pkg/database/embedded/wal"
)

type BenchProcess struct {
	ID       string `json:"id"`
	Colony   string `json:"colony"`
	State    int    `json:"state"`
	Priority int64  `json:"priority"`
	Value    int    `json:"value"`
}

func benchIdentity(s string) string { return s }

// BenchmarkE2ESchedulerAssign simulates the scheduler hot path:
// insert processes, then find the highest-priority waiting process for a colony.
func BenchmarkE2ESchedulerAssign(b *testing.B) {
	dir := b.TempDir()
	w, err := wal.NewFileWAL(filepath.Join(dir, "wal"), wal.SyncNone)
	if err != nil {
		b.Fatal(err)
	}
	defer w.Close()

	s := store.NewStore[string, BenchProcess](store.Config[string, BenchProcess]{
		WAL:        w,
		EntityName: "processes",
		KeyToStr:   benchIdentity,
		StrToKey:   benchIdentity,
	})

	stateIdx := index.NewCompoundIndex()

	// Populate with 10k processes across 10 colonies
	for i := 0; i < 10000; i++ {
		id := fmt.Sprintf("p-%d", i)
		colony := fmt.Sprintf("colony-%d", i%10)
		state := 0 // all waiting
		p := &BenchProcess{
			ID:       id,
			Colony:   colony,
			State:    state,
			Priority: int64(i),
		}
		s.Put(id, p)
		stateIdx.Add(colony, state, index.IndexEntry[string]{
			SortKey:    int64(i),
			PrimaryKey: id,
		})
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Find first waiting process for colony-0
		stateIdx.AscendFirst("colony-0", 0, 1, func(e index.IndexEntry[string]) bool {
			s.Get(e.PrimaryKey)
			return true
		})
	}
}

// BenchmarkE2EPutAndIndex simulates inserting a process with index maintenance.
func BenchmarkE2EPutAndIndex(b *testing.B) {
	s := store.NewStore[string, BenchProcess](store.Config[string, BenchProcess]{
		EntityName: "processes",
		KeyToStr:   benchIdentity,
		StrToKey:   benchIdentity,
	})

	colonyIdx := index.NewMapIndex[string, string]()
	stateIdx := index.NewCompoundIndex()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("p-%d", i)
		colony := fmt.Sprintf("colony-%d", i%10)
		p := &BenchProcess{
			ID:       id,
			Colony:   colony,
			State:    0,
			Priority: int64(i),
		}
		s.Put(id, p)
		colonyIdx.Add(id, colony)
		stateIdx.Add(colony, 0, index.IndexEntry[string]{
			SortKey:    int64(i),
			PrimaryKey: id,
		})
	}
}

// BenchmarkE2EDeleteAndIndex simulates removing a process with index cleanup.
func BenchmarkE2EDeleteAndIndex(b *testing.B) {
	s := store.NewStore[string, BenchProcess](store.Config[string, BenchProcess]{
		EntityName: "processes",
		KeyToStr:   benchIdentity,
		StrToKey:   benchIdentity,
	})

	colonyIdx := index.NewMapIndex[string, string]()
	stateIdx := index.NewCompoundIndex()

	// Pre-populate
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("p-%d", i)
		colony := fmt.Sprintf("colony-%d", i%10)
		p := &BenchProcess{
			ID:       id,
			Colony:   colony,
			State:    0,
			Priority: int64(i),
		}
		s.Put(id, p)
		colonyIdx.Add(id, colony)
		stateIdx.Add(colony, 0, index.IndexEntry[string]{
			SortKey:    int64(i),
			PrimaryKey: id,
		})
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("p-%d", i)
		p, ok := s.Get(id)
		if !ok {
			continue
		}
		colonyIdx.Remove(id, p.Colony)
		stateIdx.Remove(p.Colony, p.State, index.IndexEntry[string]{
			SortKey:    p.Priority,
			PrimaryKey: id,
		})
		s.Delete(id)
	}
}
