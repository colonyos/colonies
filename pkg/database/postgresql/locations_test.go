package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddLocation(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	location := utils.CreateTestLocation(colony.Name, "test_location")
	err = db.AddLocation(location)
	assert.Nil(t, err)

	locationFromDB, err := db.GetLocationByID(location.ID)
	assert.Nil(t, err)
	assert.True(t, location.Equals(locationFromDB))
}

func TestAddLocationNil(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	err = db.AddLocation(nil)
	assert.NotNil(t, err)
}

func TestAddLocationDuplicate(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	location := utils.CreateTestLocation(colony.Name, "test_location")
	err = db.AddLocation(location)
	assert.Nil(t, err)

	// Try to add a location with the same name - should fail
	location2 := utils.CreateTestLocation(colony.Name, "test_location")
	err = db.AddLocation(location2)
	assert.NotNil(t, err)
}

func TestGetLocationByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	location := utils.CreateTestLocation(colony.Name, "test_location")
	err = db.AddLocation(location)
	assert.Nil(t, err)

	locationFromDB, err := db.GetLocationByID(location.ID)
	assert.Nil(t, err)
	assert.True(t, location.Equals(locationFromDB))
}

func TestGetLocationByIDNotFound(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	locationFromDB, err := db.GetLocationByID("non_existent_id")
	assert.Nil(t, err)
	assert.Nil(t, locationFromDB)
}

func TestGetLocationByName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	location := utils.CreateTestLocation(colony.Name, "test_location")
	err = db.AddLocation(location)
	assert.Nil(t, err)

	locationFromDB, err := db.GetLocationByName(colony.Name, "test_location")
	assert.Nil(t, err)
	assert.True(t, location.Equals(locationFromDB))
}

func TestGetLocationByNameNotFound(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	locationFromDB, err := db.GetLocationByName(colony.Name, "non_existent_name")
	assert.Nil(t, err)
	assert.Nil(t, locationFromDB)
}

func TestGetLocationsByColonyName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	location1 := utils.CreateTestLocation(colony.Name, "test_location1")
	err = db.AddLocation(location1)
	assert.Nil(t, err)

	location2 := utils.CreateTestLocation(colony.Name, "test_location2")
	err = db.AddLocation(location2)
	assert.Nil(t, err)

	locationsFromDB, err := db.GetLocationsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Len(t, locationsFromDB, 2)
}

func TestGetLocationsByColonyNameEmpty(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	locationsFromDB, err := db.GetLocationsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Len(t, locationsFromDB, 0)
}

func TestRemoveLocationByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	location := utils.CreateTestLocation(colony.Name, "test_location")
	err = db.AddLocation(location)
	assert.Nil(t, err)

	err = db.RemoveLocationByID(location.ID)
	assert.Nil(t, err)

	locationFromDB, err := db.GetLocationByID(location.ID)
	assert.Nil(t, err)
	assert.Nil(t, locationFromDB)
}

func TestRemoveLocationByName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	location := utils.CreateTestLocation(colony.Name, "test_location")
	err = db.AddLocation(location)
	assert.Nil(t, err)

	err = db.RemoveLocationByName(colony.Name, "test_location")
	assert.Nil(t, err)

	locationFromDB, err := db.GetLocationByName(colony.Name, "test_location")
	assert.Nil(t, err)
	assert.Nil(t, locationFromDB)
}

func TestRemoveLocationsByColonyName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	location1 := utils.CreateTestLocation(colony.Name, "test_location1")
	err = db.AddLocation(location1)
	assert.Nil(t, err)

	location2 := utils.CreateTestLocation(colony.Name, "test_location2")
	err = db.AddLocation(location2)
	assert.Nil(t, err)

	locationsFromDB, err := db.GetLocationsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Len(t, locationsFromDB, 2)

	err = db.RemoveLocationsByColonyName(colony.Name)
	assert.Nil(t, err)

	locationsFromDB, err = db.GetLocationsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Len(t, locationsFromDB, 0)
}

func TestLocationCoordinates(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	location := core.CreateLocation(core.GenerateRandomID(), "test_location", colony.Name, "test_desc", -122.4194, 37.7749)
	err = db.AddLocation(location)
	assert.Nil(t, err)

	locationFromDB, err := db.GetLocationByID(location.ID)
	assert.Nil(t, err)
	assert.Equal(t, -122.4194, locationFromDB.Long)
	assert.Equal(t, 37.7749, locationFromDB.Lat)
}

func TestLocationsDeletedWhenColonyDeleted(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Add multiple locations to the colony
	location1 := utils.CreateTestLocation(colony.Name, "test_location1")
	err = db.AddLocation(location1)
	assert.Nil(t, err)

	location2 := utils.CreateTestLocation(colony.Name, "test_location2")
	err = db.AddLocation(location2)
	assert.Nil(t, err)

	location3 := utils.CreateTestLocation(colony.Name, "test_location3")
	err = db.AddLocation(location3)
	assert.Nil(t, err)

	// Verify locations exist
	locationsFromDB, err := db.GetLocationsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Len(t, locationsFromDB, 3)

	// Delete the colony
	err = db.RemoveColonyByName(colony.Name)
	assert.Nil(t, err)

	// Verify all locations are deleted
	locationsFromDB, err = db.GetLocationsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Len(t, locationsFromDB, 0)

	// Verify individual location lookups also return nil
	locationFromDB, err := db.GetLocationByID(location1.ID)
	assert.Nil(t, err)
	assert.Nil(t, locationFromDB)

	locationFromDB, err = db.GetLocationByID(location2.ID)
	assert.Nil(t, err)
	assert.Nil(t, locationFromDB)

	locationFromDB, err = db.GetLocationByID(location3.ID)
	assert.Nil(t, err)
	assert.Nil(t, locationFromDB)
}
