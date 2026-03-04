package diskstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// DiskStore is a generic file-per-record store on disk.
// Each record is stored as a JSON file keyed by primary key.
type DiskStore[V any] struct {
	mu      sync.RWMutex
	baseDir string
}

// NewDiskStore creates a new DiskStore. The baseDir/entityName directory is
// created if it does not exist.
func NewDiskStore[V any](baseDir string, entityName string) (*DiskStore[V], error) {
	dir := filepath.Join(baseDir, entityName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return &DiskStore[V]{baseDir: dir}, nil
}

var keySanitizer = strings.NewReplacer(
	"/", "__SLASH__",
	"\\", "__BSLASH__",
	":", "__COLON__",
	"*", "__STAR__",
	"?", "__QMARK__",
	"\"", "__QUOTE__",
	"<", "__LT__",
	">", "__GT__",
	"|", "__PIPE__",
)

var keyDesanitizer = strings.NewReplacer(
	"__SLASH__", "/",
	"__BSLASH__", "\\",
	"__COLON__", ":",
	"__STAR__", "*",
	"__QMARK__", "?",
	"__QUOTE__", "\"",
	"__LT__", "<",
	"__GT__", ">",
	"__PIPE__", "|",
)

// sanitizeKey replaces characters that are unsafe for filenames.
func sanitizeKey(key string) string {
	return keySanitizer.Replace(key)
}

// desanitizeKey reverses sanitizeKey.
func desanitizeKey(name string) string {
	return keyDesanitizer.Replace(name)
}

func (d *DiskStore[V]) filePath(key string) string {
	return filepath.Join(d.baseDir, sanitizeKey(key)+".json")
}

// Write marshals the value to JSON and writes it to disk.
// The write is atomic: it writes to a temp file then renames.
func (d *DiskStore[V]) Write(key string, value V) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal record %s: %w", key, err)
	}

	target := d.filePath(key)

	// Write to temp file first for atomicity
	tmp, err := os.CreateTemp(d.baseDir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpName, target); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to rename temp file to %s: %w", target, err)
	}

	return nil
}

// Read reads a record from disk and unmarshals it.
func (d *DiskStore[V]) Read(key string) (V, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var zero V
	data, err := os.ReadFile(d.filePath(key))
	if err != nil {
		if os.IsNotExist(err) {
			return zero, fmt.Errorf("record not found: %s", key)
		}
		return zero, fmt.Errorf("failed to read record %s: %w", key, err)
	}

	var value V
	if err := json.Unmarshal(data, &value); err != nil {
		return zero, fmt.Errorf("failed to unmarshal record %s: %w", key, err)
	}

	return value, nil
}

// Delete removes a record from disk.
func (d *DiskStore[V]) Delete(key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	err := os.Remove(d.filePath(key))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("record not found: %s", key)
		}
		return fmt.Errorf("failed to delete record %s: %w", key, err)
	}

	return nil
}

// List returns all keys in the store, sorted alphabetically.
func (d *DiskStore[V]) List() ([]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	entries, err := os.ReadDir(d.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", d.baseDir, err)
	}

	var keys []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		if strings.HasPrefix(name, ".tmp-") {
			continue
		}
		key := desanitizeKey(strings.TrimSuffix(name, ".json"))
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys, nil
}

// Scan iterates over all records in sorted key order, calling fn for each.
// If fn returns an error, iteration stops and the error is returned.
func (d *DiskStore[V]) Scan(fn func(key string, value V) error) error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	entries, err := os.ReadDir(d.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", d.baseDir, err)
	}

	keys := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") || strings.HasPrefix(name, ".tmp-") {
			continue
		}
		keys = append(keys, desanitizeKey(strings.TrimSuffix(name, ".json")))
	}
	sort.Strings(keys)

	for _, key := range keys {
		data, err := os.ReadFile(d.filePath(key))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to read record %s: %w", key, err)
		}
		var value V
		if err := json.Unmarshal(data, &value); err != nil {
			return fmt.Errorf("failed to unmarshal record %s: %w", key, err)
		}
		if err := fn(key, value); err != nil {
			return err
		}
	}

	return nil
}
