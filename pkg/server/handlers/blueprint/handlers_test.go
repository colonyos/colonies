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
	sd.Metadata.Namespace = env.ColonyName

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
	sd.Metadata.Namespace = env.ColonyName

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
	sd.Metadata.Namespace = env.ColonyName

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
	sd.Metadata.Namespace = env.ColonyName
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
	sdDB.Metadata.Namespace = env.ColonyName
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
	sdSvc.Metadata.Namespace = env.ColonyName
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
	sd.Metadata.Namespace = env.ColonyName
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
	sd.Metadata.Namespace = env.ColonyName
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
	sd.Metadata.Namespace = env.ColonyName
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
	sd.Metadata.Namespace = env.ColonyName
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
	sd.Metadata.Namespace = env.ColonyName
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
	sd.Metadata.Namespace = env.ColonyName

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
	sd.Metadata.Namespace = env.ColonyName
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
	sd1.Metadata.Namespace = env.ColonyName
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
	sd2.Metadata.Namespace = env.ColonyName
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
	sd.Metadata.Namespace = env.ColonyName
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
	sd.Metadata.Namespace = env.ColonyName

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
	sd.Metadata.Namespace = env.ColonyName

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
