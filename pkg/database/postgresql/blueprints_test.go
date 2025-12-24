package postgresql

import (
	"fmt"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestAddGetBlueprintDefinition(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	sd := core.CreateBlueprintDefinition(
		"executor-deployment",
		"compute.colonies.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"executor-controller",
		"reconcile",
	)
	sd.Metadata.ColonyName = "test-colony"

	err = db.AddBlueprintDefinition(sd)
	assert.Nil(t, err)

	// Get by ID
	sdFromDB, err := db.GetBlueprintDefinitionByID(sd.ID)
	assert.Nil(t, err)
	assert.NotNil(t, sdFromDB)
	assert.Equal(t, sd.ID, sdFromDB.ID)
	assert.Equal(t, sd.Metadata.Name, sdFromDB.Metadata.Name)
	assert.Equal(t, sd.Spec.Group, sdFromDB.Spec.Group)
	assert.Equal(t, sd.Spec.Version, sdFromDB.Spec.Version)

	// Get by name
	sdFromDB2, err := db.GetBlueprintDefinitionByName(sd.Metadata.ColonyName, sd.Metadata.Name)
	assert.Nil(t, err)
	assert.NotNil(t, sdFromDB2)
	assert.Equal(t, sd.ID, sdFromDB2.ID)

	// Get all
	sds, err := db.GetBlueprintDefinitions()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sds))

	// Count
	count, err := db.CountBlueprintDefinitions()
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
}

func TestGetBlueprintDefinitionByKind(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create two blueprint definitions with different kinds
	sd1 := core.CreateBlueprintDefinition(
		"executor-deployment",
		"compute.colonies.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"executor-controller",
		"reconcile",
	)
	sd1.Metadata.ColonyName = "test-colony"

	sd2 := core.CreateBlueprintDefinition(
		"service-deployment",
		"compute.colonies.io",
		"v1",
		"ServiceDeployment",
		"servicedeployments",
		"Namespaced",
		"service-controller",
		"reconcile",
	)
	sd2.Metadata.ColonyName = "test-colony"

	err = db.AddBlueprintDefinition(sd1)
	assert.Nil(t, err)
	err = db.AddBlueprintDefinition(sd2)
	assert.Nil(t, err)

	// Test GetBlueprintDefinitionByKind - find ExecutorDeployment
	foundDef, err := db.GetBlueprintDefinitionByKind("ExecutorDeployment")
	assert.Nil(t, err)
	assert.NotNil(t, foundDef)
	assert.Equal(t, "ExecutorDeployment", foundDef.Spec.Names.Kind)
	assert.Equal(t, sd1.ID, foundDef.ID)

	// Test GetBlueprintDefinitionByKind - find ServiceDeployment
	foundDef2, err := db.GetBlueprintDefinitionByKind("ServiceDeployment")
	assert.Nil(t, err)
	assert.NotNil(t, foundDef2)
	assert.Equal(t, "ServiceDeployment", foundDef2.Spec.Names.Kind)
	assert.Equal(t, sd2.ID, foundDef2.ID)

	// Test GetBlueprintDefinitionByKind - non-existent kind returns nil
	notFound, err := db.GetBlueprintDefinitionByKind("NonExistentKind")
	assert.Nil(t, err)
	assert.Nil(t, notFound)
}

