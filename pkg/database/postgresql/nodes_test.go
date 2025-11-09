package postgresql

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestNodeClosedDB(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	db.Close()

	node := core.CreateNode("test-node", "test-colony", "test-location")
	err = db.AddNode(node)
	assert.NotNil(t, err)

	_, err = db.GetNodes("")
	assert.NotNil(t, err)

	_, err = db.GetNodeByID("invalid_id")
	assert.NotNil(t, err)

	_, err = db.GetNodeByName("invalid_colony", "invalid_name")
	assert.NotNil(t, err)

	_, err = db.GetNodesByLocation("invalid_colony", "invalid_location")
	assert.NotNil(t, err)

	err = db.UpdateNode(node)
	assert.NotNil(t, err)

	err = db.RemoveNodeByID("invalid_id")
	assert.NotNil(t, err)

	err = db.RemoveNodeByName("invalid_colony", "invalid_name")
	assert.NotNil(t, err)

	err = db.RemoveNodesByColonyName("invalid_colony")
	assert.NotNil(t, err)

	_, err = db.CountNodes("")
	assert.NotNil(t, err)
}

func TestAddNode(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	node := core.CreateNode("test-node", colony.Name, "test-location")
	node.Platform = "linux"
	node.Architecture = "amd64"
	node.CPU = 32
	node.Memory = 64000
	node.GPU = 2
	node.Capabilities = []string{"docker", "gpu"}
	node.Labels["gpu.0.name"] = "NVIDIA RTX 3080"
	node.Labels["region"] = "us-west"

	err = db.AddNode(node)
	assert.Nil(t, err)

	nodes, err := db.GetNodes(colony.Name)
	assert.Nil(t, err)
	assert.Len(t, nodes, 1)

	nodeFromDB := nodes[0]
	assert.True(t, node.Equals(nodeFromDB))
	assert.Equal(t, "test-node", nodeFromDB.Name)
	assert.Equal(t, colony.Name, nodeFromDB.ColonyName)
	assert.Equal(t, "test-location", nodeFromDB.Location)
	assert.Equal(t, "linux", nodeFromDB.Platform)
	assert.Equal(t, "amd64", nodeFromDB.Architecture)
	assert.Equal(t, 32, nodeFromDB.CPU)
	assert.Equal(t, int64(64000), nodeFromDB.Memory)
	assert.Equal(t, 2, nodeFromDB.GPU)
	assert.Equal(t, 2, len(nodeFromDB.Capabilities))
	assert.Equal(t, "docker", nodeFromDB.Capabilities[0])
	assert.Equal(t, "gpu", nodeFromDB.Capabilities[1])
	assert.Equal(t, "NVIDIA RTX 3080", nodeFromDB.Labels["gpu.0.name"])
	assert.Equal(t, "us-west", nodeFromDB.Labels["region"])
	assert.Equal(t, "ready", nodeFromDB.State)
}

func TestAddNodeNil(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	err = db.AddNode(nil)
	assert.NotNil(t, err)
	assert.Equal(t, "Node is nil", err.Error())
}

func TestAddMultipleNodes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	node1 := core.CreateNode("node1", colony.Name, "location1")
	node1.Platform = "linux"
	err = db.AddNode(node1)
	assert.Nil(t, err)

	node2 := core.CreateNode("node2", colony.Name, "location2")
	node2.Platform = "darwin"
	err = db.AddNode(node2)
	assert.Nil(t, err)

	node3 := core.CreateNode("node3", colony.Name, "location1")
	node3.Platform = "windows"
	err = db.AddNode(node3)
	assert.Nil(t, err)

	nodes, err := db.GetNodes(colony.Name)
	assert.Nil(t, err)
	assert.Len(t, nodes, 3)
}

func TestGetNodeByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	node1 := core.CreateNode("node1", colony.Name, "location1")
	err = db.AddNode(node1)
	assert.Nil(t, err)

	node2 := core.CreateNode("node2", colony.Name, "location2")
	err = db.AddNode(node2)
	assert.Nil(t, err)

	nodeFromDB, err := db.GetNodeByID("invalid_id")
	assert.Nil(t, err)
	assert.Nil(t, nodeFromDB)

	nodeFromDB, err = db.GetNodeByID(node1.ID)
	assert.Nil(t, err)
	assert.NotNil(t, nodeFromDB)
	assert.True(t, node1.Equals(nodeFromDB))
}

func TestGetNodeByName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	node1 := core.CreateNode("node1", colony.Name, "location1")
	err = db.AddNode(node1)
	assert.Nil(t, err)

	node2 := core.CreateNode("node2", colony.Name, "location2")
	err = db.AddNode(node2)
	assert.Nil(t, err)

	nodeFromDB, err := db.GetNodeByName("invalid_colony", "node1")
	assert.Nil(t, err)
	assert.Nil(t, nodeFromDB)

	nodeFromDB, err = db.GetNodeByName(colony.Name, "invalid_name")
	assert.Nil(t, err)
	assert.Nil(t, nodeFromDB)

	nodeFromDB, err = db.GetNodeByName(colony.Name, "node1")
	assert.Nil(t, err)
	assert.NotNil(t, nodeFromDB)
	assert.True(t, node1.Equals(nodeFromDB))
}

func TestGetNodes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	node1 := core.CreateNode("node1", colony1.Name, "location1")
	err = db.AddNode(node1)
	assert.Nil(t, err)

	node2 := core.CreateNode("node2", colony1.Name, "location2")
	err = db.AddNode(node2)
	assert.Nil(t, err)

	node3 := core.CreateNode("node3", colony2.Name, "location3")
	err = db.AddNode(node3)
	assert.Nil(t, err)

	// Get all nodes
	allNodes, err := db.GetNodes("")
	assert.Nil(t, err)
	assert.Len(t, allNodes, 3)

	// Get nodes for colony1
	colony1Nodes, err := db.GetNodes(colony1.Name)
	assert.Nil(t, err)
	assert.Len(t, colony1Nodes, 2)

	// Get nodes for colony2
	colony2Nodes, err := db.GetNodes(colony2.Name)
	assert.Nil(t, err)
	assert.Len(t, colony2Nodes, 1)

	// Get nodes for non-existent colony
	noNodes, err := db.GetNodes("non_existent_colony")
	assert.Nil(t, err)
	assert.Len(t, noNodes, 0)
}

func TestGetNodesByLocation(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	node1 := core.CreateNode("node1", colony.Name, "us-west")
	err = db.AddNode(node1)
	assert.Nil(t, err)

	node2 := core.CreateNode("node2", colony.Name, "us-west")
	err = db.AddNode(node2)
	assert.Nil(t, err)

	node3 := core.CreateNode("node3", colony.Name, "eu-north")
	err = db.AddNode(node3)
	assert.Nil(t, err)

	// Get nodes in us-west
	usWestNodes, err := db.GetNodesByLocation(colony.Name, "us-west")
	assert.Nil(t, err)
	assert.Len(t, usWestNodes, 2)

	// Get nodes in eu-north
	euNorthNodes, err := db.GetNodesByLocation(colony.Name, "eu-north")
	assert.Nil(t, err)
	assert.Len(t, euNorthNodes, 1)

	// Get nodes for non-existent location
	noNodes, err := db.GetNodesByLocation(colony.Name, "asia-east")
	assert.Nil(t, err)
	assert.Len(t, noNodes, 0)
}

