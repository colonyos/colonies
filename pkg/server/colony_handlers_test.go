package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddColony(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	server.Shutdown()
	<-done
}

func TestDeleteColony(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	coloniesFromServer, err := client.GetColonies(serverPrvKey)
	assert.Nil(t, err)
	assert.Len(t, coloniesFromServer, 1)

	err = client.DeleteColony(addedColony.ID, serverPrvKey)
	assert.Nil(t, err)

	coloniesFromServer, err = client.GetColonies(serverPrvKey)
	assert.Nil(t, err)
	assert.Len(t, coloniesFromServer, 0)

	server.Shutdown()
	<-done
}

func TestGetColony(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	runtime, runtimePrvKey, err := utils.CreateTestRuntimeWithKey(colony.ID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(runtime, colonyPrvKey)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime.ID, colonyPrvKey)
	assert.Nil(t, err)

	colonyFromServer, err := client.GetColonyByID(colony.ID, runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(colonyFromServer))

	server.Shutdown()
	<-done
}

func TestGetColonies(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony1, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	colony2, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	var colonies []*core.Colony
	colonies = append(colonies, colony1)
	colonies = append(colonies, colony2)

	coloniesFromServer, err := client.GetColonies(serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, core.IsColonyArraysEqual(colonies, coloniesFromServer))

	server.Shutdown()
	<-done
}
