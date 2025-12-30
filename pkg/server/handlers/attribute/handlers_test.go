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

// TestAddAttributeProcessNotRunning tests adding attribute to a pending process
func TestAddAttributeProcessNotRunning(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	addedProcess, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	// Try to add attribute to a pending (not running) process
	attribute := core.CreateAttribute(addedProcess.ID, env.ColonyName, "", core.OUT, "result", "helloworld")
	_, err = client.AddAttribute(attribute, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestAddAttributeWrongExecutor tests adding attribute by a different executor
func TestAddAttributeWrongExecutor(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create second executor
	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor2.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err = client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Assign to first executor
	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to add attribute from second executor
	attribute := core.CreateAttribute(assignedProcess.ID, env.ColonyName, "", core.OUT, "result", "helloworld")
	_, err = client.AddAttribute(attribute, executor2PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestAddAttributeUnauthorized tests adding attribute from different colony
func TestAddAttributeUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	funcSpec := utils.CreateTestFunctionSpec(colony1.Name)
	_, err = client.Submit(funcSpec, executor1PrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(colony1.Name, -1, "", "", executor1PrvKey)
	assert.Nil(t, err)

	// Try to add attribute from executor in colony2
	attribute := core.CreateAttribute(assignedProcess.ID, colony1.Name, "", core.OUT, "result", "helloworld")
	_, err = client.AddAttribute(attribute, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetAttributeUnauthorized tests getting attribute from different colony
func TestGetAttributeUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	funcSpec := utils.CreateTestFunctionSpec(colony1.Name)
	_, err = client.Submit(funcSpec, executor1PrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(colony1.Name, -1, "", "", executor1PrvKey)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(assignedProcess.ID, colony1.Name, "", core.OUT, "result", "helloworld")
	addedAttribute, err := client.AddAttribute(attribute, executor1PrvKey)
	assert.Nil(t, err)

	// Try to get attribute from executor in colony2
	_, err = client.GetAttribute(addedAttribute.ID, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}
