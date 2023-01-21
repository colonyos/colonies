package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddGetAttributes(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	processSpec := utils.CreateTestProcessSpec(env.colonyID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.AssignProcess(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(core.GenerateRandomID(), env.colonyID, "", core.OUT, "result", "helloworld")
	_, err = client.AddAttribute(attribute, env.executorPrvKey)
	assert.NotNil(t, err)

	attribute = core.CreateAttribute(assignedProcess.ID, env.colonyID, "", core.OUT, "result", "helloworld")
	addedAttribute, err := client.AddAttribute(attribute, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, attribute.ID, addedAttribute.ID)

	assignedProcessFromServer, err := client.GetProcess(assignedProcess.ID, env.executorPrvKey)

	out := make(map[string]string)
	for _, attribute := range assignedProcessFromServer.Attributes {
		out[attribute.Key] = attribute.Value
	}

	assert.Equal(t, "helloworld", out["result"])

	_, err = client.GetAttribute(core.GenerateRandomID(), env.executorPrvKey)
	assert.NotNil(t, err) // Will not work, invalid target

	attributeFromServer, err := client.GetAttribute(attribute.ID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, attribute.ID, attributeFromServer.ID)

	server.Shutdown()
	<-done
}
