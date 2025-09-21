package database

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/colonyos/colonies/pkg/database/postgresql"
)

// PrepareTests creates a test database based on COLONIES_DB_TYPE environment variable
// Defaults to KVStore if not specified for faster test execution
func PrepareTests() (Database, error) {
	return PrepareTestsWithPrefix("TEST_")
}

// PrepareTestsWithPrefix creates a test database with a custom prefix
func PrepareTestsWithPrefix(prefix string) (Database, error) {
	log.SetOutput(ioutil.Discard)

	dbType := os.Getenv("COLONIES_DB_TYPE")
	if dbType == "" {
		dbType = "kvstore" // Default to KVStore for faster tests
	}

	switch dbType {
	case "kvstore":
		// Create KVStore database
		config := DatabaseConfig{
			Type: KVStore,
		}
		return CreateDatabase(config)

	case "postgresql":
		// Create PostgreSQL database (existing behavior)
		db, err := postgresql.PrepareTestsWithPrefix(prefix)
		if err != nil {
			return nil, err
		}
		return db, nil

	default:
		// Fallback to KVStore
		config := DatabaseConfig{
			Type: KVStore,
		}
		return CreateDatabase(config)
	}
}