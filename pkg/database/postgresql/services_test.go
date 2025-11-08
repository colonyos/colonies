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

	service := core.CreateResource("ExecutorDeployment", "web-server", "production")
	service.SetSpec("image", "nginx:1.21")
	service.SetSpec("replicas", 3)
	service.SetStatus("phase", "Running")

	err = db.AddResource(service)
	assert.Nil(t, err)

	// Get by ID
	resourceFromDB, err := db.GetResourceByID(service.ID)
	assert.Nil(t, err)
	assert.NotNil(t, resourceFromDB)
	assert.Equal(t, service.ID, resourceFromDB.ID)
	assert.Equal(t, service.Metadata.Name, resourceFromDB.Metadata.Name)
	assert.Equal(t, service.Metadata.Namespace, resourceFromDB.Metadata.Namespace)
	assert.Equal(t, service.Kind, resourceFromDB.Kind)

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
	resourceFromDB2, err := db.GetResourceByName(service.Metadata.Namespace, service.Metadata.Name)
	assert.Nil(t, err)
	assert.NotNil(t, resourceFromDB2)
	assert.Equal(t, service.ID, resourceFromDB2.ID)

	// Get all
	services, err := db.GetResources()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(services))

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

	service := core.CreateResource("ExecutorDeployment", "web-server", "production")
	service.SetSpec("replicas", 3)

	err = db.AddResource(service)
	assert.Nil(t, err)

	// Update service
	service.SetSpec("replicas", 5)
	err = db.UpdateResource(service)
	assert.Nil(t, err)

	// Verify update
	resourceFromDB, err := db.GetResourceByID(service.ID)
	assert.Nil(t, err)
	replicas, ok := resourceFromDB.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, float64(5), replicas)
}

func TestUpdateResourceStatus(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	service := core.CreateResource("ExecutorDeployment", "web-server", "production")
	service.SetSpec("replicas", 3)
	service.SetStatus("phase", "Pending")

	err = db.AddResource(service)
	assert.Nil(t, err)

	// Update status only
	newStatus := map[string]interface{}{
		"phase": "Running",
		"ready": 3,
	}
	err = db.UpdateResourceStatus(service.ID, newStatus)
	assert.Nil(t, err)

	// Verify update
	resourceFromDB, err := db.GetResourceByID(service.ID)
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

	service := core.CreateResource("ExecutorDeployment", "web-server", "production")
	err = db.AddResource(service)
	assert.Nil(t, err)

	// Remove by ID
	err = db.RemoveResourceByID(service.ID)
	assert.Nil(t, err)

	// Verify removed
	resourceFromDB, err := db.GetResourceByID(service.ID)
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

func TestAddGetResourceHistory(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a service
	service := core.CreateResource("ExecutorDeployment", "web-server", "production")
	service.SetSpec("replicas", 3)
	service.SetStatus("phase", "Running")

	err = db.AddResource(service)
	assert.Nil(t, err)

	t.Logf("Service generation after creation: %d", service.Metadata.Generation)

	// Create history entry for service creation
	history := core.CreateResourceHistory(service, "test-user", "create")
	t.Logf("Creating history with generation: %d, ID: %s", history.Generation, history.ID)
	err = db.AddResourceHistory(history)
	assert.Nil(t, err)

	// Get history
	histories, err := db.GetResourceHistory(service.ID, 10)
	assert.Nil(t, err)
	t.Logf("Got %d history entries:", len(histories))
	for i, h := range histories {
		t.Logf("  History[%d]: ID=%s, Generation=%d, ChangeType=%s, ChangedBy=%s", i, h.ID, h.Generation, h.ChangeType, h.ChangedBy)
	}
	assert.Equal(t, 1, len(histories))
	assert.Equal(t, service.ID, histories[0].ResourceID)
	assert.Equal(t, "ExecutorDeployment", histories[0].Kind)
	assert.Equal(t, "production", histories[0].Namespace)
	assert.Equal(t, "web-server", histories[0].Name)
	assert.Equal(t, service.Metadata.Generation, histories[0].Generation)
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

func TestResourceHistoryMultipleVersions(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a service
	service := core.CreateResource("ExecutorDeployment", "web-server", "production")
	service.SetSpec("replicas", 3)

	err = db.AddResource(service)
	assert.Nil(t, err)

	initialGen := service.Metadata.Generation

	// Create initial history entry
	history1 := core.CreateResourceHistory(service, "user1", "create")
	err = db.AddResourceHistory(history1)
	assert.Nil(t, err)

	// Update service
	service.SetSpec("replicas", 5)
	service.Metadata.Generation = initialGen + 1

	// Create second history entry
	history2 := core.CreateResourceHistory(service, "user2", "update")
	err = db.AddResourceHistory(history2)
	assert.Nil(t, err)

	// Update again
	service.SetSpec("replicas", 10)
	service.Metadata.Generation = initialGen + 2

	// Create third history entry
	history3 := core.CreateResourceHistory(service, "user3", "update")
	err = db.AddResourceHistory(history3)
	assert.Nil(t, err)

	// Get all history (no limit)
	allHistories, err := db.GetResourceHistory(service.ID, 0)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(allHistories))

	// Verify they're ordered by timestamp DESC (most recent first)
	assert.Equal(t, initialGen+2, allHistories[0].Generation)
	assert.Equal(t, initialGen+1, allHistories[1].Generation)
	assert.Equal(t, initialGen, allHistories[2].Generation)

	// Get limited history
	limitedHistories, err := db.GetResourceHistory(service.ID, 2)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(limitedHistories))
	assert.Equal(t, initialGen+2, limitedHistories[0].Generation)
	assert.Equal(t, initialGen+1, limitedHistories[1].Generation)
}

