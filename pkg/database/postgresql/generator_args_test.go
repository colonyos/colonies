package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestGeneratorArg(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()
	generatorID := core.GenerateRandomID()
	generatorArg := core.CreateGeneratorArg(generatorID, colonyID, "arg")
	generatorArg2 := core.CreateGeneratorArg(generatorID, colonyID, "arg")

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

	defer db.Close()
}

func TestDeleteGeneratorArgByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()
	generatorID := core.GenerateRandomID()
	generatorArg := core.CreateGeneratorArg(generatorID, colonyID, "arg")

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

	defer db.Close()
}

func TestDeleteGeneratorArgByGeneratorID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()
	generatorID1 := core.GenerateRandomID()
	generatorArg := core.CreateGeneratorArg(generatorID1, colonyID, "arg")
	generatorID2 := core.GenerateRandomID()
	generatorArg2 := core.CreateGeneratorArg(generatorID2, colonyID, "arg")

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

	defer db.Close()
}

func TestDeleteGeneratorArgByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()
	generatorID1 := core.GenerateRandomID()
	generatorArg := core.CreateGeneratorArg(generatorID1, colonyID, "arg")
	generatorID2 := core.GenerateRandomID()
	generatorArg2 := core.CreateGeneratorArg(generatorID2, colonyID, "arg")

	err = db.AddGeneratorArg(generatorArg)
	assert.Nil(t, err)
	err = db.AddGeneratorArg(generatorArg2)
	assert.Nil(t, err)

	err = db.DeleteAllGeneratorArgsByColonyID(colonyID)
	assert.Nil(t, err)

	count, err := db.CountGeneratorArgs(generatorID1)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	count, err = db.CountGeneratorArgs(generatorID2)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	defer db.Close()
}