func TestAddGetBlueprint(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
	blueprint.SetSpec("image", "nginx:1.21")
	blueprint.SetSpec("replicas", 3)
	blueprint.SetStatus("phase", "Running")

	err = db.AddBlueprint(blueprint)
	assert.Nil(t, err)

	// Get by ID
	blueprintFromDB, err := db.GetBlueprintByID(blueprint.ID)
	assert.Nil(t, err)
	assert.NotNil(t, blueprintFromDB)
	assert.Equal(t, blueprint.ID, blueprintFromDB.ID)
	assert.Equal(t, blueprint.Metadata.Name, blueprintFromDB.Metadata.Name)
	assert.Equal(t, blueprint.Metadata.ColonyName, blueprintFromDB.Metadata.ColonyName)
	assert.Equal(t, blueprint.Kind, blueprintFromDB.Kind)

	// Verify spec
	image, ok := blueprintFromDB.GetSpec("image")
	assert.True(t, ok)
	assert.Equal(t, "nginx:1.21", image)

	replicas, ok := blueprintFromDB.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, float64(3), replicas) // JSON unmarshaling converts to float64

	// Verify status
	phase, ok := blueprintFromDB.GetStatus("phase")
	assert.True(t, ok)
	assert.Equal(t, "Running", phase)

	// Get by name
	blueprintFromDB2, err := db.GetBlueprintByName(blueprint.Metadata.ColonyName, blueprint.Metadata.Name)
	assert.Nil(t, err)
	assert.NotNil(t, blueprintFromDB2)
	assert.Equal(t, blueprint.ID, blueprintFromDB2.ID)

	// Get all
	blueprints, err := db.GetBlueprints()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(blueprints))

	// Count
	count, err := db.CountBlueprints()
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
}

func TestGetBlueprintsByNamespace(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	blueprint1 := core.CreateBlueprint("ExecutorDeployment", "web-1", "production")
	blueprint1.SetSpec("image", "nginx:1.21")

	blueprint2 := core.CreateBlueprint("ExecutorDeployment", "web-2", "production")
	blueprint2.SetSpec("image", "nginx:1.22")

	blueprint3 := core.CreateBlueprint("ExecutorDeployment", "web-3", "staging")
	blueprint3.SetSpec("image", "nginx:1.21")

	err = db.AddBlueprint(blueprint1)
	assert.Nil(t, err)
	err = db.AddBlueprint(blueprint2)
	assert.Nil(t, err)
	err = db.AddBlueprint(blueprint3)
	assert.Nil(t, err)

	// Get by namespace
	prodBlueprints, err := db.GetBlueprintsByNamespace("production")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(prodBlueprints))

	stagingBlueprints, err := db.GetBlueprintsByNamespace("staging")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(stagingBlueprints))

	// Count by namespace
	prodCount, err := db.CountBlueprintsByNamespace("production")
	assert.Nil(t, err)
	assert.Equal(t, 2, prodCount)
}

func TestGetBlueprintsByKind(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	blueprint1 := core.CreateBlueprint("ExecutorDeployment", "web-1", "production")
	blueprint1.SetSpec("image", "nginx:1.21")

	blueprint2 := core.CreateBlueprint("Database", "db-1", "production")
	blueprint2.SetSpec("engine", "postgres")

	err = db.AddBlueprint(blueprint1)
	assert.Nil(t, err)
	err = db.AddBlueprint(blueprint2)
	assert.Nil(t, err)

	// Get by kind
	executorDeployments, err := db.GetBlueprintsByKind("ExecutorDeployment")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(executorDeployments))

	databases, err := db.GetBlueprintsByKind("Database")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(databases))
}

func TestUpdateBlueprint(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
	blueprint.SetSpec("replicas", 3)

	err = db.AddBlueprint(blueprint)
	assert.Nil(t, err)

	// Update service
	blueprint.SetSpec("replicas", 5)
	err = db.UpdateBlueprint(blueprint)
	assert.Nil(t, err)

	// Verify update
	blueprintFromDB, err := db.GetBlueprintByID(blueprint.ID)
	assert.Nil(t, err)
	replicas, ok := blueprintFromDB.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, float64(5), replicas)
}

func TestUpdateBlueprintStatus(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
	blueprint.SetSpec("replicas", 3)
	blueprint.SetStatus("phase", "Pending")

	err = db.AddBlueprint(blueprint)
	assert.Nil(t, err)

	// Update status only
	newStatus := map[string]interface{}{
		"phase": "Running",
		"ready": 3,
	}
	err = db.UpdateBlueprintStatus(blueprint.ID, newStatus)
	assert.Nil(t, err)

	// Verify update
	blueprintFromDB, err := db.GetBlueprintByID(blueprint.ID)
	assert.Nil(t, err)
	phase, ok := blueprintFromDB.GetStatus("phase")
	assert.True(t, ok)
	assert.Equal(t, "Running", phase)
	ready, ok := blueprintFromDB.GetStatus("ready")
	assert.True(t, ok)
	assert.Equal(t, float64(3), ready)
}

