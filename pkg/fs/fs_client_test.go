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
	fileNames, err := coloniesClient.GetFilenames(env.colonyID, label, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, fileNames, 1)
	orgFilename := filepath.Base(f.Name())
	assert.Equal(t, fileNames[0], orgFilename)
	coloniesFile, err := coloniesClient.GetFileByName(env.colonyID, label, orgFilename, env.executorPrvKey)
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
	coloniesFile := utils.CreateTestFile(env.colonyID)
	coloniesFile.Checksum = "710ff3fb242a5dee1220f1cb0e6a519891fb67f2f828a6cab4ef8894633b1f51"
	coloniesFile.Name = "remote_file"
	coloniesFile.Label = label
	_, err = coloniesClient.AddFile(coloniesFile, env.executorPrvKey)
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
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
	coloniesFile := utils.CreateTestFile(env.colonyID)
	coloniesFile.Checksum = "810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"
	coloniesFile.Name = "same_file"
	coloniesFile.Label = label
	_, err = coloniesClient.AddFile(coloniesFile, env.executorPrvKey)
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
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
	coloniesFile := utils.CreateTestFile(env.colonyID)
	coloniesFile.Checksum = "710ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50" // Different checksum
	coloniesFile.Name = "same_file"
	coloniesFile.Label = label
	_, err = coloniesClient.AddFile(coloniesFile, env.executorPrvKey)
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
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
	fsClient, err := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	//printSyncPlan(syncPlan)
	err = fsClient.ApplySyncPlan(env.colonyID, syncPlan)
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
	fsClient, err := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	// Upload the file to the server
	err = fsClient.ApplySyncPlan(env.colonyID, syncPlan)
	assert.Nil(t, err)

	syncDir2, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)

	syncPlan2, err := fsClient.CalcSyncPlan(syncDir2, label, true)
	assert.Nil(t, err)
	err = fsClient.ApplySyncPlan(env.colonyID, syncPlan2)
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
	fsClient, err := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	// Upload the file to the server
	err = fsClient.ApplySyncPlan(env.colonyID, syncPlan)
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
	err = fsClient.ApplySyncPlan(env.colonyID, syncPlan2)
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
	fsClient, err := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	// Upload the file to the server
	err = fsClient.ApplySyncPlan(env.colonyID, syncPlan)
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
	err = fsClient.ApplySyncPlan(env.colonyID, syncPlan)
	assert.Nil(t, err)

	// Verify that local file is not replaced
	tmpFile1Filename := filepath.Base(tmpFile1.Name())
	fileContent, err := os.ReadFile(syncDir + "/" + tmpFile1Filename)
	assert.Nil(t, err)
	assert.Equal(t, string(data2), (string(fileContent)))

	remoteColoniesFile, err := coloniesClient.GetFileByName(env.colonyID, label, tmpFile1Filename, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, remoteColoniesFile, 2)

	remoteColoniesFile, err = coloniesClient.GetLatestFileByName(env.colonyID, label, tmpFile1Filename, env.executorPrvKey)
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
	fsClient, err := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	// Upload the file to the server
	err = fsClient.ApplySyncPlan(env.colonyID, syncPlan)
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
	err = fsClient.ApplySyncPlan(env.colonyID, syncPlan)
	assert.Nil(t, err)

	// Verify that local file is replaced
	tmpFile1Filename := filepath.Base(tmpFile1.Name())
	fileContent, err := os.ReadFile(syncDir + "/" + tmpFile1Filename)
	assert.Nil(t, err)
	assert.Equal(t, string(data1), (string(fileContent)))

	remoteColoniesFile, err := coloniesClient.GetFileByName(env.colonyID, label, tmpFile1Filename, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, remoteColoniesFile, 1) // We did not upload another revision

	remoteColoniesFile, err = coloniesClient.GetLatestFileByName(env.colonyID, label, tmpFile1Filename, env.executorPrvKey)
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
	fsClient, err := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label, true)
	assert.Nil(t, err)

	// Upload the file to the server
	err = fsClient.ApplySyncPlan(env.colonyID, syncPlan)
	assert.Nil(t, err)

	coloniesFile, err := fsClient.coloniesClient.GetFileByName(env.colonyID, label, tmpFile1Filename, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, coloniesFile, 1)
	fileID := coloniesFile[0].ID

	downloadDir, err := ioutil.TempDir("/tmp/", "download")
	assert.Nil(t, err)

	err = fsClient.Download(env.colonyID, fileID, downloadDir)
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
