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

// TestAddFileUnauthorized tests adding file from different colony
func TestAddFileUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, _, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Try to add file to colony1 with executor2's key
	file := utils.CreateTestFile(colony1.Name)
	_, err = client.AddFile(file, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetFileUnauthorized tests getting file from different colony
func TestGetFileUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Add file to colony1
	file := utils.CreateTestFile(colony1.Name)
	addedFile, err := client.AddFile(file, executor1PrvKey)
	assert.Nil(t, err)

	// Try to get file from colony1 with executor2's key
	_, err = client.GetFileByID(colony1.Name, addedFile.ID, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetFilesUnauthorized tests getting files from different colony
func TestGetFilesUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Add file to colony1
	file := utils.CreateTestFile(colony1.Name)
	file.Label = "/testlabel"
	_, err = client.AddFile(file, executor1PrvKey)
	assert.Nil(t, err)

	// Try to get files from colony1 with executor2's key
	_, err = client.GetFileData(colony1.Name, "/testlabel", executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestRemoveFileUnauthorized tests removing file from different colony
func TestRemoveFileUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Add file to colony1
	file := utils.CreateTestFile(colony1.Name)
	addedFile, err := client.AddFile(file, executor1PrvKey)
	assert.Nil(t, err)

	// Try to remove file from colony1 with executor2's key
	err = client.RemoveFileByID(colony1.Name, addedFile.ID, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetFileByIDNotFound tests getting non-existent file by ID
func TestGetFileByIDNotFound(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	_, err := client.GetFileByID(env.ColonyName, "non_existent_id", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}
