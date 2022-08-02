package server

import (
	"testing"

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
