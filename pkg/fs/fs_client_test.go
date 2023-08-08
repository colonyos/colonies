package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

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

func TestSync(t *testing.T) {
	env, coloniesClient, coloniesServer, _, done := setupTestEnv(t)

	label := "test_label"

	// Create a local file
	syncDir, err := ioutil.TempDir("/tmp/", "sync")
	assert.Nil(t, err)
	f, err := os.Create(syncDir + "/local_file")
	assert.Nil(t, err)
	_, err = f.Write([]byte("testdata"))
	assert.Nil(t, err)

	// Create a remote file
	file := utils.CreateTestFile(env.colonyID)
	file.Checksum = "710ff3fb242a5dee1220f1cb0e6a519891fb67f2f828a6cab4ef8894633b1f51"
	file.Name = "remote_file"
	file.Label = label
	_, err = coloniesClient.AddFile(file, env.executorPrvKey)

	// Calculate a sync plan
	fsClient := CreateFSClient(coloniesClient, env.colonyID, env.executorPrvKey)
	localMissingFiles, remoteMissingFiles, localOverwrite, remoteOverwrite, err := fsClient.CalcSyncPlan(syncDir, label)
	assert.Nil(t, err)
	assert.Len(t, localMissingFiles, 1)
	assert.Len(t, remoteMissingFiles, 1)
	assert.Equal(t, localMissingFiles[0].Name, "remote_file")
	assert.Equal(t, remoteMissingFiles[0].Name, "local_file")

	for _, fileInfo := range localMissingFiles {
		fmt.Println("Missing local:", fileInfo)
	}
	for _, fileInfo := range localOverwrite {
		fmt.Println("Overwrite local:", fileInfo)
	}

	for _, fileInfo := range remoteMissingFiles {
		fmt.Println("Missing remote:", fileInfo)
	}
	for _, fileInfo := range remoteOverwrite {
		fmt.Println("Overwrite remote:", fileInfo)
	}

	// Clean up
	err = os.RemoveAll(syncDir)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}