func TestRemoveBlueprint(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
	err = db.AddBlueprint(blueprint)
	assert.Nil(t, err)

	// Remove by ID
	err = db.RemoveBlueprintByID(blueprint.ID)
	assert.Nil(t, err)

	// Verify removed
	blueprintFromDB, err := db.GetBlueprintByID(blueprint.ID)
	assert.Nil(t, err)
	assert.Nil(t, blueprintFromDB)

	count, err := db.CountBlueprints()
	assert.Nil(t, err)
	assert.Equal(t, 0, count)
}

func TestRemoveBlueprintsByNamespace(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	blueprint1 := core.CreateBlueprint("ExecutorDeployment", "web-1", "production")
	blueprint2 := core.CreateBlueprint("ExecutorDeployment", "web-2", "production")
	blueprint3 := core.CreateBlueprint("ExecutorDeployment", "web-3", "staging")

	err = db.AddBlueprint(blueprint1)
	assert.Nil(t, err)
	err = db.AddBlueprint(blueprint2)
	assert.Nil(t, err)
	err = db.AddBlueprint(blueprint3)
	assert.Nil(t, err)

	// Remove production namespace
	err = db.RemoveBlueprintsByNamespace("production")
	assert.Nil(t, err)

	// Verify
	prodCount, err := db.CountBlueprintsByNamespace("production")
	assert.Nil(t, err)
	assert.Equal(t, 0, prodCount)

	stagingCount, err := db.CountBlueprintsByNamespace("staging")
	assert.Nil(t, err)
	assert.Equal(t, 1, stagingCount)
}

func TestAddGetBlueprintHistory(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a service
	blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
	blueprint.SetSpec("replicas", 3)
	blueprint.SetStatus("phase", "Running")

	err = db.AddBlueprint(blueprint)
	assert.Nil(t, err)

	t.Logf("Blueprint generation after creation: %d", blueprint.Metadata.Generation)

	// Create history entry for service creation
	history := core.CreateBlueprintHistory(blueprint, "test-user", "create")
	t.Logf("Creating history with generation: %d, ID: %s", history.Generation, history.ID)
	err = db.AddBlueprintHistory(history)
	assert.Nil(t, err)

	// Get history
	histories, err := db.GetBlueprintHistory(blueprint.ID, 10)
	assert.Nil(t, err)
	t.Logf("Got %d history entries:", len(histories))
	for i, h := range histories {
		t.Logf("  History[%d]: ID=%s, Generation=%d, ChangeType=%s, ChangedBy=%s", i, h.ID, h.Generation, h.ChangeType, h.ChangedBy)
	}
	assert.Equal(t, 1, len(histories))
	assert.Equal(t, blueprint.ID, histories[0].BlueprintID)
	assert.Equal(t, "ExecutorDeployment", histories[0].Kind)
	assert.Equal(t, "production", histories[0].Namespace)
	assert.Equal(t, "web-server", histories[0].Name)
	assert.Equal(t, blueprint.Metadata.Generation, histories[0].Generation)
	assert.Equal(t, "test-user", histories[0].ChangedBy)
	assert.Equal(t, "create", histories[0].ChangeType)

	// Verify spec in history
	replicas, ok := histories[0].Spec["replicas"]
	assert.True(t, ok)
	assert.Equal(t, float64(3), replicas)

	// Verify status in history
	phase, ok := histories[0].Status["phase"]
	assert.True(t, ok)
	assert.Equal(t, "Running", phase)
}

