package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateNode(t *testing.T) {
	nodeName := "test-node"
	colonyName := "test-colony"
	location := "test-location"

	node := CreateNode(nodeName, colonyName, location)

	assert.Len(t, node.ID, 64)
	assert.Equal(t, nodeName, node.Name)
	assert.Equal(t, colonyName, node.ColonyName)
	assert.Equal(t, location, node.Location)
	assert.Equal(t, "ready", node.State)
	assert.NotNil(t, node.Labels)
	assert.NotNil(t, node.Capabilities)
	assert.Equal(t, 0, len(node.Capabilities))
	assert.False(t, node.Created.IsZero())
	assert.False(t, node.LastSeen.IsZero())
}

func TestNodeEquals(t *testing.T) {
	node1 := CreateNode("node1", "colony1", "location1")
	node1.Platform = "linux"
	node1.Architecture = "amd64"

	node2 := CreateNode("node1", "colony1", "location1")
	node2.ID = node1.ID
	node2.Platform = "linux"
	node2.Architecture = "amd64"

	node3 := CreateNode("node2", "colony1", "location1")
	node4 := CreateNode("node1", "colony2", "location1")
	node5 := CreateNode("node1", "colony1", "location2")

	assert.True(t, node1.Equals(node1))
	assert.True(t, node1.Equals(node2))
	assert.False(t, node1.Equals(node3))
	assert.False(t, node1.Equals(node4))
	assert.False(t, node1.Equals(node5))
	assert.False(t, node1.Equals(nil))
}

func TestNodeToJSON(t *testing.T) {
	node1 := CreateNode("test-node", "test-colony", "test-location")
	node1.Platform = "linux"
	node1.Architecture = "amd64"
	node1.CPU = 32
	node1.Memory = 64000
	node1.GPU = 1
	node1.Capabilities = []string{"docker", "gpu"}
	node1.Labels = map[string]string{"key1": "value1", "key2": "value2"}

	jsonString, err := node1.ToJSON()
	assert.Nil(t, err)
	assert.NotEmpty(t, jsonString)

	node2, err := ConvertJSONToNode(jsonString + "error")
	assert.NotNil(t, err)

	node2, err = ConvertJSONToNode(jsonString)
	assert.Nil(t, err)
	assert.True(t, node2.Equals(node1))
	assert.Equal(t, node1.CPU, node2.CPU)
	assert.Equal(t, node1.Memory, node2.Memory)
	assert.Equal(t, node1.GPU, node2.GPU)
	assert.Equal(t, len(node1.Capabilities), len(node2.Capabilities))
	assert.Equal(t, len(node1.Labels), len(node2.Labels))
}

func TestConvertNodesToJSON(t *testing.T) {
	node1 := CreateNode("node1", "colony1", "location1")
	node1.Platform = "linux"
	node2 := CreateNode("node2", "colony1", "location2")
	node2.Platform = "darwin"

	nodes := []*Node{node1, node2}

	jsonString, err := ConvertNodesToJSON(nodes)
	assert.Nil(t, err)
	assert.NotEmpty(t, jsonString)

	nodes2, err := ConvertJSONToNodes(jsonString + "error")
	assert.NotNil(t, err)

	nodes2, err = ConvertJSONToNodes(jsonString)
	assert.Nil(t, err)
	assert.Equal(t, len(nodes), len(nodes2))
	assert.Equal(t, node1.Name, nodes2[0].Name)
	assert.Equal(t, node2.Name, nodes2[1].Name)
}

func TestTouchLastSeen(t *testing.T) {
	node := CreateNode("test-node", "test-colony", "test-location")

	originalLastSeen := node.LastSeen
	time.Sleep(10 * time.Millisecond)

	node.TouchLastSeen()

	assert.True(t, node.LastSeen.After(originalLastSeen))
}

func TestUpdateFromMetadata(t *testing.T) {
	node := CreateNode("test-node", "test-colony", "original-location")

	metadata := &NodeMetadata{
		Hostname:     "test-hostname",
		Location:     "new-location",
		Platform:     "linux",
		Architecture: "amd64",
		CPU:          32,
		Memory:       64000,
		GPU:          1,
		Capabilities: []string{"docker", "gpu"},
		Labels: map[string]string{
			"gpu.0.name": "NVIDIA RTX 3080",
			"region":     "us-west",
		},
	}

	node.UpdateFromMetadata(metadata)

	assert.Equal(t, "new-location", node.Location)
	assert.Equal(t, "linux", node.Platform)
	assert.Equal(t, "amd64", node.Architecture)
	assert.Equal(t, 32, node.CPU)
	assert.Equal(t, int64(64000), node.Memory)
	assert.Equal(t, 1, node.GPU)
	assert.Equal(t, 2, len(node.Capabilities))
	assert.Equal(t, "docker", node.Capabilities[0])
	assert.Equal(t, "gpu", node.Capabilities[1])
	assert.Equal(t, 2, len(node.Labels))
	assert.Equal(t, "NVIDIA RTX 3080", node.Labels["gpu.0.name"])
	assert.Equal(t, "us-west", node.Labels["region"])
	assert.Equal(t, "ready", node.State)
}

func TestUpdateFromMetadataWithNil(t *testing.T) {
	node := CreateNode("test-node", "test-colony", "test-location")
	node.Platform = "linux"
	originalPlatform := node.Platform

	node.UpdateFromMetadata(nil)

	// Should not panic and should not change anything
	assert.Equal(t, originalPlatform, node.Platform)
}

func TestUpdateFromMetadataEmptyLocation(t *testing.T) {
	node := CreateNode("test-node", "test-colony", "original-location")

	metadata := &NodeMetadata{
		Platform:     "linux",
		Architecture: "amd64",
		Location:     "", // Empty location
	}

	node.UpdateFromMetadata(metadata)

	// Location should remain unchanged if metadata location is empty
	assert.Equal(t, "original-location", node.Location)
	assert.Equal(t, "linux", node.Platform)
}

func TestUpdateFromMetadataLabelMerge(t *testing.T) {
	node := CreateNode("test-node", "test-colony", "test-location")
	node.Labels["existing"] = "label"

	metadata := &NodeMetadata{
		Platform: "linux",
		Labels: map[string]string{
			"new":      "label",
			"existing": "updated",
		},
	}

	node.UpdateFromMetadata(metadata)

	assert.Equal(t, 2, len(node.Labels))
	assert.Equal(t, "updated", node.Labels["existing"]) // Should be updated
	assert.Equal(t, "label", node.Labels["new"])        // Should be added
}
