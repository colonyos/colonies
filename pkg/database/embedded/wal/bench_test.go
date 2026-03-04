package wal

import (
	"fmt"
	"path/filepath"
	"testing"
)

func BenchmarkWALAppendSyncNone(b *testing.B) {
	dir := b.TempDir()
	w, err := NewFileWAL(filepath.Join(dir, "wal"), SyncNone)
	if err != nil {
		b.Fatal(err)
	}
	defer w.Close()

	entry := Entry{
		Op:     OpPut,
		Entity: "processes",
		Key:    "abc123-def456-ghi789",
		Data:   make([]byte, 512),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := w.Append(entry); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWALAppendSyncAlways(b *testing.B) {
	dir := b.TempDir()
	w, err := NewFileWAL(filepath.Join(dir, "wal"), SyncAlways)
	if err != nil {
		b.Fatal(err)
	}
	defer w.Close()

	entry := Entry{
		Op:     OpPut,
		Entity: "processes",
		Key:    "abc123-def456-ghi789",
		Data:   make([]byte, 512),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := w.Append(entry); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWALAppendSmallEntry(b *testing.B) {
	dir := b.TempDir()
	w, err := NewFileWAL(filepath.Join(dir, "wal"), SyncNone)
	if err != nil {
		b.Fatal(err)
	}
	defer w.Close()

	entry := Entry{
		Op:     OpDelete,
		Entity: "proc",
		Key:    "k1",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := w.Append(entry); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWALAppendLargeEntry(b *testing.B) {
	dir := b.TempDir()
	w, err := NewFileWAL(filepath.Join(dir, "wal"), SyncNone)
	if err != nil {
		b.Fatal(err)
	}
	defer w.Close()

	entry := Entry{
		Op:     OpPut,
		Entity: "processes",
		Key:    "abc123-def456-ghi789-jkl012-mno345",
		Data:   make([]byte, 8192),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := w.Append(entry); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWALReplay(b *testing.B) {
	for _, count := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("entries_%d", count), func(b *testing.B) {
			dir := b.TempDir()
			w, err := NewFileWAL(filepath.Join(dir, "wal"), SyncNone)
			if err != nil {
				b.Fatal(err)
			}

			entry := Entry{
				Op:     OpPut,
				Entity: "processes",
				Key:    "abc123-def456-ghi789",
				Data:   make([]byte, 256),
			}
			for i := 0; i < count; i++ {
				if err := w.Append(entry); err != nil {
					b.Fatal(err)
				}
			}
			w.Close()

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				w2, err := NewFileWAL(filepath.Join(dir, "wal"), SyncNone)
				if err != nil {
					b.Fatal(err)
				}
				err = w2.Replay(func(e Entry) error { return nil })
				if err != nil {
					b.Fatal(err)
				}
				w2.Close()
			}
		})
	}
}

func BenchmarkEncodeEntry(b *testing.B) {
	entry := Entry{
		Op:     OpPut,
		Entity: "processes",
		Key:    "abc123-def456-ghi789",
		Data:   make([]byte, 512),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = encodeEntry(entry)
	}
}
