package validator

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCheckIfColonyExists(t *testing.T) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)

	ownership := createOwnership(db)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	err = ownership.checkIfColonyExists(core.GenerateRandomID())
	assert.NotNil(t, err)

	err = ownership.checkIfColonyExists(colony.ID)
	assert.Nil(t, err)

	defer db.Close()
}

func TestCheckIfRuntimeIsValid(t *testing.T) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)

	ownership := createOwnership(db)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	runtime := utils.CreateTestRuntime(colony.ID)
	err = db.AddRuntime(runtime)
	assert.Nil(t, err)

	approvedRuntime := utils.CreateTestRuntime(colony.ID)
	err = db.AddRuntime(approvedRuntime)
	assert.Nil(t, err)
	err = db.ApproveRuntime(approvedRuntime)
	assert.Nil(t, err)

	err = ownership.checkIfRuntimeIsValid(runtime.ID, colony.ID, false)
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

	defer db.Close()
}
