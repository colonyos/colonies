package service_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/stretchr/testify/assert"
)

func TestAddServiceDefinition(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create a ServiceDefinition
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
	sd.Metadata.Namespace = env.ColonyName

	// Add ServiceDefinition with colony owner key
	addedSD, err := client.AddServiceDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedSD)
	assert.Equal(t, sd.Metadata.Name, addedSD.Metadata.Name)
	assert.Equal(t, sd.Spec.Group, addedSD.Spec.Group)
	assert.Equal(t, sd.Spec.Version, addedSD.Spec.Version)

	// Try to add duplicate ServiceDefinition - should fail
	_, err = client.AddServiceDefinition(sd, env.ColonyPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestGetServiceDefinition(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create and add ServiceDefinition
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
	sd.Metadata.Namespace = env.ColonyName

	addedSD, err := client.AddServiceDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Get ServiceDefinition (using executor key since only members can get)
	retrievedSD, err := client.GetServiceDefinition(env.ColonyName, sd.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedSD)
	assert.Equal(t, addedSD.ID, retrievedSD.ID)
	assert.Equal(t, addedSD.Metadata.Name, retrievedSD.Metadata.Name)

	server.Shutdown()
	<-done
}

func TestAddService(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// First add a ServiceDefinition
	sd := core.CreateServiceDefinition(
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

	_, err := client.AddServiceDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create a Service instance
	service := core.CreateService("Database", "test-database", env.ColonyName)
	service.SetSpec("host", "localhost")
	service.SetSpec("port", 5432)

	// Add Service
	addedService, err := client.AddService(service, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedService)
	assert.Equal(t, service.Metadata.Name, addedService.Metadata.Name)
	assert.Equal(t, service.Kind, addedService.Kind)

	// Verify spec was preserved
	host, ok := addedService.GetSpec("host")
	assert.True(t, ok)
	assert.Equal(t, "localhost", host)

	server.Shutdown()
	<-done
}

func TestGetService(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add ServiceDefinition
	sd := core.CreateServiceDefinition(
		"service",
		"example.com",
		"v1",
		"Service",
		"services",
		"Namespaced",
		"service_controller",
		"reconcile_service",
	)
	sd.Metadata.Namespace = env.ColonyName
	_, err := client.AddServiceDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Service
	service := core.CreateService("Service", "web-service", env.ColonyName)
	service.SetSpec("port", 8080)
	addedService, err := client.AddService(service, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Get Service
	retrievedService, err := client.GetService(env.ColonyName, service.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedService)
	assert.Equal(t, addedService.ID, retrievedService.ID)
	assert.Equal(t, addedService.Metadata.Name, retrievedService.Metadata.Name)

	server.Shutdown()
	<-done
}

func TestGetServices(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add ServiceDefinition for Database
	sdDB := core.CreateServiceDefinition(
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
	_, err := client.AddServiceDefinition(sdDB, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add ServiceDefinition for Service
	sdSvc := core.CreateServiceDefinition(
		"service",
		"example.com",
		"v1",
		"Service",
		"services",
		"Namespaced",
		"service_controller",
		"reconcile_service",
	)
	sdSvc.Metadata.Namespace = env.ColonyName
	_, err = client.AddServiceDefinition(sdSvc, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add some Database services
	db1 := core.CreateService("Database", "db1", env.ColonyName)
	db2 := core.CreateService("Database", "db2", env.ColonyName)
	_, err = client.AddService(db1, env.ExecutorPrvKey)
	assert.Nil(t, err)
	_, err = client.AddService(db2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Add some Service services
	svc1 := core.CreateService("Service", "svc1", env.ColonyName)
	_, err = client.AddService(svc1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Get all services
	allServices, err := client.GetServices(env.ColonyName, "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(allServices))

	// Get only Database services
	dbServices, err := client.GetServices(env.ColonyName, "Database", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(dbServices))

	// Get only Service services
	svcServices, err := client.GetServices(env.ColonyName, "Service", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(svcServices))

	server.Shutdown()
	<-done
}

func TestUpdateService(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add ServiceDefinition
	sd := core.CreateServiceDefinition(
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
	_, err := client.AddServiceDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Service
	service := core.CreateService("Application", "my-app", env.ColonyName)
	service.SetSpec("version", "1.0.0")
	addedService, err := client.AddService(service, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Update Service
	addedService.SetSpec("version", "1.1.0")
	updatedService, err := client.UpdateService(addedService, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, updatedService)

	version, ok := updatedService.GetSpec("version")
	assert.True(t, ok)
	assert.Equal(t, "1.1.0", version)

	server.Shutdown()
	<-done
}

func TestRemoveService(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add ServiceDefinition
	sd := core.CreateServiceDefinition(
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
	_, err := client.AddServiceDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add Service
	service := core.CreateService("Cache", "redis-cache", env.ColonyName)
	addedService, err := client.AddService(service, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Remove Service
	err = client.RemoveService(env.ColonyName, addedService.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify it's removed
	_, err = client.GetService(env.ColonyName, addedService.Metadata.Name, env.ExecutorPrvKey)
	assert.NotNil(t, err) // Should fail because service doesn't exist

	server.Shutdown()
	<-done
}

func TestServiceWithComplexSpec(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add ServiceDefinition
	sd := core.CreateServiceDefinition(
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
	_, err := client.AddServiceDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create Service with complex spec
	service := core.CreateService("Deployment", "web-deployment", env.ColonyName)
	service.SetSpec("image", "nginx:1.21")
	service.SetSpec("replicas", 3)
	service.SetSpec("env", map[string]interface{}{
		"DATABASE_URL": "postgres://localhost/db",
		"PORT":         "8080",
	})
	service.Metadata.Labels = map[string]string{
		"app":     "web",
		"version": "v1.0.0",
	}
	service.Metadata.Annotations = map[string]string{
		"description": "My test application",
	}

	// Add Service
	_, err = client.AddService(service, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Retrieve and verify
	retrievedService, err := client.GetService(env.ColonyName, service.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	image, ok := retrievedService.GetSpec("image")
	assert.True(t, ok)
	assert.Equal(t, "nginx:1.21", image)

	replicas, ok := retrievedService.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, float64(3), replicas) // JSON unmarshaling converts to float64

	envSpec, ok := retrievedService.GetSpec("env")
	assert.True(t, ok)
	envMap := envSpec.(map[string]interface{})
	assert.Equal(t, "postgres://localhost/db", envMap["DATABASE_URL"])

	assert.Equal(t, "v1.0.0", retrievedService.Metadata.Labels["version"])
	assert.Equal(t, "My test application", retrievedService.Metadata.Annotations["description"])

	server.Shutdown()
	<-done
}

// TestAddServiceRequiresServiceDefinition tests that adding a service without a ServiceDefinition fails
func TestAddServiceRequiresServiceDefinition(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Try to add a Service WITHOUT adding its ServiceDefinition first
	service := core.CreateService("NonExistentKind", "test-service", env.ColonyName)
	service.SetSpec("field", "value")

	// This should fail because ServiceDefinition doesn't exist
	_, err := client.AddService(service, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Adding service without ServiceDefinition should fail")
	assert.Contains(t, err.Error(), "ServiceDefinition for kind 'NonExistentKind' not found")

	server.Shutdown()
	<-done
}

// TestAddServiceWithSchemaValidation tests that services are validated against the ServiceDefinition schema
func TestAddServiceWithSchemaValidation(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create ServiceDefinition with schema validation
	sd := core.CreateServiceDefinition(
		"validated-service",
		"example.com",
		"v1",
		"ValidatedService",
		"validatedservices",
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
				Description: "Service name",
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

	_, err := client.AddServiceDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Test 1: Valid service should succeed
	validService := core.CreateService("ValidatedService", "valid-res", env.ColonyName)
	validService.SetSpec("name", "test")
	validService.SetSpec("replicas", 3)
	validService.SetSpec("protocol", "TCP")

	addedService, err := client.AddService(validService, env.ExecutorPrvKey)
	assert.Nil(t, err, "Valid service should be added successfully")
	assert.NotNil(t, addedService)

	// Test 2: Service missing required field should fail
	invalidService1 := core.CreateService("ValidatedService", "invalid-res-1", env.ColonyName)
	invalidService1.SetSpec("name", "test") // Missing required 'replicas'

	_, err = client.AddService(invalidService1, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Service missing required field should fail")
	assert.Contains(t, err.Error(), "required field 'replicas' is missing")

	// Test 3: Service with invalid type should fail
	invalidService2 := core.CreateService("ValidatedService", "invalid-res-2", env.ColonyName)
	invalidService2.SetSpec("name", "test")
	invalidService2.SetSpec("replicas", "not-a-number") // Should be number

	_, err = client.AddService(invalidService2, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Service with invalid type should fail")
	assert.Contains(t, err.Error(), "must be a number")

	// Test 4: Service with invalid enum value should fail
	invalidService3 := core.CreateService("ValidatedService", "invalid-res-3", env.ColonyName)
	invalidService3.SetSpec("name", "test")
	invalidService3.SetSpec("replicas", 3)
	invalidService3.SetSpec("protocol", "HTTP") // Not in enum [TCP, UDP]

	_, err = client.AddService(invalidService3, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Service with invalid enum value should fail")
	assert.Contains(t, err.Error(), "must be one of")

	server.Shutdown()
	<-done
}

// TestRemoveServiceDefinitionWithActiveServices tests that removing a ServiceDefinition with active services fails
func TestRemoveServiceDefinitionWithActiveServices(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Add ServiceDefinition
	sd := core.CreateServiceDefinition(
		"protected-service",
		"example.com",
		"v1",
		"ProtectedService",
		"protectedservices",
		"Namespaced",
		"protected_controller",
		"reconcile_protected",
	)
	sd.Metadata.Namespace = env.ColonyName
	addedSD, err := client.AddServiceDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add some services of this kind
	service1 := core.CreateService("ProtectedService", "res-1", env.ColonyName)
	_, err = client.AddService(service1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	service2 := core.CreateService("ProtectedService", "res-2", env.ColonyName)
	_, err = client.AddService(service2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to remove ServiceDefinition while services exist - should fail
	err = client.RemoveServiceDefinition(env.ColonyName, addedSD.Metadata.Name, env.ColonyPrvKey)
	assert.NotNil(t, err, "Removing ServiceDefinition with active services should fail")
	assert.Contains(t, err.Error(), "2 service(s) of kind 'ProtectedService' still exist")

	// Remove one service
	err = client.RemoveService(env.ColonyName, service1.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try again - should still fail because one service remains
	err = client.RemoveServiceDefinition(env.ColonyName, addedSD.Metadata.Name, env.ColonyPrvKey)
	assert.NotNil(t, err, "Removing ServiceDefinition with 1 active service should still fail")
	assert.Contains(t, err.Error(), "1 service(s) of kind 'ProtectedService' still exist")

	// Remove the last service
	err = client.RemoveService(env.ColonyName, service2.Metadata.Name, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Now removal should succeed
	err = client.RemoveServiceDefinition(env.ColonyName, addedSD.Metadata.Name, env.ColonyPrvKey)
	assert.Nil(t, err, "Removing ServiceDefinition with no active services should succeed")

	server.Shutdown()
	<-done
}

func TestGetServiceHistory(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// First create a ServiceDefinition
	sd := core.CreateServiceDefinition(
		"testresource",
		"example.com",
		"v1",
		"TestService",
		"testservices",
		"Namespaced",
		"test_controller",
		"reconcile_testresource",
	)
	sd.Metadata.Namespace = env.ColonyName

	_, err := client.AddServiceDefinition(sd, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Create a Service
	service := core.CreateService("TestService", "test-service-1", env.ColonyName)
	service.SetSpec("replicas", 3)
	service.SetStatus("phase", "Running")

	// Add Service
	addedService, err := client.AddService(service, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedService)

	// Update the service to create more history
	addedService.SetSpec("replicas", 5)
	updatedService, err := client.UpdateService(addedService, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, updatedService)

	// Get service history
	histories, err := client.GetServiceHistory(addedService.ID, 10, env.ExecutorPrvKey)
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
