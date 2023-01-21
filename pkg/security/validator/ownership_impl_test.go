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

func TestCheckIfExecutorIsValid(t *testing.T) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)

	ownership := createOwnership(db)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	approvedExecutor := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(approvedExecutor)
	assert.Nil(t, err)
	err = db.ApproveExecutor(approvedExecutor)
	assert.Nil(t, err)

	err = ownership.checkIfExecutorIsValid(executor.ID, colony.ID, false)
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

	defer db.Close()
}
