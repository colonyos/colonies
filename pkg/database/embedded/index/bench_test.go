package index

import (
	"fmt"
	"testing"
)

// --- MapIndex benchmarks ---

func BenchmarkMapIndexAdd(b *testing.B) {
	idx := NewMapIndex[string, string]()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		idx.Add(fmt.Sprintf("key-%d", i), fmt.Sprintf("sec-%d", i%100))
	}
}

func BenchmarkMapIndexLookup(b *testing.B) {
	idx := NewMapIndex[string, string]()
	for i := 0; i < 10000; i++ {
		idx.Add(fmt.Sprintf("key-%d", i), fmt.Sprintf("sec-%d", i%100))
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = idx.Lookup(fmt.Sprintf("sec-%d", i%100))
	}
}

func BenchmarkMapIndexRemove(b *testing.B) {
	idx := NewMapIndex[string, string]()
	for i := 0; i < b.N; i++ {
		idx.Add(fmt.Sprintf("key-%d", i), "sec-0")
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		idx.Remove(fmt.Sprintf("key-%d", i), "sec-0")
	}
}

// --- OrderedIndex benchmarks ---

func BenchmarkOrderedIndexAdd(b *testing.B) {
	idx := NewStringOrderedIndex()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		idx.Add(IndexEntry[string]{
			SortKey:    int64(i),
			PrimaryKey: fmt.Sprintf("key-%d", i),
		})
	}
}

func BenchmarkOrderedIndexRemove(b *testing.B) {
	idx := NewStringOrderedIndex()
	entries := make([]IndexEntry[string], b.N)
	for i := 0; i < b.N; i++ {
		entries[i] = IndexEntry[string]{
			SortKey:    int64(i),
			PrimaryKey: fmt.Sprintf("key-%d", i),
		}
		idx.Add(entries[i])
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		idx.Remove(entries[i])
	}
}

func BenchmarkOrderedIndexAscendFirst(b *testing.B) {
	for _, size := range []int{100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			idx := NewStringOrderedIndex()
			for i := 0; i < size; i++ {
				idx.Add(IndexEntry[string]{
					SortKey:    int64(i),
					PrimaryKey: fmt.Sprintf("key-%d", i),
				})
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				idx.AscendFirst(1, func(e IndexEntry[string]) bool {
					return true
				})
			}
		})
	}
}

func BenchmarkOrderedIndexAscendFirst10(b *testing.B) {
	idx := NewStringOrderedIndex()
	for i := 0; i < 100000; i++ {
		idx.Add(IndexEntry[string]{
			SortKey:    int64(i),
			PrimaryKey: fmt.Sprintf("key-%d", i),
		})
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		idx.AscendFirst(10, func(e IndexEntry[string]) bool {
			return true
		})
	}
}

// --- CompoundIndex benchmarks ---

func BenchmarkCompoundIndexAdd(b *testing.B) {
	idx := NewCompoundIndex()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		idx.Add(
			fmt.Sprintf("colony-%d", i%10),
			i%3,
			IndexEntry[string]{
				SortKey:    int64(i),
				PrimaryKey: fmt.Sprintf("key-%d", i),
			},
		)
	}
}

func BenchmarkCompoundIndexAscendFirst1(b *testing.B) {
	idx := NewCompoundIndex()
	for i := 0; i < 100000; i++ {
		idx.Add(
			fmt.Sprintf("colony-%d", i%10),
			i%3,
			IndexEntry[string]{
				SortKey:    int64(i),
				PrimaryKey: fmt.Sprintf("key-%d", i),
			},
		)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		idx.AscendFirst("colony-0", 0, 1, func(e IndexEntry[string]) bool {
			return true
		})
	}
}

func BenchmarkCompoundIndexCount(b *testing.B) {
	idx := NewCompoundIndex()
	for i := 0; i < 100000; i++ {
		idx.Add(
			fmt.Sprintf("colony-%d", i%10),
			i%3,
			IndexEntry[string]{
				SortKey:    int64(i),
				PrimaryKey: fmt.Sprintf("key-%d", i),
			},
		)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = idx.Count("colony-0", 0)
	}
}

func BenchmarkCompoundIndexCountAll(b *testing.B) {
	idx := NewCompoundIndex()
	for i := 0; i < 100000; i++ {
		idx.Add(
			fmt.Sprintf("colony-%d", i%10),
			i%3,
			IndexEntry[string]{
				SortKey:    int64(i),
				PrimaryKey: fmt.Sprintf("key-%d", i),
			},
		)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = idx.CountAll("colony-0")
	}
}

// --- Memory usage benchmarks ---

func BenchmarkMapIndexMemory10k(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		idx := NewMapIndex[string, string]()
		for j := 0; j < 10000; j++ {
			idx.Add(fmt.Sprintf("key-%d", j), fmt.Sprintf("sec-%d", j%100))
		}
	}
}

func BenchmarkOrderedIndexMemory10k(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		idx := NewStringOrderedIndex()
		for j := 0; j < 10000; j++ {
			idx.Add(IndexEntry[string]{
				SortKey:    int64(j),
				PrimaryKey: fmt.Sprintf("key-%d", j),
			})
		}
	}
}
