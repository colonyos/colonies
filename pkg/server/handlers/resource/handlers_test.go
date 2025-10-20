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
		"test_executor_type",
		"reconcile_database",
	)
	rd.Metadata.Namespace = env.ColonyName

	_, err := client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create a Resource instance
	resource := core.CreateResource("example.com/v1", "Database", "my-database", env.ColonyName)
	resource.SetSpec("host", "localhost")
	resource.SetSpec("port", 5432)
	resource.SetSpec("name", "testdb")

	// Add Resource with executor key
	addedResource, err := client.AddResource(resource, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedResource)
	assert.Equal(t, resource.Metadata.Name, addedResource.Metadata.Name)
	assert.Equal(t, resource.Kind, addedResource.Kind)

	// Try to add duplicate Resource - should fail (same namespace + name)
	_, err = client.AddResource(resource, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestGetResource(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Setup ResourceDefinition
	rd := core.CreateResourceDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	rd.Metadata.Namespace = env.ColonyName
	_, err := client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Resource
	resource := core.CreateResource("example.com/v1", "Database", "my-database", env.ColonyName)
	resource.SetSpec("host", "localhost")

	addedResource, err := client.AddResource(resource, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Get Resource
	retrievedResource, err := client.GetResource(env.ColonyName, resource.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedResource)
	assert.Equal(t, addedResource.ID, retrievedResource.ID)
	assert.Equal(t, addedResource.Metadata.Name, retrievedResource.Metadata.Name)

	// Verify spec data
	host, ok := retrievedResource.GetSpec("host")
	assert.True(t, ok)
	assert.Equal(t, "localhost", host)

	server.Shutdown()
	<-done
}

func TestGetResources(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Setup ResourceDefinitions
	rd1 := core.CreateResourceDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	rd1.Metadata.Namespace = env.ColonyName
	_, err := client.AddResourceDefinition(rd1, env.ColonyPrvKey)
	assert.Nil(t, err)

	rd2 := core.CreateResourceDefinition(
		"queue",
		"example.com",
		"v1",
		"Queue",
		"queues",
		"Namespaced",
		"test_executor_type",
		"reconcile_queue",
	)
	rd2.Metadata.Namespace = env.ColonyName
	_, err = client.AddResourceDefinition(rd2, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add multiple Resources of different kinds
	db1 := core.CreateResource("example.com/v1", "Database", "db1", env.ColonyName)
	db1.SetSpec("host", "localhost")
	_, err = client.AddResource(db1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	db2 := core.CreateResource("example.com/v1", "Database", "db2", env.ColonyName)
	db2.SetSpec("host", "remotehost")
	_, err = client.AddResource(db2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	queue1 := core.CreateResource("example.com/v1", "Queue", "queue1", env.ColonyName)
	queue1.SetSpec("maxSize", 1000)
	_, err = client.AddResource(queue1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Get all resources in namespace
	allResources, err := client.GetResources(env.ColonyName, "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, allResources, 3)

	// Get only Database resources
	databases, err := client.GetResources(env.ColonyName, "Database", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, databases, 2)

	// Get only Queue resources
	queues, err := client.GetResources(env.ColonyName, "Queue", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, queues, 1)

	server.Shutdown()
	<-done
}

func TestUpdateResource(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Setup ResourceDefinition
	rd := core.CreateResourceDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	rd.Metadata.Namespace = env.ColonyName
	_, err := client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Resource
	resource := core.CreateResource("example.com/v1", "Database", "my-database", env.ColonyName)
	resource.SetSpec("host", "localhost")
	resource.SetSpec("port", 5432)

	addedResource, err := client.AddResource(resource, env.ExecutorPrvKey)
	assert.Nil(t, err)
	originalGeneration := addedResource.Metadata.Generation

	// Update Resource spec
	addedResource.SetSpec("port", 5433)
	addedResource.SetStatus("state", "ready")

	updatedResource, err := client.UpdateResource(addedResource, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, updatedResource)

	// Verify updates
	port, ok := updatedResource.GetSpec("port")
	assert.True(t, ok)
	assert.Equal(t, float64(5433), port) // JSON unmarshaling converts to float64

	state, ok := updatedResource.GetStatus("state")
	assert.True(t, ok)
	assert.Equal(t, "ready", state)

	// Generation should be incremented
	assert.Greater(t, updatedResource.Metadata.Generation, originalGeneration)

	server.Shutdown()
	<-done
}

func TestRemoveResource(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Setup ResourceDefinition
	rd := core.CreateResourceDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	rd.Metadata.Namespace = env.ColonyName
	_, err := client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Resources
	resource1 := core.CreateResource("example.com/v1", "Database", "db1", env.ColonyName)
	_, err = client.AddResource(resource1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	resource2 := core.CreateResource("example.com/v1", "Database", "db2", env.ColonyName)
	_, err = client.AddResource(resource2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify both exist
	resources, err := client.GetResources(env.ColonyName, "Database", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, resources, 2)

	// Remove one resource
	err = client.RemoveResource(env.ColonyName, "db1", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify only one remains
	resources, err = client.GetResources(env.ColonyName, "Database", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, resources, 1)
	assert.Equal(t, "db2", resources[0].Metadata.Name)

	// Try to get removed resource - should fail
	_, err = client.GetResource(env.ColonyName, "db1", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestResourceWithComplexSpec(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Setup ResourceDefinition
	rd := core.CreateResourceDefinition(
		"application",
		"example.com",
		"v1",
		"Application",
		"applications",
		"Namespaced",
		"test_executor_type",
		"reconcile_application",
	)
	rd.Metadata.Namespace = env.ColonyName
	_, err := client.AddResourceDefinition(rd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create Resource with complex nested spec
	resource := core.CreateResource("example.com/v1", "Application", "my-app", env.ColonyName)
	resource.SetSpec("replicas", 3)
	resource.SetSpec("image", "myapp:v1.0.0")
	resource.SetSpec("ports", []interface{}{8080, 8443})
	resource.SetSpec("env", map[string]interface{}{
		"DATABASE_URL": "postgres://localhost/db",
		"CACHE_URL":    "redis://localhost",
	})
	resource.Metadata.Labels["version"] = "v1.0.0"
	resource.Metadata.Annotations["description"] = "My test application"

	// Add Resource
	_, err = client.AddResource(resource, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Retrieve and verify all fields
	retrievedResource, err := client.GetResource(env.ColonyName, resource.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

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
