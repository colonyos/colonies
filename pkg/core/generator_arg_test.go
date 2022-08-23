package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateGeneratorArg(t *testing.T) {
	colonyID := GenerateRandomID()
	id := GenerateRandomID()
	generatorArg := CreateGeneratorArg(id, colonyID, "arg")
	assert.Equal(t, generatorArg.Arg, "arg")
	assert.Equal(t, generatorArg.GeneratorID, id)
	assert.Equal(t, generatorArg.ColonyID, colonyID)
	assert.Len(t, generatorArg.ID, 64)
}
