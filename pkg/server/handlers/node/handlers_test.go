package node_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetNodes(t *testing.T) {
	env, client, _, _, done := server.SetupTestEnv2(t)

	// Create executors with node metadata to automatically create nodes
	executor1, _, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	executor1.NodeMetadata = &core.NodeMetadata{
		Hostname:     "test-node-1",
		Location:     "us-west",
		Platform:     "linux",
		Architecture: "amd64",
		CPU:          16,
	}
	_, err = client.AddExecutor(executor1, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor1.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	executor2, _, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	executor2.NodeMetadata = &core.NodeMetadata{
		Hostname:     "test-node-2",
		Location:     "eu-north",
		Platform:     "darwin",
		Architecture: "arm64",
		CPU:          8,
	}
	_, err = client.AddExecutor(executor2, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor2.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Get nodes using the client
	nodes, err := client.GetNodes(env.ColonyName, env.ColonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, nodes)
	assert.Len(t, nodes, 2)

	// Verify node details
	foundNode1 := false
	foundNode2 := false
	for _, node := range nodes {
		if node.Name == "test-node-1" {
			foundNode1 = true
			assert.Equal(t, "us-west", node.Location)
			assert.Equal(t, "linux", node.Platform)
			assert.Equal(t, "amd64", node.Architecture)
			assert.Equal(t, 16, node.CPU)
		}
		if node.Name == "test-node-2" {
			foundNode2 = true
			assert.Equal(t, "eu-north", node.Location)
			assert.Equal(t, "darwin", node.Platform)
			assert.Equal(t, "arm64", node.Architecture)
			assert.Equal(t, 8, node.CPU)
		}
	}
	assert.True(t, foundNode1)
	assert.True(t, foundNode2)

	<-done
}

func TestGetNode(t *testing.T) {
	env, client, _, _, done := server.SetupTestEnv2(t)

	// Create an executor with node metadata
	executor, _, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	executor.NodeMetadata = &core.NodeMetadata{
		Hostname:     "test-node",
		Location:     "test-location",
		Platform:     "linux",
		Architecture: "amd64",
		CPU:          32,
		Memory:       64000,
		GPU:          2,
		Capabilities: []string{"docker", "gpu"},
		Labels: map[string]string{
			"gpu.0.name": "NVIDIA RTX 3080",
		},
	}
	_, err = client.AddExecutor(executor, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Get node by name
	retrievedNode, err := client.GetNode(env.ColonyName, "test-node", env.ColonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedNode)
	assert.Equal(t, "test-node", retrievedNode.Name)
	assert.Equal(t, "test-location", retrievedNode.Location)
	assert.Equal(t, "linux", retrievedNode.Platform)
	assert.Equal(t, "amd64", retrievedNode.Architecture)
	assert.Equal(t, 32, retrievedNode.CPU)
	assert.Equal(t, int64(64000), retrievedNode.Memory)
	assert.Equal(t, 2, retrievedNode.GPU)
	assert.Equal(t, 2, len(retrievedNode.Capabilities))
	assert.Equal(t, "NVIDIA RTX 3080", retrievedNode.Labels["gpu.0.name"])

	<-done
}

func TestGetNodeNotFound(t *testing.T) {
	env, client, _, _, done := server.SetupTestEnv2(t)

	// Try to get a node that doesn't exist
	node, err := client.GetNode(env.ColonyName, "non-existent-node", env.ColonyPrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, node)

	<-done
}

func TestGetNodesByLocation(t *testing.T) {
	env, client, _, _, done := server.SetupTestEnv2(t)

	// Create executors in different locations
	executor1, _, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	executor1.NodeMetadata = &core.NodeMetadata{
		Hostname: "node1",
		Location: "us-west",
		Platform: "linux",
	}
	_, err = client.AddExecutor(executor1, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor1.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	executor2, _, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	executor2.NodeMetadata = &core.NodeMetadata{
		Hostname: "node2",
		Location: "us-west",
		Platform: "linux",
	}
	_, err = client.AddExecutor(executor2, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor2.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	executor3, _, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	executor3.NodeMetadata = &core.NodeMetadata{
		Hostname: "node3",
		Location: "eu-north",
		Platform: "darwin",
	}
	_, err = client.AddExecutor(executor3, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor3.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Get nodes in us-west
	usWestNodes, err := client.GetNodesByLocation(env.ColonyName, "us-west", env.ColonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, usWestNodes)
	assert.Len(t, usWestNodes, 2)
	for _, node := range usWestNodes {
		assert.Equal(t, "us-west", node.Location)
	}

	// Get nodes in eu-north
	euNorthNodes, err := client.GetNodesByLocation(env.ColonyName, "eu-north", env.ColonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, euNorthNodes)
	assert.Len(t, euNorthNodes, 1)
	assert.Equal(t, "node3", euNorthNodes[0].Name)

	// Get nodes in non-existent location
	emptyNodes, err := client.GetNodesByLocation(env.ColonyName, "asia-east", env.ColonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, emptyNodes)
	assert.Len(t, emptyNodes, 0)

	<-done
}

func TestGetNodesUnauthorized(t *testing.T) {
	env, client, _, serverPrvKey, done := server.SetupTestEnv2(t)

	// Create another colony
	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	// Create an executor with node metadata in env.ColonyName
	executor, _, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	executor.NodeMetadata = &core.NodeMetadata{
		Hostname: "test-node",
		Location: "test-location",
	}
	_, err = client.AddExecutor(executor, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Try to get nodes from env.ColonyName using colony2's private key (should fail)
	_, err = client.GetNodes(env.ColonyName, colony2PrvKey)
	assert.NotNil(t, err)

	<-done
}