func TestBlueprintHistoryMultipleVersions(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a service
	blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
	blueprint.SetSpec("replicas", 3)

	err = db.AddBlueprint(blueprint)
	assert.Nil(t, err)

	initialGen := blueprint.Metadata.Generation

	// Create initial history entry
	history1 := core.CreateBlueprintHistory(blueprint, "user1", "create")
	err = db.AddBlueprintHistory(history1)
	assert.Nil(t, err)

	// Update service
	blueprint.SetSpec("replicas", 5)
	blueprint.Metadata.Generation = initialGen + 1

	// Create second history entry
	history2 := core.CreateBlueprintHistory(blueprint, "user2", "update")
	err = db.AddBlueprintHistory(history2)
	assert.Nil(t, err)

	// Update again
	blueprint.SetSpec("replicas", 10)
	blueprint.Metadata.Generation = initialGen + 2

	// Create third history entry
	history3 := core.CreateBlueprintHistory(blueprint, "user3", "update")
	err = db.AddBlueprintHistory(history3)
	assert.Nil(t, err)

	// Get all history (no limit)
	allHistories, err := db.GetBlueprintHistory(blueprint.ID, 0)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(allHistories))

	// Verify they're ordered by timestamp DESC (most recent first)
	assert.Equal(t, initialGen+2, allHistories[0].Generation)
	assert.Equal(t, initialGen+1, allHistories[1].Generation)
	assert.Equal(t, initialGen, allHistories[2].Generation)

	// Get limited history
	limitedHistories, err := db.GetBlueprintHistory(blueprint.ID, 2)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(limitedHistories))
	assert.Equal(t, initialGen+2, limitedHistories[0].Generation)
	assert.Equal(t, initialGen+1, limitedHistories[1].Generation)
}

func TestGetBlueprintHistoryByGeneration(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a service
	blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
	blueprint.SetSpec("replicas", 3)

	err = db.AddBlueprint(blueprint)
	assert.Nil(t, err)

	// Create multiple history entries
	history1 := core.CreateBlueprintHistory(blueprint, "user1", "create")
	err = db.AddBlueprintHistory(history1)
	assert.Nil(t, err)

	blueprint.SetSpec("replicas", 5)
	blueprint.Metadata.Generation = 2
	history2 := core.CreateBlueprintHistory(blueprint, "user2", "update")
	err = db.AddBlueprintHistory(history2)
	assert.Nil(t, err)

	// Get specific generation
	historyGen2, err := db.GetBlueprintHistoryByGeneration(blueprint.ID, 2)
	assert.Nil(t, err)
	assert.NotNil(t, historyGen2)
	assert.Equal(t, int64(2), historyGen2.Generation)
	assert.Equal(t, "user2", historyGen2.ChangedBy)

	replicas, ok := historyGen2.Spec["replicas"]
	assert.True(t, ok)
	assert.Equal(t, float64(5), replicas)

	// Get generation that doesn't exist
	historyGen99, err := db.GetBlueprintHistoryByGeneration(blueprint.ID, 99)
	assert.Nil(t, err)
	assert.Nil(t, historyGen99)
}

func TestRemoveBlueprintHistory(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a service
	blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
	err = db.AddBlueprint(blueprint)
	assert.Nil(t, err)

	// Create history entries
	history1 := core.CreateBlueprintHistory(blueprint, "user1", "create")
	err = db.AddBlueprintHistory(history1)
	assert.Nil(t, err)

	blueprint.Metadata.Generation = 2
	history2 := core.CreateBlueprintHistory(blueprint, "user2", "update")
	err = db.AddBlueprintHistory(history2)
	assert.Nil(t, err)

	// Verify history exists
	histories, err := db.GetBlueprintHistory(blueprint.ID, 0)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(histories))

	// Remove all history for this service
	err = db.RemoveBlueprintHistory(blueprint.ID)
	assert.Nil(t, err)

	// Verify history is removed
	historiesAfter, err := db.GetBlueprintHistory(blueprint.ID, 0)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(historiesAfter))
}

