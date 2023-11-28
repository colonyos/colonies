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

	err = db.DeleteGeneratorArgByID("invalid_id")
	assert.NotNil(t, err)

	err = db.DeleteAllGeneratorArgsByGeneratorID("invalid_id")
	assert.NotNil(t, err)

	err = db.DeleteAllGeneratorArgsByColonyName("invalid_name")
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

func TestDeleteGeneratorArgByID(t *testing.T) {
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

	err = db.DeleteGeneratorArgByID(generatorArg.ID)
	assert.Nil(t, err)

	count, err = db.CountGeneratorArgs(generatorID)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)
}

func TestDeleteGeneratorArgByGeneratorID(t *testing.T) {
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

	err = db.DeleteAllGeneratorArgsByGeneratorID(generatorID1)
	assert.Nil(t, err)

	count, err := db.CountGeneratorArgs(generatorID1)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	count, err = db.CountGeneratorArgs(generatorID2)
	assert.Nil(t, err)
	assert.Equal(t, count, 1)
}

func TestDeleteGeneratorArgByColonyName(t *testing.T) {
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

	err = db.DeleteAllGeneratorArgsByColonyName(colonyName)
	assert.Nil(t, err)

	count, err := db.CountGeneratorArgs(generatorID1)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	count, err = db.CountGeneratorArgs(generatorID2)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)
}
