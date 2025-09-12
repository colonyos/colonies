package kvstore

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
)

func TestColonyOperationsDetailed(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create a test colony
	testColony := &core.Colony{
		ID:   "test-colony-id-12345",
		Name: "test-colony-name",
	}

	t.Logf("Adding colony: %+v", testColony)

	// Test adding colony
	err = db.AddColony(testColony)
	if err != nil {
		t.Fatalf("Failed to add colony: %v", err)
	}

	t.Logf("Colony added successfully")

	// Test getting colony by name
	retrievedColony, err := db.GetColonyByName(testColony.Name)
	if err != nil {
		t.Fatalf("Failed to get colony by name: %v", err)
	}

	if retrievedColony.ID != testColony.ID {
		t.Fatalf("Expected colony ID %s, got %s", testColony.ID, retrievedColony.ID)
	}
	if retrievedColony.Name != testColony.Name {
		t.Fatalf("Expected colony name %s, got %s", testColony.Name, retrievedColony.Name)
	}

	t.Logf("Retrieved colony successfully: %+v", retrievedColony)

	// Debug: Let's see what's actually stored and search behavior
	t.Logf("Debug: Searching for colony with ID: %s", testColony.ID)
	
	// Try to get all items in colonies
	allResults, err := db.store.FindRecursive("/colonies", "", "")
	if err != nil {
		t.Logf("Debug: FindRecursive all error: %v", err)
	} else {
		t.Logf("Debug: Found %d total items in /colonies", len(allResults))
		for i, result := range allResults {
			t.Logf("Debug: Item %d - Path: %s, Value: %+v", i, result.Path, result.Value)
			if colony, ok := result.Value.(*core.Colony); ok {
				t.Logf("Debug: Colony - ID: %s, Name: %s", colony.ID, colony.Name)
			}
		}
	}
	
	// Try searching with different field names
	testCases := []string{"ID", "id", "ColonyID", "colonyid"}
	for _, fieldName := range testCases {
		t.Logf("Debug: Trying field name '%s'", fieldName)
		searchResults, err := db.store.FindRecursive("/colonies", fieldName, testColony.ID)
		if err != nil {
			t.Logf("Debug: FindRecursive error for %s: %v", fieldName, err)
		} else {
			t.Logf("Debug: FindRecursive for %s returned %d results", fieldName, len(searchResults))
		}
	}

	// Test getting colony by ID
	retrievedByID, err := db.GetColonyByID(testColony.ID)
	if err != nil {
		t.Fatalf("Failed to get colony by ID: %v", err)
	}

	if retrievedByID.Name != testColony.Name {
		t.Fatalf("Expected colony name %s, got %s", testColony.Name, retrievedByID.Name)
	}

	t.Logf("Retrieved colony by ID successfully: %+v", retrievedByID)

	// Test listing all colonies
	colonies, err := db.GetColonies()
	if err != nil {
		t.Fatalf("Failed to get all colonies: %v", err)
	}

	if len(colonies) != 1 {
		t.Fatalf("Expected 1 colony, got %d", len(colonies))
	}

	if colonies[0].Name != testColony.Name {
		t.Fatalf("Expected colony name %s, got %s", testColony.Name, colonies[0].Name)
	}

	t.Logf("Listed all colonies successfully: %d colonies found", len(colonies))
}

func TestAddDuplicateColony(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create a test colony
	testColony := &core.Colony{
		ID:   "test-colony-id-12345",
		Name: "test-colony-name",
	}

	// Add colony first time
	err = db.AddColony(testColony)
	if err != nil {
		t.Fatalf("Failed to add colony first time: %v", err)
	}

	// Try to add same colony again - should fail
	err = db.AddColony(testColony)
	if err == nil {
		t.Fatal("Expected error when adding duplicate colony, but got nil")
	}

	expectedError := "colony with name test-colony-name already exists"
	if err.Error() != expectedError {
		t.Fatalf("Expected error '%s', got '%s'", expectedError, err.Error())
	}

	t.Logf("Correctly rejected duplicate colony with error: %v", err)
}