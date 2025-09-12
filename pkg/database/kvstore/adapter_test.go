package kvstore

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
)

// TestNewKVStoreDatabase tests the creation of a new KVStore database adapter
func TestNewKVStoreDatabase(t *testing.T) {
	db := NewKVStoreDatabase()
	if db == nil {
		t.Fatal("NewKVStoreDatabase() returned nil")
	}

	if db.store == nil {
		t.Fatal("KVStore database has nil store")
	}
}

// TestDatabaseInitialization tests the initialization of the database
func TestDatabaseInitialization(t *testing.T) {
	db := NewKVStoreDatabase()
	
	err := db.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Test that we can initialize multiple times without error
	err = db.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed on second call: %v", err)
	}
}

// TestUserOperations tests basic user database operations
func TestUserOperations(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Create test user
	user := &core.User{
		ID:         "user123",
		Name:       "testuser",
		ColonyName: "testcolony",
		Email:      "test@example.com",
	}

	// Add user
	err = db.AddUser(user)
	if err != nil {
		t.Fatalf("AddUser() failed: %v", err)
	}

	// Get user by ID
	retrievedUser, err := db.GetUserByID("testcolony", "user123")
	if err != nil {
		t.Fatalf("GetUserByID() failed: %v", err)
	}

	if retrievedUser.Name != "testuser" {
		t.Fatalf("Retrieved user name mismatch: expected 'testuser', got '%s'", retrievedUser.Name)
	}

	// Get user by name
	retrievedUser2, err := db.GetUserByName("testcolony", "testuser")
	if err != nil {
		t.Fatalf("GetUserByName() failed: %v", err)
	}

	if retrievedUser2.ID != "user123" {
		t.Fatalf("Retrieved user ID mismatch: expected 'user123', got '%s'", retrievedUser2.ID)
	}

	// Remove user
	err = db.RemoveUserByID("testcolony", "user123")
	if err != nil {
		t.Fatalf("RemoveUserByID() failed: %v", err)
	}

	// Verify user is removed
	_, err = db.GetUserByID("testcolony", "user123")
	if err == nil {
		t.Fatal("Expected error when getting removed user, but got nil")
	}
}

// TestColonyOperations tests basic colony database operations
func TestColonyOperations(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Create test colony
	colony := &core.Colony{
		ID:   "colony123",
		Name: "testcolony",
	}

	// Add colony
	err = db.AddColony(colony)
	if err != nil {
		t.Fatalf("AddColony() failed: %v", err)
	}

	// Get colony by name
	retrievedColony, err := db.GetColonyByName("testcolony")
	if err != nil {
		t.Fatalf("GetColonyByName() failed: %v", err)
	}

	if retrievedColony.ID != "colony123" {
		t.Fatalf("Retrieved colony ID mismatch: expected 'colony123', got '%s'", retrievedColony.ID)
	}

	// Count colonies
	count, err := db.CountColonies()
	if err != nil {
		t.Fatalf("CountColonies() failed: %v", err)
	}

	if count != 1 {
		t.Fatalf("Expected 1 colony, got %d", count)
	}

	// Remove colony
	err = db.RemoveColonyByName("testcolony")
	if err != nil {
		t.Fatalf("RemoveColonyByName() failed: %v", err)
	}

	// Verify colony is removed
	retrievedColony, err = db.GetColonyByName("testcolony")
	if err != nil {
		t.Fatalf("GetColonyByName() failed: %v", err)
	}
	if retrievedColony != nil {
		t.Fatal("Expected nil colony when getting removed colony, but got a colony")
	}
}

// TestProcessOperations tests basic process database operations
func TestProcessOperations(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Create test process
	process := &core.Process{
		ID:    "process123",
		State: 0, // WAITING
		FunctionSpec: core.FunctionSpec{
			FuncName: "testfunction",
			Conditions: core.Conditions{
				ColonyName:   "testcolony",
				ExecutorType: "testtype",
			},
		},
	}

	// Add process
	err = db.AddProcess(process)
	if err != nil {
		t.Fatalf("AddProcess() failed: %v", err)
	}

	// Get process by ID
	retrievedProcess, err := db.GetProcessByID("process123")
	if err != nil {
		t.Fatalf("GetProcessByID() failed: %v", err)
	}

	if retrievedProcess.State != 0 {
		t.Fatalf("Retrieved process state mismatch: expected 0, got %d", retrievedProcess.State)
	}

	// Set process state
	err = db.SetProcessState("process123", 1) // RUNNING
	if err != nil {
		t.Fatalf("SetProcessState() failed: %v", err)
	}

	// Verify state change
	updatedProcess, err := db.GetProcessByID("process123")
	if err != nil {
		t.Fatalf("GetProcessByID() after state change failed: %v", err)
	}

	if updatedProcess.State != 1 {
		t.Fatalf("Process state not updated: expected 1, got %d", updatedProcess.State)
	}

	// Remove process
	err = db.RemoveProcessByID("process123")
	if err != nil {
		t.Fatalf("RemoveProcessByID() failed: %v", err)
	}
}

// TestDatabaseClose tests database closing
func TestDatabaseClose(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Close should not error
	db.Close()

	// After close, should be able to initialize again
	err = db.Initialize()
	if err != nil {
		t.Fatalf("Initialize() after Close() failed: %v", err)
	}
}