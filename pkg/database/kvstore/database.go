package kvstore

import (
	"errors"
	"time"
)

// KVStoreDatabase implements all database interfaces using the kvstore
type KVStoreDatabase struct {
	store       *MixedKVStore
	initialized bool
	locked      bool
	lockTimeout time.Time
}

// NewKVStoreDatabase creates a new KVStore-based database adapter
func NewKVStoreDatabase() *KVStoreDatabase {
	return &KVStoreDatabase{
		store: NewMixedKVStore(),
	}
}

// KVStoreDatabase implements all required database interfaces

// =====================================
// DatabaseCore Interface Implementation
// =====================================

// Close closes the database connection
func (db *KVStoreDatabase) Close() {
	db.initialized = false
	db.locked = false
}

// Initialize sets up the database structure
func (db *KVStoreDatabase) Initialize() error {
	if db.initialized {
		return nil
	}

	// For KVStore adapter, we don't need to pre-create structure
	// The store will handle path creation dynamically
	db.initialized = true
	return nil
}

// Drop removes all data from the database
func (db *KVStoreDatabase) Drop() error {
	err := db.store.Clear()
	if err != nil {
		return err
	}
	db.initialized = false
	return nil
}

// Lock locks the database with a timeout
func (db *KVStoreDatabase) Lock(timeout int) error {
	if db.locked && time.Now().Before(db.lockTimeout) {
		return errors.New("database is already locked")
	}
	db.locked = true
	db.lockTimeout = time.Now().Add(time.Duration(timeout) * time.Second)
	return nil
}

// Unlock unlocks the database
func (db *KVStoreDatabase) Unlock() error {
	db.locked = false
	db.lockTimeout = time.Time{}
	return nil
}

// ApplyRetentionPolicy removes old data based on retention period
func (db *KVStoreDatabase) ApplyRetentionPolicy(retentionPeriod int64) error {
	// KVStore implementation doesn't support automatic retention yet
	// This would need to be implemented based on timestamps
	return nil
}

// IsInitialized returns whether the database is initialized
func (db *KVStoreDatabase) IsInitialized() bool {
	return db.initialized
}