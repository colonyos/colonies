package fs

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func printSyncPlan(syncPlan *SyncPlan) {
	fmt.Println("Missing local:", len(syncPlan.LocalMissing))
	for _, fileInfo := range syncPlan.LocalMissing {
		fmt.Println("  file:", fileInfo.Name)
	}

	fmt.Println("Missing remote:", len(syncPlan.LocalMissing))
	for _, fileInfo := range syncPlan.RemoteMissing {
		fmt.Println("  file:", fileInfo.Name)
	}

	fmt.Println("Conflicts:", len(syncPlan.Conflicts))
	for _, fileInfo := range syncPlan.Conflicts {
		fmt.Println("  file:", fileInfo.Name)
	}
}

func checkFile(t *testing.T, env *testEnv, label string, coloniesClient *client.ColoniesClient, f *os.File) {
	fileNames, err := coloniesClient.GetFilenames(env.colonyName, label, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, fileNames, 1)
	orgFilename := filepath.Base(f.Name())
	assert.Equal(t, fileNames[0], orgFilename)
	coloniesFile, err := coloniesClient.GetFileByName(env.colonyName, label, orgFilename, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, coloniesFile, 1)
	assert.Equal(t, coloniesFile[0].Name, orgFilename)
}

func generateRandomData(size int) []byte {
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, size)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return b
}

func TestChecksum(t *testing.T) {
	testDir, err := ioutil.TempDir("/tmp/", "test")
	assert.Nil(t, err)
	f, err := os.Create(testDir + "/test_file")
	assert.Nil(t, err)
	_, err = f.Write([]byte("testdata"))
	assert.Nil(t, err)

	checksum, err := checksum(testDir + "/test_file")
	assert.Nil(t, err)
	assert.Equal(t, checksum, "810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50")

	err = os.RemoveAll(testDir)
	assert.Nil(t, err)
}

// Scenario:
//
//	Remote: "remote_file"
//	Local:  "local_file"
//
//	Expected result:
//	  Remote missing: ["local_file"]
//	  Local missing: ["remote_file"]
//	  Conflicts: []
func TestCalcSyncPlan1(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"

	// Create a local file
	syncDir, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)
	localFile, err := os.Create(syncDir + "/" + "local_file")
	assert.Nil(t, err)
	_, err = localFile.Write([]byte("testdata"))
	assert.Nil(t, err)

	// Create a remote file
	coloniesFile := utils.CreateTestFile(env.colonyName)
	coloniesFile.Checksum = "710ff3fb242a5dee1220f1cb0e6a519891fb67f2f828a6cab4ef8894633b1f51"
	coloniesFile.Name = "remote_file"
	coloniesFile.Label = label
	_, err = coloniesClient.AddFile(coloniesFile, env.executorPrvKey)
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)
	assert.Len(t, syncPlan.LocalMissing, 1)
	assert.Len(t, syncPlan.RemoteMissing, 1)
	assert.Equal(t, syncPlan.LocalMissing[0].Name, "remote_file")
	assert.Equal(t, syncPlan.RemoteMissing[0].Name, "local_file")

	//printSyncPlan(syncPlan)

	// Clean up
	localFile.Close()
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

// Scenario:
//
//	Remote: "same_file"    <- two identical files
//	Local:  "same_file"
//
//	Expected result:
//	  Remote missing: []
//	  Local missing: []
//	  Conflicts: []
func TestCalcSyncPlan2(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"

	// Create a local file
	syncDir, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)
	sameFile, err := os.Create(syncDir + "/" + "same_file")
	assert.Nil(t, err)
	_, err = sameFile.Write([]byte("testdata"))
	assert.Nil(t, err)

	// Create a remote file
	coloniesFile := utils.CreateTestFile(env.colonyName)
	coloniesFile.Checksum = "810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"
	coloniesFile.Name = "same_file"
	coloniesFile.Label = label
	_, err = coloniesClient.AddFile(coloniesFile, env.executorPrvKey)
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)
	assert.Len(t, syncPlan.LocalMissing, 0)
	assert.Len(t, syncPlan.RemoteMissing, 0)

	//printSyncPlan(syncPlan)

	// Clean up
	sameFile.Close()
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