func TestUpdateNode(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	node := core.CreateNode("test-node", colony.Name, "original-location")
	node.Platform = "linux"
	node.CPU = 16
	err = db.AddNode(node)
	assert.Nil(t, err)

	// Update node metadata
	node.Location = "new-location"
	node.Platform = "darwin"
	node.Architecture = "arm64"
	node.CPU = 32
	node.Memory = 128000
	node.GPU = 4
	node.Capabilities = []string{"docker", "kubernetes"}
	node.Labels["new-label"] = "new-value"
	node.State = "offline"

	// Sleep to ensure LastSeen is different
	time.Sleep(10 * time.Millisecond)
	node.TouchLastSeen()

	err = db.UpdateNode(node)
	assert.Nil(t, err)

	nodeFromDB, err := db.GetNodeByID(node.ID)
	assert.Nil(t, err)
	assert.NotNil(t, nodeFromDB)
	assert.Equal(t, "new-location", nodeFromDB.Location)
	assert.Equal(t, "darwin", nodeFromDB.Platform)
	assert.Equal(t, "arm64", nodeFromDB.Architecture)
	assert.Equal(t, 32, nodeFromDB.CPU)
	assert.Equal(t, int64(128000), nodeFromDB.Memory)
	assert.Equal(t, 4, nodeFromDB.GPU)
	assert.Equal(t, 2, len(nodeFromDB.Capabilities))
	assert.Equal(t, "new-value", nodeFromDB.Labels["new-label"])
	assert.Equal(t, "offline", nodeFromDB.State)
	assert.True(t, nodeFromDB.LastSeen.After(node.Created))
}

func TestUpdateNodeNil(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	err = db.UpdateNode(nil)
	assert.NotNil(t, err)
	assert.Equal(t, "Node is nil", err.Error())
}

func TestRemoveNodeByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	node1 := core.CreateNode("node1", colony.Name, "location1")
	err = db.AddNode(node1)
	assert.Nil(t, err)

	node2 := core.CreateNode("node2", colony.Name, "location2")
	err = db.AddNode(node2)
	assert.Nil(t, err)

	// Verify both nodes exist
	nodes, err := db.GetNodes(colony.Name)
	assert.Nil(t, err)
	assert.Len(t, nodes, 2)

	// Remove node1
	err = db.RemoveNodeByID(node1.ID)
	assert.Nil(t, err)

	// Verify node1 is gone
	nodeFromDB, err := db.GetNodeByID(node1.ID)
	assert.Nil(t, err)
	assert.Nil(t, nodeFromDB)

	// Verify node2 still exists
	nodes, err = db.GetNodes(colony.Name)
	assert.Nil(t, err)
	assert.Len(t, nodes, 1)
	assert.Equal(t, node2.ID, nodes[0].ID)
}

func TestRemoveNodeByName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	node1 := core.CreateNode("node1", colony.Name, "location1")
	err = db.AddNode(node1)
	assert.Nil(t, err)

	node2 := core.CreateNode("node2", colony.Name, "location2")
	err = db.AddNode(node2)
	assert.Nil(t, err)

	// Remove node1 by name
	err = db.RemoveNodeByName(colony.Name, "node1")
	assert.Nil(t, err)

	// Verify node1 is gone
	nodeFromDB, err := db.GetNodeByName(colony.Name, "node1")
	assert.Nil(t, err)
	assert.Nil(t, nodeFromDB)

	// Verify node2 still exists
	nodes, err := db.GetNodes(colony.Name)
	assert.Nil(t, err)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "node2", nodes[0].Name)
}

func TestRemoveNodesByColonyName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	node1 := core.CreateNode("node1", colony1.Name, "location1")
	err = db.AddNode(node1)
	assert.Nil(t, err)

	node2 := core.CreateNode("node2", colony1.Name, "location2")
	err = db.AddNode(node2)
	assert.Nil(t, err)

	node3 := core.CreateNode("node3", colony2.Name, "location3")
	err = db.AddNode(node3)
	assert.Nil(t, err)

	// Verify all nodes exist
	allNodes, err := db.GetNodes("")
	assert.Nil(t, err)
	assert.Len(t, allNodes, 3)

	// Remove all nodes in colony1
	err = db.RemoveNodesByColonyName(colony1.Name)
	assert.Nil(t, err)

	// Verify colony1 nodes are gone
	colony1Nodes, err := db.GetNodes(colony1.Name)
	assert.Nil(t, err)
	assert.Len(t, colony1Nodes, 0)

	// Verify colony2 node still exists
	colony2Nodes, err := db.GetNodes(colony2.Name)
	assert.Nil(t, err)
	assert.Len(t, colony2Nodes, 1)
	assert.Equal(t, node3.ID, colony2Nodes[0].ID)
}

