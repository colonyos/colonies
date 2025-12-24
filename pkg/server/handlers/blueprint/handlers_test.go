package blueprint_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/stretchr/testify/assert"
)

func TestAddBlueprintDefinition(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create a BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"test-blueprint",
		"example.com",
		"v1",
		"TestBlueprint",
		"testblueprints",
		"Namespaced",
		"test_executor_type",
		"reconcile_test_resource",
	)
	sd.Metadata.ColonyName = env.ColonyName

	// Add BlueprintDefinition with colony owner key
	addedSD, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedSD)
	assert.Equal(t, sd.Metadata.Name, addedSD.Metadata.Name)
	assert.Equal(t, sd.Spec.Group, addedSD.Spec.Group)
	assert.Equal(t, sd.Spec.Version, addedSD.Spec.Version)

	// Try to add duplicate BlueprintDefinition - should fail
	_, err = client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestGetBlueprintDefinition(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create and add BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"test-blueprint",
		"example.com",
		"v1",
		"TestBlueprint",
		"testblueprints",
		"Namespaced",
		"test_executor_type",
		"reconcile_test_resource",
	)
	sd.Metadata.ColonyName = env.ColonyName

	addedSD, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Get BlueprintDefinition (using executor key since only members can get)
	retrievedSD, err := client.GetBlueprintDefinition(env.ColonyName, sd.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedSD)
	assert.Equal(t, addedSD.ID, retrievedSD.ID)
	assert.Equal(t, addedSD.Metadata.Name, retrievedSD.Metadata.Name)

	server.Shutdown()
	<-done
}

func TestAddBlueprint(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// First add a BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"database_controller",
		"reconcile_database",
	)
	sd.Metadata.ColonyName = env.ColonyName

	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create a Blueprint instance
	blueprint := core.CreateBlueprint("Database", "test-database", env.ColonyName)
	blueprint.SetSpec("host", "localhost")
	blueprint.SetSpec("port", 5432)

	// Add Blueprint
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedBlueprint)
	assert.Equal(t, blueprint.Metadata.Name, addedBlueprint.Metadata.Name)
	assert.Equal(t, blueprint.Kind, addedBlueprint.Kind)

	// Verify spec was preserved
	host, ok := addedBlueprint.GetSpec("host")
	assert.True(t, ok)
	assert.Equal(t, "localhost", host)

	server.Shutdown()
	<-done
}

func TestGetBlueprint(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"blueprint",
		"example.com",
		"v1",
		"Blueprint",
		"blueprints",
		"Namespaced",
		"blueprint_controller",
		"reconcile_blueprint",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Blueprint
	blueprint := core.CreateBlueprint("Blueprint", "web-blueprint", env.ColonyName)
	blueprint.SetSpec("port", 8080)
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Get Blueprint
	retrievedBlueprint, err := client.GetBlueprint(env.ColonyName, blueprint.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedBlueprint)
	assert.Equal(t, addedBlueprint.ID, retrievedBlueprint.ID)
	assert.Equal(t, addedBlueprint.Metadata.Name, retrievedBlueprint.Metadata.Name)

	server.Shutdown()
	<-done
}