func TestBlueprintHistoryWithStatusChanges(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a service
	blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
	blueprint.SetSpec("replicas", 3)
	blueprint.SetStatus("phase", "Pending")
	blueprint.SetStatus("ready", 0)

	err = db.AddBlueprint(blueprint)
	assert.Nil(t, err)

	// Create initial history
	history1 := core.CreateBlueprintHistory(blueprint, "controller", "create")
	err = db.AddBlueprintHistory(history1)
	assert.Nil(t, err)

	// Update status only (status update via reconciliation)
	blueprint.SetStatus("phase", "Running")
	blueprint.SetStatus("ready", 3)
	blueprint.Metadata.Generation = 2

	history2 := core.CreateBlueprintHistory(blueprint, "reconciler", "status-update")
	err = db.AddBlueprintHistory(history2)
	assert.Nil(t, err)

	// Get history and verify status changes are tracked
	histories, err := db.GetBlueprintHistory(blueprint.ID, 0)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(histories))

	// Check latest status
	phase, ok := histories[0].Status["phase"]
	assert.True(t, ok)
	assert.Equal(t, "Running", phase)
	ready, ok := histories[0].Status["ready"]
	assert.True(t, ok)
	assert.Equal(t, float64(3), ready)

	// Check original status
	phaseOld, ok := histories[1].Status["phase"]
	assert.True(t, ok)
	assert.Equal(t, "Pending", phaseOld)
	readyOld, ok := histories[1].Status["ready"]
	assert.True(t, ok)
	assert.Equal(t, float64(0), readyOld)
}

func TestGetBlueprintsByLocationCaseInsensitive(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create blueprint with lowercase location "home"
	blueprint1 := core.CreateBlueprint("ExecutorDeployment", "web-1", "production")
	blueprint1.Metadata.LocationName = "home"
	blueprint1.SetSpec("image", "nginx:1.21")
	err = db.AddBlueprint(blueprint1)
	assert.Nil(t, err)

	// Create blueprint with uppercase location "HOME"
	blueprint2 := core.CreateBlueprint("ExecutorDeployment", "web-2", "production")
	blueprint2.Metadata.LocationName = "HOME"
	blueprint2.SetSpec("image", "nginx:1.22")
	err = db.AddBlueprint(blueprint2)
	assert.Nil(t, err)

	// Create blueprint with mixed case location "Home"
	blueprint3 := core.CreateBlueprint("ExecutorDeployment", "web-3", "production")
	blueprint3.Metadata.LocationName = "Home"
	blueprint3.SetSpec("image", "nginx:1.23")
	err = db.AddBlueprint(blueprint3)
	assert.Nil(t, err)

	// Create blueprint at different location "office"
	blueprint4 := core.CreateBlueprint("ExecutorDeployment", "web-4", "production")
	blueprint4.Metadata.LocationName = "office"
	blueprint4.SetSpec("image", "nginx:1.24")
	err = db.AddBlueprint(blueprint4)
	assert.Nil(t, err)

	// Query with "Home" should return all three home blueprints (case-insensitive)
	blueprints, err := db.GetBlueprintsByNamespaceKindAndLocation("production", "ExecutorDeployment", "Home")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(blueprints))

	// Query with "home" should also return all three
	blueprints, err = db.GetBlueprintsByNamespaceKindAndLocation("production", "ExecutorDeployment", "home")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(blueprints))

	// Query with "HOME" should also return all three
	blueprints, err = db.GetBlueprintsByNamespaceKindAndLocation("production", "ExecutorDeployment", "HOME")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(blueprints))

	// Query with "office" should return only blueprint4
	blueprints, err = db.GetBlueprintsByNamespaceKindAndLocation("production", "ExecutorDeployment", "office")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(blueprints))
	assert.Equal(t, "web-4", blueprints[0].Metadata.Name)

	// Query with empty location should return all four
	blueprints, err = db.GetBlueprintsByNamespaceKindAndLocation("production", "ExecutorDeployment", "")
	assert.Nil(t, err)
	assert.Equal(t, 4, len(blueprints))
}

