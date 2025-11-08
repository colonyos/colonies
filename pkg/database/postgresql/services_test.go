package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestAddGetServiceDefinition(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	sd := core.CreateServiceDefinition(
		"executordeployments.compute.colonies.io",
		"compute.colonies.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"executor-controller",
		"reconcile",
	)
	sd.Metadata.Namespace = "test-colony"

	err = db.AddServiceDefinition(sd)
	assert.Nil(t, err)

	// Get by ID
	sdFromDB, err := db.GetServiceDefinitionByID(sd.ID)
	assert.Nil(t, err)
	assert.NotNil(t, sdFromDB)
	assert.Equal(t, sd.ID, sdFromDB.ID)
	assert.Equal(t, sd.Metadata.Name, sdFromDB.Metadata.Name)
	assert.Equal(t, sd.Spec.Group, sdFromDB.Spec.Group)
	assert.Equal(t, sd.Spec.Version, sdFromDB.Spec.Version)

	// Get by name
	sdFromDB2, err := db.GetServiceDefinitionByName(sd.Metadata.Namespace, sd.Metadata.Name)
	assert.Nil(t, err)
	assert.NotNil(t, sdFromDB2)
	assert.Equal(t, sd.ID, sdFromDB2.ID)

	// Get all
	sds, err := db.GetServiceDefinitions()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sds))

	// Count
	count, err := db.CountServiceDefinitions()
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
}

func TestAddGetService(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	service := core.CreateService("ExecutorDeployment", "web-server", "production")
	service.SetSpec("image", "nginx:1.21")
	service.SetSpec("replicas", 3)
	service.SetStatus("phase", "Running")

	err = db.AddService(service)
	assert.Nil(t, err)

	// Get by ID
	serviceFromDB, err := db.GetServiceByID(service.ID)
	assert.Nil(t, err)
	assert.NotNil(t, serviceFromDB)
	assert.Equal(t, service.ID, serviceFromDB.ID)
	assert.Equal(t, service.Metadata.Name, serviceFromDB.Metadata.Name)
	assert.Equal(t, service.Metadata.Namespace, serviceFromDB.Metadata.Namespace)
	assert.Equal(t, service.Kind, serviceFromDB.Kind)

	// Verify spec
	image, ok := serviceFromDB.GetSpec("image")
	assert.True(t, ok)
	assert.Equal(t, "nginx:1.21", image)

	replicas, ok := serviceFromDB.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, float64(3), replicas) // JSON unmarshaling converts to float64

	// Verify status
	phase, ok := serviceFromDB.GetStatus("phase")
	assert.True(t, ok)
	assert.Equal(t, "Running", phase)

	// Get by name
	serviceFromDB2, err := db.GetServiceByName(service.Metadata.Namespace, service.Metadata.Name)
	assert.Nil(t, err)
	assert.NotNil(t, serviceFromDB2)
	assert.Equal(t, service.ID, serviceFromDB2.ID)

	// Get all
	services, err := db.GetServices()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(services))

	// Count
	count, err := db.CountServices()
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
}

func TestGetServicesByNamespace(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	service1 := core.CreateService("ExecutorDeployment", "web-1", "production")
	service1.SetSpec("image", "nginx:1.21")

	service2 := core.CreateService("ExecutorDeployment", "web-2", "production")
	service2.SetSpec("image", "nginx:1.22")

	service3 := core.CreateService("ExecutorDeployment", "web-3", "staging")
	service3.SetSpec("image", "nginx:1.21")

	err = db.AddService(service1)
	assert.Nil(t, err)
	err = db.AddService(service2)
	assert.Nil(t, err)
	err = db.AddService(service3)
	assert.Nil(t, err)

	// Get by namespace
	prodServices, err := db.GetServicesByNamespace("production")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(prodServices))

	stagingServices, err := db.GetServicesByNamespace("staging")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(stagingServices))

	// Count by namespace
	prodCount, err := db.CountServicesByNamespace("production")
	assert.Nil(t, err)
	assert.Equal(t, 2, prodCount)
}

func TestGetServicesByKind(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	service1 := core.CreateService("ExecutorDeployment", "web-1", "production")
	service1.SetSpec("image", "nginx:1.21")

	service2 := core.CreateService("Database", "db-1", "production")
	service2.SetSpec("engine", "postgres")

	err = db.AddService(service1)
	assert.Nil(t, err)
	err = db.AddService(service2)
	assert.Nil(t, err)

	// Get by kind
	executorDeployments, err := db.GetServicesByKind("ExecutorDeployment")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(executorDeployments))

	databases, err := db.GetServicesByKind("Database")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(databases))
}

