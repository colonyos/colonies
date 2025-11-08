package service_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddServiceDefinitionSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Create a ServiceDefinition for colony1
	sd := core.CreateServiceDefinition(
		"test-service",
		"example.com",
		"v1",
		"TestService",
		"testservices",
		"Namespaced",
		"test_executor_type",
		"reconcile_test_resource",
	)
	sd.Metadata.Namespace = env.Colony1Name

	// Only colony owner should be able to add ServiceDefinitions

	// Try with executor key - should FAIL
	_, err := client.AddServiceDefinition(sd, env.Executor1PrvKey)
	assert.NotNil(t, err)

	// Try with another colony's owner key - should FAIL
	_, err = client.AddServiceDefinition(sd, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with correct colony owner key - should SUCCEED
	_, err = client.AddServiceDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetServiceDefinitionSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Create and add ServiceDefinition for colony1
	sd := core.CreateServiceDefinition(
		"test-service",
		"example.com",
		"v1",
		"TestService",
		"testservices",
		"Namespaced",
		"test_executor_type",
		"reconcile_test_resource",
	)
	sd.Metadata.Namespace = env.Colony1Name

	_, err := client.AddServiceDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Colony members should be able to get ServiceDefinitions

	// Try with executor from same colony - should SUCCEED
	_, err = client.GetServiceDefinition(env.Colony1Name, sd.Metadata.Name, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Try with executor from different colony - should FAIL
	_, err = client.GetServiceDefinition(env.Colony1Name, sd.Metadata.Name, env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	_, err = client.GetServiceDefinition(env.Colony1Name, sd.Metadata.Name, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with correct colony owner - should SUCCEED
	_, err = client.GetServiceDefinition(env.Colony1Name, sd.Metadata.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestAddServiceSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Setup ServiceDefinition for colony1
	sd := core.CreateServiceDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	sd.Metadata.Namespace = env.Colony1Name

	_, err := client.AddServiceDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Create a Service instance for colony1
	service := core.CreateService("Database", "my-database", env.Colony1Name)
	service.SetSpec("host", "localhost")

	// Colony members should be able to add Services

	// Try with executor from different colony - should FAIL
	_, err = client.AddService(service, env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	_, err = client.AddService(service, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with executor from same colony - should SUCCEED
	_, err = client.AddService(service, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Create another service to test with colony owner
	service2 := core.CreateService("Database", "another-database", env.Colony1Name)
	service2.SetSpec("host", "remotehost")

	// Try with colony owner - should also SUCCEED
	_, err = client.AddService(service2, env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetServiceSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Setup ServiceDefinition
	sd := core.CreateServiceDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	sd.Metadata.Namespace = env.Colony1Name
	_, err := client.AddServiceDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Add Service to colony1
	service := core.CreateService("Database", "my-database", env.Colony1Name)
	service.SetSpec("host", "localhost")
	_, err = client.AddService(service, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Try to get service from different colony - should FAIL
	_, err = client.GetService(env.Colony1Name, service.Metadata.Name, env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	_, err = client.GetService(env.Colony1Name, service.Metadata.Name, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with executor from same colony - should SUCCEED
	_, err = client.GetService(env.Colony1Name, service.Metadata.Name, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Try with colony owner - should SUCCEED
	_, err = client.GetService(env.Colony1Name, service.Metadata.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetServicesSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Setup ServiceDefinition
	sd := core.CreateServiceDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	sd.Metadata.Namespace = env.Colony1Name
	_, err := client.AddServiceDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Add Services to colony1
	service1 := core.CreateService("Database", "db1", env.Colony1Name)
	_, err = client.AddService(service1, env.Executor1PrvKey)
	assert.Nil(t, err)

	service2 := core.CreateService("Database", "db2", env.Colony1Name)
	_, err = client.AddService(service2, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Try to list services from different colony - should FAIL
	_, err = client.GetServices(env.Colony1Name, "Database", env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	_, err = client.GetServices(env.Colony1Name, "Database", env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with executor from same colony - should SUCCEED
	services, err := client.GetServices(env.Colony1Name, "Database", env.Executor1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, services, 2)

	// Try with colony owner - should SUCCEED
	services, err = client.GetServices(env.Colony1Name, "Database", env.Colony1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, services, 2)

	server.Shutdown()
	<-done
}

func TestUpdateServiceSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Setup ServiceDefinition
	sd := core.CreateServiceDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	sd.Metadata.Namespace = env.Colony1Name
	_, err := client.AddServiceDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Add Service to colony1
	service := core.CreateService("Database", "my-database", env.Colony1Name)
	service.SetSpec("host", "localhost")
	addedService, err := client.AddService(service, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Update the service spec
	addedService.SetSpec("port", 5432)

	// Try to update from different colony executor - should FAIL
	_, err = client.UpdateService(addedService, env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	_, err = client.UpdateService(addedService, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with executor from same colony - should SUCCEED
	_, err = client.UpdateService(addedService, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Update again with colony owner - should SUCCEED
	addedService.SetSpec("port", 5433)
	_, err = client.UpdateService(addedService, env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestRemoveServiceSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Setup ServiceDefinition
	sd := core.CreateServiceDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	sd.Metadata.Namespace = env.Colony1Name
	_, err := client.AddServiceDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Add Services to colony1
	service1 := core.CreateService("Database", "db1", env.Colony1Name)
	_, err = client.AddService(service1, env.Executor1PrvKey)
	assert.Nil(t, err)

	service2 := core.CreateService("Database", "db2", env.Colony1Name)
	_, err = client.AddService(service2, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Try to remove from different colony executor - should FAIL
	err = client.RemoveService(env.Colony1Name, "db1", env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	err = client.RemoveService(env.Colony1Name, "db1", env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with executor from same colony - should SUCCEED
	err = client.RemoveService(env.Colony1Name, "db1", env.Executor1PrvKey)
	assert.Nil(t, err)

	// Verify it was removed
	_, err = client.GetService(env.Colony1Name, "db1", env.Executor1PrvKey)
	assert.NotNil(t, err)

	// Remove with colony owner - should SUCCEED
	err = client.RemoveService(env.Colony1Name, "db2", env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestCrossColonyServiceIsolation(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Create ServiceDefinitions for both colonies
	sd1 := core.CreateServiceDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	sd1.Metadata.Namespace = env.Colony1Name
	_, err := client.AddServiceDefinition(sd1, env.Colony1PrvKey)
	assert.Nil(t, err)

	sd2 := core.CreateServiceDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	sd2.Metadata.Namespace = env.Colony2Name
	_, err = client.AddServiceDefinition(sd2, env.Colony2PrvKey)
	assert.Nil(t, err)

	// Add Services to both colonies with same name
	service1 := core.CreateService("Database", "shared-name", env.Colony1Name)
	service1.SetSpec("colonyId", "colony1")
	_, err = client.AddService(service1, env.Executor1PrvKey)
	assert.Nil(t, err)

	service2 := core.CreateService("Database", "shared-name", env.Colony2Name)
	service2.SetSpec("colonyId", "colony2")
	_, err = client.AddService(service2, env.Executor2PrvKey)
	assert.Nil(t, err)

	// Each colony should only see its own service
	s1, err := client.GetService(env.Colony1Name, "shared-name", env.Executor1PrvKey)
	assert.Nil(t, err)
	colonyId1, _ := s1.GetSpec("colonyId")
	assert.Equal(t, "colony1", colonyId1)

	s2, err := client.GetService(env.Colony2Name, "shared-name", env.Executor2PrvKey)
	assert.Nil(t, err)
	colonyId2, _ := s2.GetSpec("colonyId")
	assert.Equal(t, "colony2", colonyId2)

	// Verify isolation - executor1 cannot see colony2 services
	_, err = client.GetService(env.Colony2Name, "shared-name", env.Executor1PrvKey)
	assert.NotNil(t, err)

	// Verify isolation - executor2 cannot see colony1 services
	_, err = client.GetService(env.Colony1Name, "shared-name", env.Executor2PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestServiceDefinitionOnlyColonyOwner(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Create additional executor in colony1
	executor3, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor3, env.Colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.Colony1Name, executor3.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Create ServiceDefinition
	sd := core.CreateServiceDefinition(
		"test-service",
		"example.com",
		"v1",
		"TestService",
		"testservices",
		"Namespaced",
		"test_executor_type",
		"reconcile_test_resource",
	)
	sd.Metadata.Namespace = env.Colony1Name

	// None of the executors should be able to add ServiceDefinitions
	_, err = client.AddServiceDefinition(sd, env.Executor1PrvKey)
	assert.NotNil(t, err)

	_, err = client.AddServiceDefinition(sd, executor3PrvKey)
	assert.NotNil(t, err)

	// Only colony owner can add ServiceDefinitions
	_, err = client.AddServiceDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// But all colony members can read ServiceDefinitions
	_, err = client.GetServiceDefinition(env.Colony1Name, sd.Metadata.Name, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetServiceDefinition(env.Colony1Name, sd.Metadata.Name, executor3PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}
