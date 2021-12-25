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

func TestColonyEquals(t *testing.T) {
	name := "test_colony_name"
	id := GenerateRandomID()
	colony1 := CreateColony(id, name)
	assert.True(t, colony1.Equals(colony1))

	colony2 := CreateColony(id+"X", name)
	assert.False(t, colony2.Equals(colony1))
	colony2 = CreateColony(id, name+"X")
	assert.False(t, colony2.Equals(colony1))
}

func TestIsColonyArraysEqual(t *testing.T) {
	colony1 := CreateColony(GenerateRandomID(), "test_colony_name_1")
	colony2 := CreateColony(GenerateRandomID(), "test_colony_name_1")
	colony3 := CreateColony(GenerateRandomID(), "test_colony_name_1")
	colony4 := CreateColony(GenerateRandomID(), "test_colony_name_1")

	var colonies1 []*Colony
	colonies1 = append(colonies1, colony1)
	colonies1 = append(colonies1, colony2)
	colonies1 = append(colonies1, colony3)

	var colonies2 []*Colony
	colonies2 = append(colonies2, colony2)
	colonies2 = append(colonies2, colony3)
	colonies2 = append(colonies2, colony1)

	var colonies3 []*Colony
	colonies3 = append(colonies3, colony2)
	colonies3 = append(colonies3, colony3)
	colonies3 = append(colonies3, colony4)

	var colonies4 []*Colony

	assert.True(t, IsColonyArraysEqual(colonies1, colonies1))
	assert.True(t, IsColonyArraysEqual(colonies1, colonies2))
	assert.False(t, IsColonyArraysEqual(colonies1, colonies3))
	assert.False(t, IsColonyArraysEqual(colonies1, colonies4))
	assert.True(t, IsColonyArraysEqual(colonies4, colonies4))
	assert.True(t, IsColonyArraysEqual(nil, nil))
	assert.False(t, IsColonyArraysEqual(nil, colonies2))
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
