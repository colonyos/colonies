package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddFile(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	file := utils.CreateTestFile(env.colonyName)
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

	file := utils.CreateTestFile(env.colonyName)
	addedFile, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	fileFromServer, err := client.GetFileByID(env.colonyName, addedFile.ID, env.executorPrvKey)
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

	label := "/testprefix"
	name := "testfile"

	file := utils.CreateTestFile(env.colonyName)
	file.Label = label
	file.Name = name
	file.Size = 1
	_, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	fileFromServer, err := client.GetFileByName(env.colonyName, label, name, env.executorPrvKey)
	assert.Len(t, fileFromServer, 1)
	file2 := fileFromServer[0]
	file.ID = file2.ID
	file.Added = file2.Added
	file.SequenceNumber = file2.SequenceNumber
	assert.True(t, file.Equals(file2))
	assert.Equal(t, file.Size, int64(1))

	// Add another file so that there are two revisions
	file = utils.CreateTestFile(env.colonyName)
	file.Label = label
	file.Name = name
	file.Size = 2
	_, err = client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	fileFromServer, err = client.GetFileByName(env.colonyName, label, name, env.executorPrvKey)
	assert.Len(t, fileFromServer, 2)

	var sum int64
	for _, f := range fileFromServer {
		sum += f.Size
	}
	assert.Equal(t, sum, int64(3))

	// Try to get the latest revision
	fileFromServer, err = client.GetLatestFileByName(env.colonyName, label, name, env.executorPrvKey)
	assert.Len(t, fileFromServer, 1)
	file2 = fileFromServer[0]
	assert.Equal(t, file2.Size, int64(2))

	server.Shutdown()
	<-done
}

func TestGetFiles(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	label := "/testprefix"

	filenames, err := client.GetFilenames(env.colonyName, label, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, filenames, 0)

	file := utils.CreateTestFile(env.colonyName)
	file.Label = label
	file.Name = "testfile1"
	file.Size = 1
	_, err = client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.colonyName)
	file.Label = label
	file.Name = "testfile2"
	file.Size = 1
	_, err = client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	filenames, err = client.GetFilenames(env.colonyName, "prefix_does_not_exists", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, filenames, 0)

	filenames, err = client.GetFilenames("colony_does_not_exists", label, env.executorPrvKey)
	assert.NotNil(t, err) // Should not work

	filenames, err = client.GetFilenames(env.colonyName, label, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, filenames, 2)

	server.Shutdown()
	<-done
}

func TestGetFileLabels(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	file := utils.CreateTestFile(env.colonyName)
	file.Label = "/testlabel1"
	file.Name = "testfile1"
	file.Size = 1
	_, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.colonyName)
	file.Label = "/testlabel2"
	file.Name = "testfile2"
	file.Size = 1
	_, err = client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	labels, err := client.GetFileLabels("colony_does_not_exists", env.executorPrvKey)
	assert.NotNil(t, err)

	labels, err = client.GetFileLabels(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, labels, 2)

	server.Shutdown()
	<-done
}

func TestGetFileLabelsByName(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	file := utils.CreateTestFile(env.colonyName)
	file.Label = "/testlabel1"
	file.Name = "testfile1"
	file.Size = 1
	_, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.colonyName)
	file.Label = "/testlabel2"
	file.Name = "testfile2"
	file.Size = 1
	_, err = client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.colonyName)
	file.Label = "/testlabel2/sublabel1"
	file.Name = "testfile3"
	file.Size = 1
	_, err = client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	labels, err := client.GetFileLabelsByName(env.colonyName, "/testlabel1", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, labels, 1)
	assert.Equal(t, labels[0].Name, "/testlabel1")

	labels, err = client.GetFileLabelsByName(env.colonyName, "/testlabel2", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, labels, 2)

	counter := 0
	for _, label := range labels {
		if label.Name == "/testlabel2" {
			counter++
		}
		if label.Name == "/testlabel2/sublabel1" {
			counter++
		}
	}
	assert.Equal(t, counter, 2)

	labels, err = client.GetFileLabelsByName(env.colonyName, "/testlabel", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, labels, 3)

	labels, err = client.GetFileLabelsByName(env.colonyName, "/", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, labels, 3)

	labels, err = client.GetFileLabelsByName(env.colonyName, "", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, labels, 3)

	labels, err = client.GetFileLabelsByName(env.colonyName, "does_not_exists", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, labels, 0)

	server.Shutdown()
	<-done
}

func TestRemoveFile(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	file := utils.CreateTestFile(env.colonyName)
	file.Label = "/testlabel"
	file.Name = "testfile2"
	file.Size = 1
	_, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.colonyName)
	file.Label = "/testlabel"
	file.Name = "testfile2"
	file.Size = 1
	_, err = client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.colonyName)
	file.Label = "/testlabel"
	file.Name = "testfile2"
	file.Size = 1
	addedFile, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	fileFromServer, err := client.GetFileByID(env.colonyName, addedFile.ID, env.executorPrvKey)
	assert.Len(t, fileFromServer, 1)

	err = client.RemoveFileByID(env.colonyName, addedFile.ID, env.executorPrvKey)
	assert.Nil(t, err)

	_, err = client.GetFileByID(env.colonyName, addedFile.ID, env.executorPrvKey)
	assert.NotNil(t, err)

	fileCount, err := server.db.CountFiles(env.colonyName)
	assert.Nil(t, err)
	assert.Equal(t, fileCount, 2)

	err = client.RemoveFileByName(env.colonyName, "label_does_not_exists", "testfile2", env.executorPrvKey)
	assert.Nil(t, err) // NOP

	err = client.RemoveFileByName(env.colonyName, "/testlabel", "file_does_not_exist", env.executorPrvKey)
	assert.Nil(t, err) // NOP

	err = client.RemoveFileByName("colony_does_not_exists", "/testlabel", "testfile2", env.executorPrvKey)
	assert.NotNil(t, err)

	err = client.RemoveFileByName(env.colonyName, "/testlabel", "testfile2", env.executorPrvKey)
	assert.Nil(t, err)

	fileCount, err = server.db.CountFiles(env.colonyName)
	assert.Nil(t, err)
	assert.Equal(t, fileCount, 0)

	server.Shutdown()
	<-done
}
