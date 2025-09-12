package kvstore

import (
	"testing"
)

func TestSetGetServerID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	testServerID := "test-server-id-12345"

	// Test setting server ID
	err = db.SetServerID("", testServerID)
	if err != nil {
		t.Fatalf("Failed to set server ID: %v", err)
	}

	// Test getting server ID
	retrievedID, err := db.GetServerID()
	if err != nil {
		t.Fatalf("Failed to get server ID: %v", err)
	}

	if retrievedID != testServerID {
		t.Fatalf("Expected server ID %s, got %s", testServerID, retrievedID)
	}

	t.Logf("Successfully set and retrieved server ID: %s", retrievedID)
}

func TestServerIDNotFound(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Try to get server ID when none is set
	_, err = db.GetServerID()
	if err == nil {
		t.Fatal("Expected error when server ID not set, but got nil")
	}

	expectedError := "server ID not found"
	if err.Error() != expectedError {
		t.Fatalf("Expected error '%s', got '%s'", expectedError, err.Error())
	}

	t.Logf("Correctly returned error when server ID not found: %v", err)
}

func TestUpdateServerID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Set initial server ID
	initialID := "initial-server-id"
	err = db.SetServerID("", initialID)
	if err != nil {
		t.Fatalf("Failed to set initial server ID: %v", err)
	}

	// Update server ID
	newID := "updated-server-id"
	err = db.SetServerID(initialID, newID)
	if err != nil {
		t.Fatalf("Failed to update server ID: %v", err)
	}

	// Verify the update
	retrievedID, err := db.GetServerID()
	if err != nil {
		t.Fatalf("Failed to get updated server ID: %v", err)
	}

	if retrievedID != newID {
		t.Fatalf("Expected updated server ID %s, got %s", newID, retrievedID)
	}

	t.Logf("Successfully updated server ID from %s to %s", initialID, newID)
}