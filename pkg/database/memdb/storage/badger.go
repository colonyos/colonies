package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/colonyos/colonies/pkg/database/memdb/schema"
	"github.com/coocood/freecache"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

// BadgerStorage implements storage using BadgerDB with FreeCache
type BadgerStorage struct {
	db     *badger.DB
	cache  *freecache.Cache
	mu     sync.RWMutex
	config *BadgerConfig
}

// BadgerConfig holds configuration for BadgerDB storage
type BadgerConfig struct {
	DataDir     string
	CacheSize   int // Size in MB for FreeCache
	SyncWrites  bool
	InMemory    bool
	TTL         time.Duration
}

// Document represents a stored document
type Document struct {
	ID       string                 `json:"id"`
	Fields   map[string]interface{} `json:"fields"`
	Version  uint64                 `json:"version"`
	Created  time.Time             `json:"created"`
	Modified time.Time             `json:"modified"`
}

// NewBadgerStorage creates a new BadgerDB storage instance
func NewBadgerStorage(config *BadgerConfig) (*BadgerStorage, error) {
	opts := badger.DefaultOptions(config.DataDir).
		WithSyncWrites(config.SyncWrites).
		WithLogger(nil) // Disable badger logging

	if config.InMemory {
		opts = opts.WithInMemory(true)
	}

	db, err := badger.Open(opts)
	if err != nil {
		return &BadgerStorage{}, fmt.Errorf("failed to open BadgerDB: %w", err)
	}

	// Initialize FreeCache (default 100MB if not specified)
	cacheSize := config.CacheSize
	if cacheSize == 0 {
		cacheSize = 100
	}
	cache := freecache.NewCache(cacheSize * 1024 * 1024) // Convert MB to bytes

	storage := &BadgerStorage{
		db:     db,
		cache:  cache,
		config: config,
	}

	// Start background cleanup
	go storage.runGC()

	return storage, nil
}

// CreateCollection creates a new collection with schema
func (s *BadgerStorage) CreateCollection(ctx context.Context, name string, sch *schema.Schema) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	collectionKey := s.collectionKey(name)
	
	// Check if collection already exists
	err := s.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(collectionKey))
		if err == nil {
			return fmt.Errorf("collection '%s' already exists", name)
		}
		if err != badger.ErrKeyNotFound {
			return err
		}
		return nil
	})
	
	if err != nil {
		return err
	}

	// Store collection metadata
	metadata := map[string]interface{}{
		"name":    name,
		"schema":  sch,
		"created": time.Now(),
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal collection metadata: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(collectionKey), data)
	})
}

// DropCollection removes a collection and all its documents
func (s *BadgerStorage) DropCollection(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete all documents in the collection
	prefix := s.documentPrefix(name)
	
	err := s.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		var keysToDelete [][]byte
		
		// Collect all keys to delete
		for it.Seek([]byte(prefix)); it.Valid(); it.Next() {
			key := it.Item().Key()
			if !strings.HasPrefix(string(key), prefix) {
				break
			}
			keysToDelete = append(keysToDelete, append([]byte(nil), key...))
		}

		// Delete all collected keys
		for _, key := range keysToDelete {
			if err := txn.Delete(key); err != nil {
				return err
			}
		}

		// Delete collection metadata
		return txn.Delete([]byte(s.collectionKey(name)))
	})

	if err != nil {
		return fmt.Errorf("failed to drop collection: %w", err)
	}

	// Clear cache entries for this collection
	s.cache.Clear()

	return nil
}

// Insert adds a new document to a collection
func (s *BadgerStorage) Insert(ctx context.Context, collection string, doc *Document) error {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	
	doc.Version = 1
	doc.Created = time.Now()
	doc.Modified = doc.Created

	key := s.documentKey(collection, doc.ID)
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	err = s.db.Update(func(txn *badger.Txn) error {
		// Check if document already exists
		_, err := txn.Get([]byte(key))
		if err == nil {
			return fmt.Errorf("document with ID '%s' already exists", doc.ID)
		}
		if err != badger.ErrKeyNotFound {
			return err
		}

		// Insert the document
		entry := badger.NewEntry([]byte(key), data)
		if s.config.TTL > 0 {
			entry = entry.WithTTL(s.config.TTL)
		}
		
		return txn.SetEntry(entry)
	})

	if err != nil {
		return err
	}

	// Cache the document
	s.cache.Set([]byte(key), data, int(s.config.TTL.Seconds()))

	return nil
}

