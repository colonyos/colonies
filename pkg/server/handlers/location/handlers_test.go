package location_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddLocation(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	location := utils.CreateTestLocation(colony.Name, "test_location")
	addedLocation, err := client.AddLocation(location, colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedLocation)
	assert.Equal(t, "test_location", addedLocation.Name)
	assert.Equal(t, colony.Name, addedLocation.ColonyName)

	s.Shutdown()
	<-done
}

func TestAddLocationDuplicate(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	location := utils.CreateTestLocation(colony.Name, "test_location")
	_, err = client.AddLocation(location, colonyPrvKey)
	assert.Nil(t, err)

	// Try to add same location again - should fail
	location2 := utils.CreateTestLocation(colony.Name, "test_location")
	_, err = client.AddLocation(location2, colonyPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

func TestAddLocationNotColonyOwner(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	// Try to add location with executor key - should fail
	location := utils.CreateTestLocation(colony.Name, "test_location")
	_, err = client.AddLocation(location, executorPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

func TestGetLocations(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	location1 := utils.CreateTestLocation(colony.Name, "test_location1")
	_, err = client.AddLocation(location1, colonyPrvKey)
	assert.Nil(t, err)

	location2 := utils.CreateTestLocation(colony.Name, "test_location2")
	_, err = client.AddLocation(location2, colonyPrvKey)
	assert.Nil(t, err)

	locationsFromServer, err := client.GetLocations(colony.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, locationsFromServer, 2)

	s.Shutdown()
	<-done
}

func TestGetLocationsEmpty(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	locationsFromServer, err := client.GetLocations(colony.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, locationsFromServer, 0)

	s.Shutdown()
	<-done
}

func TestGetLocation(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	location := core.CreateLocation(core.GenerateRandomID(), "test_location", colony.Name, "my description", 12.34, 56.78)
	_, err = client.AddLocation(location, colonyPrvKey)
	assert.Nil(t, err)

	locationFromServer, err := client.GetLocation(colony.Name, "test_location", executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, "test_location", locationFromServer.Name)
	assert.Equal(t, "my description", locationFromServer.Description)
	assert.Equal(t, 12.34, locationFromServer.Long)
	assert.Equal(t, 56.78, locationFromServer.Lat)

	s.Shutdown()
	<-done
}

func TestGetLocationNotFound(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	_, err = client.GetLocation(colony.Name, "non_existent_location", executorPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

func TestRemoveLocation(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	location := utils.CreateTestLocation(colony.Name, "test_location")
	_, err = client.AddLocation(location, colonyPrvKey)
	assert.Nil(t, err)

	locationsFromServer, err := client.GetLocations(colony.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, locationsFromServer, 1)

	err = client.RemoveLocation(colony.Name, "test_location", colonyPrvKey)
	assert.Nil(t, err)

	locationsFromServer, err = client.GetLocations(colony.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, locationsFromServer, 0)

	s.Shutdown()
	<-done
}

func TestRemoveLocationNotColonyOwner(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	location := utils.CreateTestLocation(colony.Name, "test_location")
	_, err = client.AddLocation(location, colonyPrvKey)
	assert.Nil(t, err)

	// Try to remove location with executor key - should fail
	err = client.RemoveLocation(colony.Name, "test_location", executorPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

func TestRemoveLocationNotFound(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	err = client.RemoveLocation(colony.Name, "non_existent_location", colonyPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

func TestGetLocationsNotMember(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor.Name, colonyPrvKey2)
	assert.Nil(t, err)

	location := utils.CreateTestLocation(colony1.Name, "test_location")
	_, err = client.AddLocation(location, colonyPrvKey1)
	assert.Nil(t, err)

	// Try to get locations from colony1 with executor from colony2 - should fail
	_, err = client.GetLocations(colony1.Name, executorPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}