// TestUpdateBlueprintStatusConcurrent tests that concurrent updates to UpdateBlueprintStatus
// do not lose updates. This test verifies the fix for the race condition where the previous
// read-modify-write pattern caused lost updates.
//
// The fix uses PostgreSQL's jsonb_set to atomically update the status field in a single
// SQL statement, eliminating the race condition.
//
// Note: Each call to UpdateBlueprintStatus replaces the entire status object, so this test
// verifies that the last write wins correctly (no corruption), not that all fields merge.
func TestUpdateBlueprintStatusConcurrent(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	// Create a blueprint
	blueprint := core.CreateBlueprint("ExecutorDeployment", "concurrent-test", "test-colony")
	blueprint.SetSpec("image", "nginx:1.21")
	blueprint.SetStatus("initial", "value")

	err = db.AddBlueprint(blueprint)
	assert.Nil(t, err)

	// Number of concurrent updaters
	numUpdaters := 10

	// Channel to synchronize goroutine start
	startChan := make(chan struct{})

	// Channel to collect errors
	errChan := make(chan error, numUpdaters)

	// Launch goroutines that will update status concurrently
	for i := 0; i < numUpdaters; i++ {
		go func(index int) {
			// Wait for start signal
			<-startChan

			// Each goroutine sets a complete status with its index
			status := map[string]interface{}{
				"updater": index,
				"value":   fmt.Sprintf("update_%d", index),
			}

			err := db.UpdateBlueprintStatus(blueprint.ID, status)
			errChan <- err
		}(i)
	}

	// Start all goroutines simultaneously
	close(startChan)

	// Wait for all updates to complete - all should succeed without error
	for i := 0; i < numUpdaters; i++ {
		err := <-errChan
		assert.Nil(t, err, "Update should not return error")
	}

	// Read the final blueprint state
	finalBlueprint, err := db.GetBlueprintByID(blueprint.ID)
	assert.Nil(t, err)

	// Verify that the status has valid data from one of the updaters
	// (last write wins, but should be consistent)
	updater, ok := finalBlueprint.GetStatus("updater")
	assert.True(t, ok, "Status should have 'updater' field")
	assert.NotNil(t, updater, "Updater value should not be nil")

	value, ok := finalBlueprint.GetStatus("value")
	assert.True(t, ok, "Status should have 'value' field")
	assert.NotNil(t, value, "Value should not be nil")

	t.Logf("Final status from updater: %v, value: %v", updater, value)
}

// TestUpdateBlueprintStatusAtomicUpdate verifies that the atomic update using jsonb_set
// works correctly and preserves other fields in the blueprint data.
func TestUpdateBlueprintStatusAtomicUpdate(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	// Create a blueprint with spec and initial status
	blueprint := core.CreateBlueprint("ExecutorDeployment", "atomic-test", "test-colony")
	blueprint.SetSpec("image", "nginx:1.21")
	blueprint.SetSpec("replicas", 3)
	blueprint.SetStatus("phase", "Pending")

	err = db.AddBlueprint(blueprint)
	assert.Nil(t, err)

	// Update status
	newStatus := map[string]interface{}{
		"phase":    "Running",
		"replicas": 3,
		"ready":    true,
	}
	err = db.UpdateBlueprintStatus(blueprint.ID, newStatus)
	assert.Nil(t, err)

	// Read back and verify spec is preserved
	updatedBlueprint, err := db.GetBlueprintByID(blueprint.ID)
	assert.Nil(t, err)

	// Verify spec is still intact
	image, ok := updatedBlueprint.GetSpec("image")
	assert.True(t, ok, "Spec 'image' should be preserved")
	assert.Equal(t, "nginx:1.21", image)

	replicas, ok := updatedBlueprint.GetSpec("replicas")
	assert.True(t, ok, "Spec 'replicas' should be preserved")
	assert.Equal(t, float64(3), replicas)

	// Verify status was updated
	phase, ok := updatedBlueprint.GetStatus("phase")
	assert.True(t, ok, "Status 'phase' should exist")
	assert.Equal(t, "Running", phase)

	ready, ok := updatedBlueprint.GetStatus("ready")
	assert.True(t, ok, "Status 'ready' should exist")
	assert.Equal(t, true, ready)

	// Verify metadata is preserved
	assert.Equal(t, "atomic-test", updatedBlueprint.Metadata.Name)
	assert.Equal(t, "test-colony", updatedBlueprint.Metadata.ColonyName)
}