// Scenario:
//
//	Remote: "same_file"    <- different checksum
//	Local:  "same_file"
//
//	Expected result:
//	Remote missing: []
//	Local missing: []
//	  Conflicts: ["same_file"]
func TestCalcSyncPlan3(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"

	// Create a local file
	syncDir, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)
	sameFile, err := os.Create(syncDir + "/" + "same_file")
	assert.Nil(t, err)
	_, err = sameFile.Write([]byte("testdata"))
	assert.Nil(t, err)

	// Create a remote file
	coloniesFile := utils.CreateTestFile(env.colonyName)
	coloniesFile.Checksum = "710ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50" // Different checksum
	coloniesFile.Name = "same_file"
	coloniesFile.Label = label
	_, err = coloniesClient.AddFile(coloniesFile, env.executorPrvKey)
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)
	assert.Len(t, syncPlan.LocalMissing, 0)
	assert.Len(t, syncPlan.RemoteMissing, 0)
	assert.Len(t, syncPlan.Conflicts, 1)

	//printSyncPlan(syncPlan)

	// Clean up
	sameFile.Close()
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

// Scenario:
//
//	Remote:
//	Local:  tmpFile1
//
//	Expected result: tmpFile1 is uploaded to server
func TestApplySyncPlan1(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"

	// Create tmpFile1
	syncDir, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)
	tmpFile1, err := ioutil.TempFile(syncDir, "test")
	assert.Nil(t, err)
	_, err = tmpFile1.Write([]byte("testdata"))
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	//printSyncPlan(syncPlan)
	err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
	assert.Nil(t, err)
	checkFile(t, env, label, coloniesClient, tmpFile1)

	// Clean up
	tmpFile1.Close()
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

// Scenario:
//
//	Remote: tmpFile1
//	Local:
//
//	Expected result: tmpFile1 is downloaded to client
func TestApplySyncPlan2(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"

	// Create tmpFile1
	syncDir, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)
	tmpFile1, err := ioutil.TempFile(syncDir, "test")
	assert.Nil(t, err)
	_, err = tmpFile1.Write([]byte("testdata"))
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	// Upload the file to the server
	err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
	assert.Nil(t, err)

	syncDir2, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)

	syncPlan2, err := fsClient.CalcSyncPlan(syncDir2, label, true)
	assert.Nil(t, err)
	err = fsClient.ApplySyncPlan(env.colonyName, syncPlan2)
	assert.Nil(t, err)

	tmpFile1Filename := filepath.Base(tmpFile1.Name())
	fileContent, err := os.ReadFile(syncDir2 + "/" + tmpFile1Filename)
	assert.Nil(t, err)
	assert.Equal(t, "testdata", (string(fileContent)))

	// Clean up
	tmpFile1.Close()
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)
	err = os.RemoveAll(syncDir2)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

// Scenario:
//
//	Remote: tmpFile1
//	Local: tmpFile2
//
//	Expected result: tmpFile1 is downloaded and temp tmpFile2 is uploaded
func TestApplySyncPlan3(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"
	fileSize := 1000

	// Create tmpFile1
	syncDir, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)
	tmpFile1, err := ioutil.TempFile(syncDir, "test")
	assert.Nil(t, err)
	data1 := generateRandomData(fileSize)
	_, err = tmpFile1.Write(data1)
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	// Upload the file to the server
	err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
	assert.Nil(t, err)

	// Create tmpFile2
	syncDir2, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)
	tmpFile2, err := ioutil.TempFile(syncDir2, "test")
	assert.Nil(t, err)
	data2 := generateRandomData(fileSize)
	_, err = tmpFile2.Write(data2)
	assert.Nil(t, err)

	syncPlan2, err := fsClient.CalcSyncPlan(syncDir2, label, true)
	assert.Nil(t, err)
	err = fsClient.ApplySyncPlan(env.colonyName, syncPlan2)
	assert.Nil(t, err)

	// Check that we got both files
	tmpFile1Filename := filepath.Base(tmpFile1.Name())
	fileContent, err := os.ReadFile(syncDir2 + "/" + tmpFile1Filename)
	assert.Nil(t, err)
	assert.Equal(t, string(data1), (string(fileContent)))

	tmpFile2Filename := filepath.Base(tmpFile2.Name())
	fileContent, err = os.ReadFile(syncDir2 + "/" + tmpFile2Filename)
	assert.Nil(t, err)
	assert.Equal(t, string(data2), (string(fileContent)))

	// Clean up
	tmpFile1.Close()
	tmpFile2.Close()
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)
	err = os.RemoveAll(syncDir2)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

