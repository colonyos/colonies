package blueprint_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddBlueprintDefinitionSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Create a BlueprintDefinition for colony1
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
	sd.Metadata.ColonyName = env.Colony1Name

	// Only colony owner should be able to add BlueprintDefinitions

	// Try with executor key - should FAIL
	_, err := client.AddBlueprintDefinition(sd, env.Executor1PrvKey)
	assert.NotNil(t, err)

	// Try with another colony's owner key - should FAIL
	_, err = client.AddBlueprintDefinition(sd, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with correct colony owner key - should SUCCEED
	_, err = client.AddBlueprintDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetBlueprintDefinitionSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Create and add BlueprintDefinition for colony1
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
	sd.Metadata.ColonyName = env.Colony1Name

	_, err := client.AddBlueprintDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Colony members should be able to get BlueprintDefinitions

	// Try with executor from same colony - should SUCCEED
	_, err = client.GetBlueprintDefinition(env.Colony1Name, sd.Metadata.Name, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Try with executor from different colony - should FAIL
	_, err = client.GetBlueprintDefinition(env.Colony1Name, sd.Metadata.Name, env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	_, err = client.GetBlueprintDefinition(env.Colony1Name, sd.Metadata.Name, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with correct colony owner - should SUCCEED
	_, err = client.GetBlueprintDefinition(env.Colony1Name, sd.Metadata.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestAddBlueprintSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	// Setup BlueprintDefinition for colony1
	sd := core.CreateBlueprintDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	sd.Metadata.ColonyName = env.Colony1Name

	_, err := client.AddBlueprintDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Create a Blueprint instance for colony1
	blueprint := core.CreateBlueprint("Database", "my-database", env.Colony1Name)
	blueprint.SetSpec("host", "localhost")

	// Colony members should be able to add Blueprints

	// Try with executor from different colony - should FAIL
	_, err = client.AddBlueprint(blueprint, env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	_, err = client.AddBlueprint(blueprint, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with executor from same colony - should SUCCEED
	_, err = client.AddBlueprint(blueprint, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Create another service to test with colony owner
	blueprint2 := core.CreateBlueprint("Database", "another-database", env.Colony1Name)
	blueprint2.SetSpec("host", "remotehost")

	// Try with colony owner - should also SUCCEED
	_, err = client.AddBlueprint(blueprint2, env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetBlueprintSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Setup BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	sd.Metadata.ColonyName = env.Colony1Name
	_, err := client.AddBlueprintDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Add Blueprint to colony1
	blueprint := core.CreateBlueprint("Database", "my-database", env.Colony1Name)
	blueprint.SetSpec("host", "localhost")
	_, err = client.AddBlueprint(blueprint, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Try to get service from different colony - should FAIL
	_, err = client.GetBlueprint(env.Colony1Name, blueprint.Metadata.Name, env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	_, err = client.GetBlueprint(env.Colony1Name, blueprint.Metadata.Name, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with executor from same colony - should SUCCEED
	_, err = client.GetBlueprint(env.Colony1Name, blueprint.Metadata.Name, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Try with colony owner - should SUCCEED
	_, err = client.GetBlueprint(env.Colony1Name, blueprint.Metadata.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetBlueprintsSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Setup BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	sd.Metadata.ColonyName = env.Colony1Name
	_, err := client.AddBlueprintDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Add Blueprints to colony1
	blueprint1 := core.CreateBlueprint("Database", "db1", env.Colony1Name)
	_, err = client.AddBlueprint(blueprint1, env.Executor1PrvKey)
	assert.Nil(t, err)

	blueprint2 := core.CreateBlueprint("Database", "db2", env.Colony1Name)
	_, err = client.AddBlueprint(blueprint2, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Try to list blueprints from different colony - should FAIL
	_, err = client.GetBlueprints(env.Colony1Name, "Database", env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	_, err = client.GetBlueprints(env.Colony1Name, "Database", env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with executor from same colony - should SUCCEED
	blueprints, err := client.GetBlueprints(env.Colony1Name, "Database", env.Executor1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, blueprints, 2)

	// Try with colony owner - should SUCCEED
	blueprints, err = client.GetBlueprints(env.Colony1Name, "Database", env.Colony1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, blueprints, 2)

	server.Shutdown()
	<-done
}

func TestUpdateBlueprintSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Setup BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	sd.Metadata.ColonyName = env.Colony1Name
	_, err := client.AddBlueprintDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Add Blueprint to colony1
	blueprint := core.CreateBlueprint("Database", "my-database", env.Colony1Name)
	blueprint.SetSpec("host", "localhost")
	addedBlueprint, err := client.AddBlueprint(blueprint, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Update the blueprint spec
	addedBlueprint.SetSpec("port", 5432)

	// Try to update from different colony executor - should FAIL
	_, err = client.UpdateBlueprint(addedBlueprint, env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	_, err = client.UpdateBlueprint(addedBlueprint, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with executor from same colony - should SUCCEED
	_, err = client.UpdateBlueprint(addedBlueprint, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Update again with colony owner - should SUCCEED
	addedBlueprint.SetSpec("port", 5433)
	_, err = client.UpdateBlueprint(addedBlueprint, env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestRemoveBlueprintSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Setup BlueprintDefinition
	sd := core.CreateBlueprintDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	sd.Metadata.ColonyName = env.Colony1Name
	_, err := client.AddBlueprintDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Add Blueprints to colony1
	blueprint1 := core.CreateBlueprint("Database", "db1", env.Colony1Name)
	_, err = client.AddBlueprint(blueprint1, env.Executor1PrvKey)
	assert.Nil(t, err)

	blueprint2 := core.CreateBlueprint("Database", "db2", env.Colony1Name)
	_, err = client.AddBlueprint(blueprint2, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Try to remove from different colony executor - should FAIL
	err = client.RemoveBlueprint(env.Colony1Name, "db1", env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Try with colony owner from different colony - should FAIL
	err = client.RemoveBlueprint(env.Colony1Name, "db1", env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try with executor from same colony - should SUCCEED
	err = client.RemoveBlueprint(env.Colony1Name, "db1", env.Executor1PrvKey)
	assert.Nil(t, err)

	// Verify it was removed
	_, err = client.GetBlueprint(env.Colony1Name, "db1", env.Executor1PrvKey)
	assert.NotNil(t, err)

	// Remove with colony owner - should SUCCEED
	err = client.RemoveBlueprint(env.Colony1Name, "db2", env.Colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestCrossColonyBlueprintIsolation(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Create BlueprintDefinitions for both colonies
	sd1 := core.CreateBlueprintDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	sd1.Metadata.ColonyName = env.Colony1Name
	_, err := client.AddBlueprintDefinition(sd1, env.Colony1PrvKey)
	assert.Nil(t, err)

	sd2 := core.CreateBlueprintDefinition(
		"database",
		"example.com",
		"v1",
		"Database",
		"databases",
		"Namespaced",
		"test_executor_type",
		"reconcile_database",
	)
	sd2.Metadata.ColonyName = env.Colony2Name
	_, err = client.AddBlueprintDefinition(sd2, env.Colony2PrvKey)
	assert.Nil(t, err)

	// Add Blueprints to both colonies with same name
	blueprint1 := core.CreateBlueprint("Database", "shared-name", env.Colony1Name)
	blueprint1.SetSpec("colonyId", "colony1")
	_, err = client.AddBlueprint(blueprint1, env.Executor1PrvKey)
	assert.Nil(t, err)

	blueprint2 := core.CreateBlueprint("Database", "shared-name", env.Colony2Name)
	blueprint2.SetSpec("colonyId", "colony2")
	_, err = client.AddBlueprint(blueprint2, env.Executor2PrvKey)
	assert.Nil(t, err)

	// Each colony should only see its own service
	s1, err := client.GetBlueprint(env.Colony1Name, "shared-name", env.Executor1PrvKey)
	assert.Nil(t, err)
	colonyId1, _ := s1.GetSpec("colonyId")
	assert.Equal(t, "colony1", colonyId1)

	s2, err := client.GetBlueprint(env.Colony2Name, "shared-name", env.Executor2PrvKey)
	assert.Nil(t, err)
	colonyId2, _ := s2.GetSpec("colonyId")
	assert.Equal(t, "colony2", colonyId2)

	// Verify isolation - executor1 cannot see colony2 services
	_, err = client.GetBlueprint(env.Colony2Name, "shared-name", env.Executor1PrvKey)
	assert.NotNil(t, err)

	// Verify isolation - executor2 cannot see colony1 services
	_, err = client.GetBlueprint(env.Colony1Name, "shared-name", env.Executor2PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestBlueprintDefinitionOnlyColonyOwner(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Create additional executor in colony1
	executor3, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor3, env.Colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.Colony1Name, executor3.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Create BlueprintDefinition
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
	sd.Metadata.ColonyName = env.Colony1Name

	// None of the executors should be able to add BlueprintDefinitions
	_, err = client.AddBlueprintDefinition(sd, env.Executor1PrvKey)
	assert.NotNil(t, err)

	_, err = client.AddBlueprintDefinition(sd, executor3PrvKey)
	assert.NotNil(t, err)

	// Only colony owner can add BlueprintDefinitions
	_, err = client.AddBlueprintDefinition(sd, env.Colony1PrvKey)
	assert.Nil(t, err)

	// But all colony members can read BlueprintDefinitions
	_, err = client.GetBlueprintDefinition(env.Colony1Name, sd.Metadata.Name, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetBlueprintDefinition(env.Colony1Name, sd.Metadata.Name, executor3PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}
