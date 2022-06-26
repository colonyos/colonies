package postgresql

import (
	"fmt"
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
	generator := fakeGenerator(t, colonyID)
	err = db.AddGenerator(generator)
	assert.Nil(t, err)

	generator = fakeGenerator(t, colonyID)
	err = db.AddGenerator(generator)
	assert.Nil(t, err)

	generatorsFromDB, err := db.FindGeneratorsByColonyID(colonyID, 100)
	assert.Nil(t, err)

	// TODO
	fmt.Println(generatorsFromDB)

	defer db.Close()
}
