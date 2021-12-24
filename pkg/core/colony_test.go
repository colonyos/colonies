package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateColony(t *testing.T) {
	name := "test_colony_name"
	colony := CreateColony(GenerateRandomID(), name)

	assert.Equal(t, colony.Name, name)
	assert.Len(t, colony.ID, 64)
}

func TestColonyToJSON(t *testing.T) {
	name := "test_colony_name"
	colony := CreateColony(GenerateRandomID(), name)

	jsonString, err := colony.ToJSON()
	assert.Nil(t, err)

	colony2, err := ConvertJSONToColony(jsonString)
	assert.Nil(t, err)
	assert.True(t, colony2.Equals(colony))
}

func TestColonyToJSONArray(t *testing.T) {
	var colonies []*Colony

	colonies = append(colonies, CreateColony(GenerateRandomID(), "test_colony_name1"))
	colonies = append(colonies, CreateColony(GenerateRandomID(), "test_colony_name2"))

	jsonString, err := ConvertColonyArrayToJSON(colonies)
	assert.Nil(t, err)

	colonies2, err := ConvertJSONToColonyArray(jsonString)
	assert.Nil(t, err)
	assert.True(t, IsColonyArraysEqual(colonies, colonies2))
}
