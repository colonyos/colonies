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