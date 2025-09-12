# KVStore Database Usage Example

The KVStore database adapter can now be used through the database factory. Here's how:

## Using the Database Factory

```go
package main

import (
    "log"
    "github.com/colonyos/colonies/pkg/database"
    "github.com/colonyos/colonies/pkg/core"
)

func main() {
    // Create KVStore database configuration
    config := database.DatabaseConfig{
        Type: database.KVStore,
        // Note: KVStore doesn't need Host, Port, User, Password, etc.
        // It's an in-memory database
    }

    // Create the database using the factory
    db, err := database.CreateDatabase(config)
    if err != nil {
        log.Fatalf("Failed to create database: %v", err)
    }
    defer db.Close()

    // Now you can use the database with any Database interface methods
    
    // Add a colony
    colony := &core.Colony{
        ID:   "test-colony-id",
        Name: "test-colony",
    }
    
    err = db.AddColony(colony)
    if err != nil {
        log.Fatalf("Failed to add colony: %v", err)
    }

    // Retrieve the colony
    retrievedColony, err := db.GetColonyByName("test-colony")
    if err != nil {
        log.Fatalf("Failed to get colony: %v", err)
    }
    
    log.Printf("Retrieved colony: %+v", retrievedColony)
}
```

## Direct Usage (Alternative)

```go
package main

import (
    "log"
    "github.com/colonyos/colonies/pkg/database/kvstore"
    "github.com/colonyos/colonies/pkg/core"
)

func main() {
    // Create KVStore database directly
    db := kvstore.NewKVStoreDatabase()
    err := db.Initialize()
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.Close()

    // Use the database...
}
```

## Environment Variable Configuration

You can now set the database type using environment variables or configuration files:

```bash
# Use KVStore database
export COLONIES_DB_TYPE=kvstore

# Use PostgreSQL database (existing)
export COLONIES_DB_TYPE=postgresql
export COLONIES_DB_HOST=localhost
export COLONIES_DB_PORT=5432
# ... other PostgreSQL settings
```

## Benefits of KVStore Database

- **In-memory**: Ultra-fast operations, perfect for development and testing
- **Zero setup**: No external database server required
- **Full compatibility**: Implements all the same interfaces as PostgreSQL
- **Thread-safe**: Safe for concurrent operations
- **Drop-in replacement**: Can replace PostgreSQL in any existing code
- **CLI Integration**: Full support for `COLONIES_DB_TYPE=kvstore make test`
- **Server ID Support**: Complete SecurityDatabase interface implementation

## Use Cases

- **Development**: Fast local development without PostgreSQL setup
- **Testing**: Unit and integration tests that need database functionality
- **Embedded systems**: Applications that need lightweight database functionality
- **Prototyping**: Quick prototyping without database infrastructure