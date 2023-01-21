package validator

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCheckIfColonyExistsMock(t *testing.T) {
	ownership := createOwnershipMock()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	ownership.addColony(colony.ID)

	err := ownership.checkIfColonyExists(core.GenerateRandomID())
	assert.NotNil(t, err)

	err = ownership.checkIfColonyExists(colony.ID)
	assert.Nil(t, err)
}

func TestCheckIfExecutorBelongsToColonyMock(t *testing.T) {
	ownership := createOwnershipMock()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	ownership.addColony(colony.ID)

	executor := utils.CreateTestExecutor(colony.ID)
	ownership.addExecutor(executor.ID, colony.ID)

	assert.Nil(t, ownership.checkIfExecutorBelongsToColony(executor.ID, colony.ID))
	assert.NotNil(t, ownership.checkIfExecutorBelongsToColony(core.GenerateRandomID(), colony.ID))
	assert.NotNil(t, ownership.checkIfExecutorBelongsToColony(executor.ID, core.GenerateRandomID()))
	assert.NotNil(t, ownership.checkIfExecutorBelongsToColony(core.GenerateRandomID(), core.GenerateRandomID()))
}

func TestCheckIfExecutorIsValidMock(t *testing.T) {
	ownership := createOwnershipMock()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	ownership.addColony(colony.ID)

	executor := utils.CreateTestExecutor(colony.ID)
	ownership.addExecutor(executor.ID, colony.ID)

	approvedExecutor := utils.CreateTestExecutor(colony.ID)
	ownership.addExecutor(approvedExecutor.ID, colony.ID)
	ownership.approveExecutor(approvedExecutor.ID, colony.ID)

	err := ownership.checkIfExecutorIsValid(executor.ID, colony.ID, false)
	assert.Nil(t, err)
	err = ownership.checkIfExecutorIsValid(executor.ID, colony.ID, true)
	assert.NotNil(t, err)
	err = ownership.checkIfExecutorIsValid(core.GenerateRandomID(), colony.ID, true)
	assert.NotNil(t, err)
	err = ownership.checkIfExecutorIsValid(executor.ID, core.GenerateRandomID(), true)
	assert.NotNil(t, err)
	err = ownership.checkIfExecutorIsValid(core.GenerateRandomID(), core.GenerateRandomID(), true)
	assert.NotNil(t, err)
	err = ownership.checkIfExecutorIsValid(approvedExecutor.ID, colony.ID, true)
	assert.Nil(t, err)
	err = ownership.checkIfExecutorIsValid(approvedExecutor.ID, colony.ID, false)
	assert.Nil(t, err)
}
