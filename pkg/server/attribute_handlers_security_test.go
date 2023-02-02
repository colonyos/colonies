package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddAttributeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.AssignProcess(env.colony1ID, -1, env.executor1PrvKey)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(assignedProcess.ID, env.colony1ID, "", core.OUT, "result", "helloworld")
	_, err = client.AddAttribute(attribute, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Add another executor to colony1 and try to set an attribute in the assigned processes assigned to
	// executor1, it should not be possible
	executor3, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.colony1ID)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor3, env.colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(executor3.ID, env.colony1PrvKey)
	assert.Nil(t, err)
	_, err = client.AddAttribute(attribute, executor3PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddAttribute(attribute, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetAttributeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.AssignProcess(env.colony1ID, -1, env.executor1PrvKey)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(assignedProcess.ID, env.colony1ID, "", core.OUT, "result", "helloworld")
	_, err = client.AddAttribute(attribute, env.executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetAttribute(attribute.ID, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetAttribute(attribute.ID, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}