// Scenario: Test conflict management (keep local)
//
//	Remote: tmpFile1    <- file is changed
//	Local: tmpFile1
//
//	Expected result: tmpFile1 at server is replaced
func TestApplySyncPlan4(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"
	fileSize := 1000

	// Create tmpFile1
	syncDir, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)
	tmpFile1, err := ioutil.TempFile(syncDir, "test")
	assert.Nil(t, err)
	data1 := generateRandomData(fileSize)
	_, err = tmpFile1.Write(data1)
	assert.Nil(t, err)
	orgChecksum, err := checksum(tmpFile1.Name())
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	// Upload the file to the server
	err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
	assert.Nil(t, err)

	// Replace content in tmpFile1
	err = tmpFile1.Truncate(0)
	assert.Nil(t, err)
	_, err = tmpFile1.Seek(0, 0)
	assert.Nil(t, err)
	data2 := generateRandomData(fileSize)
	_, err = tmpFile1.Write(data2)
	assert.Nil(t, err)
	replacedChecksum, err := checksum(tmpFile1.Name())
	assert.Nil(t, err)

	// Make another sync
	keepLocal := true
	syncPlan, err = fsClient.CalcSyncPlan(syncDir, label, keepLocal)
	assert.Nil(t, err)
	assert.Len(t, syncPlan.Conflicts, 1)
	err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
	assert.Nil(t, err)

	// Verify that local file is not replaced
	tmpFile1Filename := filepath.Base(tmpFile1.Name())
	fileContent, err := os.ReadFile(syncDir + "/" + tmpFile1Filename)
	assert.Nil(t, err)
	assert.Equal(t, string(data2), (string(fileContent)))

	remoteColoniesFile, err := coloniesClient.GetFileByName(env.colonyName, label, tmpFile1Filename, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, remoteColoniesFile, 2)

	remoteColoniesFile, err = coloniesClient.GetLatestFileByName(env.colonyName, label, tmpFile1Filename, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, remoteColoniesFile, 1)
	remoteChecksum := remoteColoniesFile[0].Checksum

	localChecksum, err := checksum(tmpFile1.Name())
	assert.Nil(t, err)
	assert.Equal(t, remoteChecksum, localChecksum)
	assert.NotEqual(t, remoteChecksum, orgChecksum)
	assert.Equal(t, remoteChecksum, replacedChecksum)

	// Clean up
	tmpFile1.Close()
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

