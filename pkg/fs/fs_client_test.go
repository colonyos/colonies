package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

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
	f, err := os.Create(syncDir + "/" + "local_file")
	assert.Nil(t, err)
	_, err = f.Write([]byte("testdata"))
	assert.Nil(t, err)

	// Create a remote file
	file := utils.CreateTestFile(env.colonyID)
	file.Checksum = "710ff3fb242a5dee1220f1cb0e6a519891fb67f2f828a6cab4ef8894633b1f51"
	file.Name = "remote_file"
	file.Label = label
	_, err = coloniesClient.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label)
	assert.Nil(t, err)
	assert.Len(t, syncPlan.LocalMissing, 1)
	assert.Len(t, syncPlan.RemoteMissing, 1)
	assert.Equal(t, syncPlan.LocalMissing[0].Name, "remote_file")
	assert.Equal(t, syncPlan.RemoteMissing[0].Name, "local_file")

	printSyncPlan(syncPlan)

	// Clean up
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
	f, err := os.Create(syncDir + "/" + "same_file")
	assert.Nil(t, err)
	_, err = f.Write([]byte("testdata"))
	assert.Nil(t, err)

	// Create a remote file
	file := utils.CreateTestFile(env.colonyID)
	file.Checksum = "810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"
	file.Name = "same_file"
	file.Label = label
	_, err = coloniesClient.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label)
	assert.Nil(t, err)
	assert.Len(t, syncPlan.LocalMissing, 0)
	assert.Len(t, syncPlan.RemoteMissing, 0)

	printSyncPlan(syncPlan)

	// Clean up
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
	f, err := os.Create(syncDir + "/" + "same_file")
	assert.Nil(t, err)
	_, err = f.Write([]byte("testdata"))
	assert.Nil(t, err)

	// Create a remote file
	file := utils.CreateTestFile(env.colonyID)
	file.Checksum = "710ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50" // Different checksum
	file.Name = "same_file"
	file.Label = label
	_, err = coloniesClient.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label)
	assert.Nil(t, err)
	assert.Len(t, syncPlan.LocalMissing, 0)
	assert.Len(t, syncPlan.RemoteMissing, 0)
	assert.Len(t, syncPlan.Conflicts, 1)

	printSyncPlan(syncPlan)

	// Clean up
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

// Scenario:
//
//	Remote: ""
//	Local:  "local_file"
//
//	Expected result: File is uploaded to server
func TestApplySyncPlan1(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"

	// Create a local file
	syncDir, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)
	f, err := ioutil.TempFile(syncDir, "test")
	assert.Nil(t, err)
	_, err = f.Write([]byte("testdata"))
	assert.Nil(t, err)

	// Calculate a sync plan
	fsClient, err := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	syncPlan, err := fsClient.CalcSyncPlan(syncDir, label)
	assert.Nil(t, err)

	printSyncPlan(syncPlan)

	err = fsClient.ApplySyncPlan(env.colonyID, syncPlan, true)

	// Clean up
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}
