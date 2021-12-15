package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateColony(t *testing.T) {
	name := "test_colony_name"
	colony := CreateColony(GenerateRandomID(), name)

	assert.Equal(t, colony.Name(), name)
	assert.Len(t, colony.ID(), 64)
}

func TestColonyToJSON(t *testing.T) {
	name := "test_colony_name"
	colony := CreateColony(GenerateRandomID(), name)

	jsonString, err := colony.ToJSON()
	assert.Nil(t, err)

	colony2, err := CreateColonyFromJSON(jsonString)
	assert.Nil(t, err)

	assert.Equal(t, colony.Name(), colony2.Name())
	assert.Equal(t, colony.ID(), colony2.ID())
}
