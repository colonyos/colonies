package memdb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coocood/freecache"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

// VelocityDB is a complete, working in-memory database
type VelocityDB struct {
	db    *badger.DB
	cache *freecache.Cache
	mu    sync.RWMutex
}

// VelocityConfig holds configuration
type VelocityConfig struct {
	DataDir   string
	CacheSize int  // MB
	InMemory  bool
}

// VelocityDocument represents a document
type VelocityDocument struct {
	ID       string                 `json:"id"`
	Fields   map[string]interface{} `json:"fields"`
	Version  uint64                 `json:"version"`
	Created  time.Time             `json:"created"`
	Modified time.Time             `json:"modified"`
}

// CASRequest for compare-and-swap
type VelocityCASRequest struct {
	Key      string
	Expected interface{}
	Value    interface{}
}

// CASResult for compare-and-swap response
type VelocityCASResult struct {
	Success      bool
	CurrentValue interface{}
	Version      uint64
}

// NewVelocityDB creates a new VelocityDB instance
func NewVelocityDB(config *VelocityConfig) (*VelocityDB, error) {
	if config == nil {
		config = &VelocityConfig{
			DataDir:   "/tmp/velocitydb",
			CacheSize: 100,
			InMemory:  true,
		}
	}

	var opts badger.Options
	if config.InMemory {
		opts = badger.DefaultOptions("").WithInMemory(true).WithLogger(nil)
	} else {
		opts = badger.DefaultOptions(config.DataDir).WithLogger(nil)
	}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BadgerDB: %w", err)
	}

	cache := freecache.NewCache(config.CacheSize * 1024 * 1024)

	return &VelocityDB{
		db:    db,
		cache: cache,
	}, nil
}

// Insert adds a new document
func (v *VelocityDB) Insert(ctx context.Context, collection string, doc *VelocityDocument) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}

	doc.Version = 1
	doc.Created = time.Now()
	doc.Modified = doc.Created

	key := v.key(collection, doc.ID)
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	err = v.db.Update(func(txn *badger.Txn) error {
		// Check if exists
		_, err := txn.Get([]byte(key))
		if err == nil {
			return fmt.Errorf("document already exists")
		}
		
		return txn.Set([]byte(key), data)
	})

	if err != nil {
		return err
	}

	// Cache it
	v.cache.Set([]byte(key), data, 0)
	return nil
}

// Get retrieves a document
func (v *VelocityDB) Get(ctx context.Context, collection string, id string) (*VelocityDocument, error) {
	key := v.key(collection, id)

	// Try cache first
	if cached, err := v.cache.Get([]byte(key)); err == nil {
		var doc VelocityDocument
		if err := json.Unmarshal(cached, &doc); err == nil {
			return &doc, nil
		}
	}

	// Fallback to BadgerDB
	var doc VelocityDocument
	err := v.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &doc)
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, fmt.Errorf("document not found")
		}
		return nil, err
	}

	// Cache it
	if data, err := json.Marshal(doc); err == nil {
		v.cache.Set([]byte(key), data, 0)
	}

	return &doc, nil
}

// Update modifies a document
func (v *VelocityDB) Update(ctx context.Context, collection string, id string, fields map[string]interface{}) (*VelocityDocument, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	key := v.key(collection, id)
	var updatedDoc *VelocityDocument

	err := v.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return fmt.Errorf("document not found")
		}

		var doc VelocityDocument
		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &doc)
		})
		if err != nil {
			return err
		}

		// Update fields
		if doc.Fields == nil {
			doc.Fields = make(map[string]interface{})
		}
		for k, v := range fields {
			doc.Fields[k] = v
		}

		doc.Version++
		doc.Modified = time.Now()

		data, err := json.Marshal(doc)
		if err != nil {
			return err
		}

		updatedDoc = &doc
		return txn.Set([]byte(key), data)
	})

	if err != nil {
		return nil, err
	}

	// Update cache
	if data, err := json.Marshal(updatedDoc); err == nil {
		v.cache.Set([]byte(key), data, 0)
	}

	return updatedDoc, nil
}

