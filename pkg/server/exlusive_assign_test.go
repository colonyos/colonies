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
	host := testServer.controller.thisNode.Host
	apiPort := testServer.controller.thisNode.APIPort

	client := client.CreateColoniesClient(host, apiPort, true, true)

	colony1, colony1PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, runningCluster[testServerIndex].ServerPrvKey)
	assert.Nil(t, err)

	runtime1, runtime1PrvKey, err := utils.CreateTestRuntimeWithKey(colony1.ID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(runtime1, colony1PrvKey)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime1.ID, colony1PrvKey)
	assert.Nil(t, err)

	processSpec1 := utils.CreateTestProcessSpec(colony1.ID)
	addedProcess1, err := client.SubmitProcessSpec(processSpec1, runtime1PrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.AssignProcess(colony1.ID, -1, runtime1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, assignedProcess.ID, addedProcess1.ID)

	runningCluster[0].Server.Shutdown()
	runningCluster[1].Server.Shutdown()
	runningCluster[2].Server.Shutdown()

	for _, s := range runningCluster {
		<-s.Done
	}
}
