package node_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetNodesSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Create an executor with node metadata in colony1
	executor1, _, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)
	executor1.NodeMetadata = &core.NodeMetadata{
		Hostname:     "colony1-node",
		Location:     "us-west",
		Platform:     "linux",
		Architecture: "amd64",
		CPU:          16,
	}
	_, err = client.AddExecutor(executor1, env.Colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.Colony1Name, executor1.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Try to get nodes from colony1 using colony2's private key (should fail)
	_, err = client.GetNodes(env.Colony1Name, env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try to get nodes from colony1 using executor2's private key (should fail)
	_, err = client.GetNodes(env.Colony1Name, env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Get nodes from colony1 using colony1's private key (should succeed)
	nodes, err := client.GetNodes(env.Colony1Name, env.Colony1PrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, nodes)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "colony1-node", nodes[0].Name)

	// Get nodes from colony1 using executor1's private key (should succeed)
	nodes, err = client.GetNodes(env.Colony1Name, env.Executor1PrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, nodes)
	assert.Len(t, nodes, 1)

	server.Shutdown()
	<-done
}

func TestGetNodeSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Create an executor with node metadata in colony1
	executor1, _, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)
	executor1.NodeMetadata = &core.NodeMetadata{
		Hostname:     "colony1-node",
		Location:     "us-west",
		Platform:     "linux",
		Architecture: "amd64",
		CPU:          32,
	}
	_, err = client.AddExecutor(executor1, env.Colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.Colony1Name, executor1.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Try to get node from colony1 using colony2's private key (should fail)
	_, err = client.GetNode(env.Colony1Name, "colony1-node", env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try to get node from colony1 using executor2's private key (should fail)
	_, err = client.GetNode(env.Colony1Name, "colony1-node", env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Get node from colony1 using colony1's private key (should succeed)
	node, err := client.GetNode(env.Colony1Name, "colony1-node", env.Colony1PrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, "colony1-node", node.Name)
	assert.Equal(t, "us-west", node.Location)
	assert.Equal(t, 32, node.CPU)

	// Get node from colony1 using executor1's private key (should succeed)
	node, err = client.GetNode(env.Colony1Name, "colony1-node", env.Executor1PrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, "colony1-node", node.Name)

	server.Shutdown()
	<-done
}

func TestGetNodesByLocationSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Create executors with node metadata in colony1
	executor1, _, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)
	executor1.NodeMetadata = &core.NodeMetadata{
		Hostname:     "colony1-node-west",
		Location:     "us-west",
		Platform:     "linux",
		Architecture: "amd64",
	}
	_, err = client.AddExecutor(executor1, env.Colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.Colony1Name, executor1.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	executor2, _, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)
	executor2.NodeMetadata = &core.NodeMetadata{
		Hostname:     "colony1-node-east",
		Location:     "us-east",
		Platform:     "linux",
		Architecture: "amd64",
	}
	_, err = client.AddExecutor(executor2, env.Colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.Colony1Name, executor2.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Try to get nodes by location from colony1 using colony2's private key (should fail)
	_, err = client.GetNodesByLocation(env.Colony1Name, "us-west", env.Colony2PrvKey)
	assert.NotNil(t, err)

	// Try to get nodes by location from colony1 using executor2's private key (should fail)
	_, err = client.GetNodesByLocation(env.Colony1Name, "us-west", env.Executor2PrvKey)
	assert.NotNil(t, err)

	// Get nodes by location from colony1 using colony1's private key (should succeed)
	nodes, err := client.GetNodesByLocation(env.Colony1Name, "us-west", env.Colony1PrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, nodes)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "colony1-node-west", nodes[0].Name)

	// Get nodes by location from colony1 using executor1's private key (should succeed)
	nodes, err = client.GetNodesByLocation(env.Colony1Name, "us-east", env.Executor1PrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, nodes)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "colony1-node-east", nodes[0].Name)

	server.Shutdown()
	<-done
}

func TestGetNodesAcrossColonies(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// Create executor with node metadata in colony1
	executor1, _, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)
	executor1.NodeMetadata = &core.NodeMetadata{
		Hostname: "colony1-node",
		Location: "us-west",
	}
	_, err = client.AddExecutor(executor1, env.Colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.Colony1Name, executor1.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	// Create executor with node metadata in colony2
	executor2, _, err := utils.CreateTestExecutorWithKey(env.Colony2Name)
	assert.Nil(t, err)
	executor2.NodeMetadata = &core.NodeMetadata{
		Hostname: "colony2-node",
		Location: "us-west",
	}
	_, err = client.AddExecutor(executor2, env.Colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.Colony2Name, executor2.Name, env.Colony2PrvKey)
	assert.Nil(t, err)

	// Get nodes from colony1 - should only see colony1's nodes
	nodes, err := client.GetNodes(env.Colony1Name, env.Colony1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "colony1-node", nodes[0].Name)

	// Get nodes from colony2 - should only see colony2's nodes
	nodes, err = client.GetNodes(env.Colony2Name, env.Colony2PrvKey)
	assert.Nil(t, err)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "colony2-node", nodes[0].Name)

	// Get nodes by location from colony1 - should only see colony1's nodes
	nodes, err = client.GetNodesByLocation(env.Colony1Name, "us-west", env.Colony1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "colony1-node", nodes[0].Name)

	// Get nodes by location from colony2 - should only see colony2's nodes
	nodes, err = client.GetNodesByLocation(env.Colony2Name, "us-west", env.Colony2PrvKey)
	assert.Nil(t, err)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "colony2-node", nodes[0].Name)

	server.Shutdown()
	<-done
}