func TestUpdateService(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	service := core.CreateService("ExecutorDeployment", "web-server", "production")
	service.SetSpec("replicas", 3)

	err = db.AddService(service)
	assert.Nil(t, err)

	// Update service
	service.SetSpec("replicas", 5)
	err = db.UpdateService(service)
	assert.Nil(t, err)

	// Verify update
	serviceFromDB, err := db.GetServiceByID(service.ID)
	assert.Nil(t, err)
	replicas, ok := serviceFromDB.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, float64(5), replicas)
}

func TestUpdateServiceStatus(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	service := core.CreateService("ExecutorDeployment", "web-server", "production")
	service.SetSpec("replicas", 3)
	service.SetStatus("phase", "Pending")

	err = db.AddService(service)
	assert.Nil(t, err)

	// Update status only
	newStatus := map[string]interface{}{
		"phase": "Running",
		"ready": 3,
	}
	err = db.UpdateServiceStatus(service.ID, newStatus)
	assert.Nil(t, err)

	// Verify update
	serviceFromDB, err := db.GetServiceByID(service.ID)
	assert.Nil(t, err)
	phase, ok := serviceFromDB.GetStatus("phase")
	assert.True(t, ok)
	assert.Equal(t, "Running", phase)
	ready, ok := serviceFromDB.GetStatus("ready")
	assert.True(t, ok)
	assert.Equal(t, float64(3), ready)
}

func TestRemoveService(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	service := core.CreateService("ExecutorDeployment", "web-server", "production")
	err = db.AddService(service)
	assert.Nil(t, err)

	// Remove by ID
	err = db.RemoveServiceByID(service.ID)
	assert.Nil(t, err)

	// Verify removed
	serviceFromDB, err := db.GetServiceByID(service.ID)
	assert.Nil(t, err)
	assert.Nil(t, serviceFromDB)

	count, err := db.CountServices()
	assert.Nil(t, err)
	assert.Equal(t, 0, count)
}

func TestRemoveServicesByNamespace(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	service1 := core.CreateService("ExecutorDeployment", "web-1", "production")
	service2 := core.CreateService("ExecutorDeployment", "web-2", "production")
	service3 := core.CreateService("ExecutorDeployment", "web-3", "staging")

	err = db.AddService(service1)
	assert.Nil(t, err)
	err = db.AddService(service2)
	assert.Nil(t, err)
	err = db.AddService(service3)
	assert.Nil(t, err)

	// Remove production namespace
	err = db.RemoveServicesByNamespace("production")
	assert.Nil(t, err)

	// Verify
	prodCount, err := db.CountServicesByNamespace("production")
	assert.Nil(t, err)
	assert.Equal(t, 0, prodCount)

	stagingCount, err := db.CountServicesByNamespace("staging")
	assert.Nil(t, err)
	assert.Equal(t, 1, stagingCount)
}

func TestAddGetServiceHistory(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a service
	service := core.CreateService("ExecutorDeployment", "web-server", "production")
	service.SetSpec("replicas", 3)
	service.SetStatus("phase", "Running")

	err = db.AddService(service)
	assert.Nil(t, err)

	t.Logf("Service generation after creation: %d", service.Metadata.Generation)

	// Create history entry for service creation
	history := core.CreateServiceHistory(service, "test-user", "create")
	t.Logf("Creating history with generation: %d, ID: %s", history.Generation, history.ID)
	err = db.AddServiceHistory(history)
	assert.Nil(t, err)

	// Get history
	histories, err := db.GetServiceHistory(service.ID, 10)
	assert.Nil(t, err)
	t.Logf("Got %d history entries:", len(histories))
	for i, h := range histories {
		t.Logf("  History[%d]: ID=%s, Generation=%d, ChangeType=%s, ChangedBy=%s", i, h.ID, h.Generation, h.ChangeType, h.ChangedBy)
	}
	assert.Equal(t, 1, len(histories))
	assert.Equal(t, service.ID, histories[0].ServiceID)
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

func TestServiceHistoryMultipleVersions(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a service
	service := core.CreateService("ExecutorDeployment", "web-server", "production")
	service.SetSpec("replicas", 3)

	err = db.AddService(service)
	assert.Nil(t, err)

	initialGen := service.Metadata.Generation

	// Create initial history entry
	history1 := core.CreateServiceHistory(service, "user1", "create")
	err = db.AddServiceHistory(history1)
	assert.Nil(t, err)

	// Update service
	service.SetSpec("replicas", 5)
	service.Metadata.Generation = initialGen + 1

	// Create second history entry
	history2 := core.CreateServiceHistory(service, "user2", "update")
	err = db.AddServiceHistory(history2)
	assert.Nil(t, err)

	// Update again
	service.SetSpec("replicas", 10)
	service.Metadata.Generation = initialGen + 2

	// Create third history entry
	history3 := core.CreateServiceHistory(service, "user3", "update")
	err = db.AddServiceHistory(history3)
	assert.Nil(t, err)

	// Get all history (no limit)
	allHistories, err := db.GetServiceHistory(service.ID, 0)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(allHistories))

	// Verify they're ordered by timestamp DESC (most recent first)
	assert.Equal(t, initialGen+2, allHistories[0].Generation)
	assert.Equal(t, initialGen+1, allHistories[1].Generation)
	assert.Equal(t, initialGen, allHistories[2].Generation)

	// Get limited history
	limitedHistories, err := db.GetServiceHistory(service.ID, 2)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(limitedHistories))
	assert.Equal(t, initialGen+2, limitedHistories[0].Generation)
	assert.Equal(t, initialGen+1, limitedHistories[1].Generation)
}

