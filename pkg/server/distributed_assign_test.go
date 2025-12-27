package server

import (
	"sync"
	"testing"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestDistributedAssign verifies that distributed assignment (ExclusiveAssign=false)
// never results in double assignments under contention. Multiple executors connecting
// to different servers in the cluster should each get unique processes.
func TestDistributedAssign(t *testing.T) {
	db, err := postgresql.PrepareTests()
	defer db.Close()
	assert.Nil(t, err)

	clusterSize := 3
	numProcesses := 50
	numExecutors := 10

	// Create a cluster with ExclusiveAssign=false (distributed assignment)
	runningCluster := StartClusterDistributed(t, db, clusterSize)
	assert.Len(t, runningCluster, clusterSize)

	WaitForCluster(t, runningCluster)
	log.Info("Cluster ready for distributed assign test")

	// Use first server to set up colony and executors
	setupClient := client.CreateColoniesClient("localhost", runningCluster[0].Node.APIPort, true, true)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = setupClient.AddColony(colony, runningCluster[0].ServerPrvKey)
	assert.Nil(t, err)

	// Create multiple executors
	executors := make([]*core.Executor, numExecutors)
	executorKeys := make([]string, numExecutors)
	for i := 0; i < numExecutors; i++ {
		executor, prvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
		assert.Nil(t, err)
		_, err = setupClient.AddExecutor(executor, colonyPrvKey)
		assert.Nil(t, err)
		err = setupClient.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
		assert.Nil(t, err)
		executors[i] = executor
		executorKeys[i] = prvKey
	}

	// Submit processes
	processIDs := make([]string, numProcesses)
	for i := 0; i < numProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(colony.Name)
		addedProcess, err := setupClient.Submit(funcSpec, executorKeys[0])
		assert.Nil(t, err)
		processIDs[i] = addedProcess.ID
	}

	log.WithFields(log.Fields{
		"numProcesses": numProcesses,
		"numExecutors": numExecutors,
		"clusterSize":  clusterSize,
	}).Info("Starting concurrent assignment test")

	// Track assigned processes to detect duplicates
	var mu sync.Mutex
	assignedProcesses := make(map[string]int) // processID -> executor index that got it
	var wg sync.WaitGroup

	// Each executor tries to assign from a different server in the cluster
	for i := 0; i < numExecutors; i++ {
		wg.Add(1)
		go func(executorIdx int) {
			defer wg.Done()

			// Connect to different servers in round-robin fashion
			serverIdx := executorIdx % clusterSize
			c := client.CreateColoniesClient("localhost", runningCluster[serverIdx].Node.APIPort, true, true)

			// Keep assigning until no more processes
			for {
				assignedProcess, err := c.Assign(colony.Name, 1, "", "", executorKeys[executorIdx])
				if err != nil {
					// No more processes or timeout
					break
				}

				mu.Lock()
				if prevExecutor, exists := assignedProcesses[assignedProcess.ID]; exists {
					// This should NEVER happen - double assignment detected!
					t.Errorf("DOUBLE ASSIGNMENT DETECTED! Process %s was assigned to executor %d but already assigned to executor %d",
						assignedProcess.ID, executorIdx, prevExecutor)
				}
				assignedProcesses[assignedProcess.ID] = executorIdx
				mu.Unlock()

				// Close the process so it doesn't block
				err = c.Close(assignedProcess.ID, executorKeys[executorIdx])
				if err != nil {
					log.WithFields(log.Fields{"ProcessID": assignedProcess.ID, "Error": err}).Warn("Failed to close process")
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify results
	mu.Lock()
	totalAssigned := len(assignedProcesses)
	mu.Unlock()

	log.WithFields(log.Fields{
		"totalAssigned": totalAssigned,
		"numProcesses":  numProcesses,
	}).Info("Distributed assignment test completed")

	// All processes should have been assigned exactly once
	assert.Equal(t, numProcesses, totalAssigned, "Not all processes were assigned")

	// Shutdown cluster
	for _, s := range runningCluster {
		s.Server.Shutdown()
	}
	for _, s := range runningCluster {
		<-s.Done
	}
}

// TestDistributedAssignHighContention tests with more executors than processes
// to maximize contention and verify no double assignments
func TestDistributedAssignHighContention(t *testing.T) {
	db, err := postgresql.PrepareTests()
	defer db.Close()
	assert.Nil(t, err)

	clusterSize := 3
	numProcesses := 10
	numExecutors := 30 // Many more executors than processes

	// Create a cluster with ExclusiveAssign=false (distributed assignment)
	runningCluster := StartClusterDistributed(t, db, clusterSize)
	assert.Len(t, runningCluster, clusterSize)

	WaitForCluster(t, runningCluster)
	log.Info("Cluster ready for high contention test")

	setupClient := client.CreateColoniesClient("localhost", runningCluster[0].Node.APIPort, true, true)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = setupClient.AddColony(colony, runningCluster[0].ServerPrvKey)
	assert.Nil(t, err)

	// Create many executors
	executorKeys := make([]string, numExecutors)
	for i := 0; i < numExecutors; i++ {
		executor, prvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
		assert.Nil(t, err)
		_, err = setupClient.AddExecutor(executor, colonyPrvKey)
		assert.Nil(t, err)
		err = setupClient.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
		assert.Nil(t, err)
		executorKeys[i] = prvKey
	}

	// Submit fewer processes than executors
	for i := 0; i < numProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(colony.Name)
		_, err := setupClient.Submit(funcSpec, executorKeys[0])
		assert.Nil(t, err)
	}

	log.WithFields(log.Fields{
		"numProcesses": numProcesses,
		"numExecutors": numExecutors,
	}).Info("Starting high contention assignment test")

	var mu sync.Mutex
	assignedProcesses := make(map[string]int)
	assignmentCounts := make([]int, numExecutors) // How many processes each executor got
	var wg sync.WaitGroup

	// All executors try to assign simultaneously
	for i := 0; i < numExecutors; i++ {
		wg.Add(1)
		go func(executorIdx int) {
			defer wg.Done()

			serverIdx := executorIdx % clusterSize
			c := client.CreateColoniesClient("localhost", runningCluster[serverIdx].Node.APIPort, true, true)

			// Try to assign with short timeout
			assignedProcess, err := c.Assign(colony.Name, 1, "", "", executorKeys[executorIdx])
			if err != nil {
				return // No process available
			}

			mu.Lock()
			if prevExecutor, exists := assignedProcesses[assignedProcess.ID]; exists {
				t.Errorf("DOUBLE ASSIGNMENT! Process %s assigned to executor %d, was already assigned to %d",
					assignedProcess.ID, executorIdx, prevExecutor)
			}
			assignedProcesses[assignedProcess.ID] = executorIdx
			assignmentCounts[executorIdx]++
			mu.Unlock()

			// Close process
			c.Close(assignedProcess.ID, executorKeys[executorIdx])
		}(i)
	}

	wg.Wait()

	mu.Lock()
	totalAssigned := len(assignedProcesses)
	mu.Unlock()

	log.WithFields(log.Fields{
		"totalAssigned": totalAssigned,
		"numProcesses":  numProcesses,
	}).Info("High contention test completed")

	assert.Equal(t, numProcesses, totalAssigned, "Not all processes were assigned")

	// Shutdown cluster
	for _, s := range runningCluster {
		s.Server.Shutdown()
	}
	for _, s := range runningCluster {
		<-s.Done
	}
}
