package server

import (
	"fmt"
	"testing"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestExclusiveAssign(t *testing.T) {
	db, err := postgresql.PrepareTests()
	defer db.Close()
	assert.Nil(t, err)

	clusterSize := 3

	// Create a cluster
	runningCluster := StartCluster(t, db, clusterSize)
	assert.Len(t, runningCluster, clusterSize)

	WaitForCluster(t, runningCluster)
	log.Info("Cluster ready")

	isServer0Leader := runningCluster[0].Server.controller.isLeader()
	isServer1Leader := runningCluster[1].Server.controller.isLeader()
	isServer2Leader := runningCluster[2].Server.controller.isLeader()

	testServerIndex := 0
	if isServer0Leader {
		fmt.Println("server 0 is leader")
		testServerIndex = 1
	} else if isServer1Leader {
		fmt.Println("server 1 is leader")
		testServerIndex = 2
	} else if isServer2Leader {
		fmt.Println("server 1 is leader")
		testServerIndex = 0
	}

	testServer := runningCluster[testServerIndex].Server
	host := testServer.controller.getThisNode().Host
	apiPort := testServer.controller.getThisNode().APIPort

	client := client.CreateColoniesClient(host, apiPort, true, true)

	colony1, colony1PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, runningCluster[testServerIndex].ServerPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colony1PrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(colony1.Name, executor1.Name, colony1PrvKey)
	assert.Nil(t, err)

	funcSpec1 := utils.CreateTestFunctionSpec(colony1.Name)
	addedProcess1, err := client.Submit(funcSpec1, executor1PrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(colony1.Name, -1, executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, assignedProcess.ID, addedProcess1.ID)

	runningCluster[0].Server.Shutdown()
	runningCluster[1].Server.Shutdown()
	runningCluster[2].Server.Shutdown()

	for _, s := range runningCluster {
		<-s.Done
	}
}
