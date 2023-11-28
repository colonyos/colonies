package server

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRetention(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
	//log.SetLevel(log.DebugLevel)

	client, server, serverPrvKey, done := prepareTestsWithRetention(t, true)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(executor.ID, colonyPrvKey)
	assert.Nil(t, err)

	wf := generateSingleWorkflowSpec(colony.Name)
	_, err = client.SubmitWorkflowSpec(wf, executorPrvKey)
	assert.Nil(t, err)

	process, err := client.Assign(colony.Name, -1, executorPrvKey)
	assert.Nil(t, err)

	err = client.Close(process.ID, executorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(colony.Name, executorPrvKey)
	assert.Equal(t, stat.SuccessfulWorkflows, 1)
	assert.Equal(t, stat.SuccessfulProcesses, 1)

	time.Sleep(2 * time.Second)

	stat, err = client.ColonyStatistics(colony.Name, executorPrvKey)
	assert.Equal(t, stat.SuccessfulWorkflows, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 0)

	server.Shutdown()
	<-done
}
