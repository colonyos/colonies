package file_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddFileSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	file := utils.CreateTestFile(env.Colony1Name)
	_, err := client.AddFile(file, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFile(file, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFile(file, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFile(file, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetFileByIDSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	file := utils.CreateTestFile(env.Colony1Name)
	addedFile, err := client.AddFile(file, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetFileByID(env.Colony1Name, addedFile.ID, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByID(env.Colony1Name, addedFile.ID, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByID(env.Colony1Name, addedFile.ID, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByID(env.Colony1Name, addedFile.ID, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetFileByNameSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	file := utils.CreateTestFile(env.Colony1Name)
	addedFile, err := client.AddFile(file, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetFileByName(env.Colony1Name, addedFile.Label, addedFile.Name, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByName(env.Colony1Name, addedFile.Label, addedFile.Name, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByName(env.Colony1Name, addedFile.Label, addedFile.Name, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileByName(env.Colony1Name, addedFile.Label, addedFile.Name, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetFilenamesSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.GetFileData(env.Colony1Name, "/testprefix", env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileData(env.Colony1Name, "/testprefix", env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileData(env.Colony1Name, "/testprefix", env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileData(env.Colony1Name, "/testprefix", env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetFilePrefixesSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.GetFileLabels(env.Colony1Name, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileLabels(env.Colony1Name, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileLabels(env.Colony1Name, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFileLabels(env.Colony1Name, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestRemoveFileByIDSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	file := utils.CreateTestFile(env.Colony1Name)
	file.Label = "/testlabel"
	file.Name = "testfile2"
	file.Size = 1
	addedFile, err := client.AddFile(file, env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.RemoveFileByID(env.Colony1Name, addedFile.ID, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveFileByID(env.Colony1Name, addedFile.ID, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveFileByID(env.Colony1Name, addedFile.ID, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveFileByID(env.Colony1Name, addedFile.ID, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestRemoveFileByNameSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	file := utils.CreateTestFile(env.Colony1Name)
	file.Label = "/testlabel"
	file.Name = "testfile2"
	file.Size = 1
	addedFile, err := client.AddFile(file, env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.RemoveFileByName(env.Colony1Name, addedFile.Label, addedFile.Name, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveFileByName(env.Colony1Name, addedFile.Label, addedFile.Name, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveFileByName(env.Colony1Name, addedFile.Label, addedFile.Name, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveFileByName(env.Colony1Name, addedFile.Label, addedFile.Name, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}
