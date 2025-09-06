package processgraph_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/service"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestSubmitWorkflowSpecSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	diamond := server.GenerateDiamondtWorkflowSpec(env.Colony1Name)

	_, err := client.SubmitWorkflowSpec(diamond, env.Executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.SubmitWorkflowSpec(diamond, env.Colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.SubmitWorkflowSpec(diamond, env.Colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.SubmitWorkflowSpec(diamond, env.Executor1PrvKey)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestGetProcessGraphSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	diamond := server.GenerateDiamondtWorkflowSpec(env.Colony1Name)
	graph, err := client.SubmitWorkflowSpec(diamond, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetProcessGraph(graph.ID, env.Executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetProcessGraph(graph.ID, env.Colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetProcessGraph(graph.ID, env.Colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetProcessGraph(graph.ID, env.Executor1PrvKey)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestGetProcessGraphsSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.GetWaitingProcessGraphs(env.Colony1Name, 100, env.Executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetWaitingProcessGraphs(env.Colony1Name, 100, env.Colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetWaitingProcessGraphs(env.Colony1Name, 100, env.Colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetWaitingProcessGraphs(env.Colony1Name, 100, env.Executor1PrvKey)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveProcessGraphSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	diamond := server.GenerateDiamondtWorkflowSpec(env.Colony1Name)
	graph, err := client.SubmitWorkflowSpec(diamond, env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.RemoveProcessGraph(graph.ID, env.Executor2PrvKey)
	assert.NotNil(t, err)
	err = client.RemoveProcessGraph(graph.ID, env.Colony1PrvKey)
	assert.NotNil(t, err)
	err = client.RemoveProcessGraph(graph.ID, env.Colony2PrvKey)
	assert.NotNil(t, err)
	err = client.RemoveProcessGraph(graph.ID, env.Executor1PrvKey)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveAllProcessGraphsSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	diamond := server.GenerateDiamondtWorkflowSpec(env.Colony1Name)
	_, err := client.SubmitWorkflowSpec(diamond, env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.RemoveAllProcessGraphs(env.Colony1Name, env.Executor2PrvKey)
	assert.NotNil(t, err)
	err = client.RemoveAllProcessGraphs(env.Colony1Name, env.Colony1PrvKey)
	assert.Nil(t, err)
	err = client.RemoveAllProcessGraphs(env.Colony1Name, env.Colony2PrvKey)
	assert.NotNil(t, err)
	err = client.RemoveAllProcessGraphs(env.Colony1Name, env.Executor1PrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestAddChildSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	executor, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.Colony2Name)
	assert.Nil(t, err)
	executor3, err := client.AddExecutor(executor, env.Colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.Colony2Name, executor3.Name, env.Colony2PrvKey)
	assert.Nil(t, err)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2
	//   executor3 is member of colony2

	diamond := server.GenerateDiamondtWorkflowSpec(env.Colony2Name)
	processGraph, err := client.SubmitWorkflowSpec(diamond, env.Executor2PrvKey)
	assert.Nil(t, err)

	parentProcessID := processGraph.Roots[0]

	childFunctionSpec := utils.CreateTestFunctionSpec(env.Colony2Name)
	childFunctionSpec.NodeName = "task5"

	_, err = client.AddChild(processGraph.ID, parentProcessID, "", childFunctionSpec, false, env.Executor1PrvKey)
	assert.NotNil(t, err) // Error, executor1 not member of member of colony2

	_, err = client.AddChild(processGraph.ID, parentProcessID, "", childFunctionSpec, false, env.Colony1PrvKey)
	assert.NotNil(t, err) // Error, invalid prvkey

	_, err = client.AddChild(processGraph.ID, parentProcessID, "", childFunctionSpec, false, env.Colony2PrvKey)
	assert.NotNil(t, err) // Error, invalid prvkey

	_, err = client.AddChild(processGraph.ID, parentProcessID, "", childFunctionSpec, false, env.Executor2PrvKey)
	assert.NotNil(t, err) // Error, process must be running

	// Assign task1 to executor2
	_, err = client.Assign(env.Colony2Name, -1, "", "", env.Executor2PrvKey)
	assert.Nil(t, err)

	// Now, we should be able to add a child since we got assigned task1
	_, err = client.AddChild(processGraph.ID, parentProcessID, "", childFunctionSpec, false, executor3PrvKey)
	assert.NotNil(t, err) // Error, process is not assigned to executor3

	// But, executor2 should be able to add a child since process is assigned to executor2
	_, err = client.AddChild(processGraph.ID, parentProcessID, "", childFunctionSpec, false, env.Executor2PrvKey)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}
