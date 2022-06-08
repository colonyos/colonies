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
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(assignedProcess.ID, env.colony1ID, "", core.OUT, "result", "helloworld")
	_, err = client.AddAttribute(attribute, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Add another runtime to colony1 and try to set an attribute in the assigned processes assigned to
	// runtime1, it should not be possible
	runtime3, runtime3PrvKey, err := utils.CreateTestRuntimeWithKey(env.colony1ID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(runtime3, env.colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(runtime3.ID, env.colony1PrvKey)
	assert.Nil(t, err)
	_, err = client.AddAttribute(attribute, runtime3PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddAttribute(attribute, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetAttributeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(assignedProcess.ID, env.colony1ID, "", core.OUT, "result", "helloworld")
	_, err = client.AddAttribute(attribute, env.runtime1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetAttribute(attribute.ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetAttribute(attribute.ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}
