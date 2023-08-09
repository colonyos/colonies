package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func printResult(localMissing []*FileInfo, remoteMissing []*FileInfo, localOverwrite []*FileInfo, remoteOverwrite []*FileInfo) {
	fmt.Println("Missing local:", len(localMissing))
	for _, fileInfo := range localMissing {
		fmt.Println("  file:", fileInfo.Name)
	}

	fmt.Println("Overwrite local:", len(localOverwrite))
	for _, fileInfo := range localOverwrite {
		fmt.Println("  file:", fileInfo.Name)
	}

	fmt.Println("Missing remote:", len(localMissing))
	for _, fileInfo := range remoteMissing {
		fmt.Println("  file:", fileInfo.Name)
	}

	fmt.Println("Overwrite remote:", len(localOverwrite))
	for _, fileInfo := range remoteOverwrite {
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
//	Remote missing: ["local_file"]
//	Local missing: ["remote_file"]
//	Local overwrite: []
//	Remote overwrite: []
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
	fsClient := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
	localMissing, remoteMissing, localOverwrite, remoteOverwrite, err := fsClient.CalcSyncPlan(syncDir, label)
	assert.Nil(t, err)
	assert.Len(t, localMissing, 1)
	assert.Len(t, remoteMissing, 1)
	assert.Equal(t, localMissing[0].Name, "remote_file")
	assert.Equal(t, remoteMissing[0].Name, "local_file")

	printResult(localMissing, remoteMissing, localOverwrite, remoteOverwrite)

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
//	Remote missing: []
//	Local missing: []
//	Local overwrite: []
//	Remote overwrite: []
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
	fsClient := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
	localMissing, remoteMissing, localOverwrite, remoteOverwrite, err := fsClient.CalcSyncPlan(syncDir, label)
	assert.Nil(t, err)
	assert.Len(t, localMissing, 0)
	assert.Len(t, remoteMissing, 0)

	printResult(localMissing, remoteMissing, localOverwrite, remoteOverwrite)

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
//	Local overwrite: ["same_file"]
//	Remote overwrite: ["same_file"]
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
	fsClient := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
	localMissing, remoteMissing, localOverwrite, remoteOverwrite, err := fsClient.CalcSyncPlan(syncDir, label)
	assert.Nil(t, err)
	assert.Len(t, localMissing, 0)
	assert.Len(t, remoteMissing, 0)

	printResult(localMissing, remoteMissing, localOverwrite, remoteOverwrite)

	// Clean up
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}
