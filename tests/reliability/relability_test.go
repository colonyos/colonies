package reliability

import (
	"fmt"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
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

	// Create a cluster
	runningCluster := server.StartCluster(t, db, clusterSize)
	assert.Len(t, runningCluster, clusterSize)

	server.WaitForCluster(t, runningCluster)
	log.Info("Cluster ready")

	// Create a client connected to one of the servers
	selectedServer := runningCluster[selectServerIndex]
	c := client.CreateColoniesClient("localhost", selectedServer.Node.APIPort, true, true)

	// Setup a test environment
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

	// Kill the remaining servers, this will also end the test
	runningCluster[1].Server.Shutdown()
	runningCluster[2].Server.Shutdown()

	for _, s := range runningCluster {
		<-s.Done
	}
}

func waitForProcessGraphs(t *testing.T, c *client.ColoniesClient, colonyID string, generatorID string, runtimePrvKey string, threshold int) int {
	var graphs []*core.ProcessGraph
	var err error
	retries := 40
	for i := 0; i < retries; i++ {
		graphs, err = c.GetWaitingProcessGraphs(colonyID, 100, runtimePrvKey)
		assert.Nil(t, err)
		err = c.IncGenerator(generatorID, runtimePrvKey)
		if len(graphs) > threshold {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return len(graphs)
}

func TestGeneratorReliability(t *testing.T) {
	db, err := postgresql.PrepareTests()
	defer db.Close()
	assert.Nil(t, err)

	clusterSize := 3
	selectServerIndex := 0

	// Create a cluster
	runningCluster := server.StartCluster(t, db, clusterSize)
	assert.Len(t, runningCluster, clusterSize)

	server.WaitForCluster(t, runningCluster)
	log.Info("Cluster ready")

	// Create a client connected to one of the servers
	selectedServer := runningCluster[selectServerIndex]
	c := client.CreateColoniesClient("localhost", selectedServer.Node.APIPort, true, true)

	// Setup a test environment
	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	colony, err = c.AddColony(colony, selectedServer.ServerPrvKey)
	assert.Nil(t, err)
	runtime, runtimePrvKey, err := utils.CreateTestRuntimeWithKey(colony.ID)
	assert.Nil(t, err)
	_, err = c.AddRuntime(runtime, colonyPrvKey)
	err = c.ApproveRuntime(runtime.ID, colonyPrvKey)

	// Start a generator
	generator := utils.FakeGenerator(t, colony.ID)
	generator.Timeout = 1    // Every 1 seconds
	generator.Trigger = 1000 // A very high number since we want to trigger generators only by time
	addedGenerator, err := c.AddGenerator(generator, runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	// Test that the generator works, we need to wait 1 second
	var graphs []*core.ProcessGraph
	graphs, err = c.GetWaitingProcessGraphs(colony.ID, 100, runtimePrvKey)
	assert.Len(t, graphs, 0) // Since we have not triggered any generator yet

	nrOfgraphs := waitForProcessGraphs(t, c, colony.ID, generator.ID, runtimePrvKey, 1)
	assert.Greater(t, nrOfgraphs, 1) // Ok we got a generator

	// The leader is reponsible for the generator engine
	// Find out who the leader is, and then kill it
	clusterInfo, err := c.GetClusterInfo(selectedServer.ServerPrvKey)
	assert.Nil(t, err)
	leaderName := clusterInfo.Leader.Name
	fmt.Println("leader is", leaderName)

	// Ok, now we now who name of the leader, find out which server that is
	var leaderS server.ServerInfo
	for _, s := range runningCluster {
		if s.Node.Name == leaderName {
			leaderS = s
		}
	}
	log.Info("ColoniesServer Leader is ", leaderS.Node.Name, " kill it")

	// Now kill server 1
	leaderS.Server.Shutdown()
	server.WaitForServerToDie(t, leaderS)
	log.Info("ColoniesServer Leader is ", leaderS.Node.Name, " is dead")

	// The problem now is that our client might be connected to that ColoniesServer
	// In that case, we need to connect to another server
	err = c.CheckHealth()
	if err != nil {
		selectServerIndex = 1
		selectedServer := runningCluster[selectServerIndex]
		c = client.CreateColoniesClient("localhost", selectedServer.Node.APIPort, true, true) // Connect to another server
	}

	nrOfgraphs2 := waitForProcessGraphs(t, c, colony.ID, generator.ID, runtimePrvKey, nrOfgraphs)
	log.WithFields(log.Fields{"nrOfgraphs": nrOfgraphs, "nrOfgraphs2": nrOfgraphs2}).Info("Done waiting for processgraphs")
	assert.Greater(t, nrOfgraphs2, nrOfgraphs)

	// Kill the remaining servers, this will also end the test
	for _, s := range runningCluster {
		if s.Node.Name != leaderS.Node.Name {
			s.Server.Shutdown()
		}
	}

	for _, s := range runningCluster {
		<-s.Done
	}
}
