package resource_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/stretchr/testify/assert"
)

func TestAddResourceDefinition(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create a ResourceDefinition
	rd := core.CreateResourceDefinition(
		"test-resource",
		"example.com",
		"v1",
		"TestResource",
		"testresources",
		"Namespaced",
		"test_executor_type",
		"reconcile_test_resource",
	)
	rd.Metadata.Namespace = env.ColonyName

	// Add ResourceDefinition with colony owner key
	addedRD, err := client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedRD)
	assert.Equal(t, rd.Metadata.Name, addedRD.Metadata.Name)
	assert.Equal(t, rd.Spec.Group, addedRD.Spec.Group)
	assert.Equal(t, rd.Spec.Version, addedRD.Spec.Version)

	// Try to add duplicate ResourceDefinition - should fail
	_, err = client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestGetResourceDefinition(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create and add ResourceDefinition
	rd := core.CreateResourceDefinition(
		"test-resource",
		"example.com",
		"v1",
		"TestResource",
		"testresources",
		"Namespaced",
		"test_executor_type",
		"reconcile_test_resource",
	)
	rd.Metadata.Namespace = env.ColonyName

	addedRD, err := client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Get ResourceDefinition (using executor key since only members can get)
	retrievedRD, err := client.GetResourceDefinition(env.ColonyName, rd.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedRD)
	assert.Equal(t, addedRD.ID, retrievedRD.ID)
	assert.Equal(t, addedRD.Metadata.Name, retrievedRD.Metadata.Name)

	server.Shutdown()
	<-done
}

func TestAddResource(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// First add a ResourceDefinition
	rd := core.CreateResourceDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"database_controller",
		"reconcile_database",
	)
	rd.Metadata.Namespace = env.ColonyName

	_, err := client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create a Resource instance
	resource := core.CreateResource("Database", "test-database", env.ColonyName)
	resource.SetSpec("host", "localhost")
	resource.SetSpec("port", 5432)

	// Add Resource
	addedResource, err := client.AddResource(resource, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedResource)
	assert.Equal(t, resource.Metadata.Name, addedResource.Metadata.Name)
	assert.Equal(t, resource.Kind, addedResource.Kind)

	// Verify spec was preserved
	host, ok := addedResource.GetSpec("host")
	assert.True(t, ok)
	assert.Equal(t, "localhost", host)

	server.Shutdown()
	<-done
}

