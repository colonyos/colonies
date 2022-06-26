package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func fakeGenerator(t *testing.T, colonyID string) *core.Generator {
	workflowSpec := core.CreateWorkflowSpec(colonyID)
	processSpec1 := core.CreateEmptyProcessSpec()
	processSpec1.Name = "task1"
	processSpec2 := core.CreateEmptyProcessSpec()
	processSpec2.Name = "task2"
	processSpec2.AddDependency("task1")
	workflowSpec.AddProcessSpec(processSpec1)
	workflowSpec.AddProcessSpec(processSpec2)
	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)
	generator := core.CreateGenerator(colonyID, "test_genname", jsonStr, 10, 0, 10)
	return generator
}

func TestAddGenerator(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	generator := fakeGenerator(t, core.GenerateRandomID())
	err = db.AddGenerator(generator)
	assert.Nil(t, err)

	defer db.Close()
}

func TestGetGenerator(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	generator := fakeGenerator(t, core.GenerateRandomID())
	err = db.AddGenerator(generator)
	assert.Nil(t, err)

	generatorFromDB, err := db.GetGeneratorByID(generator.ID)
	assert.Nil(t, err)
	assert.True(t, generator.Equals(generatorFromDB))

	defer db.Close()
}

func TestFindGeneratorByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()
	generator1 := fakeGenerator(t, colonyID)
	err = db.AddGenerator(generator1)
	assert.Nil(t, err)

	generator2 := fakeGenerator(t, colonyID)
	err = db.AddGenerator(generator2)
	assert.Nil(t, err)

	generatorsFromDB, err := db.FindGeneratorsByColonyID(colonyID, 100)
	assert.Nil(t, err)
	assert.Len(t, generatorsFromDB, 2)

	count := 0
	for _, generator := range generatorsFromDB {
		if generator.ID == generator1.ID {
			count++
		}
		if generator.ID == generator2.ID {
			count++
		}
	}
	assert.True(t, count == 2)

	defer db.Close()
}

func TestDeleteGeneratorByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()
	generator1 := fakeGenerator(t, colonyID)
	err = db.AddGenerator(generator1)
	assert.Nil(t, err)

	generator2 := fakeGenerator(t, colonyID)
	err = db.AddGenerator(generator2)
	assert.Nil(t, err)

	generatorFromDB, err := db.GetGeneratorByID(generator1.ID)
	assert.Nil(t, err)
	assert.NotNil(t, generatorFromDB)

	err = db.DeleteGeneratorByID(generator1.ID)
	assert.Nil(t, err)

	generatorFromDB, err = db.GetGeneratorByID(generator1.ID)
	assert.Nil(t, err)
	assert.Nil(t, generatorFromDB)

	generatorFromDB, err = db.GetGeneratorByID(generator2.ID)
	assert.Nil(t, err)
	assert.NotNil(t, generatorFromDB)

	defer db.Close()
}

func TestDeleteAllGeneratorsByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID1 := core.GenerateRandomID()
	generator1 := fakeGenerator(t, colonyID1)
	err = db.AddGenerator(generator1)
	assert.Nil(t, err)

	generator2 := fakeGenerator(t, colonyID1)
	err = db.AddGenerator(generator2)
	assert.Nil(t, err)

	colonyID2 := core.GenerateRandomID()
	generator3 := fakeGenerator(t, colonyID2)
	err = db.AddGenerator(generator3)
	assert.Nil(t, err)

	generatorFromDB, err := db.GetGeneratorByID(generator1.ID)
	assert.Nil(t, err)
	assert.NotNil(t, generatorFromDB)

	err = db.DeleteAllGeneratorsByColonyID(colonyID1)
	assert.Nil(t, err)

	generatorFromDB, err = db.GetGeneratorByID(generator1.ID)
	assert.Nil(t, err)
	assert.Nil(t, generatorFromDB)

	generatorFromDB, err = db.GetGeneratorByID(generator2.ID)
	assert.Nil(t, err)
	assert.Nil(t, generatorFromDB)

	generatorFromDB, err = db.GetGeneratorByID(generator3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, generatorFromDB)

	defer db.Close()
}

func TestGeneratorLifeCycle(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID1 := core.GenerateRandomID()
	generator1 := fakeGenerator(t, colonyID1)
	err = db.AddGenerator(generator1)
	assert.Nil(t, err)

	generatorFromDB, err := db.GetGeneratorByID(generator1.ID)
	assert.Nil(t, err)
	assert.True(t, generatorFromDB.Counter == 0)

	err = db.IncreaseGeneratorCounter(generator1.ID)
	assert.Nil(t, err)

	generatorFromDB, err = db.GetGeneratorByID(generator1.ID)
	assert.Nil(t, err)
	assert.True(t, generatorFromDB.Counter == 1)

	err = db.ResetGenerator(generator1.ID)
	assert.Nil(t, err)

	generatorFromDB, err = db.GetGeneratorByID(generator1.ID)
	assert.Nil(t, err)
	assert.True(t, generatorFromDB.Counter == 0)

	defer db.Close()
}
