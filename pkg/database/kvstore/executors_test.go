package kvstore

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestExecutorClosedDB(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)

	db.Close()

	executor := utils.CreateTestExecutor(core.GenerateRandomID())
	
	// KVStore operations work even after close (in-memory store)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	_, err = db.GetExecutors()
	assert.Nil(t, err)

	_, err = db.GetExecutorByID("invalid_id")
	assert.NotNil(t, err) // Expected error for non-existing

	_, err = db.GetExecutorsByColonyName("invalid_colony_name")
	assert.Nil(t, err) // Returns empty slice

	// The executor we added should be retrievable
	_, err = db.GetExecutorByName(executor.ColonyName, executor.Name)
	assert.Nil(t, err)

	err = db.ApproveExecutor(executor)
	assert.Nil(t, err)

	err = db.RejectExecutor(executor)
	assert.Nil(t, err)

	err = db.MarkAlive(executor)
	assert.Nil(t, err)

	err = db.RemoveExecutorByName("invalid_colony_name", "invalid_id")
	assert.NotNil(t, err) // Expected error for non-existing

	err = db.RemoveExecutorsByColonyName("invalid_colony_name")
	assert.Nil(t, err) // No error when nothing to remove

	_, err = db.CountExecutors()
	assert.Nil(t, err)

	_, err = db.CountExecutorsByColonyName("invalid_colony_name")
	assert.Nil(t, err)
}

func TestAddExecutor(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	executor.Capabilities.Software.Name = "sw_name"
	executor.Capabilities.Software.Type = "sw_type"
	executor.Capabilities.Software.Version = "sw_version"

	executor.Capabilities.Hardware.Model = "model"
	executor.Capabilities.Hardware.Nodes = 10
	executor.Capabilities.Hardware.CPU = "1000m"
	executor.Capabilities.Hardware.Memory = "10G"
	executor.Capabilities.Hardware.Storage = "1000G"
	executor.Capabilities.Hardware.GPU.Name = "nvidia_2080ti"
	executor.Capabilities.Hardware.GPU.Count = 4000
	executor.Capabilities.Hardware.GPU.NodeCount = 4
	executor.Capabilities.Hardware.GPU.Memory = "10G"

	// Test adding nil executor
	err = db.AddExecutor(nil)
	assert.NotNil(t, err)

	// Test adding valid executor
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	executors, err := db.GetExecutors()
	assert.Nil(t, err)
	assert.Len(t, executors, 1)

	executorFromDB := executors[0]
	assert.True(t, executor.Equals(executorFromDB))
	assert.True(t, executorFromDB.IsPending())
	assert.False(t, executorFromDB.IsApproved())
	assert.False(t, executorFromDB.IsRejected())

	assert.Equal(t, executor.Capabilities.Software.Name, "sw_name")
	assert.Equal(t, executor.Capabilities.Software.Type, "sw_type")
	assert.Equal(t, executor.Capabilities.Software.Version, "sw_version")

	assert.Equal(t, executor.Capabilities.Hardware.Model, "model")
	assert.Equal(t, executor.Capabilities.Hardware.Nodes, 10)
	assert.Equal(t, executor.Capabilities.Hardware.CPU, "1000m")
	assert.Equal(t, executor.Capabilities.Hardware.Memory, "10G")
	assert.Equal(t, executor.Capabilities.Hardware.Storage, "1000G")
	assert.Equal(t, executor.Capabilities.Hardware.GPU.Name, "nvidia_2080ti")
	assert.Equal(t, executor.Capabilities.Hardware.GPU.Count, 4000)
	assert.Equal(t, executor.Capabilities.Hardware.GPU.NodeCount, 4)
	assert.Equal(t, executor.Capabilities.Hardware.GPU.Memory, "10G")

	// Test duplicate executor
	err = db.AddExecutor(executor)
	assert.NotNil(t, err)
}

func TestGetExecutorByID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	executorFromDB, err := db.GetExecutorByID(executor.ID)
	assert.Nil(t, err)
	assert.True(t, executor.Equals(executorFromDB))

	// Test non-existing executor
	_, err = db.GetExecutorByID("non_existing_id")
	assert.NotNil(t, err)
}

func TestGetExecutorByName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	executorFromDB, err := db.GetExecutorByName(colony.Name, executor.Name)
	assert.Nil(t, err)
	assert.True(t, executor.Equals(executorFromDB))

	// Test non-existing executor
	_, err = db.GetExecutorByName(colony.Name, "non_existing_name")
	assert.NotNil(t, err)

	// Test invalid colony
	_, err = db.GetExecutorByName("invalid_colony", executor.Name)
	assert.NotNil(t, err)
}

