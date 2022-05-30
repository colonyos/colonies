package server

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubmitWorkflowSpec(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	diamond := generateDiamondtWorkflowSpec(env.colonyID)
	processgraph, err := client.SubmitWorkflowSpec(diamond, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, processgraph)

	processes, err := client.GetWaitingProcesses(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processes, 4)

	assignedProcess, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	fmt.Println(assignedProcess.ProcessSpec.Name)

	// We cannot be assigned more tasks until task1 is closed
	_, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.NotNil(t, err)

	// Close task1
	err = client.CloseSuccessful(assignedProcess.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)

	fmt.Println(assignedProcess.ProcessSpec.Name)

	server.Shutdown()
	<-done
}
