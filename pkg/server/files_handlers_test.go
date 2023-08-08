package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddFile(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	file := utils.CreateTestFile(env.colonyID)
	addedFile, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedFile)
	file.ID = addedFile.ID
	file.Added = addedFile.Added
	file.SequenceNumber = addedFile.SequenceNumber
	assert.True(t, file.Equals(addedFile))

	server.Shutdown()
	<-done
}

func TestGetFileByID(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	file := utils.CreateTestFile(env.colonyID)
	addedFile, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	fileFromServer, err := client.GetFileByID(env.colonyID, addedFile.ID, env.executorPrvKey)
	assert.Len(t, fileFromServer, 1)
	file2 := fileFromServer[0]
	file.ID = file2.ID
	file.Added = file2.Added
	file.SequenceNumber = file2.SequenceNumber
	assert.True(t, file.Equals(file2))

	server.Shutdown()
	<-done
}

func TestGetFileByName(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	prefix := "/testprefix"
	name := "testfile"

	file := utils.CreateTestFile(env.colonyID)
	file.Prefix = prefix
	file.Name = name
	file.Size = 1
	_, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	fileFromServer, err := client.GetFileByName(env.colonyID, prefix, name, env.executorPrvKey)
	assert.Len(t, fileFromServer, 1)
	file2 := fileFromServer[0]
	file.ID = file2.ID
	file.Added = file2.Added
	file.SequenceNumber = file2.SequenceNumber
	assert.True(t, file.Equals(file2))
	assert.Equal(t, file.Size, int64(1))

	// Add another file so that there are two revisions
	file = utils.CreateTestFile(env.colonyID)
	file.Prefix = prefix
	file.Name = name
	file.Size = 2
	_, err = client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	fileFromServer, err = client.GetFileByName(env.colonyID, prefix, name, env.executorPrvKey)
	assert.Len(t, fileFromServer, 2)

	var sum int64
	for _, f := range fileFromServer {
		sum += f.Size
	}
	assert.Equal(t, sum, int64(3))

	// Try to get the latest revision
	fileFromServer, err = client.GetLatestFileByName(env.colonyID, prefix, name, env.executorPrvKey)
	assert.Len(t, fileFromServer, 1)
	file2 = fileFromServer[0]
	assert.Equal(t, file2.Size, int64(2))

	server.Shutdown()
	<-done
}

func TestGetFiles(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	prefix := "/testprefix"

	filenames, err := client.GetFilenames(env.colonyID, prefix, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, filenames, 0)

	file := utils.CreateTestFile(env.colonyID)
	file.Prefix = prefix
	file.Name = "testfile1"
	file.Size = 1
	_, err = client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.colonyID)
	file.Prefix = prefix
	file.Name = "testfile2"
	file.Size = 1
	_, err = client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	filenames, err = client.GetFilenames(env.colonyID, "prefix_does_not_exists", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, filenames, 0)

	filenames, err = client.GetFilenames("colony_does_not_exists", prefix, env.executorPrvKey)
	assert.NotNil(t, err) // Should not work

	filenames, err = client.GetFilenames(env.colonyID, prefix, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, filenames, 2)

	server.Shutdown()
	<-done
}

func TestGetFilePrefixes(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	file := utils.CreateTestFile(env.colonyID)
	file.Prefix = "/testprefix1"
	file.Name = "testfile1"
	file.Size = 1
	_, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.colonyID)
	file.Prefix = "/testprefix2"
	file.Name = "testfile2"
	file.Size = 1
	_, err = client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	prefixes, err := client.GetFilePrefixes("colony_does_not_exists", env.executorPrvKey)
	assert.NotNil(t, err)

	prefixes, err = client.GetFilePrefixes(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, prefixes, 2)

	server.Shutdown()
	<-done
}

func TestDeleteFile(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	file := utils.CreateTestFile(env.colonyID)
	file.Prefix = "/testprefix"
	file.Name = "testfile2"
	file.Size = 1
	_, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.colonyID)
	file.Prefix = "/testprefix"
	file.Name = "testfile2"
	file.Size = 1
	_, err = client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.colonyID)
	file.Prefix = "/testprefix"
	file.Name = "testfile2"
	file.Size = 1
	addedFile, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	fileFromServer, err := client.GetFileByID(env.colonyID, addedFile.ID, env.executorPrvKey)
	assert.Len(t, fileFromServer, 1)

	err = client.DeleteFileByID(env.colonyID, addedFile.ID, env.executorPrvKey)
	assert.Nil(t, err)

	_, err = client.GetFileByID(env.colonyID, addedFile.ID, env.executorPrvKey)
	assert.NotNil(t, err)

	fileCount, err := server.db.CountFiles(env.colonyID)
	assert.Nil(t, err)
	assert.Equal(t, fileCount, 2)

	err = client.DeleteFileByName(env.colonyID, "prefix_does_not_exists", "testfile2", env.executorPrvKey)
	assert.Nil(t, err) // NOP

	err = client.DeleteFileByName(env.colonyID, "/testprefix", "file_does_not_exist", env.executorPrvKey)
	assert.Nil(t, err) // NOP

	err = client.DeleteFileByName("colony_does_not_exists", "/testprefix", "testfile2", env.executorPrvKey)
	assert.NotNil(t, err)

	err = client.DeleteFileByName(env.colonyID, "/testprefix", "testfile2", env.executorPrvKey)
	assert.Nil(t, err)

	fileCount, err = server.db.CountFiles(env.colonyID)
	assert.Nil(t, err)
	assert.Equal(t, fileCount, 0)

	server.Shutdown()
	<-done
}
