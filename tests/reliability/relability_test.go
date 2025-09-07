package reliability

import (
	"testing"

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
	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = c.AddExecutor(executor, colonyPrvKey)
	err = c.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	_, err = c.GetColonyByName(colony.Name, executorPrvKey)
	assert.Nil(t, err)

	// Now kill server 1
	runningCluster[0].Server.Shutdown()
	server.WaitForServerToDie(t, runningCluster[0])
	log.Info("Server ", selectServerIndex, " is dead now")

	_, err = c.GetColonyByName(colony.Name, executorPrvKey)
	assert.NotNil(t, err) // Will not work, server is DEAD

	// Connect to another server in the cluster and try again
	selectServerIndex = 1
	selectedServer = runningCluster[selectServerIndex]
	c = client.CreateColoniesClient("localhost", selectedServer.Node.APIPort, true, true)

	_, err = c.GetColonyByName(colony.Name, executorPrvKey)
	assert.Nil(t, err) // Should work

	// Kill the remaining servers, this will also end the test
	runningCluster[1].Server.Shutdown()
	runningCluster[2].Server.Shutdown()

	for _, s := range runningCluster {
		<-s.Done
	}
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
	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = c.AddExecutor(executor, colonyPrvKey)
	err = c.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)

	// Start a generator
	generator := utils.FakeGenerator(t, colony.Name, executor.ID, executor.Name)
	generator.Trigger = 1
	addedGenerator, err := c.AddGenerator(generator, executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	// Test that the generator works, we need to wait 1 second
	var graphs []*core.ProcessGraph
	graphs, err = c.GetWaitingProcessGraphs(colony.Name, 100, executorPrvKey)
	assert.Len(t, graphs, 0) // Since we have not triggered any generator yet

	err = c.PackGenerator(addedGenerator.ID, "arg", executorPrvKey)
	assert.Nil(t, err)

	nrOfgraphs := server.WaitForProcessGraphs(t, c, colony.Name, addedGenerator.Name, executorPrvKey, 1)
	assert.Equal(t, nrOfgraphs, 1) // Ok we got a generator

	// The leader is reponsible for the generator engine
	// Find out who the leader is, and then kill it
	clusterInfo, err := c.GetClusterInfo(selectedServer.ServerPrvKey)
	assert.Nil(t, err)
	leaderName := clusterInfo.Leader.Name

	// Ok, now we now who name of the leader, find out which server that is
	var leaderS server.ServerInfo
	for _, s := range runningCluster {
		if s.Node.Name == leaderName {
			leaderS = s
		}
	}
	log.Info("ColoniesServer Leader is ", leaderS.Node.Name, " kill it")

	// Now kill leader server
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

	c.PackGenerator(addedGenerator.ID, "arg", executorPrvKey)

	nrOfgraphs = server.WaitForProcessGraphs(t, c, colony.Name, addedGenerator.Name, executorPrvKey, 2)
	log.WithFields(log.Fields{"nrOfgraphs": nrOfgraphs}).Info("Done waiting for processgraphs")
	assert.Equal(t, nrOfgraphs, 2)

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

func TestCronReliability(t *testing.T) {
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
	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = c.AddExecutor(executor, colonyPrvKey)
	err = c.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)

	// Start a cron
	cron := utils.FakeCron(t, colony.Name, executor.ID, executor.Name)
	cron.CronExpression = "0/1 * * * * *" // every second
	addedCron, err := c.AddCron(cron, executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	// Test that the cron works, we need to wait 1 second
	var graphs []*core.ProcessGraph
	graphs, err = c.GetWaitingProcessGraphs(colony.Name, 100, executorPrvKey)
	assert.Len(t, graphs, 0) // Since we have not triggered any cron yet

	nrOfgraphs := server.WaitForProcessGraphs(t, c, colony.Name, "", executorPrvKey, 1)
	assert.Equal(t, nrOfgraphs, 1) // Ok we got a cron

	// The leader is reponsible for the generator engine
	// Find out who the leader is, and then kill it
	clusterInfo, err := c.GetClusterInfo(selectedServer.ServerPrvKey)
	assert.Nil(t, err)
	leaderName := clusterInfo.Leader.Name

	// Ok, now we now who name of the leader, find out which server that is
	var leaderS server.ServerInfo
	for _, s := range runningCluster {
		if s.Node.Name == leaderName {
			leaderS = s
		}
	}
	log.Info("ColoniesServer Leader is ", leaderS.Node.Name, " kill it")

	// Now kill leader server
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

	nrOfgraphs2 := server.WaitForProcessGraphs(t, c, colony.Name, "", executorPrvKey, 2)
	log.WithFields(log.Fields{"nrOfgraphs": nrOfgraphs, "nrOfgraphs2": nrOfgraphs2}).Info("Done waiting for processgraphs")
	assert.Equal(t, nrOfgraphs2, 2)

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
