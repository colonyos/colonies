package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddFileSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	file := utils.CreateTestFile(env.colony1Name)
	_, err := client.AddFile(file, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFile(file, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFile(file, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFile(file, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetFileByIDSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	file := utils.CreateTestFile(env.colony1Name)
	addedFile, err := client.AddFile(file, env.executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetFileByID(env.colony1Name, addedFile.ID, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByID(env.colony1Name, addedFile.ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByID(env.colony1Name, addedFile.ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByID(env.colony1Name, addedFile.ID, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetFileByNameSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	file := utils.CreateTestFile(env.colony1Name)
	addedFile, err := client.AddFile(file, env.executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetFileByName(env.colony1Name, addedFile.Label, addedFile.Name, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByName(env.colony1Name, addedFile.Label, addedFile.Name, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByName(env.colony1Name, addedFile.Label, addedFile.Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByName(env.colony1Name, addedFile.Label, addedFile.Name, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetFilenamesSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.GetFilenames(env.colony1Name, "/testprefix", env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFilenames(env.colony1Name, "/testprefix", env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFilenames(env.colony1Name, "/testprefix", env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFilenames(env.colony1Name, "/testprefix", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetFilePrefixesSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.GetFileLabels(env.colony1Name, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileLabels(env.colony1Name, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileLabels(env.colony1Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileLabels(env.colony1Name, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestDeleteFileByIDSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	file := utils.CreateTestFile(env.colony1Name)
	file.Label = "/testlabel"
	file.Name = "testfile2"
	file.Size = 1
	addedFile, err := client.AddFile(file, env.executor1PrvKey)
	assert.Nil(t, err)

	err = client.RemoveFileByID(env.colony1Name, addedFile.ID, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveFileByID(env.colony1Name, addedFile.ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveFileByID(env.colony1Name, addedFile.ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveFileByID(env.colony1Name, addedFile.ID, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestDeleteFileByNameSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	file := utils.CreateTestFile(env.colony1Name)
	file.Label = "/testlabel"
	file.Name = "testfile2"
	file.Size = 1
	addedFile, err := client.AddFile(file, env.executor1PrvKey)
	assert.Nil(t, err)

	err = client.RemoveFileByName(env.colony1Name, addedFile.Label, addedFile.Name, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveFileByName(env.colony1Name, addedFile.Label, addedFile.Name, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveFileByName(env.colony1Name, addedFile.Label, addedFile.Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveFileByName(env.colony1Name, addedFile.Label, addedFile.Name, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}
