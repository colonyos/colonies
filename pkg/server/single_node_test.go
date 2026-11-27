package server

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// TestSingleNodeServer exercises the production code path used when
// clustering is disabled (--cluster=false): no etcd, no relay server, the
// local coordinator and a nil relay in the event handler. It runs regardless
// of COLONIES_TEST_CLUSTER so the single-node mode is always covered.
func TestSingleNodeServer(t *testing.T) {
	client, server, serverPrvKey, done := prepareTestsWithOptions(t, false, false)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	// The cluster is a singleton and the local node is the leader
	clusterConfig, err := client.GetClusterInfo(serverPrvKey)
	assert.Nil(t, err)
	assert.Len(t, clusterConfig.Nodes, 1)
	assert.Equal(t, clusterConfig.Nodes[0], clusterConfig.Leader)
	assert.True(t, server.controller.IsLeader())

	// Pause and resume go through the local coordinator
	err = client.PauseColonyAssignments(colony.Name, colonyPrvKey)
	assert.Nil(t, err)
	paused, err := client.AreColonyAssignmentsPaused(colony.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.True(t, paused)

	funcSpec := utils.CreateTestFunctionSpec(colony.Name)
	addedProcess, err := client.Submit(funcSpec, executorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(colony.Name, -1, "", "", executorPrvKey)
	assert.NotNil(t, err)

	err = client.ResumeColonyAssignments(colony.Name, colonyPrvKey)
	assert.Nil(t, err)
	paused, err = client.AreColonyAssignmentsPaused(colony.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.False(t, paused)

	// Realtime events are delivered without a relay server
	subscription, err := client.SubscribeProcess(colony.Name, addedProcess.ID, executor.Type, core.SUCCESS, 100, executorPrvKey)
	assert.Nil(t, err)
	waitForProcess := make(chan error)
	go func() {
		select {
		case <-subscription.ProcessChan:
			waitForProcess <- nil
		case err := <-subscription.ErrChan:
			waitForProcess <- err
		case <-time.After(10 * time.Second):
			waitForProcess <- assert.AnError
		}
	}()

	assignedProcess, err := client.Assign(colony.Name, -1, "", "", executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess.ID, assignedProcess.ID)

	err = client.Close(assignedProcess.ID, executorPrvKey)
	assert.Nil(t, err)
	assert.Nil(t, <-waitForProcess)

	// Stopping must not touch etcd or relay resources
	server.Shutdown()
	<-done
}
