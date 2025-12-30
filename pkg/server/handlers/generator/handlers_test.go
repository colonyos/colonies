package generator_test

import (
	"strconv"
	"testing"

	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddGenerator(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	// Create an executor for the colony
	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	addedExecutor, err := client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	// Approve the executor
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	// Create a generator using the executor's credentials
	generator := utils.FakeGenerator(t, colony.Name, addedExecutor.ID, addedExecutor.Name)
	addedGenerator, err := client.AddGenerator(generator, executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)
	assert.Equal(t, generator.Name, addedGenerator.Name)
	assert.Equal(t, generator.ColonyName, addedGenerator.ColonyName)

	s.Shutdown()
	<-done
}

func TestGetGenerator(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	// Create an executor for the colony
	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	addedExecutor, err := client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	// Approve the executor
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	// Create a generator using the executor's credentials
	generator := utils.FakeGenerator(t, colony.Name, addedExecutor.ID, addedExecutor.Name)
	addedGenerator, err := client.AddGenerator(generator, executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	retrievedGenerator, err := client.GetGenerator(addedGenerator.ID, executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedGenerator)
	assert.Equal(t, addedGenerator.ID, retrievedGenerator.ID)
	assert.Equal(t, addedGenerator.Name, retrievedGenerator.Name)
	assert.Equal(t, addedGenerator.ColonyName, retrievedGenerator.ColonyName)
	assert.Equal(t, addedGenerator.WorkflowSpec, retrievedGenerator.WorkflowSpec)

	s.Shutdown()
	<-done
}

func TestGetGenerators(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	// Create an executor for the colony
	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	addedExecutor, err := client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	// Approve the executor
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	// Create generators using the executor's credentials
	generator1 := utils.FakeGenerator(t, colony.Name, addedExecutor.ID, addedExecutor.Name)
	generator1.Name = "test_generator1"
	addedGenerator1, err := client.AddGenerator(generator1, executorPrvKey)
	assert.Nil(t, err)

	generator2 := utils.FakeGenerator(t, colony.Name, addedExecutor.ID, addedExecutor.Name)
	generator2.Name = "test_generator2"
	addedGenerator2, err := client.AddGenerator(generator2, executorPrvKey)
	assert.Nil(t, err)

	// GetGenerators requires executor permissions
	generators, err := client.GetGenerators(colony.Name, 10, executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, generators, 2)

	// Check that both generators are in the result
	found1 := false
	found2 := false
	for _, g := range generators {
		if g.ID == addedGenerator1.ID {
			found1 = true
		}
		if g.ID == addedGenerator2.ID {
			found2 = true
		}
	}
	assert.True(t, found1)
	assert.True(t, found2)

	s.Shutdown()
	<-done
}

func TestResolveGenerator(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	// Create an executor for the colony
	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	addedExecutor, err := client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	// Approve the executor
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	// Create a generator with a specific name
	generator := utils.FakeGenerator(t, colony.Name, addedExecutor.ID, addedExecutor.Name)
	generator.Name = "test_generator_name"
	addedGenerator, err := client.AddGenerator(generator, executorPrvKey)
	assert.Nil(t, err)

	// Resolve generator by name
	generatorFromServer, err := client.ResolveGenerator(colony.Name, "test_generator_name", executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, generatorFromServer.ID, addedGenerator.ID)
	assert.Equal(t, generatorFromServer.Name, addedGenerator.Name)

	s.Shutdown()
	<-done
}

func TestRemoveGenerator(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	// Create an executor for the colony
	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	addedExecutor, err := client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	// Approve the executor
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	// Create a generator using the executor's credentials
	generator := utils.FakeGenerator(t, colony.Name, addedExecutor.ID, addedExecutor.Name)
	addedGenerator, err := client.AddGenerator(generator, executorPrvKey)
	assert.Nil(t, err)

	err = client.RemoveGenerator(addedGenerator.ID, executorPrvKey)
	assert.Nil(t, err)

	// Try to get the removed generator - should fail
	_, err = client.GetGenerator(addedGenerator.ID, executorPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// Trigger mechanism tests

func TestAddGeneratorCounter(t *testing.T) {
	env, client, s, _, done := server.SetupTestEnv2(t)

	colonyName := env.ColonyName

	generator := utils.FakeGenerator(t, colonyName, env.ExecutorID, env.ExecutorName)
	generator.Trigger = 10
	addedGenerator, err := client.AddGenerator(generator, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	doneInc := make(chan bool)
	go func() {
		for i := 0; i < 73; i++ {
			err = client.PackGenerator(addedGenerator.ID, "arg"+strconv.Itoa(i), env.ExecutorPrvKey)
			assert.Nil(t, err)
		}
		doneInc <- true
	}()
	<-doneInc

	server.WaitForProcessGraphs(t, client, colonyName, addedGenerator.ID, env.ExecutorPrvKey, 7)

	graphs, err := client.GetWaitingProcessGraphs(colonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 7)

	s.Shutdown()
	<-done
}

func TestAddGeneratorTimeout(t *testing.T) {
	env, client, s, _, done := server.SetupTestEnv2(t)

	colonyName := env.ColonyName

	generator := utils.FakeGenerator(t, colonyName, env.ExecutorID, env.ExecutorName)
	generator.Trigger = 10
	generator.Timeout = 1
	addedGenerator, err := client.AddGenerator(generator, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	err = client.PackGenerator(addedGenerator.ID, "arg1", env.ExecutorPrvKey)
	assert.Nil(t, err)

	server.WaitForProcessGraphs(t, client, colonyName, addedGenerator.ID, env.ExecutorPrvKey, 1)

	graphs, err := client.GetWaitingProcessGraphs(colonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	s.Shutdown()
	<-done
}

// Tests that generator worker works correctly if it runs before it has been packed
// Also tests that the generator keeps on working
func TestAddGeneratorTimeout2(t *testing.T) {
	env, client, s, _, done := server.SetupTestEnv2(t)

	colonyName := env.ColonyName

	generator := utils.FakeGenerator(t, colonyName, env.ExecutorID, env.ExecutorName)
	generator.Trigger = 10
	generator.Timeout = 1
	addedGenerator, err := client.AddGenerator(generator, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	// Sleep so that generator worker runs before it has been packed
	// Note: Using WaitForProcessGraphs instead of time.Sleep for more reliable timing

	err = client.PackGenerator(addedGenerator.ID, "arg1", env.ExecutorPrvKey)
	assert.Nil(t, err)

	server.WaitForProcessGraphs(t, client, colonyName, addedGenerator.ID, env.ExecutorPrvKey, 1)

	graphs, err := client.GetWaitingProcessGraphs(colonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	err = client.PackGenerator(addedGenerator.ID, "arg1", env.ExecutorPrvKey)
	assert.Nil(t, err)

	server.WaitForProcessGraphs(t, client, colonyName, addedGenerator.ID, env.ExecutorPrvKey, 2)

	graphs, err = client.GetWaitingProcessGraphs(colonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 2)

	s.Shutdown()
	<-done
}

// This test tests a combination of trigger and timeout triggered workflows
func TestAddGeneratorTimeout3(t *testing.T) {
	env, client, s, _, done := server.SetupTestEnv2(t)

	colonyName := env.ColonyName

	generator := utils.FakeGenerator(t, colonyName, env.ExecutorID, env.ExecutorName)
	generator.Trigger = 10
	generator.Timeout = 1
	addedGenerator, err := client.AddGenerator(generator, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	// Since we are packing 12 args and the trigger is set to 10, a workflow should immediately be submitted
	for i := 0; i < 12; i++ {
		err = client.PackGenerator(addedGenerator.ID, "arg", env.ExecutorPrvKey)
		assert.Nil(t, err)
	}

	server.WaitForProcessGraphs(t, client, colonyName, addedGenerator.ID, env.ExecutorPrvKey, 1)

	graphs, err := client.GetWaitingProcessGraphs(colonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	// The second workflow should be submitted by the timeout trigger
	server.WaitForProcessGraphs(t, client, colonyName, addedGenerator.ID, env.ExecutorPrvKey, 2)

	graphs, err = client.GetWaitingProcessGraphs(colonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 2)

	process, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, process.FunctionSpec.Args, 10)

	process, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, process.FunctionSpec.Args, 2)

	s.Shutdown()
	<-done
}

// TestAddGeneratorWithUserAsInitiator tests that a user can create a generator (covers resolveInitiator user path)
func TestAddGeneratorWithUserAsInitiator(t *testing.T) {
	env, client, s, _, done := server.SetupTestEnv2(t)

	// Create a user
	user, userPrvKey, err := utils.CreateTestUserWithKey(env.ColonyName, "generator-user")
	assert.Nil(t, err)
	_, err = client.AddUser(user, env.ColonyPrvKey)
	assert.Nil(t, err)

	// User creates a generator
	generator := utils.FakeGenerator(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	addedGenerator, err := client.AddGenerator(generator, userPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)
	assert.Equal(t, "generator-user", addedGenerator.InitiatorName)

	s.Shutdown()
	<-done
}

// TestGetGeneratorNotFound tests getting a non-existent generator
func TestGetGeneratorNotFound(t *testing.T) {
	env, client, s, _, done := server.SetupTestEnv2(t)

	_, err := client.GetGenerator("nonexistent-generator-id", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestRemoveGeneratorNotFound tests removing a non-existent generator
func TestRemoveGeneratorNotFound(t *testing.T) {
	env, client, s, _, done := server.SetupTestEnv2(t)

	err := client.RemoveGenerator("nonexistent-generator-id", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestPackGeneratorNotFound tests packing a non-existent generator
func TestPackGeneratorNotFound(t *testing.T) {
	env, client, s, _, done := server.SetupTestEnv2(t)

	err := client.PackGenerator("nonexistent-generator-id", "arg", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestResolveGeneratorNotFound tests resolving a non-existent generator
func TestResolveGeneratorNotFound(t *testing.T) {
	env, client, s, _, done := server.SetupTestEnv2(t)

	_, err := client.ResolveGenerator(env.ColonyName, "nonexistent-generator", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetGeneratorsEmpty tests getting generators when none exist
func TestGetGeneratorsEmpty(t *testing.T) {
	env, client, s, _, done := server.SetupTestEnv2(t)

	generators, err := client.GetGenerators(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, generators, 0)

	s.Shutdown()
	<-done
}

// TestAddGeneratorUnauthorized tests that non-members cannot add generators
func TestAddGeneratorUnauthorized(t *testing.T) {
	env, client, s, serverPrvKey, done := server.SetupTestEnv2(t)

	// Create another colony
	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colony2PrvKey)
	assert.Nil(t, err)

	// Try to add generator to a different colony
	generator := utils.FakeGenerator(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	_, err = client.AddGenerator(generator, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetGeneratorUnauthorized tests that non-members cannot get generator details
func TestGetGeneratorUnauthorized(t *testing.T) {
	env, client, s, serverPrvKey, done := server.SetupTestEnv2(t)

	// Add a generator
	generator := utils.FakeGenerator(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	addedGenerator, err := client.AddGenerator(generator, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Create another colony
	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colony2PrvKey)
	assert.Nil(t, err)

	// Try to get generator from a different colony
	_, err = client.GetGenerator(addedGenerator.ID, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetGeneratorsUnauthorized tests that non-members cannot get generators
func TestGetGeneratorsUnauthorized(t *testing.T) {
	env, client, s, serverPrvKey, done := server.SetupTestEnv2(t)

	// Create another colony
	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colony2PrvKey)
	assert.Nil(t, err)

	// Try to get generators from a different colony
	_, err = client.GetGenerators(env.ColonyName, 100, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestRemoveGeneratorUnauthorized tests that non-members cannot remove generators
func TestRemoveGeneratorUnauthorized(t *testing.T) {
	env, client, s, serverPrvKey, done := server.SetupTestEnv2(t)

	// Add a generator
	generator := utils.FakeGenerator(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	addedGenerator, err := client.AddGenerator(generator, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Create another colony
	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colony2PrvKey)
	assert.Nil(t, err)

	// Try to remove generator from a different colony
	err = client.RemoveGenerator(addedGenerator.ID, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestPackGeneratorUnauthorized tests that non-members cannot pack generators
func TestPackGeneratorUnauthorized(t *testing.T) {
	env, client, s, serverPrvKey, done := server.SetupTestEnv2(t)

	// Add a generator
	generator := utils.FakeGenerator(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	addedGenerator, err := client.AddGenerator(generator, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Create another colony
	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colony2PrvKey)
	assert.Nil(t, err)

	// Try to pack generator from a different colony
	err = client.PackGenerator(addedGenerator.ID, "arg", executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestResolveGeneratorUnauthorized tests that non-members cannot resolve generators
func TestResolveGeneratorUnauthorized(t *testing.T) {
	env, client, s, serverPrvKey, done := server.SetupTestEnv2(t)

	// Add a generator
	generator := utils.FakeGenerator(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	generator.Name = "test-generator"
	_, err := client.AddGenerator(generator, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Create another colony
	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colony2PrvKey)
	assert.Nil(t, err)

	// Try to resolve generator from a different colony
	_, err = client.ResolveGenerator(env.ColonyName, "test-generator", executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestAddGeneratorInvalidWorkflow tests that invalid workflow spec is rejected
func TestAddGeneratorInvalidWorkflow(t *testing.T) {
	env, client, s, _, done := server.SetupTestEnv2(t)

	generator := utils.FakeGenerator(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	generator.WorkflowSpec = "invalid-json"
	_, err := client.AddGenerator(generator, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}