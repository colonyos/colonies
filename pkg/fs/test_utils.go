package fs

import (
	"io/ioutil"
	"math/rand"
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

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.ID)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(executor.ID, colonyPrvKey)
	assert.Nil(t, err)

	env := &testEnv{colonyID: colony.ID,
		colony:         colony,
		colonyPrvKey:   colonyPrvKey,
		executorID:     executor.ID,
		executor:       executor,
		executorPrvKey: executorPrvKey}

	return env, client, server, serverPrvKey, done
}
