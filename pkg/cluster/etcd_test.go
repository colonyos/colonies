package cluster

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateEtcdCluster(t *testing.T) {
	nodes := testClusterNodes(t, "etcd1", "etcd2", "etcd3", "etcd4")
	node1, node2, node3, node4 := nodes[0], nodes[1], nodes[2], nodes[3]

	config := Config{}
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)
	config.AddNode(node4)

	dataPath := t.TempDir()
	server1 := CreateEtcdServer(node1, config, dataPath)
	server2 := CreateEtcdServer(node2, config, dataPath)
	server3 := CreateEtcdServer(node3, config, dataPath)
	server4 := CreateEtcdServer(node4, config, dataPath)

	expectedClusterStr := fmt.Sprintf("etcd1=http://localhost:%d,etcd2=http://localhost:%d,etcd3=http://localhost:%d,etcd4=http://localhost:%d",
		node1.EtcdPeerPort, node2.EtcdPeerPort, node3.EtcdPeerPort, node4.EtcdPeerPort)
	assert.Equal(t, server1.buildInitialClusterStr(), expectedClusterStr)

	server1.Start()
	server2.Start()
	server3.Start()
	server4.Start()

	server1.WaitToStart()
	server2.WaitToStart()
	server3.WaitToStart()
	server4.WaitToStart()

	leader := server1.Leader()
	assert.Equal(t, server2.Leader(), leader)
	assert.Equal(t, server3.Leader(), leader)
	assert.Equal(t, server4.Leader(), leader)

	nodes1 := server1.Members()
	assert.Len(t, nodes1, 4)
	nodes2 := server1.Members()
	assert.Len(t, nodes2, 4)
	nodes3 := server1.Members()
	assert.Len(t, nodes3, 4)
	nodes4 := server1.Members()
	assert.Len(t, nodes4, 4)

	currentCluster := server4.CurrentCluster()
	assert.Len(t, currentCluster.Nodes, 4)

	server1.Stop()
	server2.Stop()
	server3.Stop()
	server4.Stop()

	server1.WaitToStop()
	server2.WaitToStop()
	server3.WaitToStop()
	server4.WaitToStop()

	os.RemoveAll(server1.StorageDir())
	os.RemoveAll(server2.StorageDir())
	os.RemoveAll(server3.StorageDir())
	os.RemoveAll(server4.StorageDir())
}

func TestEtcdAssignmentsPauseResume(t *testing.T) {
	node := testClusterNodes(t, "etcd1")[0]
	config := Config{}
	config.AddNode(node)

	server := CreateEtcdServer(node, config, t.TempDir())
	server.Start()
	server.WaitToStart()

	colonyName := "test_colony"

	// Test initial state - assignments should not be paused
	paused, err := server.AreColonyAssignmentsPaused(colonyName)
	assert.NoError(t, err)
	assert.False(t, paused, "Colony assignments should not be paused initially")

	// Test pause assignments
	err = server.PauseColonyAssignments(colonyName)
	assert.NoError(t, err)

	// Verify assignments are paused
	paused, err = server.AreColonyAssignmentsPaused(colonyName)
	assert.NoError(t, err)
	assert.True(t, paused, "Colony assignments should be paused after calling PauseColonyAssignments")

	// Test resume assignments
	err = server.ResumeColonyAssignments(colonyName)
	assert.NoError(t, err)

	// Verify assignments are resumed
	paused, err = server.AreColonyAssignmentsPaused(colonyName)
	assert.NoError(t, err)
	assert.False(t, paused, "Assignments should not be paused after calling ResumeAssignments")

	// Test multiple pause/resume cycles
	err = server.PauseColonyAssignments(colonyName)
	assert.NoError(t, err)
	paused, err = server.AreColonyAssignmentsPaused(colonyName)
	assert.NoError(t, err)
	assert.True(t, paused)

	err = server.ResumeColonyAssignments(colonyName)
	assert.NoError(t, err)
	paused, err = server.AreColonyAssignmentsPaused(colonyName)
	assert.NoError(t, err)
	assert.False(t, paused)

	// Cleanup
	server.Stop()
	server.WaitToStop()
	os.RemoveAll(server.StorageDir())
}

func TestEtcdAssignmentsPauseResumeWithoutClient(t *testing.T) {
	node := testClusterNodes(t, "etcd2")[0]
	config := Config{}
	config.AddNode(node)

	server := CreateEtcdServer(node, config, t.TempDir())
	colonyName := "test_colony"

	// Test methods fail when etcd client is not initialized
	paused, err := server.AreColonyAssignmentsPaused(colonyName)
	assert.Error(t, err)
	assert.False(t, paused)
	assert.Contains(t, err.Error(), "etcd client is not initialized")

	err = server.PauseColonyAssignments(colonyName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "etcd client is not initialized")

	err = server.ResumeColonyAssignments(colonyName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "etcd client is not initialized")
}

func TestEtcdAssignmentsPauseResumeMultiNode(t *testing.T) {
	nodes := testClusterNodes(t, "etcd1", "etcd2")
	node1, node2 := nodes[0], nodes[1]

	config := Config{}
	config.AddNode(node1)
	config.AddNode(node2)

	dataPath := t.TempDir()
	server1 := CreateEtcdServer(node1, config, dataPath)
	server2 := CreateEtcdServer(node2, config, dataPath)

	server1.Start()
	server2.Start()
	server1.WaitToStart()
	server2.WaitToStart()

	colonyName := "test_colony"

	// Test initial state on both nodes - assignments should not be paused
	paused1, err := server1.AreColonyAssignmentsPaused(colonyName)
	assert.NoError(t, err)
	assert.False(t, paused1, "Colony assignments should not be paused initially on node1")

	paused2, err := server2.AreColonyAssignmentsPaused(colonyName)
	assert.NoError(t, err)
	assert.False(t, paused2, "Colony assignments should not be paused initially on node2")

	// Pause assignments on server1
	err = server1.PauseColonyAssignments(colonyName)
	assert.NoError(t, err)

	// Verify pause state is visible on both nodes
	paused1, err = server1.AreColonyAssignmentsPaused(colonyName)
	assert.NoError(t, err)
	assert.True(t, paused1, "Colony assignments should be paused on node1")

	paused2, err = server2.AreColonyAssignmentsPaused(colonyName)
	assert.NoError(t, err)
	assert.True(t, paused2, "Colony assignments should be paused on node2")

	// Resume assignments on server2 (different node)
	err = server2.ResumeColonyAssignments(colonyName)
	assert.NoError(t, err)

	// Verify resume state is visible on both nodes
	paused1, err = server1.AreColonyAssignmentsPaused(colonyName)
	assert.NoError(t, err)
	assert.False(t, paused1, "Colony assignments should not be paused on node1")

	paused2, err = server2.AreColonyAssignmentsPaused(colonyName)
	assert.NoError(t, err)
	assert.False(t, paused2, "Assignments should not be paused on node2")

	// Cleanup
	server1.Stop()
	server2.Stop()
	server1.WaitToStop()
	server2.WaitToStop()
	os.RemoveAll(server1.StorageDir())
	os.RemoveAll(server2.StorageDir())
}
