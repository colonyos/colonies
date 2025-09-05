package attribute_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddGetAttributes(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	addedProcess, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(core.GenerateRandomID(), env.ColonyName, "", core.OUT, "result", "helloworld")
	_, err = client.AddAttribute(attribute, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	attribute = core.CreateAttribute(assignedProcess.ID, env.ColonyName, "", core.OUT, "result", "helloworld")
	addedAttribute, err := client.AddAttribute(attribute, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, attribute.ID, addedAttribute.ID)

	assignedProcessFromServer, err := client.GetProcess(assignedProcess.ID, env.ExecutorPrvKey)

	out := make(map[string]string)
	for _, attribute := range assignedProcessFromServer.Attributes {
		out[attribute.Key] = attribute.Value
	}

	assert.Equal(t, "helloworld", out["result"])

	_, err = client.GetAttribute(core.GenerateRandomID(), env.ExecutorPrvKey)
	assert.NotNil(t, err) // Will not work, invalid target

	attributeFromServer, err := client.GetAttribute(attribute.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, attribute.ID, attributeFromServer.ID)

	server.Shutdown()
	<-done
}