// Update modifies an existing document
func (s *BadgerStorage) Update(ctx context.Context, collection string, id string, fields map[string]interface{}) (*Document, error) {
	key := s.documentKey(collection, id)
	
	var updatedDoc *Document
	
	err := s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return fmt.Errorf("document with ID '%s' not found", id)
			}
			return err
		}

		var doc Document
		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &doc)
		})
		if err != nil {
			return fmt.Errorf("failed to unmarshal document: %w", err)
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

		// Marshal and store
		data, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("failed to marshal updated document: %w", err)
		}

		entry := badger.NewEntry([]byte(key), data)
		if s.config.TTL > 0 {
			entry = entry.WithTTL(s.config.TTL)
		}

		updatedDoc = &doc
		return txn.SetEntry(entry)
	})

	if err != nil {
		return nil, err
	}

	// Update cache
	if data, err := json.Marshal(updatedDoc); err == nil {
		s.cache.Set([]byte(key), data, int(s.config.TTL.Seconds()))
	}

	return updatedDoc, nil
}

// Delete removes a document from a collection
func (s *BadgerStorage) Delete(ctx context.Context, collection string, id string) error {
	key := s.documentKey(collection, id)
	
	err := s.db.Update(func(txn *badger.Txn) error {
		// Check if document exists
		_, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return fmt.Errorf("document with ID '%s' not found", id)
			}
			return err
		}

		return txn.Delete([]byte(key))
	})

	if err != nil {
		return err
	}

	// Remove from cache
	s.cache.Del([]byte(key))

	return nil
}

// Get retrieves a document by ID
func (s *BadgerStorage) Get(ctx context.Context, collection string, id string) (*Document, error) {
	key := s.documentKey(collection, id)
	
	// Try cache first
	if cached, err := s.cache.Get([]byte(key)); err == nil {
		var doc Document
		if err := json.Unmarshal(cached, &doc); err == nil {
			return &doc, nil
		}
	}

	// Fallback to BadgerDB
	var doc Document
	
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return fmt.Errorf("document with ID '%s' not found", id)
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &doc)
		})
	})

	if err != nil {
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(doc); err == nil {
		s.cache.Set([]byte(key), data, int(s.config.TTL.Seconds()))
	}

	return &doc, nil
}

// List returns all documents in a collection
func (s *BadgerStorage) List(ctx context.Context, collection string, limit, offset int) ([]*Document, error) {
	prefix := s.documentPrefix(collection)
	var documents []*Document
	
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = limit
		it := txn.NewIterator(opts)
		defer it.Close()

		count := 0
		skipped := 0
		
		for it.Seek([]byte(prefix)); it.Valid(); it.Next() {
			key := it.Item().Key()
			if !strings.HasPrefix(string(key), prefix) {
				break
			}

			// Skip offset
			if skipped < offset {
				skipped++
				continue
			}

			// Apply limit
			if limit > 0 && count >= limit {
				break
			}

			var doc Document
			err := it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &doc)
			})
			if err != nil {
				continue // Skip corrupted documents
			}

			documents = append(documents, &doc)
			count++
		}

		return nil
	})

	return documents, err
}

// Count returns the number of documents in a collection
func (s *BadgerStorage) Count(ctx context.Context, collection string) (int, error) {
	prefix := s.documentPrefix(collection)
	count := 0
	
	err := s.db.View(func(txn *badger.Txn) error {
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

// Close closes the storage
func (s *BadgerStorage) Close() error {
	return s.db.Close()
}

// Helper methods

func (s *BadgerStorage) collectionKey(name string) string {
	return fmt.Sprintf("_collections:%s", name)
}

func (s *BadgerStorage) documentKey(collection, id string) string {
	return fmt.Sprintf("_docs:%s:%s", collection, id)
}

func (s *BadgerStorage) documentPrefix(collection string) string {
	return fmt.Sprintf("_docs:%s:", collection)
}

// runGC runs periodic garbage collection
func (s *BadgerStorage) runGC() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		err := s.db.RunValueLogGC(0.7)
		if err != nil {
			// GC didn't run, which is fine
		}
	}
}