package cluster

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateEtcdCluster(t *testing.T) {
	node1 := Node{Name: "etcd1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := Node{Name: "etcd2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}
	node3 := Node{Name: "etcd3", Host: "localhost", EtcdClientPort: 24300, EtcdPeerPort: 23300, RelayPort: 25300, APIPort: 26300}
	node4 := Node{Name: "etcd4", Host: "localhost", EtcdClientPort: 24400, EtcdPeerPort: 23400, RelayPort: 25400, APIPort: 26400}

	config := Config{}
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)
	config.AddNode(node4)

	server1 := CreateEtcdServer(node1, config, ".")
	server2 := CreateEtcdServer(node2, config, ".")
	server3 := CreateEtcdServer(node3, config, ".")
	server4 := CreateEtcdServer(node4, config, ".")

	assert.Equal(t, server1.buildInitialClusterStr(), "etcd1=http://localhost:23100,etcd2=http://localhost:23200,etcd3=http://localhost:23300,etcd4=http://localhost:23400")

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
	node := Node{Name: "etcd1", Host: "localhost", EtcdClientPort: 24500, EtcdPeerPort: 23500, RelayPort: 25500, APIPort: 26500}
	config := Config{}
	config.AddNode(node)

	server := CreateEtcdServer(node, config, ".")
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
	node := Node{Name: "etcd2", Host: "localhost", EtcdClientPort: 24600, EtcdPeerPort: 23600, RelayPort: 25600, APIPort: 26600}
	config := Config{}
	config.AddNode(node)

	server := CreateEtcdServer(node, config, ".")
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
	node1 := Node{Name: "etcd1", Host: "localhost", EtcdClientPort: 24700, EtcdPeerPort: 23700, RelayPort: 25700, APIPort: 26700}
	node2 := Node{Name: "etcd2", Host: "localhost", EtcdClientPort: 24800, EtcdPeerPort: 23800, RelayPort: 25800, APIPort: 26800}

	config := Config{}
	config.AddNode(node1)
	config.AddNode(node2)

	server1 := CreateEtcdServer(node1, config, ".")
	server2 := CreateEtcdServer(node2, config, ".")

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
