package diskstore

import (
	"fmt"
	"testing"
)

type BenchDiskRecord struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"value"`
	Data  string `json:"data"`
}

func BenchmarkDiskStoreWrite(b *testing.B) {
	dir := b.TempDir()
	ds, err := NewDiskStore[BenchDiskRecord](dir, "bench")
	if err != nil {
		b.Fatal(err)
	}

	rec := BenchDiskRecord{ID: "r1", Name: "test", Value: 42, Data: "some data payload for benchmarking"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("r-%d", i)
		if err := ds.Write(key, rec); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDiskStoreRead(b *testing.B) {
	dir := b.TempDir()
	ds, err := NewDiskStore[BenchDiskRecord](dir, "bench")
	if err != nil {
		b.Fatal(err)
	}

	// Write records first
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("r-%d", i)
		ds.Write(key, BenchDiskRecord{ID: key, Name: "test", Value: i})
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("r-%d", i%1000)
		if _, err := ds.Read(key); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDiskStoreScan(b *testing.B) {
	for _, size := range []int{100, 1000} {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			dir := b.TempDir()
			ds, err := NewDiskStore[BenchDiskRecord](dir, "bench")
			if err != nil {
				b.Fatal(err)
			}

			for i := 0; i < size; i++ {
				key := fmt.Sprintf("r-%d", i)
				ds.Write(key, BenchDiskRecord{ID: key, Name: "test", Value: i})
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ds.Scan(func(key string, value BenchDiskRecord) error {
					return nil
				})
			}
		})
	}
}

func BenchmarkSanitizeKey(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sanitizeKey("abc123-def456-ghi789")
	}
}

func BenchmarkSanitizeKeySpecial(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sanitizeKey("path/to:file*name?with<special>chars|here")
	}
}
