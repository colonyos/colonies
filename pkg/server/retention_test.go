package server

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
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

	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	wf := generateSingleWorkflowSpec(colony.Name)
	_, err = client.SubmitWorkflowSpec(wf, executorPrvKey)
	assert.Nil(t, err)

	process, err := client.Assign(colony.Name, -1, "", "", executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)

	err = client.Close(process.ID, executorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(colony.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, stat.SuccessfulWorkflows)
	assert.Equal(t, 1, stat.SuccessfulProcesses)

	// The retention worker deletes the process and workflow once they are
	// older than the policy; poll instead of sleeping a fixed time so the
	// test tolerates a slow leader election or database under load
	assert.Eventually(t, func() bool {
		stat, err := client.ColonyStatistics(colony.Name, executorPrvKey)
		if err != nil {
			return false
		}
		return stat.SuccessfulWorkflows == 0 && stat.SuccessfulProcesses == 0
	}, 4*testRetentionPolicySeconds*time.Second, 250*time.Millisecond)

	server.Shutdown()
	<-done
}

// generateSingleWorkflowSpec creates a simple workflow spec with a single task for testing
func generateSingleWorkflowSpec(colonyName string) *core.WorkflowSpec {
	workflowSpec := core.CreateWorkflowSpec(colonyName)

	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.NodeName = "task1"
	funcSpec.Conditions.ColonyName = colonyName
	funcSpec.Conditions.ExecutorType = "test_executor_type"

	workflowSpec.AddFunctionSpec(funcSpec)

	return workflowSpec
}