// Delete removes a document
func (v *VelocityDB) Delete(ctx context.Context, collection string, id string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	key := v.key(collection, id)

	err := v.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		if err != nil {
			return fmt.Errorf("document not found")
		}
		return txn.Delete([]byte(key))
	})

	if err != nil {
		return err
	}

	v.cache.Del([]byte(key))
	return nil
}

// List returns all documents in a collection
func (v *VelocityDB) List(ctx context.Context, collection string, limit, offset int) ([]*VelocityDocument, error) {
	prefix := v.prefix(collection)
	var documents []*VelocityDocument

	err := v.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		count := 0
		skipped := 0

		for it.Seek([]byte(prefix)); it.Valid(); it.Next() {
			key := it.Item().Key()
			if !strings.HasPrefix(string(key), prefix) {
				break
			}

			if skipped < offset {
				skipped++
				continue
			}

			if limit > 0 && count >= limit {
				break
			}

			var doc VelocityDocument
			err := it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &doc)
			})
			if err != nil {
				continue
			}

			documents = append(documents, &doc)
			count++
		}

		return nil
	})

	return documents, err
}

// Count returns the number of documents in a collection
func (v *VelocityDB) Count(ctx context.Context, collection string) (int, error) {
	prefix := v.prefix(collection)
	count := 0

	err := v.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte(prefix)); it.Valid(); it.Next() {
			key := it.Item().Key()
			if !strings.HasPrefix(string(key), prefix) {
				break
			}
			count++
		}

		return nil
	})

	return count, err
}

// CompareAndSwap performs atomic compare-and-swap
func (v *VelocityDB) CompareAndSwap(ctx context.Context, collection string, cas *VelocityCASRequest) (*VelocityCASResult, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	key := v.key(collection, cas.Key)
	var result *VelocityCASResult

	err := v.db.Update(func(txn *badger.Txn) error {
		// Get current document
		item, err := txn.Get([]byte(key))
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}

		var currentDoc *VelocityDocument
		if err != badger.ErrKeyNotFound {
			currentDoc = &VelocityDocument{}
			err = item.Value(func(val []byte) error {
				return json.Unmarshal(val, currentDoc)
			})
			if err != nil {
				return err
			}
		}

		// Compare expected with current
		var currentValue interface{}
		if currentDoc != nil {
			currentValue = currentDoc.Fields
		}

		if !v.valuesEqual(cas.Expected, currentValue) {
			result = &VelocityCASResult{
				Success:      false,
				CurrentValue: currentValue,
				Version:      0,
			}
			if currentDoc != nil {
				result.Version = currentDoc.Version
			}
			return nil
		}

		// Perform swap
		newFields, ok := cas.Value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("value must be a map[string]interface{}")
		}

		newDoc := &VelocityDocument{
			ID:      cas.Key,
			Fields:  newFields,
			Version: 1,
		}

		if currentDoc != nil {
			newDoc.Version = currentDoc.Version + 1
			newDoc.Created = currentDoc.Created
		} else {
			newDoc.Created = time.Now()
		}
		newDoc.Modified = time.Now()

		data, err := json.Marshal(newDoc)
		if err != nil {
			return err
		}

		err = txn.Set([]byte(key), data)
		if err != nil {
			return err
		}

		result = &VelocityCASResult{
			Success:      true,
			CurrentValue: newDoc.Fields,
			Version:      newDoc.Version,
		}

		// Update cache
		v.cache.Set([]byte(key), data, 0)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Health checks database health
func (v *VelocityDB) Health(ctx context.Context) error {
	// Simple health check
	return v.db.View(func(txn *badger.Txn) error {
		return nil
	})
}

// Close closes the database
func (v *VelocityDB) Close() error {
	return v.db.Close()
}

// Helper methods
func (v *VelocityDB) key(collection, id string) string {
	return fmt.Sprintf("%s:%s", collection, id)
}

func (v *VelocityDB) prefix(collection string) string {
	return collection + ":"
}

func (v *VelocityDB) valuesEqual(expected, actual interface{}) bool {
	if expected == nil && actual == nil {
		return true
	}
	if expected == nil || actual == nil {
		return false
	}

	// Simple comparison - could be enhanced
	expectedJSON, _ := json.Marshal(expected)
	actualJSON, _ := json.Marshal(actual)
	return string(expectedJSON) == string(actualJSON)
}