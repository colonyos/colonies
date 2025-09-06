package generator_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/service"
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