func TestGetServiceHistoryByGeneration(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a service
	service := core.CreateService("ExecutorDeployment", "web-server", "production")
	service.SetSpec("replicas", 3)

	err = db.AddService(service)
	assert.Nil(t, err)

	// Create multiple history entries
	history1 := core.CreateServiceHistory(service, "user1", "create")
	err = db.AddServiceHistory(history1)
	assert.Nil(t, err)

	service.SetSpec("replicas", 5)
	service.Metadata.Generation = 2
	history2 := core.CreateServiceHistory(service, "user2", "update")
	err = db.AddServiceHistory(history2)
	assert.Nil(t, err)

	// Get specific generation
	historyGen2, err := db.GetServiceHistoryByGeneration(service.ID, 2)
	assert.Nil(t, err)
	assert.NotNil(t, historyGen2)
	assert.Equal(t, int64(2), historyGen2.Generation)
	assert.Equal(t, "user2", historyGen2.ChangedBy)

	replicas, ok := historyGen2.Spec["replicas"]
	assert.True(t, ok)
	assert.Equal(t, float64(5), replicas)

	// Get generation that doesn't exist
	historyGen99, err := db.GetServiceHistoryByGeneration(service.ID, 99)
	assert.Nil(t, err)
	assert.Nil(t, historyGen99)
}

func TestRemoveServiceHistory(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a service
	service := core.CreateService("ExecutorDeployment", "web-server", "production")
	err = db.AddService(service)
	assert.Nil(t, err)

	// Create history entries
	history1 := core.CreateServiceHistory(service, "user1", "create")
	err = db.AddServiceHistory(history1)
	assert.Nil(t, err)

	service.Metadata.Generation = 2
	history2 := core.CreateServiceHistory(service, "user2", "update")
	err = db.AddServiceHistory(history2)
	assert.Nil(t, err)

	// Verify history exists
	histories, err := db.GetServiceHistory(service.ID, 0)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(histories))

	// Remove all history for this service
	err = db.RemoveServiceHistory(service.ID)
	assert.Nil(t, err)

	// Verify history is removed
	historiesAfter, err := db.GetServiceHistory(service.ID, 0)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(historiesAfter))
}

func TestServiceHistoryWithStatusChanges(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create a service
	service := core.CreateService("ExecutorDeployment", "web-server", "production")
	service.SetSpec("replicas", 3)
	service.SetStatus("phase", "Pending")
	service.SetStatus("ready", 0)

	err = db.AddService(service)
	assert.Nil(t, err)

	// Create initial history
	history1 := core.CreateServiceHistory(service, "controller", "create")
	err = db.AddServiceHistory(history1)
	assert.Nil(t, err)

	// Update status only (status update via reconciliation)
	service.SetStatus("phase", "Running")
	service.SetStatus("ready", 3)
	service.Metadata.Generation = 2

	history2 := core.CreateServiceHistory(service, "reconciler", "status-update")
	err = db.AddServiceHistory(history2)
	assert.Nil(t, err)

	// Get history and verify status changes are tracked
	histories, err := db.GetServiceHistory(service.ID, 0)
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
