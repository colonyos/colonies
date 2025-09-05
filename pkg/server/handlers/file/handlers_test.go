package file_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddFile(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	file := utils.CreateTestFile(env.ColonyName)
	addedFile, err := client.AddFile(file, env.ExecutorPrvKey)
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
	env, client, server, _, done := server.SetupTestEnv2(t)

	file := utils.CreateTestFile(env.ColonyName)
	addedFile, err := client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)

	fileFromServer, err := client.GetFileByID(env.ColonyName, addedFile.ID, env.ExecutorPrvKey)
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
	env, client, server, _, done := server.SetupTestEnv2(t)

	label := "/testprefix"
	name := "testfile"

	file := utils.CreateTestFile(env.ColonyName)
	file.Label = label
	file.Name = name
	file.Size = 1
	_, err := client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)

	fileFromServer, err := client.GetFileByName(env.ColonyName, label, name, env.ExecutorPrvKey)
	assert.Len(t, fileFromServer, 1)
	file2 := fileFromServer[0]
	file.ID = file2.ID
	file.Added = file2.Added
	file.SequenceNumber = file2.SequenceNumber
	assert.True(t, file.Equals(file2))
	assert.Equal(t, file.Size, int64(1))

	// Add another file so that there are two revisions
	file = utils.CreateTestFile(env.ColonyName)
	file.Label = label
	file.Name = name
	file.Size = 2
	_, err = client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)

	fileFromServer, err = client.GetFileByName(env.ColonyName, label, name, env.ExecutorPrvKey)
	assert.Len(t, fileFromServer, 2)

	var sum int64
	for _, f := range fileFromServer {
		sum += f.Size
	}
	assert.Equal(t, sum, int64(3))

	// Try to get the latest revision
	fileFromServer, err = client.GetLatestFileByName(env.ColonyName, label, name, env.ExecutorPrvKey)
	assert.Len(t, fileFromServer, 1)
	file2 = fileFromServer[0]
	assert.Equal(t, file2.Size, int64(2))

	server.Shutdown()
	<-done
}

func TestGetFiles(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	label := "/testprefix"

	fileData, err := client.GetFileData(env.ColonyName, label, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, fileData, 0)

	file := utils.CreateTestFile(env.ColonyName)
	file.Label = label
	file.Name = "testfile1"
	file.Size = 1
	_, err = client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.ColonyName)
	file.Label = label
	file.Name = "testfile2"
	file.Size = 1
	_, err = client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)

	fileData, err = client.GetFileData(env.ColonyName, "prefix_does_not_exists", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, fileData, 0)

	fileData, err = client.GetFileData("colony_does_not_exists", label, env.ExecutorPrvKey)
	assert.NotNil(t, err) // Should not work

	fileData, err = client.GetFileData(env.ColonyName, label, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, fileData, 2)

	server.Shutdown()
	<-done
}

func TestGetFileLabels(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	file := utils.CreateTestFile(env.ColonyName)
	file.Label = "/testlabel1"
	file.Name = "testfile1"
	file.Size = 1
	_, err := client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.ColonyName)
	file.Label = "/testlabel2"
	file.Name = "testfile2"
	file.Size = 1
	_, err = client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)

	labels, err := client.GetFileLabels("colony_does_not_exists", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	labels, err = client.GetFileLabels(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, labels, 2)

	server.Shutdown()
	<-done
}

func TestGetFileLabelsByName(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	file := utils.CreateTestFile(env.ColonyName)
	file.Label = "/testlabel1"
	file.Name = "testfile1"
	file.Size = 1
	_, err := client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.ColonyName)
	file.Label = "/testlabel2"
	file.Name = "testfile2"
	file.Size = 1
	_, err = client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.ColonyName)
	file.Label = "/testlabel2/sublabel1"
	file.Name = "testfile3"
	file.Size = 1
	_, err = client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)

	labels, err := client.GetFileLabelsByName(env.ColonyName, "/testlabel1", true, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, labels, 1)
	assert.Equal(t, labels[0].Name, "/testlabel1")

	labels, err = client.GetFileLabelsByName(env.ColonyName, "/testlabel2", true, env.ExecutorPrvKey)
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

	labels, err = client.GetFileLabelsByName(env.ColonyName, "/testlabel2", true, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, labels, 2)

	labels, err = client.GetFileLabelsByName(env.ColonyName, "does_not_exists", true, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, labels, 0)

	server.Shutdown()
	<-done
}

func TestRemoveFile(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	file := utils.CreateTestFile(env.ColonyName)
	file.Label = "/testlabel"
	file.Name = "testfile2"
	file.Size = 1
	_, err := client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.ColonyName)
	file.Label = "/testlabel"
	file.Name = "testfile2"
	file.Size = 1
	_, err = client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)

	file = utils.CreateTestFile(env.ColonyName)
	file.Label = "/testlabel"
	file.Name = "testfile2"
	file.Size = 1
	addedFile, err := client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)

	fileFromServer, err := client.GetFileByID(env.ColonyName, addedFile.ID, env.ExecutorPrvKey)
	assert.Len(t, fileFromServer, 1)

	err = client.RemoveFileByID(env.ColonyName, addedFile.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.GetFileByID(env.ColonyName, addedFile.ID, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	fileCount, err := server.FileDB().CountFiles(env.ColonyName)
	assert.Nil(t, err)
	assert.Equal(t, fileCount, 2)

	err = client.RemoveFileByName(env.ColonyName, "label_does_not_exists", "testfile2", env.ExecutorPrvKey)
	assert.Nil(t, err) // NOP

	err = client.RemoveFileByName(env.ColonyName, "/testlabel", "file_does_not_exist", env.ExecutorPrvKey)
	assert.Nil(t, err) // NOP

	err = client.RemoveFileByName("colony_does_not_exists", "/testlabel", "testfile2", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	err = client.RemoveFileByName(env.ColonyName, "/testlabel", "testfile2", env.ExecutorPrvKey)
	assert.Nil(t, err)

	fileCount, err = server.FileDB().CountFiles(env.ColonyName)
	assert.Nil(t, err)
	assert.Equal(t, fileCount, 0)

	server.Shutdown()
	<-done
}
