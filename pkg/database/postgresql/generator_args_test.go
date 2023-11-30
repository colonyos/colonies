package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestGeneratorArgClosedDB(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	db.Close()

	generatorArg := core.CreateGeneratorArg("invalid_id", "invalid_id", "invalid_arh")
	err = db.AddGeneratorArg(generatorArg)
	assert.NotNil(t, err)

	_, err = db.GetGeneratorArgs("invalid_id", 1)
	assert.NotNil(t, err)

	_, err = db.CountGeneratorArgs("invalid_id")
	assert.NotNil(t, err)

	err = db.RemoveGeneratorArgByID("invalid_id")
	assert.NotNil(t, err)

	err = db.RemoveAllGeneratorArgsByGeneratorID("invalid_id")
	assert.NotNil(t, err)

	err = db.RemoveAllGeneratorArgsByColonyName("invalid_name")
	assert.NotNil(t, err)
}

func TestGeneratorArg(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	generatorID := core.GenerateRandomID()
	generatorArg := core.CreateGeneratorArg(generatorID, colonyName, "arg")
	generatorArg2 := core.CreateGeneratorArg(generatorID, colonyName, "arg")

	err = db.AddGeneratorArg(generatorArg)
	assert.Nil(t, err)
	err = db.AddGeneratorArg(generatorArg2)
	assert.Nil(t, err)

	generatorsArgFromDB, err := db.GetGeneratorArgs(generatorID, 100)
	assert.Nil(t, err)
	assert.Len(t, generatorsArgFromDB, 2)

	count, err := db.CountGeneratorArgs(generatorID)
	assert.Nil(t, err)
	assert.Equal(t, count, 2)
}

func TestRemoveGeneratorArgByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	generatorID := core.GenerateRandomID()
	generatorArg := core.CreateGeneratorArg(generatorID, colonyName, "arg")

	err = db.AddGeneratorArg(generatorArg)
	assert.Nil(t, err)

	count, err := db.CountGeneratorArgs(generatorID)
	assert.Nil(t, err)
	assert.Equal(t, count, 1)

	err = db.RemoveGeneratorArgByID(generatorArg.ID)
	assert.Nil(t, err)

	count, err = db.CountGeneratorArgs(generatorID)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)
}

func TestRemoveGeneratorArgByGeneratorID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	generatorID1 := core.GenerateRandomID()
	generatorArg := core.CreateGeneratorArg(generatorID1, colonyName, "arg")
	generatorID2 := core.GenerateRandomID()
	generatorArg2 := core.CreateGeneratorArg(generatorID2, colonyName, "arg")

	err = db.AddGeneratorArg(generatorArg)
	assert.Nil(t, err)
	err = db.AddGeneratorArg(generatorArg2)
	assert.Nil(t, err)

	err = db.RemoveAllGeneratorArgsByGeneratorID(generatorID1)
	assert.Nil(t, err)

	count, err := db.CountGeneratorArgs(generatorID1)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	count, err = db.CountGeneratorArgs(generatorID2)
	assert.Nil(t, err)
	assert.Equal(t, count, 1)
}

func TestRemoveGeneratorArgByColonyName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	generatorID1 := core.GenerateRandomID()
	generatorArg := core.CreateGeneratorArg(generatorID1, colonyName, "arg")
	generatorID2 := core.GenerateRandomID()
	generatorArg2 := core.CreateGeneratorArg(generatorID2, colonyName, "arg")

	err = db.AddGeneratorArg(generatorArg)
	assert.Nil(t, err)
	err = db.AddGeneratorArg(generatorArg2)
	assert.Nil(t, err)

	err = db.RemoveAllGeneratorArgsByColonyName(colonyName)
	assert.Nil(t, err)

	count, err := db.CountGeneratorArgs(generatorID1)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	count, err = db.CountGeneratorArgs(generatorID2)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)
}
