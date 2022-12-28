package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestSubmitWorkflowSpecSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	diamond := generateDiamondtWorkflowSpec(env.colony1ID)

	_, err := client.SubmitWorkflowSpec(diamond, env.runtime2PrvKey)
	assert.NotNil(t, err)
	_, err = client.SubmitWorkflowSpec(diamond, env.colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.SubmitWorkflowSpec(diamond, env.colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.SubmitWorkflowSpec(diamond, env.runtime1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetProcessGraphSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	diamond := generateDiamondtWorkflowSpec(env.colony1ID)
	graph, err := client.SubmitWorkflowSpec(diamond, env.runtime1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetProcessGraph(graph.ID, env.runtime2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetProcessGraph(graph.ID, env.colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetProcessGraph(graph.ID, env.colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetProcessGraph(graph.ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetProcessGraphsSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	_, err := client.GetWaitingProcessGraphs(env.colony1ID, 100, env.runtime2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetWaitingProcessGraphs(env.colony1ID, 100, env.colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetWaitingProcessGraphs(env.colony1ID, 100, env.colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetWaitingProcessGraphs(env.colony1ID, 100, env.runtime1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestDeleteProcessGraphSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	diamond := generateDiamondtWorkflowSpec(env.colony1ID)
	graph, err := client.SubmitWorkflowSpec(diamond, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = client.DeleteProcessGraph(graph.ID, env.runtime2PrvKey)
	assert.NotNil(t, err)
	err = client.DeleteProcessGraph(graph.ID, env.colony1PrvKey)
	assert.NotNil(t, err)
	err = client.DeleteProcessGraph(graph.ID, env.colony2PrvKey)
	assert.NotNil(t, err)
	err = client.DeleteProcessGraph(graph.ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestDeleteAllProcessGraphsSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	diamond := generateDiamondtWorkflowSpec(env.colony1ID)
	_, err := client.SubmitWorkflowSpec(diamond, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = client.DeleteAllProcessGraphs(env.colony1ID, env.runtime2PrvKey)
	assert.NotNil(t, err)
	err = client.DeleteAllProcessGraphs(env.colony1ID, env.colony1PrvKey)
	assert.Nil(t, err)
	err = client.DeleteAllProcessGraphs(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err)
	err = client.DeleteAllProcessGraphs(env.colony1ID, env.runtime1PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestAddChildSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	runtime, runtime3PrvKey, err := utils.CreateTestRuntimeWithKey(env.colony2ID)
	assert.Nil(t, err)
	runtime3, err := client.AddRuntime(runtime, env.colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(runtime3.ID, env.colony2PrvKey)
	assert.Nil(t, err)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2
	//   runtime3 is member of colony2

	diamond := generateDiamondtWorkflowSpec(env.colony2ID)
	processGraph, err := client.SubmitWorkflowSpec(diamond, env.runtime2PrvKey)
	assert.Nil(t, err)

	parentProcessID := processGraph.Roots[0]

	childProcessSpec := utils.CreateTestProcessSpec(env.colony2ID)
	childProcessSpec.Name = "task5"

	_, err = client.AddChild(processGraph.ID, parentProcessID, childProcessSpec, env.runtime1PrvKey)
	assert.NotNil(t, err) // Error, runtime1 not member of member of colony2

	_, err = client.AddChild(processGraph.ID, parentProcessID, childProcessSpec, env.colony1PrvKey)
	assert.NotNil(t, err) // Error, invalid prvkey

	_, err = client.AddChild(processGraph.ID, parentProcessID, childProcessSpec, env.colony2PrvKey)
	assert.NotNil(t, err) // Error, invalid prvkey

	_, err = client.AddChild(processGraph.ID, parentProcessID, childProcessSpec, env.runtime2PrvKey)
	assert.NotNil(t, err) // Error, process must be running

	// Assign task1 to runtime2
	_, err = client.AssignProcess(env.colony2ID, -1, env.runtime2PrvKey)
	assert.Nil(t, err)

	// Now, we should be able to add a child since we got assigned task1
	_, err = client.AddChild(processGraph.ID, parentProcessID, childProcessSpec, runtime3PrvKey)
	assert.NotNil(t, err) // Error, process is not assigned to runtime3

	// But, runtime2 should be able to add a child since process is assigned to runtime2
	_, err = client.AddChild(processGraph.ID, parentProcessID, childProcessSpec, env.runtime2PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}
