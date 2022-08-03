package reliability

import (
	"testing"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestReliability(t *testing.T) {
	db, err := postgresql.PrepareTests()
	defer db.Close()
	assert.Nil(t, err)

	clusterSize := 3
	selectServerIndex := 0

	runningCluster := server.StartCluster(t, db, clusterSize)
	assert.Len(t, runningCluster, clusterSize)

	server.WaitForCluster(t, runningCluster)
	log.Info("Cluster ready")

	selectedServer := runningCluster[selectServerIndex]
	c := client.CreateColoniesClient("localhost", selectedServer.Node.APIPort, true, true)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	colony, err = c.AddColony(colony, selectedServer.ServerPrvKey)
	assert.Nil(t, err)
	runtime, runtimePrvKey, err := utils.CreateTestRuntimeWithKey(colony.ID)
	assert.Nil(t, err)
	_, err = c.AddRuntime(runtime, colonyPrvKey)
	err = c.ApproveRuntime(runtime.ID, colonyPrvKey)
	_, err = c.GetColonyByID(colony.ID, runtimePrvKey)
	assert.Nil(t, err)

	// Now kill server 1
	runningCluster[0].Server.Shutdown()
	server.WaitForServerToDie(t, runningCluster[0])
	log.Info("Server ", selectServerIndex, " is dead now")

	_, err = c.GetColonyByID(colony.ID, runtimePrvKey)
	assert.NotNil(t, err) // Will not work, server is DEAD

	// Connect to another server in the cluster and try again
	selectServerIndex = 1
	selectedServer = runningCluster[selectServerIndex]
	c = client.CreateColoniesClient("localhost", selectedServer.Node.APIPort, true, true)

	_, err = c.GetColonyByID(colony.ID, runtimePrvKey)
	assert.Nil(t, err) // Should work

	// Kill the remaining nodes, this will also end the test
	runningCluster[1].Server.Shutdown()
	runningCluster[2].Server.Shutdown()

	for _, s := range runningCluster {
		<-s.Done
	}
}