func TestCountNodes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	// Count nodes when empty
	count, err := db.CountNodes("")
	assert.Nil(t, err)
	assert.Equal(t, 0, count)

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	node1 := core.CreateNode("node1", colony1.Name, "location1")
	err = db.AddNode(node1)
	assert.Nil(t, err)

	node2 := core.CreateNode("node2", colony1.Name, "location2")
	err = db.AddNode(node2)
	assert.Nil(t, err)

	node3 := core.CreateNode("node3", colony2.Name, "location3")
	err = db.AddNode(node3)
	assert.Nil(t, err)

	// Count all nodes
	count, err = db.CountNodes("")
	assert.Nil(t, err)
	assert.Equal(t, 3, count)

	// Count nodes for colony1
	count, err = db.CountNodes(colony1.Name)
	assert.Nil(t, err)
	assert.Equal(t, 2, count)

	// Count nodes for colony2
	count, err = db.CountNodes(colony2.Name)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)

	// Count nodes for non-existent colony
	count, err = db.CountNodes("non_existent_colony")
	assert.Nil(t, err)
	assert.Equal(t, 0, count)
}

func TestNodeLabelsAndCapabilities(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	node := core.CreateNode("test-node", colony.Name, "test-location")
	node.Capabilities = []string{"docker", "kubernetes", "gpu", "tpu"}
	node.Labels["gpu.0.name"] = "NVIDIA A100"
	node.Labels["gpu.0.memory"] = "40GB"
	node.Labels["gpu.1.name"] = "NVIDIA A100"
	node.Labels["gpu.1.memory"] = "40GB"
	node.Labels["region"] = "us-west-2"
	node.Labels["availability-zone"] = "us-west-2a"

	err = db.AddNode(node)
	assert.Nil(t, err)

	nodeFromDB, err := db.GetNodeByID(node.ID)
	assert.Nil(t, err)
	assert.NotNil(t, nodeFromDB)
	assert.Equal(t, 4, len(nodeFromDB.Capabilities))
	assert.Contains(t, nodeFromDB.Capabilities, "docker")
	assert.Contains(t, nodeFromDB.Capabilities, "kubernetes")
	assert.Contains(t, nodeFromDB.Capabilities, "gpu")
	assert.Contains(t, nodeFromDB.Capabilities, "tpu")
	assert.Equal(t, 6, len(nodeFromDB.Labels))
	assert.Equal(t, "NVIDIA A100", nodeFromDB.Labels["gpu.0.name"])
	assert.Equal(t, "40GB", nodeFromDB.Labels["gpu.0.memory"])
	assert.Equal(t, "us-west-2", nodeFromDB.Labels["region"])
	assert.Equal(t, "us-west-2a", nodeFromDB.Labels["availability-zone"])
}

func TestNodeTimestamps(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	node := core.CreateNode("test-node", colony.Name, "test-location")
	originalCreated := node.Created
	originalLastSeen := node.LastSeen

	err = db.AddNode(node)
	assert.Nil(t, err)

	nodeFromDB, err := db.GetNodeByID(node.ID)
	assert.Nil(t, err)
	assert.NotNil(t, nodeFromDB)

	// Verify timestamps are preserved
	assert.Equal(t, originalCreated.Unix(), nodeFromDB.Created.Unix())
	assert.Equal(t, originalLastSeen.Unix(), nodeFromDB.LastSeen.Unix())
	assert.False(t, nodeFromDB.Created.IsZero())
	assert.False(t, nodeFromDB.LastSeen.IsZero())

	// Update LastSeen
	time.Sleep(100 * time.Millisecond)
	node.TouchLastSeen()
	err = db.UpdateNode(node)
	assert.Nil(t, err)

	nodeFromDB, err = db.GetNodeByID(node.ID)
	assert.Nil(t, err)
	assert.True(t, nodeFromDB.LastSeen.After(originalLastSeen))
	assert.Equal(t, originalCreated.Unix(), nodeFromDB.Created.Unix())
}