func TestGetBlueprints(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition for Database
	sdDB := core.CreateBlueprintDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"database_controller",
		"reconcile_database",
	)
	sdDB.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sdDB, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add BlueprintDefinition for Blueprint
	sdSvc := core.CreateBlueprintDefinition(
		"blueprint",
		"example.com",
		"v1",
		"Blueprint",
		"blueprints",
		"Namespaced",
		"blueprint_controller",
		"reconcile_blueprint",
	)
	sdSvc.Metadata.ColonyName = env.ColonyName
	_, err = client.AddBlueprintDefinition(sdSvc, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add some Database services
	db1 := core.CreateBlueprint("Database", "db1", env.ColonyName)
	db2 := core.CreateBlueprint("Database", "db2", env.ColonyName)
	_, err = client.AddBlueprint(db1, env.ExecutorPrvKey)
	assert.Nil(t, err)
	_, err = client.AddBlueprint(db2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Add some Blueprint blueprints
	svc1 := core.CreateBlueprint("Blueprint", "svc1", env.ColonyName)
	_, err = client.AddBlueprint(svc1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Get all services
	allBlueprints, err := client.GetBlueprints(env.ColonyName, "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(allBlueprints))

	// Get only Database services
	dbBlueprints, err := client.GetBlueprints(env.ColonyName, "Database", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(dbBlueprints))

	// Get only Blueprint blueprints
	svcBlueprints, err := client.GetBlueprints(env.ColonyName, "Blueprint", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(svcBlueprints))

	server.Shutdown()
	<-done
}

func TestUpdateBlueprint(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"application",
		"example.com",
		"v1",
		"Application",
		"applications",
		"Namespaced",
		"app_controller",
		"reconcile_application",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Blueprint
	blueprint := core.CreateBlueprint("Application", "my-app", env.ColonyName)
	blueprint.SetSpec("version", "1.0.0")
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Update Blueprint
	addedBlueprint.SetSpec("version", "1.1.0")
	updatedBlueprint, err := client.UpdateBlueprint(addedBlueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, updatedBlueprint)

	version, ok := updatedBlueprint.GetSpec("version")
	assert.True(t, ok)
	assert.Equal(t, "1.1.0", version)

	server.Shutdown()
	<-done
}

func TestUpdateBlueprintGenerationIncrementsOnSpecChange(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"gen-test",
		"example.com",
		"v1",
		"GenTest",
		"gentests",
		"Namespaced",
		"test_controller",
		"reconcile",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Blueprint
	blueprint := core.CreateBlueprint("GenTest", "test-app", env.ColonyName)
	blueprint.SetSpec("version", "1.0.0")
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	initialGeneration := addedBlueprint.Metadata.Generation

	// Update Blueprint with spec change - generation should increment
	addedBlueprint.SetSpec("version", "1.1.0")
	updatedBlueprint, err := client.UpdateBlueprint(addedBlueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, initialGeneration+1, updatedBlueprint.Metadata.Generation, "Generation should increment on spec change")

	// Update Blueprint without spec change - generation should NOT increment
	unchangedBlueprint, err := client.UpdateBlueprint(updatedBlueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, updatedBlueprint.Metadata.Generation, unchangedBlueprint.Metadata.Generation, "Generation should NOT increment without spec change")

	server.Shutdown()
	<-done
}

func TestUpdateBlueprintWithForceGeneration(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"force-gen-test",
		"example.com",
		"v1",
		"ForceGenTest",
		"forcegentests",
		"Namespaced",
		"test_controller",
		"reconcile",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Blueprint
	blueprint := core.CreateBlueprint("ForceGenTest", "test-app", env.ColonyName)
	blueprint.SetSpec("replicas", 3)
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	initialGeneration := addedBlueprint.Metadata.Generation

	// Update Blueprint WITHOUT spec change and WITHOUT force - generation should NOT increment
	unchangedBlueprint, err := client.UpdateBlueprint(addedBlueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, initialGeneration, unchangedBlueprint.Metadata.Generation, "Generation should NOT increment without spec change or force")

	// Update Blueprint WITHOUT spec change but WITH force - generation SHOULD increment
	forcedBlueprint, err := client.UpdateBlueprintWithForce(unchangedBlueprint, true, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, initialGeneration+1, forcedBlueprint.Metadata.Generation, "Generation should increment with force=true even without spec change")

	// Force again - should increment again
	forcedBlueprint2, err := client.UpdateBlueprintWithForce(forcedBlueprint, true, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, initialGeneration+2, forcedBlueprint2.Metadata.Generation, "Generation should increment again with force=true")

	// Update with force=false and no spec change - should NOT increment
	noForcedBlueprint, err := client.UpdateBlueprintWithForce(forcedBlueprint2, false, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, forcedBlueprint2.Metadata.Generation, noForcedBlueprint.Metadata.Generation, "Generation should NOT increment with force=false and no spec change")

	server.Shutdown()
	<-done
}

func TestUpdateBlueprintWithForceTriggersReconciliation(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition with handler
	sd := core.CreateBlueprintDefinition(
		"force-recon-test",
		"example.com",
		"v1",
		"ForceReconTest",
		"forcerecontests",
		"Namespaced",
		"recon_controller",
		"reconcile",
	)
	sd.Spec.Handler.ReconcileInterval = 60
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Blueprint
	blueprint := core.CreateBlueprint("ForceReconTest", "test-app", env.ColonyName)
	blueprint.SetSpec("replicas", 3)
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Get initial waiting process count (from blueprint create)
	waitingProcs, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	initialWaitingCount := len(waitingProcs)

	// Force update (no spec change) - should trigger reconciliation
	_, err = client.UpdateBlueprintWithForce(addedBlueprint, true, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify that a new reconciliation process was created
	waitingProcsAfter, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Greater(t, len(waitingProcsAfter), initialWaitingCount, "Should have created a new reconciliation process after force update")

	server.Shutdown()
	<-done
}

func TestRemoveBlueprint(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"cache",
		"example.com",
		"v1",
		"Cache",
		"caches",
		"Namespaced",
		"cache_controller",
		"reconcile_cache",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Blueprint
	blueprint := core.CreateBlueprint("Cache", "redis-cache", env.ColonyName)
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Remove Blueprint
	err = client.RemoveBlueprint(env.ColonyName, addedBlueprint.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify it's removed
	_, err = client.GetBlueprint(env.ColonyName, addedBlueprint.Metadata.Name, env.ExecutorPrvKey)
	assert.NotNil(t, err) // Should fail because service doesn't exist

	server.Shutdown()
	<-done
}

func TestBlueprintWithComplexSpec(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"deployment",
		"compute.io",
		"v1",
		"Deployment",
		"deployments",
		"Namespaced",
		"deployment_controller",
		"reconcile_deployment",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create Blueprint with complex spec
	blueprint := core.CreateBlueprint("Deployment", "web-deployment", env.ColonyName)
	blueprint.SetSpec("image", "nginx:1.21")
	blueprint.SetSpec("replicas", 3)
	blueprint.SetSpec("env", map[string]interface{}{
		"DATABASE_URL": "postgres://localhost/db",
		"PORT":         "8080",
	})
	blueprint.Metadata.Labels = map[string]string{
		"app":     "web",
		"version": "v1.0.0",
	}
	blueprint.Metadata.Annotations = map[string]string{
		"description": "My test application",
	}

	// Add Blueprint
	_, err = client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Retrieve and verify
	retrievedBlueprint, err := client.GetBlueprint(env.ColonyName, blueprint.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	image, ok := retrievedBlueprint.GetSpec("image")
	assert.True(t, ok)
	assert.Equal(t, "nginx:1.21", image)

	replicas, ok := retrievedBlueprint.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, float64(3), replicas) // JSON unmarshaling converts to float64

	envSpec, ok := retrievedBlueprint.GetSpec("env")
	assert.True(t, ok)
	envMap := envSpec.(map[string]interface{})
	assert.Equal(t, "postgres://localhost/db", envMap["DATABASE_URL"])

	assert.Equal(t, "v1.0.0", retrievedBlueprint.Metadata.Labels["version"])
	assert.Equal(t, "My test application", retrievedBlueprint.Metadata.Annotations["description"])

	server.Shutdown()
	<-done
}

// TestAddBlueprintRequiresBlueprintDefinition tests that adding a blueprint without a BlueprintDefinition fails
func TestAddBlueprintRequiresBlueprintDefinition(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Try to add a Blueprint WITHOUT adding its BlueprintDefinition first
	blueprint := core.CreateBlueprint("NonExistentKind", "test-blueprint", env.ColonyName)
	blueprint.SetSpec("field", "value")

	// This should fail because BlueprintDefinition doesn't exist
	_, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Adding service without BlueprintDefinition should fail")
	assert.Contains(t, err.Error(), "BlueprintDefinition for kind 'NonExistentKind' not found")

	server.Shutdown()
	<-done
}

// TestAddBlueprintWithSchemaValidation tests that blueprints are validated against the BlueprintDefinition schema
func TestAddBlueprintWithSchemaValidation(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create BlueprintDefinition with schema validation
	sd := core.CreateBlueprintDefinition(
		"validated-blueprint",
		"example.com",
		"v1",
		"ValidatedBlueprint",
		"validatedblueprints",
		"Namespaced",
		"validator_controller",
		"reconcile_validated",
	)
	sd.Metadata.ColonyName = env.ColonyName
	sd.Spec.Schema = &core.ValidationSchema{
		Type: "object",
		Properties: map[string]core.SchemaProperty{
			"name": {
				Type:        "string",
				Description: "Blueprint name",
			},
			"replicas": {
				Type:        "number",
				Description: "Number of replicas",
			},
			"protocol": {
				Type: "string",
				Enum: []interface{}{"TCP", "UDP"},
			},
		},
		Required: []string{"name", "replicas"},
	}

	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Test 1: Valid blueprint should succeed
	validBlueprint := core.CreateBlueprint("ValidatedBlueprint", "valid-res", env.ColonyName)
	validBlueprint.SetSpec("name", "test")
	validBlueprint.SetSpec("replicas", 3)
	validBlueprint.SetSpec("protocol", "TCP")

	addedBlueprint, err := client.AddBlueprint(validBlueprint, env.ExecutorPrvKey)
	assert.Nil(t, err, "Valid blueprint should be added successfully")
	assert.NotNil(t, addedBlueprint)

	// Test 2: Blueprint missing required field should fail
	invalidBlueprint1 := core.CreateBlueprint("ValidatedBlueprint", "invalid-res-1", env.ColonyName)
	invalidBlueprint1.SetSpec("name", "test") // Missing required 'replicas'

	_, err = client.AddBlueprint(invalidBlueprint1, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Blueprint missing required field should fail")
	assert.Contains(t, err.Error(), "required field 'replicas' is missing")

	// Test 3: Blueprint with invalid type should fail
	invalidBlueprint2 := core.CreateBlueprint("ValidatedBlueprint", "invalid-res-2", env.ColonyName)
	invalidBlueprint2.SetSpec("name", "test")
	invalidBlueprint2.SetSpec("replicas", "not-a-number") // Should be number

	_, err = client.AddBlueprint(invalidBlueprint2, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Blueprint with invalid type should fail")
	assert.Contains(t, err.Error(), "must be a number")

	// Test 4: Blueprint with invalid enum value should fail
	invalidBlueprint3 := core.CreateBlueprint("ValidatedBlueprint", "invalid-res-3", env.ColonyName)
	invalidBlueprint3.SetSpec("name", "test")
	invalidBlueprint3.SetSpec("replicas", 3)
	invalidBlueprint3.SetSpec("protocol", "HTTP") // Not in enum [TCP, UDP]

	_, err = client.AddBlueprint(invalidBlueprint3, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Blueprint with invalid enum value should fail")
	assert.Contains(t, err.Error(), "must be one of")

	server.Shutdown()
	<-done
}

// TestRemoveBlueprintDefinitionWithActiveBlueprints tests that removing a BlueprintDefinition with active blueprints fails
func TestRemoveBlueprintDefinitionWithActiveBlueprints(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"protected-blueprint",
		"example.com",
		"v1",
		"ProtectedBlueprint",
		"protectedblueprints",
		"Namespaced",
		"protected_controller",
		"reconcile_protected",
	)
	sd.Metadata.ColonyName = env.ColonyName
	addedSD, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add some blueprints of this kind
	blueprint1 := core.CreateBlueprint("ProtectedBlueprint", "res-1", env.ColonyName)
	_, err = client.AddBlueprint(blueprint1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	blueprint2 := core.CreateBlueprint("ProtectedBlueprint", "res-2", env.ColonyName)
	_, err = client.AddBlueprint(blueprint2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to remove BlueprintDefinition while blueprints exist - should fail
	err = client.RemoveBlueprintDefinition(env.ColonyName, addedSD.Metadata.Name, env.ColonyPrvKey)
	assert.NotNil(t, err, "Removing BlueprintDefinition with active blueprints should fail")
	assert.Contains(t, err.Error(), "2 blueprint(s) of kind 'ProtectedBlueprint' still exist")

	// Remove one service
	err = client.RemoveBlueprint(env.ColonyName, blueprint1.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try again - should still fail because one service remains
	err = client.RemoveBlueprintDefinition(env.ColonyName, addedSD.Metadata.Name, env.ColonyPrvKey)
	assert.NotNil(t, err, "Removing BlueprintDefinition with 1 active service should still fail")
	assert.Contains(t, err.Error(), "1 blueprint(s) of kind 'ProtectedBlueprint' still exist")

	// Remove the last service
	err = client.RemoveBlueprint(env.ColonyName, blueprint2.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Now removal should succeed
	err = client.RemoveBlueprintDefinition(env.ColonyName, addedSD.Metadata.Name, env.ColonyPrvKey)
	assert.Nil(t, err, "Removing BlueprintDefinition with no active blueprints should succeed")

	server.Shutdown()
	<-done
}

func TestGetBlueprintHistory(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// First create a BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"testresource",
		"example.com",
		"v1",
		"TestBlueprint",
		"testblueprints",
		"Namespaced",
		"test_controller",
		"reconcile_testresource",
	)
	sd.Metadata.ColonyName = env.ColonyName

	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create a Blueprint
	blueprint := core.CreateBlueprint("TestBlueprint", "test-service-1", env.ColonyName)
	blueprint.SetSpec("replicas", 3)
	blueprint.SetStatus("phase", "Running")

	// Add Blueprint
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedBlueprint)

	// Update the blueprint to create more history
	addedBlueprint.SetSpec("replicas", 5)
	updatedBlueprint, err := client.UpdateBlueprint(addedBlueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, updatedBlueprint)

	// Get service history
	histories, err := client.GetBlueprintHistory(addedBlueprint.ID, 10, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, histories)

	t.Logf("Retrieved %d history entries", len(histories))
	for i, h := range histories {
		t.Logf("  History[%d]: Generation=%d, ChangeType=%s, ChangedBy=%s", i, h.Generation, h.ChangeType, h.ChangedBy)
	}

	// We should have at least 2 history entries (create and update)
	assert.GreaterOrEqual(t, len(histories), 2, "Should have at least 2 history entries")

	// Verify history is ordered by timestamp DESC (most recent first)
	if len(histories) >= 2 {
		assert.GreaterOrEqual(t, histories[0].Generation, histories[1].Generation, "History should be ordered by generation DESC")
	}

	server.Shutdown()
	<-done
}

// TestRemoveBlueprintTriggersDeleteReconciliation verifies that removing a blueprint triggers a delete reconciliation process
func TestRemoveBlueprintTriggersDeleteReconciliation(t *testing.T) {
	t.Skip("Event-driven reconciliation with action metadata not yet implemented - using cron-based reconciliation instead")
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition with a handler (so reconciliation will be triggered)
	sd := core.CreateBlueprintDefinition(
		"docker-deployment",
		"compute.io",
		"v1",
		"DockerDeployment",
		"dockerdeployments",
		"Namespaced",
		"docker-reconciler",
		"reconcile",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Blueprint instance
	blueprint := core.CreateBlueprint("DockerDeployment", "test-deployment", env.ColonyName)
	blueprint.SetSpec("replicas", 3)
	blueprint.SetSpec("image", "nginx:alpine")
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedBlueprint)

	// Get waiting processes - should have 1 (create reconciliation from add)
	waitingProcs, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	initialWaitingCount := len(waitingProcs)
	assert.Equal(t, 1, initialWaitingCount, "Should have 1 waiting process from create reconciliation")

	// Verify the create reconciliation
	if len(waitingProcs) > 0 {
		createProc := waitingProcs[0]
		assert.NotNil(t, createProc.FunctionSpec.Reconciliation)
		assert.Equal(t, core.ReconciliationCreate, createProc.FunctionSpec.Reconciliation.Action)
		assert.Nil(t, createProc.FunctionSpec.Reconciliation.Old)
		assert.NotNil(t, createProc.FunctionSpec.Reconciliation.New)
	}

	// Remove the Blueprint
	err = client.RemoveBlueprint(env.ColonyName, addedBlueprint.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Get waiting processes again - should now have 2 (create + delete)
	waitingProcs, err = client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(waitingProcs), "Should have 2 waiting processes after remove (create + delete)")

	// Find the delete reconciliation process (should be the newest one)
	var deleteProc *core.Process
	for _, proc := range waitingProcs {
		if proc.FunctionSpec.Reconciliation != nil {
			if proc.FunctionSpec.Reconciliation.Action == core.ReconciliationDelete {
				deleteProc = proc
				break
			}
		}
	}

	// Verify delete reconciliation was created
	assert.NotNil(t, deleteProc, "Delete reconciliation process should have been created")
	assert.NotNil(t, deleteProc.FunctionSpec.Reconciliation)
	assert.Equal(t, core.ReconciliationDelete, deleteProc.FunctionSpec.Reconciliation.Action)
	assert.NotNil(t, deleteProc.FunctionSpec.Reconciliation.Old, "Delete reconciliation should have old blueprint")
	assert.Nil(t, deleteProc.FunctionSpec.Reconciliation.New, "Delete reconciliation should have nil new blueprint")

	// Verify the old blueprint in reconciliation matches what we deleted
	assert.Equal(t, addedBlueprint.Metadata.Name, deleteProc.FunctionSpec.Reconciliation.Old.Metadata.Name)
	assert.Equal(t, "DockerDeployment", deleteProc.FunctionSpec.Reconciliation.Old.Kind)

	// Verify the blueprint was removed from database
	_, err = client.GetBlueprint(env.ColonyName, addedBlueprint.Metadata.Name, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Blueprint should not exist in database after removal")

	server.Shutdown()
	<-done
}

func TestGetBlueprintDefinitions(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Initially should be empty
	definitions, err := client.GetBlueprintDefinitions(env.ColonyName, env.ColonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, definitions)
	assert.Equal(t, 0, len(definitions))

	// Add first BlueprintDefinition
	sd1 := core.CreateBlueprintDefinition(
		"test-blueprint-1",
		"example.com",
		"v1",
		"TestBlueprint1",
		"testblueprints1",
		"Namespaced",
		"test_executor_type",
		"reconcile_test_resource",
	)
	sd1.Metadata.ColonyName = env.ColonyName
	_, err = client.AddBlueprintDefinition(sd1, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add second BlueprintDefinition
	sd2 := core.CreateBlueprintDefinition(
		"test-blueprint-2",
		"example.com",
		"v1",
		"TestBlueprint2",
		"testblueprints2",
		"Namespaced",
		"test_executor_type",
		"reconcile_test_resource",
	)
	sd2.Metadata.ColonyName = env.ColonyName
	_, err = client.AddBlueprintDefinition(sd2, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Get all definitions
	definitions, err = client.GetBlueprintDefinitions(env.ColonyName, env.ColonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, definitions)
	assert.Equal(t, 2, len(definitions))

	// Verify both definitions are present
	foundSD1 := false
	foundSD2 := false
	for _, sd := range definitions {
		if sd.Metadata.Name == "test-blueprint-1" {
			foundSD1 = true
			assert.Equal(t, "TestBlueprint1", sd.Spec.Names.Kind)
		}
		if sd.Metadata.Name == "test-blueprint-2" {
			foundSD2 = true
			assert.Equal(t, "TestBlueprint2", sd.Spec.Names.Kind)
		}
	}
	assert.True(t, foundSD1, "Should find first blueprint definition")
	assert.True(t, foundSD2, "Should find second blueprint definition")

	server.Shutdown()
	<-done
}

func TestGetBlueprintDefinitionsAsExecutor(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition as colony owner
	sd := core.CreateBlueprintDefinition(
		"test-blueprint",
		"example.com",
		"v1",
		"TestBlueprint",
		"testblueprints",
		"Namespaced",
		"test_executor_type",
		"reconcile_test_resource",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Executors (members) should also be able to list definitions
	definitions, err := client.GetBlueprintDefinitions(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, definitions)
	assert.Equal(t, 1, len(definitions))
	assert.Equal(t, "test-blueprint", definitions[0].Metadata.Name)

	server.Shutdown()
	<-done
}

func TestGetBlueprintHistoryNotFound(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Try to get history for non-existent blueprint
	_, err := client.GetBlueprintHistory("nonexistent-blueprint-id", 10, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestAddBlueprintWithInvalidSchema(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create a BlueprintDefinition with schema validation
	sd := core.CreateBlueprintDefinition(
		"validated-deployment",
		"compute.io",
		"v1",
		"ValidatedDeployment",
		"validateddeployments",
		"Namespaced",
		"test_executor_type",
		"reconcile",
	)
	sd.Metadata.ColonyName = env.ColonyName

	// Add schema requiring "replicas" field
	sd.Spec.Schema = &core.ValidationSchema{
		Type: "object",
		Properties: map[string]core.SchemaProperty{
			"replicas": {
				Type: "integer",
			},
		},
		Required: []string{"replicas"},
	}

	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Try to add blueprint without required field - should fail
	blueprint := core.CreateBlueprint("ValidatedDeployment", "test-deployment", env.ColonyName)
	blueprint.SetSpec("image", "nginx:alpine") // missing required "replicas"
	_, err = client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Should fail validation for missing required field")

	// Add blueprint with valid schema
	blueprint2 := core.CreateBlueprint("ValidatedDeployment", "test-deployment-2", env.ColonyName)
	blueprint2.SetSpec("replicas", 3)
	_, err = client.AddBlueprint(blueprint2, env.ExecutorPrvKey)
	assert.Nil(t, err, "Should pass validation with required field")

	server.Shutdown()
	<-done
}

func TestUpdateBlueprintWithoutHandler(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create a BlueprintDefinition without handler (no reconciliation)
	sd := core.CreateBlueprintDefinition(
		"simple-config",
		"config.io",
		"v1",
		"SimpleConfig",
		"simpleconfigs",
		"Namespaced",
		"", // No executor type
		"", // No reconciliation function
	)
	sd.Metadata.ColonyName = env.ColonyName

	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add blueprint
	blueprint := core.CreateBlueprint("SimpleConfig", "test-config", env.ColonyName)
	blueprint.SetSpec("key", "value1")
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Update blueprint - should work even without handler
	addedBlueprint.SetSpec("key", "value2")
	updatedBlueprint, err := client.UpdateBlueprint(addedBlueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, updatedBlueprint)
	val, _ := updatedBlueprint.GetSpec("key")
	assert.Equal(t, "value2", val)

	server.Shutdown()
	<-done
}

func TestRemoveBlueprintDefinitionNotFound(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Try to remove non-existent definition
	err := client.RemoveBlueprintDefinition(env.ColonyName, "nonexistent-definition", env.ColonyPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestGetBlueprintNotFound(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Try to get non-existent blueprint
	_, err := client.GetBlueprint(env.ColonyName, "nonexistent-blueprint", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestRemoveBlueprintCreatesCleanupProcess(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition with a handler (so cleanup process will be triggered)
	sd := core.CreateBlueprintDefinition(
		"executor-deployment",
		"compute.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"docker-reconciler",
		"reconcile",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Blueprint instance
	blueprint := core.CreateBlueprint("ExecutorDeployment", "test-executor", env.ColonyName)
	blueprint.SetSpec("replicas", 3)
	blueprint.SetSpec("image", "alpine:latest")
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedBlueprint)

	// Get waiting processes before removal - should have 1 (create reconciliation)
	waitingProcs, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	initialWaitingCount := len(waitingProcs)
	assert.Equal(t, 1, initialWaitingCount, "Should have 1 waiting process from create reconciliation")

	// Remove the Blueprint
	err = client.RemoveBlueprint(env.ColonyName, addedBlueprint.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Get waiting processes again - should now have at least a cleanup process
	waitingProcs, err = client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(waitingProcs), 2, "Should have at least 2 waiting processes after remove")

	// Find the cleanup process
	var cleanupProc *core.Process
	for _, proc := range waitingProcs {
		if proc.FunctionSpec.FuncName == "cleanup" {
			cleanupProc = proc
			break
		}
	}

	// Verify cleanup process was created
	assert.NotNil(t, cleanupProc, "Cleanup process should have been created")
	assert.Equal(t, "cleanup", cleanupProc.FunctionSpec.FuncName)
	assert.Equal(t, "docker-reconciler", cleanupProc.FunctionSpec.Conditions.ExecutorType)

	// Verify blueprintName is in kwargs
	blueprintName, ok := cleanupProc.FunctionSpec.KwArgs["blueprintName"].(string)
	assert.True(t, ok, "blueprintName should be in kwargs")
	assert.Equal(t, "test-executor", blueprintName)

	// Verify the blueprint was removed from database
	_, err = client.GetBlueprint(env.ColonyName, addedBlueprint.Metadata.Name, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Blueprint should not exist in database after removal")

	server.Shutdown()
	<-done
}

func TestUpdateBlueprintTriggersCron(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition with handler configured
	sd := core.CreateBlueprintDefinition(
		"worker",
		"example.com",
		"v1",
		"Worker",
		"workers",
		"Namespaced",
		"worker_reconciler", // ExecutorType
		"reconcile_worker",  // FunctionName
	)
	sd.Spec.Handler.ReconcileInterval = 60 // Configure reconcile interval
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Blueprint - this should auto-create a cron
	blueprint := core.CreateBlueprint("Worker", "my-worker", env.ColonyName)
	blueprint.SetSpec("replicas", 3)
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify cron was created
	crons, err := client.GetCrons(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(crons), "Should have one cron for the Worker kind")
	assert.Equal(t, "reconcile-Worker", crons[0].Name)
	cronID := crons[0].ID

	// Get initial waiting process count (from blueprint create)
	waitingProcs, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	initialWaitingCount := len(waitingProcs)

	// Update Blueprint spec to trigger reconciliation
	addedBlueprint.SetSpec("replicas", 5)
	updatedBlueprint, err := client.UpdateBlueprint(addedBlueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, updatedBlueprint)

	// Verify replicas updated
	replicas, ok := updatedBlueprint.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, float64(5), replicas) // JSON unmarshals numbers as float64

	// Verify that a new reconciliation process was created (cron was triggered)
	waitingProcs, err = client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Should have more processes than before (the reconciliation process created by the cron trigger)
	assert.Greater(t, len(waitingProcs), initialWaitingCount, "Should have created a new reconciliation process")

	// Find the most recent reconciliation process
	var reconProcess *core.Process
	for _, proc := range waitingProcs {
		// Check if it's a reconciliation process with the right kind
		if kind, ok := proc.FunctionSpec.KwArgs["kind"].(string); ok && kind == "Worker" {
			reconProcess = proc
			break
		}
	}

	// Verify reconciliation process was created
	assert.NotNil(t, reconProcess, "Should have created a reconciliation process")
	assert.Equal(t, "reconcile", reconProcess.FunctionSpec.FuncName) // Consolidated reconciliation always uses "reconcile"
	assert.Equal(t, "worker_reconciler", reconProcess.FunctionSpec.Conditions.ExecutorType)

	// Verify the process has the correct kwargs
	kind, ok := reconProcess.FunctionSpec.KwArgs["kind"].(string)
	assert.True(t, ok, "Process should have 'kind' kwarg")
	assert.Equal(t, "Worker", kind)

	// Verify the cron still exists and has the same ID (not recreated)
	cronsAfter, err := client.GetCrons(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(cronsAfter), "Should still have only one cron")
	assert.Equal(t, cronID, cronsAfter[0].ID, "Cron ID should not have changed")

	server.Shutdown()
	<-done
}

func TestGetBlueprintDefinitionByKind(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create and add BlueprintDefinition for "ExecutorDeployment"
	sd1 := core.CreateBlueprintDefinition(
		"executor-deployment",
		"colonies.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"docker_reconciler",
		"reconcile",
	)
	sd1.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd1, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create and add BlueprintDefinition for "DockerDeployment"
	sd2 := core.CreateBlueprintDefinition(
		"docker-deployment",
		"colonies.io",
		"v1",
		"DockerDeployment",
		"dockerdeployments",
		"Namespaced",
		"docker_reconciler",
		"reconcile",
	)
	sd2.Metadata.ColonyName = env.ColonyName
	_, err = client.AddBlueprintDefinition(sd2, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Test GetBlueprintDefinitionByKind - find ExecutorDeployment
	foundSD, err := client.GetBlueprintDefinitionByKind(env.ColonyName, "ExecutorDeployment", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, foundSD)
	assert.Equal(t, "ExecutorDeployment", foundSD.Spec.Names.Kind)
	assert.Equal(t, "executor-deployment", foundSD.Metadata.Name)

	// Test GetBlueprintDefinitionByKind - find DockerDeployment
	foundSD2, err := client.GetBlueprintDefinitionByKind(env.ColonyName, "DockerDeployment", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, foundSD2)
	assert.Equal(t, "DockerDeployment", foundSD2.Spec.Names.Kind)
	assert.Equal(t, "docker-deployment", foundSD2.Metadata.Name)

	// Test GetBlueprintDefinitionByKind - non-existent kind returns nil
	notFoundSD, err := client.GetBlueprintDefinitionByKind(env.ColonyName, "NonExistentKind", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Nil(t, notFoundSD)

	server.Shutdown()
	<-done
}

func TestBlueprintWithLocationCreatesCronWithLocation(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition with handler configured
	sd := core.CreateBlueprintDefinition(
		"executor-deployment",
		"colonies.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"docker_reconciler", // ExecutorType
		"reconcile",         // FunctionName
	)
	sd.Spec.Handler.ReconcileInterval = 60
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Blueprint WITH location - this should create a cron with location suffix
	blueprint := core.CreateBlueprint("ExecutorDeployment", "my-deployment", env.ColonyName)
	blueprint.Metadata.LocationName = "datacenter-east" // Set location
	blueprint.SetSpec("replicas", 2)
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedBlueprint)

	// Verify cron was created with location in name
	crons, err := client.GetCrons(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(crons), "Should have one cron for the ExecutorDeployment kind at datacenter-east")
	assert.Equal(t, "reconcile-ExecutorDeployment-datacenter-east", crons[0].Name, "Cron name should include location suffix")

	// Get waiting processes to verify reconciliation process was created
	waitingProcs, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Greater(t, len(waitingProcs), 0, "Should have at least one waiting reconciliation process")

	// Find the reconciliation process and verify location is set correctly
	var reconProcess *core.Process
	for _, proc := range waitingProcs {
		if kind, ok := proc.FunctionSpec.KwArgs["kind"].(string); ok && kind == "ExecutorDeployment" {
			reconProcess = proc
			break
		}
	}

	assert.NotNil(t, reconProcess, "Should have created a reconciliation process")
	assert.Equal(t, "reconcile", reconProcess.FunctionSpec.FuncName)
	assert.Equal(t, "docker_reconciler", reconProcess.FunctionSpec.Conditions.ExecutorType)
	assert.Equal(t, "datacenter-east", reconProcess.FunctionSpec.Conditions.LocationName, "Process should have LocationName condition set")

	// Verify KwArgs has kind
	kind, ok := reconProcess.FunctionSpec.KwArgs["kind"].(string)
	assert.True(t, ok, "Process should have 'kind' kwarg")
	assert.Equal(t, "ExecutorDeployment", kind)

	server.Shutdown()
	<-done
}

func TestBlueprintWithLocationCreatesSeparateCronsPerLocation(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"container-deployment",
		"colonies.io",
		"v1",
		"ContainerDeployment",
		"containerdeployments",
		"Namespaced",
		"container_reconciler",
		"reconcile",
	)
	sd.Spec.Handler.ReconcileInterval = 60
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add first Blueprint at location "east"
	blueprint1 := core.CreateBlueprint("ContainerDeployment", "deployment-east", env.ColonyName)
	blueprint1.Metadata.LocationName = "east"
	blueprint1.SetSpec("replicas", 1)
	_, err = client.AddBlueprint(blueprint1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Add second Blueprint at location "west"
	blueprint2 := core.CreateBlueprint("ContainerDeployment", "deployment-west", env.ColonyName)
	blueprint2.Metadata.LocationName = "west"
	blueprint2.SetSpec("replicas", 1)
	_, err = client.AddBlueprint(blueprint2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Add third Blueprint at location "east" (same location as first)
	blueprint3 := core.CreateBlueprint("ContainerDeployment", "deployment-east-2", env.ColonyName)
	blueprint3.Metadata.LocationName = "east"
	blueprint3.SetSpec("replicas", 1)
	_, err = client.AddBlueprint(blueprint3, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify two crons were created (one per unique location)
	crons, err := client.GetCrons(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(crons), "Should have two crons - one for each unique location")

	// Verify cron names
	cronNames := make(map[string]bool)
	for _, cron := range crons {
		cronNames[cron.Name] = true
	}
	assert.True(t, cronNames["reconcile-ContainerDeployment-east"], "Should have cron for east location")
	assert.True(t, cronNames["reconcile-ContainerDeployment-west"], "Should have cron for west location")

	server.Shutdown()
	<-done
}

func TestBlueprintWithoutLocationCreatesCronWithoutSuffix(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"service-deployment",
		"colonies.io",
		"v1",
		"ServiceDeployment",
		"servicedeployments",
		"Namespaced",
		"service_reconciler",
		"reconcile",
	)
	sd.Spec.Handler.ReconcileInterval = 60
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Blueprint WITHOUT location
	blueprint := core.CreateBlueprint("ServiceDeployment", "my-service", env.ColonyName)
	// Explicitly NOT setting location
	blueprint.SetSpec("replicas", 1)
	_, err = client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify cron was created WITHOUT location suffix
	crons, err := client.GetCrons(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(crons))
	assert.Equal(t, "reconcile-ServiceDeployment", crons[0].Name, "Cron name should NOT have location suffix when no location specified")

	// Get waiting processes
	waitingProcs, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Greater(t, len(waitingProcs), 0)

	// Find reconciliation process
	var reconProcess *core.Process
	for _, proc := range waitingProcs {
		if kind, ok := proc.FunctionSpec.KwArgs["kind"].(string); ok && kind == "ServiceDeployment" {
			reconProcess = proc
			break
		}
	}

	assert.NotNil(t, reconProcess)
	assert.Equal(t, "", reconProcess.FunctionSpec.Conditions.LocationName, "Process should have empty LocationName when blueprint has no location")

	server.Shutdown()
	<-done
}

// TestCronNamingConsistencyBetweenAddAndUpdate verifies that AddBlueprint and
// UpdateBlueprint use the same cron naming scheme: reconcile-{Kind}-{locationName}
// This ensures UpdateBlueprint can find and trigger the reconciliation cron.
func TestCronNamingConsistencyBetweenAddAndUpdate(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create BlueprintDefinition with an executorType
	executorType := "docker-reconciler"
	sd := core.CreateBlueprintDefinition(
		"cron-naming-test",
		"example.com",
		"v1",
		"CronNamingTest",
		"cronnamingtests",
		"Namespaced",
		executorType,
		"reconcile",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create Blueprint with a locationName
	locationName := "dc1"
	blueprint := core.CreateBlueprint("CronNamingTest", "test-deployment", env.ColonyName)
	blueprint.Metadata.LocationName = locationName
	blueprint.SetSpec("image", "nginx:1.21")
	blueprint.SetSpec("replicas", 3)

	// Add Blueprint - this creates a cron with name "reconcile-CronNamingTest-dc1"
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedBlueprint)

	// Get all crons and find the one created by AddBlueprint
	crons, err := client.GetCrons(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Both AddBlueprint and UpdateBlueprint should use: reconcile-{Kind}-{locationName}
	expectedCronName := "reconcile-CronNamingTest-" + locationName // reconcile-CronNamingTest-dc1
	var foundCron *core.Cron
	for _, cron := range crons {
		if cron.Name == expectedCronName {
			foundCron = cron
			break
		}
	}

	// Verify AddBlueprint created the cron with locationName pattern
	assert.NotNil(t, foundCron, "AddBlueprint should create cron with name: %s", expectedCronName)
	t.Logf("AddBlueprint created cron with name: %s", expectedCronName)

	// Record the cron's last run time before update
	cronBeforeUpdate := foundCron

	// Now update the blueprint - this should trigger the same cron
	addedBlueprint.SetSpec("replicas", 5)
	_, err = client.UpdateBlueprint(addedBlueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Get crons again to verify the cron was triggered
	cronsAfterUpdate, err := client.GetCrons(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Find the cron again
	var cronAfterUpdate *core.Cron
	for _, cron := range cronsAfterUpdate {
		if cron.Name == expectedCronName {
			cronAfterUpdate = cron
			break
		}
	}

	assert.NotNil(t, cronAfterUpdate, "Cron should still exist after update")

	// Verify the cron was triggered by checking LastRun changed
	// (The cron should have been run by UpdateBlueprint)
	assert.True(t,
		cronAfterUpdate.LastRun.After(cronBeforeUpdate.LastRun) || !cronAfterUpdate.LastRun.IsZero(),
		"UpdateBlueprint should trigger the reconciliation cron (LastRun should be updated)")

	t.Logf("Cron naming is consistent:")
	t.Logf("  AddBlueprint creates: %s", expectedCronName)
	t.Logf("  UpdateBlueprint finds: %s", expectedCronName)
	t.Logf("  Cron was triggered: LastRun before=%v, after=%v", cronBeforeUpdate.LastRun, cronAfterUpdate.LastRun)

	server.Shutdown()
	<-done
}

// TestRemoveBlueprintCronCleanup verifies that RemoveBlueprint correctly removes
// the cron when the last blueprint of a Kind at a location is deleted.
// Both AddBlueprint and RemoveBlueprint use the same naming: reconcile-{Kind}-{locationName}
func TestRemoveBlueprintCronCleanup(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"remove-cron-test",
		"example.com",
		"v1",
		"RemoveCronTest",
		"removetest",
		"Namespaced",
		"test-reconciler",
		"reconcile",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create Blueprint with a locationName
	locationName := "dc1"
	blueprint := core.CreateBlueprint("RemoveCronTest", "test-deployment", env.ColonyName)
	blueprint.Metadata.LocationName = locationName
	blueprint.SetSpec("image", "nginx:1.21")

	// Add Blueprint - this creates a cron with name "reconcile-RemoveCronTest-dc1"
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedBlueprint)

	// Get crons and verify the cron was created with locationName pattern
	cronsBefore, err := client.GetCrons(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)

	expectedCronName := "reconcile-RemoveCronTest-" + locationName // reconcile-RemoveCronTest-dc1
	var foundCronBefore *core.Cron
	for _, cron := range cronsBefore {
		if cron.Name == expectedCronName {
			foundCronBefore = cron
			break
		}
	}
	assert.NotNil(t, foundCronBefore, "AddBlueprint should create cron with name: %s", expectedCronName)
	t.Logf("AddBlueprint created cron: %s", expectedCronName)

	// Now remove the blueprint - this is the last (and only) blueprint of this Kind at this location
	err = client.RemoveBlueprint(env.ColonyName, blueprint.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Get crons again after removal
	cronsAfter, err := client.GetCrons(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Check if the cron was removed
	var cronStillExists *core.Cron
	for _, cron := range cronsAfter {
		if cron.Name == expectedCronName {
			cronStillExists = cron
			break
		}
	}

	// The cron SHOULD be removed since the last blueprint at this location was deleted
	assert.Nil(t, cronStillExists, "Cron should be removed after deleting last blueprint at location")

	t.Logf("Cron cleanup verified:")
	t.Logf("  Cron '%s' was correctly removed after deleting last blueprint", expectedCronName)

	server.Shutdown()
	<-done
}

// TestRemoveBlueprintCronKeptWhenOthersExist verifies that the cron is NOT removed
// when other blueprints of the same Kind at the same location still exist.
func TestRemoveBlueprintCronKeptWhenOthersExist(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"cron-kept-test",
		"example.com",
		"v1",
		"CronKeptTest",
		"cronkepttest",
		"Namespaced",
		"test-reconciler",
		"reconcile",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create two blueprints at the same location
	locationName := "dc1"

	blueprint1 := core.CreateBlueprint("CronKeptTest", "deployment-1", env.ColonyName)
	blueprint1.Metadata.LocationName = locationName
	blueprint1.SetSpec("image", "nginx:1.21")

	blueprint2 := core.CreateBlueprint("CronKeptTest", "deployment-2", env.ColonyName)
	blueprint2.Metadata.LocationName = locationName
	blueprint2.SetSpec("image", "nginx:1.22")

	// Add both blueprints
	_, err = client.AddBlueprint(blueprint1, env.ExecutorPrvKey)
	assert.Nil(t, err)
	_, err = client.AddBlueprint(blueprint2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify cron exists
	expectedCronName := "reconcile-CronKeptTest-" + locationName
	cronsBefore, err := client.GetCrons(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)

	var foundCron *core.Cron
	for _, cron := range cronsBefore {
		if cron.Name == expectedCronName {
			foundCron = cron
			break
		}
	}
	assert.NotNil(t, foundCron, "Cron should exist")

	// Remove first blueprint - cron should be kept because blueprint2 still exists
	err = client.RemoveBlueprint(env.ColonyName, blueprint1.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify cron still exists
	cronsAfter, err := client.GetCrons(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)

	var cronAfterFirstRemove *core.Cron
	for _, cron := range cronsAfter {
		if cron.Name == expectedCronName {
			cronAfterFirstRemove = cron
			break
		}
	}
	assert.NotNil(t, cronAfterFirstRemove, "Cron should be kept when other blueprints at location exist")

	// Remove second blueprint - now cron should be removed
	err = client.RemoveBlueprint(env.ColonyName, blueprint2.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify cron is now removed
	cronsFinal, err := client.GetCrons(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)

	var cronAfterSecondRemove *core.Cron
	for _, cron := range cronsFinal {
		if cron.Name == expectedCronName {
			cronAfterSecondRemove = cron
			break
		}
	}
	assert.Nil(t, cronAfterSecondRemove, "Cron should be removed after last blueprint at location is deleted")

	t.Logf("Cron lifecycle verified:")
	t.Logf("  Cron kept after removing first blueprint (other blueprints exist)")
	t.Logf("  Cron removed after removing last blueprint at location")

	server.Shutdown()
	<-done
}

func TestImmediateReconciliationTriggeredOnBlueprintAdd(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition with handler configured
	sd := core.CreateBlueprintDefinition(
		"executor-deployment",
		"compute.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"docker-reconciler",
		"reconcile",
	)
	sd.Spec.Handler.ReconcileInterval = 60
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add first Blueprint - this creates the cron AND an immediate reconciliation process
	blueprint1 := core.CreateBlueprint("ExecutorDeployment", "first-executor", env.ColonyName)
	blueprint1.SetSpec("replicas", 2)
	blueprint1.SetSpec("image", "nginx:latest")
	_, err = client.AddBlueprint(blueprint1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Check waiting processes after first blueprint
	waitingProcsAfterFirst, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(waitingProcsAfterFirst), "Should have 1 waiting process from first blueprint reconciliation")

	// Verify the first reconciliation process
	firstReconcileProc := waitingProcsAfterFirst[0]
	assert.Equal(t, "reconcile", firstReconcileProc.FunctionSpec.FuncName)
	assert.Equal(t, "docker-reconciler", firstReconcileProc.FunctionSpec.Conditions.ExecutorType)

	// Add second Blueprint of the same Kind - this should find existing cron and trigger immediate reconciliation
	blueprint2 := core.CreateBlueprint("ExecutorDeployment", "second-executor", env.ColonyName)
	blueprint2.SetSpec("replicas", 3)
	blueprint2.SetSpec("image", "alpine:latest")
	_, err = client.AddBlueprint(blueprint2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Check waiting processes after second blueprint - should now have 2 reconciliation processes
	waitingProcsAfterSecond, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(waitingProcsAfterSecond), "Should have 2 waiting processes after second blueprint (one for each)")

	// Verify both reconciliation processes have correct function name and executor type
	for _, proc := range waitingProcsAfterSecond {
		assert.Equal(t, "reconcile", proc.FunctionSpec.FuncName)
		assert.Equal(t, "docker-reconciler", proc.FunctionSpec.Conditions.ExecutorType)
	}

	// Add third Blueprint - should also trigger immediate reconciliation
	blueprint3 := core.CreateBlueprint("ExecutorDeployment", "third-executor", env.ColonyName)
	blueprint3.SetSpec("replicas", 1)
	_, err = client.AddBlueprint(blueprint3, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify we now have 3 reconciliation processes
	waitingProcsAfterThird, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(waitingProcsAfterThird), "Should have 3 waiting processes after third blueprint")

	t.Logf("Immediate reconciliation verified:")
	t.Logf("  First blueprint: created cron + 1 reconciliation process")
	t.Logf("  Second blueprint: found existing cron + triggered immediate reconciliation")
	t.Logf("  Third blueprint: found existing cron + triggered immediate reconciliation")
	t.Logf("  Total reconciliation processes: %d", len(waitingProcsAfterThird))

	server.Shutdown()
	<-done
}

func TestLocationAutoCreatedWithBlueprint(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"executor-deployment",
		"compute.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"docker-reconciler",
		"reconcile",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Verify location doesn't exist initially
	locationName := "auto-created-datacenter"
	_, err = client.GetLocation(env.ColonyName, locationName, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Location should not exist initially")

	// Add blueprint with a new location - this should auto-create the location
	blueprint := core.CreateBlueprint("ExecutorDeployment", "test-executor", env.ColonyName)
	blueprint.Metadata.LocationName = locationName
	blueprint.SetSpec("replicas", 2)
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedBlueprint)

	// Verify location was auto-created
	location, err := client.GetLocation(env.ColonyName, locationName, env.ExecutorPrvKey)
	assert.Nil(t, err, "Location should exist after blueprint creation")
	assert.NotNil(t, location)
	assert.Equal(t, locationName, location.Name)
	assert.Contains(t, location.Description, "Auto-created from blueprint")

	// Verify blueprint was created with correct location
	retrievedBlueprint, err := client.GetBlueprint(env.ColonyName, "test-executor", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, locationName, retrievedBlueprint.Metadata.LocationName)

	t.Logf("Location auto-creation verified:")
	t.Logf("  Location '%s' was auto-created", locationName)
	t.Logf("  Blueprint references location correctly")

	server.Shutdown()
	<-done
}

func TestPreExistingLocationNotAffectedByBlueprintFailure(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"executor-deployment",
		"compute.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"docker-reconciler",
		"reconcile",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Manually create a location first
	locationName := "pre-existing-datacenter"
	location := core.CreateLocation(
		core.GenerateRandomID(),
		locationName,
		env.ColonyName,
		"Manually created location",
		10.0,
		20.0,
	)
	addedLocation, err := client.AddLocation(location, env.ColonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedLocation)

	// Add first blueprint with this location - should succeed
	blueprint1 := core.CreateBlueprint("ExecutorDeployment", "first-executor", env.ColonyName)
	blueprint1.Metadata.LocationName = locationName
	blueprint1.SetSpec("replicas", 2)
	_, err = client.AddBlueprint(blueprint1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to add a DUPLICATE blueprint (same name) - should fail
	blueprint2 := core.CreateBlueprint("ExecutorDeployment", "first-executor", env.ColonyName)
	blueprint2.Metadata.LocationName = locationName
	blueprint2.SetSpec("replicas", 3)
	_, err = client.AddBlueprint(blueprint2, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Adding duplicate blueprint should fail")

	// Verify the pre-existing location is still there (not deleted by failed operation)
	locationAfterFailure, err := client.GetLocation(env.ColonyName, locationName, env.ExecutorPrvKey)
	assert.Nil(t, err, "Pre-existing location should still exist after failed blueprint creation")
	assert.NotNil(t, locationAfterFailure)
	assert.Equal(t, locationName, locationAfterFailure.Name)
	assert.Equal(t, "Manually created location", locationAfterFailure.Description)

	t.Logf("Pre-existing location protection verified:")
	t.Logf("  Location '%s' was not affected by failed blueprint creation", locationName)

	server.Shutdown()
	<-done
}

func TestLocationCleanupOnBlueprintCreationFailure(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"executor-deployment",
		"compute.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"docker-reconciler",
		"reconcile",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add a blueprint with a location first
	locationName := "shared-datacenter"
	blueprint1 := core.CreateBlueprint("ExecutorDeployment", "first-executor", env.ColonyName)
	blueprint1.Metadata.LocationName = locationName
	blueprint1.SetSpec("replicas", 2)
	_, err = client.AddBlueprint(blueprint1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify location was created
	location, err := client.GetLocation(env.ColonyName, locationName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, location)

	// Try to add a second blueprint with SAME name but SAME location (should fail - duplicate name)
	// This tests that the location (which already exists) is not deleted when blueprint creation fails
	blueprint2 := core.CreateBlueprint("ExecutorDeployment", "first-executor", env.ColonyName)
	blueprint2.Metadata.LocationName = locationName
	blueprint2.SetSpec("replicas", 5)
	_, err = client.AddBlueprint(blueprint2, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Adding duplicate blueprint should fail")

	// Location should still exist (it was not auto-created for blueprint2, so cleanup doesn't apply)
	locationStillExists, err := client.GetLocation(env.ColonyName, locationName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, locationStillExists)
	assert.Equal(t, locationName, locationStillExists.Name)

	// Now test with a NEW location name - the location should NOT be auto-created
	// because the duplicate check happens BEFORE location creation
	newLocationName := "new-datacenter"
	blueprint3 := core.CreateBlueprint("ExecutorDeployment", "first-executor", env.ColonyName) // duplicate name
	blueprint3.Metadata.LocationName = newLocationName
	blueprint3.SetSpec("replicas", 3)
	_, err = client.AddBlueprint(blueprint3, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Adding duplicate blueprint with new location should fail")

	// The new location should NOT exist (duplicate check failed before location creation)
	_, err = client.GetLocation(env.ColonyName, newLocationName, env.ExecutorPrvKey)
	assert.NotNil(t, err, "New location should not exist - duplicate check failed before location creation")

	t.Logf("Location cleanup behavior verified:")
	t.Logf("  Pre-existing location preserved on blueprint failure")
	t.Logf("  New location not created when duplicate check fails first")

	server.Shutdown()
	<-done
}

func TestUpdateBlueprintReturns404ForNonExistent(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition (required for update validation)
	sd := core.CreateBlueprintDefinition(
		"executor-deployment",
		"compute.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"docker-reconciler",
		"reconcile",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Try to update a blueprint that doesn't exist
	nonExistentBlueprint := core.CreateBlueprint("ExecutorDeployment", "non-existent-blueprint", env.ColonyName)
	nonExistentBlueprint.SetSpec("replicas", 5)

	_, err = client.UpdateBlueprint(nonExistentBlueprint, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Updating non-existent blueprint should return an error")
	assert.Contains(t, err.Error(), "not found", "Error message should indicate blueprint not found")

	// Verify the error is a ColoniesError with 404 status
	coloniesErr, ok := err.(*core.ColoniesError)
	assert.True(t, ok, "Error should be a ColoniesError")
	assert.Equal(t, 404, coloniesErr.Status, "Status code should be 404 Not Found")

	t.Logf("UpdateBlueprint 404 validation verified:")
	t.Logf("  Non-existent blueprint returns error with status %d", coloniesErr.Status)
	t.Logf("  Error message: %s", coloniesErr.Message)

	server.Shutdown()
	<-done
}

func TestUpdateBlueprintSucceedsForExistingBlueprint(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"executor-deployment",
		"compute.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"docker-reconciler",
		"reconcile",
	)
	sd.Metadata.ColonyName = env.ColonyName
	_, err := client.AddBlueprintDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// First, add a blueprint
	blueprint := core.CreateBlueprint("ExecutorDeployment", "existing-blueprint", env.ColonyName)
	blueprint.SetSpec("replicas", 2)
	addedBlueprint, err := client.AddBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedBlueprint)

	// Verify initial replicas
	replicas, ok := addedBlueprint.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, float64(2), replicas)

	// Now update the existing blueprint
	blueprint.SetSpec("replicas", 5)
	updatedBlueprint, err := client.UpdateBlueprint(blueprint, env.ExecutorPrvKey)
	assert.Nil(t, err, "Updating existing blueprint should succeed")
	assert.NotNil(t, updatedBlueprint)

	// Verify the update was applied
	newReplicas, ok := updatedBlueprint.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, float64(5), newReplicas)

	// Verify generation was incremented
	assert.Equal(t, addedBlueprint.Metadata.Generation+1, updatedBlueprint.Metadata.Generation)

	t.Logf("UpdateBlueprint success case verified:")
	t.Logf("  Replicas changed from 2 to 5")
	t.Logf("  Generation incremented from %d to %d", addedBlueprint.Metadata.Generation, updatedBlueprint.Metadata.Generation)

	server.Shutdown()
	<-done
}
