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

	file := utils.CreateTestFile(env.colony1ID)
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

	file := utils.CreateTestFile(env.colony1ID)
	addedFile, err := client.AddFile(file, env.executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetFileByID(env.colony1ID, addedFile.ID, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByID(env.colony1ID, addedFile.ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByID(env.colony1ID, addedFile.ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByID(env.colony1ID, addedFile.ID, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetFileByNameSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	file := utils.CreateTestFile(env.colony1ID)
	addedFile, err := client.AddFile(file, env.executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetFileByName(env.colony1ID, addedFile.Label, addedFile.Name, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByName(env.colony1ID, addedFile.Label, addedFile.Name, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByName(env.colony1ID, addedFile.Label, addedFile.Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByName(env.colony1ID, addedFile.Label, addedFile.Name, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetFilenamesSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.GetFilenames(env.colony1ID, "/testprefix", env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFilenames(env.colony1ID, "/testprefix", env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFilenames(env.colony1ID, "/testprefix", env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFilenames(env.colony1ID, "/testprefix", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetFilePrefixesSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.GetFileLabels(env.colony1ID, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileLabels(env.colony1ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileLabels(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileLabels(env.colony1ID, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestDeleteFileByIDSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	file := utils.CreateTestFile(env.colony1ID)
	file.Label = "/testlabel"
	file.Name = "testfile2"
	file.Size = 1
	addedFile, err := client.AddFile(file, env.executor1PrvKey)
	assert.Nil(t, err)

	err = client.DeleteFileByID(env.colony1ID, addedFile.ID, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteFileByID(env.colony1ID, addedFile.ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteFileByID(env.colony1ID, addedFile.ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteFileByID(env.colony1ID, addedFile.ID, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestDeleteFileByNameSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	file := utils.CreateTestFile(env.colony1ID)
	file.Label = "/testlabel"
	file.Name = "testfile2"
	file.Size = 1
	addedFile, err := client.AddFile(file, env.executor1PrvKey)
	assert.Nil(t, err)

	err = client.DeleteFileByName(env.colony1ID, addedFile.Label, addedFile.Name, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteFileByName(env.colony1ID, addedFile.Label, addedFile.Name, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteFileByName(env.colony1ID, addedFile.Label, addedFile.Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteFileByName(env.colony1ID, addedFile.Label, addedFile.Name, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}