// TestUpdateBlueprintStatusSequentialUpdates verifies that sequential updates
// each properly replace the previous status.
func TestUpdateBlueprintStatusSequentialUpdates(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	blueprint := core.CreateBlueprint("ExecutorDeployment", "sequential-test", "test-colony")
	blueprint.SetSpec("image", "nginx:1.21")

	err = db.AddBlueprint(blueprint)
	assert.Nil(t, err)

	// Perform sequential updates
	statuses := []map[string]interface{}{
		{"phase": "Pending", "message": "Waiting for resources"},
		{"phase": "Creating", "message": "Creating containers"},
		{"phase": "Running", "message": "All containers running", "ready": true},
	}

	for i, status := range statuses {
		err = db.UpdateBlueprintStatus(blueprint.ID, status)
		assert.Nil(t, err, "Update %d should succeed", i)

		// Verify the update took effect
		bp, err := db.GetBlueprintByID(blueprint.ID)
		assert.Nil(t, err)

		phase, ok := bp.GetStatus("phase")
		assert.True(t, ok)
		assert.Equal(t, status["phase"], phase)

		message, ok := bp.GetStatus("message")
		assert.True(t, ok)
		assert.Equal(t, status["message"], message)
	}

	// Final verification
	finalBp, err := db.GetBlueprintByID(blueprint.ID)
	assert.Nil(t, err)

	phase, _ := finalBp.GetStatus("phase")
	assert.Equal(t, "Running", phase)

	ready, ok := finalBp.GetStatus("ready")
	assert.True(t, ok)
	assert.Equal(t, true, ready)
}

func TestGetBlueprintHistoryParameterizedLimit(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a blueprint
	blueprint := core.CreateBlueprint("ExecutorDeployment", "limit-test", "production")
	blueprint.SetSpec("replicas", 1)

	err = db.AddBlueprint(blueprint)
	assert.Nil(t, err)

	// Create 5 history entries
	for i := 1; i <= 5; i++ {
		blueprint.SetSpec("replicas", i)
		blueprint.Metadata.Generation = int64(i)
		history := core.CreateBlueprintHistory(blueprint, "user", "update")
		err = db.AddBlueprintHistory(history)
		assert.Nil(t, err)
	}

	// Test: limit = 0 returns all entries
	allHistories, err := db.GetBlueprintHistory(blueprint.ID, 0)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(allHistories), "limit=0 should return all 5 entries")

	// Test: limit = 1 returns exactly 1 entry
	oneHistory, err := db.GetBlueprintHistory(blueprint.ID, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(oneHistory), "limit=1 should return exactly 1 entry")
	assert.Equal(t, int64(5), oneHistory[0].Generation, "Should return most recent (generation 5)")

	// Test: limit = 3 returns exactly 3 entries
	threeHistories, err := db.GetBlueprintHistory(blueprint.ID, 3)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(threeHistories), "limit=3 should return exactly 3 entries")
	assert.Equal(t, int64(5), threeHistories[0].Generation, "First should be generation 5")
	assert.Equal(t, int64(4), threeHistories[1].Generation, "Second should be generation 4")
	assert.Equal(t, int64(3), threeHistories[2].Generation, "Third should be generation 3")

	// Test: limit > total entries returns all entries
	manyHistories, err := db.GetBlueprintHistory(blueprint.ID, 100)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(manyHistories), "limit=100 should return all 5 entries (not more)")

	// Test: negative limit treated as no limit (returns all)
	negativeLimit, err := db.GetBlueprintHistory(blueprint.ID, -1)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(negativeLimit), "negative limit should return all entries")

	// Test: non-existent blueprint returns empty slice
	noHistories, err := db.GetBlueprintHistory("nonexistent-id", 10)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(noHistories), "non-existent blueprint should return empty slice")
}
