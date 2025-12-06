package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateLocation(t *testing.T) {
	location := CreateLocation("test_id", "test_name", "test_colony", "test_description", 12.34, 56.78)
	assert.Equal(t, "test_id", location.ID)
	assert.Equal(t, "test_name", location.Name)
	assert.Equal(t, "test_colony", location.ColonyName)
	assert.Equal(t, "test_description", location.Description)
	assert.Equal(t, 12.34, location.Long)
	assert.Equal(t, 56.78, location.Lat)
}

func TestLocationToJSON(t *testing.T) {
	location := CreateLocation("test_id", "test_name", "test_colony", "test_description", 12.34, 56.78)
	jsonStr, err := location.ToJSON()
	assert.Nil(t, err)
	assert.Contains(t, jsonStr, "test_id")
	assert.Contains(t, jsonStr, "test_name")
	assert.Contains(t, jsonStr, "test_colony")
	assert.Contains(t, jsonStr, "test_description")
}

func TestConvertJSONToLocation(t *testing.T) {
	location := CreateLocation("test_id", "test_name", "test_colony", "test_description", 12.34, 56.78)
	jsonStr, err := location.ToJSON()
	assert.Nil(t, err)

	location2, err := ConvertJSONToLocation(jsonStr)
	assert.Nil(t, err)
	assert.True(t, location.Equals(location2))
}

func TestConvertJSONToLocationInvalid(t *testing.T) {
	_, err := ConvertJSONToLocation("invalid json")
	assert.NotNil(t, err)
}

func TestLocationEquals(t *testing.T) {
	location1 := CreateLocation("test_id", "test_name", "test_colony", "test_description", 12.34, 56.78)
	location2 := CreateLocation("test_id", "test_name", "test_colony", "test_description", 12.34, 56.78)
	assert.True(t, location1.Equals(location2))

	location3 := CreateLocation("different_id", "test_name", "test_colony", "test_description", 12.34, 56.78)
	assert.False(t, location1.Equals(location3))

	location4 := CreateLocation("test_id", "different_name", "test_colony", "test_description", 12.34, 56.78)
	assert.False(t, location1.Equals(location4))

	location5 := CreateLocation("test_id", "test_name", "different_colony", "test_description", 12.34, 56.78)
	assert.False(t, location1.Equals(location5))

	location6 := CreateLocation("test_id", "test_name", "test_colony", "different_description", 12.34, 56.78)
	assert.False(t, location1.Equals(location6))

	location7 := CreateLocation("test_id", "test_name", "test_colony", "test_description", 99.99, 56.78)
	assert.False(t, location1.Equals(location7))

	location8 := CreateLocation("test_id", "test_name", "test_colony", "test_description", 12.34, 99.99)
	assert.False(t, location1.Equals(location8))
}

func TestLocationEqualsNil(t *testing.T) {
	location := CreateLocation("test_id", "test_name", "test_colony", "test_description", 12.34, 56.78)
	assert.False(t, location.Equals(nil))
}

func TestConvertLocationArrayToJSON(t *testing.T) {
	location1 := CreateLocation("test_id1", "test_name1", "test_colony", "test_description1", 12.34, 56.78)
	location2 := CreateLocation("test_id2", "test_name2", "test_colony", "test_description2", 98.76, 54.32)
	locations := []*Location{location1, location2}

	jsonStr, err := ConvertLocationArrayToJSON(locations)
	assert.Nil(t, err)
	assert.Contains(t, jsonStr, "test_id1")
	assert.Contains(t, jsonStr, "test_id2")
}

func TestConvertJSONToLocationArray(t *testing.T) {
	location1 := CreateLocation("test_id1", "test_name1", "test_colony", "test_description1", 12.34, 56.78)
	location2 := CreateLocation("test_id2", "test_name2", "test_colony", "test_description2", 98.76, 54.32)
	locations := []*Location{location1, location2}

	jsonStr, err := ConvertLocationArrayToJSON(locations)
	assert.Nil(t, err)

	locationsFromJSON, err := ConvertJSONToLocationArray(jsonStr)
	assert.Nil(t, err)
	assert.Len(t, locationsFromJSON, 2)
	assert.True(t, IsLocationArraysEqual(locations, locationsFromJSON))
}

func TestConvertJSONToLocationArrayInvalid(t *testing.T) {
	_, err := ConvertJSONToLocationArray("invalid json")
	assert.NotNil(t, err)
}

func TestIsLocationArraysEqual(t *testing.T) {
	location1 := CreateLocation("test_id1", "test_name1", "test_colony", "test_description1", 12.34, 56.78)
	location2 := CreateLocation("test_id2", "test_name2", "test_colony", "test_description2", 98.76, 54.32)

	locations1 := []*Location{location1, location2}
	locations2 := []*Location{location2, location1}
	assert.True(t, IsLocationArraysEqual(locations1, locations2))

	location3 := CreateLocation("test_id3", "test_name3", "test_colony", "test_description3", 11.11, 22.22)
	locations3 := []*Location{location1, location3}
	assert.False(t, IsLocationArraysEqual(locations1, locations3))

	locations4 := []*Location{location1}
	assert.False(t, IsLocationArraysEqual(locations1, locations4))
}

func TestIsLocationArraysEqualEmpty(t *testing.T) {
	var locations1 []*Location
	var locations2 []*Location
	assert.True(t, IsLocationArraysEqual(locations1, locations2))
}