// Scenario: Test conflict management (keep remote)
//
//	Remote: tmpFile1    <- file is changed
//	Local: tmpFile1
//
//	Expected result: local tmpFile is replaced
func TestApplySyncPlan5(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"
	fileSize := 1000

	// Create tmpFile1
	syncDir, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)
	tmpFile1, err := ioutil.TempFile(syncDir, "test")
	assert.Nil(t, err)
	data1 := generateRandomData(fileSize)
	_, err = tmpFile1.Write(data1)
	assert.Nil(t, err)
	orgChecksum, err := checksum(tmpFile1.Name())
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	// Upload the file to the server
	err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
	assert.Nil(t, err)

	// Replace content in tmpFile1
	err = tmpFile1.Truncate(0)
	assert.Nil(t, err)
	_, err = tmpFile1.Seek(0, 0)
	assert.Nil(t, err)
	data2 := generateRandomData(fileSize)
	_, err = tmpFile1.Write(data2)
	assert.Nil(t, err)
	replacedChecksum, err := checksum(tmpFile1.Name())
	assert.Nil(t, err)

	// Make another sync
	keepLocal := false // keep remote
	syncPlan, err = fsClient.CalcSyncPlan(syncDir, label, keepLocal)
	assert.Nil(t, err)
	assert.Len(t, syncPlan.Conflicts, 1)
	err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
	assert.Nil(t, err)

	// Verify that local file is replaced
	tmpFile1Filename := filepath.Base(tmpFile1.Name())
	fileContent, err := os.ReadFile(syncDir + "/" + tmpFile1Filename)
	assert.Nil(t, err)
	assert.Equal(t, string(data1), (string(fileContent)))

	remoteColoniesFile, err := coloniesClient.GetFileByName(env.colonyName, label, tmpFile1Filename, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, remoteColoniesFile, 1) // We did not upload another revision

	remoteColoniesFile, err = coloniesClient.GetLatestFileByName(env.colonyName, label, tmpFile1Filename, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, remoteColoniesFile, 1)
	remoteChecksum := remoteColoniesFile[0].Checksum

	localChecksum, err := checksum(tmpFile1.Name())
	assert.Nil(t, err)
	assert.Equal(t, remoteChecksum, localChecksum)
	assert.Equal(t, remoteChecksum, orgChecksum)
	assert.NotEqual(t, remoteChecksum, replacedChecksum)

	// Clean up
	tmpFile1.Close()
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

