package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestAddGetResourceDefinition(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	rd := core.CreateResourceDefinition(
		"executordeployments.compute.colonies.io",
		"compute.colonies.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"executor-controller",
		"reconcile",
	)
	rd.Metadata.Namespace = "test-colony"

	err = db.AddResourceDefinition(rd)
	assert.Nil(t, err)

	// Get by ID
	rdFromDB, err := db.GetResourceDefinitionByID(rd.ID)
	assert.Nil(t, err)
	assert.NotNil(t, rdFromDB)
	assert.Equal(t, rd.ID, rdFromDB.ID)
	assert.Equal(t, rd.Metadata.Name, rdFromDB.Metadata.Name)
	assert.Equal(t, rd.Spec.Group, rdFromDB.Spec.Group)
	assert.Equal(t, rd.Spec.Version, rdFromDB.Spec.Version)

	// Get by name
	rdFromDB2, err := db.GetResourceDefinitionByName(rd.Metadata.Namespace, rd.Metadata.Name)
	assert.Nil(t, err)
	assert.NotNil(t, rdFromDB2)
	assert.Equal(t, rd.ID, rdFromDB2.ID)

	// Get all
	rds, err := db.GetResourceDefinitions()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rds))

	// Count
	count, err := db.CountResourceDefinitions()
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
}

func TestAddGetResource(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	resource := core.CreateResource("ExecutorDeployment", "web-server", "production")
	resource.SetSpec("image", "nginx:1.21")
	resource.SetSpec("replicas", 3)
	resource.SetStatus("phase", "Running")

	err = db.AddResource(resource)
	assert.Nil(t, err)

	// Get by ID
	resourceFromDB, err := db.GetResourceByID(resource.ID)
	assert.Nil(t, err)
	assert.NotNil(t, resourceFromDB)
	assert.Equal(t, resource.ID, resourceFromDB.ID)
	assert.Equal(t, resource.Metadata.Name, resourceFromDB.Metadata.Name)
	assert.Equal(t, resource.Metadata.Namespace, resourceFromDB.Metadata.Namespace)
	assert.Equal(t, resource.Kind, resourceFromDB.Kind)

	// Verify spec
	image, ok := resourceFromDB.GetSpec("image")
	assert.True(t, ok)
	assert.Equal(t, "nginx:1.21", image)

	replicas, ok := resourceFromDB.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, float64(3), replicas) // JSON unmarshaling converts to float64

	// Verify status
	phase, ok := resourceFromDB.GetStatus("phase")
	assert.True(t, ok)
	assert.Equal(t, "Running", phase)

	// Get by name
	resourceFromDB2, err := db.GetResourceByName(resource.Metadata.Namespace, resource.Metadata.Name)
	assert.Nil(t, err)
	assert.NotNil(t, resourceFromDB2)
	assert.Equal(t, resource.ID, resourceFromDB2.ID)

	// Get all
	resources, err := db.GetResources()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resources))

	// Count
	count, err := db.CountResources()
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
}

func TestGetResourcesByNamespace(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	resource1 := core.CreateResource("ExecutorDeployment", "web-1", "production")
	resource1.SetSpec("image", "nginx:1.21")

	resource2 := core.CreateResource("ExecutorDeployment", "web-2", "production")
	resource2.SetSpec("image", "nginx:1.22")

	resource3 := core.CreateResource("ExecutorDeployment", "web-3", "staging")
	resource3.SetSpec("image", "nginx:1.21")

	err = db.AddResource(resource1)
	assert.Nil(t, err)
	err = db.AddResource(resource2)
	assert.Nil(t, err)
	err = db.AddResource(resource3)
	assert.Nil(t, err)

	// Get by namespace
	prodResources, err := db.GetResourcesByNamespace("production")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(prodResources))

	stagingResources, err := db.GetResourcesByNamespace("staging")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(stagingResources))

	// Count by namespace
	prodCount, err := db.CountResourcesByNamespace("production")
	assert.Nil(t, err)
	assert.Equal(t, 2, prodCount)
}

func TestGetResourcesByKind(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	resource1 := core.CreateResource("ExecutorDeployment", "web-1", "production")
	resource1.SetSpec("image", "nginx:1.21")

	resource2 := core.CreateResource("Database", "db-1", "production")
	resource2.SetSpec("engine", "postgres")

	err = db.AddResource(resource1)
	assert.Nil(t, err)
	err = db.AddResource(resource2)
	assert.Nil(t, err)

	// Get by kind
	executorDeployments, err := db.GetResourcesByKind("ExecutorDeployment")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(executorDeployments))

	databases, err := db.GetResourcesByKind("Database")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(databases))
}

func TestUpdateResource(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	resource := core.CreateResource("ExecutorDeployment", "web-server", "production")
	resource.SetSpec("replicas", 3)

	err = db.AddResource(resource)
	assert.Nil(t, err)

	// Update resource
	resource.SetSpec("replicas", 5)
	err = db.UpdateResource(resource)
	assert.Nil(t, err)

	// Verify update
	resourceFromDB, err := db.GetResourceByID(resource.ID)
	assert.Nil(t, err)
	replicas, ok := resourceFromDB.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, float64(5), replicas)
}

func TestUpdateResourceStatus(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	resource := core.CreateResource("ExecutorDeployment", "web-server", "production")
	resource.SetSpec("replicas", 3)
	resource.SetStatus("phase", "Pending")

	err = db.AddResource(resource)
	assert.Nil(t, err)

	// Update status only
	newStatus := map[string]interface{}{
		"phase": "Running",
		"ready": 3,
	}
	err = db.UpdateResourceStatus(resource.ID, newStatus)
	assert.Nil(t, err)

	// Verify update
	resourceFromDB, err := db.GetResourceByID(resource.ID)
	assert.Nil(t, err)
	phase, ok := resourceFromDB.GetStatus("phase")
	assert.True(t, ok)
	assert.Equal(t, "Running", phase)
	ready, ok := resourceFromDB.GetStatus("ready")
	assert.True(t, ok)
	assert.Equal(t, float64(3), ready)
}

func TestRemoveResource(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	resource := core.CreateResource("ExecutorDeployment", "web-server", "production")
	err = db.AddResource(resource)
	assert.Nil(t, err)

	// Remove by ID
	err = db.RemoveResourceByID(resource.ID)
	assert.Nil(t, err)

	// Verify removed
	resourceFromDB, err := db.GetResourceByID(resource.ID)
	assert.Nil(t, err)
	assert.Nil(t, resourceFromDB)

	count, err := db.CountResources()
	assert.Nil(t, err)
	assert.Equal(t, 0, count)
}

func TestRemoveResourcesByNamespace(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	resource1 := core.CreateResource("ExecutorDeployment", "web-1", "production")
	resource2 := core.CreateResource("ExecutorDeployment", "web-2", "production")
	resource3 := core.CreateResource("ExecutorDeployment", "web-3", "staging")

	err = db.AddResource(resource1)
	assert.Nil(t, err)
	err = db.AddResource(resource2)
	assert.Nil(t, err)
	err = db.AddResource(resource3)
	assert.Nil(t, err)

	// Remove production namespace
	err = db.RemoveResourcesByNamespace("production")
	assert.Nil(t, err)

	// Verify
	prodCount, err := db.CountResourcesByNamespace("production")
	assert.Nil(t, err)
	assert.Equal(t, 0, prodCount)

	stagingCount, err := db.CountResourcesByNamespace("staging")
	assert.Nil(t, err)
	assert.Equal(t, 1, stagingCount)
}
