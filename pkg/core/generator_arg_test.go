package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateGeneratorArg(t *testing.T) {
	colonyName := GenerateRandomID()
	id := GenerateRandomID()
	generatorArg := CreateGeneratorArg(id, colonyName, "arg")
	assert.Equal(t, generatorArg.Arg, "arg")
	assert.Equal(t, generatorArg.GeneratorID, id)
	assert.Equal(t, generatorArg.ColonyName, colonyName)
	assert.Len(t, generatorArg.ID, 64)
}