// Scenario:
//
//	Remote: tmpFile1
//	Local:
//
//	Expected result: tmpFile1 is downloaded to client
func TestDownload(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"

	// Create tmpFile1
	syncDir, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)
	tmpFile1, err := ioutil.TempFile(syncDir, "test")
	assert.Nil(t, err)
	tmpFile1Filename := filepath.Base(tmpFile1.Name())
	_, err = tmpFile1.Write([]byte("testdata"))
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	// Upload the file to the server
	err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
	assert.Nil(t, err)

	coloniesFile, err := fsClient.coloniesClient.GetFileByName(env.colonyName, label, tmpFile1Filename, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, coloniesFile, 1)
	fileID := coloniesFile[0].ID

	downloadDir, err := ioutil.TempDir("/tmp/", "download")
	assert.Nil(t, err)

	err = fsClient.Download(env.colonyName, fileID, downloadDir)
	assert.Nil(t, err)

	fileContent, err := os.ReadFile(downloadDir + "/" + tmpFile1Filename)
	assert.Nil(t, err)
	assert.Equal(t, "testdata", (string(fileContent)))

	// Clean up
	tmpFile1.Close()
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)
	err = os.RemoveAll(downloadDir)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveByID(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"

	// Create tmpFile1
	syncDir, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)
	tmpFile1, err := ioutil.TempFile(syncDir, "test")
	assert.Nil(t, err)
	tmpFile1Filename := filepath.Base(tmpFile1.Name())
	_, err = tmpFile1.Write([]byte("testdata"))
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	// Upload the file to the server
	err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
	assert.Nil(t, err)

	coloniesFile, err := fsClient.coloniesClient.GetFileByName(env.colonyName, label, tmpFile1Filename, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, coloniesFile, 1)
	fileID := coloniesFile[0].ID

	err = fsClient.RemoveFileByID(env.colonyName, fileID)
	assert.Nil(t, err)

	coloniesFile, err = fsClient.coloniesClient.GetFileByID(env.colonyName, fileID, env.executorPrvKey)
	assert.NotNil(t, err)

	// Clean up
	tmpFile1.Close()
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveByName(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"

	// Create tmpFile1
	syncDir, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)
	tmpFile1, err := ioutil.TempFile(syncDir, "test")
	assert.Nil(t, err)
	tmpFile1Filename := filepath.Base(tmpFile1.Name())
	_, err = tmpFile1.Write([]byte("testdata"))
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	// Upload the file to the server
	err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
	assert.Nil(t, err)

	// Modify the file so that we get two revisions
	err = tmpFile1.Truncate(0)
	assert.Nil(t, err)
	_, err = tmpFile1.Seek(0, 0)
	assert.Nil(t, err)
	_, err = tmpFile1.Write([]byte("testdata2"))
	assert.Nil(t, err)

	// Make another sync
	keepLocal := true
	syncPlan, err = fsClient.CalcSyncPlan(syncDir, label, keepLocal)
	assert.Nil(t, err)
	err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
	assert.Nil(t, err)

	// Get the file
	coloniesFile, err := fsClient.coloniesClient.GetFileByName(env.colonyName, label, tmpFile1Filename, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, coloniesFile, 2)

	err = fsClient.RemoveFileByName(env.colonyName, label, tmpFile1Filename)
	assert.Nil(t, err)

	coloniesFile, err = fsClient.coloniesClient.GetFileByName(env.colonyName, label, tmpFile1Filename, env.executorPrvKey)
	assert.NotNil(t, err)

	// Clean up
	tmpFile1.Close()
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestDownloadSnapshot(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"

	// Create tmpFile1
	syncDir, err := ioutil.TempDir("/tmp/", "sync1")
	assert.Nil(t, err)
	tmpFile1, err := ioutil.TempFile(syncDir, "test")
	assert.Nil(t, err)
	tmpFile1Filename := filepath.Base(tmpFile1.Name())
	_, err = tmpFile1.Write([]byte("testdata1"))
	assert.Nil(t, err)

	// Create tmpFile2
	tmpFile2, err := ioutil.TempFile(syncDir, "test2")
	assert.Nil(t, err)
	tmpFile2Filename := filepath.Base(tmpFile2.Name())
	_, err = tmpFile2.Write([]byte("testdata2"))
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	// Upload the file to the server
	err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
	assert.Nil(t, err)

	// Create a snapshot
	snapshot, err := fsClient.coloniesClient.CreateSnapshot(env.colonyName, label, "test_snapshot1", env.executorPrvKey)
	assert.Nil(t, err)

	// Download files in snapshot
	downloadDir, err := ioutil.TempDir("/tmp/", "download")
	err = fsClient.DownloadSnapshot(snapshot.ID, downloadDir)
	assert.Nil(t, err)

	// Get the files
	fileContent, err := os.ReadFile(downloadDir + "/" + tmpFile1Filename)
	assert.Nil(t, err)
	assert.Equal(t, "testdata1", (string(fileContent)))
	fileContent, err = os.ReadFile(downloadDir + "/" + tmpFile2Filename)
	assert.Nil(t, err)
	assert.Equal(t, "testdata2", (string(fileContent)))

	// Clean up
	tmpFile1.Close()
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)
	err = os.RemoveAll(downloadDir)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveAllFilesWithLabel(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"

	// Create tmpFile1
	syncDir, err := ioutil.TempDir("/tmp/", "sync1")
	assert.Nil(t, err)
	tmpFile1, err := ioutil.TempFile(syncDir, "test")
	assert.Nil(t, err)
	tmpFile1Filename := filepath.Base(tmpFile1.Name())
	_, err = tmpFile1.Write([]byte("testdata1"))
	assert.Nil(t, err)

	// Create tmpFile2
	tmpFile2, err := ioutil.TempFile(syncDir, "test2")
	assert.Nil(t, err)
	tmpFile2Filename := filepath.Base(tmpFile2.Name())
	_, err = tmpFile2.Write([]byte("testdata2"))
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	// Upload the file to the server
	err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
	assert.Nil(t, err)

	filenames, err := fsClient.coloniesClient.GetFilenames(env.colonyName, label, env.executorPrvKey)
	tmpFile1S3Object := ""
	tmpFile2S3Object := ""
	for _, filename := range filenames {
		file, err := fsClient.coloniesClient.GetFileByName(env.colonyName, label, filename, env.executorPrvKey)
		assert.Nil(t, err)
		assert.Len(t, file, 1)
		if file[0].Name == tmpFile1Filename {
			tmpFile1S3Object = file[0].Reference.S3Object.Object
		}
		if file[0].Name == tmpFile2Filename {
			tmpFile2S3Object = file[0].Reference.S3Object.Object
		}
	}

	assert.True(t, fsClient.s3Client.Exists(tmpFile1S3Object))
	assert.True(t, fsClient.s3Client.Exists(tmpFile2S3Object))

	// Remove all files
	err = fsClient.RemoveAllFilesWithLabel(label)
	assert.Nil(t, err)

	// Verify that files are gone
	filenames, err = fsClient.coloniesClient.GetFilenames(fsClient.colonyName, label, fsClient.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, filenames, 0)

	assert.False(t, fsClient.s3Client.Exists(tmpFile1S3Object))
	assert.False(t, fsClient.s3Client.Exists(tmpFile2S3Object))

	// Clean up
	tmpFile1.Close()
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestAddFilesRecursively(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"

	syncDir, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)

	subDirPath1 := syncDir + "/subdir1"
	err = os.Mkdir(subDirPath1, 0755)
	assert.Nil(t, err)

	subDirPath2 := syncDir + "/subdir2"
	err = os.Mkdir(subDirPath2, 0755)
	assert.Nil(t, err)

	subSubDirPath1 := subDirPath1 + "/subsubdir1"
	err = os.Mkdir(subSubDirPath1, 0755)
	assert.Nil(t, err)

	tmpFile1, err := ioutil.TempFile(syncDir, "file1")
	assert.Nil(t, err)
	filepath.Base(tmpFile1.Name())
	_, err = tmpFile1.Write([]byte("testdata1"))
	assert.Nil(t, err)

	tmpFile2, err := ioutil.TempFile(syncDir, "file2")
	assert.Nil(t, err)
	filepath.Base(tmpFile2.Name())
	_, err = tmpFile2.Write([]byte("testdata2"))
	assert.Nil(t, err)

	tmpFile3, err := ioutil.TempFile(subDirPath1, "file3")
	assert.Nil(t, err)
	filepath.Base(tmpFile3.Name())
	_, err = tmpFile3.Write([]byte("testdata3"))
	assert.Nil(t, err)

	tmpFile4, err := ioutil.TempFile(subDirPath1, "file4")
	assert.Nil(t, err)
	filepath.Base(tmpFile4.Name())
	_, err = tmpFile4.Write([]byte("testdata4"))
	assert.Nil(t, err)

	tmpFile5, err := ioutil.TempFile(subDirPath2, "file5")
	assert.Nil(t, err)
	filepath.Base(tmpFile5.Name())
	_, err = tmpFile5.Write([]byte("testdata5"))
	assert.Nil(t, err)

	tmpFile6, err := ioutil.TempFile(subSubDirPath1, "file6")
	assert.Nil(t, err)
	filepath.Base(tmpFile6.Name())
	_, err = tmpFile6.Write([]byte("testdata6"))
	assert.Nil(t, err)

	// We now have this file structure:
	//   /tmp/sync2289845301/file2939843634
	//   /tmp/sync2289845301/file11729054073
	//   /tmp/sync2289845301/subdir1/file31004664384
	//   /tmp/sync2289845301/subdir1/file42329025229
	//   /tmp/sync2289845301/subdir1/subsubdir1/file63200468450
	//   /tmp/sync2289845301/subdir2/file53082049703

	fsClient, err := CreateFSClient(coloniesClient, env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlans, err := fsClient.CalcSyncPlans(syncDir, label, true)
	assert.Nil(t, err)

	for _, syncPlan := range syncPlans {
		err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
		assert.Nil(t, err)
	}

	syncDir2, err := ioutil.TempDir("/tmp/", "sync2")
	assert.Nil(t, err)

	syncPlans, err = fsClient.CalcSyncPlans(syncDir2, label, true)
	for _, syncPlan := range syncPlans {
		err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
		assert.Nil(t, err)
	}

	same, err := areDirsSame(syncDir, syncDir2)
	assert.Nil(t, err)
	assert.True(t, same)

	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)

	err = os.RemoveAll(syncDir2)
	assert.Nil(t, err)

	labelsAtServer, err := coloniesClient.GetFileLabels(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, labelsAtServer, 4)

	coloniesServer.Shutdown()
	<-done
}