func TestGetResource(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add ResourceDefinition
	rd := core.CreateResourceDefinition(
		"service",
		"example.com",
		"v1",
		"Service",
		"services",
		"Namespaced",
		"service_controller",
		"reconcile_service",
	)
	rd.Metadata.Namespace = env.ColonyName
	_, err := client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Resource
	resource := core.CreateResource("Service", "web-service", env.ColonyName)
	resource.SetSpec("port", 8080)
	addedResource, err := client.AddResource(resource, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Get Resource
	retrievedResource, err := client.GetResource(env.ColonyName, resource.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedResource)
	assert.Equal(t, addedResource.ID, retrievedResource.ID)
	assert.Equal(t, addedResource.Metadata.Name, retrievedResource.Metadata.Name)

	server.Shutdown()
	<-done
}

func TestGetResources(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add ResourceDefinition for Database
	rdDB := core.CreateResourceDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"database_controller",
		"reconcile_database",
	)
	rdDB.Metadata.Namespace = env.ColonyName
	_, err := client.AddResourceDefinition(rdDB, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add ResourceDefinition for Service
	rdSvc := core.CreateResourceDefinition(
		"service",
		"example.com",
		"v1",
		"Service",
		"services",
		"Namespaced",
		"service_controller",
		"reconcile_service",
	)
	rdSvc.Metadata.Namespace = env.ColonyName
	_, err = client.AddResourceDefinition(rdSvc, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add some Database resources
	db1 := core.CreateResource("Database", "db1", env.ColonyName)
	db2 := core.CreateResource("Database", "db2", env.ColonyName)
	_, err = client.AddResource(db1, env.ExecutorPrvKey)
	assert.Nil(t, err)
	_, err = client.AddResource(db2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Add some Service resources
	svc1 := core.CreateResource("Service", "svc1", env.ColonyName)
	_, err = client.AddResource(svc1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Get all resources
	allResources, err := client.GetResources(env.ColonyName, "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(allResources))

	// Get only Database resources
	dbResources, err := client.GetResources(env.ColonyName, "Database", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(dbResources))

	// Get only Service resources
	svcResources, err := client.GetResources(env.ColonyName, "Service", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(svcResources))

	server.Shutdown()
	<-done
}

func TestUpdateResource(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add ResourceDefinition
	rd := core.CreateResourceDefinition(
		"application",
		"example.com",
		"v1",
		"Application",
		"applications",
		"Namespaced",
		"app_controller",
		"reconcile_application",
	)
	rd.Metadata.Namespace = env.ColonyName
	_, err := client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Resource
	resource := core.CreateResource("Application", "my-app", env.ColonyName)
	resource.SetSpec("version", "1.0.0")
	addedResource, err := client.AddResource(resource, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Update Resource
	addedResource.SetSpec("version", "1.1.0")
	updatedResource, err := client.UpdateResource(addedResource, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, updatedResource)

	version, ok := updatedResource.GetSpec("version")
	assert.True(t, ok)
	assert.Equal(t, "1.1.0", version)

	server.Shutdown()
	<-done
}

func TestRemoveResource(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add ResourceDefinition
	rd := core.CreateResourceDefinition(
		"cache",
		"example.com",
		"v1",
		"Cache",
		"caches",
		"Namespaced",
		"cache_controller",
		"reconcile_cache",
	)
	rd.Metadata.Namespace = env.ColonyName
	_, err := client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Resource
	resource := core.CreateResource("Cache", "redis-cache", env.ColonyName)
	addedResource, err := client.AddResource(resource, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Remove Resource
	err = client.RemoveResource(env.ColonyName, addedResource.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify it's removed
	_, err = client.GetResource(env.ColonyName, addedResource.Metadata.Name, env.ExecutorPrvKey)
	assert.NotNil(t, err) // Should fail because resource doesn't exist

	server.Shutdown()
	<-done
}

func TestResourceWithComplexSpec(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add ResourceDefinition
	rd := core.CreateResourceDefinition(
		"deployment",
		"compute.io",
		"v1",
		"Deployment",
		"deployments",
		"Namespaced",
		"deployment_controller",
		"reconcile_deployment",
	)
	rd.Metadata.Namespace = env.ColonyName
	_, err := client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create Resource with complex spec
	resource := core.CreateResource("Deployment", "web-deployment", env.ColonyName)
	resource.SetSpec("image", "nginx:1.21")
	resource.SetSpec("replicas", 3)
	resource.SetSpec("env", map[string]interface{}{
		"DATABASE_URL": "postgres://localhost/db",
		"PORT":         "8080",
	})
	resource.Metadata.Labels = map[string]string{
		"app":     "web",
		"version": "v1.0.0",
	}
	resource.Metadata.Annotations = map[string]string{
		"description": "My test application",
	}

	// Add Resource
	_, err = client.AddResource(resource, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Retrieve and verify
	retrievedResource, err := client.GetResource(env.ColonyName, resource.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	image, ok := retrievedResource.GetSpec("image")
	assert.True(t, ok)
	assert.Equal(t, "nginx:1.21", image)

	replicas, ok := retrievedResource.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, float64(3), replicas) // JSON unmarshaling converts to float64

	envSpec, ok := retrievedResource.GetSpec("env")
	assert.True(t, ok)
	envMap := envSpec.(map[string]interface{})
	assert.Equal(t, "postgres://localhost/db", envMap["DATABASE_URL"])

	assert.Equal(t, "v1.0.0", retrievedResource.Metadata.Labels["version"])
	assert.Equal(t, "My test application", retrievedResource.Metadata.Annotations["description"])

	server.Shutdown()
	<-done
}

// TestAddResourceRequiresResourceDefinition tests that adding a resource without a ResourceDefinition fails
func TestAddResourceRequiresResourceDefinition(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Try to add a Resource WITHOUT adding its ResourceDefinition first
	resource := core.CreateResource("NonExistentKind", "test-resource", env.ColonyName)
	resource.SetSpec("field", "value")

	// This should fail because ResourceDefinition doesn't exist
	_, err := client.AddResource(resource, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Adding resource without ResourceDefinition should fail")
	assert.Contains(t, err.Error(), "ResourceDefinition for kind 'NonExistentKind' not found")

	server.Shutdown()
	<-done
}

// TestAddResourceWithSchemaValidation tests that resources are validated against the ResourceDefinition schema
func TestAddResourceWithSchemaValidation(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create ResourceDefinition with schema validation
	rd := core.CreateResourceDefinition(
		"validated-resource",
		"example.com",
		"v1",
		"ValidatedResource",
		"validatedresources",
		"Namespaced",
		"validator_controller",
		"reconcile_validated",
	)
	rd.Metadata.Namespace = env.ColonyName
	rd.Spec.Schema = &core.ValidationSchema{
		Type: "object",
		Properties: map[string]core.SchemaProperty{
			"name": {
				Type:        "string",
				Description: "Resource name",
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

	_, err := client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Test 1: Valid resource should succeed
	validResource := core.CreateResource("ValidatedResource", "valid-res", env.ColonyName)
	validResource.SetSpec("name", "test")
	validResource.SetSpec("replicas", 3)
	validResource.SetSpec("protocol", "TCP")

	addedResource, err := client.AddResource(validResource, env.ExecutorPrvKey)
	assert.Nil(t, err, "Valid resource should be added successfully")
	assert.NotNil(t, addedResource)

	// Test 2: Resource missing required field should fail
	invalidResource1 := core.CreateResource("ValidatedResource", "invalid-res-1", env.ColonyName)
	invalidResource1.SetSpec("name", "test") // Missing required 'replicas'

	_, err = client.AddResource(invalidResource1, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Resource missing required field should fail")
	assert.Contains(t, err.Error(), "required field 'replicas' is missing")

	// Test 3: Resource with invalid type should fail
	invalidResource2 := core.CreateResource("ValidatedResource", "invalid-res-2", env.ColonyName)
	invalidResource2.SetSpec("name", "test")
	invalidResource2.SetSpec("replicas", "not-a-number") // Should be number

	_, err = client.AddResource(invalidResource2, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Resource with invalid type should fail")
	assert.Contains(t, err.Error(), "must be a number")

	// Test 4: Resource with invalid enum value should fail
	invalidResource3 := core.CreateResource("ValidatedResource", "invalid-res-3", env.ColonyName)
	invalidResource3.SetSpec("name", "test")
	invalidResource3.SetSpec("replicas", 3)
	invalidResource3.SetSpec("protocol", "HTTP") // Not in enum [TCP, UDP]

	_, err = client.AddResource(invalidResource3, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Resource with invalid enum value should fail")
	assert.Contains(t, err.Error(), "must be one of")

	server.Shutdown()
	<-done
}

// TestRemoveResourceDefinitionWithActiveResources tests that removing a ResourceDefinition with active resources fails
func TestRemoveResourceDefinitionWithActiveResources(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add ResourceDefinition
	rd := core.CreateResourceDefinition(
		"protected-resource",
		"example.com",
		"v1",
		"ProtectedResource",
		"protectedresources",
		"Namespaced",
		"protected_controller",
		"reconcile_protected",
	)
	rd.Metadata.Namespace = env.ColonyName
	addedRD, err := client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add some resources of this kind
	resource1 := core.CreateResource("ProtectedResource", "res-1", env.ColonyName)
	_, err = client.AddResource(resource1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	resource2 := core.CreateResource("ProtectedResource", "res-2", env.ColonyName)
	_, err = client.AddResource(resource2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to remove ResourceDefinition while resources exist - should fail
	err = client.RemoveResourceDefinition(env.ColonyName, addedRD.Metadata.Name, env.ColonyPrvKey)
	assert.NotNil(t, err, "Removing ResourceDefinition with active resources should fail")
	assert.Contains(t, err.Error(), "2 resource(s) of kind 'ProtectedResource' still exist")

	// Remove one resource
	err = client.RemoveResource(env.ColonyName, resource1.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try again - should still fail because one resource remains
	err = client.RemoveResourceDefinition(env.ColonyName, addedRD.Metadata.Name, env.ColonyPrvKey)
	assert.NotNil(t, err, "Removing ResourceDefinition with 1 active resource should still fail")
	assert.Contains(t, err.Error(), "1 resource(s) of kind 'ProtectedResource' still exist")

	// Remove the last resource
	err = client.RemoveResource(env.ColonyName, resource2.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Now removal should succeed
	err = client.RemoveResourceDefinition(env.ColonyName, addedRD.Metadata.Name, env.ColonyPrvKey)
	assert.Nil(t, err, "Removing ResourceDefinition with no active resources should succeed")

	server.Shutdown()
	<-done
}

func TestGetResourceHistory(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// First create a ResourceDefinition
	rd := core.CreateResourceDefinition(
		"testresource",
		"example.com",
		"v1",
		"TestResource",
		"testresources",
		"Namespaced",
		"test_controller",
		"reconcile_testresource",
	)
	rd.Metadata.Namespace = env.ColonyName

	_, err := client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create a Resource
	resource := core.CreateResource("TestResource", "test-resource-1", env.ColonyName)
	resource.SetSpec("replicas", 3)
	resource.SetStatus("phase", "Running")

	// Add Resource
	addedResource, err := client.AddResource(resource, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedResource)

	// Update the resource to create more history
	addedResource.SetSpec("replicas", 5)
	updatedResource, err := client.UpdateResource(addedResource, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, updatedResource)

	// Get resource history
	histories, err := client.GetResourceHistory(addedResource.ID, 10, env.ExecutorPrvKey)
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