func TestGetExecutorsByColonyName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	executors, err := db.GetExecutorsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Len(t, executors, 2)

	// Test invalid colony
	executors, err = db.GetExecutorsByColonyName("invalid_colony")
	assert.Nil(t, err)
	assert.Empty(t, executors)
}

func TestApproveExecutor(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	// Executor should be pending initially
	executorFromDB, err := db.GetExecutorByID(executor.ID)
	assert.Nil(t, err)
	assert.True(t, executorFromDB.IsPending())

	// Approve executor
	err = db.ApproveExecutor(executor)
	assert.Nil(t, err)

	executorFromDB, err = db.GetExecutorByID(executor.ID)
	assert.Nil(t, err)
	assert.True(t, executorFromDB.IsApproved())
	assert.False(t, executorFromDB.IsPending())
	assert.False(t, executorFromDB.IsRejected())

	// Test approve nil executor
	err = db.ApproveExecutor(nil)
	assert.NotNil(t, err)
}

func TestRejectExecutor(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	// Reject executor
	err = db.RejectExecutor(executor)
	assert.Nil(t, err)

	executorFromDB, err := db.GetExecutorByID(executor.ID)
	assert.Nil(t, err)
	assert.True(t, executorFromDB.IsRejected())
	assert.False(t, executorFromDB.IsPending())
	assert.False(t, executorFromDB.IsApproved())

	// Test reject nil executor
	err = db.RejectExecutor(nil)
	assert.NotNil(t, err)
}

func TestMarkAlive(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	// Mark alive
	err = db.MarkAlive(executor)
	assert.Nil(t, err)

	executorFromDB, err := db.GetExecutorByID(executor.ID)
	assert.Nil(t, err)
	// LastHeardFromTime should be updated
	assert.True(t, executorFromDB.LastHeardFromTime.After(executor.LastHeardFromTime))

	// Test mark alive nil executor
	err = db.MarkAlive(nil)
	assert.NotNil(t, err)
}

func TestRemoveExecutor(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	// Verify executor exists
	count, err := db.CountExecutorsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, count, 1)

	// Remove executor
	err = db.RemoveExecutorByName(colony.Name, executor.Name)
	assert.Nil(t, err)

	// Verify executor is gone
	count, err = db.CountExecutorsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	_, err = db.GetExecutorByID(executor.ID)
	assert.NotNil(t, err)

	// Test remove non-existing executor
	err = db.RemoveExecutorByName(colony.Name, "non_existing")
	assert.NotNil(t, err)

	// Test remove from invalid colony
	err = db.RemoveExecutorByName("invalid_colony", executor.Name)
	assert.NotNil(t, err)
}

func TestRemoveExecutorsByColonyName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Add multiple executors
	executor1 := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	// Verify executors exist
	count, err := db.CountExecutorsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, count, 2)

	// Remove all executors
	err = db.RemoveExecutorsByColonyName(colony.Name)
	assert.Nil(t, err)

	// Verify all executors are gone
	count, err = db.CountExecutorsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	executors, err := db.GetExecutorsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Empty(t, executors)

	// Test remove from invalid colony - should not error
	err = db.RemoveExecutorsByColonyName("invalid_colony")
	assert.Nil(t, err)
}

func TestCountExecutors(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "colony1")
	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "colony2")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	// Add executors to different colonies
	executor1 := utils.CreateTestExecutor(colony1.Name)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony1.Name)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	executor3 := utils.CreateTestExecutor(colony2.Name)
	err = db.AddExecutor(executor3)
	assert.Nil(t, err)

	// Test total count
	totalCount, err := db.CountExecutors()
	assert.Nil(t, err)
	assert.Equal(t, totalCount, 3)

	// Test colony-specific counts
	count1, err := db.CountExecutorsByColonyName(colony1.Name)
	assert.Nil(t, err)
	assert.Equal(t, count1, 2)

	count2, err := db.CountExecutorsByColonyName(colony2.Name)
	assert.Nil(t, err)
	assert.Equal(t, count2, 1)

	// Test invalid colony
	invalidCount, err := db.CountExecutorsByColonyName("invalid_colony")
	assert.Nil(t, err)
	assert.Equal(t, invalidCount, 0)
}