func TestDownloadSnapshopRecursively(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "/test_label"

	syncDir, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)

	subDirPath1 := syncDir + "/subdir1"
	err = os.Mkdir(subDirPath1, 0755)
	assert.Nil(t, err)

	subDirPath2 := syncDir + "/subdir2"
	err = os.Mkdir(subDirPath2, 0755)
	assert.Nil(t, err)

	subSubDirPath1 := subDirPath1 + "/subsubdir1"
	err = os.Mkdir(subSubDirPath1, 0755)
	assert.Nil(t, err)

	tmpFile1, err := ioutil.TempFile(syncDir, "file1")
	assert.Nil(t, err)
	filepath.Base(tmpFile1.Name())
	_, err = tmpFile1.Write([]byte("testdata1"))
	assert.Nil(t, err)

	tmpFile2, err := ioutil.TempFile(syncDir, "file2")
	assert.Nil(t, err)
	filepath.Base(tmpFile2.Name())
	_, err = tmpFile2.Write([]byte("testdata2"))
	assert.Nil(t, err)

	tmpFile3, err := ioutil.TempFile(subDirPath1, "file3")
	assert.Nil(t, err)
	filepath.Base(tmpFile3.Name())
	_, err = tmpFile3.Write([]byte("testdata3"))
	assert.Nil(t, err)

	tmpFile4, err := ioutil.TempFile(subDirPath1, "file4")
	assert.Nil(t, err)
	filepath.Base(tmpFile4.Name())
	_, err = tmpFile4.Write([]byte("testdata4"))
	assert.Nil(t, err)

	tmpFile5, err := ioutil.TempFile(subDirPath2, "file5")
	assert.Nil(t, err)
	filepath.Base(tmpFile5.Name())
	_, err = tmpFile5.Write([]byte("testdata5"))
	assert.Nil(t, err)

	tmpFile6, err := ioutil.TempFile(subSubDirPath1, "file6")
	assert.Nil(t, err)
	filepath.Base(tmpFile6.Name())
	_, err = tmpFile6.Write([]byte("testdata6"))
	assert.Nil(t, err)

	// We now have this file structure:
	//   /tmp/sync2289845301/file2939843634
	//   /tmp/sync2289845301/file11729054073
	//   /tmp/sync2289845301/subdir1/file31004664384
	//   /tmp/sync2289845301/subdir1/file42329025229
	//   /tmp/sync2289845301/subdir1/subsubdir1/file63200468450
	//   /tmp/sync2289845301/subdir2/file53082049703

	fsClient, err := CreateFSClient(coloniesClient, env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlans, err := fsClient.CalcSyncPlans(syncDir, label, true)
	assert.Nil(t, err)

	for _, syncPlan := range syncPlans {
		err = fsClient.ApplySyncPlan(env.colonyName, syncPlan)
		assert.Nil(t, err)
	}

	// Create a snapshot
	snapshot, err := fsClient.coloniesClient.CreateSnapshot(env.colonyName, label, "/test_snapshot1", env.executorPrvKey)
	assert.Nil(t, err)

	// Download files in snapshot
	downloadDir, err := ioutil.TempDir("/tmp/", "download")
	err = fsClient.DownloadSnapshot(snapshot.ID, downloadDir)
	assert.Nil(t, err)

	same, err := areDirsSame(syncDir, downloadDir)
	assert.Nil(t, err)
	assert.True(t, same)

	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)

	err = os.RemoveAll(downloadDir)
	assert.Nil(t, err)

	labelsAtServer, err := coloniesClient.GetFileLabels(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, labelsAtServer, 4)

	coloniesServer.Shutdown()
	<-done
}
