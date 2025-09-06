package attribute_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/service"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddAttributeSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	addedProcess, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(assignedProcess.ID, env.Colony1Name, "", core.OUT, "result", "helloworld")
	_, err = client.AddAttribute(attribute, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Add another executor to colony1 and try to set an attribute in the assigned processes assigned to
	// executor1, it should not be possible
	executor3, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor3, env.Colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.Colony1Name, executor3.Name, env.Colony1PrvKey)
	assert.Nil(t, err)
	_, err = client.AddAttribute(attribute, executor3PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddAttribute(attribute, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetAttributeSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	addedProcess, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(assignedProcess.ID, env.Colony1Name, "", core.OUT, "result", "helloworld")
	_, err = client.AddAttribute(attribute, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetAttribute(attribute.ID, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetAttribute(attribute.ID, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}
