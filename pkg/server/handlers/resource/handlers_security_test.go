package resource_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddResourceDefinitionSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Create a ResourceDefinition for colony1
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
	rd.Metadata.Namespace = env.Colony1Name

	// Only colony owner should be able to add ResourceDefinitions

	// Try with executor key - should FAIL
	_, err := client.AddResourceDefinition(rd, env.Executor1PrvKey)
	assert.NotNil(t, err)

	// Try with another colony's owner key - should FAIL
	_, err = client.AddResourceDefinition(rd, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with correct colony owner key - should SUCCEED
	_, err = client.AddResourceDefinition(rd, env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetResourceDefinitionSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Create and add ResourceDefinition for colony1
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
	rd.Metadata.Namespace = env.Colony1Name

	_, err := client.AddResourceDefinition(rd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Colony members should be able to get ResourceDefinitions

	// Try with executor from same colony - should SUCCEED
	_, err = client.GetResourceDefinition(env.Colony1Name, rd.Metadata.Name, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Try with executor from different colony - should FAIL
	_, err = client.GetResourceDefinition(env.Colony1Name, rd.Metadata.Name, env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	_, err = client.GetResourceDefinition(env.Colony1Name, rd.Metadata.Name, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with correct colony owner - should SUCCEED
	_, err = client.GetResourceDefinition(env.Colony1Name, rd.Metadata.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestAddResourceSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Setup ResourceDefinition for colony1
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
	rd.Metadata.Namespace = env.Colony1Name

	_, err := client.AddResourceDefinition(rd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Create a Resource instance for colony1
	resource := core.CreateResource("Database", "my-database", env.Colony1Name)
	resource.SetSpec("host", "localhost")

	// Colony members should be able to add Resources

	// Try with executor from different colony - should FAIL
	_, err = client.AddResource(resource, env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	_, err = client.AddResource(resource, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with executor from same colony - should SUCCEED
	_, err = client.AddResource(resource, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Create another resource to test with colony owner
	resource2 := core.CreateResource("Database", "another-database", env.Colony1Name)
	resource2.SetSpec("host", "remotehost")

	// Try with colony owner - should also SUCCEED
	_, err = client.AddResource(resource2, env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetResourceSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

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
	rd.Metadata.Namespace = env.Colony1Name
	_, err := client.AddResourceDefinition(rd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Add Resource to colony1
	resource := core.CreateResource("Database", "my-database", env.Colony1Name)
	resource.SetSpec("host", "localhost")
	_, err = client.AddResource(resource, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Try to get resource from different colony - should FAIL
	_, err = client.GetResource(env.Colony1Name, resource.Metadata.Name, env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	_, err = client.GetResource(env.Colony1Name, resource.Metadata.Name, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with executor from same colony - should SUCCEED
	_, err = client.GetResource(env.Colony1Name, resource.Metadata.Name, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Try with colony owner - should SUCCEED
	_, err = client.GetResource(env.Colony1Name, resource.Metadata.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetResourcesSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

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
	rd.Metadata.Namespace = env.Colony1Name
	_, err := client.AddResourceDefinition(rd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Add Resources to colony1
	resource1 := core.CreateResource("Database", "db1", env.Colony1Name)
	_, err = client.AddResource(resource1, env.Executor1PrvKey)
	assert.Nil(t, err)

	resource2 := core.CreateResource("Database", "db2", env.Colony1Name)
	_, err = client.AddResource(resource2, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Try to list resources from different colony - should FAIL
	_, err = client.GetResources(env.Colony1Name, "Database", env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	_, err = client.GetResources(env.Colony1Name, "Database", env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with executor from same colony - should SUCCEED
	resources, err := client.GetResources(env.Colony1Name, "Database", env.Executor1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, resources, 2)

	// Try with colony owner - should SUCCEED
	resources, err = client.GetResources(env.Colony1Name, "Database", env.Colony1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, resources, 2)

	server.Shutdown()
	<-done
}

func TestUpdateResourceSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

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
	rd.Metadata.Namespace = env.Colony1Name
	_, err := client.AddResourceDefinition(rd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Add Resource to colony1
	resource := core.CreateResource("Database", "my-database", env.Colony1Name)
	resource.SetSpec("host", "localhost")
	addedResource, err := client.AddResource(resource, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Update the resource spec
	addedResource.SetSpec("port", 5432)

	// Try to update from different colony executor - should FAIL
	_, err = client.UpdateResource(addedResource, env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	_, err = client.UpdateResource(addedResource, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with executor from same colony - should SUCCEED
	_, err = client.UpdateResource(addedResource, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Update again with colony owner - should SUCCEED
	addedResource.SetSpec("port", 5433)
	_, err = client.UpdateResource(addedResource, env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestRemoveResourceSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

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
	rd.Metadata.Namespace = env.Colony1Name
	_, err := client.AddResourceDefinition(rd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Add Resources to colony1
	resource1 := core.CreateResource("Database", "db1", env.Colony1Name)
	_, err = client.AddResource(resource1, env.Executor1PrvKey)
	assert.Nil(t, err)

	resource2 := core.CreateResource("Database", "db2", env.Colony1Name)
	_, err = client.AddResource(resource2, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Try to remove from different colony executor - should FAIL
	err = client.RemoveResource(env.Colony1Name, "db1", env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	err = client.RemoveResource(env.Colony1Name, "db1", env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with executor from same colony - should SUCCEED
	err = client.RemoveResource(env.Colony1Name, "db1", env.Executor1PrvKey)
	assert.Nil(t, err)

	// Verify it was removed
	_, err = client.GetResource(env.Colony1Name, "db1", env.Executor1PrvKey)
	assert.NotNil(t, err)

	// Remove with colony owner - should SUCCEED
	err = client.RemoveResource(env.Colony1Name, "db2", env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestCrossColonyResourceIsolation(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Create ResourceDefinitions for both colonies
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
	rd1.Metadata.Namespace = env.Colony1Name
	_, err := client.AddResourceDefinition(rd1, env.Colony1PrvKey)
	assert.Nil(t, err)

	rd2 := core.CreateResourceDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	rd2.Metadata.Namespace = env.Colony2Name
	_, err = client.AddResourceDefinition(rd2, env.Colony2PrvKey)
	assert.Nil(t, err)

	// Add Resources to both colonies with same name
	resource1 := core.CreateResource("Database", "shared-name", env.Colony1Name)
	resource1.SetSpec("colonyId", "colony1")
	_, err = client.AddResource(resource1, env.Executor1PrvKey)
	assert.Nil(t, err)

	resource2 := core.CreateResource("Database", "shared-name", env.Colony2Name)
	resource2.SetSpec("colonyId", "colony2")
	_, err = client.AddResource(resource2, env.Executor2PrvKey)
	assert.Nil(t, err)

	// Each colony should only see its own resource
	r1, err := client.GetResource(env.Colony1Name, "shared-name", env.Executor1PrvKey)
	assert.Nil(t, err)
	colonyId1, _ := r1.GetSpec("colonyId")
	assert.Equal(t, "colony1", colonyId1)

	r2, err := client.GetResource(env.Colony2Name, "shared-name", env.Executor2PrvKey)
	assert.Nil(t, err)
	colonyId2, _ := r2.GetSpec("colonyId")
	assert.Equal(t, "colony2", colonyId2)

	// Verify isolation - executor1 cannot see colony2 resources
	_, err = client.GetResource(env.Colony2Name, "shared-name", env.Executor1PrvKey)
	assert.NotNil(t, err)

	// Verify isolation - executor2 cannot see colony1 resources
	_, err = client.GetResource(env.Colony1Name, "shared-name", env.Executor2PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestResourceDefinitionOnlyColonyOwner(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Create additional executor in colony1
	executor3, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor3, env.Colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.Colony1Name, executor3.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Create ResourceDefinition
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
	rd.Metadata.Namespace = env.Colony1Name

	// None of the executors should be able to add ResourceDefinitions
	_, err = client.AddResourceDefinition(rd, env.Executor1PrvKey)
	assert.NotNil(t, err)

	_, err = client.AddResourceDefinition(rd, executor3PrvKey)
	assert.NotNil(t, err)

	// Only colony owner can add ResourceDefinitions
	_, err = client.AddResourceDefinition(rd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// But all colony members can read ResourceDefinitions
	_, err = client.GetResourceDefinition(env.Colony1Name, rd.Metadata.Name, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetResourceDefinition(env.Colony1Name, rd.Metadata.Name, executor3PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}