func TestGetResourceHistoryByGeneration(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a service
	service := core.CreateResource("ExecutorDeployment", "web-server", "production")
	service.SetSpec("replicas", 3)

	err = db.AddResource(service)
	assert.Nil(t, err)

	// Create multiple history entries
	history1 := core.CreateResourceHistory(service, "user1", "create")
	err = db.AddResourceHistory(history1)
	assert.Nil(t, err)

	service.SetSpec("replicas", 5)
	service.Metadata.Generation = 2
	history2 := core.CreateResourceHistory(service, "user2", "update")
	err = db.AddResourceHistory(history2)
	assert.Nil(t, err)

	// Get specific generation
	historyGen2, err := db.GetResourceHistoryByGeneration(service.ID, 2)
	assert.Nil(t, err)
	assert.NotNil(t, historyGen2)
	assert.Equal(t, int64(2), historyGen2.Generation)
	assert.Equal(t, "user2", historyGen2.ChangedBy)

	replicas, ok := historyGen2.Spec["replicas"]
	assert.True(t, ok)
	assert.Equal(t, float64(5), replicas)

	// Get generation that doesn't exist
	historyGen99, err := db.GetResourceHistoryByGeneration(service.ID, 99)
	assert.Nil(t, err)
	assert.Nil(t, historyGen99)
}

func TestRemoveResourceHistory(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a service
	service := core.CreateResource("ExecutorDeployment", "web-server", "production")
	err = db.AddResource(service)
	assert.Nil(t, err)

	// Create history entries
	history1 := core.CreateResourceHistory(service, "user1", "create")
	err = db.AddResourceHistory(history1)
	assert.Nil(t, err)

	service.Metadata.Generation = 2
	history2 := core.CreateResourceHistory(service, "user2", "update")
	err = db.AddResourceHistory(history2)
	assert.Nil(t, err)

	// Verify history exists
	histories, err := db.GetResourceHistory(service.ID, 0)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(histories))

	// Remove all history for this service
	err = db.RemoveResourceHistory(service.ID)
	assert.Nil(t, err)

	// Verify history is removed
	historiesAfter, err := db.GetResourceHistory(service.ID, 0)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(historiesAfter))
}

func TestResourceHistoryWithStatusChanges(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a service
	service := core.CreateResource("ExecutorDeployment", "web-server", "production")
	service.SetSpec("replicas", 3)
	service.SetStatus("phase", "Pending")
	service.SetStatus("ready", 0)

	err = db.AddResource(service)
	assert.Nil(t, err)

	// Create initial history
	history1 := core.CreateResourceHistory(service, "controller", "create")
	err = db.AddResourceHistory(history1)
	assert.Nil(t, err)

	// Update status only (status update via reconciliation)
	service.SetStatus("phase", "Running")
	service.SetStatus("ready", 3)
	service.Metadata.Generation = 2

	history2 := core.CreateResourceHistory(service, "reconciler", "status-update")
	err = db.AddResourceHistory(history2)
	assert.Nil(t, err)

	// Get history and verify status changes are tracked
	histories, err := db.GetResourceHistory(service.ID, 0)
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
