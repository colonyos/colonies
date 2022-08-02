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

func TestCheckIfRuntimeBelongsToColonyMock(t *testing.T) {
	ownership := createOwnershipMock()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	ownership.addColony(colony.ID)

	runtime := utils.CreateTestRuntime(colony.ID)
	ownership.addRuntime(runtime.ID, colony.ID)

	assert.Nil(t, ownership.checkIfRuntimeBelongsToColony(runtime.ID, colony.ID))
	assert.NotNil(t, ownership.checkIfRuntimeBelongsToColony(core.GenerateRandomID(), colony.ID))
	assert.NotNil(t, ownership.checkIfRuntimeBelongsToColony(runtime.ID, core.GenerateRandomID()))
	assert.NotNil(t, ownership.checkIfRuntimeBelongsToColony(core.GenerateRandomID(), core.GenerateRandomID()))
}

func TestCheckIfRuntimeIsValidMock(t *testing.T) {
	ownership := createOwnershipMock()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	ownership.addColony(colony.ID)

	runtime := utils.CreateTestRuntime(colony.ID)
	ownership.addRuntime(runtime.ID, colony.ID)

	approvedRuntime := utils.CreateTestRuntime(colony.ID)
	ownership.addRuntime(approvedRuntime.ID, colony.ID)
	ownership.approveRuntime(approvedRuntime.ID, colony.ID)

	err := ownership.checkIfRuntimeIsValid(runtime.ID, colony.ID, false)
	assert.Nil(t, err)
	err = ownership.checkIfRuntimeIsValid(runtime.ID, colony.ID, true)
	assert.NotNil(t, err)
	err = ownership.checkIfRuntimeIsValid(core.GenerateRandomID(), colony.ID, true)
	assert.NotNil(t, err)
	err = ownership.checkIfRuntimeIsValid(runtime.ID, core.GenerateRandomID(), true)
	assert.NotNil(t, err)
	err = ownership.checkIfRuntimeIsValid(core.GenerateRandomID(), core.GenerateRandomID(), true)
	assert.NotNil(t, err)
	err = ownership.checkIfRuntimeIsValid(approvedRuntime.ID, colony.ID, true)
	assert.Nil(t, err)
	err = ownership.checkIfRuntimeIsValid(approvedRuntime.ID, colony.ID, false)
	assert.Nil(t, err)
}
