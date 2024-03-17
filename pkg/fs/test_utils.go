package fs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type testEnv struct {
	colonyID       string
	colonyName     string
	colony         *core.Colony
	colonyPrvKey   string
	executorID     string
	executor       *core.Executor
	executorPrvKey string
}

func setupTestEnv(t *testing.T) (*testEnv, *client.ColoniesClient, *server.ColoniesServer, string, chan bool) {
	rand.Seed(time.Now().UTC().UnixNano())

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
	//log.SetLevel(log.DebugLevel)
	client, server, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	env := &testEnv{
		colonyName:     colony.Name,
		colonyID:       colony.ID,
		colony:         colony,
		colonyPrvKey:   colonyPrvKey,
		executorID:     executor.ID,
		executor:       executor,
		executorPrvKey: executorPrvKey}

	return env, client, server, serverPrvKey, done
}

func areDirsSame(dir1, dir2 string) (bool, error) {
	var isSame = true

	err := filepath.Walk(dir1, func(path1 string, info1 os.FileInfo, err1 error) error {
		if err1 != nil {
			return err1
		}

		// Construct the corresponding path in dir2
		relPath, _ := filepath.Rel(dir1, path1)
		path2 := filepath.Join(dir2, relPath)

		cfsFile := filepath.Base(path2)

		if cfsFile != ".cfs" {
			info2, err2 := os.Stat(path2)
			if err2 != nil {
				if os.IsNotExist(err2) {
					isSame = false
					return fmt.Errorf("%s does not exist in %s", path1, dir2)
				}
				return err2
			}

			// If one is dir and the other is not
			if info1.IsDir() != info2.IsDir() {
				isSame = false
				return fmt.Errorf("%s and %s are not the same type", path1, path2)
			}

			// Compare file contents
			if !info1.IsDir() {
				content1, _ := ioutil.ReadFile(path1)
				content2, _ := ioutil.ReadFile(path2)
				if !bytes.Equal(content1, content2) {
					isSame = false
					return fmt.Errorf("content of %s and %s is not the same", path1, path2)
				}
			}
		}

		return nil
	})

	if !isSame || err != nil {
		return false, err
	}

	// Check the other way around to ensure dir2 doesn't have extra files/dirs
	err = filepath.Walk(dir2, func(path2 string, info2 os.FileInfo, err2 error) error {
		if err2 != nil {
			return err2
		}

		relPath, _ := filepath.Rel(dir2, path2)
		path1 := filepath.Join(dir1, relPath)

		_, err1 := os.Stat(path1)
		if err1 != nil && os.IsNotExist(err1) {
			isSame = false
			return fmt.Errorf("%s does not exist in %s", path2, dir1)
		}

		return nil
	})

	return isSame && err == nil, err
}
